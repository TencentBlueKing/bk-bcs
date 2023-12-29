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
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/types"
)

// AppTemplateBinding 未命名版本服务的模版绑定
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
//
//nolint:lll
type AppTemplateBindingSpec struct {
	TemplateSpaceIDs    types.Uint32Slice `json:"template_space_ids" gorm:"column:template_space_ids;type:json;default:'[]'"`
	TemplateSetIDs      types.Uint32Slice `json:"template_set_ids" gorm:"column:template_set_ids;type:json;default:'[]'"`
	TemplateIDs         types.Uint32Slice `json:"template_ids" gorm:"column:template_ids;type:json;default:'[]'"`
	TemplateRevisionIDs types.Uint32Slice `json:"template_revision_ids" gorm:"column:template_revision_ids;type:json;default:'[]'"`
	LatestTemplateIDs   types.Uint32Slice `json:"latest_template_ids" gorm:"column:latest_template_ids;type:json;default:'[]'"`
	Bindings            TemplateBindings  `json:"bindings" gorm:"column:bindings;type:json;default:'[]'"`
}

// TemplateBindings is []*TemplateBinding
type TemplateBindings []*TemplateBinding

// TemplateBinding is relation between template set id and template revisions
type TemplateBinding struct {
	TemplateSetID     uint32                     `json:"template_set_id"`
	TemplateRevisions []*TemplateRevisionBinding `json:"template_revisions"`
}

// TemplateRevisionBinding is template revision binding
type TemplateRevisionBinding struct {
	TemplateID         uint32 `json:"template_id"`
	TemplateRevisionID uint32 `json:"template_revision_id"`
	IsLatest           bool   `json:"is_latest"`
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

// ValidateCreate validate AppTemplateBinding spec when it is created.
func (t *AppTemplateBindingSpec) ValidateCreate() error {
	return validateBindingsCreate(t.Bindings)
}

// ValidateUpdate validate AppTemplateBinding spec when it is updated.
func (t *AppTemplateBindingSpec) ValidateUpdate() error {
	return validateBindingsUpdate(t.Bindings)
}

func validateBindingsCreate(bindings TemplateBindings) error {
	if len(bindings) == 0 {
		return errors.New("bindings can't be empty")
	}
	for _, b := range bindings {
		if b.TemplateSetID <= 0 {
			return fmt.Errorf("invalid template set id of bindings member: %d", b.TemplateSetID)
		}
	}
	return nil
}

func validateBindingsUpdate(bindings TemplateBindings) error {
	for _, b := range bindings {
		if b.TemplateSetID <= 0 {
			return fmt.Errorf("invalid template set id of bindings member: %d", b.TemplateSetID)
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
