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
)

// ReleasedAppTemplate 已生成版本服务的模版
type ReleasedAppTemplate struct {
	ID         uint32                         `json:"id" gorm:"primaryKey"`
	Spec       *ReleasedAppTemplateSpec       `json:"spec" gorm:"embedded"`
	Attachment *ReleasedAppTemplateAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision                      `json:"revision" gorm:"embedded"`
}

// TableName is the ReleasedAppTemplate's database table name.
func (r *ReleasedAppTemplate) TableName() string {
	return "released_app_templates"
}

// AppID AuditRes interface
func (r *ReleasedAppTemplate) AppID() uint32 {
	return r.Attachment.AppID
}

// ResID AuditRes interface
func (r *ReleasedAppTemplate) ResID() uint32 {
	return r.ID
}

// ResType AuditRes interface
func (r *ReleasedAppTemplate) ResType() string {
	return "released_app_template"
}

// RatiList is released app templates
type RatiList []*ReleasedAppTemplate

// AppID AuditRes interface
func (rs RatiList) AppID() uint32 {
	if len(rs) > 0 {
		return rs[0].Attachment.AppID
	}
	return 0
}

// ResID AuditRes interface
func (rs RatiList) ResID() uint32 {
	if len(rs) > 0 {
		return rs[0].ID
	}
	return 0
}

// ResType AuditRes interface
func (rs RatiList) ResType() string {
	return "released_app_template"
}

// ValidateCreate validate ReleasedAppTemplate is valid or not when create ir.
func (r *ReleasedAppTemplate) ValidateCreate() error {
	if r.ID > 0 {
		return errors.New("id should not be set")
	}

	if r.Spec == nil {
		return errors.New("spec not set")
	}

	if err := r.Spec.ValidateCreate(); err != nil {
		return err
	}

	if r.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := r.Attachment.Validate(); err != nil {
		return err
	}

	if r.Revision == nil {
		return errors.New("revision not set")
	}

	return nil
}

// ReleasedAppTemplateSpec defines all the specifics for ReleasedAppTemplate set by user.
type ReleasedAppTemplateSpec struct {
	ReleaseID            uint32 `json:"release_id" gorm:"column:release_id"`
	TemplateSpaceID      uint32 `json:"template_space_id" gorm:"column:template_space_id"`
	TemplateSpaceName    string `json:"template_space_name" gorm:"column:template_space_name"`
	TemplateSetID        uint32 `json:"template_set_id" gorm:"column:template_set_id"`
	TemplateSetName      string `json:"template_set_name" gorm:"column:template_set_name"`
	TemplateID           uint32 `json:"template_id" gorm:"column:template_id"`
	Name                 string `json:"name" gorm:"column:name"`
	Path                 string `json:"path" gorm:"column:path"`
	TemplateRevisionID   uint32 `json:"template_revision_id" gorm:"column:template_revision_id"`
	IsLatest             bool   `json:"is_latest" gorm:"column:is_latest"`
	TemplateRevisionName string `json:"template_revision_name" gorm:"column:template_revision_name"`
	TemplateRevisionMemo string `json:"template_revision_memo" gorm:"column:template_revision_memo"`
	FileType             string `json:"file_type" gorm:"column:file_type"`
	FileMode             string `json:"file_mode" gorm:"column:file_mode"`
	User                 string `json:"user" gorm:"column:user"`
	UserGroup            string `json:"user_group" gorm:"column:user_group"`
	Privilege            string `json:"privilege" gorm:"column:privilege"`
	Signature            string `json:"signature" gorm:"column:signature"`
	ByteSize             uint64 `json:"byte_size" gorm:"column:byte_size"`
	OriginSignature      string `json:"origin_signature" gorm:"column:origin_signature"`
	OriginByteSize       uint64 `json:"origin_byte_size" gorm:"column:origin_byte_size"`
}

// ValidateCreate validate ReleasedAppTemplate spec when it is created.
func (r *ReleasedAppTemplateSpec) ValidateCreate() error {
	if r.ReleaseID <= 0 {
		return errors.New("invalid release id")
	}

	if r.TemplateSpaceID <= 0 {
		return errors.New("invalid template space id")
	}

	if r.TemplateSetID <= 0 {
		return errors.New("invalid template set id")
	}

	if r.TemplateID <= 0 {
		return errors.New("invalid template id")
	}

	if r.TemplateRevisionID <= 0 {
		return errors.New("invalid template revision id")
	}

	if r.TemplateSpaceName == "" {
		return errors.New("template space name is empty")
	}

	if r.TemplateSetName == "" {
		return errors.New("template set name is empty")
	}

	if r.TemplateRevisionName == "" {
		return errors.New("template revision name is empty")
	}

	if r.Name == "" {
		return errors.New("template config name is empty")
	}

	if r.Path == "" {
		return errors.New("template config path is empty")
	}

	return nil
}

// ReleasedAppTemplateAttachment defines the ReleasedAppTemplate attachments.
type ReleasedAppTemplateAttachment struct {
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
	AppID uint32 `json:"app_id" gorm:"column:app_id"`
}

// Validate whether ReleasedAppTemplate attachment is valid or not.
func (r *ReleasedAppTemplateAttachment) Validate() error {
	if r.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if r.AppID <= 0 {
		return errors.New("invalid attachment app id")
	}

	return nil
}
