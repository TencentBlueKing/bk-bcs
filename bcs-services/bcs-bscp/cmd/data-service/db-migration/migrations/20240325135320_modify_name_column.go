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
		Version: "20240325135320",
		Name:    "20240325135320_modify_name_column",
		Mode:    migrator.GormMode,
		Up:      mig20240325135320Up,
		Down:    mig20240325135320Down,
	})
}

func modifyApp(tx *gorm.DB) error {
	// Applications : 服务
	type Applications struct {
		Name  string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
		Alias string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
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
	return nil
}

func modifyConfigItems(tx *gorm.DB) error {
	// ConfigItems : 配置项
	type ConfigItems struct {
		Name string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
		Path string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
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
	return nil
}

func modifyCredentials(tx *gorm.DB) error {
	// Credentials : 密钥
	type Credentials struct {
		Name string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&Credentials{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Credentials{}, "Name"); err != nil {
			return err
		}
	}
	return nil
}

func modifyGroups(tx *gorm.DB) error {
	// Groups : 分组
	type Groups struct {
		Name string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&Groups{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Groups{}, "Name"); err != nil {
			return err
		}
	}
	return nil
}

func modifyHookRevisions(tx *gorm.DB) error {
	// HookRevisions : 脚本版本
	type HookRevisions struct {
		Name string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&HookRevisions{}, "Name") {
		if err := tx.Migrator().AlterColumn(&HookRevisions{}, "Name"); err != nil {
			return err
		}
	}
	return nil
}

func modifyHooks(tx *gorm.DB) error {
	// Hooks : 脚本
	type Hooks struct {
		Name string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&Hooks{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Hooks{}, "Name"); err != nil {
			return err
		}
	}
	return nil
}

func modifyKvs(tx *gorm.DB) error {
	// Kvs : 键值对配置
	type Kvs struct {
		Key string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&Kvs{}, "Key") {
		if err := tx.Migrator().AlterColumn(&Kvs{}, "Key"); err != nil {
			return err
		}
	}
	return nil
}

func modifyReleases(tx *gorm.DB) error {
	// Releases : 版本
	type Releases struct {
		Name string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&Releases{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Releases{}, "Name"); err != nil {
			return err
		}
	}
	return nil
}

func modifyReleasedKvs(tx *gorm.DB) error {
	// ReleasedKvs : 已发布Kvs
	type ReleasedKvs struct {
		Key string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&ReleasedKvs{}, "Key") {
		if err := tx.Migrator().AlterColumn(&ReleasedKvs{}, "Key"); err != nil {
			return err
		}
	}
	return nil
}

func modifyTemplateRevisions(tx *gorm.DB) error {
	// TemplateRevisions : 模版版本
	type TemplateRevisions struct {
		RevisionName string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&TemplateRevisions{}, "RevisionName") {
		if err := tx.Migrator().AlterColumn(&TemplateRevisions{}, "RevisionName"); err != nil {
			return err
		}
	}
	return nil
}

func modifyTemplateSets(tx *gorm.DB) error {
	// TemplateSets : 模版套餐
	type TemplateSets struct {
		Name string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&TemplateSets{}, "Name") {
		if err := tx.Migrator().AlterColumn(&TemplateSets{}, "Name"); err != nil {
			return err
		}
	}
	return nil
}

func modifyTemplateSpaces(tx *gorm.DB) error {
	// TemplateSpaces : 模版空间
	type TemplateSpaces struct {
		Name string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&TemplateSpaces{}, "Name") {
		if err := tx.Migrator().AlterColumn(&TemplateSpaces{}, "Name"); err != nil {
			return err
		}
	}
	return nil
}

func modifyTemplateVariables(tx *gorm.DB) error {
	// TemplateVariables : 模版变量
	type TemplateVariables struct {
		Name string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&TemplateVariables{}, "Name") {
		if err := tx.Migrator().AlterColumn(&TemplateVariables{}, "Name"); err != nil {
			return err
		}
	}
	return nil
}

func modifyTemplates(tx *gorm.DB) error {
	// Templates : 模版
	type Templates struct {
		Name string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
		Path string `gorm:"type:varchar(255) collate utf8mb4_bin not null"`
	}
	if tx.Migrator().HasColumn(&Templates{}, "Name") {
		if err := tx.Migrator().AlterColumn(&Templates{}, "Name"); err != nil {
			return err
		}
	}
	if tx.Migrator().HasColumn(&Templates{}, "Path") {
		if err := tx.Migrator().AlterColumn(&Templates{}, "Path"); err != nil {
			return err
		}
	}
	return nil
}

func doModify(tx *gorm.DB, fs ...func(tx *gorm.DB) error) error {
	for _, f := range fs {
		if err := f(tx); err != nil {
			return err
		}
	}
	return nil
}

// mig20240325135320Up for up migration
// 字段变更，设置为collate为utf8mb4_bin，查询时区分大小写，场景如：字段唯一性校验时，字符串"App"和"app"能区分为不同
func mig20240325135320Up(tx *gorm.DB) error {
	return doModify(tx,
		modifyApp,
		modifyConfigItems,
		modifyCredentials,
		modifyGroups,
		modifyHookRevisions,
		modifyHooks,
		modifyKvs,
		modifyReleases,
		modifyReleasedKvs,
		modifyTemplateRevisions,
		modifyTemplateSets,
		modifyTemplateSpaces,
		modifyTemplateVariables,
		modifyTemplates)
}

// mig20240325135320Down for down migration
func mig20240325135320Down(tx *gorm.DB) error {
	return nil
}
