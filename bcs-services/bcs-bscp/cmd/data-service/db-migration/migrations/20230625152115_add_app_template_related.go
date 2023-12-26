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
		Version: "20230625152115",
		Name:    "20230625152115_gorm_add_app_template_related",
		Mode:    migrator.GormMode,
		Up:      mig20230625152115Up,
		Down:    mig20230625152115Down,
	})
}

// mig20230625152115Up for up migration
func mig20230625152115Up(tx *gorm.DB) error { //nolint:funlen
	// AppTemplateBindings : 记录未命名版本服务版本与模版套餐及模版版本的绑定情况
	type AppTemplateBindings struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		TemplateSpaceIDs    string `gorm:"type:json not null"`
		TemplateSetIDs      string `gorm:"type:json not null"`
		TemplateIDs         string `gorm:"type:json not null"`
		TemplateRevisionIDs string `gorm:"type:json not null"`
		LatestTemplateIDs   string `gorm:"type:json not null"`
		Bindings            string `gorm:"type:json not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID,priority:1"`
		AppID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_appID,priority:2"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// ReleasedAppTemplates : 记录已发布服务版本的模版情况
	type ReleasedAppTemplates struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		ReleaseID            uint   `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_relID,priority:3"`
		TemplateSpaceID      uint   `gorm:"type:bigint(1) unsigned not null"`
		TemplateSpaceName    string `gorm:"type:varchar(255) not null"`
		TemplateSetID        uint   `gorm:"type:bigint(1) unsigned not null"`
		TemplateSetName      string `gorm:"type:varchar(255) not null"`
		TemplateID           uint   `gorm:"type:bigint(1) unsigned not null"`
		Name                 string `gorm:"type:varchar(255) not null"`
		Path                 string `gorm:"type:varchar(255) not null"`
		TemplateRevisionID   uint   `gorm:"type:bigint(1) unsigned not null"`
		IsLatest             bool   `gorm:"boolean default false"`
		TemplateRevisionName string `gorm:"type:varchar(255) not null"`
		TemplateRevisionMemo string `gorm:"type:varchar(256) default ''"`
		FileType             string `gorm:"type:varchar(20) not null"`
		FileMode             string `gorm:"type:varchar(20) not null"`
		User                 string `gorm:"type:varchar(64) not null"`
		UserGroup            string `gorm:"type:varchar(64) not null"`
		Privilege            string `gorm:"type:varchar(64) not null"`
		Signature            string `gorm:"type:varchar(64) not null"`
		ByteSize             uint   `gorm:"type:bigint(1) unsigned not null"`
		OriginSignature      string `gorm:"type:varchar(64) not null"`
		OriginByteSize       uint   `gorm:"type:bigint(1) unsigned not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_relID,priority:1"`
		AppID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID_relID,priority:2"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// TemplateVariables : 用于模版变量的管理
	type TemplateVariables struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		// Name查询时区分大小写，设置collate为utf8_bin
		Name       string `gorm:"type:varchar(255) collate utf8_bin not null;uniqueIndex:idx_bizID_name,priority:2"`
		Type       string `gorm:"type:varchar(20) not null"`
		DefaultVal string `gorm:"type:mediumtext"`
		Memo       string `gorm:"type:varchar(256) default ''"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_name,priority:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
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
		ReleaseID uint   `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_releaseID,priority:1"`
		Variables string `gorm:"type:json not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID,priority:1"`
		AppID uint `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID,priority:2"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").AutoMigrate(
		&AppTemplateBindings{},
		&ReleasedAppTemplates{},
		&TemplateVariables{},
		&AppTemplateVariables{},
		&ReleasedAppTemplateVariables{},
	); err != nil {
		return err
	}

	if result := tx.Create([]IDGenerators{
		{Resource: "app_template_bindings", MaxID: 0},
		{Resource: "released_app_templates", MaxID: 0},
		{Resource: "template_variables", MaxID: 0},
		{Resource: "app_template_variables", MaxID: 0},
		{Resource: "released_app_template_variables", MaxID: 0},
	}); result.Error != nil {
		return result.Error
	}

	return nil

}

// mig20230625152115Down for down migration
func mig20230625152115Down(tx *gorm.DB) error {
	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Migrator().DropTable(
		"app_template_bindings",
		"released_app_templates",
		"template_variables",
		"app_template_variables",
		"released_app_template_variables",
	); err != nil {
		return err
	}

	var resources = []string{
		"app_template_bindings",
		"released_app_templates",
		"template_variables",
		"app_template_variables",
		"released_app_template_variables",
	}
	if result := tx.Where("resource in ?", resources).Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	return nil
}
