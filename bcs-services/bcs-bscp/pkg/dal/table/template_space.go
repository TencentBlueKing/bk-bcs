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

	"bscp.io/pkg/criteria/validator"
)

// TemplateSpace 模版空间
type TemplateSpace struct {
	ID         uint32                   `json:"id" gorm:"primaryKey"`
	Spec       *TemplateSpaceSpec       `json:"spec" gorm:"embedded"`
	Attachment *TemplateSpaceAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision                `json:"revision" gorm:"embedded"`
}

// TableName is the TemplateSpace's database table name.
func (s *TemplateSpace) TableName() string {
	return "template_spaces"
}

// AppID AuditRes interface
func (s *TemplateSpace) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (s *TemplateSpace) ResID() uint32 {
	return s.ID
}

// ResType AuditRes interface
func (s *TemplateSpace) ResType() string {
	return "template_space"
}

// ValidateCreate validate TemplateSpace is valid or not when create it.
func (s TemplateSpace) ValidateCreate() error {
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

// ValidateUpdate validate TemplateSpace is valid or not when update it.
func (s TemplateSpace) ValidateUpdate() error {

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

// ValidateDelete validate the TemplateSpace's info when delete it.
func (s TemplateSpace) ValidateDelete() error {
	if s.ID <= 0 {
		return errors.New("TemplateSpace id should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}

// TemplateSpaceSpec defines all the specifics for TemplateSpace set by user.
type TemplateSpaceSpec struct {
	Name string `json:"name" gorm:"column:name"`
	Memo string `json:"memo" gorm:"column:memo"`
}

// TemplateSpaceType is the type of TemplateSpace
type TemplateSpaceType string

// ValidateCreate validate TemplateSpace spec when it is created.
func (s TemplateSpaceSpec) ValidateCreate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if err := validator.ValidateAppName(s.Name); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate TemplateSpace spec when it is updated.
func (s TemplateSpaceSpec) ValidateUpdate() error {
	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if err := validator.ValidateMemo(s.Memo, false); err != nil {
		return err
	}

	return nil
}

// TemplateSpaceAttachment defines the TemplateSpace attachments.
type TemplateSpaceAttachment struct {
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
}

// IsEmpty test whether TemplateSpace attachment is empty or not.
func (s TemplateSpaceAttachment) IsEmpty() bool {
	return s.BizID == 0
}

// Validate whether TemplateSpace attachment is valid or not.
func (s TemplateSpaceAttachment) Validate() error {
	if s.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	return nil
}
