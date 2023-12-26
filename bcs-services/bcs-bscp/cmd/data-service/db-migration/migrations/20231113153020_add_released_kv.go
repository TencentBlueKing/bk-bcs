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

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrator"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20231113153020",
		Name:    "20231113153020_add_released_kv",
		Mode:    migrator.GormMode,
		Up:      mig20231113153020Up,
		Down:    mig20231113153020Down,
	})
}

// mig20231113153020Up for up migration
func mig20231113153020Up(tx *gorm.DB) error {

	// ReleasedKvs :已生成版本的kv
	type ReleasedKvs struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Key       string `gorm:"type:varchar(255) not null;uniqueIndex:relID_key,priority:1"`
		Version   uint   `gorm:"type:bigint(1) unsigned not null;"`
		ReleaseID uint   `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_ID,priority:3;uniqueIndex:relID_key,priority:2"` //nolint:lll
		KvType    string `gorm:"type:varchar(64) not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_ID,priority:1"`
		AppID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_ID,priority:2"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").AutoMigrate(
		&ReleasedKvs{},
	); err != nil {
		return err
	}

	if result := tx.Create([]IDGenerators{
		{Resource: "released_kvs", MaxID: 0},
	}); result.Error != nil {
		return result.Error
	}

	return nil

}

// mig20231113153020Down for down migration
func mig20231113153020Down(tx *gorm.DB) error {

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Migrator().
		DropTable("released_kvs"); err != nil {
		return err
	}

	var resources = []string{
		"released_kvs",
	}
	if result := tx.Where("resource in ?", resources).Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	return nil
}
