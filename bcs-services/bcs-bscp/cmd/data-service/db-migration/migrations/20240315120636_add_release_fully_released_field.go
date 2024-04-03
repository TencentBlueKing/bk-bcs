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
		Version: "20240315120636",
		Name:    "20240315120636_add_release_fully_released_field",
		Mode:    migrator.GormMode,
		Up:      mig20240315120636Up,
		Down:    mig20240315120636Down,
	})
}

// **这里只需要DB操作用到的字段**

// Release20240315120636 版本
type Release20240315120636 struct {
	ID    uint32 `gorm:"primaryKey"`
	BizID uint32 `gorm:"column:biz_id"`
	// FullyReleased is a flag to indicate whether the release was once fully released.
	FullyReleased bool `gorm:"column:fully_released"`
}

// TableName 表名
func (Release20240315120636) TableName() string {
	return "releases"
}

// Strategy20240315120636 发布策略
type Strategy20240315120636 struct {
	ID        uint32 `gorm:"primaryKey"`
	BizID     uint32 `gorm:"column:biz_id"`
	ReleaseID uint32 `gorm:"column:release_id"`
}

// mig20240315120636Up for up migration
func mig20240315120636Up(tx *gorm.DB) error {

	/*** 字段变更 ***/
	if !tx.Migrator().HasColumn(&Release20240315120636{}, "FullyReleased") {
		if err := tx.Migrator().AddColumn(&Release20240315120636{}, "FullyReleased"); err != nil {
			return err
		}
	}

	// 所有发布策略中，如果满足以下两个条件之一
	// 1. 包含默认分组: strategies.as_default = true
	// 2. 全量发布: JSON_LENGTH(strategies.scope->'$.groups') = 0
	// 则将发布策略对应的版本标记为全量发布过
	// gorm 限制了 UPDATE 必须有 WHERE 条件，所以这里使用了 Where("1 = 1")
	return tx.Model(&Release20240315120636{}).Where("1 = 1").UpdateColumn("fully_released",
		gorm.Expr("EXISTS(SELECT 1 FROM strategies WHERE strategies.release_id = releases.id"+
			" AND (strategies.as_default = 1 OR JSON_LENGTH(strategies.scope->'$.groups') = 0))")).Error
}

// mig20240315120636Down for down migration
func mig20240315120636Down(tx *gorm.DB) error {

	/*** 字段变更 ***/
	if tx.Migrator().HasColumn(&Release20240315120636{}, "FullyReleased") {
		if err := tx.Migrator().DropColumn(&Release20240315120636{}, "FullyReleased"); err != nil {
			return err
		}
	}
	return nil
}
