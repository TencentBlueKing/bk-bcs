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

// Kvs20240125175500 kv
type Kvs20240125175500 struct {
	Signature string `gorm:"type:varchar(64) not null"`
	ByteSize  uint   `gorm:"type:bigint(1) unsigned not null"`
}

// TableName gorm table name
func (Kvs20240125175500) TableName() string {
	return "kvs"
}

// ReleasedKvs20240125175500 已生成版本的kv
type ReleasedKvs20240125175500 struct {
	Signature string `gorm:"type:varchar(64) not null"`
	ByteSize  uint   `gorm:"type:bigint(1) unsigned not null"`
}

// TableName gorm table name
func (ReleasedKvs20240125175500) TableName() string {
	return "released_kvs"
}

// mig20240125175500Up for up migration
func mig20240125175500Up(tx *gorm.DB) error {
	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&Kvs20240125175500{}); err != nil {
		return err
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&ReleasedKvs20240125175500{}); err != nil {
		return err
	}

	return nil
}

// mig20240125175500Down for down migration
func mig20240125175500Down(tx *gorm.DB) error {
	// delete kvs old column
	if tx.Migrator().HasColumn(&Kvs20240125175500{}, "Signature") {
		if err := tx.Migrator().DropColumn(&Kvs20240125175500{}, "Signature"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&Kvs20240125175500{}, "ByteSize") {
		if err := tx.Migrator().DropColumn(&Kvs20240125175500{}, "ByteSize"); err != nil {
			return err
		}
	}

	// delete release_kvs old column
	if tx.Migrator().HasColumn(&ReleasedKvs20240125175500{}, "Signature") {
		if err := tx.Migrator().DropColumn(&ReleasedKvs20240125175500{}, "Signature"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&ReleasedKvs20240125175500{}, "ByteSize") {
		if err := tx.Migrator().DropColumn(&ReleasedKvs20240125175500{}, "ByteSize"); err != nil {
			return err
		}
	}

	return nil
}
