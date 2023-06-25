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
		Version: "20230617152513",
		Name:    "20230617152513_add_template",
		Mode:    migrator.GormMode,
		Up:      mig20230617152513GormUp,
		Down:    mig20230617152513GormDown,
	})
}

// mig20230617152513GormUp for up migration
func mig20230617152513GormUp(tx *gorm.DB) error {
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
		ReleaseName string `gorm:"type:varchar(255) not null;uniqueIndex:idx_tempID_relName,priority:2"`
		ReleaseMemo string `gorm:"type:varchar(256) default ''"`
		Name        string `gorm:"type:varchar(255) not null"`
		Path        string `gorm:"type:varchar(255) not null"`
		FileType    string `gorm:"type:varchar(20) not null"`
		FileMode    string `gorm:"type:varchar(20) not null"`
		User        string `gorm:"type:varchar(64) not null"`
		UserGroup   string `gorm:"type:varchar(64) not null"`
		Privilege   string `gorm:"type:varchar(64) not null"`
		Signature   string `gorm:"type:varchar(64) not null"`
		ByteSize    uint   `gorm:"type:bigint(1) unsigned not null"`

		// Attachment is attachment info of the resource
		BizID           uint `gorm:"type:bigint(1) unsigned not null"`
		TemplateSpaceID uint `gorm:"type:bigint(1) unsigned not null"`
		TemplateID      uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_tempID_relName,priority:1"`

		// CreatedRevision is reversion info of the resource being created
		Creator   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// TemplateSets : 模版套餐
	type TemplateSets struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name        string `gorm:"type:varchar(255) not null;uniqueIndex:idx_tempSpaID_name,priority:2"`
		Memo        string `gorm:"type:varchar(256) default ''"`
		TemplateIDs string `gorm:"type:json not null"`
		Public      bool   `gorm:"type:boolean default false"`
		BoundApps   string `gorm:"type:json not null"`

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

// mig20230617152513GormDown for down migration
func mig20230617152513GormDown(tx *gorm.DB) error {
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
