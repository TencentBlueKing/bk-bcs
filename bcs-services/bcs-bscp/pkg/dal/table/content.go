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
	"strings"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// ContentColumns defines 's columns
var ContentColumns = mergeColumns(ContentColumnDescriptor)

// ContentColumnDescriptor is Content's column descriptors.
var ContentColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", ContentSpecColumnDescriptor),
	mergeColumnDescriptors("attachment", ContentAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", CreatedRevisionColumnDescriptor))

// Content is definition for content.
type Content struct {
	// ID is an auto-increased value, which is this content's
	// unique identity.
	ID         uint32             `db:"id" json:"id" gorm:"primaryKey"`
	Spec       *ContentSpec       `db:"spec" json:"spec" gorm:"embedded"`
	Attachment *ContentAttachment `db:"attachment" json:"attachment" gorm:"embedded"`
	Revision   *CreatedRevision   `db:"revision" json:"revision" gorm:"embedded"`
}

// AppID AuditRes interface
func (c *Content) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (c *Content) ResID() uint32 {
	return c.ID
}

// ResType AuditRes interface
func (c *Content) ResType() string {
	return "content"
}

// TableName is the content's database table name.
func (c Content) TableName() Name {
	return ContentTable
}

// ValidateCreate validate create information when content is created.
func (c Content) ValidateCreate(kit *kit.Kit) error {
	if c.ID != 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "content id can not set"))
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

	if err := c.Attachment.Validate(kit); err != nil {
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

// ContentSpecColumns defines ContentSpec's columns
var ContentSpecColumns = mergeColumns(ContentSpecColumnDescriptor)

// ContentSpecColumnDescriptor is ContentSpec's column descriptors.
var ContentSpecColumnDescriptor = ColumnDescriptors{
	{Column: "signature", NamedC: "signature", Type: enumor.String},
	{Column: "byte_size", NamedC: "byte_size", Type: enumor.Numeric}}

// ContentSpec is a collection of a content specifics.
// all the fields under the content spec can not be updated.
type ContentSpec struct {
	// Signature is the sha256 value of a configuration file's
	// content, it can not be updated.
	Signature string `db:"signature" json:"signature" gorm:"column:signature"`
	// ByteSize is the size of this content in byte.
	// can not be updated
	ByteSize uint64 `db:"byte_size" json:"byte_size" gorm:"column:byte_size"`
	// Md5 is the md5 value of a configuration file's content.
	// it can not be updated.
	Md5 string `db:"md5" json:"md5" gorm:"column:md5"`
}

// ReleasedContentSpec is a collection of a released content specifics.
// all the fields under the content spec can not be updated.
type ReleasedContentSpec struct {
	// Signature is the sha256 value of a configuration file's
	// content after render, it can not be updated.
	Signature string `db:"signature" json:"signature" gorm:"column:signature"`
	// ByteSize is the size of this content in byte after render.
	// can not be updated
	ByteSize uint64 `db:"byte_size" json:"byte_size" gorm:"column:byte_size"`
	// OriginSignature is the sha256 value of a configuration file's
	// content before render, it can not be updated.
	OriginSignature string `db:"origin_signature" json:"origin_signature" gorm:"column:origin_signature"`
	// OriginByteSize is the size of this content in byte before render.
	// can not be updated
	OriginByteSize uint64 `db:"origin_byte_size" json:"origin_byte_size" gorm:"column:origin_byte_size"`
	// Md5 is the md5 value of a configuration file's content after render.
	// it can not be updated.
	Md5 string `db:"md5" json:"md5" gorm:"column:md5"`
}

// Validate content's spec
func (cs ContentSpec) Validate(kit *kit.Kit) error {
	// a file's sha256 signature value's length is 64.
	if len(cs.Signature) != 64 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid content signature, should be config's sha256 value"))
	}

	if cs.Signature != strings.ToLower(cs.Signature) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "content signature should be lowercase"))
	}
	return nil
}

// Validate released content's spec
func (cs ReleasedContentSpec) Validate(kit *kit.Kit) error {
	// a file's sha256 signature value's length is 64.
	if len(cs.Signature) != 64 {
		return errf.Errorf(errf.InvalidArgument,
			i18n.T(kit, "invalid content signature, should be config's sha256 value"))
	}
	if len(cs.OriginSignature) != 64 {
		return errf.Errorf(errf.InvalidArgument,
			i18n.T(kit, "invalid origin content signature, should be config's sha256 value"))
	}

	if cs.Signature != strings.ToLower(cs.Signature) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "content signature should be lowercase"))
	}
	if cs.OriginSignature != strings.ToLower(cs.OriginSignature) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "origin content signature should be lowercase"))
	}
	return nil
}

// ContentAttachmentColumns defines ContentAttachment's columns
var ContentAttachmentColumns = mergeColumns(ContentAttachmentColumnDescriptor)

// ContentAttachmentColumnDescriptor is ContentAttachment's column descriptors.
var ContentAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
	{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric},
	{Column: "config_item_id", NamedC: "config_item_id", Type: enumor.Numeric}}

// ContentAttachment defines content's attachment information
type ContentAttachment struct {
	BizID        uint32 `db:"biz_id" json:"biz_id" gorm:"column:biz_id"`
	AppID        uint32 `db:"app_id" json:"app_id" gorm:"column:app_id"`
	ConfigItemID uint32 `db:"config_item_id" json:"config_item_id" gorm:"column:config_item_id"`
}

// Validate content attachment.
func (c ContentAttachment) Validate(kit *kit.Kit) error {
	if c.BizID <= 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid biz id"))
	}

	if c.AppID <= 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid app id"))
	}

	if c.ConfigItemID <= 0 {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "invalid config item id"))
	}

	return nil
}
