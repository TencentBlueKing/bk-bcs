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
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20230511114513",
		Name:    "20230511114513_add_template",
		Mode:    migrator.GormMode,
		Up:      mig20230511114513GormTestUp,
		Down:    mig20230511114513GormDown,
	})
}

func mig20230511114513GormTestUp(tx *gorm.DB) error {
	// TemplateSpaces ：模版空间
	type TemplateSpaces struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_name,priority:2"`
		Memo string `gorm:"type:varchar(256) default ''"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_name,priority:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// Templates : 配置模版
	type Templates struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name string `gorm:"type:varchar(255) not null;uniqueIndex:idx_tempSpaID_name_path,priority:2"`
		Path string `gorm:"type:varchar(255) not null;uniqueIndex:idx_tempSpaID_name_path,priority:3"`
		Memo string `gorm:"type:varchar(256) default ''"`

		// Attachment is attachment info of the resource
		BizID           uint `gorm:"type:bigint(1) unsigned not null"`
		TemplateSpaceID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_tempSpaID_name_path,priority:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// TemplateReleases : 模版版本
	type TemplateReleases struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		TemplateName string `gorm:"type:varchar(255) not null"`
		TemplatePath string `gorm:"type:varchar(255) not null"`
		Name         string `gorm:"type:varchar(255) not null;uniqueIndex:idx_tempID_name,priority:2"`
		Memo         string `gorm:"type:varchar(256) default ''"`

		// Attachment is attachment info of the resource
		BizID           uint `gorm:"type:bigint(1) unsigned not null"`
		TemplateSpaceID uint `gorm:"type:bigint(1) unsigned not null"`
		TemplateID      uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_tempID_name,priority:1"`

		// CreatedRevision is reversion info of the resource being created
		Creator   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// TemplateSets : 模版套餐
	type TemplateSets struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name               string `gorm:"type:varchar(255) not null;uniqueIndex:idx_tempSpaID_name,priority:2"`
		Memo               string `gorm:"type:varchar(256) default ''"`
		TemplateReleaseIDs string `gorm:"type:json not null"`

		// Attachment is attachment info of the resource
		BizID           uint `gorm:"type:bigint(1) unsigned not null"`
		TemplateSpaceID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_tempSpaID_name,priority:1"`

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
		AutoMigrate(&TemplateSpaces{}, &Templates{}, &TemplateReleases{}, &TemplateSets{}); err != nil {
		return err
	}

	now := time.Now()
	if result := tx.Create([]IDGenerators{
		{Resource: "template_spaces", MaxID: 0, UpdatedAt: now},
		{Resource: "templates", MaxID: 0, UpdatedAt: now},
		{Resource: "template_releases", MaxID: 0, UpdatedAt: now},
		{Resource: "template_sets", MaxID: 0, UpdatedAt: now},
	}); result.Error != nil {
		return result.Error
	}

	return nil

}

func mig20230511114513GormDown(tx *gorm.DB) error {
	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Migrator().
		DropTable("template_spaces", "templates", "template_releases", "template_sets"); err != nil {
		return err
	}

	var resources = []string{
		"template_spaces",
		"templates",
		"template_releases",
		"template_sets",
	}
	if result := tx.Where("resource in ?", resources).Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	return nil
}
