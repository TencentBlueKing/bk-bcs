/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/runtime/filter"
)

// ListGroupsOption defines options to list group.
type ListGroupsOption struct {
	BizID  uint32             `json:"biz_id"`
	Filter *filter.Expression `json:"filter"`
	Page   *BasePage          `json:"page"`
}

// Validate the list group options
func (opt *ListGroupsOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is nil")
	}

	exprOpt := &filter.ExprOption{
		// remove biz_id,app_id because it's a required field in the option.
		RuleFields: table.GroupColumns.WithoutColumn("biz_id", "app_id"),
	}
	if err := opt.Filter.Validate(exprOpt); err != nil {
		return err
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// ListGroupDetails defines the response details of requested ListGroupsOption.
type ListGroupDetails struct {
	Count   uint32         `json:"count"`
	Details []*table.Group `json:"details"`
}

// ListGroupRleasesdAppsOption defines options to list group's published apps and their release details.
type ListGroupRleasesdAppsOption struct {
	BizID   uint32 `json:"biz_id"`
	GroupID uint32 `json:"group_id"`
	Start   uint32 `json:"start"`
	Limit   uint32 `json:"limit"`
}

// Validate the list group's published apps options
func (opt *ListGroupRleasesdAppsOption) Validate() error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.GroupID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid group id, should >= 1")
	}

	if opt.Start < 0 {
		return errf.New(errf.InvalidParameter, "invalid start, should >= 0")
	}

	if opt.Limit <= 0 {
		return errf.New(errf.InvalidParameter, "invalid limit, should >= 1")
	}
	return nil
}

// ListGroupRleasesdAppsData defines the response detail data of requested ListGroupRleasesdAppsOption.
type ListGroupRleasesdAppsData struct {
	AppID       uint32 `gorm:"column:app_id" json:"app_id"`
	AppName     string `gorm:"column:app_name" json:"app_name"`
	ReleaseID   uint32 `gorm:"column:release_id" json:"release_id"`
	ReleaseName string `gorm:"column:release_name" json:"release_name"`
	Edited      bool   `gorm:"column:edited" json:"edited"`
}

// ListGroupRleasesdAppsDetails defines the response details of requested ListGroupRleasesdAppsOption.
type ListGroupRleasesdAppsDetails struct {
	Count   uint32                       `json:"count"`
	Details []*ListGroupRleasesdAppsData `json:"details"`
}
