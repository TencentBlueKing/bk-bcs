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
		Version: "20240914104733",
		Name:    "20240914104733_modif_client",
		Mode:    migrator.GormMode,
		Up:      mig20240914104733Up,
		Down:    mig20240914104733Down,
	})
}

// mig20240914104733Up for up migration
func mig20240914104733Up(tx *gorm.DB) error {

	// Clients  : clients
	type Clients struct {
		TotalSeconds float64 `gorm:"column:total_seconds;type:double unsigned;default:0;NOT NULL"`
	}

	// Clients add new column
	if !tx.Migrator().HasColumn(&Clients{}, "total_seconds") {
		if err := tx.Migrator().AddColumn(&Clients{}, "total_seconds"); err != nil {
			return err
		}
	}

	return nil
}

// mig20240914104733Down for down migration
func mig20240914104733Down(tx *gorm.DB) error {

	// Clients  : clients
	type Clients struct {
		TotalSeconds float64 `gorm:"column:total_seconds;type:double unsigned;default:0;NOT NULL"`
	}

	// Clients drop column
	if tx.Migrator().HasColumn(&Clients{}, "total_seconds") {
		if err := tx.Migrator().DropColumn(&Clients{}, "total_seconds"); err != nil {
			return err
		}
	}

	return nil
}
