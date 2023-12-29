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
		Version: "20230905163915",
		Name:    "20230905163915_modify_released_ci",
		Mode:    migrator.GormMode,
		Up:      mig20230905163915Up,
		Down:    mig20230905163915Down,
	})
}

// mig20230905163915Up for up migration
func mig20230905163915Up(tx *gorm.DB) error {
	// ReleasedConfigItems : 已生成版本的配置项
	type ReleasedConfigItems struct {
		// 这里只需要DB操作用到的字段
		OriginSignature string `gorm:"type:varchar(64) not null"`
		OriginByteSize  uint   `gorm:"type:bigint(1) unsigned not null"`

		BizID     uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_relID,priority:1"`
		AppID     uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_relID,priority:2"`
		ReleaseID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_relID,priority:3"`
	}

	/*** 字段变更 ***/
	// add new column
	if !tx.Migrator().HasColumn(&ReleasedConfigItems{}, "OriginSignature") {
		if err := tx.Migrator().AddColumn(&ReleasedConfigItems{}, "OriginSignature"); err != nil {
			return err
		}
	}
	if !tx.Migrator().HasColumn(&ReleasedConfigItems{}, "OriginByteSize") {
		if err := tx.Migrator().AddColumn(&ReleasedConfigItems{}, "OriginByteSize"); err != nil {
			return err
		}
	}

	/*** 索引变更 ***/
	// delete old index
	if tx.Migrator().HasIndex(&ReleasedConfigItems{}, "idx_releaseID_commitID") {
		if err := tx.Migrator().DropIndex(&ReleasedConfigItems{}, "idx_releaseID_commitID"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasIndex(&ReleasedConfigItems{}, "idx_bizID_appID") {
		if err := tx.Migrator().DropIndex(&ReleasedConfigItems{}, "idx_bizID_appID"); err != nil {
			return err
		}
	}

	// create new index
	if !tx.Migrator().HasIndex(&ReleasedConfigItems{}, "idx_bizID_appID_relID") {
		if err := tx.Migrator().CreateIndex(&ReleasedConfigItems{}, "idx_bizID_appID_relID"); err != nil {
			return err
		}
	}

	return nil

}

// mig20230905163915Down for down migration
func mig20230905163915Down(tx *gorm.DB) error {
	return nil
}
