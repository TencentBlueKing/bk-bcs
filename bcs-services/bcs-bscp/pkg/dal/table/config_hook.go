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
)

// ConfigHook 配置脚本
type ConfigHook struct {
	// ID is an auto-increased value, which is a unique identity of a hook.
	ID         uint32                `json:"id" gorm:"primaryKey"`
	Spec       *ConfigHookSpec       `json:"spec" gorm:"embedded"`
	Attachment *ConfigHookAttachment `json:"attachment" gorm:"embedded"`
	Revision   *Revision             `json:"revision" gorm:"embedded"`
}

// TableName is the ConfigHook's database table name.
func (s *ConfigHook) TableName() string {
	return "config_hooks"
}

// AppID AuditRes interface
func (s *ConfigHook) AppID() uint32 {
	return s.Attachment.AppID
}

// ResID AuditRes interface
func (s *ConfigHook) ResID() uint32 {
	return s.ID
}

// ResType AuditRes interface
func (s *ConfigHook) ResType() string {
	return "config_hook"
}

// ConfigHookSpec defines all the specifics for ConfigHook set by user.
type ConfigHookSpec struct {
	PreHookID         uint32 `json:"pre_hook_id" gorm:"pre_hook_id"`
	PreHookReleaseID  uint32 `json:"pre_hook_release_id" gorm:"pre_hook_release_id"`
	PostHookID        uint32 `json:"post_hook_id" gorm:"post_hook_id"`
	PostHookReleaseID uint32 `json:"post_hook_release_id" gorm:"post_hook_release_id"`
}

// ConfigHookAttachment defines the ConfigHook attachments.
type ConfigHookAttachment struct {
	BizID uint32 `json:"biz_id" gorm:"column:biz_id"`
	AppID uint32 `json:"app_id" gorm:"column:app_id"`
}

// ValidateCreate validate ConfigHook is valid or not when create it.
func (s ConfigHook) ValidateCreate() error {
	if s.ID > 0 {
		return errors.New("id should not be set")
	}

	if s.Spec == nil {
		return errors.New("spec not set")
	}

	if err := s.Spec.ValidateCreate(); err != nil {
		return err
	}

	if s.Attachment == nil {
		return errors.New("attachment not set")
	}

	if err := s.Attachment.Validate(); err != nil {
		return err
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if err := s.Revision.ValidateCreate(); err != nil {
		return err
	}

	return nil
}

// ValidateCreate validate ConfigHook spec when it is created.
func (s ConfigHookSpec) ValidateCreate() error {

	if (s.PostHookID <= 0) && (s.PreHookID <= 0) {
		return errors.New("preHookID and postHookID should not be set")

	}

	return nil
}

// Validate whether ConfigHook attachment is valid or not.
func (s ConfigHookAttachment) Validate() error {
	if s.BizID <= 0 {
		return errors.New("invalid attachment biz id")
	}

	if s.AppID <= 0 {
		return errors.New("invalid attachment app id")
	}

	return nil
}

// ValidateUpdate validate ConfigHook is valid or not when update it.
func (s ConfigHook) ValidateUpdate() error {

	if s.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if s.Attachment.AppID <= 0 {
		return errors.New("app id should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if len(s.Revision.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	if len(s.Revision.Creator) != 0 {
		return errors.New("creator can not be updated")
	}

	return nil
}

// ValidateEnable validate ConfigHook is valid or not when update it.
func (s ConfigHook) ValidateEnable() error {

	if s.Attachment == nil {
		return errors.New("attachment should be set")
	}

	if s.Attachment.AppID <= 0 {
		return errors.New("app id should be set")
	}

	if s.Attachment.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	if s.Revision == nil {
		return errors.New("revision not set")
	}

	if len(s.Revision.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	if len(s.Revision.Creator) != 0 {
		return errors.New("creator can not be updated")
	}

	return nil
}
