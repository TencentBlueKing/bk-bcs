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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrator"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20230617152513",
		Name:    "20230617152513_add_template",
		Mode:    migrator.GormMode,
		Up:      mig20230617152513Up,
		Down:    mig20230617152513Down,
	})
}

// mig20230617152513Up for up migration
func mig20230617152513Up(tx *gorm.DB) error { //nolint:funlen
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
		Name string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_tempSpaID_name_path,priority:3"`
		Path string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_tempSpaID_name_path,priority:4"`
		Memo string `gorm:"type:varchar(256) default ''"`

		// Attachment is attachment info of the resource
		BizID           uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_tempSpaID_name_path,priority:1"`
		TemplateSpaceID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_tempSpaID_name_path,priority:2"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// TemplateRevisions : 模版版本
	type TemplateRevisions struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		RevisionName string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_tempID_revName,priority:3"`
		RevisionMemo string `gorm:"type:varchar(256) default ''"`
		Name         string `gorm:"type:varchar(255) not null"`
		Path         string `gorm:"type:varchar(255) not null"`
		FileType     string `gorm:"type:varchar(20) not null"`
		FileMode     string `gorm:"type:varchar(20) not null"`
		User         string `gorm:"type:varchar(64) not null"`
		UserGroup    string `gorm:"type:varchar(64) not null"`
		Privilege    string `gorm:"type:varchar(64) not null"`
		Signature    string `gorm:"type:varchar(64) not null"`
		ByteSize     uint   `gorm:"type:bigint(1) unsigned not null"`

		// Attachment is attachment info of the resource
		BizID           uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_tempID_revName,priority:1"`
		TemplateSpaceID uint `gorm:"type:bigint(1) unsigned not null"`
		TemplateID      uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_tempID_revName,priority:2"`

		// CreatedRevision is reversion info of the resource being created
		Creator   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// TemplateSets : 模版套餐
	type TemplateSets struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name        string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_tempSpaID_name,priority:3"`
		Memo        string `gorm:"type:varchar(256) default ''"`
		TemplateIDs string `gorm:"type:json not null"`
		Public      bool   `gorm:"type:boolean default false"`
		BoundApps   string `gorm:"type:json not null"`

		// Attachment is attachment info of the resource
		BizID           uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_tempSpaID_name,priority:1"`
		TemplateSpaceID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_tempSpaID_name,priority:2"`

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
		AutoMigrate(&TemplateSpaces{}, &Templates{}, &TemplateRevisions{}, &TemplateSets{}); err != nil {
		return err
	}

	if result := tx.Create([]IDGenerators{
		{Resource: "template_spaces", MaxID: 0},
		{Resource: "templates", MaxID: 0},
		{Resource: "template_revisions", MaxID: 0},
		{Resource: "template_sets", MaxID: 0},
	}); result.Error != nil {
		return result.Error
	}

	return nil

}

// mig20230617152513Down for down migration
func mig20230617152513Down(tx *gorm.DB) error {
	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Migrator().
		DropTable("template_spaces", "templates", "template_revisions", "template_sets"); err != nil {
		return err
	}

	var resources = []string{
		"template_spaces",
		"templates",
		"template_revisions",
		"template_sets",
	}
	if result := tx.Where("resource in ?", resources).Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	return nil
}
