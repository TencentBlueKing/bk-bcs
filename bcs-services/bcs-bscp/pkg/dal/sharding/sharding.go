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

// Package sharding NOTES
package sharding

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/sharding"
	"k8s.io/klog/v2"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
)

// InitSharding initialize a sharding management instance.
func InitSharding(sd *cc.Sharding) (*Sharding, error) {

	db, err := connect(sd.AdminDatabase)
	if err != nil {
		return nil, err
	}

	s := &Sharding{
		one: &One{
			shardingUid: uuid.UUID(),
			db:          db,
		},
	}

	return s, nil
}

// Sharding is used to manage all the mysql instances
// which works for all the biz and admin resources.
type Sharding struct {
	// we support only one db just for now.
	// Note: support sharding management later.
	one *One
	// db pool, like a connection pool
}

// MustSharding get a db instance with biz id.
// It does not check the biz's value, caller should to
// guarantee that biz is > 0; Otherwise, it will panic.
func (s *Sharding) MustSharding(biz uint32) *sqlx.DB {
	return s.one.db
}

// ShardingOne get a db instance with biz id.
func (s *Sharding) ShardingOne(biz uint32) *One {
	if biz <= 0 {
		return &One{hitErr: fmt.Errorf("invalid sharding one, because biz: %d is invalid", biz)}
	}

	return s.one
}

// Admin get the admin db instance
func (s *Sharding) Admin() *One {
	return s.one
}

// Audit get the audit db instance
func (s *Sharding) Audit() *One {
	return s.one
}

// Event get the event db instance
func (s *Sharding) Event() *One {
	return s.one
}

// Healthz check mysql healthz.
func (s *Sharding) Healthz() error {
	if err := s.one.db.Ping(); err != nil {
		return errors.New("mysql ping failed, err: " + err.Error())
	}

	return nil
}

// checkAuditorDB check db connection
func checkAuditorDB(auditorDB *gorm.DB) error {
	db, err := auditorDB.DB()
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	// 初始化 Gen 配置
	genM := gen.Use(auditorDB)
	q := genM.Audit.WithContext(context.Background())

	if _, err := q.Limit(1).Find(); err != nil {
		return err
	}

	return nil
}

// MustShardingAuditor auditorDB 不存在，使用 adminDB
func MustShardingAuditor(adminDB *gorm.DB) *gorm.DB {
	auditorDB, err := InitAuditorSharding(adminDB)
	if err != nil {
		klog.InfoS("init auditor sharding failed, fallback to admin db", "dbname", adminDB.Migrator().CurrentDatabase(),
			"err", err)
		return adminDB
	}

	klog.InfoS("init auditor sharding done", "dbname", auditorDB.Migrator().CurrentDatabase())
	return auditorDB
}

// InitAuditorSharding 审计表分库
func InitAuditorSharding(adminDB *gorm.DB) (*gorm.DB, error) {
	// 初始化配置
	conf := adminDB.Dialector.(*mysql.Dialector).Config.DSNConfig.Clone()
	conf.DBName = "bk_bscp_auditor"

	auditorDB, err := gorm.Open(mysql.Open(conf.FormatDSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}

	if err := checkAuditorDB(auditorDB); err != nil {
		return nil, err
	}

	if err := auditorDB.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		return nil, err
	}

	return auditorDB, nil
}

// bizPrimaryKeyGeneratorFn biz 主键生成算法
func bizPrimaryKeyGeneratorFn(index int64) int64 {
	return 0
}

// bizShardingAlgorithm biz切表算法
func bizShardingAlgorithm(value interface{}) (suffix string, err error) {
	id := 0
	switch value := value.(type) {
	case int:
		id = value
	case int64:
		id = int(value)
	case uint32:
		id = int(value)
	default:
		return "", fmt.Errorf("not valid biz type")
	}

	// 特定业务才切换表
	if id == 0 {
		return fmt.Sprintf("_%d", id), nil
	}

	return "", nil
}

// InitBizSharding 按业务ID分表
func InitBizSharding(db *gorm.DB) error {
	// 初始化 Gen 配置
	genM := gen.Use(db)

	// 使用 biz_id 分表
	sh := sharding.Register(
		sharding.Config{
			ShardingKey:           genM.TemplateSpace.BizID.ColumnName().String(),
			PrimaryKeyGenerator:   sharding.PKCustom,
			ShardingAlgorithm:     bizShardingAlgorithm,
			PrimaryKeyGeneratorFn: bizPrimaryKeyGeneratorFn,
		},
		genM.TemplateSpace.TableName(),
	)

	if err := db.Use(sh); err != nil {
		return errors.Wrap(err, "init biz sharding")
	}
	return nil
}
