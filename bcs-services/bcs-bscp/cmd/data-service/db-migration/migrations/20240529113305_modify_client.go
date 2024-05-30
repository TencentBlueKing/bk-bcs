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
		Version: "20240529113305",
		Name:    "20240529113305_modify_client",
		Mode:    migrator.GormMode,
		Up:      mig20240529113305Up,
		Down:    mig20240529113305Down,
	})
}

// mig20240529113305Up for up migration
func mig20240529113305Up(tx *gorm.DB) error {
	// Clients : client
	type Clients struct {
		FailedDetailReason string `gorm:"column:failed_detail_reason;type:varchar(1030);default:'';NOT NULL"`
	}

	// ClientEvents : client_events
	type ClientEvents struct {
		FailedDetailReason string `gorm:"column:failed_detail_reason;type:varchar(1030);default:'';NOT NULL"`
	}

	if tx.Migrator().HasColumn(&Clients{}, "failed_detail_reason") {
		if err := tx.Migrator().AlterColumn(&Clients{}, "failed_detail_reason"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&ClientEvents{}, "failed_detail_reason") {
		if err := tx.Migrator().AlterColumn(&ClientEvents{}, "failed_detail_reason"); err != nil {
			return err
		}
	}

	return nil
}

// mig20240529113305Down for down migration
func mig20240529113305Down(tx *gorm.DB) error {
	// Clients : client
	type Clients struct {
		FailedDetailReason string `gorm:"column:failed_detail_reason;type:varchar(600);default:'';NOT NULL"`
	}

	// ClientEvents : client_events
	type ClientEvents struct {
		FailedDetailReason string `gorm:"column:failed_detail_reason;type:varchar(600);default:'';NOT NULL"`
	}

	if tx.Migrator().HasColumn(&Clients{}, "failed_detail_reason") {
		if err := tx.Migrator().AlterColumn(&Clients{}, "failed_detail_reason"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&ClientEvents{}, "failed_detail_reason") {
		if err := tx.Migrator().AlterColumn(&ClientEvents{}, "failed_detail_reason"); err != nil {
			return err
		}
	}

	return nil
}
