/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package dao

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao/mysqlrate"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

type driver struct {
	db         *gorm.DB
	rateClient mysqlrate.RateInterface
}

var (
	globalDB *driver
)

func GlobalDB() Interface {
	return globalDB
}

// NewDriver creates the MySQL instance
func NewDriver(dbCfg *common.DBConfig) (Interface, error) {
	connArgs := fmt.Sprintf("%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbCfg.Username, dbCfg.Password, dbCfg.Addr, dbCfg.Database)
	var err error
	globalDB, err = newDriver(connArgs)
	if err != nil {
		return nil, err
	}
	if dbCfg.LimitQPS == 0 {
		dbCfg.LimitQPS = 200
	}
	globalDB.rateClient = mysqlrate.NewRateLimit(globalDB.db, dbCfg.LimitQPS)
	return globalDB, nil
}

func newDriver(connArgs string) (*driver, error) {
	db, err := gorm.Open(mysql.Open(connArgs), &gorm.Config{})
	if err != nil {
		blog.Errorf("Connect to MySQL '%s' failed, err: %s", connArgs, err.Error())
		return nil, err
	}
	return &driver{
		db: db,
	}, nil
}

// Init will auto create the tables if not exist
func (d *driver) Init() error {
	if err := d.autoCreateTable(); err != nil {
		return errors.Wrapf(err, "db driver init failed")
	}
	return nil
}

func (d *driver) autoCreateTable() error {
	if err := d.createTable(tableActivityUser, &ActivityUser{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableActivityUser)
	}
	if err := d.createTable(tableResourcePreference, &ResourcePreference{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableResourcePreference)
	}
	if err := d.createTable(tableSyncInfo, &SyncInfo{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableSyncInfo)
	}
	if err := d.createTable(tableHistoryManifest, &ApplicationHistoryManifest{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableHistoryManifest)
	}
	return nil
}

func (d *driver) createTable(tableName string, obj interface{}) error {
	if d.db.Migrator().HasTable(tableName) {
		blog.Infof("[DB] table '%s' existed.", tableName)
	} else {
		if err := d.db.Table(tableName).Set("gorm:table_options", "ENGINE=InnoDB AUTO_INCREMENT = 0").
			AutoMigrate(obj); err != nil {
			return errors.Wrapf(err, "create table '%s' failed", tableName)
		}
		blog.Infof("[DB] create table '%s' success.", tableName)
	}
	return nil
}

// ListActivityUser return the activity users for project
func (d *driver) ListActivityUser(project string) ([]ActivityUser, error) {
	rows, err := d.rateClient.Table(tableActivityUser).Where("project = ?", project).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "query activity users failed")
	}
	defer rows.Close()

	result := make([]ActivityUser, 0)
	for rows.Next() {
		activityUser := new(ActivityUser)
		if err = d.db.ScanRows(rows, activityUser); err != nil {
			return nil, errors.Wrapf(err, "scan activity user rows failed")
		}
		result = append(result, *activityUser)
	}
	return result, nil
}

// SaveActivityUser save the activity user
func (d *driver) SaveActivityUser(user *ActivityUser) error {
	return d.rateClient.Table(tableActivityUser).Save(user).Error
}

// GetActivityUser get activity user
func (d *driver) GetActivityUser(project, user string) (*ActivityUser, error) {
	rows, err := d.rateClient.Table(tableActivityUser).Where("project = ?", project).
		Where("userName = ?", user).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "get activity user failed")
	}
	defer rows.Close()

	result := make([]*ActivityUser, 0)
	for rows.Next() {
		activityUser := new(ActivityUser)
		if err = d.db.ScanRows(rows, activityUser); err != nil {
			return nil, errors.Wrapf(err, "scan activity user rows failed")
		}
		result = append(result, activityUser)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result[0], nil
}

// UpdateActivityUser update the activity user
func (d *driver) UpdateActivityUser(user *ActivityUser) error {
	if err := d.rateClient.Table(tableActivityUser).Where("id = ?", user.ID).UpdateColumns(map[string]interface{}{
		"operateNum":       user.OperateNum,
		"lastActivityTime": time.Now(),
	}).Error; err != nil {
		return errors.Wrapf(err, "update activity_user '%d' failed", user.ID)
	}
	return nil
}

// ListSyncInfosForProject list the sync infos for project
func (d *driver) ListSyncInfosForProject(project string) ([]SyncInfo, error) {
	rows, err := d.rateClient.Table(tableSyncInfo).Where("project = ?", project).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "query resource preferences failed")
	}
	defer rows.Close()

	result := make([]SyncInfo, 0)
	for rows.Next() {
		syncInfo := new(SyncInfo)
		if err = d.db.ScanRows(rows, syncInfo); err != nil {
			return nil, errors.Wrapf(err, "scan syncinfo rows failed")
		}
		result = append(result, *syncInfo)
	}
	return result, nil
}

// GetSyncInfo get the sync info for project
func (d *driver) GetSyncInfo(project, cluster, app, phase string) (*SyncInfo, error) {
	rows, err := d.rateClient.Table(tableSyncInfo).Where("project = ?", project).
		Where("cluster = ?", cluster).
		Where("application = ?", app).
		Where("phase = ?", phase).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "get syncinfo failed")
	}
	defer rows.Close()

	result := make([]SyncInfo, 0)
	for rows.Next() {
		syncInfo := new(SyncInfo)
		if err = d.db.ScanRows(rows, syncInfo); err != nil {
			return nil, errors.Wrapf(err, "scan syncinfo rows failed")
		}
		result = append(result, *syncInfo)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}

// SaveSyncInfo save the sync info
func (d *driver) SaveSyncInfo(info *SyncInfo) error {
	info.UpdateTime = time.Now()
	return d.rateClient.Table(tableSyncInfo).Save(info).Error
}

// UpdateSyncInfo update the sync info
func (d *driver) UpdateSyncInfo(info *SyncInfo) error {
	if err := d.rateClient.Table(tableSyncInfo).Where("id = ?", info.ID).UpdateColumns(map[string]interface{}{
		"syncTotal":    info.SyncTotal,
		"previousSync": info.PreviousSync,
		"updateTime":   time.Now(),
	}).Error; err != nil {
		return errors.Wrapf(err, "update sync_info failed")
	}
	return nil
}

// SaveResourcePreference save resource preference
func (d *driver) SaveResourcePreference(prefer *ResourcePreference) error {
	return d.db.Table(tableResourcePreference).Save(prefer).Error
}

// DeleteResourcePreference delete the resource preference
func (d *driver) DeleteResourcePreference(project, resourceType, name string) error {
	return d.db.Table(tableResourcePreference).Where("project = ?", project).
		Where("resourceType = ?", resourceType).
		Where("name = ?", name).Delete(&ResourcePreference{}).Error
}

// ListResourcePreferences list all the resource preferences for project
func (d *driver) ListResourcePreferences(project, resourceType string) ([]ResourcePreference, error) {
	rows, err := d.db.Table(tableResourcePreference).Where("project = ?", project).
		Where("resourceType = ?", resourceType).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "query resource preferences failed")
	}
	defer rows.Close()

	result := make([]ResourcePreference, 0)
	for rows.Next() {
		prefer := new(ResourcePreference)
		if err = d.db.ScanRows(rows, prefer); err != nil {
			return nil, errors.Wrapf(err, "scan preference rows failed")
		}
		result = append(result, *prefer)
	}
	return result, nil
}

// CreateApplicationHistoryManifest create application history manifest object
func (d *driver) CreateApplicationHistoryManifest(hm *ApplicationHistoryManifest) error {
	return d.db.Table(tableHistoryManifest).Save(hm).Error
}

// GetApplicationHistoryManifest get application history manifest
func (d *driver) GetApplicationHistoryManifest(appName, appUID string,
	historyID int64) (*ApplicationHistoryManifest, error) {
	rows, err := d.rateClient.Table(tableHistoryManifest).Where("name = ?", appName).
		Where("applicationUID = ?", appUID).
		Where("historyID = ?", historyID).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "get application history manifest failed")
	}
	defer rows.Close()

	result := make([]ApplicationHistoryManifest, 0)
	for rows.Next() {
		appHM := new(ApplicationHistoryManifest)
		if err = d.db.ScanRows(rows, appHM); err != nil {
			return nil, errors.Wrapf(err, "scan syncinfo rows failed")
		}
		result = append(result, *appHM)
	}
	if len(result) == 0 {
		return nil, nil
	}
	return &result[0], nil
}
