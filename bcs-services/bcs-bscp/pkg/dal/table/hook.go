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
	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/validator"
	"errors"
	"fmt"
)

// HookColumns defines Hook's columns
var HookColumns = mergeColumns(HookColumnDescriptor)

// HookColumnDescriptor is Hook's column descriptors.
var HookColumnDescriptor = mergeColumnDescriptors("",
	ColumnDescriptors{{Column: "id", NamedC: "id", Type: enumor.Numeric}},
	mergeColumnDescriptors("spec", HookSpecColumnDescriptor),
	mergeColumnDescriptors("attachment", HookAttachmentColumnDescriptor),
	mergeColumnDescriptors("revision", RevisionColumnDescriptor))

// Hook defines a hook for an app to publish.
// it contains the selector to define the scope of the matched instances.
type Hook struct {
	// ID is an auto-increased value, which is a unique identity of a hook.
	ID         uint32          `json:"id" gorm:"primaryKey"`
	Spec       *HookSpec       `json:"spec" gorm:"embedded"`
	Attachment *HookAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision       `json:"revision" gorm:"embedded"`
}

// TableName is the Hook's database table name.
func (h *Hook) TableName() string {
	return "hooks"
}

// AppID HookRes interface
func (h *Hook) AppID() uint32 {
	return 0
}

// ResID HookRes interface
func (h *Hook) ResID() uint32 {
	return h.ID
}

// ResType HookRes interface
func (h *Hook) ResType() string {
	return "hook"
}

// ValidateCreate validate hook is valid or not when create it.
func (s Hook) ValidateCreate() error {

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

// ValidateUpdate validate hook is valid or not when update it.
func (s Hook) ValidateUpdate() error {

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

// ValidateDelete validate the hook's info when delete it.
func (s Hook) ValidateDelete() error {
	if s.ID <= 0 {
		return errors.New("hook id should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if err := s.Spec.ValidateDelete(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the TemplateSpace's info when delete it.
func (s HookSpec) ValidateDelete() error {

	return nil
}

// HookSpecColumns defines HookSpec's columns
var HookSpecColumns = mergeColumns(HookSpecColumnDescriptor)

// HookSpecColumnDescriptor is HookSpec's column descriptors.
var HookSpecColumnDescriptor = ColumnDescriptors{
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "tag", NamedC: "tag", Type: enumor.String},
}

// HookSpec defines all the specifics for hook set by user.
type HookSpec struct {
	Name string `json:"name" gorm:"column:name"`
	// Type is the hook type of hook
	Type HookType `json:"type" gorm:"column:type"`
	// Tag
	Tag  string `json:"tag" gorm:"column:tag"`
	Memo string `json:"memo" gorm:"column:memo"`
}

const (
	// Shell is the type for shell hook
	Shell HookType = "shell"

	// Python is the type for python hook
	Python HookType = "python"
)

// HookType is the type of hook
type HookType string

// Validate validate the hook type
func (s HookType) Validate() error {
	if s == "" {
		return nil
	}
	switch s {
	case Shell:
	case Python:
	default:
		return fmt.Errorf("unsupported hook type: %s", s)
	}

	return nil
}

// ValidateCreate validate hook spec when it is created.
func (s HookSpec) ValidateCreate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if err := s.Type.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate hook spec when it is updated.
func (s HookSpec) ValidateUpdate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	return nil
}

// ValidateShellHookSecurity validate security of shell hook content
func (s HookSpec) ValidateShellHookSecurity(hookContent string) error {
	// TODO implement this
	return nil
}

// ValidatePythonHookSecurity validate security of python hook content
func (s HookSpec) ValidatePythonHookSecurity(hookContent string) error {
	// TODO implement this
	return nil
}

// HookAttachmentColumns defines HookAttachment's columns
var HookAttachmentColumns = mergeColumns(HookAttachmentColumnDescriptor)

// HookAttachmentColumnDescriptor is HookAttachment's column descriptors.
var HookAttachmentColumnDescriptor = ColumnDescriptors{
	{Column: "biz_id", NamedC: "biz_id", Type: enumor.Numeric},
}

// HookAttachment defines the hook attachments.
type HookAttachment struct {
	BizID uint32 `db:"biz_id" gorm:"column:biz_id"`
}

// IsEmpty test whether hook attachment is empty or not.
func (s HookAttachment) IsEmpty() bool {
	return s.BizID == 0
}

// Validate whether hook attachment is valid or not.
func (s HookAttachment) Validate() error {
	if s.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	return nil
}
