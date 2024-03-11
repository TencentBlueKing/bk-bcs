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
		Version: "20240221103510",
		Name:    "20240221103510_kv_add_md5",
		Mode:    migrator.GormMode,
		Up:      mig20240221103510Up,
		Down:    mig20240221103510Down,
	})
}

// mig20240221103510Up for up migration
func mig20240221103510Up(tx *gorm.DB) error {
	// 调用kv add signature 自动执行 md5 migrate
	return mig20240125175500Up(tx)
}

// mig20240221103510Down for down migration
func mig20240221103510Down(tx *gorm.DB) error {
	// 保持幂等
	// delete kvs old column
	if tx.Migrator().HasColumn(&Kvs20240125175500{}, "Md5") {
		if err := tx.Migrator().DropColumn(&Kvs20240125175500{}, "Md5"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&ReleasedKvs20240125175500{}, "Md5") {
		if err := tx.Migrator().DropColumn(&Kvs20240125175500{}, "Md5"); err != nil {
			return err
		}
	}

	return nil
}
