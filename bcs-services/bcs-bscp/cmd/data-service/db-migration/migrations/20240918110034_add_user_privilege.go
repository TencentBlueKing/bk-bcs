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
		Version: "20240918110034",
		Name:    "20240918110034_add_user_privilege",
		Mode:    migrator.GormMode,
		Up:      mig20240918110034Up,
		Down:    mig20240918110034Down,
	})
}

// mig20240918110034Up for up migration
// nolint:funlen
func mig20240918110034Up(tx *gorm.DB) error {
	// UserPrivileges : user privileges
	type UserPrivileges struct {
		ID              uint      `gorm:"column:id;type:bigint(1) unsigned;primary_key"`
		BizID           uint      `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID,priority:1"`
		AppID           uint      `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID,priority:2"`
		Uid             uint      `gorm:"column:uid;type:bigint(1) unsigned;default:0;NOT NULL"`
		TemplateSpaceID uint      `gorm:"column:template_space_id;type:bigint(1) unsigned;default:0;NOT NULL"`
		User            string    `gorm:"column:user;type:varchar(64);NOT NULL"`
		PrivilegeType   string    `gorm:"column:privilege_type;type:varchar(20);NOT NULL"`
		ReadOnly        uint      `gorm:"column:read_only;type:tinyint(1) unsigned;default:0;NOT NULL"`
		Creator         string    `gorm:"column:creator;type:varchar(64);NOT NULL"`
		Reviser         string    `gorm:"column:reviser;type:varchar(64);NOT NULL"`
		CreatedAt       time.Time `gorm:"column:created_at;type:datetime(6);NOT NULL"`
		UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime(6);NOT NULL"`
	}

	// UserGroupPrivileges : user group privileges
	type UserGroupPrivileges struct {
		ID              uint      `gorm:"column:id;type:bigint(1) unsigned;primary_key"`
		BizID           uint      `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID,priority:1"`
		AppID           uint      `gorm:"type:bigint(1) unsigned not null;index:idx_bizID_appID,priority:2"`
		Gid             uint      `gorm:"column:gid;type:bigint(1) unsigned;default:0;NOT NULL"`
		TemplateSpaceID uint      `gorm:"column:template_space_id;type:bigint(1) unsigned;default:0;NOT NULL"`
		UserGroup       string    `gorm:"column:user_group;type:varchar(64);NOT NULL"`
		PrivilegeType   string    `gorm:"column:privilege_type;type:varchar(20);NOT NULL"`
		ReadOnly        uint      `gorm:"column:read_only;type:tinyint(1) unsigned;default:0;NOT NULL"`
		Creator         string    `gorm:"column:creator;type:varchar(64);NOT NULL"`
		Reviser         string    `gorm:"column:reviser;type:varchar(64);NOT NULL"`
		CreatedAt       time.Time `gorm:"column:created_at;type:datetime(6);NOT NULL"`
		UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime(6);NOT NULL"`
	}

	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&UserPrivileges{}, &UserGroupPrivileges{}); err != nil {
		return err
	}

	now := time.Now()

	// 新增默认数据
	tx.Create([]UserPrivileges{
		{
			ID:              1,
			BizID:           0,
			AppID:           0,
			TemplateSpaceID: 0,
			Uid:             0,
			User:            "root",
			PrivilegeType:   "system",
			ReadOnly:        1,
			Creator:         "system",
			Reviser:         "",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	})

	// 新增默认数据
	tx.Create([]UserGroupPrivileges{
		{
			ID:              1,
			BizID:           0,
			AppID:           0,
			TemplateSpaceID: 0,
			Gid:             0,
			UserGroup:       "root",
			PrivilegeType:   "system",
			ReadOnly:        1,
			Creator:         "system",
			Reviser:         "",
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	})

	if result := tx.Create([]IDGenerators{
		{Resource: "user_privileges", MaxID: 1, UpdatedAt: now},
		{Resource: "user_group_privileges", MaxID: 1, UpdatedAt: now},
	}); result.Error != nil {
		return result.Error
	}

	// ReleasedAppTemplates  : released_app_templates
	type ReleasedAppTemplates struct {
		Uid uint64 `gorm:"column:uid;type:bigint(1) unsigned;default:0;NOT NULL"`
		Gid uint64 `gorm:"column:gid;type:bigint(1) unsigned;default:0;NOT NULL"`
	}

	if !tx.Migrator().HasColumn(&ReleasedAppTemplates{}, "uid") {
		if err := tx.Migrator().AddColumn(&ReleasedAppTemplates{}, "uid"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&ReleasedAppTemplates{}, "gid") {
		if err := tx.Migrator().AddColumn(&ReleasedAppTemplates{}, "gid"); err != nil {
			return err
		}
	}

	// ReleasedConfigItems  : released_config_items
	type ReleasedConfigItems struct {
		Uid uint64 `gorm:"column:uid;type:bigint(1) unsigned;default:0;NOT NULL"`
		Gid uint64 `gorm:"column:gid;type:bigint(1) unsigned;default:0;NOT NULL"`
	}

	if !tx.Migrator().HasColumn(&ReleasedConfigItems{}, "uid") {
		if err := tx.Migrator().AddColumn(&ReleasedConfigItems{}, "uid"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&ReleasedConfigItems{}, "gid") {
		if err := tx.Migrator().AddColumn(&ReleasedConfigItems{}, "gid"); err != nil {
			return err
		}
	}

	return nil
}

// mig20240918110034Down for down migration
func mig20240918110034Down(tx *gorm.DB) error {
	// IDGenerators : ID生成器
	type IDGenerators struct {
		ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
		Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
		MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
		UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
	}

	var resources = []string{
		"user_privileges",
		"user_group_privileges",
	}
	if result := tx.Where("resource IN ?", resources).Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	if err := tx.Migrator().DropTable("user_privileges"); err != nil {
		return err
	}

	if err := tx.Migrator().DropTable("user_group_privileges"); err != nil {
		return err
	}

	// ReleasedAppTemplates  : released_app_templates
	type ReleasedAppTemplates struct {
		Uid uint64 `gorm:"column:uid;type:bigint(1) unsigned;default:0;NOT NULL"`
		Gid uint64 `gorm:"column:gid;type:bigint(1) unsigned;default:0;NOT NULL"`
	}

	if tx.Migrator().HasColumn(&ReleasedAppTemplates{}, "uid") {
		if err := tx.Migrator().DropColumn(&ReleasedAppTemplates{}, "uid"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&ReleasedAppTemplates{}, "gid") {
		if err := tx.Migrator().DropColumn(&ReleasedAppTemplates{}, "gid"); err != nil {
			return err
		}
	}

	// ReleasedConfigItems  : released_config_items
	type ReleasedConfigItems struct {
		Uid uint64 `gorm:"column:uid;type:bigint(1) unsigned;default:0;NOT NULL"`
		Gid uint64 `gorm:"column:gid;type:bigint(1) unsigned;default:0;NOT NULL"`
	}

	if tx.Migrator().HasColumn(&ReleasedConfigItems{}, "uid") {
		if err := tx.Migrator().DropColumn(&ReleasedConfigItems{}, "uid"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&ReleasedConfigItems{}, "gid") {
		if err := tx.Migrator().DropColumn(&ReleasedConfigItems{}, "gid"); err != nil {
			return err
		}
	}

	return nil
}
