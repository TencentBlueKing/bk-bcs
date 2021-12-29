/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dbsharding

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"bk-bscp/internal/database"
	pbcommon "bk-bscp/internal/protocol/common"
)

var (
	// RECORDNOTFOUND returns a "record not found error".
	// Occurs only when attempting to query the database with a struct,
	// querying with a slice won't return this error.
	RECORDNOTFOUND = gorm.ErrRecordNotFound

	// BSCPDBKEY is the key of bscp system sharding db.
	BSCPDBKEY = "BSCPDBKEY"
)

// DBService is DB service instance.
type DBService struct {
	// DB service id.
	ID string

	// DB service host.
	Host string

	// DB service port.
	Port int

	// DB username.
	User string

	// DB password.
	Password string
}

// DBConfigTemplate is DB service config template.
type DBConfigTemplate struct {
	// mysql connect timeout.
	ConnTimeout time.Duration

	// mysql read timeout.
	ReadTimeout time.Duration

	// mysql write timeout.
	WriteTimeout time.Duration

	// max num of connections.
	MaxOpenConns int

	// max num of idle connections.
	MaxIdleConns int

	// max life time of connection.
	KeepAlive time.Duration
}

// ShardingDB is database sharding result.
type ShardingDB struct {
	// DB service id.
	DBID string

	// database name.
	DBName string

	// gorm database handler.
	db *gorm.DB
}

// DB returns DB handler.
func (sd *ShardingDB) DB() *gorm.DB {
	return sd.db
}

// Config of ShardingMgr.
type Config struct {
	// DBHost is database service host.
	DBHost string

	// DBPort is database service port.
	DBPort int

	// DBUser is database username.
	DBUser string

	// DBPasswd is database password string.
	DBPasswd string

	// Size is dbsharding manager cache size.
	Size int

	// PurgeInterval is the purge internal of sharding cache.
	PurgeInterval time.Duration
}

// ShardingManager is sharing manager.
type ShardingManager struct {
	// config for sharding manager.
	config *Config

	// gorm database handler for sharding manager.
	db *gorm.DB

	// config template of db service instance.
	configTemplate *DBConfigTemplate

	// local cache for sharding databases, KEY -> *ShardingDB
	shardings gcache.Cache

	// local cache for db service instance, DBID -> *DBService
	dbServices gcache.Cache

	// local cache for gorm db handler, DBID -> DB handler(*gorm.DB)
	dbs gcache.Cache

	// rw mutex used for db client update without repetition.
	repeatMu sync.RWMutex
}

// NewShardingMgr creates a new sharding manager.
func NewShardingMgr(config *Config, configTemplate *DBConfigTemplate) *ShardingManager {
	mgr := &ShardingManager{config: config, configTemplate: configTemplate}

	if mgr.config.PurgeInterval < time.Minute {
		// just minutes level interval, do not purge too frequent.
		mgr.config.PurgeInterval = time.Minute
	}

	// build caches.
	mgr.shardings = gcache.New(config.Size).EvictType(gcache.TYPE_LRU).Build()
	mgr.dbServices = gcache.New(config.Size).EvictType(gcache.TYPE_LRU).Build()
	mgr.dbs = gcache.New(config.Size).EvictType(gcache.TYPE_LRU).EvictedFunc(mgr.evicteDB).Build()

	return mgr
}

// Init initializes new sharding manager.
func (mgr *ShardingManager) Init() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s&charset=%s",
		mgr.config.DBUser,
		mgr.config.DBPasswd,
		mgr.config.DBHost,
		mgr.config.DBPort,
		database.BSCPDB,
		mgr.configTemplate.ConnTimeout,
		mgr.configTemplate.ReadTimeout,
		mgr.configTemplate.WriteTimeout,
		database.BSCPCHARSET,
	)

	db, err := gorm.Open(mysql.Open(dsn),
		&gorm.Config{NowFunc: func() time.Time { return time.Now().Local().Round(time.Microsecond) }})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(mgr.configTemplate.MaxOpenConns)
	sqlDB.SetMaxIdleConns(mgr.configTemplate.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(mgr.configTemplate.KeepAlive)

	mgr.db = db

	// start purging sharding cache.
	go func() {
		ticker := time.NewTicker(mgr.config.PurgeInterval)
		defer ticker.Stop()

		for {
			<-ticker.C
			mgr.PurgeSharding()
			log.Print("purge sharding cache success!")
		}
	}()

	return nil
}

// newDB make a new db handler base on gorm database connection.
func (mgr *ShardingManager) newDB(service *DBService, dbname string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s&charset=%s",
		service.User,
		service.Password,
		service.Host,
		service.Port,
		dbname,
		mgr.configTemplate.ConnTimeout,
		mgr.configTemplate.ReadTimeout,
		mgr.configTemplate.WriteTimeout,
		database.BSCPCHARSET,
	)

	db, err := gorm.Open(mysql.Open(dsn),
		&gorm.Config{NowFunc: func() time.Time { return time.Now().Local().Round(time.Microsecond) }})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(mgr.configTemplate.MaxOpenConns)
	sqlDB.SetMaxIdleConns(mgr.configTemplate.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(mgr.configTemplate.KeepAlive)

	return db, nil
}

func (mgr *ShardingManager) evicteDB(k, v interface{}) {
	db, ok := v.(*gorm.DB)
	if !ok || db == nil {
		return
	}

	sqlDB, err := db.DB()
	if err != nil || sqlDB == nil {
		return
	}
	sqlDB.Close()
}

func (mgr *ShardingManager) getSharding(key string) (*ShardingDB, error) {
	var target *ShardingDB

	sd, err := mgr.shardings.Get(key)
	if err != nil || sd == nil {
		var st database.Sharding

		if err := mgr.db.Where("Fkey = ?", key).First(&st).Error; err != nil {
			return nil, err
		}

		target = &ShardingDB{DBID: st.DBID, DBName: st.DBName}
		mgr.shardings.Set(key, target)

	} else {
		v, ok := sd.(*ShardingDB)
		if !ok || v == nil {
			return nil, errors.New("can't get sharding database, invalid shardings cache struct")
		}
		target = v
	}

	return target, nil
}

func (mgr *ShardingManager) getDBService(dbID string) (*DBService, error) {
	var target *DBService

	service, err := mgr.dbServices.Get(dbID)
	if err != nil || service == nil {
		var st database.ShardingDB

		if err := mgr.db.
			Where("Fdb_id = ? AND Fstate = ?", dbID, pbcommon.CommonState_CS_VALID).
			First(&st).Error; err != nil {
			return nil, err
		}

		target = &DBService{
			ID:       st.DBID,
			Host:     st.Host,
			Port:     int(st.Port),
			User:     st.User,
			Password: st.Password,
		}
		mgr.dbServices.Set(st.DBID, target)

	} else {
		v, ok := service.(*DBService)
		if !ok || v == nil {
			return nil, errors.New("can't get sharding database, invalid dbServices cache struct")
		}
		target = v
	}

	return target, nil
}

func (mgr *ShardingManager) dbSDKey(dbID, dbName string) string {
	return dbID + "-" + dbName
}

func (mgr *ShardingManager) getDB(dbID, dbName string) (*gorm.DB, error) {
	var target *gorm.DB

	dbSDKey := mgr.dbSDKey(dbID, dbName)

	db, err := mgr.dbs.Get(dbSDKey)
	if err != nil || db == nil {
		service, err := mgr.getDBService(dbID)
		if err != nil {
			return nil, err
		}

		// make a new db connection with db service instance.
		newDB, err := mgr.newDB(service, dbName)
		if err != nil {
			return nil, err
		}

		// update new database client without repetition.
		mgr.repeatMu.Lock()
		defer mgr.repeatMu.Unlock()

		oldDB, err := mgr.dbs.Get(dbSDKey)
		if err != nil {
			target = newDB
			mgr.dbs.Set(dbSDKey, target)

		} else {
			v, ok := oldDB.(*gorm.DB)
			if !ok || v == nil {
				return nil, errors.New("can't get sharding database, invalid dbs cache struct")
			}
			target = v

			sqlDB, err := newDB.DB()
			if err == nil {
				sqlDB.Close()
			}
		}

	} else {
		v, ok := db.(*gorm.DB)
		if !ok || v == nil {
			return nil, errors.New("can't get sharding database, invalid dbs cache struct")
		}
		target = v
	}

	return target, nil
}

// ShardingDB gets the sharding result of target key.
func (mgr *ShardingManager) ShardingDB(key string) (*ShardingDB, error) {
	// default key is the BSCP system sharding database.
	if len(key) == 0 {
		key = BSCPDBKEY
	}

	if key == BSCPDBKEY {
		return &ShardingDB{DBName: database.BSCPDB, db: mgr.db}, nil
	}

	// get ShardingDB from shardings cache.
	sd, err := mgr.getSharding(key)
	if err != nil {
		return nil, err
	}

	// target db service instance client.
	db, err := mgr.getDB(sd.DBID, sd.DBName)
	if err != nil {
		return nil, err
	}

	// return target sharding result, include dbname and db handler.
	shardingDB := &ShardingDB{
		DBID:   sd.DBID,
		DBName: sd.DBName,
		db:     db,
	}

	return shardingDB, nil
}

// PurgeSharding purges sharding cache.
func (mgr *ShardingManager) PurgeSharding() {
	mgr.shardings.Purge()
}

// Close closes sharding manager.
func (mgr *ShardingManager) Close() error {
	sqlDB, err := mgr.db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Close(); err != nil {
		return err
	}

	for _, dbSDKey := range mgr.dbs.Keys(false) {
		v, err := mgr.dbs.Get(dbSDKey)
		if err != nil {
			continue
		}

		db, ok := v.(*gorm.DB)
		if !ok {
			continue
		}

		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}

	return nil
}

// CreateShardingDB create a new sharding database.
func (mgr *ShardingManager) CreateShardingDB(db *pbcommon.ShardingDB) error {
	st := &database.ShardingDB{
		DBID:     db.DbId,
		Host:     db.Host,
		Port:     db.Port,
		User:     db.User,
		Password: db.Password,
		Memo:     db.Memo,
		State:    db.State,
	}
	return mgr.db.Create(st).Error
}

// QueryShardingDB returns target sharding database.
func (mgr *ShardingManager) QueryShardingDB(dbID string) (*pbcommon.ShardingDB, error) {
	var st database.ShardingDB
	if err := mgr.db.Where("Fdb_id = ?", dbID).First(&st).Error; err != nil {
		return nil, err
	}

	db := &pbcommon.ShardingDB{
		DbId:      st.DBID,
		Host:      st.Host,
		Port:      st.Port,
		User:      st.User,
		Password:  st.Password,
		Memo:      st.Memo,
		State:     st.State,
		CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	return db, nil
}

// QueryShardingDBList returns all sharding databases.
func (mgr *ShardingManager) QueryShardingDBList() ([]*pbcommon.ShardingDB, error) {
	var sts []database.ShardingDB

	if err := mgr.db.
		Where("Fstate = ?", pbcommon.CommonState_CS_VALID).
		Find(&sts).Error; err != nil {
		return nil, err
	}

	var dbs []*pbcommon.ShardingDB

	for _, st := range sts {
		db := &pbcommon.ShardingDB{
			DbId:      st.DBID,
			Host:      st.Host,
			Port:      st.Port,
			User:      st.User,
			Password:  st.Password,
			Memo:      st.Memo,
			State:     st.State,
			CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		dbs = append(dbs, db)
	}

	return dbs, nil
}

// QueryShardingDBCount returns all sharding databases count.
func (mgr *ShardingManager) QueryShardingDBCount() (int64, error) {
	var totalCount int64

	if err := mgr.db.
		Model(&database.ShardingDB{}).
		Where("Fstate = ?", pbcommon.CommonState_CS_VALID).
		Count(&totalCount).Error; err != nil {
		return 0, err
	}
	return totalCount, nil
}

// UpdateShardingDB updates target sharding database.
func (mgr *ShardingManager) UpdateShardingDB(db *pbcommon.ShardingDB) error {
	ups := map[string]interface{}{
		"Host":     db.Host,
		"Port":     db.Port,
		"User":     db.User,
		"Password": db.Password,
		"Memo":     db.Memo,
		"State":    db.State,
	}
	return mgr.db.Model(&database.ShardingDB{}).Where("Fdb_id = ?", db.DbId).Updates(ups).Error
}

// CreateSharding creates a new sharding relation.
func (mgr *ShardingManager) CreateSharding(sharding *pbcommon.Sharding) error {
	st := &database.Sharding{
		Key:    sharding.Key,
		DBID:   sharding.DbId,
		DBName: sharding.DbName,
		Memo:   sharding.Memo,
		State:  sharding.State,
	}
	return mgr.db.Create(st).Error
}

// QuerySharding returns target sharding by key.
func (mgr *ShardingManager) QuerySharding(key string) (*pbcommon.Sharding, error) {
	var st database.Sharding
	if err := mgr.db.Where("Fkey = ?", key).First(&st).Error; err != nil {
		return nil, err
	}

	sharding := &pbcommon.Sharding{
		Key:       st.Key,
		DbId:      st.DBID,
		DbName:    st.DBName,
		Memo:      st.Memo,
		State:     st.State,
		CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	return sharding, nil
}

// QueryShardingList returns all shardings.
func (mgr *ShardingManager) QueryShardingList() ([]*pbcommon.Sharding, error) {
	var sts []database.Sharding

	if err := mgr.db.
		Where("Fstate = ?", pbcommon.CommonState_CS_VALID).
		Find(&sts).Error; err != nil {
		return nil, err
	}

	var shardings []*pbcommon.Sharding

	for _, st := range sts {
		sharding := &pbcommon.Sharding{
			Key:       st.Key,
			DbId:      st.DBID,
			DbName:    st.DBName,
			Memo:      st.Memo,
			State:     st.State,
			CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		shardings = append(shardings, sharding)
	}

	return shardings, nil
}

// QueryShardingCount returns all shardings count.
func (mgr *ShardingManager) QueryShardingCount() (int64, error) {
	var totalCount int64

	if err := mgr.db.
		Model(&database.Sharding{}).
		Where("Fstate = ?", pbcommon.CommonState_CS_VALID).
		Count(&totalCount).Error; err != nil {
		return 0, err
	}
	return totalCount, nil
}

// UpdateSharding updates target sharding relation.
func (mgr *ShardingManager) UpdateSharding(sharding *pbcommon.Sharding) error {
	ups := map[string]interface{}{
		"DBID":   sharding.DbId,
		"DBName": sharding.DbName,
		"Memo":   sharding.Memo,
		"State":  sharding.State,
	}
	return mgr.db.Model(&database.Sharding{}).Where("Fkey = ?", sharding.Key).Updates(ups).Error
}

// ShowQuestionsStatus returns db status of target variable 'questions'.
func (mgr *ShardingManager) ShowQuestionsStatus() (*DBStatus, error) {
	return mgr.ShowStatus(DBStatusQuestions)
}

// ShowThreadsConnectedStatus returns db status of target variable 'Threads_connected'.
func (mgr *ShardingManager) ShowThreadsConnectedStatus() (*DBStatus, error) {
	return mgr.ShowStatus(DBStatusThreadsConnected)
}

// ShowStatus returns db status of target variable.
func (mgr *ShardingManager) ShowStatus(variableName string) (*DBStatus, error) {
	st := DBStatus{}
	sql := fmt.Sprintf("show global status like '%s';", variableName)

	if err := mgr.db.Raw(sql).Scan(&st).Error; err != nil {
		return nil, err
	}
	return &st, nil
}
