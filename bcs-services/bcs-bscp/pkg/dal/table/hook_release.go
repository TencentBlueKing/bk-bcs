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
	"fmt"

	"bscp.io/pkg/criteria/validator"
)

// HookRelease 脚本版本
type HookRelease struct {
	// ID is an auto-increased value, which is a unique identity of a hook.
	ID uint32 `db:"id" json:"id"`

	Spec       *HookReleaseSpec       `json:"spec" gorm:"embedded"`
	Attachment *HookReleaseAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision              `json:"revision" gorm:"embedded"`
}

// HookReleaseSpec defines all the specifics for hook set by user.
type HookReleaseSpec struct {
	Name       string            `json:"name" gorm:"column:name"`
	PublishNum uint32            `json:"publish_num" gorm:"column:publish_num"`
	State      HookReleaseStatus `json:"state" gorm:"column:state"`
	Content    string            `json:"content" gorm:"column:content"`
	Memo       string            `json:"memo" gorm:"column:memo"`
}

// HookReleaseAttachment defines the hook attachments.
type HookReleaseAttachment struct {
	BizID  uint32 `json:"biz_id" gorm:"column:biz_id"`
	HookID uint32 `json:"hook_id" gorm:"column:hook_id"`
}

// TableName is the hook's database table name.
func (r *HookRelease) TableName() Name {
	return "hook_releases"
}

// AppID AuditRes interface
func (r *HookRelease) AppID() uint32 {
	return 0
}

// ResID AuditRes interface
func (r *HookRelease) ResID() uint32 {
	return r.ID
}

// ResType AuditRes interface
func (r *HookRelease) ResType() string {
	return "hook_releases"
}

// ValidateCreate validate hook is valid or not when create it.
func (r *HookRelease) ValidateCreate() error {

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

	if err := r.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateCreate validate spec when created.
func (s *HookReleaseSpec) ValidateCreate() error {

	if err := validator.ValidateName(s.Name); err != nil {
		return err
	}

	if err := validator.ValidateMemo(s.Memo, false); err != nil {
		return err
	}

	return nil
}

// Validate validate Attachment.
func (a HookReleaseAttachment) Validate() error {

	if a.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if a.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	return nil
}

// ValidateDelete validate the hook release info when delete it.
func (r HookRelease) ValidateDelete() error {
	if r.ID <= 0 {
		return errors.New("hook release id should be set")
	}

	if r.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if r.Attachment.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	return nil
}

// ValidateDeleteByHookID validate the hook release info when delete it.
func (r HookRelease) ValidateDeleteByHookID() error {

	if r.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if r.Attachment.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	return nil
}

// ValidatePublish validate the Publish
func (r HookRelease) ValidatePublish() error {

	if r.ID <= 0 {
		return errors.New("hook release id should be set")
	}

	if r.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if r.Attachment.HookID <= 0 {
		return errors.New("hook id should be set")
	}

	return nil
}

const (
	// NotDeployedHookReleased ....
	NotDeployedHookReleased HookReleaseStatus = "not_deployed"

	// DeployedHookReleased ...
	DeployedHookReleased HookReleaseStatus = "deployed"

	// ShutdownHookReleased ...
	ShutdownHookReleased HookReleaseStatus = "shutdown"
)

// HookReleaseStatus defines hook release status.
type HookReleaseStatus string

// String returns hook release status string.
func (s HookReleaseStatus) String() string {
	return string(s)
}

// Validate strategy set type.
func (s HookReleaseStatus) Validate() error {
	switch s {
	case ShutdownHookReleased:
	case NotDeployedHookReleased:
	case DeployedHookReleased:
	default:
		return fmt.Errorf("unsupported hook release released status: %s", s)
	}

	return nil
}
