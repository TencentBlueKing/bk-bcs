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
	"fmt"

	"gorm.io/gorm"

	"bscp.io/cmd/data-service/db-migration/migrator"
	"bscp.io/pkg/dal/table"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20231113103022",
		Name:    "20231113103022_modify_app",
		Mode:    migrator.GormMode,
		Up:      mig20231113103022Up,
		Down:    mig20231113103022Down,
	})
}

// mig20231113103022Up for up migration
func mig20231113103022Up(tx *gorm.DB) error {

	// Applications:
	type Applications struct {
		Alias        string `gorm:"type:varchar(255) not null;"`
		CredentialID uint   `gorm:"type:bigint(1) unsigned not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").AutoMigrate(
		&Applications{},
	); err != nil {
		return err
	}

	// set default value
	var apps []table.App
	tx.Model(&table.App{}).Find(&apps)
	for _, app := range apps {
		if app.Spec.Alias == "" {
			app.Spec.Alias = fmt.Sprintf("%s_alias", app.Spec.Name)
			tx.Save(&apps)
		}
	}

	return nil

}

// mig20231113103022Down for down migration
func mig20231113103022Down(tx *gorm.DB) error {

	// Applications:
	type Applications struct {
		Alias        string `gorm:"type:varchar(255) not null;"`
		CredentialID uint   `gorm:"type:bigint(1) unsigned not null"`
	}

	// delete old column
	if tx.Migrator().HasColumn(&Applications{}, "alias") {
		if err := tx.Migrator().DropColumn(&Applications{}, "alias"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&Applications{}, "credential_id") {
		if err := tx.Migrator().DropColumn(&Applications{}, "credential_id"); err != nil {
			return err
		}
	}

	return nil

}
