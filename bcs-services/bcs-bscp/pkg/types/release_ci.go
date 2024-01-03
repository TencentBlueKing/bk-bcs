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

// ListReleasedCIsOption defines options to list released config item.
type ListReleasedCIsOption struct {
	BizID     uint32    `json:"biz_id"`
	ReleaseID uint32    `json:"release_id"`
	SearchKey string    `json:"search_key"`
	Page      *BasePage `json:"page"`
}

// Validate the list released config item options
func (opt *ListReleasedCIsOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page is null")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// ListReleasedCIsDetails defines the response details of requested ListReleasedCIsOption.
type ListReleasedCIsDetails struct {
	Count   uint32                      `json:"count"`
	Details []*table.ReleasedConfigItem `json:"details"`
}
