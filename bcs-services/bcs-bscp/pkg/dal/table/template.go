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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// Template 模版
type Template struct {
	ID         uint32              `json:"id" gorm:"primaryKey"`
	Spec       *TemplateSpec       `json:"spec" gorm:"embedded"`
	Attachment *TemplateAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision           `json:"revision" gorm:"embedded"`
}

// TableName is the template's database table name.
func (t *Template) TableName() string {
	return "templates"
}

// AppID AuditRes interface
func (t *Template) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (t *Template) ResID() uint32 {
	return t.ID
}

// ResType AuditRes interface
func (t *Template) ResType() string {
	return "template"
}

// ValidateCreate validate template is valid or not when create it.
func (t *Template) ValidateCreate(kit *kit.Kit) error {
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

// ValidateUpdate validate template is valid or not when update it.
func (t *Template) ValidateUpdate(kit *kit.Kit) error {

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

// ValidateDelete validate the template's info when delete it.
func (t *Template) ValidateDelete() error {
	if t.ID <= 0 {
		return errors.New("template id should be set")
	}

	if t.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	return nil
}

// TemplateSpec defines all the specifics for template set by user.
type TemplateSpec struct {
	Name string `json:"name" gorm:"column:name"`
	Path string `json:"path" gorm:"column:path"`
	Memo string `json:"memo" gorm:"column:memo"`
}

// ValidateCreate validate template spec when it is created.
func (t *TemplateSpec) ValidateCreate(kit *kit.Kit) error {
	if err := validator.ValidateFileName(kit, t.Name); err != nil {
		return err
	}

	if err := ValidatePath(kit, t.Path, Unix); err != nil {
		return fmt.Errorf("%s err: %v", t.Path, err)
	}

	return nil
}

// ValidateUpdate validate template spec when it is updated.
func (t *TemplateSpec) ValidateUpdate(kit *kit.Kit) error {
	if err := validator.ValidateMemo(kit, t.Memo, false); err != nil {
		return err
	}

	return nil
}

// TemplateAttachment defines the template attachments.
type TemplateAttachment struct {
	BizID           uint32 `json:"biz_id" gorm:"column:biz_id"`
	TemplateSpaceID uint32 `json:"template_space_id" gorm:"column:template_space_id"`
}

// Validate whether template attachment is valid or not.
func (t *TemplateAttachment) Validate() error {
	if t.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if t.TemplateSpaceID <= 0 {
		return errors.New("invalid attachment template space id")
	}

	return nil
}
