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
		Version: "20230625152115",
		Name:    "20230625152115_gorm_add_app_template_related",
		Mode:    migrator.GormMode,
		Up:      mig20230625152115GormAddAppTemplateRelatedUp,
		Down:    mig20230625152115GormAddAppTemplateRelatedDown,
	})
}

// mig20230625152115GormAddAppTemplateRelatedUp for up migration
func mig20230625152115GormAddAppTemplateRelatedUp(tx *gorm.DB) error {
	// AppTemplateBindings : 记录未命名服务版本与模版套餐及模版版本的绑定情况
	type AppTemplateBindings struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		TemplateSpaceIDs   string `gorm:"type:json not null"`
		TemplateSetIDs     string `gorm:"type:json not null"`
		TemplateIDs        string `gorm:"type:json not null"`
		TemplateReleaseIDs string `gorm:"type:json not null"`
		Bindings           string `gorm:"type:json not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID,priority:1"`
		AppID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID,priority:2"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// ReleasedAppTemplateBindings : 记录已发布服务版本与模版套餐及模版版本的绑定情况
	type ReleasedAppTemplateBindings struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		TemplateSpaceIDs   string `gorm:"type:json not null"`
		TemplateSetIDs     string `gorm:"type:json not null"`
		TemplateIDs        string `gorm:"type:json not null"`
		TemplateReleaseIDs string `gorm:"type:json not null"`
		Bindings           string `gorm:"type:json not null"`
		ReleaseID          uint   `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_RelID,priority:3"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_RelID,priority:1"`
		AppID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_RelID,priority:2"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// AppTemplateVariables : 记录未命名版本服务引用的模版配置项渲染用的变量(变量kv存为json串：variables字段)
	type AppTemplateVariables struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Variables string `gorm:"type:json not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID,priority:1"`
		AppID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID,priority:2"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// ReleasedAppTemplateVariables : 记录已发布服务版本引用的模版配置项渲染用的变量(变量kv存为json串：variables字段)
	type ReleasedAppTemplateVariables struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Variables string `gorm:"type:json not null"`
		ReleaseID uint   `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_RelID,priority:3"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_RelID,priority:1"`
		AppID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID_RelID,priority:2"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").AutoMigrate(
		&AppTemplateBindings{},
		&ReleasedAppTemplateBindings{},
		&AppTemplateVariables{},
		&ReleasedAppTemplateVariables{},
	); err != nil {
		return err
	}

	return nil

}

// mig20230625152115GormAddAppTemplateRelatedDown for down migration
func mig20230625152115GormAddAppTemplateRelatedDown(tx *gorm.DB) error {
	if err := tx.Migrator().DropTable(
		"app_template_bindings",
		"released_app_template_bindings",
		"app_template_variables",
		"released_template_variables",
	); err != nil {
		return err
	}

	return nil
}
