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
		Version: "20231123143015",
		Name:    "20231123143015_modify_app",
		Mode:    migrator.GormMode,
		Up:      mig20231123143015Up,
		Down:    mig20231123143015Down,
	})
}

// mig20231123143015Up for up migration
func mig20231123143015Up(tx *gorm.DB) error {

	// Applications: 服务
	type Applications struct {
		Alias    string `gorm:"type:varchar(255) not null;"`
		DataType string `gorm:"type:varchar(255) not null;"`
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
		}
		if app.Spec.ConfigType == table.KV {
			// 原有的kv类型app，数据类型全部记为any
			app.Spec.DataType = table.KvAny
		}
		tx.Save(&apps)
	}

	return nil

}

// mig20231123143015Down for down migration
func mig20231123143015Down(tx *gorm.DB) error {

	// Applications: 服务
	type Applications struct {
		Alias    string `gorm:"type:varchar(255) not null;"`
		DataType string `gorm:"type:varchar(255) not null;"`
	}

	// delete old column
	if tx.Migrator().HasColumn(&Applications{}, "alias") {
		if err := tx.Migrator().DropColumn(&Applications{}, "alias"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&Applications{}, "data_type") {
		if err := tx.Migrator().DropColumn(&Applications{}, "data_type"); err != nil {
			return err
		}
	}

	return nil
}
