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
		Version: "20240125175500",
		Name:    "20240125175500_kv_add_signature",
		Mode:    migrator.GormMode,
		Up:      mig20240125175500Up,
		Down:    mig20240125175500Down,
	})
}

// mig20240125175500Up for up migration
func mig20240125175500Up(tx *gorm.DB) error {
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

		Signature string `gorm:"type:varchar(64) not null"`
		ByteSize  uint   `gorm:"type:bigint(1) unsigned not null"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// ReleasedKvs :已生成版本的kv
	type ReleasedKvs struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Key       string `gorm:"type:varchar(255) not null;uniqueIndex:relID_key,priority:1"`
		Version   uint   `gorm:"type:bigint(1) unsigned not null;"`
		ReleaseID uint   `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_ID,priority:3;uniqueIndex:relID_key,priority:2"` //nolint:lll
		KvType    string `gorm:"type:varchar(64) not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_ID,priority:1"`
		AppID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_ID,priority:2"`

		Signature string `gorm:"type:varchar(64) not null"`
		ByteSize  uint   `gorm:"type:bigint(1) unsigned not null"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&Kvs{}); err != nil {
		return err
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&ReleasedKvs{}); err != nil {
		return err
	}

	return nil

}

// mig20240125175500Down for down migration
func mig20240125175500Down(tx *gorm.DB) error {
	// Kvs: 服务
	type Kvs struct {
		Signature string `gorm:"type:varchar(64) not null"`
		ByteSize  uint   `gorm:"type:bigint(1) unsigned not null"`
	}

	// ReleasedKvs: 服务
	type ReleasedKvs struct {
		Signature string `gorm:"type:varchar(64) not null"`
		ByteSize  uint   `gorm:"type:bigint(1) unsigned not null"`
	}

	// delete old column
	if tx.Migrator().HasColumn(&Kvs{}, "signature") {
		if err := tx.Migrator().DropColumn(&Kvs{}, "signature"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&Kvs{}, "byte_size") {
		if err := tx.Migrator().DropColumn(&Kvs{}, "byte_size"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&ReleasedKvs{}, "signature") {
		if err := tx.Migrator().DropColumn(&ReleasedKvs{}, "signature"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&ReleasedKvs{}, "byte_size") {
		if err := tx.Migrator().DropColumn(&ReleasedKvs{}, "byte_size"); err != nil {
			return err
		}
	}

	return nil
}
