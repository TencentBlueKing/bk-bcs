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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"bscp.io/pkg/dal/types"
)

// AppTemplateBinding 应用模版绑定
type AppTemplateBinding struct {
	ID         uint32                        `json:"id" gorm:"primaryKey"`
	Spec       *AppTemplateBindingSpec       `json:"spec" gorm:"embedded"`
	Attachment *AppTemplateBindingAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision                     `json:"revision" gorm:"embedded"`
}

// TableName is the AppTemplateBinding's database table name.
func (t *AppTemplateBinding) TableName() string {
	return "app_template_bindings"
}

// AppID AuditRes interface
func (t *AppTemplateBinding) AppID() uint32 {
	return t.Attachment.AppID
}

// ResID AuditRes interface
func (t *AppTemplateBinding) ResID() uint32 {
	return t.ID
}

// ResType AuditRes interface
func (t *AppTemplateBinding) ResType() string {
	return "app_template_binding"
}

// ValidateCreate validate AppTemplateBinding is valid or not when create it.
func (t *AppTemplateBinding) ValidateCreate() error {
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

	if err := t.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate AppTemplateBinding is valid or not when update it.
func (t *AppTemplateBinding) ValidateUpdate() error {

	if t.ID <= 0 {
		return errors.New("id should be set")
	}

	if t.Spec != nil {
		if err := t.Spec.ValidateUpdate(); err != nil {
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

// ValidateDelete validate the AppTemplateBinding's info when delete it.
func (t *AppTemplateBinding) ValidateDelete() error {
	if t.ID <= 0 {
		return errors.New("AppTemplateBinding id should be set")
	}

	if t.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	return nil
}

// AppTemplateBindingSpec defines all the specifics for AppTemplateBinding set by user.
type AppTemplateBindingSpec struct {
	TemplateSpaceIDs   types.Uint32Slice `json:"template_space_ids" gorm:"column:template_space_ids;type:json"`
	TemplateSetIDs     types.Uint32Slice `json:"template_set_ids" gorm:"column:template_set_ids;type:json"`
	TemplateIDs        types.Uint32Slice `json:"template_ids" gorm:"column:template_ids;type:json"`
	TemplateReleaseIDs types.Uint32Slice `json:"template_release_ids" gorm:"column:template_release_ids;type:json"`
	Bindings           TemplateBindings  `json:"bindings" gorm:"column:bindings;type:json"`
}

type TemplateBindings []*TemplateBinding

type TemplateBinding struct {
	TemplateSetID      uint32   `json:"template_set_id"`
	TemplateReleaseIDs []uint32 `json:"template_release_ids"`
}

// Value implements the driver.Valuer interface
// See gorm document about customizing data types: https://gorm.io/docs/data_types.html
func (u TemplateBindings) Value() (driver.Value, error) {
	// Convert the TemplateBinding to a JSON-encoded string
	data, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

// Scan implements the sql.Scanner interface
// See gorm document about customizing data types: https://gorm.io/docs/data_types.html
func (u *TemplateBindings) Scan(value interface{}) error {
	// Check if the value is nil
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// The value is of type []byte (MySQL driver representation for JSON columns)
		// Unmarshal the JSON-encoded value to TemplateBinding
		err := json.Unmarshal(v, u)
		if err != nil {
			return err
		}
	case string:
		// The value is of type string (fallback for older versions of MySQL driver)
		// Unmarshal the JSON-encoded value to TemplateBinding
		err := json.Unmarshal([]byte(v), u)
		if err != nil {
			return err
		}
	default:
		return errors.New("unsupported Scan type for TemplateBinding")
	}

	return nil
}

// AppTemplateBindingType is the type of AppTemplateBinding
type AppTemplateBindingType string

// ValidateCreate validate AppTemplateBinding spec when it is created.
func (t *AppTemplateBindingSpec) ValidateCreate() error {
	return validateBindings(t.Bindings)
}

// ValidateUpdate validate AppTemplateBinding spec when it is updated.
func (t *AppTemplateBindingSpec) ValidateUpdate() error {
	return validateBindings(t.Bindings)
}

func validateBindings(bindings TemplateBindings) error {
	if len(bindings) == 0 {
		return errors.New("bindings can't be empty")
	}
	for _, b := range bindings {
		if b.TemplateSetID <= 0 {
			return fmt.Errorf("invalid template set id of bindings member: %d", b.TemplateSetID)
		}
		if len(b.TemplateReleaseIDs) == 0 {
			return errors.New("template release ids of bindings member can't be empty")
		}
		for _, id := range b.TemplateReleaseIDs {
			if id <= 0 {
				return fmt.Errorf("invalid template release id of bindings member: %d", id)
			}
		}
	}
	return nil
}

// AppTemplateBindingAttachment defines the AppTemplateBinding attachments.
type AppTemplateBindingAttachment struct {
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
	AppID uint32 `json:"app_id" gorm:"column:app_id"`
}

// Validate whether AppTemplateBinding attachment is valid or not.
func (t *AppTemplateBindingAttachment) Validate() error {
	if t.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if t.AppID <= 0 {
		return errors.New("invalid attachment app id")
	}

	return nil
}
