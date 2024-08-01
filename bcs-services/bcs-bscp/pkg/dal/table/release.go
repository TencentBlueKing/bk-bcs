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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// ReleaseColumns defines Release's columns
var ReleaseColumns = mergeColumns(ReleaseColumnDescriptor)

// ReleaseColumnDescriptor is Release's column descriptors.
var ReleaseColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", ReleaseSpecColumnDescriptor),
	mergeColumnDescriptors("attachment", ReleaseAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", CreatedRevisionColumnDescriptor))

// Release is a release version of an app's all configuration items.
// A release is not editable once created.
type Release struct {
	// ID is an auto-increased value, which is a unique identity
	// of a commit.
	ID         uint32             `db:"id" json:"id"`
	Spec       *ReleaseSpec       `db:"spec" json:"spec" gorm:"embedded"`
	Attachment *ReleaseAttachment `db:"attachment" json:"attachment" gorm:"embedded"`
	Revision   *CreatedRevision   `db:"revision" json:"revision" gorm:"embedded"`
}

// TableName is the release's database table name.
func (r Release) TableName() Name {
	return ReleaseTable
}

// AppID AuditRes interface
func (r *Release) AppID() uint32 {
	return r.Attachment.AppID
}

// ResID AuditRes interface
func (r *Release) ResID() uint32 {
	return r.ID
}

// ResType AuditRes interface
func (r *Release) ResType() string {
	return "release"
}

// ValidateCreate a release's information
func (r Release) ValidateCreate(kit *kit.Kit) error {
	if r.ID != 0 {
		return errors.New("id should not set")
	}

	if r.Spec == nil {
		return errors.New("spec should be set")
	}

	if err := r.Spec.Validate(kit); err != nil {
		return err
	}

	if r.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := r.Attachment.Validate(); err != nil {
		return err
	}

	if r.Revision == nil {
		return errors.New("revision should be set")
	}

	if err := r.Revision.Validate(); err != nil {
		return err
	}

	return nil
}

// ReleaseSpecColumns defines ReleaseSpec's columns
var ReleaseSpecColumns = mergeColumns(ReleaseSpecColumnDescriptor)

// ReleaseSpecColumnDescriptor is ReleaseSpec's column descriptors.
var ReleaseSpecColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{
		{Column: "name", NamedC: "name", Type: enumor.String},
		{Column: "memo", NamedC: "memo", Type: enumor.String},
		{Column: "deprecated", NamedC: "deprecated", Type: enumor.Boolean},
		{Column: "publish_num", NamedC: "publish_num", Type: enumor.Numeric},
	},
	mergeColumnDescriptors("hook", HookColumnDescriptor))

// HookColumnDescriptor is hook column descriptor
var HookColumnDescriptor = ColumnDescriptors{
	{Column: "pre_hook_id", NamedC: "pre_hook_id", Type: enumor.Numeric},
	{Column: "pre_hook_revision_id", NamedC: "pre_hook_revision_id", Type: enumor.Numeric},
	{Column: "post_hook_id", NamedC: "post_hook_id", Type: enumor.Numeric},
	{Column: "post_hook_revision_id", NamedC: "post_hook_revision_id", Type: enumor.Numeric},
}

// ReleaseSpec defines all the specifics related with a release, which is set by user.
type ReleaseSpec struct {
	Name       string `db:"name" json:"name"`
	Memo       string `db:"memo" json:"memo"`
	Deprecated bool   `db:"deprecated" json:"deprecated"`
	PublishNum uint32 `db:"publish_num" json:"publish_num"`
	// 是否全量发布过
	FullyReleased bool `db:"fully_released" json:"fully_released"`
}

// Validate a release specifics when it is created.
func (r ReleaseSpec) Validate(kit *kit.Kit) error {
	if err := validator.ValidateReleaseName(kit, r.Name); err != nil {
		return err
	}

	if err := validator.ValidateMemo(kit, r.Memo, false); err != nil {
		return err
	}

	return nil
}

const (
	// FullReleased means that app all groups were released in this release,
	// must include default group.
	FullReleased ReleaseStatus = "full_released"

	// PartialReleased means that app not all groups released in this release,
	PartialReleased ReleaseStatus = "partial_released"

	// NotReleased means that no group released in this release.
	NotReleased ReleaseStatus = "not_released"
)

// ReleaseStatus defines release status.
type ReleaseStatus string

// String returns release status string.
func (s ReleaseStatus) String() string {
	return string(s)
}

// Validate strategy set type.
func (s ReleaseStatus) Validate() error {
	switch s {
	case FullReleased:
	case PartialReleased:
	default:
		return fmt.Errorf("unsupported release released status: %s", s)
	}

	return nil
}

// ReleaseAttachmentColumns defines ReleaseAttachment's columns
var ReleaseAttachmentColumns = mergeColumns(ReleaseAttachmentColumnDescriptor)

// ReleaseAttachmentColumnDescriptor is ReleaseAttachment's column descriptors.
var ReleaseAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric}}

// ReleaseAttachment defines release related information.
type ReleaseAttachment struct {
	BizID uint32 `db:"biz_id" json:"biz_id" gorm:"column:biz_id"`
	AppID uint32 `db:"app_id" json:"app_id" gorm:"column:app_id"`
}

// IsEmpty test whether this release attachment is empty or not.
func (r ReleaseAttachment) IsEmpty() bool {
	return r.BizID == 0 && r.AppID == 0
}

// Validate release's attachment information
func (r ReleaseAttachment) Validate() error {
	if r.BizID <= 0 {
		return errors.New("invalid biz id")
	}

	if r.AppID <= 0 {
		return errors.New("invalid app id")
	}

	return nil
}
