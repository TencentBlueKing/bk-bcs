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
		Version: "20240930182339",
		Name:    "20240930182339_modify_strategy",
		Mode:    migrator.GormMode,
		Up:      mig20240930182339Up,
		Down:    mig20240930182339Down,
	})
}

// Strategies  : strategies
type Strategies struct {
	PublishType       string `gorm:"column:publish_type;type:varchar(20);default:'';NOT NULL"`
	PublishTime       string `gorm:"column:publish_time;type:varchar(20);default:NULL"`
	PublishStatus     string `gorm:"column:publish_status;type:varchar(20);default:'';NOT NULL"`
	RejectReason      string `gorm:"column:reject_reason;type:varchar(256);default:NULL"`
	Approver          string `gorm:"column:approver;type:varchar(256);default:NULL"`
	ApproverProgress  string `gorm:"column:approver_progress;type:varchar(256);default:NULL"`
	ItsmTicketType    string `gorm:"column:itsm_ticket_type;type:varchar(256);default:NULL"`
	ItsmTicketUrl     string `gorm:"column:itsm_ticket_url;type:varchar(256);default:NULL"`
	ItsmTicketSn      string `gorm:"column:itsm_ticket_sn;type:varchar(256);default:NULL"`
	ItsmTicketStatus  string `gorm:"column:itsm_ticket_status;type:varchar(256);default:NULL"`
	ItsmTicketStateId int    `gorm:"column:itsm_ticket_state_id;type:int(11);default:NULL"`
}

// mig20240930182339Up for up migration
func mig20240930182339Up(tx *gorm.DB) error {
	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "publish_type") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "publish_type"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "publish_time") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "publish_time"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "publish_status") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "publish_status"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "reject_reason") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "reject_reason"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "approver") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "approver"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "approver_progress") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "approver_progress"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "itsm_ticket_type") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "itsm_ticket_type"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "itsm_ticket_url") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "itsm_ticket_url"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "itsm_ticket_sn") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "itsm_ticket_sn"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "itsm_ticket_status") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "itsm_ticket_status"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "itsm_ticket_state_id") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "itsm_ticket_state_id"); err != nil {
			return err
		}
	}
	return nil
}

// mig20240930182339Down for down migration
func mig20240930182339Down(tx *gorm.DB) error {
	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "publish_type") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "publish_type"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "publish_time") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "publish_time"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "publish_status") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "publish_status"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "reject_reason") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "reject_reason"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "approver") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "approver"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "approver_progress") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "approver_progress"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "itsm_ticket_type") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "itsm_ticket_type"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "itsm_ticket_url") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "itsm_ticket_url"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "itsm_ticket_sn") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "itsm_ticket_sn"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "itsm_ticket_status") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "itsm_ticket_status"); err != nil {
			return err
		}
	}

	// Strategies add new column
	if !tx.Migrator().HasColumn(&Strategies{}, "itsm_ticket_state_id") {
		if err := tx.Migrator().AddColumn(&Strategies{}, "itsm_ticket_state_id"); err != nil {
			return err
		}
	}
	return nil
}
