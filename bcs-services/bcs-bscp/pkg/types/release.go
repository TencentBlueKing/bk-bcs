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
)

// ListReleasesOption defines options to list release.
type ListReleasesOption struct {
	BizID      uint32    `json:"biz_id"`
	AppID      uint32    `json:"app_id"`
	Deprecated bool      `json:"deprecated"`
	SearchKey  string    `json:"search_key"`
	Page       *BasePage `json:"page"`
}

// Validate the list release options
func (opt *ListReleasesOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.AppID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid app id, should >= 1")
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// ListReleaseDetails defines the response details of requested ListReleasesOption.
type ListReleaseDetails struct {
	Count   uint32           `json:"count"`
	Details []*table.Release `json:"details"`
}

// ListReleasesStrategies defines model to list release strategie.
type ListReleasesStrategies struct {
	PublishTime   string      `gorm:"column:publish_time" json:"publish_time"`
	Name          string      `gorm:"column:name" json:"name"`
	Scope         table.Scope `gorm:"column:scope;type:json" json:"scope"`
	Creator       string      `gorm:"column:creator" json:"creator"`
	FullyReleased bool        `gorm:"column:fully_released" json:"fully_released"`
}
