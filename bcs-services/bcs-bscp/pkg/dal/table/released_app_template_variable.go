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

	"bscp.io/pkg/kit"
)

// ReleasedAppTemplateVariable 已生成版本服务的应用模版变量
type ReleasedAppTemplateVariable struct {
	ID         uint32                                 `json:"id" gorm:"primaryKey"`
	Spec       *ReleasedAppTemplateVariableSpec       `json:"spec" gorm:"embedded"`
	Attachment *ReleasedAppTemplateVariableAttachment `json:"attachment" gorm:"embedded"`
	Revision   *CreatedRevision                       `json:"revision" gorm:"embedded"`
}

// TableName is the ReleasedAppTemplateVariable's database table name.
func (t *ReleasedAppTemplateVariable) TableName() string {
	return "released_app_template_variables"
}

// AppID AuditRes interface
func (t *ReleasedAppTemplateVariable) AppID() uint32 {
	return t.Attachment.AppID
}

// ResID AuditRes interface
func (t *ReleasedAppTemplateVariable) ResID() uint32 {
	return t.ID
}

// ResType AuditRes interface
func (t *ReleasedAppTemplateVariable) ResType() string {
	return "released_app_template_variable"
}

// ValidateCreate validate ReleasedAppTemplateVariable is valid or not when created.
func (t *ReleasedAppTemplateVariable) ValidateCreate(kit *kit.Kit) error {
	if t.Spec != nil {
		if err := t.Spec.ValidateCreate(kit); err != nil {
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

	if err := t.Revision.Validate(); err != nil {
		return err
	}

	return nil
}

// ReleasedAppTemplateVariableSpec defines all the specifics for ReleasedAppTemplateVariable set by user.
type ReleasedAppTemplateVariableSpec struct {
	ReleaseID uint32       `json:"release_id" gorm:"column:release_id"`
	Variables AppVariables `json:"variables" gorm:"column:variables;type:json;default:'[]'"`
}

// ValidateCreate validate ReleasedAppTemplateVariable spec when it is created.
func (t *ReleasedAppTemplateVariableSpec) ValidateCreate(kit *kit.Kit) error {
	for _, v := range t.Variables {
		if err := v.ValidateCreate(kit); err != nil {
			return err
		}
	}
	return nil
}

// ReleasedAppTemplateVariableAttachment defines the ReleasedAppTemplateVariable attachments.
type ReleasedAppTemplateVariableAttachment struct {
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
	AppID uint32 `json:"app_id" gorm:"column:app_id"`
}

// Validate whether ReleasedAppTemplateVariable attachment is valid or not.
func (t *ReleasedAppTemplateVariableAttachment) Validate() error {
	if t.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if t.AppID <= 0 {
		return errors.New("invalid attachment app id")
	}

	return nil
}
