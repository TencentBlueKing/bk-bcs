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
	"time"

	"gorm.io/gorm"

	"bscp.io/cmd/data-service/db-migration/migrator"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20231009103050",
		Name:    "20231009103050_add_kv",
		Mode:    migrator.GormMode,
		Up:      mig20231009103050GormUp,
		Down:    mig20231009103050GormDown,
	})
}

// mig20231009103050GormUp for up migration
func mig20231009103050GormUp(tx *gorm.DB) error {

	// app
	type Applications struct {
		Alias        string `gorm:"type:varchar(255) not null;"`
		CredentialID uint   `gorm:"type:bigint(1) unsigned not null"`
	}

	// kvs
	type Kvs struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_appID_name,priority:1"`
		Type string `gorm:"type:varchar(255) not null;"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_name,priority:2"`
		APPID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_name,priority:3"`

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

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&Applications{}, &Kvs{}); err != nil {
		return err
	}

	now := time.Now()
	if result := tx.Create([]IDGenerators{
		{Resource: "kvs", MaxID: 0, UpdatedAt: now},
	}); result.Error != nil {
		return result.Error
	}

	return nil
}

// mig20231009103050GormDown for down migration
func mig20231009103050GormDown(tx *gorm.DB) error {

	// app
	type Applications struct {
		Alias        string `gorm:"type:varchar(255) not null;"`
		CredentialID uint   `gorm:"type:bigint(1) unsigned not null"`
	}

	// kvs
	type Kvs struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_appID_name,priority:1"`
		Type string `gorm:"type:varchar(255) not null;"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_name,priority:2"`
		APPID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_name,priority:3"`

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

	if err := tx.Migrator().DropTable(Kvs{}); err != nil {
		return err
	}

	if err := tx.Migrator().DropColumn(Applications{}, "alias"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(Applications{}, "credential_id"); err != nil {
		return err
	}

	var resources = []string{
		"kvs",
	}
	if result := tx.Where("resource IN ?", resources).Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	return nil
}
