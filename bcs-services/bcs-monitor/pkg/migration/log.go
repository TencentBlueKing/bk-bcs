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

// Package migration log rule
package migration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

var (
	// mysql
	mysqlHost     = os.Getenv("MYSQL_HOST")
	mysqlPort     = os.Getenv("MYSQL_PORT")
	mysqlUsername = os.Getenv("MYSQL_USERNAME")
	mysqlPassword = os.Getenv("MYSQL_PASSWORD")
	mysqlDatabase = os.Getenv("MYSQL_DATABASE")
)

// MigrateLogRule migrate log rule from saas db
// 数据迁移分为两个部分，1.28 saas 对接 bklog 的规则数据和旧版本日志的索引集
func MigrateLogRule() error {
	blog.Info("start to migrate log rule")

	// init mysql
	mysqlDB, err := initDB()
	if err != nil {
		blog.Errorf("init mysql db failed, err %s", err.Error())
		return err
	}

	ctx := context.Background()

	migrateLogRule(ctx, storage.GlobalStorage, mysqlDB)
	migrateLogIndexSet(ctx, storage.GlobalStorage, mysqlDB)
	return nil
}

func migrateLogRule(ctx context.Context, model storage.Storage, mysqlDB *gorm.DB) {
	// load old log rule
	var err error
	rules := loadOldLogRules(mysqlDB)
	cond := operator.NewLeafCondition(operator.Gte, operator.M{
		entity.FieldKeyRuleID: 0,
	})
	_, newRules, err := model.ListLogRules(ctx, cond, &utils.ListOption{})
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Fatalf("list rules failed, err %s", err.Error())
		return
	}

	// save log rules
	count := 0
	for _, v := range rules {
		// create if not exist
		exist := false
		for _, r := range newRules {
			if v.RuleID == r.RuleID {
				exist = true
				break
			}
		}
		if exist {
			continue
		}
		_, err = model.CreateLogRule(ctx, &entity.LogRule{
			Name:      v.RuleName,
			RuleName:  v.RuleName,
			RuleID:    v.RuleID,
			ProjectID: v.ProjectID,
			ClusterID: v.ClusterID,
			CreatedAt: utils.JSONTime{Time: v.CreateTime},
			UpdatedAt: utils.JSONTime{Time: v.UpdateTime},
			Creator:   v.Creator,
			Updator:   v.Updator,
			Status:    entity.SuccessStatus,
		})
		if err != nil {
			blog.Fatalf("create rules failed, err %s", err.Error())
		}
		count++
	}
	blog.Infof("migrate %d rules from saas db", count)
}

func migrateLogIndexSet(ctx context.Context, model storage.Storage, mysqlDB *gorm.DB) {
	// load log index sets
	indexs := loadOldIndexSet(mysqlDB)

	// save log index sets
	count := 0
	for i := range indexs {
		oldIndex, err := model.GetOldIndexSetID(ctx, indexs[i].ProjectID)
		exist := true
		if err != nil {
			blog.Fatalf("get old index failed, err %s", err.Error())
		}
		if oldIndex == nil {
			exist = false
		}
		// create if not exist
		if !exist {
			err = model.CreateOldIndexSetID(ctx, &indexs[i])
			if err != nil {
				blog.Fatalf("create index failed, err %s", err.Error())
			}
		}
		count++
	}
	blog.Infof("migrate %d index sets from saas db", count)
}

func loadOldLogRules(db *gorm.DB) []logCollectorMetadata {
	var rules []logCollectorMetadata
	err := db.Raw("SELECT creator, updator, created, updated, project_id, cluster_id, config_id, " +
		"config_name FROM log_collect_logcollectmetadata WHERE is_deleted=0").
		Scan(&rules).Error
	if err != nil {
		blog.Fatalf("get old rules failed, err %s", err.Error())
	}
	blog.Infof("get %d rules from saas db", len(rules))
	return rules
}

func loadOldIndexSet(db *gorm.DB) []entity.LogIndex {
	if !db.Migrator().HasTable(&entity.LogIndex{}) {
		blog.Info("not found log index table")
		return nil
	}
	var indexs []entity.LogIndex
	err := db.Raw("SELECT project_id, std_data_id, file_data_id, file_index_set_id, std_index_set_id, cc_app_id " +
		"FROM datalog_datalogplan").
		Scan(&indexs).Error
	if err != nil {
		blog.Fatalf("get indexs failed, err %s", err.Error())
	}
	blog.Infof("get %d indexs from saas db", len(indexs))
	return indexs
}

func initDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		mysqlUsername, mysqlPassword, mysqlHost, mustInt(mysqlPort), mysqlDatabase)
	mysqlDB, err := gorm.Open(mysql.Open(dsn))
	if err != nil {
		return nil, err
	}
	return mysqlDB, nil
}

func mustInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

// 新版日志 saas 实现产生的数据
type logCollectorMetadata struct {
	Creator    string    `json:"creator"`
	Updator    string    `json:"updator"`
	CreateTime time.Time `json:"createTime" gorm:"column:created"`
	UpdateTime time.Time `json:"updateTime" gorm:"column:updated"`
	ProjectID  string    `json:"projectID" gorm:"column:project_id"`
	ClusterID  string    `json:"clusterID" gorm:"column:cluster_id"`
	RuleID     int       `json:"ruleID" gorm:"column:config_id"`
	RuleName   string    `json:"ruleName" gorm:"column:config_name"`
}
