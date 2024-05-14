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
		Version: "20240318212059",
		Name:    "20240318212059_modify_app",
		Mode:    migrator.GormMode,
		Up:      mig20240318212059Up,
		Down:    mig20240318212059Down,
	})
}

// mig20240318212059Up for up migration
func mig20240318212059Up(tx *gorm.DB) error {
	// Applications : 服务
	type Applications struct {
		Mode           string
		ReloadType     string
		ReloadFilePath string
	}

	// Strategies : 发布策略
	type Strategies struct {
		Mode string
	}

	/*** 字段变更 ***/
	// delete column
	if tx.Migrator().HasColumn(&Applications{}, "Mode") {
		if err := tx.Migrator().DropColumn(&Applications{}, "Mode"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&Applications{}, "ReloadType") {
		if err := tx.Migrator().DropColumn(&Applications{}, "ReloadType"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&Applications{}, "ReloadFilePath") {
		if err := tx.Migrator().DropColumn(&Applications{}, "ReloadFilePath"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&Strategies{}, "Mode") {
		if err := tx.Migrator().DropColumn(&Strategies{}, "Mode"); err != nil {
			return err
		}
	}

	return nil
}

// mig20240318212059Down for down migration
func mig20240318212059Down(tx *gorm.DB) error {
	return nil
}
