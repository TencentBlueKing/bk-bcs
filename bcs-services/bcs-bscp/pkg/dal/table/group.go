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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
)

// GroupColumns defines Group's columns
var GroupColumns = mergeColumns(GroupColumnDescriptor)

// GroupColumnDescriptor is Group's column descriptors.
var GroupColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", GroupSpecColumnDescriptor),
	mergeColumnDescriptors("attachment", GroupAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor))

// Group defines a group for an app to publish.
// it contains the selector to define the scope of the matched instances.
type Group struct {
	// ID is an auto-increased value, which is a unique identity of a group.
	ID         uint32           `db:"id" json:"id" gorm:"primaryKey"`
	Spec       *GroupSpec       `db:"spec" json:"spec" gorm:"embedded"`
	Attachment *GroupAttachment `db:"attachment" json:"attachment" gorm:"embedded"`
	Revision   *Revision        `db:"revision" json:"revision" gorm:"embedded"`
}

// TableName is the group's database table name.
func (g Group) TableName() string {
	return "groups"
}

// AppID AuditRes interface
func (g Group) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (g Group) ResID() uint32 {
	return g.ID
}

// ResType AuditRes interface
func (g Group) ResType() string {
	return "app"
}

// ValidateCreate validate group is valid or not when create it.
func (g Group) ValidateCreate(kit *kit.Kit) error {

	if g.ID > 0 {
		return errors.New("id should not be set")
	}

	if g.Spec == nil {
		return errors.New("spec not set")
	}

	if err := g.Spec.ValidateCreate(kit); err != nil {
		return err
	}

	if g.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := g.Attachment.Validate(); err != nil {
		return err
	}

	if g.Revision == nil {
		return errors.New("revision not set")
	}

	if err := g.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate group is valid or not when update it.
func (g Group) ValidateUpdate(kit *kit.Kit) error {

	if g.ID <= 0 {
		return errors.New("id should be set")
	}

	changed := false
	if g.Spec != nil {
		changed = true
		if err := g.Spec.ValidateUpdate(kit); err != nil {
			return err
		}
	}

	if g.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if g.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if !changed {
		return errors.New("nothing is found to be change")
	}

	if g.Revision == nil {
		return errors.New("revision not set")
	}

	if err := g.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the group's info when delete it.
func (g Group) ValidateDelete() error {
	if g.ID <= 0 {
		return errors.New("group id should be set")
	}

	if g.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

// GroupSpecColumns defines GroupSpec's columns
var GroupSpecColumns = mergeColumns(GroupSpecColumnDescriptor)

// GroupSpecColumnDescriptor is GroupSpec's column descriptors.
var GroupSpecColumnDescriptor = ColumnDescriptors{
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "public", NamedC: "public", Type: enumor.Boolean},
	{Column: "mode", NamedC: "mode", Type: enumor.String},
	{Column: "selector", NamedC: "selector", Type: enumor.String},
	{Column: "uid", NamedC: "uid", Type: enumor.String},
}

// GroupSpec defines all the specifics for group set by user.
type GroupSpec struct {
	Name string `db:"name" json:"name"`
	// Public defines weather group can be used by all apps.
	// It can not be updated once it is created.
	Public   bool               `db:"public" json:"public" gorm:"column:public"`
	Mode     GroupMode          `db:"mode" json:"mode" gorm:"column:mode"`
	Selector *selector.Selector `db:"selector" json:"selector" gorm:"column:selector;type:json"`
	UID      string             `db:"uid" json:"uid" gorm:"column:uid"`
}

const (
	// Custom means this is a user customed group, it's selector is defined by user
	Custom GroupMode = "custom"
	// Debug means that this group can noly set UID,
	// in other word can only select specific instance
	Debug GroupMode = "debug"
	// Default will select instances that won't be selected by any other released groups
	Default GroupMode = "default"
	// BuiltIn define bscp built-in group,eg. ClusterID, Namespace, CMDBModuleID...
	// Note: BuiltIn define bscp built-in group,eg. ClusterID, Namespace, CMDBModuleID...
	BuiltIn GroupMode = "builtin"
)

// GroupMode is the mode of an group works in
type GroupMode string

// String returns the string value of GroupMode.
func (g GroupMode) String() string {
	return string(g)
}

// Validate strategy set type.
func (g GroupMode) Validate() error {
	switch g {
	case Custom:
	case Debug:
	case Default:
	default:
		return fmt.Errorf("unsupported group working mode: %s", g)
	}

	return nil
}

// ValidateCreate validate group spec when it is created.
func (g GroupSpec) ValidateCreate(kit *kit.Kit) error {
	if err := validator.ValidateName(kit, g.Name); err != nil {
		return err
	}
	if err := g.Mode.Validate(); err != nil {
		return err
	}
	switch g.Mode {
	case Custom:
		if g.Selector == nil || g.Selector.IsEmpty() {
			return errors.New("group works in custom mode, selector should be set")
		}
		if err := g.Selector.Validate(); err != nil {
			return fmt.Errorf("group works in custom mode, selector is invalid, err: %v", err)
		}
	case Debug:
		if g.UID == "" {
			return errors.New("group works in debug mode, uid should be set")
		}
	default:
		return fmt.Errorf("unsupported group working mode: %s", g.Mode.String())
	}
	return nil
}

// ValidateUpdate validate group spec when it is updated.
func (g GroupSpec) ValidateUpdate(kit *kit.Kit) error {
	if err := validator.ValidateName(kit, g.Name); err != nil {
		return err
	}

	if g.Mode != "" {
		return errors.New("group's mode can not be updated")
	}

	if g.Selector == nil {
		return errors.New("group's selector should be set")
	}

	// Note: at present, noly custom group's selector can be updated,
	// so we don't need to check other mode's selector.
	if err := g.Selector.Validate(); err != nil {
		return fmt.Errorf("group's selector is invalid, err: %v", err)
	}

	return nil
}

// GroupAttachmentColumns defines GroupAttachment's columns
var GroupAttachmentColumns = mergeColumns(GroupAttachmentColumnDescriptor)

// GroupAttachmentColumnDescriptor is GroupAttachment's column descriptors.
var GroupAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric}}

// GroupAttachment defines the group attachments.
type GroupAttachment struct {
	BizID uint32 `db:"biz_id" json:"biz_id" gorm:"column:biz_id"`
}

// IsEmpty test whether group attachment is empty or not.
func (g GroupAttachment) IsEmpty() bool {
	return g.BizID == 0
}

// Validate whether group attachment is valid or not.
func (g GroupAttachment) Validate() error {
	if g.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}
	return nil
}
