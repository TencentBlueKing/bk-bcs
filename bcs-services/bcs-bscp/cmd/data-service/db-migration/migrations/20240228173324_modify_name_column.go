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
		Version: "20240228173324",
		Name:    "20240228173324_modify_name_column",
		Mode:    migrator.GormMode,
		Up:      mig20240228173324Up,
		Down:    mig20240228173324Down,
	})
}

// mig20240228173324Up for up migration
// 字段变更，设置为collate为utf8_bin，查询时区分大小写，场景如：字段唯一性校验时，字符串"App"和"app"能区分为不同
//
//nolint:funlen
func mig20240228173324Up(tx *gorm.DB) error {
	// Applications : 服务
	type Applications struct {
		Name  string `gorm:"type:varchar(255) collate utf8_bin not null"`
		Alias string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&Applications{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Applications{}, "Name"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&Applications{}, "Alias") {
		if err := tx.Migrator().AlterColumn(&Applications{}, "Alias"); err != nil {
			return err
		}
	}

	// ConfigItems : 配置项
	type ConfigItems struct {
		Name string `gorm:"type:varchar(255) collate utf8_bin not null"`
		Path string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&ConfigItems{}, "Name") {
		if err := tx.Migrator().AlterColumn(&ConfigItems{}, "Name"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&ConfigItems{}, "Path") {
		if err := tx.Migrator().AlterColumn(&ConfigItems{}, "Path"); err != nil {
			return err
		}
	}

	// Credentials : 密钥
	type Credentials struct {
		Name string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&Credentials{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Credentials{}, "Name"); err != nil {
			return err
		}
	}

	// Groups : 分组
	type Groups struct {
		Name string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&Groups{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Groups{}, "Name"); err != nil {
			return err
		}
	}

	// HookRevisions : 脚本版本
	type HookRevisions struct {
		Name string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&HookRevisions{}, "Name") {
		if err := tx.Migrator().AlterColumn(&HookRevisions{}, "Name"); err != nil {
			return err
		}
	}

	// Hooks : 脚本
	type Hooks struct {
		Name string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&Hooks{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Hooks{}, "Name"); err != nil {
			return err
		}
	}

	// Kvs : 键值对配置
	type Kvs struct {
		Key string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&Kvs{}, "Key") {
		if err := tx.Migrator().AlterColumn(&Kvs{}, "Key"); err != nil {
			return err
		}
	}

	// Releases : 版本
	type Releases struct {
		Name string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&Releases{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Releases{}, "Name"); err != nil {
			return err
		}
	}

	// TemplateRevisions : 模版版本
	type TemplateRevisions struct {
		RevisionName string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&TemplateRevisions{}, "RevisionName") {
		if err := tx.Migrator().AlterColumn(&TemplateRevisions{}, "RevisionName"); err != nil {
			return err
		}
	}

	// TemplateSets : 模版套餐
	type TemplateSets struct {
		Name string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&TemplateSets{}, "Name") {
		if err := tx.Migrator().AlterColumn(&TemplateSets{}, "Name"); err != nil {
			return err
		}
	}

	// TemplateSpaces : 模版空间
	type TemplateSpaces struct {
		Name string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&TemplateSpaces{}, "Name") {
		if err := tx.Migrator().AlterColumn(&TemplateSpaces{}, "Name"); err != nil {
			return err
		}
	}

	// Templates : 模版
	type Templates struct {
		Name string `gorm:"type:varchar(255) collate utf8_bin not null"`
	}
	if tx.Migrator().HasColumn(&Templates{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Templates{}, "Name"); err != nil {
			return err
		}
	}

	return nil

}

// mig20240228173324Down for down migration
func mig20240228173324Down(tx *gorm.DB) error {
	return nil
}
