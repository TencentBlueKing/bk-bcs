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
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// AppTemplateVariable 未命名版本服务的模版变量
type AppTemplateVariable struct {
	ID         uint32                         `json:"id" gorm:"primaryKey"`
	Spec       *AppTemplateVariableSpec       `json:"spec" gorm:"embedded"`
	Attachment *AppTemplateVariableAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision                      `json:"revision" gorm:"embedded"`
}

// TableName is the AppTemplateVariable's database table name.
func (t *AppTemplateVariable) TableName() string {
	return "app_template_variables"
}

// AppID AuditRes interface
func (t *AppTemplateVariable) AppID() uint32 {
	return t.Attachment.AppID
}

// ResID AuditRes interface
func (t *AppTemplateVariable) ResID() uint32 {
	return t.ID
}

// ResType AuditRes interface
func (t *AppTemplateVariable) ResType() string {
	return "app_template_variable"
}

// ValidateUpsert validate AppTemplateVariable is valid or not when create or update it.
func (t *AppTemplateVariable) ValidateUpsert(kit *kit.Kit) error {
	if t.Spec != nil {
		if err := t.Spec.ValidateUpsert(kit); err != nil {
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

// AppTemplateVariableSpec defines all the specifics for AppTemplateVariable set by user.
type AppTemplateVariableSpec struct {
	Variables AppVariables `json:"variables" gorm:"column:variables;type:json;default:'[]'"`
}

// AppVariables is []*AppVariable
type AppVariables []*TemplateVariableSpec

// Value implements the driver.Valuer interface
// See gorm document about customizing data types: https://gorm.io/docs/data_types.html
func (u AppVariables) Value() (driver.Value, error) {
	// Convert the AppVariables to a JSON-encoded string
	data, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan implements the sql.Scanner interface
// See gorm document about customizing data types: https://gorm.io/docs/data_types.html
func (u *AppVariables) Scan(value interface{}) error {
	// Check if the value is nil
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// The value is of type []byte (MySQL driver representation for JSON columns)
		// Unmarshal the JSON-encoded value to AppVariables
		err := json.Unmarshal(v, u)
		if err != nil {
			return err
		}
	case string:
		// The value is of type string (fallback for older versions of MySQL driver)
		// Unmarshal the JSON-encoded value to AppVariables
		err := json.Unmarshal([]byte(v), u)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported Scan type for AppVariables")
	}

	return nil
}

// ValidateUpsert validate AppTemplateVariable spec when it is created or updated.
func (t *AppTemplateVariableSpec) ValidateUpsert(kit *kit.Kit) error {
	for _, v := range t.Variables {
		if err := v.ValidateCreate(kit); err != nil {
			return err
		}
	}
	return nil
}

// AppTemplateVariableAttachment defines the AppTemplateVariable attachments.
type AppTemplateVariableAttachment struct {
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
	AppID uint32 `json:"app_id" gorm:"column:app_id"`
}

// Validate whether AppTemplateVariable attachment is valid or not.
func (t *AppTemplateVariableAttachment) Validate() error {
	if t.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if t.AppID <= 0 {
		return errors.New("invalid attachment app id")
	}

	return nil
}
