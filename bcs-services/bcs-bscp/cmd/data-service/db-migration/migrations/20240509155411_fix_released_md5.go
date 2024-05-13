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
	"fmt"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/db-migration/migrator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

func init() {
	// add current migration to migrator
	migrator.GetMigrator().AddMigration(&migrator.Migration{
		Version: "20240509155411",
		Name:    "20240509155411_add_content_md5",
		Mode:    migrator.GormMode,
		Up:      mig20240509155411Up,
		Down:    mig20240509155411Down,
	})
}

// **这里只需要DB操作用到的字段**

// ReleasedConfigItem20240509155411 已生成版本的配置项
type ReleasedConfigItem20240509155411 struct {
	ID        uint32 `gorm:"primaryKey"`
	BizID     uint32 `gorm:"column:biz_id"`
	Signature string `gorm:"type:varchar(64) not null"`
	Md5       string `gorm:"type:varchar(64) not null"`
}

// TableName return table name
func (ReleasedConfigItem20240509155411) TableName() string {
	return "released_config_items"
}

// ReleasedAppTemplate20240509155411 已生成版本服务的模版
// 这里只需要DB操作用到的字段
type ReleasedAppTemplate20240509155411 struct {
	ID        uint32 `gorm:"primaryKey"`
	BizID     uint32 `gorm:"column:biz_id"`
	Signature string `gorm:"type:varchar(64) not null"`
	Md5       string `gorm:"type:varchar(64) not null"`
}

// TableName return table name
func (ReleasedAppTemplate20240509155411) TableName() string {
	return "released_app_templates"
}

// mig20240509155411Up for up migration
func mig20240509155411Up(tx *gorm.DB) error {

	kt := kit.New()

	provider, err := repository.NewProvider(cc.DataService().Repo)
	if err != nil {
		return err
	}
	md5Map := map[string]string{}

	if err := batchUpdateReleasedConfigItemMd520240509155411(kt, tx, provider, md5Map); err != nil {
		return err
	}

	if err := batchUpdateReleasedAppTemplateMd520240509155411(kt, tx, provider, md5Map); err != nil {
		return err
	}

	return nil

}

// mig20240509155411Down for down migration
func mig20240509155411Down(tx *gorm.DB) error {
	return nil
}

func batchUpdateReleasedConfigItemMd520240509155411(kt *kit.Kit, tx *gorm.DB, provider repository.Provider,
	md5Map map[string]string) error {
	var currentMaxID uint32
	releasedCIs := []*ReleasedConfigItem20240509155411{}
	if err := tx.Model(&ReleasedConfigItem20240509155411{}).Select("COALESCE(MAX(id), 0)").Row().
		Scan(&currentMaxID); err != nil {
		return err
	}
	if err := tx.Model(&ReleasedConfigItem20240509155411{}).Where("id <= ?", currentMaxID).
		Find(&releasedCIs).Error; err != nil {
		return err
	}

	successCount := 0
	failedIDs := []uint32{}
	for _, releasedCI := range releasedCIs {
		kt.BizID = releasedCI.BizID
		if releasedCI.Md5 != "" {
			continue
		}
		var md5 string
		if md5Map[releasedCI.Signature] != "" {
			md5 = md5Map[releasedCI.Signature]
		} else {
			metadata, err := provider.Metadata(kt, releasedCI.Signature)
			if err != nil {
				fmt.Printf("get metadata for released_config_item %s failed, err: %s\n",
					releasedCI.Signature, err.Error())
				failedIDs = append(failedIDs, releasedCI.ID)
				continue
			}
			md5 = metadata.Md5
			md5Map[releasedCI.Signature] = md5
		}
		if err := tx.Model(&ReleasedConfigItem20240509155411{}).Where("id = ?", releasedCI.ID).
			Update("md5", md5).Error; err != nil {
			fmt.Printf("update released_config_item %d md5 failed, err: %s\n", releasedCI.ID, err.Error())
			failedIDs = append(failedIDs, releasedCI.ID)
			continue
		}
		successCount++
	}
	fmt.Printf("batch update released_config_itemt md5 success count: %d, failed count: %d\n",
		successCount, len(failedIDs))
	if len(failedIDs) > 0 {
		fmt.Printf("failed released_config_item ids: %v\n", failedIDs)
	}
	return nil
}

func batchUpdateReleasedAppTemplateMd520240509155411(kt *kit.Kit, tx *gorm.DB, provider repository.Provider,
	md5Map map[string]string) error {
	var currentMaxID uint32
	releasedATs := []*ReleasedAppTemplate20240509155411{}
	if err := tx.Model(&ReleasedAppTemplate20240509155411{}).Select("COALESCE(MAX(id), 0)").Row().
		Scan(&currentMaxID); err != nil {
		return err
	}
	if err := tx.Model(&ReleasedAppTemplate20240509155411{}).Where("id <= ?", currentMaxID).
		Find(&releasedATs).Error; err != nil {
		return err
	}

	successCount := 0
	failedIDs := []uint32{}
	for _, releasedAT := range releasedATs {
		kt.BizID = releasedAT.BizID
		var md5 string
		if md5Map[releasedAT.Signature] != "" {
			md5 = md5Map[releasedAT.Signature]
		} else {
			metadata, err := provider.Metadata(kt, releasedAT.Signature)
			if err != nil {
				fmt.Printf("get metadata for released_app_template %s failed, err: %s\n",
					releasedAT.Signature, err.Error())
				failedIDs = append(failedIDs, releasedAT.ID)
				continue
			}
			md5 = metadata.Md5
			md5Map[releasedAT.Signature] = md5
		}
		if releasedAT.Md5 == md5 {
			continue
		}
		if err := tx.Model(&ReleasedAppTemplate20240509155411{}).Where("id = ?", releasedAT.ID).
			Update("md5", md5).Error; err != nil {
			fmt.Printf("update released_app_template %d md5 failed, err: %s\n", releasedAT.ID, err.Error())
			failedIDs = append(failedIDs, releasedAT.ID)
			continue
		}
		successCount++
	}
	fmt.Printf("batch update released_app_template md5 success count: %d, failed count: %d\n",
		successCount, len(failedIDs))
	if len(failedIDs) > 0 {
		fmt.Printf("failed released_app_template ids: %v\n", failedIDs)
	}
	return nil
}
