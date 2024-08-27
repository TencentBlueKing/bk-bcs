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

package storage

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	_ "github.com/go-sql-driver/mysql" // nolint
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-pre-check/pkg/storage/mysqlrate"
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
	if err := d.createTable(tablePreCheckTask, &PreCheckTask{}); err != nil {
		return errors.Wrapf(err, "create table '%s' failed", tablePreCheckTask)
	}
	return nil
}

func (d *driver) createTable(tableName string, obj interface{}) error {
	if d.db.Migrator().HasTable(tableName) {
		blog.Infof("[DB] table '%s' existed.", tableName)
		if !d.db.Table(tableName).Migrator().HasColumn(obj, "labelSelector") {
			if err := d.db.Table(tableName).AutoMigrate(obj); err != nil {
				return errors.Wrapf(err, "update table '%s' failed", tableName)
			}
		}
	} else {
		if err := d.db.Table(tableName).Set("gorm:table_options", "ENGINE=InnoDB AUTO_INCREMENT = 0").
			AutoMigrate(obj); err != nil {
			return errors.Wrapf(err, "create table '%s' failed", tableName)
		}
		blog.Infof("[DB] create table '%s' success.", tableName)
	}
	return nil
}

func (d *driver) CreatePreCheckTask(task *PreCheckTask) (*PreCheckTask, error) {
	if err := d.db.Table(tablePreCheckTask).Create(task).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "create precheck task failed")
	}
	return task, nil
}

func (d *driver) UpdatePreCheckTask(task *PreCheckTask) error {
	if err := d.rateClient.Table(tablePreCheckTask).Save(task).Error; err != nil {
		return errors.Wrapf(err, "update task '%d' failed", task.ID)
	}
	return nil
}

func (d *driver) GetPreCheckTask(id int, project string) (*PreCheckTask, error) {
	result := make([]*PreCheckTask, 0)
	rows, err := d.db.Table(tablePreCheckTask).Where("id = ?", id).Rows()
	if err != nil {
		return nil, errors.Wrapf(err, "get preCheck task failed")
	}
	defer rows.Close() // nolint

	for rows.Next() {
		obj := new(PreCheckTask)
		if err = d.db.ScanRows(rows, obj); err != nil {
			return nil, errors.Wrapf(err, "scan preCheck task failed")
		}
		result = append(result, obj)
	}
	if len(result) == 0 {
		return nil, nil
	}
	if len(result) > 1 {
		blog.Errorf("duplicated id %d", id)
		for _, task := range result {
			if validateProject(task, project) {
				return task, nil
			}
		}
	}
	if validateProject(result[0], project) {
		return result[0], nil
	}
	return nil, nil
}

func (d *driver) ListPreCheckTask(query *PreCheckTaskQuery) ([]*PreCheckTask, error) {
	dbQuery := d.db.Table(tablePreCheckTask).Where("project IN (?) and needReplaceRepo = ?", query.Projects, false).
		Or("project IN (?) and needReplaceRepo = ? and replaceProject = ?", query.Projects, true, "").
		Or("replaceProject IN (?) and needReplaceRepo = ?", query.Projects, true)

	if len(query.Repositories) != 0 {
		dbQuery.Where("repository_addr IN (?)", query.Repositories)
	}
	if query.StartTime != "" {
		dbQuery.Where("create_time >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		dbQuery.Where("update_time <>>= ?", query.EndTime)
	}
	if !query.WithDetail {
		dbQuery.Omit("check_detail")
	}
	tasks := make([]*PreCheckTask, 0)
	if err := dbQuery.Order("id desc").Offset(query.Offset).Limit(query.Limit).Find(&tasks).Error; err != nil {
		return nil, errors.Wrapf(err, "query task failed")
	}
	return tasks, nil
}

func validateProject(task *PreCheckTask, project string) bool {
	if project == "" {
		return true
	}
	if (task.Project == project && !task.NeedReplaceRepo) || (task.ReplaceProject == project && task.NeedReplaceRepo) ||
		(task.Project == project && task.NeedReplaceRepo && task.ReplaceProject == "") {
		return true
	}
	return false
}
