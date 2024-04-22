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
		Version: "20240410112938",
		Name:    "20240410112938_modify_client",
		Mode:    migrator.GormMode,
		Up:      mig20240410112938Up,
		Down:    mig20240410112938Down,
	})
}

// mig20240410112938Up for up migration
func mig20240410112938Up(tx *gorm.DB) error {
	// Clients : client
	type Clients struct {
		CpuMinUsage               float64 `gorm:"column:cpu_min_usage;type:double unsigned;default:0;NOT NULL"`
		CpuAvgUsage               float64 `gorm:"column:cpu_avg_usage;type:double unsigned;default:0;NOT NULL"`
		MemoryMinUsage            uint64  `gorm:"column:memory_min_usage;type:bigint(1) unsigned;default:0;NOT NULL"`
		MemoryAvgUsage            uint64  `gorm:"column:memory_avg_usage;type:bigint(1) unsigned;default:0;NOT NULL"`
		SpecificFailedReason      string  `gorm:"column:specific_failed_reason;type:varchar(50);default:'';NOT NULL"`
		ReleaseChangeFailedReason string  `gorm:"column:release_change_failed_reason;type:varchar(50);default:'';NOT NULL"`
		FailedDetailReason        string  `gorm:"column:failed_detail_reason;type:varchar(600);default:'';NOT NULL"`
	}

	// ClientEvents : client_events
	type ClientEvents struct {
		SpecificFailedReason      string `gorm:"column:specific_failed_reason;type:varchar(20);default:'';NOT NULL"`
		ReleaseChangeFailedReason string `gorm:"column:release_change_failed_reason;type:varchar(50);default:'';NOT NULL"`
		FailedDetailReason        string `gorm:"column:failed_detail_reason;type:varchar(600);default:'';NOT NULL"`
	}

	// add new column
	if !tx.Migrator().HasColumn(&Clients{}, "cpu_min_usage") {
		if err := tx.Migrator().AddColumn(&Clients{}, "cpu_min_usage"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&Clients{}, "cpu_avg_usage") {
		if err := tx.Migrator().AddColumn(&Clients{}, "cpu_avg_usage"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&Clients{}, "memory_min_usage") {
		if err := tx.Migrator().AddColumn(&Clients{}, "memory_min_usage"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&Clients{}, "memory_avg_usage") {
		if err := tx.Migrator().AddColumn(&Clients{}, "memory_avg_usage"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&Clients{}, "specific_failed_reason") {
		if err := tx.Migrator().AddColumn(&Clients{}, "specific_failed_reason"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&Clients{}, "release_change_failed_reason") {
		if err := tx.Migrator().AlterColumn(&Clients{}, "release_change_failed_reason"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&Clients{}, "failed_detail_reason") {
		if err := tx.Migrator().AlterColumn(&Clients{}, "failed_detail_reason"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&ClientEvents{}, "specific_failed_reason") {
		if err := tx.Migrator().AddColumn(&ClientEvents{}, "specific_failed_reason"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&ClientEvents{}, "release_change_failed_reason") {
		if err := tx.Migrator().AlterColumn(&ClientEvents{}, "release_change_failed_reason"); err != nil {
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

// mig20240410112938Down for down migration
func mig20240410112938Down(tx *gorm.DB) error {
	// Clients : client
	type Clients struct {
		CpuMinUsage               float64 `gorm:"column:cpu_min_usage;type:double unsigned;default:0;NOT NULL"`
		CpuAvgUsage               float64 `gorm:"column:cpu_avg_usage;type:double unsigned;default:0;NOT NULL"`
		MemoryMinUsage            uint64  `gorm:"column:memory_min_usage;type:bigint(1) unsigned;default:0;NOT NULL"`
		MemoryAvgUsage            uint64  `gorm:"column:memory_avg_usage;type:bigint(1) unsigned;default:0;NOT NULL"`
		SpecificFailedReason      string  `gorm:"column:specific_failed_reason;type:varchar(20);default:'';NOT NULL"`
		ReleaseChangeFailedReason string  `gorm:"column:release_change_failed_reason;type:varchar(50);default:'';NOT NULL"`
		FailedDetailReason        string  `gorm:"column:failed_detail_reason;type:varchar(600);default:'';NOT NULL"`
	}

	// ClientEvents : client_events
	type ClientEvents struct {
		SpecificFailedReason      string `gorm:"column:specific_failed_reason;type:varchar(20);default:'';NOT NULL"`
		ReleaseChangeFailedReason string `gorm:"column:release_change_failed_reason;type:varchar(50);default:'';NOT NULL"`
		FailedDetailReason        string `gorm:"column:failed_detail_reason;type:varchar(600);default:'';NOT NULL"`
	}

	// add new column
	if tx.Migrator().HasColumn(&Clients{}, "cpu_min_usage") {
		if err := tx.Migrator().DropColumn(&Clients{}, "cpu_min_usage"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&Clients{}, "cpu_avg_usage") {
		if err := tx.Migrator().DropColumn(&Clients{}, "cpu_avg_usage"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&Clients{}, "memory_min_usage") {
		if err := tx.Migrator().DropColumn(&Clients{}, "memory_min_usage"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&Clients{}, "memory_avg_usage") {
		if err := tx.Migrator().DropColumn(&Clients{}, "memory_avg_usage"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&Clients{}, "specific_failed_reason") {
		if err := tx.Migrator().DropColumn(&Clients{}, "specific_failed_reason"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&ClientEvents{}, "specific_failed_reason") {
		if err := tx.Migrator().DropColumn(&ClientEvents{}, "specific_failed_reason"); err != nil {
			return err
		}
	}

	return nil
}
