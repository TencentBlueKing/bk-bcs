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

package table

import (
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// ReleasedConfigItemColumns defines ReleasedConfigItem's columns
var ReleasedConfigItemColumns = mergeColumns(ReleasedCIColumnDescriptor)

// ReleasedCIColumnDescriptor is ReleasedConfigItem's column descriptors.
var ReleasedCIColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "id", NamedC: "id", Type: enumor.Numeric},
		{Column: "release_id", NamedC: "release_id", Type: enumor.Numeric},
		{Column: "config_item_id", NamedC: "config_item_id", Type: enumor.Numeric},
		{Column: "commit_id", NamedC: "commit_id", Type: enumor.Numeric}},
	mergeColumnDescriptors("commit_spec", CommitSpecColumnDescriptor),
	mergeColumnDescriptors("config_item_spec", RCISpecColumnDescriptor),
	mergeColumnDescriptors("attachment", CIAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor))

// ReleasedConfigItem records all the information when a config item
// is released. it is not editable after created.
type ReleasedConfigItem struct {
	// ID is an auto-increased value, which is a unique identity
	// of a released app config items.
	ID uint32 `db:"id" json:"id" gorm:"primaryKey"`

	// ReleaseID is this app's config item's release id
	ReleaseID uint32 `db:"release_id" json:"release_id" gorm:"column:release_id"`

	// CommitID is this config item's commit id when it is released.
	CommitID uint32 `db:"commit_id" json:"commit_id" gorm:"column:commit_id"`

	// ConfigItemID is the config item's origin id when it is released.
	ConfigItemID uint32 `db:"config_item_id" json:"config_item_id" gorm:"column:config_item_id"`

	// CommitSpec is this config item's commit spec when it is released.
	// which is same with the commits' spec information with the upper
	// CommitID
	CommitSpec *ReleasedCommitSpec `db:"commit_spec" json:"commit_spec" gorm:"embedded"`

	// ConfigItemSpec is this config item's spec when it is released, which
	// means it is same with the config item's spec information when it is
	// released.
	ConfigItemSpec *ConfigItemSpec       `db:"config_item_spec" json:"config_item_spec" gorm:"embedded"`
	Attachment     *ConfigItemAttachment `db:"attachment" json:"attachment" gorm:"embedded"`
	Revision       *Revision             `db:"revision" json:"revision" gorm:"embedded"`
}

// TableName is the released app config's database table name.
func (r *ReleasedConfigItem) TableName() string {
	return "released_config_items"
}

// AppID AuditRes interface
func (r *ReleasedConfigItem) AppID() uint32 {
	return r.Attachment.AppID
}

// ResID AuditRes interface
func (r *ReleasedConfigItem) ResID() uint32 {
	return r.ID
}

// ResType AuditRes interface
func (r *ReleasedConfigItem) ResType() string {
	return "released_config_item"
}

// RciList is released config items
type RciList []*ReleasedConfigItem

// AppID AuditRes interface
func (rs RciList) AppID() uint32 {
	if len(rs) > 0 {
		return rs[0].Attachment.AppID
	}
	return 0
}

// ResID AuditRes interface
func (rs RciList) ResID() uint32 {
	if len(rs) > 0 {
		return rs[0].ID
	}
	return 0
}

// ResType AuditRes interface
func (rs RciList) ResType() string {
	return "released_config_item"
}

// Validate the released config item information.
func (r *ReleasedConfigItem) Validate(kit *kit.Kit) error {
	if r.ID != 0 {
		return errors.New("id should not set")
	}

	if r.ReleaseID <= 0 {
		return errors.New("invalid release id")
	}

	if r.CommitSpec == nil {
		return errors.New("commit spec is empty")
	}

	// when config item id = 0 ,it is a rendered template config item
	// when config item id > 0, it is a normal config item (not rendered from template)
	if r.ConfigItemID > 0 {
		if r.CommitID <= 0 {
			return errors.New("invalid commit id")
		}

		if err := r.CommitSpec.Validate(kit); err != nil {
			return err
		}
	} else {
		// for rendered template config item, need to validate content signature
		if err := r.CommitSpec.Content.Validate(kit); err != nil {
			return err
		}
	}

	if r.ConfigItemSpec == nil {
		return errors.New("config item spec is empty")
	}

	if err := r.ConfigItemSpec.ValidateCreate(kit); err != nil {
		return fmt.Errorf("invalid config item spec, err: %v", err)
	}

	if r.Attachment == nil {
		return errors.New("attachment is empty")
	}

	if err := r.Attachment.Validate(); err != nil {
		return err
	}

	if r.Revision == nil {
		return errors.New("revision is empty")
	}

	if err := r.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}
