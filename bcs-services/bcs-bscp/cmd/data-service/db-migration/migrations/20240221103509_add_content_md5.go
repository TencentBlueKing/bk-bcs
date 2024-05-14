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
		Version: "20240221103509",
		Name:    "20240221103509_add_content_md5",
		Mode:    migrator.GormMode,
		Up:      mig20240221103509Up,
		Down:    mig20240221103509Down,
	})
}

// **这里只需要DB操作用到的字段**

// Content20240221103509 文件内容
type Content20240221103509 struct {
	ID        uint32 `gorm:"primaryKey"`
	BizID     uint32 `gorm:"column:biz_id"`
	Signature string `gorm:"type:varchar(64) not null"`
	// Md5 is the md5 value of a configuration file's content.
	// it can not be updated.
	Md5 string `gorm:"type:varchar(64) not null"`
}

// TableName return table name
func (Content20240221103509) TableName() string {
	return "contents"
}

// Commit20240221103509 文件修改记录
type Commit20240221103509 struct {
	ID        uint32 `gorm:"primaryKey"`
	BizID     uint32 `gorm:"column:biz_id"`
	Signature string `gorm:"type:varchar(64) not null"`
	Md5       string `gorm:"type:varchar(64) not null"`
}

// TableName return table name
func (Commit20240221103509) TableName() string {
	return "commits"
}

// ReleasedConfigItem20240221103509 已生成版本的配置项
type ReleasedConfigItem20240221103509 struct {
	ID        uint32 `gorm:"primaryKey"`
	BizID     uint32 `gorm:"column:biz_id"`
	Signature string `gorm:"type:varchar(64) not null"`
	Md5       string `gorm:"type:varchar(64) not null"`
}

// TableName return table name
func (ReleasedConfigItem20240221103509) TableName() string {
	return "released_config_items"
}

// ReleasedAppTemplate20240221103509 已生成版本服务的模版
// 这里只需要DB操作用到的字段
type ReleasedAppTemplate20240221103509 struct {
	ID        uint32 `gorm:"primaryKey"`
	BizID     uint32 `gorm:"column:biz_id"`
	Signature string `gorm:"type:varchar(64) not null"`
	Md5       string `gorm:"type:varchar(64) not null"`
}

// TableName return table name
func (ReleasedAppTemplate20240221103509) TableName() string {
	return "released_app_templates"
}

// TemplateRevision20240221103509 模版版本
type TemplateRevision20240221103509 struct {
	ID        uint32 `gorm:"primaryKey"`
	BizID     uint32 `gorm:"column:biz_id"`
	Signature string `gorm:"type:varchar(64) not null"`
	Md5       string `gorm:"type:varchar(64) not null"`
}

// TableName return table name
func (TemplateRevision20240221103509) TableName() string {
	return "template_revisions"
}

// mig20240221103509Up for up migration
func mig20240221103509Up(tx *gorm.DB) error {

	/*** 字段变更 ***/
	if !tx.Migrator().HasColumn(&Content20240221103509{}, "Md5") {
		if err := tx.Migrator().AddColumn(&Content20240221103509{}, "Md5"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&Commit20240221103509{}, "Md5") {
		if err := tx.Migrator().AddColumn(&Commit20240221103509{}, "Md5"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&ReleasedConfigItem20240221103509{}, "Md5") {
		if err := tx.Migrator().AddColumn(&ReleasedConfigItem20240221103509{}, "Md5"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&ReleasedAppTemplate20240221103509{}, "Md5") {
		if err := tx.Migrator().AddColumn(&ReleasedAppTemplate20240221103509{}, "Md5"); err != nil {
			return err
		}
	}

	if !tx.Migrator().HasColumn(&TemplateRevision20240221103509{}, "Md5") {
		if err := tx.Migrator().AddColumn(&TemplateRevision20240221103509{}, "Md5"); err != nil {
			return err
		}
	}

	kt := kit.New()

	provider, err := repository.NewProvider(cc.DataService().Repo)
	if err != nil {
		return err
	}
	md5Map := map[string]string{}

	if err := batchUpdateContentMd5(kt, tx, provider, md5Map); err != nil {
		return err
	}

	if err := batchUpdateCommitMd5(kt, tx, provider, md5Map); err != nil {
		return err
	}

	if err := batchUpdateReleasedConfigItemMd5(kt, tx, provider, md5Map); err != nil {
		return err
	}

	if err := batchUpdateReleasedAppTemplateMd5(kt, tx, provider, md5Map); err != nil {
		return err
	}

	if err := batchUpdateTemplateRevisionMd5(kt, tx, provider, md5Map); err != nil {
		return err
	}

	return nil

}

// mig20240221103509Down for down migration
func mig20240221103509Down(tx *gorm.DB) error {

	/*** 字段变更 ***/

	if tx.Migrator().HasColumn(&Content20240221103509{}, "Md5") {
		if err := tx.Migrator().DropColumn(&Content20240221103509{}, "Md5"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&Commit20240221103509{}, "Md5") {
		if err := tx.Migrator().DropColumn(&Commit20240221103509{}, "Md5"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&ReleasedConfigItem20240221103509{}, "Md5") {
		if err := tx.Migrator().DropColumn(&ReleasedConfigItem20240221103509{}, "Md5"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&ReleasedAppTemplate20240221103509{}, "Md5") {
		if err := tx.Migrator().DropColumn(&ReleasedAppTemplate20240221103509{}, "Md5"); err != nil {
			return err
		}
	}

	if tx.Migrator().HasColumn(&TemplateRevision20240221103509{}, "Md5") {
		if err := tx.Migrator().DropColumn(&TemplateRevision20240221103509{}, "Md5"); err != nil {
			return err
		}
	}

	return nil
}

func batchUpdateContentMd5(kt *kit.Kit, tx *gorm.DB, provider repository.Provider, md5Map map[string]string) error {
	var currentMaxID uint32
	contents := []*Content20240221103509{}
	if err := tx.Model(&Content20240221103509{}).Select("COALESCE(MAX(id), 0)").Row().Scan(&currentMaxID); err != nil {
		return err
	}
	if err := tx.Model(&Content20240221103509{}).Where("id <= ?", currentMaxID).Find(&contents).Error; err != nil {
		return err
	}

	successCount := 0
	failedIDs := []uint32{}
	for _, content := range contents {
		kt.BizID = content.BizID
		if content.Md5 != "" {
			continue
		}
		var md5 string
		if md5Map[content.Signature] != "" {
			md5 = md5Map[content.Signature]
		} else {
			metadata, err := provider.Metadata(kt, content.Signature)
			if err != nil {
				fmt.Printf("get metadata for content %s failed, err: %s\n", content.Signature, err.Error())
				failedIDs = append(failedIDs, content.ID)
				continue
			}
			md5 = metadata.Md5
			md5Map[content.Signature] = md5
		}
		if err := tx.Model(&Content20240221103509{}).Where("id = ?", content.ID).Update("md5", md5).Error; err != nil {
			fmt.Printf("update content %d md5 failed, err: %s\n", content.ID, err.Error())
			failedIDs = append(failedIDs, content.ID)
			continue
		}
		successCount++
	}
	fmt.Printf("batch update content md5 success count: %d, failed count: %d\n", successCount, len(failedIDs))
	if len(failedIDs) > 0 {
		fmt.Printf("failed content ids: %v\n", failedIDs)
	}
	return nil
}

func batchUpdateCommitMd5(kt *kit.Kit, tx *gorm.DB, provider repository.Provider, md5Map map[string]string) error {
	var currentMaxID uint32
	commits := []*Commit20240221103509{}
	if err := tx.Model(&Commit20240221103509{}).Select("COALESCE(MAX(id), 0)").Row().Scan(&currentMaxID); err != nil {
		return err
	}
	if err := tx.Model(&Commit20240221103509{}).Where("id <= ?", currentMaxID).Find(&commits).Error; err != nil {
		return err
	}

	successCount := 0
	failedIDs := []uint32{}
	for _, commit := range commits {
		kt.BizID = commit.BizID
		if commit.Md5 != "" {
			continue
		}
		var md5 string
		if md5Map[commit.Signature] != "" {
			md5 = md5Map[commit.Signature]
		} else {
			metadata, err := provider.Metadata(kt, commit.Signature)
			if err != nil {
				fmt.Printf("get metadata for commit %s failed, err: %s\n", commit.Signature, err.Error())
				failedIDs = append(failedIDs, commit.ID)
				continue
			}
			md5 = metadata.Md5
			md5Map[commit.Signature] = md5
		}
		if err := tx.Model(&Commit20240221103509{}).Where("id = ?", commit.ID).Update("md5", md5).Error; err != nil {
			fmt.Printf("update commit %d md5 failed, err: %s\n", commit.ID, err.Error())
			failedIDs = append(failedIDs, commit.ID)
			continue
		}
		successCount++
	}
	fmt.Printf("batch update commit md5 success count: %d, failed count: %d\n", successCount, len(failedIDs))
	if len(failedIDs) > 0 {
		fmt.Printf("failed commit ids: %v\n", failedIDs)
	}
	return nil
}

func batchUpdateReleasedConfigItemMd5(kt *kit.Kit, tx *gorm.DB, provider repository.Provider,
	md5Map map[string]string) error {
	var currentMaxID uint32
	releasedCIs := []*ReleasedConfigItem20240221103509{}
	if err := tx.Model(&ReleasedConfigItem20240221103509{}).Select("COALESCE(MAX(id), 0)").Row().
		Scan(&currentMaxID); err != nil {
		return err
	}
	if err := tx.Model(&ReleasedConfigItem20240221103509{}).Where("id <= ?", currentMaxID).
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
		if err := tx.Model(&ReleasedConfigItem20240221103509{}).Where("id = ?", releasedCI.ID).
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

func batchUpdateReleasedAppTemplateMd5(kt *kit.Kit, tx *gorm.DB, provider repository.Provider,
	md5Map map[string]string) error {
	var currentMaxID uint32
	releasedATs := []*ReleasedAppTemplate20240221103509{}
	if err := tx.Model(&ReleasedAppTemplate20240221103509{}).Select("COALESCE(MAX(id), 0)").Row().
		Scan(&currentMaxID); err != nil {
		return err
	}
	if err := tx.Model(&ReleasedAppTemplate20240221103509{}).Where("id <= ?", currentMaxID).
		Find(&releasedATs).Error; err != nil {
		return err
	}

	successCount := 0
	failedIDs := []uint32{}
	for _, releasedAT := range releasedATs {
		kt.BizID = releasedAT.BizID
		if releasedAT.Md5 != "" {
			continue
		}
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
		if err := tx.Model(&ReleasedAppTemplate20240221103509{}).Where("id = ?", releasedAT.ID).
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

func batchUpdateTemplateRevisionMd5(kt *kit.Kit, tx *gorm.DB, provider repository.Provider,
	md5Map map[string]string) error {
	var currentMaxID uint32
	templateRevisions := []*TemplateRevision20240221103509{}
	if err := tx.Model(&TemplateRevision20240221103509{}).Select("COALESCE(MAX(id), 0)").Row().
		Scan(&currentMaxID); err != nil {
		return err
	}
	if err := tx.Model(&TemplateRevision20240221103509{}).Where("id <= ?", currentMaxID).
		Find(&templateRevisions).Error; err != nil {
		return err
	}

	successCount := 0
	failedIDs := []uint32{}
	for _, templateRevision := range templateRevisions {
		kt.BizID = templateRevision.BizID
		if templateRevision.Md5 != "" {
			continue
		}
		var md5 string
		if md5Map[templateRevision.Signature] != "" {
			md5 = md5Map[templateRevision.Signature]
		} else {
			metadata, err := provider.Metadata(kt, templateRevision.Signature)
			if err != nil {
				fmt.Printf("get metadata for template_revision %s failed, err: %s\n",
					templateRevision.Signature, err.Error())
				failedIDs = append(failedIDs, templateRevision.ID)
				continue
			}
			md5 = metadata.Md5
			md5Map[templateRevision.Signature] = md5
		}
		if err := tx.Model(&TemplateRevision20240221103509{}).Where("id = ?", templateRevision.ID).
			Update("md5", md5).Error; err != nil {
			fmt.Printf("update template_revision %d md5 failed, err: %s\n", templateRevision.ID, err.Error())
			failedIDs = append(failedIDs, templateRevision.ID)
			continue
		}
		successCount++
	}
	fmt.Printf("batch update template_revision md5 success count: %d, failed count: %d\n",
		successCount, len(failedIDs))
	if len(failedIDs) > 0 {
		fmt.Printf("failed template_revision ids: %v\n", failedIDs)
	}
	return nil
}
