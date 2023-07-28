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
	"time"
)

// ReleasedHook defines a released hook with release info and hook revision info
type ReleasedHook struct {
	// ID is an auto-increased value, which is a group app's
	// unique identity.
	ID               uint32     `db:"id" json:"id" gorm:"primaryKey"`
	AppID            uint32     `db:"app_id" json:"app_id" gorm:"column:app_id"`
	ReleaseID        uint32     `db:"release_id" json:"release_id" gorm:"column:release_id"`
	HookID           uint32     `db:"hook_id" json:"hook_id" gorm:"column:hook_id"`
	HookRevisionID   uint32     `db:"hook_revision_id" json:"hook_revision_id" gorm:"column:hook_revision_id"`
	HookName         string     `db:"hook_name" json:"hook_name" gorm:"column:hook_name"`
	HookRevisionName string     `db:"hook_revision_name" json:"hook_revision_name" gorm:"column:hook_revision_name"`
	Content          string     `db:"content" json:"content" gorm:"column:content"`
	ScriptType       ScriptType `db:"script_type" json:"script_type" gorm:"column:script_type"`
	HookType         HookType   `db:"hook_type" json:"hook_type" gorm:"column:hook_type"`
	BizID            uint32     `db:"biz_id" json:"biz_id" gorm:"column:biz_id"`
	Reviser          string     `db:"reviser" json:"reviser" gorm:"column:reviser"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at" gorm:"column:updated_at"`
}

// TableName is the released hook's database table name.
func (c ReleasedHook) TableName() string {
	return "released_hooks"
}

const (
	// PreHook is the type for pre hook
	PreHook HookType = "pre_hook"

	// PostHook is the type for post hook
	PostHook HookType = "post_hook"
)

// HookType is the type of hook
type HookType string

// String returns the string value of hook type
func (s HookType) String() string {
	return string(s)
}

// Validate validate the hook type
func (s HookType) Validate() error {
	if s == "" {
		return nil
	}
	switch s {
	case PreHook:
	case PostHook:
	default:
		return fmt.Errorf("unsupported hook type: %s", s)
	}

	return nil
}

// ValidateCreate validate the group app's specific when create it.
func (c ReleasedHook) ValidateCreate() error {
	if c.ID != 0 {
		return errors.New("group app id can not be set")
	}
	if c.AppID <= 0 {
		return errors.New("app id should be set")
	}
	if c.HookID <= 0 {
		return errors.New("hook id should be set")
	}
	if c.HookName == "" {
		return errors.New("hook name should be set")
	}
	if c.HookRevisionID <= 0 {
		return errors.New("hook revision id should be set")
	}
	if c.HookRevisionName == "" {
		return errors.New("hook revision name should be set")
	}
	if c.Content == "" {
		return errors.New("content should be set")
	}
	if err := c.ScriptType.Validate(); err != nil {
		return err
	}
	if err := c.HookType.Validate(); err != nil {
		return err
	}
	if c.BizID <= 0 {
		return errors.New("biz id should be set")
	}
	if c.Reviser == "" {
		return errors.New("reviser should be set")
	}

	return nil
}

// ValidateUpdate validate the group app's specific when update it.
func (c ReleasedHook) ValidateUpdate() error {
	if c.ID <= 0 {
		return errors.New("group app id should be set")
	}
	if c.BizID <= 0 {
		return errors.New("biz id should be set")
	}
	if c.Reviser == "" {
		return errors.New("reviser should be set")
	}

	return nil
}

// ValidateDelete validate the group app's info when delete it.
func (c ReleasedHook) ValidateDelete() error {
	if c.ID <= 0 {
		return errors.New("group app id should be set")
	}
	if c.BizID <= 0 {
		return errors.New("biz id should be set")
	}

	return nil
}
