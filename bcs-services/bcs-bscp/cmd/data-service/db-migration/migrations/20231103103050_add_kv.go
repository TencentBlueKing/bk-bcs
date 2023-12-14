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

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrator"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20231103103050",
		Name:    "20231103103050_add_kv",
		Mode:    migrator.GormMode,
		Up:      mig20231103103050Up,
		Down:    mig20231103103050Down,
	})
}

// mig20231103103050Up for up migration
func mig20231103103050Up(tx *gorm.DB) error {

	// Kvs : kv
	type Kvs struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Key     string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_appID_key_kvState,priority:1"`
		Version uint   `gorm:"type:bigint(1) unsigned not null;"`
		KvType  string `gorm:"type:varchar(64) not null"`
		KvState string `gorm:"type:varchar(64) not null;uniqueIndex:idx_bizID_appID_key_kvState,priority:2"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_key_kvState,priority:3"`
		APPID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_key_kvState,priority:4"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&Kvs{}); err != nil {
		return err
	}

	now := time.Now()
	if result := tx.Create([]IDGenerators{
		{Resource: "kvs", MaxID: 0, UpdatedAt: now},
	}); result.Error != nil {
		return result.Error
	}

	return nil
}

// mig20231103103050Down for down migration
func mig20231103103050Down(tx *gorm.DB) error {

	// Kvs : kv
	type Kvs struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Key     string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_appID_key_state,priority:1"`
		Version uint   `gorm:"type:bigint(1) unsigned not null;"`
		KvType  string `gorm:"type:varchar(64) not null"`
		KvState string `gorm:"type:varchar(64) not null;uniqueIndex:idx_bizID_appID_key_state,priority:2"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_key_state,priority:3"`
		APPID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_key_state,priority:4"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Migrator().DropTable(Kvs{}); err != nil {
		return err
	}

	var resources = []string{
		"kvs",
	}
	if result := tx.Where("resource IN ?", resources).Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	return nil
}
