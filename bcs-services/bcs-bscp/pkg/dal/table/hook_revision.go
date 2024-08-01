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
	"strings"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/validator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// HookRevision 脚本版本
type HookRevision struct {
	// ID is an auto-increased value, which is a unique identity of a hook.
	ID uint32 `db:"id" json:"id"`

	Spec       *HookRevisionSpec       `json:"spec" gorm:"embedded"`
	Attachment *HookRevisionAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision               `json:"revision" gorm:"embedded"`
}

// HookRevisionSpec defines all the specifics for hook set by user.
type HookRevisionSpec struct {
	Name    string             `json:"name" gorm:"column:name"`
	State   HookRevisionStatus `json:"state" gorm:"column:state"`
	Content string             `json:"content" gorm:"column:content"`
	Memo    string             `json:"memo" gorm:"column:memo"`
}

// HookRevisionAttachment defines the hook attachments.
type HookRevisionAttachment struct {
	BizID  uint32 `json:"biz_id" gorm:"column:biz_id"`
	HookID uint32 `json:"hook_id" gorm:"column:hook_id"`
}

// TableName is the hook's database table name.
func (r *HookRevision) TableName() Name {
	return "hook_revisions"
}

// AppID AuditRes interface
func (r *HookRevision) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (r *HookRevision) ResID() uint32 {
	return r.ID
}

// ResType AuditRes interface
func (r *HookRevision) ResType() string {
	return "hook_revisions"
}

// ValidateCreate validate hook is valid or not when create it.
func (r *HookRevision) ValidateCreate(kit *kit.Kit) error {

	if r.Spec == nil {
		return errors.New("spec not set")
	}

	if err := r.Spec.ValidateCreate(kit); err != nil {
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

	if err := r.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateCreate validate spec when created.
func (s *HookRevisionSpec) ValidateCreate(kit *kit.Kit) error {

	if err := validator.ValidateReleaseName(kit, s.Name); err != nil {
		return err
	}

	if err := validator.ValidateMemo(kit, s.Memo, false); err != nil {
		return err
	}

	if strings.Trim(strings.Trim(s.Content, ""), "\n") == "" {
		return errors.New("content should not be empty")
	}

	return nil
}

// ValidateUpdate validate spec when updated.
func (s *HookRevisionSpec) ValidateUpdate(kit *kit.Kit) error {

	if err := validator.ValidateReleaseName(kit, s.Name); err != nil {
		return err
	}

	if err := validator.ValidateMemo(kit, s.Memo, false); err != nil {
		return err
	}

	if strings.Trim(strings.Trim(s.Content, ""), "\n") == "" {
		return errors.New("content should not be empty")
	}

	return nil
}

// Validate validate Attachment.
func (a HookRevisionAttachment) Validate() error {

	if a.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if a.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	return nil
}

// ValidateDelete validate the hook revision info when delete it.
func (r HookRevision) ValidateDelete() error {
	if r.ID <= 0 {
		return errors.New("hook revision id should be set")
	}

	if r.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if r.Attachment.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	return nil
}

// ValidateDeleteByHookID validate the hook revision info when delete it.
func (r HookRevision) ValidateDeleteByHookID() error {

	if r.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if r.Attachment.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	return nil
}

// ValidatePublish validate the Publish
func (r HookRevision) ValidatePublish() error {

	if r.ID <= 0 {
		return errors.New("hook revision id should be set")
	}

	if r.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if r.Attachment.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	return nil
}

// ValidateUpdate validate the update
func (r HookRevision) ValidateUpdate(kit *kit.Kit) error {

	if r.ID <= 0 {
		return errors.New("hook revision id should be set")
	}

	if r.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if r.Attachment.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	if r.Spec == nil {
		return errors.New("spec not set")
	}

	if err := r.Spec.ValidateUpdate(kit); err != nil {
		return err
	}

	if r.Revision == nil {
		return errors.New("revision not set")
	}

	if err := r.Revision.ValidateUpdate(); err != nil {
		return err
	}

	return nil
}

const (
	// HookRevisionStatusNotDeployed ....
	HookRevisionStatusNotDeployed HookRevisionStatus = "not_deployed"

	// HookRevisionStatusDeployed ...
	HookRevisionStatusDeployed HookRevisionStatus = "deployed"

	// HookRevisionStatusShutdown ...
	HookRevisionStatusShutdown HookRevisionStatus = "shutdown"
)

// HookRevisionStatus defines hook revision status.
type HookRevisionStatus string

// String returns hook revision status string.
func (s HookRevisionStatus) String() string {
	return string(s)
}

// Validate strategy set type.
func (s HookRevisionStatus) Validate() error {
	switch s {
	case HookRevisionStatusShutdown:
	case HookRevisionStatusNotDeployed:
	case HookRevisionStatusDeployed:
	default:
		return fmt.Errorf("unsupported hook revision status: %s", s)
	}

	return nil
}
