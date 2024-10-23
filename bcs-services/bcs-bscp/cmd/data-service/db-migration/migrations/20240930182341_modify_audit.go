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
		Version: "20240930182341",
		Name:    "20240930182341_modify_audit",
		Mode:    migrator.GormMode,
		Up:      mig20240930182341Up,
		Down:    mig20240930182341Down,
	})
}

// Audits  : audits
type Audits struct {
	ResInstance string `gorm:"column:res_instance;type:varchar(256);default:'';NOT NULL"`
	OperateWay  string `gorm:"column:operate_way;type:varchar(20);default:'';NOT NULL"`
	Status      string `gorm:"column:status;type:varchar(20);default:'';NOT NULL"`
	StrategyId  uint   `gorm:"column:strategy_id;type:bigint(1) unsigned;default:NULL;index:idx_strategyID,priority:1"`
	IsCompare   uint   `gorm:"column:is_compare;type:tinyint(1) unsigned;default:0;NOT NULL"`
}

// mig20240930182341Up for up migration
func mig20240930182341Up(tx *gorm.DB) error {
	// Audits add new column
	if !tx.Migrator().HasColumn(&Audits{}, "res_instance") {
		if err := tx.Migrator().AddColumn(&Audits{}, "res_instance"); err != nil {
			return err
		}
	}

	// Audits add new column
	if !tx.Migrator().HasColumn(&Audits{}, "operate_way") {
		if err := tx.Migrator().AddColumn(&Audits{}, "operate_way"); err != nil {
			return err
		}
	}

	// Audits add new column
	if !tx.Migrator().HasColumn(&Audits{}, "status") {
		if err := tx.Migrator().AddColumn(&Audits{}, "status"); err != nil {
			return err
		}
	}

	// Audits add new column
	if !tx.Migrator().HasColumn(&Audits{}, "strategy_id") {
		if err := tx.Migrator().AddColumn(&Audits{}, "strategy_id"); err != nil {
			return err
		}
	}

	// Audits add new column
	if !tx.Migrator().HasColumn(&Audits{}, "is_compare") {
		if err := tx.Migrator().AddColumn(&Audits{}, "is_compare"); err != nil {
			return err
		}
	}

	return nil
}

// mig20240930182341Down for down migration
func mig20240930182341Down(tx *gorm.DB) error {
	// Audits add new column
	if !tx.Migrator().HasColumn(&Audits{}, "res_instance") {
		if err := tx.Migrator().AddColumn(&Audits{}, "res_instance"); err != nil {
			return err
		}
	}

	// Audits add new column
	if !tx.Migrator().HasColumn(&Audits{}, "operate_way") {
		if err := tx.Migrator().AddColumn(&Audits{}, "operate_way"); err != nil {
			return err
		}
	}

	// Audits add new column
	if !tx.Migrator().HasColumn(&Audits{}, "status") {
		if err := tx.Migrator().AddColumn(&Audits{}, "status"); err != nil {
			return err
		}
	}

	// Audits add new column
	if !tx.Migrator().HasColumn(&Audits{}, "strategy_id") {
		if err := tx.Migrator().AddColumn(&Audits{}, "strategy_id"); err != nil {
			return err
		}
	}

	// Audits add new column
	if !tx.Migrator().HasColumn(&Audits{}, "is_compare") {
		if err := tx.Migrator().AddColumn(&Audits{}, "is_compare"); err != nil {
			return err
		}
	}

	return nil
}
