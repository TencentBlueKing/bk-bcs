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
		Version: "20241008111330",
		Name:    "20241008111330_add_itsm_config",
		Mode:    migrator.GormMode,
		Up:      mig20241008111330Up,
		Down:    mig20241008111330Down,
	})
}

// ItsmConfigs : ItsmConfigs
type ItsmConfigs struct {
	ID             uint   `gorm:"column:id;type:bigint(1) unsigned;primary_key"`
	Key            string `gorm:"column:key;type:varchar(256);default:'';NOT NULL;index:idx_key,priority:1"`
	Value          uint   `gorm:"column:value;type:bigint(1) unsigned;default:0;NOT NULL"`
	WorkflowId     uint   `gorm:"column:workflow_id;type:bigint(1) unsigned;default:0;NOT NULL"`
	StateApproveId uint   `gorm:"column:state_approve_id;type:bigint(1) unsigned;default:0;NOT NULL;"`
}

// IDGenerators : ID生成器
type IDGenerators struct {
	ID        uint      `gorm:"type:bigint(1) unsigned not null;primaryKey"`
	Resource  string    `gorm:"type:varchar(50) not null;uniqueIndex:idx_resource"`
	MaxID     uint      `gorm:"type:bigint(1) unsigned not null"`
	UpdatedAt time.Time `gorm:"type:datetime(6) not null"`
}

// mig20241008111330Up for up migration
func mig20241008111330Up(tx *gorm.DB) error {

	if err := tx.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4").
		AutoMigrate(&ItsmConfigs{}); err != nil {
		return err
	}

	now := time.Now()

	if result := tx.Create([]IDGenerators{
		{Resource: "itsm_configs", MaxID: 0, UpdatedAt: now},
	}); result.Error != nil {
		return result.Error
	}

	return nil
}

// mig20241008111330Down for down migration
func mig20241008111330Down(tx *gorm.DB) error {
	if err := tx.Migrator().DropTable(ItsmConfigs{}); err != nil {
		return err
	}

	var resources = []string{
		"itsm_configs",
	}
	if result := tx.Where("resource IN ?", resources).Delete(&IDGenerators{}); result.Error != nil {
		return result.Error
	}

	return nil
}
