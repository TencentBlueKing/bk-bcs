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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// TemplateSpace 模版空间
type TemplateSpace struct {
	ID         uint32                   `json:"id" gorm:"primaryKey"`
	Spec       *TemplateSpaceSpec       `json:"spec" gorm:"embedded"`
	Attachment *TemplateSpaceAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision                `json:"revision" gorm:"embedded"`
}

// TableName is the template space's database table name.
func (t *TemplateSpace) TableName() string {
	return "template_spaces"
}

// AppID AuditRes interface
func (t *TemplateSpace) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (t *TemplateSpace) ResID() uint32 {
	return t.ID
}

// ResType AuditRes interface
func (t *TemplateSpace) ResType() string {
	return "template_space"
}

// ValidateCreate validate template space is valid or not when create it.
func (t *TemplateSpace) ValidateCreate(kit *kit.Kit) error {
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

// ValidateUpdate validate template space is valid or not when update it.
func (t *TemplateSpace) ValidateUpdate(kit *kit.Kit) error {

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

// ValidateDelete validate the template space's info when delete it.
func (t *TemplateSpace) ValidateDelete() error {
	if t.ID <= 0 {
		return errors.New("template space id should be set")
	}

	if t.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	return nil
}

// TemplateSpaceSpec defines all the specifics for template space set by user.
type TemplateSpaceSpec struct {
	Name string `json:"name" gorm:"column:name"`
	Memo string `json:"memo" gorm:"column:memo"`
}

// ValidateCreate validate template space spec when it is created.
func (t *TemplateSpaceSpec) ValidateCreate(kit *kit.Kit) error {
	if err := validator.ValidateName(kit, t.Name); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate template space spec when it is updated.
func (t *TemplateSpaceSpec) ValidateUpdate(kit *kit.Kit) error {
	if err := validator.ValidateMemo(kit, t.Memo, false); err != nil {
		return err
	}

	return nil
}

// TemplateSpaceAttachment defines the template space attachments.
type TemplateSpaceAttachment struct {
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
}

// Validate whether template space attachment is valid or not.
func (t *TemplateSpaceAttachment) Validate() error {
	if t.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	return nil
}
