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

// TemplateRevision 模版版本
type TemplateRevision struct {
	ID         uint32                      `json:"id" gorm:"primaryKey"`
	Spec       *TemplateRevisionSpec       `json:"spec" gorm:"embedded"`
	Attachment *TemplateRevisionAttachment `json:"attachment" gorm:"embedded"`
	Revision   *CreatedRevision            `json:"revision" gorm:"embedded"`
}

// TableName is the template revision's database table name.
func (t *TemplateRevision) TableName() string {
	return "template_revisions"
}

// AppID AuditRes interface
func (t *TemplateRevision) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (t *TemplateRevision) ResID() uint32 {
	return t.ID
}

// ResType AuditRes interface
func (t *TemplateRevision) ResType() string {
	return "template_revision"
}

// ValidateCreate validate template revision is valid or not when create it.
func (t *TemplateRevision) ValidateCreate(kit *kit.Kit) error {
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

	if err := t.Revision.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateDelete validate the template revision's info when delete it.
func (t *TemplateRevision) ValidateDelete() error {
	if t.ID <= 0 {
		return errors.New("template revision id should be set")
	}

	if t.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if err := t.Attachment.Validate(); err != nil {
		return err
	}

	return nil
}

// TemplateRevisionSpec defines all the specifics for template revision set by user.
type TemplateRevisionSpec struct {
	RevisionName string          `json:"revision_name" gorm:"column:revision_name"`
	RevisionMemo string          `json:"revision_memo" gorm:"column:revision_memo"`
	Name         string          `json:"name" gorm:"column:name"`
	Path         string          `json:"path" gorm:"column:path"`
	FileType     FileFormat      `json:"file_type" gorm:"column:file_type"`
	FileMode     FileMode        `json:"file_mode" gorm:"column:file_mode"`
	Permission   *FilePermission `json:"permission" gorm:"embedded"`
	ContentSpec  *ContentSpec    `json:"content" gorm:"embedded"`
}

// ValidateCreate validate template revision spec when it is created.
func (t *TemplateRevisionSpec) ValidateCreate(kit *kit.Kit) error {
	if err := validator.ValidateReleaseName(kit, t.RevisionName); err != nil {
		return err
	}

	if err := validator.ValidateMemo(kit, t.RevisionMemo, false); err != nil {
		return err
	}

	if err := validator.ValidateFileName(kit, t.Name); err != nil {
		return err
	}

	if err := t.FileType.Validate(kit); err != nil {
		return err
	}

	if err := t.FileMode.Validate(kit); err != nil {
		return err
	}

	if err := ValidatePath(kit, t.Path, Unix); err != nil {
		return err
	}

	if err := t.Permission.Validate(kit, t.FileMode); err != nil {
		return err
	}

	if err := t.ContentSpec.Validate(kit); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate validate template revision spec when it is updated.
func (t *TemplateRevisionSpec) ValidateUpdate(kit *kit.Kit) error {
	if err := validator.ValidateMemo(kit, t.RevisionMemo, false); err != nil {
		return err
	}

	return nil
}

// TemplateRevisionAttachment defines the template revision attachments.
type TemplateRevisionAttachment struct {
	BizID           uint32 `json:"biz_id" gorm:"column:biz_id"`
	TemplateSpaceID uint32 `json:"template_space_id" gorm:"column:template_space_id"`
	TemplateID      uint32 `json:"template_id" gorm:"column:template_id"`
}

// Validate whether template revision attachment is valid or not.
func (t *TemplateRevisionAttachment) Validate() error {
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
