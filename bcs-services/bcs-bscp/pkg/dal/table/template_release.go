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

// TemplateRelease is template config item
type TemplateRelease struct {
	ID         uint32                     `json:"id" gorm:"primaryKey"`
	Spec       *TemplateReleaseSpec       `json:"spec" gorm:"embedded"`
	Attachment *TemplateReleaseAttachment `json:"attachment" gorm:"embedded"`
	Revision   *CreatedRevision           `json:"revision" gorm:"embedded"`
}

// TableName is the TemplateRelease's database table name.
func (t *TemplateRelease) TableName() string {
	return "template_releases"
}

// AppID AuditRes interface
func (t *TemplateRelease) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (t *TemplateRelease) ResID() uint32 {
	return t.ID
}

// ResType AuditRes interface
func (t *TemplateRelease) ResType() string {
	return "template_release"
}

// ValidateCreate validate TemplateRelease is valid or not when create it.
func (t *TemplateRelease) ValidateCreate() error {
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

	if err := t.Revision.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the TemplateRelease's info when delete it.
func (t *TemplateRelease) ValidateDelete() error {
	if t.ID <= 0 {
		return errors.New("TemplateRelease id should be set")
	}

	if t.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	return nil
}

// TemplateReleaseSpec defines all the specifics for TemplateRelease set by user.
type TemplateReleaseSpec struct {
	ReleaseName string          `json:"release_name" gorm:"column:release_name"`
	ReleaseMemo string          `json:"release_memo" gorm:"column:release_memo"`
	Name        string          `json:"name" gorm:"column:name"`
	Path        string          `json:"path" gorm:"column:path"`
	FileType    FileFormat      `json:"file_type" gorm:"column:file_type"`
	FileMode    FileMode        `json:"file_mode" gorm:"column:file_mode"`
	Permission  *FilePermission `json:"permission" gorm:"embedded"`
	ContentSpec *ContentSpec    `json:"content" gorm:"embedded"`
}

// TemplateReleaseType is the type of TemplateRelease
type TemplateReleaseType string

// ValidateCreate validate TemplateRelease spec when it is created.
func (t *TemplateReleaseSpec) ValidateCreate() error {
	if err := validator.ValidateCfgItemName(t.Name); err != nil {
		return err
	}

	if err := t.FileType.Validate(); err != nil {
		return err
	}

	if err := t.FileMode.Validate(); err != nil {
		return err
	}

	if err := ValidatePath(t.Path, Unix); err != nil {
		return err
	}

	if err := t.Permission.Validate(t.FileMode); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate TemplateRelease spec when it is updated.
func (t *TemplateReleaseSpec) ValidateUpdate() error {
	if err := validator.ValidateMemo(t.ReleaseMemo, false); err != nil {
		return err
	}

	return nil
}

// TemplateReleaseAttachment defines the TemplateRelease attachments.
type TemplateReleaseAttachment struct {
	BizID           uint32 `json:"biz_id" gorm:"column:biz_id"`
	TemplateSpaceID uint32 `json:"template_space_id" gorm:"column:template_space_id"`
	TemplateID      uint32 `json:"template_id" gorm:"column:template_id"`
}

// Validate whether TemplateRelease attachment is valid or not.
func (t *TemplateReleaseAttachment) Validate() error {
	if t.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if t.TemplateSpaceID <= 0 {
		return errors.New("invalid attachment template space id")
	}

	if t.TemplateID <= 0 {
		return errors.New("invalid attachment template id")
	}

	return nil
}
