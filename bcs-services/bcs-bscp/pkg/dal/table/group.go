/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package table

import (
	"errors"
	"fmt"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/validator"
	"bscp.io/pkg/runtime/selector"
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
	ID         uint32           `db:"id" json:"id"`
	Spec       *GroupSpec       `db:"spec" json:"spec"`
	Attachment *GroupAttachment `db:"attachment" json:"attachment"`
	Revision   *Revision        `db:"revision" json:"revision"`
}

// TableName is the group's database table name.
func (s Group) TableName() Name {
	return GroupTable
}

// ValidateCreate validate group is valid or not when create it.
func (s Group) ValidateCreate() error {

	if s.ID > 0 {
		return errors.New("id should not be set")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.ValidateCreate(); err != nil {
		return err
	}

	if s.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := s.Attachment.Validate(); err != nil {
		return err
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if err := s.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate group is valid or not when update it.
func (s Group) ValidateUpdate() error {

	if s.ID <= 0 {
		return errors.New("id should be set")
	}

	changed := false
	if s.Spec != nil {
		changed = true
		if err := s.Spec.ValidateUpdate(); err != nil {
			return err
		}
	}

	if s.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if !changed {
		return errors.New("nothing is found to be change")
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if err := s.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the group's info when delete it.
func (s Group) ValidateDelete() error {
	if s.ID <= 0 {
		return errors.New("group id should be set")
	}

	if s.Attachment.BizID <= 0 {
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
	Public   bool               `db:"public" json:"public"`
	Mode     GroupMode          `db:"mode" json:"mode"`
	Selector *selector.Selector `db:"selector" json:"selector"`
	UID      string             `db:"uid" json:"uid"`
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
	// TODO: BuiltIn define bscp built-in group,eg. ClusterID, Namespace, CMDBModuleID...
	BuiltIn GroupMode = "builtin"
)

// GroupMode is the mode of an group works in
type GroupMode string

// String returns the string value of GroupMode.
func (s GroupMode) String() string {
	return string(s)
}

// Validate strategy set type.
func (s GroupMode) Validate() error {
	switch s {
	case Custom:
	case Debug:
	case Default:
	default:
		return fmt.Errorf("unsupported group working mode: %s", s)
	}

	return nil
}

// ValidateCreate validate group spec when it is created.
func (s GroupSpec) ValidateCreate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}
	if err := s.Mode.Validate(); err != nil {
		return err
	}
	switch s.Mode {
	case Custom:
		if s.Selector == nil || s.Selector.IsEmpty() {
			return errors.New("group works in custom mode, selector should be set")
		}
	case Debug:
		if s.UID == "" {
			return errors.New("group works in debug mode, uid should be set")
		}
	}
	return nil
}

// ValidateUpdate validate group spec when it is updated.
func (s GroupSpec) ValidateUpdate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if s.Mode != "" {
		return errors.New("group's mode can not be updated")
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
	BizID uint32 `db:"biz_id" json:"biz_id"`
}

// IsEmpty test whether group attachment is empty or not.
func (s GroupAttachment) IsEmpty() bool {
	return s.BizID == 0
}

// Validate whether group attachment is valid or not.
func (s GroupAttachment) Validate() error {
	if s.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}
	return nil
}
