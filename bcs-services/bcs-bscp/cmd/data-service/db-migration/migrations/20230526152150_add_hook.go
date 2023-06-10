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
		Version: "mig20230526152150",
		Name:    "20230526152150_add_hook",
		Mode:    migrator.GormMode,
		Up:      mig20230526152150GormTestUp,
		Down:    mig20230526152150GormDown,
	})
}

func mig20230526152150GormTestUp(tx *gorm.DB) error {

	// Hook 脚本
	type Hook struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name       string `gorm:"type:varchar(255) not null;uniqueIndex:idx_bizID_name,priority:2"`
		Memo       string `gorm:"type:varchar(256) default ''"`
		PublishNum uint   `gorm:"type:bigint(1) unsigned not null"`
		Type       string `gorm:"type:varchar(64) not null"`
		Tag        string `gorm:"type:varchar(64) not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_name,priority:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// HookRelease 脚本版本
	type HookRelease struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name       string `gorm:"type:varchar(255) not null;;uniqueIndex:idx_bizID_releaseName,priority:2"`
		Memo       string `gorm:"type:varchar(256) default ''"`
		PublishNum uint   `gorm:"type:bigint(1) unsigned not null"`
		State      string `gorm:"type:varchar(64) not null"`
		Content    string `gorm:"type:longtext"`

		// Attachment is attachment info of the resource
		BizID  uint `gorm:"type:bigint(1) unsigned not null"`
		HookID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_releaseName,priority:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// ConfigHook : 配置脚本
	type ConfigHook struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		PreHookID         uint `gorm:"type:bigint(1) unsigned not null"`
		PreHookReleaseID  uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookID        uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookReleaseID uint `gorm:"type:bigint(1) unsigned not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null"`
		APPID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_AppID:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// Release mapped from table <releases>
	type Release struct {
		PreHookID         uint `gorm:"type:bigint(1) unsigned not null"`
		PreHookReleaseID  uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookID        uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookReleaseID uint `gorm:"type:bigint(1) unsigned not null"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&Hook{}, &HookRelease{}, &ConfigHook{}, &Release{}); err != nil {
		return err
	}

	now := time.Now()
	if result := tx.Create([]IDGenerators{
		{Resource: "hooks", MaxID: 0, UpdatedAt: now},
		{Resource: "hook_releases", MaxID: 0, UpdatedAt: now},
		{Resource: "config_hooks", MaxID: 0, UpdatedAt: now},
	}); result.Error != nil {
		return result.Error
	}

	return nil
}

func mig20230526152150GormDown(tx *gorm.DB) error {

	// Hook 脚本
	type Hook struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name       string `gorm:"type:varchar(64) not null;uniqueIndex:idx_bizID_name,priority:2"`
		Meme       string `gorm:"type:varchar(64) not null"`
		PublishNum uint   `gorm:"type:bigint(1) unsigned not null"`
		Type       string `gorm:"type:varchar(64) not null"`
		Tag        string `gorm:"type:varchar(64) not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_name,priority:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// HookRelease 脚本版本
	type HookRelease struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		Name       string `gorm:"type:varchar(64) not null;uniqueIndex:idx_bizID_releaseName,priority:2"`
		Meme       string `gorm:"type:varchar(64) not null"`
		PublishNum uint   `gorm:"type:bigint(1) unsigned not null"`
		State      string `gorm:"type:varchar(64) not null"`
		Content    string `gorm:"type:longtext"`

		// Attachment is attachment info of the resource
		BizID  uint `gorm:"type:bigint(1) unsigned not null"`
		HookID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_bizID_releaseName,priority:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// ConfigHook : 配置脚本
	type ConfigHook struct {
		ID uint `gorm:"type:bigint(1) unsigned not null;primaryKey;autoIncrement:false"`

		// Spec is specifics of the resource defined with user
		PreHookID         uint `gorm:"type:bigint(1) unsigned not null"`
		PreHookReleaseID  uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookID        uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookReleaseID uint `gorm:"type:bigint(1) unsigned not null"`

		// Attachment is attachment info of the resource
		BizID uint `gorm:"type:bigint(1) unsigned not null"`
		APPID uint `gorm:"type:bigint(1) unsigned not null;uniqueIndex:idx_AppID:1"`

		// Revision is revision info of the resource
		Creator   string    `gorm:"type:varchar(64) not null"`
		Reviser   string    `gorm:"type:varchar(64) not null"`
		CreatedAt time.Time `gorm:"type:datetime(6) not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	// Release mapped from table <releases>
	type Release struct {
		PreHookID         uint `gorm:"type:bigint(1) unsigned not null"`
		PreHookReleaseID  uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookID        uint `gorm:"type:bigint(1) unsigned not null"`
		PostHookReleaseID uint `gorm:"type:bigint(1) unsigned not null"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Migrator().DropTable(ConfigHook{}); err != nil {
		return err
	}

	if err := tx.Migrator().DropColumn(Release{}, "pre_hook_id"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(Release{}, "pre_hook_release_id"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(Release{}, "post_hook_id"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(Release{}, "post_hook_release_id"); err != nil {
		return err
	}

	if err := tx.Migrator().
		DropTable("hooks", "hook_releases", "config_hooks"); err != nil {
		return err
	}

	var resources = []string{
		"hooks",
		"hook_releases",
		"config_hooks",
	}
	if result := tx.Where("resource IN ?", resources).Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	return nil
}
