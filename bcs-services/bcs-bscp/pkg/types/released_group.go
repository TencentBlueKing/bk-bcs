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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/filter"
)

// CountGroupsReleasedAppsOption defines options to count each group's published apps.
type CountGroupsReleasedAppsOption struct {
	BizID  uint32   `json:"biz_id"`
	Groups []uint32 `json:"groups"`
}

// Validate the count group's published apps options
func (opt *CountGroupsReleasedAppsOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}
	return nil
}

// GroupPublishedAppsCount defines the response details of requested CountGroupsReleasedAppsOption.
type GroupPublishedAppsCount struct {
	GroupID uint32 `gorm:"column:group_id" json:"group_id"`
	Counts  uint32 `gorm:"column:counts" json:"counts"`
	Edited  bool   `gorm:"column:edited" json:"edited"`
}

// ListReleasedGroupsOption defines options to list group current releases.
type ListReleasedGroupsOption struct {
	BizID  uint32             `json:"biz_id"`
	Filter *filter.Expression `json:"filter"`
}

// Validate the list group current release options
func (opt *ListReleasedGroupsOption) Validate() error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is nil")
	}

	exprOpt := &filter.ExprOption{
		// remove biz_id because it's a required field in the option.
		RuleFields: table.ReleasedGroupColumns.WithoutColumn("biz_id"),
	}
	if err := opt.Filter.Validate(exprOpt); err != nil {
		return err
	}

	return nil
}
