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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// TemplateVariable 模版变量
type TemplateVariable struct {
	ID         uint32                      `json:"id" gorm:"primaryKey"`
	Spec       *TemplateVariableSpec       `json:"spec" gorm:"embedded"`
	Attachment *TemplateVariableAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision                   `json:"revision" gorm:"embedded"`
}

// TableName is the template variable's database table name.
func (t *TemplateVariable) TableName() string {
	return "template_variables"
}

// AppID AuditRes interface
func (t *TemplateVariable) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (t *TemplateVariable) ResID() uint32 {
	return t.ID
}

// ResType AuditRes interface
func (t *TemplateVariable) ResType() string {
	return "template_variable"
}

// ValidateCreate validate template variable is valid or not when create it.
func (t *TemplateVariable) ValidateCreate(kit *kit.Kit) error {
	if t.ID > 0 {
		return errf.ErrWithIDF(kit)
	}

	if t.Spec == nil {
		return errf.ErrNoSpecF(kit)
	}

	if err := t.Spec.ValidateCreate(kit); err != nil {
		return err
	}

	if t.Attachment == nil {
		return errf.ErrNoAttachmentF(kit)
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	if t.Revision == nil {
		return errf.ErrNoRevisionF(kit)
	}

	if err := t.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate template variable is valid or not when update it.
func (t *TemplateVariable) ValidateUpdate(kit *kit.Kit) error {

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

// ValidateDelete validate the template variable's info when delete it.
func (t *TemplateVariable) ValidateDelete() error {
	if t.ID <= 0 {
		return errors.New("template variable id should be set")
	}

	if t.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	return nil
}

// TemplateVariableSpec defines all the specifics for template variable set by user.
type TemplateVariableSpec struct {
	Name       string       `json:"name" gorm:"column:name"`
	Type       VariableType `json:"type" gorm:"column:type"`
	DefaultVal string       `json:"default_val" gorm:"column:default_val"`
	Memo       string       `json:"memo" gorm:"column:memo"`
}

// ValidateCreate validate template variable spec when it is created.
func (t *TemplateVariableSpec) ValidateCreate(kit *kit.Kit) error {
	if err := t.ValidateDefaultVal(kit); err != nil {
		return err
	}

	if err := validator.ValidateVariableName(kit, t.Name); err != nil {
		return err
	}

	if err := t.Type.Validate(kit); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate template variable spec when it is updated.
func (t *TemplateVariableSpec) ValidateUpdate(kit *kit.Kit) error {
	if err := t.ValidateDefaultVal(kit); err != nil {
		return err
	}

	if err := validator.ValidateMemo(kit, t.Memo, false); err != nil {
		return err
	}

	return nil
}

// ValidateDefaultVal validate template variable default value.
func (t *TemplateVariableSpec) ValidateDefaultVal(kit *kit.Kit) error {
	if t.Type == NumberVar && !tools.IsNumber(t.DefaultVal) {
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "default_val %s is not a number type", t.DefaultVal))
	}

	return nil
}

// TemplateVariableAttachment defines the template variable attachments.
type TemplateVariableAttachment struct {
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
}

// Validate whether template variable attachment is valid or not.
func (t *TemplateVariableAttachment) Validate() error {
	if t.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	return nil
}

const (
	// StringVar is string type variable
	StringVar VariableType = "string"
	// NumberVar is number type variable
	NumberVar VariableType = "number"
)

// VariableType is template variable type
type VariableType string

// Validate the file format is supported or not.
func (t VariableType) Validate(kit *kit.Kit) error {
	switch t {
	case StringVar:
	case NumberVar:
	default:
		return errf.Errorf(errf.InvalidArgument, i18n.T(kit, "unsupported variable type: %s", t))
	}

	return nil
}
