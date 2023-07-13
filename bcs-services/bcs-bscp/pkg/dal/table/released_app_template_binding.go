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

	"bscp.io/pkg/dal/types"
)

// ReleasedAppTemplateBinding 已发布的应用模版绑定
type ReleasedAppTemplateBinding struct {
	ID         uint32                                `json:"id" gorm:"primaryKey"`
	Spec       *ReleasedAppTemplateBindingSpec       `json:"spec" gorm:"embedded"`
	Attachment *ReleasedAppTemplateBindingAttachment `json:"attachment" gorm:"embedded"`
	Revision   *CreatedRevision                      `json:"revision" gorm:"embedded"`
}

// TableName is the ReleasedAppTemplateBinding's database table name.
func (t *ReleasedAppTemplateBinding) TableName() string {
	return "released_app_template_bindings"
}

// AppID AuditRes interface
func (t *ReleasedAppTemplateBinding) AppID() uint32 {
	return t.Attachment.AppID
}

// ResID AuditRes interface
func (t *ReleasedAppTemplateBinding) ResID() uint32 {
	return t.ID
}

// ResType AuditRes interface
func (t *ReleasedAppTemplateBinding) ResType() string {
	return "released_app_template_binding"
}

// ValidateCreate validate ReleasedAppTemplateBinding is valid or not when create it.
func (t *ReleasedAppTemplateBinding) ValidateCreate() error {
	if t.ID > 0 {
		return errors.New("id should not be set")
	}

	if t.Spec == nil {
		return errors.New("spec not set")
	}

	if err := t.Spec.ValidateCreate(); err != nil {
		return err
	}

	if t.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	if t.Revision == nil {
		return errors.New("revision not set")
	}

	return nil
}

// ValidateDelete validate the ReleasedAppTemplateBinding's info when delete it.
func (t *ReleasedAppTemplateBinding) ValidateDelete() error {
	if t.ID <= 0 {
		return errors.New("ReleasedAppTemplateBinding id should be set")
	}

	if t.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	return nil
}

// ReleasedAppTemplateBindingSpec defines all the specifics for ReleasedAppTemplateBinding set by user.
type ReleasedAppTemplateBindingSpec struct {
	TemplateSpaceIDs   types.Uint32Slice `json:"template_space_ids" gorm:"column:template_space_ids;type:json;default:'[]'"`
	TemplateSetIDs     types.Uint32Slice `json:"template_set_ids" gorm:"column:template_set_ids;type:json;default:'[]'"`
	TemplateIDs        types.Uint32Slice `json:"template_ids" gorm:"column:template_ids;type:json;default:'[]'"`
	TemplateReleaseIDs types.Uint32Slice `json:"template_release_ids" gorm:"column:template_release_ids;type:json;default:'[]'"`
	Bindings           TemplateBindings  `json:"bindings" gorm:"column:bindings;type:json;default:'[]'"`
	ReleaseID          uint32            `json:"release_id" gorm:"column:release_id"`
}

// ReleasedAppTemplateBindingType is the type of ReleasedAppTemplateBinding
type ReleasedAppTemplateBindingType string

// ValidateCreate validate ReleasedAppTemplateBinding spec when it is created.
func (t *ReleasedAppTemplateBindingSpec) ValidateCreate() error {
	return validateBindings(t.Bindings)
}

// ValidateUpdate validate ReleasedAppTemplateBinding spec when it is updated.
func (t *ReleasedAppTemplateBindingSpec) ValidateUpdate() error {
	return validateBindings(t.Bindings)
}

// ReleasedAppTemplateBindingAttachment defines the ReleasedAppTemplateBinding attachments.
type ReleasedAppTemplateBindingAttachment struct {
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
	AppID uint32 `json:"app_id" gorm:"column:app_id"`
}

// Validate whether ReleasedAppTemplateBinding attachment is valid or not.
func (t *ReleasedAppTemplateBindingAttachment) Validate() error {
	if t.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if t.AppID <= 0 {
		return errors.New("invalid attachment app id")
	}

	return nil
}
