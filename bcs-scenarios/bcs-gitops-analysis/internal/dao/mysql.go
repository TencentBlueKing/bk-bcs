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
 */

package dao

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/internal/dao/mysqlrate"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/options"
)

type driver struct {
	db         *gorm.DB
	rateClient mysqlrate.RateInterface
}

var (
	globalDB *driver
)

// GlobalDB global db
func GlobalDB() Interface {
	return globalDB
}

// NewDriver creates the MySQL instance
func NewDriver() (Interface, error) {
	dbCfg := options.GlobalOptions().DBConfig
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
	if err := d.createTable(tableSyncInfo, &SyncInfo{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableSyncInfo)
	}
	if err := d.createTable(tableResourceInfo, &ResourceInfo{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tableResourceInfo)
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
	rows, err := d.rateClient.Table(tableActivityUser).Where("project = ?", project).
		Order("lastActivityTime DESC").Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "query activity users failed")
	}
	defer rows.Close() // nolint

	result := make([]ActivityUser, 0)
	appeared := make(map[string]struct{})
	for rows.Next() {
		activityUser := new(ActivityUser)
		if err = d.db.ScanRows(rows, activityUser); err != nil {
			return nil, errors.Wrapf(err, "scan activity user rows failed")
		}
		// 防止因脏数据导致的数据不一致
		if _, ok := appeared[activityUser.Project+activityUser.UserName]; ok {
			continue
		}
		appeared[activityUser.Project+activityUser.UserName] = struct{}{}
		result = append(result, *activityUser)
	}
	return result, nil
}

// List7DayActivityUsers return last 7day activity user
func (d *driver) List7DayActivityUsers() ([]ActivityUser, error) {
	t := time.Now()
	v := t.Add(-7 * 24 * time.Hour)
	rows, err := d.rateClient.Table(tableActivityUser).Where("lastActivityTime > ?", v).
		Order("lastActivityTime ASC").Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "query 7day activity users failed")
	}
	defer rows.Close()

	result := make([]ActivityUser, 0)
	appeared := make(map[string]struct{})
	for rows.Next() {
		activityUser := new(ActivityUser)
		if err = d.db.ScanRows(rows, activityUser); err != nil {
			return nil, errors.Wrapf(err, "scan activity user rows failed")
		}
		// 防止因脏数据导致的数据不一致
		if _, ok := appeared[activityUser.Project+"/"+activityUser.UserName]; ok {
			continue
		}
		appeared[activityUser.Project+"/"+activityUser.UserName] = struct{}{}
		result = append(result, *activityUser)
	}
	return result, nil
}

// ListSyncInfosForProject list the sync infos for project
func (d *driver) ListSyncInfosForProject(project string) ([]SyncInfo, error) {
	var rows *sql.Rows
	var err error
	if project != "" {
		rows, err = d.rateClient.Table(tableSyncInfo).Where("project = ?", project).Rows()
	} else {
		rows, err = d.rateClient.Table(tableSyncInfo).Rows()
	}
	if err != nil {
		return nil, errors.Wrapf(err, "query resource preferences failed")
	}
	defer rows.Close() // nolint

	syncs := make(map[string]*SyncInfo)
	for rows.Next() {
		syncInfo := new(SyncInfo)
		if err = d.db.ScanRows(rows, syncInfo); err != nil {
			return nil, errors.Wrapf(err, "scan syncinfo rows failed")
		}
		oldSyncInfo, ok := syncs[syncInfo.Application]
		if !ok {
			syncs[syncInfo.Application] = syncInfo
			continue
		}
		// 应用可能出现被删除后重建同名应用的情况，将应用的同步合并
		if oldSyncInfo.UpdateTime.Before(syncInfo.UpdateTime) {
			syncInfo.SyncTotal += oldSyncInfo.SyncTotal
			syncs[syncInfo.Application] = syncInfo
		} else {
			oldSyncInfo.SyncTotal += syncInfo.SyncTotal
		}
	}

	result := make([]SyncInfo, 0)
	for _, syncInfo := range syncs {
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
	defer rows.Close() // nolint

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

// SaveOrUpdateResourceInfo save the resource info object
func (d *driver) SaveOrUpdateResourceInfo(info *ResourceInfo) error {
	info.UpdateTime = time.Now()
	result := d.rateClient.Table(tableResourceInfo).
		Where("project = ? AND application = ?", info.Project, info.Application).
		UpdateColumns(map[string]interface{}{
			"resources":  info.Resources,
			"updateTime": time.Now(),
		})
	if err := result.Error; err != nil {
		return errors.Wrapf(err, "update resource info '%s' failed", info.Application)
	}
	if result.RowsAffected == 0 {
		return d.rateClient.Table(tableResourceInfo).Save(info).Error
	}
	return nil
}

// ListResourceInfosByProject list resource info by project
func (d *driver) ListResourceInfosByProject(projects []string) ([]ResourceInfo, error) {
	var rows *sql.Rows
	var err error
	if len(projects) == 0 {
		rows, err = d.rateClient.Table(tableResourceInfo).Rows()
	} else {
		rows, err = d.rateClient.Table(tableResourceInfo).Where("project IN (?)", projects).Rows()
	}
	if err != nil {
		return nil, errors.Wrapf(err, "query resource preferences failed")
	}
	defer rows.Close() // nolint

	result := make([]ResourceInfo, 0)
	for rows.Next() {
		syncInfo := new(ResourceInfo)
		if err = d.db.ScanRows(rows, syncInfo); err != nil {
			return nil, errors.Wrapf(err, "scan syncinfo rows failed")
		}
		result = append(result, *syncInfo)
	}
	return result, nil
}

// GetResourceInfo get the resource info by project and application
func (d *driver) GetResourceInfo(project, app string) (*ResourceInfo, error) {
	rows, err := d.rateClient.Table(tableResourceInfo).Where("project = ?", project).
		Where("application = ?", app).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "get syncinfo failed")
	}
	defer rows.Close() // nolint

	result := make([]ResourceInfo, 0)
	for rows.Next() {
		syncInfo := new(ResourceInfo)
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
