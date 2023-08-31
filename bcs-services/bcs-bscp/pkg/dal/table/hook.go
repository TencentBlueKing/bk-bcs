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
)

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
func (h Hook) ValidateCreate() error {

	if h.ID > 0 {
		return errors.New("id should not be set")
	}

	if h.Spec == nil {
		return errors.New("spec not set")
	}

	if err := h.Spec.ValidateCreate(); err != nil {
		return err
	}

	if h.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := h.Attachment.Validate(); err != nil {
		return err
	}

	if h.Revision == nil {
		return errors.New("revision not set")
	}

	if err := h.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the hook's info when delete it.
func (h Hook) ValidateDelete() error {
	if h.ID <= 0 {
		return errors.New("hook id should be set")
	}

	if h.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

// HookSpec defines all the specifics for hook set by user.
type HookSpec struct {
	Name string `json:"name" gorm:"column:name"`
	// Type is the hook type of hook
	Type ScriptType `json:"type" gorm:"column:type"`
	// Tag
	Tag  string `json:"tag" gorm:"column:tag"`
	Memo string `json:"memo" gorm:"column:memo"`
}

const (
	// Shell is the type for shell hook
	Shell ScriptType = "shell"

	// Python is the type for python hook
	Python ScriptType = "python"
)

// ScriptType is the type of hook script
type ScriptType string

// String returns string value of ScriptType
func (s ScriptType) String() string {
	return string(s)
}

// Validate validate the hook type
func (s ScriptType) Validate() error {
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
