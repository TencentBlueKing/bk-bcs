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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

// ListHooksWithReferOption defines options to list group.
type ListHooksWithReferOption struct {
	BizID     uint32    `json:"biz_id"`
	Name      string    `json:"name"`
	Tag       string    `json:"tag"`
	All       bool      `json:"all"`
	NotTag    bool      `json:"not_tag"`
	Page      *BasePage `json:"page"`
	SearchKey string    `json:"search_key"`
}

// Validate the list group options
func (opt *ListHooksWithReferOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if opt.Page == nil {
		return errors.New("page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// ListHooksWithReferDetail defines the response details.
type ListHooksWithReferDetail struct {
	Hook                *table.Hook `json:"hook" gorm:"embedded"`
	ReferCount          int64       `json:"refer_count" gorm:"column:refer_count"`
	BoundEditingRelease bool        `json:"refer_editing_release" gorm:"column:refer_editing_release"`
	PublishedRevisionID uint32      `json:"published_revision_id" gorm:"column:published_revision_id"`
}

// ListHookReferencesOption defines options to list hook references.
type ListHookReferencesOption struct {
	BizID     uint32 `json:"biz_id"`
	HookID    uint32 `json:"hook_id"`
	SearchKey string `json:"search_key"`
	Page      *BasePage
}

// Validate the list hook references options
func (opt *ListHookReferencesOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errors.New("invalid biz id, should >= 1")
	}

	if opt.Page == nil {
		return errors.New("page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// ListHookReferencesDetail defines the response details.
type ListHookReferencesDetail struct {
	HookRevisionID   uint32 `gorm:"column:hook_revision_id" json:"hook_revision_id"`
	HookRevisionName string `gorm:"column:hook_revision_name" json:"hook_revision_name"`
	AppID            uint32 `gorm:"column:app_id" json:"app_id"`
	AppName          string `gorm:"column:app_name" json:"app_name"`
	ReleaseID        uint32 `gorm:"column:release_id" json:"release_id"`
	ReleaseName      string `gorm:"column:release_name" json:"release_name"`
	HookType         string `gorm:"column:hook_type" json:"hook_type"`
	Deprecated       bool   `gorm:"column:deprecated" json:"deprecated"`
}

// HookTagCount defines the response details of requested CountHookTag.
type HookTagCount struct {
	Tag    string `gorm:"column:tag" json:"tag"`
	Counts uint32 `gorm:"column:counts" json:"counts"`
}
