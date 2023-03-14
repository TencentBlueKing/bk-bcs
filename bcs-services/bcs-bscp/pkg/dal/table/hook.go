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
	ID         uint32          `db:"id" json:"id"`
	Spec       *HookSpec       `db:"spec" json:"spec"`
	Attachment *HookAttachment `db:"attachment" json:"attachment"`
	Revision   *Revision       `db:"revision" json:"revision"`
}

// TableName is the hook's database table name.
func (s Hook) TableName() Name {
	return HookTable
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

	if s.Attachment.AppID <= 0 {
		return errors.New("app id should be set")
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

	return nil
}

// HookSpecColumns defines HookSpec's columns
var HookSpecColumns = mergeColumns(HookSpecColumnDescriptor)

// HookSpecColumnDescriptor is HookSpec's column descriptors.
var HookSpecColumnDescriptor = ColumnDescriptors{
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "pre_type", NamedC: "pre_type", Type: enumor.String},
	{Column: "pre_hook", NamedC: "pre_hook", Type: enumor.String},
	{Column: "post_type", NamedC: "post_type", Type: enumor.String},
	{Column: "post_hook", NamedC: "post_hook", Type: enumor.String},
}

// HookSpec defines all the specifics for hook set by user.
type HookSpec struct {
	Name string `db:"name" json:"name"`
	// PreType is the hook type of pre hook
	PreType HookType `db:"pre_type" json:"pre_type"`
	// PreHook is the content of pre hook
	PreHook string `db:"pre_hook" json:"pre_hook"`
	// PostType is the hook type of post hook
	PostType HookType `db:"post_type" json:"post_type"`
	// PostHook is the content of post hook
	PostHook string `db:"post_hook" json:"post_hook"`
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
	if err := s.PreType.Validate(); err != nil {
		return err
	}
	if err := s.PostType.Validate(); err != nil {
		return err
	}

	if err := s.ValidateHookContentSecurity(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate hook spec when it is updated.
func (s HookSpec) ValidateUpdate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if err := s.ValidateHookContentSecurity(); err != nil {
		return err
	}

	return nil
}

// ValidateHookContentSecurity validate security of hook content
func (s HookSpec) ValidateHookContentSecurity() error {
	if s.PreHook != "" {
		switch s.PreType {
		case Shell:
			if err := s.ValidateShellHookSecurity(s.PreHook); err != nil {
				return err
			}
		case Python:
			if err := s.ValidatePythonHookSecurity(s.PreHook); err != nil {
				return err
			}
		case "":
			return fmt.Errorf("pre hook must set a hook type")
		}
	}

	if s.PostHook != "" {
		switch s.PostType {
		case Shell:
			if err := s.ValidateShellHookSecurity(s.PostHook); err != nil {
				return err
			}
		case Python:
			if err := s.ValidatePythonHookSecurity(s.PostHook); err != nil {
				return err
			}
		case "":
			return fmt.Errorf("post hook must set a hook type")
		}
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
	{Column: "app_id", NamedC: "app_id", Type: enumor.Numeric},
	{Column: "release_id", NamedC: "release_id", Type: enumor.Numeric}}

// HookAttachment defines the hook attachments.
type HookAttachment struct {
	BizID     uint32 `db:"biz_id" json:"biz_id"`
	AppID     uint32 `db:"app_id" json:"app_id"`
	ReleaseID uint32 `db:"release_id" json:"release_id"`
}

// IsEmpty test whether hook attachment is empty or not.
func (s HookAttachment) IsEmpty() bool {
	return s.BizID == 0 && s.AppID == 0
}

// Validate whether hook attachment is valid or not.
func (s HookAttachment) Validate() error {
	if s.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if s.AppID <= 0 {
		return errors.New("invalid attachment app id")
	}

	return nil
}
