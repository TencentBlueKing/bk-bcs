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

	"bscp.io/cmd/data-service/db-migration/migrator"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20231010143023",
		Name:    "20231010143023_add_credential",
		Mode:    migrator.GormMode,
		Up:      mig20231010143023Up,
		Down:    mig20231010143023Down,
	})
}

// mig20231010143023Up for up migration
func mig20231010143023Up(tx *gorm.DB) error {

	// Credentials ：服务密钥
	type Credentials struct {
		BizID uint   `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_name,priority:1"`
		Name  string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_name,priority:2"`
	}

	// add new column
	if !tx.Migrator().HasColumn(&Credentials{}, "name") {
		if err := tx.Migrator().AddColumn(&Credentials{}, "name"); err != nil {
			return err
		}
	}

	// create new index
	if !tx.Migrator().HasIndex(&Credentials{}, "idx_bizID_name") {
		if err := tx.Migrator().CreateIndex(&Credentials{}, "idx_bizID_name"); err != nil {
			return err
		}
	}

	return nil

}

// mig20231010143023Down for down migration
func mig20231010143023Down(tx *gorm.DB) error {

	// Credentials ：服务密钥
	type Credentials struct {
		BizID uint   `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_name,priority:1"`
		Name  string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_name,priority:2"`
	}

	// delete column
	if tx.Migrator().HasColumn(&Credentials{}, "name") {
		if err := tx.Migrator().DropColumn(&Credentials{}, "name"); err != nil {
			return err
		}
	}

	// delete old index
	if tx.Migrator().HasIndex(&Credentials{}, "idx_bizID_name") {
		if err := tx.Migrator().DropIndex(&Credentials{}, "idx_bizID_name"); err != nil {
			return err
		}
	}

	return nil

}
