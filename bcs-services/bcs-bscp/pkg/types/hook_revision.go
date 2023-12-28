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

package types

import (
	"errors"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

// ListHookRevisionsOption defines the response details of requested ListHookRevisionsOption.
type ListHookRevisionsOption struct {
	BizID     uint32                   `json:"biz_id"`
	HookID    uint32                   `json:"hook_id"`
	Page      *BasePage                `json:"page"`
	SearchKey string                   `json:"search_key"`
	State     table.HookRevisionStatus `json:"state"`
}

// ListHookRevisionDetails defines the response details of requested ListHookRevisionsReleaseOption.
type ListHookRevisionDetails struct {
	Count   uint32                `json:"count"`
	Details []*table.HookRevision `json:"details"`
}

// Validate the list revisions options
func (opt *ListHookRevisionsOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if opt.HookID <= 0 {
		return errors.New("invalid hook id id, should >= 1")
	}

	if opt.Page == nil {
		return errors.New("page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	if opt.State.String() != "" {
		if err := opt.State.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ListHookRevisionsWithReferDetail defines the response details.
type ListHookRevisionsWithReferDetail struct {
	HookRevision        *table.HookRevision `json:"hook_revision" gorm:"embedded"`
	ReferCount          int64               `json:"refer_count" gorm:"column:refer_count"`
	BoundEditingRelease bool                `json:"refer_editing_release" gorm:"column:refer_editing_release"`
}

// ListHookRevisionReferencesDetail defines the response details.
type ListHookRevisionReferencesDetail struct {
	RevisionID   uint32 `gorm:"column:revision_id" json:"revision_id"`
	RevisionName string `gorm:"column:revision_name" json:"revision_name"`
	AppID        uint32 `gorm:"column:app_id" json:"app_id"`
	AppName      string `gorm:"column:app_name" json:"app_name"`
	ReleaseID    uint32 `gorm:"column:release_id" json:"release_id"`
	ReleaseName  string `gorm:"column:release_name" json:"release_name"`
	HookType     string `gorm:"column:hook_type" json:"hook_type"`
	Deprecated   bool   `gorm:"column:deprecated" json:"deprecated"`
}

// GetByPubStateOption defines options to get hr by State
type GetByPubStateOption struct {
	BizID  uint32
	HookID uint32
	State  table.HookRevisionStatus
}

// Validate the get ByPubState option
func (opt *GetByPubStateOption) Validate() error {
	if opt.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if opt.HookID <= 0 {
		return errors.New("invalid hook id id, should >= 1")
	}

	if err := opt.State.Validate(); err != nil {
		return err
	}

	return nil
}

// ListHookRevisionReferencesOption defines the response details of requested ListHookRevisionReferencesOption.
type ListHookRevisionReferencesOption struct {
	BizID           uint32    `json:"biz_id"`
	HookID          uint32    `json:"hook_id"`
	HookRevisionsID uint32    `json:"hook_revision_id"`
	SearchKey       string    `json:"search_key"`
	Page            *BasePage `json:"page"`
}

// ListHookRevisionReferences defines the response details of requested ListHookRevisionReferencesOption.
type ListHookRevisionReferences struct {
	AppID             uint32 `json:"app_id"`
	ConfigReleaseName string `json:"config_revision_name"`
	ConfigReleaseID   uint32 `json:"config_revision_id"`
	HookRevisionName  string `json:"hook_revision_name"`
	HookRevisionID    uint32 `json:"hook_revision_id"`
	AppName           string `json:"app_name"`
	PubSate           string `json:"pub_sate"`
}

// Validate the list revision options
func (opt *ListHookRevisionReferencesOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if opt.HookID <= 0 {
		return errors.New("invalid hook id id, should >= 1")
	}

	if opt.Page == nil {
		return errors.New("page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}
