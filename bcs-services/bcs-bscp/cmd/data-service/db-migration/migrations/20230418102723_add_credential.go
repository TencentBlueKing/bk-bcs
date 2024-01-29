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
	"strings"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrator"
)

const mig20230418102723 = "20230418102723_add_credential"

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20230418102723",
		Name:    "20230418102723_add_credential",
		Mode:    migrator.SqlMode,
		Up:      mig20230418102723Up,
		Down:    mig20230418102723Down,
	})
}

// mig20230418102723Up for up migration
func mig20230418102723Up(tx *gorm.DB) error {
	sqlArr := strings.Split(migrator.GetMigrator().MigrationSQLs[migrator.GetUpSQLKey(mig20230418102723)], ";")
	for _, sql := range sqlArr {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		if result := tx.Exec(sql); result.Error != nil {
			return result.Error
		}
	}

	return nil

}

// mig20230418102723Down for down migration
func mig20230418102723Down(tx *gorm.DB) error {
	sqlArr := strings.Split(migrator.GetMigrator().MigrationSQLs[migrator.GetDownSQLKey(mig20230418102723)], ";")
	for _, sql := range sqlArr {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		if result := tx.Exec(sql); result.Error != nil {
			return result.Error
		}
	}

	return nil
}
