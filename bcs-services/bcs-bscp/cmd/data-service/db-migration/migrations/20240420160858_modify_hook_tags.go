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
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrator"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20240420160858",
		Name:    "20240420160858_modify_hook_tags",
		Mode:    migrator.GormMode,
		Up:      mig20240420160858Up,
		Down:    mig20240420160858Down,
	})
}

// Hook20240420160858 脚本
type Hook20240420160858 struct {
	ID   uint32 `gorm:"primaryKey"`
	Tag  string `gorm:"type:varchar(64) not null"`
	Tags string `gorm:"type:json not null"`
}

// TableName return table name
func (Hook20240420160858) TableName() string {
	return "hooks"
}

// mig20240420160858Up for up migration
func mig20240420160858Up(tx *gorm.DB) error {
	/*** 字段变更 ***/
	if !tx.Migrator().HasColumn(&Hook20240420160858{}, "Tags") {
		if err := tx.Migrator().AddColumn(&Hook20240420160858{}, "Tags"); err != nil {
			return err
		}
	}

	if err := migrateFromTagToTags(tx); err != nil {
		return err
	}

	if tx.Migrator().HasColumn(&Hook20240420160858{}, "Tag") {
		if err := tx.Migrator().DropColumn(&Hook20240420160858{}, "Tag"); err != nil {
			return err
		}
	}

	return nil
}

// mig20240420160858Down for down migration
func mig20240420160858Down(tx *gorm.DB) error {
	/*** 字段变更 ***/
	if !tx.Migrator().HasColumn(&Hook20240420160858{}, "Tag") {
		if err := tx.Migrator().AddColumn(&Hook20240420160858{}, "Tag"); err != nil {
			return err
		}
	}

	if err := migrateFromTagsToTag(tx); err != nil {
		return err
	}

	if tx.Migrator().HasColumn(&Hook20240420160858{}, "Tags") {
		if err := tx.Migrator().DropColumn(&Hook20240420160858{}, "Tags"); err != nil {
			return err
		}
	}

	return nil
}

// migrateFromTagToTags migrate data from field tag to tags
func migrateFromTagToTags(tx *gorm.DB) error {
	var hooks []Hook20240420160858
	if err := tx.Model(&Hook20240420160858{}).Find(&hooks).Error; err != nil {
		return err
	}
	for _, h := range hooks {
		h.Tags = "[]"
		if h.Tag != "" {
			// 转为golang中[]string类型的json串
			h.Tags = "[\"" + h.Tag + "\"]"
		}
		// 只更新必须的字段, 不刷新updated_at
		if err := tx.Select("Tags").UpdateColumns(&h).Error; err != nil {
			fmt.Printf("update tags for hook %#v failed, err: %s\n", h, err)
		}
	}
	return nil
}

// migrateFromTagsToTag migrate data from field tags to tag
func migrateFromTagsToTag(tx *gorm.DB) error {
	var hooks []Hook20240420160858
	if err := tx.Model(&Hook20240420160858{}).Find(&hooks).Error; err != nil {
		return err
	}
	for _, h := range hooks {
		if h.Tags == "[]" {
			continue
		}
		var t []string
		if err := json.Unmarshal([]byte(h.Tags), &t); err != nil {
			fmt.Printf("unmarshal hook tags %s failed, err: %s\n", h.Tags, err)
			continue
		}
		h.Tag = t[0]
		// 只更新必须的字段, 不刷新updated_at
		if err := tx.Select("Tag").UpdateColumns(&h).Error; err != nil {
			fmt.Printf("update tags for hook %#v failed, err: %s\n", h, err)
		}
	}
	return nil
}
