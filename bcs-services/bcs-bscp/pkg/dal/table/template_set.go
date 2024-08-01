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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// TemplateSet 模版套餐
type TemplateSet struct {
	ID         uint32                 `json:"id" gorm:"primaryKey"`
	Spec       *TemplateSetSpec       `json:"spec" gorm:"embedded"`
	Attachment *TemplateSetAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision              `json:"revision" gorm:"embedded"`
}

// TableName is the template set's database table name.
func (t *TemplateSet) TableName() string {
	return "template_sets"
}

// AppID AuditRes interface
func (t *TemplateSet) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (t *TemplateSet) ResID() uint32 {
	return t.ID
}

// ResType AuditRes interface
func (t *TemplateSet) ResType() string {
	return "template_set"
}

// ValidateCreate validate template set is valid or not when create it.
func (t *TemplateSet) ValidateCreate(kit *kit.Kit) error {
	if t.ID > 0 {
		return errors.New("id should not be set")
	}

	if t.Spec == nil {
		return errors.New("spec not set")
	}

	if err := t.Spec.ValidateCreate(kit); err != nil {
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

	if err := t.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate template set is valid or not when update it.
func (t *TemplateSet) ValidateUpdate(kit *kit.Kit) error {

	if t.ID <= 0 {
		return errors.New("id should be set")
	}

	if t.Spec != nil {
		if err := t.Spec.ValidateUpdate(kit); err != nil {
			return err
		}
	}

	if t.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	if t.Revision == nil {
		return errors.New("revision not set")
	}

	if err := t.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the template set's info when delete it.
func (t *TemplateSet) ValidateDelete() error {
	if t.ID <= 0 {
		return errors.New("template set id should be set")
	}

	if t.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	return nil
}

// TemplateSetSpec defines all the specifics for template set set by user.
type TemplateSetSpec struct {
	Name        string            `json:"name" gorm:"column:name"`
	Memo        string            `json:"memo" gorm:"column:memo"`
	TemplateIDs types.Uint32Slice `json:"template_ids" gorm:"column:template_ids;type:json;default:'[]'"`
	Public      bool              `json:"public" gorm:"column:public"`
	BoundApps   types.Uint32Slice `json:"bound_apps" gorm:"column:bound_apps;type:json;default:'[]'"`
}

// ValidateCreate validate template set spec when it is created.
func (t *TemplateSetSpec) ValidateCreate(kit *kit.Kit) error {
	if err := validator.ValidateName(kit, t.Name); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate template set spec when it is updated.
func (t *TemplateSetSpec) ValidateUpdate(kit *kit.Kit) error {
	if err := validator.ValidateMemo(kit, t.Memo, false); err != nil {
		return err
	}

	return nil
}

// TemplateSetAttachment defines the template set attachments.
type TemplateSetAttachment struct {
	BizID           uint32 `json:"biz_id" gorm:"column:biz_id"`
	TemplateSpaceID uint32 `json:"template_space_id" gorm:"column:template_space_id"`
}

// Validate whether TemplateSet attachment is valid or not.
func (t *TemplateSetAttachment) Validate() error {
	if t.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if t.TemplateSpaceID <= 0 {
		return errors.New("invalid attachment template space id")
	}

	return nil
}
