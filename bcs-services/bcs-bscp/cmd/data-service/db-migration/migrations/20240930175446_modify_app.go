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
		Version: "20240930175446",
		Name:    "20240930175446_modify_app",
		Mode:    migrator.GormMode,
		Up:      mig20240930175446Up,
		Down:    mig20240930175446Down,
	})
}

// Applications  : applications
type Applications struct {
	ApproveType string `gorm:"column:approve_type;type:varchar(20);default:NULL"`
	Approver    string `gorm:"column:approver;type:varchar(256);default:NULL"`
	IsApprove   uint   `gorm:"column:is_approve;type:tinyint(1) unsigned;default:0;NOT NULL"`
}

// mig20240930175446Up for up migration
func mig20240930175446Up(tx *gorm.DB) error {

	// Applications add new column
	if !tx.Migrator().HasColumn(&Applications{}, "is_approve") {
		if err := tx.Migrator().AddColumn(&Applications{}, "is_approve"); err != nil {
			return err
		}
	}

	// Applications add new column
	if !tx.Migrator().HasColumn(&Applications{}, "approve_type") {
		if err := tx.Migrator().AddColumn(&Applications{}, "approve_type"); err != nil {
			return err
		}
	}

	// Applications add new column
	if !tx.Migrator().HasColumn(&Applications{}, "approver") {
		if err := tx.Migrator().AddColumn(&Applications{}, "approver"); err != nil {
			return err
		}
	}

	return nil
}

// mig20240930175446Down for down migration
func mig20240930175446Down(tx *gorm.DB) error {
	// Applications add new column
	if !tx.Migrator().HasColumn(&Applications{}, "is_approve") {
		if err := tx.Migrator().AddColumn(&Applications{}, "is_approve"); err != nil {
			return err
		}
	}

	// Applications add new column
	if !tx.Migrator().HasColumn(&Applications{}, "approve_type") {
		if err := tx.Migrator().AddColumn(&Applications{}, "approve_type"); err != nil {
			return err
		}
	}

	// Applications add new column
	if !tx.Migrator().HasColumn(&Applications{}, "approver") {
		if err := tx.Migrator().AddColumn(&Applications{}, "approver"); err != nil {
			return err
		}
	}
	return nil
}
