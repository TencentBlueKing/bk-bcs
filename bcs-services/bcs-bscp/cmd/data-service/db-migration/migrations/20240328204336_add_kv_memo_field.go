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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20240328204336",
		Name:    "20240328204336_add_kv_memo_field",
		Mode:    migrator.GormMode,
		Up:      mig20240328204336Up,
		Down:    mig20240328204336Down,
	})
}

// Kvs20240328204336 kv
type Kvs20240328204336 struct {
	ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

	// Spec is specifics of the resource defined with user
	Memo string `gorm:"type:varchar(256) default ''"`
}

// TableName gorm table name
func (Kvs20240328204336) TableName() string {
	t := &table.Kv{}
	return t.TableName()
}

// ReleasedKvs20240328204336 已生成版本的kv
type ReleasedKvs20240328204336 struct {
	ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

	// Spec is specifics of the resource defined with user
	Memo string `gorm:"type:varchar(256) default ''"`
}

// TableName gorm table name
func (ReleasedKvs20240328204336) TableName() string {
	t := &table.ReleasedKv{}
	return t.TableName()
}

// mig20240328204336Up for up migration
func mig20240328204336Up(tx *gorm.DB) error {

	if !tx.Migrator().HasColumn(&Kvs20240328204336{}, "Memo") {
		if err := tx.Migrator().AddColumn(&Kvs20240328204336{}, "Memo"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&ReleasedKvs20240328204336{}, "Memo") {
		if err := tx.Migrator().AddColumn(&ReleasedKvs20240328204336{}, "Memo"); err != nil {
			return err
		}
	}

	return nil
}

// mig20240328204336Down for down migration
func mig20240328204336Down(tx *gorm.DB) error {

	if tx.Migrator().HasColumn(&Kvs20240328204336{}, "Memo") {
		if err := tx.Migrator().DropColumn(&Kvs20240328204336{}, "Memo"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&ReleasedKvs20240328204336{}, "Memo") {
		if err := tx.Migrator().DropColumn(&ReleasedKvs20240328204336{}, "Memo"); err != nil {
			return err
		}
	}

	return nil
}
