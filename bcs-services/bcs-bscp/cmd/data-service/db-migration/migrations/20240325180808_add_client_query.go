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

package migrations

import (
	"time"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrator"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20240325180808",
		Name:    "20240325180808_add_client_query",
		Mode:    migrator.GormMode,
		Up:      mig20240325180808Up,
		Down:    mig20240325180808Down,
	})
}

// mig20240325180808Up for up migration
func mig20240325180808Up(tx *gorm.DB) error {
	// ClientQuerys : ClientQuerys
	type ClientQuerys struct {
		ID uint `gorm:"column:id;type:bigint(1) unsigned;primary_key"`

		BizID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_creator,priority:1"`
		AppID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_creator,priority:2"`

		Creator         string    `gorm:"column:creator;type:varchar(64);NOT NULL;index:idx_bizID_appID_creator,priority:3"`
		SearchName      string    `gorm:"column:search_name;type:varchar(64);default:'';NOT NULL"`
		SearchType      string    `gorm:"column:search_type;type:varchar(20);NOT NULL"`
		SearchCondition string    `gorm:"column:search_condition;type:json;NOT NULL"`
		CreatedAt       time.Time `gorm:"column:created_at;type:datetime(6);NOT NULL"`
		UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime(6);NOT NULL"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&ClientQuerys{}); err != nil {
		return err
	}

	tx.Create([]ClientQuerys{
		{
			ID:              1,
			BizID:           0,
			AppID:           0,
			SearchName:      "配置拉取失败",
			SearchType:      "common",
			SearchCondition: "{\"release_change_status\": [\"failed\"]}",
			Creator:         "system",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              2,
			BizID:           0,
			AppID:           0,
			SearchName:      "离线客户端",
			SearchType:      "common",
			SearchCondition: "{\"online_status\": [\"offline\"]}",
			Creator:         "system",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	})

	now := time.Now()
	if result := tx.Create([]IDGenerators{
		{Resource: "client_querys", MaxID: 2, UpdatedAt: now},
	}); result.Error != nil {
		return result.Error
	}

	return nil
}

// mig20240325180808Down for down migration
func mig20240325180808Down(tx *gorm.DB) error {
	if err := tx.Migrator().DropTable("client_querys"); err != nil {
		return err
	}

	return nil
}
