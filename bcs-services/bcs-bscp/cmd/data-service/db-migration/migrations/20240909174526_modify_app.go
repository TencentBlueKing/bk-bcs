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
	"time"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrator"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20240909174526",
		Name:    "20240909174526_modify_app",
		Mode:    migrator.GormMode,
		Up:      mig20240909174526Up,
		Down:    mig20240909174526Down,
	})
}

// mig20240909174526Up for up migration
func mig20240909174526Up(tx *gorm.DB) error {
	// Applications  : applications
	type Applications struct {
		LastConsumedTime time.Time `gorm:"column:last_consumed_time;type:datetime(6);default:NULL"`
	}

	// Applications add new column
	if !tx.Migrator().HasColumn(&Applications{}, "last_consumed_time") {
		if err := tx.Migrator().AddColumn(&Applications{}, "last_consumed_time"); err != nil {
			return err
		}
	}

	return nil
}

// mig20240909174526Down for down migration
func mig20240909174526Down(tx *gorm.DB) error {

	// Applications  : applications
	type Applications struct {
		LastConsumedTime time.Time `gorm:"column:last_consumed_time;type:datetime(6);default:NULL"`
	}

	// Applications drop column
	if tx.Migrator().HasColumn(&Applications{}, "last_consumed_time") {
		if err := tx.Migrator().DropColumn(&Applications{}, "last_consumed_time"); err != nil {
			return err
		}
	}

	return nil
}
