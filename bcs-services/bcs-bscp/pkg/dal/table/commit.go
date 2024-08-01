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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// CommitsColumns defines all the commits' columns.
var CommitsColumns = mergeColumns(CommitsColumnDescriptor)

// CommitsColumnDescriptor is Commit' column descriptors.
var CommitsColumnDescriptor = mergeColumnDescriptors("", ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", CommitSpecColumnDescriptor),
	mergeColumnDescriptors("attachment", CommitAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", CreatedRevisionColumnDescriptor))

// Commit record each change of a configuration item.
// a commit is not editable after created.
type Commit struct {
	// ID is an auto-increased value, which is a unique identity
	// of a commit.
	ID         uint32            `db:"id" json:"id" gorm:"primaryKey"`
	Spec       *CommitSpec       `db:"spec" json:"spec" gorm:"embedded"`
	Attachment *CommitAttachment `db:"attachment" json:"attachment" gorm:"embedded"`
	Revision   *CreatedRevision  `db:"revision" json:"revision" gorm:"embedded"`
}

// AppID AuditRes interface
func (c *Commit) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (c *Commit) ResID() uint32 {
	return c.ID
}

// ResType AuditRes interface
func (c *Commit) ResType() string {
	return "commit"
}

// TableName is the commits' database table name.
func (c Commit) TableName() Name {
	return CommitsTable
}

// ValidateCreate a commit related information when it be created.
func (c Commit) ValidateCreate(kit *kit.Kit) error {
	if c.ID != 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "id should not be set"))
	}

	if c.Spec == nil {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "spec should be set"))
	}

	if err := c.Spec.Validate(kit); err != nil {
		return err
	}

	if c.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := c.Attachment.Validate(); err != nil {
		return err
	}

	if c.Revision == nil {
		return errors.New("revision should be set")
	}

	if err := c.Revision.Validate(); err != nil {
		return err
	}

	return nil
}

// CommitSpecColumns defines all the commit spec's columns.
var CommitSpecColumns = mergeColumns(CommitSpecColumnDescriptor)

// CommitSpecColumnDescriptor is CommitSpec's column descriptors.
var CommitSpecColumnDescriptor = mergeColumnDescriptors("", ColumnDescriptors{
	{Column: "content_id", NamedC: "content_id", Type: enumor.Numeric},
	{Column: "memo", NamedC: "memo", Type: enumor.String}},
	mergeColumnDescriptors("content", ContentSpecColumnDescriptor))

// CommitSpec is the specifics of this committed configuration file.
type CommitSpec struct {
	// ContentID is the identity id of a content.
	ContentID uint32       `db:"content_id" json:"content_id" gorm:"column:content_id"`
	Content   *ContentSpec `db:"content" json:"content" gorm:"embedded"`
	Memo      string       `db:"memo" json:"memo" gorm:"column:memo"`
}

// ReleasedCommitSpec is the specifics of this released committed configuration file.
type ReleasedCommitSpec struct {
	// ContentID is the identity id of a content.
	ContentID uint32               `db:"content_id" json:"content_id" gorm:"column:content_id"`
	Content   *ReleasedContentSpec `db:"content" json:"content" gorm:"embedded"`
	Memo      string               `db:"memo" json:"memo" gorm:"column:memo"`
}

// Validate commit specifics.
func (c CommitSpec) Validate(kit *kit.Kit) error {
	if c.ContentID <= 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid commit spec's content id"))
	}

	if c.Content == nil {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "commit spec's content is empty"))
	}

	if err := validator.ValidateMemo(kit, c.Memo, false); err != nil {
		return err
	}

	return c.Content.Validate(kit)
}

// Validate released commit specifics.
func (c ReleasedCommitSpec) Validate(kit *kit.Kit) error {
	if c.ContentID <= 0 {
		return errors.New("invalid commit spec's content id")
	}

	if c.Content == nil {
		return errors.New("commit spec's content is empty")
	}

	if err := validator.ValidateMemo(kit, c.Memo, false); err != nil {
		return err
	}

	return c.Content.Validate(kit)
}

// CommitAttachmentColumns defines commit attachment's columns
var CommitAttachmentColumns = mergeColumns(CommitAttachmentColumnDescriptor)

// CommitAttachmentColumnDescriptor is CommitAttachment's column descriptors.
var CommitAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric},
	{Column: "config_item_id", NamedC: "config_item_id", Type: enumor.Numeric},
}

// CommitAttachment is the related information of this commit.
type CommitAttachment struct {
	BizID        uint32 `db:"biz_id" json:"biz_id" gorm:"column:biz_id"`
	AppID        uint32 `db:"app_id" json:"app_id" gorm:"column:app_id"`
	ConfigItemID uint32 `db:"config_item_id" json:"config_item_id" gorm:"column:config_item_id"`
}

// Validate commit related information.
func (c CommitAttachment) Validate() error {
	if c.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if c.AppID <= 0 {
		return errors.New("invalid app id")
	}

	if c.ConfigItemID <= 0 {
		return errors.New("invalid config id")
	}

	return nil
}
