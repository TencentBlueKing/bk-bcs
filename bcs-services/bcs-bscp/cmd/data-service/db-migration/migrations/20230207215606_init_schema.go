/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package migrations

import (
	"strings"

	"gorm.io/gorm"

	"bscp.io/cmd/data-service/db-migration/migrator"
)

const mig20230207215606 = "20230207215606_init_schema"

func init() {
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20230207215606",
		Name:    "20230207215606_init_schema",
		Mode:    migrator.SqlMode,
		Up:      mig20230207215606InitSchemaUp,
		Down:    mig20230207215606InitSchemaDown,
	})
}

func mig20230207215606InitSchemaUp(tx *gorm.DB) error {
	sqlArr := strings.Split(migrator.GetMigrator().MigrationSQLs[migrator.GetUpSQLKey(mig20230207215606)], ";")
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

func mig20230207215606InitSchemaDown(tx *gorm.DB) error {
	sqlArr := strings.Split(migrator.GetMigrator().MigrationSQLs[migrator.GetDownSQLKey(mig20230207215606)], ";")
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
