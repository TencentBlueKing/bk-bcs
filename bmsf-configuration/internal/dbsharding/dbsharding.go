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
	"github.com/jinzhu/gorm"

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

// enableDBLog is database inner log flag.
var enableDBLog = false

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

var (
	// tbMap is a map that stored tables base informations
	// for auto migration, dbid:dbname:tablename -> EXISTSFLAG.
	tbMap = make(map[string]bool)

	// tbMapMu makes the ops on tbMap safe.
	tbMapMu = sync.RWMutex{}
)

// ShardingDB is database sharding result.
type ShardingDB struct {
	// DB service id.
	DBid string

	// database name.
	DBName string

	// gorm database handler.
	db *gorm.DB
}

// DB returns DB handler.
func (sd *ShardingDB) DB() *gorm.DB {
	return sd.db
}

func (sd *ShardingDB) tbmapKey(tbname string) string {
	return sd.DBid + ":" + sd.DBName + ":" + tbname
}

// AutoMigrate run auto migration for given models.
func (sd *ShardingDB) AutoMigrate(tb database.Table) {
	tbMapMu.RLock()
	_, ok := tbMap[sd.tbmapKey(tb.TableName())]
	tbMapMu.RUnlock()

	if ok {
		return
	}

	if sd.db.HasTable(tb) {
		tbMapMu.Lock()
		tbMap[sd.tbmapKey(tb.TableName())] = true
		tbMapMu.Unlock()
	} else {
		// only create table, do not alter table, ignore the error
		// when there is a repeated db operation。
		sd.db.Set("gorm:table_options", "CHARSET="+database.BSCPCHARSET).CreateTable(tb)
	}
}

// Config of ShardingMgr.
type Config struct {
	// Dialect is database driver name.
	Dialect string

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
	// initializes system database.
	db, err := gorm.Open(mgr.config.Dialect,
		fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s&charset=%s",
			mgr.config.DBUser,
			mgr.config.DBPasswd,
			mgr.config.DBHost,
			mgr.config.DBPort,
			mgr.configTemplate.ConnTimeout,
			mgr.configTemplate.ReadTimeout,
			mgr.configTemplate.WriteTimeout,
			database.BSCPCHARSET,
		))
	if err != nil {
		return err
	}
	if err = db.Exec("CREATE DATABASE IF NOT EXISTS " + database.BSCPSHARDINGDB).Error; err != nil {
		return err
	}
	if err = db.Close(); err != nil {
		return err
	}

	if db, err = gorm.Open(mgr.config.Dialect,
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s&charset=%s",
			mgr.config.DBUser,
			mgr.config.DBPasswd,
			mgr.config.DBHost,
			mgr.config.DBPort,
			database.BSCPSHARDINGDB,
			mgr.configTemplate.ConnTimeout,
			mgr.configTemplate.ReadTimeout,
			mgr.configTemplate.WriteTimeout,
			database.BSCPCHARSET,
		)); err != nil {
		return err
	}

	db.DB().SetMaxOpenConns(mgr.configTemplate.MaxOpenConns)
	db.DB().SetMaxIdleConns(mgr.configTemplate.MaxIdleConns)
	db.DB().SetConnMaxLifetime(mgr.configTemplate.KeepAlive)

	db.LogMode(enableDBLog)
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
	db, err := gorm.Open(mgr.config.Dialect,
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s&charset=%s",
			service.User,
			service.Password,
			service.Host,
			service.Port,
			dbname,
			mgr.configTemplate.ConnTimeout,
			mgr.configTemplate.ReadTimeout,
			mgr.configTemplate.WriteTimeout,
			database.BSCPCHARSET,
		))
	if err != nil {
		return nil, err
	}

	db.DB().SetMaxOpenConns(mgr.configTemplate.MaxOpenConns)
	db.DB().SetMaxIdleConns(mgr.configTemplate.MaxIdleConns)
	db.DB().SetConnMaxLifetime(mgr.configTemplate.KeepAlive)

	db.LogMode(enableDBLog)

	return db, nil
}

func (mgr *ShardingManager) evicteDB(k, v interface{}) {
	db, ok := v.(*gorm.DB)
	if !ok {
		return
	}

	if db != nil {
		db.Close()
	}
}

func (mgr *ShardingManager) getSharding(key string) (*ShardingDB, error) {
	var target *ShardingDB

	sd, err := mgr.shardings.Get(key)
	if err != nil || sd == nil {
		if !mgr.db.HasTable(&database.Sharding{}) {
			mgr.db.Set("gorm:table_options", "CHARSET="+database.BSCPCHARSET).CreateTable(&database.Sharding{})
		}

		var stsd database.Sharding
		if err := mgr.db.Where("Fkey = ?", key).First(&stsd).Error; err != nil {
			return nil, err
		}
		target = &ShardingDB{DBid: stsd.DBid, DBName: stsd.DBName}

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

func (mgr *ShardingManager) getDBService(dbid string) (*DBService, error) {
	var target *DBService

	service, err := mgr.dbServices.Get(dbid)
	if err != nil || service == nil {
		if !mgr.db.HasTable(&database.ShardingDB{}) {
			mgr.db.Set("gorm:table_options", "CHARSET="+database.BSCPCHARSET).CreateTable(&database.ShardingDB{})
		}

		var stdb database.ShardingDB
		if err := mgr.db.Where("Fdbid = ? AND Fstate = ?", dbid, pbcommon.CommonState_CS_VALID).
			First(&stdb).Error; err != nil {
			return nil, err
		}

		target = &DBService{
			ID:       stdb.DBid,
			Host:     stdb.Host,
			Port:     int(stdb.Port),
			User:     stdb.User,
			Password: stdb.Password,
		}

		mgr.dbServices.Set(stdb.DBid, target)
	} else {
		v, ok := service.(*DBService)
		if !ok || v == nil {
			return nil, errors.New("can't get sharding database, invalid dbServices cache struct")
		}
		target = v
	}

	return target, nil
}

func (mgr *ShardingManager) dbSDKey(dbid, dbname string) string {
	return dbid + "-" + dbname
}

func (mgr *ShardingManager) getDB(dbid, dbname string) (*gorm.DB, error) {
	var target *gorm.DB

	db, err := mgr.dbs.Get(mgr.dbSDKey(dbid, dbname))
	if err != nil || db == nil {
		service, err := mgr.getDBService(dbid)
		if err != nil {
			return nil, err
		}

		// make a new db connection with db service instance.
		newDB, err := mgr.newDB(service, dbname)
		if err != nil {
			return nil, err
		}

		// update new database client without repetition.
		mgr.repeatMu.Lock()
		defer mgr.repeatMu.Unlock()

		odb, err := mgr.dbs.Get(mgr.dbSDKey(dbid, dbname))
		if err != nil {
			target = newDB
			mgr.dbs.Set(mgr.dbSDKey(dbid, dbname), target)
		} else {
			v, ok := odb.(*gorm.DB)
			if !ok || v == nil {
				return nil, errors.New("can't get sharding database, invalid dbs cache struct")
			}
			target = v
			newDB.Close()
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
		return &ShardingDB{DBName: database.BSCPSHARDINGDB, db: mgr.db}, nil
	}

	// get ShardingDB from shardings cache.
	sd, err := mgr.getSharding(key)
	if err != nil {
		return nil, err
	}

	// target db service instance client.
	db, err := mgr.getDB(sd.DBid, sd.DBName)
	if err != nil {
		return nil, err
	}

	// return target sharding result, include dbname and db handler.
	shardingDB := &ShardingDB{
		DBid:   sd.DBid,
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
	if err := mgr.db.Close(); err != nil {
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
		db.Close()
	}

	return nil
}

// CreateShardingDB create a new sharding database.
func (mgr *ShardingManager) CreateShardingDB(db *pbcommon.ShardingDB) error {
	if !mgr.db.HasTable(&database.ShardingDB{}) {
		mgr.db.Set("gorm:table_options", "CHARSET="+database.BSCPCHARSET).CreateTable(&database.ShardingDB{})
	}

	st := &database.ShardingDB{
		DBid:     db.Dbid,
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
func (mgr *ShardingManager) QueryShardingDB(dbid string) (*pbcommon.ShardingDB, error) {
	if !mgr.db.HasTable(&database.ShardingDB{}) {
		mgr.db.Set("gorm:table_options", "CHARSET="+database.BSCPCHARSET).CreateTable(&database.ShardingDB{})
	}

	var st database.ShardingDB
	if err := mgr.db.Where("Fdbid = ?", dbid).First(&st).Error; err != nil {
		return nil, err
	}

	db := &pbcommon.ShardingDB{
		Dbid:      st.DBid,
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
	if !mgr.db.HasTable(&database.ShardingDB{}) {
		mgr.db.Set("gorm:table_options", "CHARSET="+database.BSCPCHARSET).CreateTable(&database.ShardingDB{})
	}

	var sts []database.ShardingDB
	if err := mgr.db.Where("Fstate = ?", pbcommon.CommonState_CS_VALID).Find(&sts).Error; err != nil {
		return nil, err
	}

	var dbs []*pbcommon.ShardingDB
	for _, st := range sts {
		db := &pbcommon.ShardingDB{
			Dbid:      st.DBid,
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

// UpdateShardingDB updates target sharding database.
func (mgr *ShardingManager) UpdateShardingDB(db *pbcommon.ShardingDB) error {
	if !mgr.db.HasTable(&database.ShardingDB{}) {
		mgr.db.Set("gorm:table_options", "CHARSET="+database.BSCPCHARSET).CreateTable(&database.ShardingDB{})
	}

	ups := map[string]interface{}{
		"Host":     db.Host,
		"Port":     db.Port,
		"User":     db.User,
		"Password": db.Password,
		"Memo":     db.Memo,
		"State":    db.State,
	}
	return mgr.db.Model(&database.ShardingDB{}).Where("Fdbid = ?", db.Dbid).Updates(ups).Error
}

// CreateSharding creates a new sharding relation.
func (mgr *ShardingManager) CreateSharding(sharding *pbcommon.Sharding) error {
	if !mgr.db.HasTable(&database.Sharding{}) {
		mgr.db.Set("gorm:table_options", "CHARSET="+database.BSCPCHARSET).CreateTable(&database.Sharding{})
	}

	st := &database.Sharding{
		Key:    sharding.Key,
		DBid:   sharding.Dbid,
		DBName: sharding.Dbname,
		Memo:   sharding.Memo,
		State:  sharding.State,
	}
	return mgr.db.Create(st).Error
}

// QuerySharding returns target sharding by key.
func (mgr *ShardingManager) QuerySharding(key string) (*pbcommon.Sharding, error) {
	if !mgr.db.HasTable(&database.Sharding{}) {
		mgr.db.Set("gorm:table_options", "CHARSET="+database.BSCPCHARSET).CreateTable(&database.Sharding{})
	}

	var st database.Sharding
	if err := mgr.db.Where("Fkey = ?", key).First(&st).Error; err != nil {
		return nil, err
	}

	sharding := &pbcommon.Sharding{
		Key:       st.Key,
		Dbid:      st.DBid,
		Dbname:    st.DBName,
		Memo:      st.Memo,
		State:     st.State,
		CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	return sharding, nil
}

// UpdateSharding updates target sharding relation.
func (mgr *ShardingManager) UpdateSharding(sharding *pbcommon.Sharding) error {
	if !mgr.db.HasTable(&database.Sharding{}) {
		mgr.db.Set("gorm:table_options", "CHARSET="+database.BSCPCHARSET).CreateTable(&database.Sharding{})
	}

	ups := map[string]interface{}{
		"DBid":   sharding.Dbid,
		"DBName": sharding.Dbname,
		"Memo":   sharding.Memo,
		"State":  sharding.State,
	}
	return mgr.db.Model(&database.Sharding{}).Where("Fkey = ?", sharding.Key).Updates(ups).Error
}
