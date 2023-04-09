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
	GroupID uint32 `db:"group_id" json:"group_id"`
	Counts  uint32 `db:"counts" json:"counts"`
	Edited  bool   `db:"edited" json:"edited"`
}
