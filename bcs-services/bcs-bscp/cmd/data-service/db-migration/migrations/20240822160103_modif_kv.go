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
		Version: "20240822160103",
		Name:    "20240822160103_modif_kv",
		Mode:    migrator.GormMode,
		Up:      mig20240822160103Up,
		Down:    mig20240822160103Down,
	})
}

// mig20240822160103Up for up migration
func mig20240822160103Up(tx *gorm.DB) error {

	// Kvs  : KV
	type Kvs struct {
		SecretType   string `gorm:"column:secret_type;type:varchar(20);default:'';NOT NULL"`
		SecretHidden uint   `gorm:"column:secret_hidden;type:tinyint(1) unsigned;default:0;NOT NULL"`
	}

	// ReleasedKvs  : 发布的KV
	type ReleasedKvs struct {
		SecretType   string `gorm:"column:secret_type;type:varchar(20);default:'';NOT NULL"`
		SecretHidden uint   `gorm:"column:secret_hidden;type:tinyint(1) unsigned;default:0;NOT NULL"`
	}

	// Kvs add new column
	if !tx.Migrator().HasColumn(&Kvs{}, "secret_type") {
		if err := tx.Migrator().AddColumn(&Kvs{}, "secret_type"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&Kvs{}, "secret_hidden") {
		if err := tx.Migrator().AddColumn(&Kvs{}, "secret_hidden"); err != nil {
			return err
		}
	}

	// ReleasedKvs add new column
	if !tx.Migrator().HasColumn(&ReleasedKvs{}, "secret_type") {
		if err := tx.Migrator().AddColumn(&ReleasedKvs{}, "secret_type"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&ReleasedKvs{}, "secret_hidden") {
		if err := tx.Migrator().AddColumn(&ReleasedKvs{}, "secret_hidden"); err != nil {
			return err
		}
	}

	return nil
}

// mig20240822160103Down for down migration
func mig20240822160103Down(tx *gorm.DB) error {

	// Kvs  : KV
	type Kvs struct {
		SecretType   string `gorm:"column:secret_type;type:varchar(20);default:'';NOT NULL"`
		SecretHidden uint   `gorm:"column:secret_hidden;type:tinyint(1) unsigned;default:0;NOT NULL"`
	}

	// ReleasedKvs  : 发布的KV
	type ReleasedKvs struct {
		SecretType   string `gorm:"column:secret_type;type:varchar(20);default:'';NOT NULL"`
		SecretHidden uint   `gorm:"column:secret_hidden;type:tinyint(1) unsigned;default:0;NOT NULL"`
	}

	// Kvs drop column
	if tx.Migrator().HasColumn(&Kvs{}, "secret_type") {
		if err := tx.Migrator().DropColumn(&Kvs{}, "secret_type"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&Kvs{}, "secret_hidden") {
		if err := tx.Migrator().DropColumn(&Kvs{}, "secret_hidden"); err != nil {
			return err
		}
	}

	// ReleasedKvs drop column
	if tx.Migrator().HasColumn(&ReleasedKvs{}, "secret_type") {
		if err := tx.Migrator().DropColumn(&ReleasedKvs{}, "secret_type"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&ReleasedKvs{}, "secret_hidden") {
		if err := tx.Migrator().DropColumn(&ReleasedKvs{}, "secret_hidden"); err != nil {
			return err
		}
	}

	return nil
}
