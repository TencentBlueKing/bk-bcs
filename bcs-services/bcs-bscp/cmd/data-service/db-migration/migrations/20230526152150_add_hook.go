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
		Version: "20230526152150",
		Name:    "20230526152150_add_hook",
		Mode:    migrator.GormMode,
		Up:      mig20230526152150Up,
		Down:    mig20230526152150Down,
	})
}

// mig20230526152150Up for up migration
func mig20230526152150Up(tx *gorm.DB) error {

	// Hook 脚本
	type Hook struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_name,priority:2"`
		Memo string `gorm:"type:varchar(256) default ''"`
		Type string `gorm:"type:varchar(64) not null"`
		Tag  string `gorm:"type:varchar(64) not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_name,priority:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// HookRevision 脚本版本
	type HookRevision struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name    string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_revisionName,priority:2"`
		Memo    string `gorm:"type:varchar(256) default ''"`
		State   string `gorm:"type:varchar(64) not null"`
		Content string `gorm:"type:longtext"`

		// Attachment is attachment info of the resource
		BizID  uint `gorm:"type:bigint(1) unsigned not null"`
		HookID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_revisionName,priority:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// ReleasedHook : 已随配置项版本发布的配置脚本
	type ReleasedHook struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		APPID     uint   `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_appID_releaseID_hookType,priority:1"`
		ReleaseID uint   `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_appID_releaseID_hookType,priority:2"`
		HookType  string `gorm:"type:varchar(64) not null;uniqueIndex:idx_appID_releaseID_hookType,priority:3"`

		HookID           uint      `gorm:"type:bigint(1) unsigned not null"`
		HookRevisionID   uint      `gorm:"type:bigint(1) unsigned not null"`
		HookName         string    `gorm:"type:varchar(64) not null"`
		HookRevisionName string    `gorm:"type:varchar(64) not null"`
		Content          string    `gorm:"type:longtext"`
		ScriptType       string    `gorm:"type:varchar(64) not null"`
		BizID            uint      `gorm:"type:bigint(1) unsigned not null"`
		Reviser          string    `gorm:"type:varchar(64) not null"`
		UpdatedAt        time.Time `gorm:"type:datetime(6) not null"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&Hook{}, &HookRevision{}, &ReleasedHook{}); err != nil {
		return err
	}

	now := time.Now()
	if result := tx.Create([]IDGenerators{
		{Resource: "hooks", MaxID: 0, UpdatedAt: now},
		{Resource: "hook_revisions", MaxID: 0, UpdatedAt: now},
		{Resource: "released_hooks", MaxID: 0, UpdatedAt: now},
	}); result.Error != nil {
		return result.Error
	}

	return nil
}

// mig20230526152150Down for down migration
func mig20230526152150Down(tx *gorm.DB) error {

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Migrator().
		DropTable("hooks", "hook_revisions", "released_hooks"); err != nil {
		return err
	}

	var resources = []string{
		"hooks",
		"hook_revisions",
		"released_hooks",
	}
	if result := tx.Where("resource IN ?", resources).Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	return nil
}
