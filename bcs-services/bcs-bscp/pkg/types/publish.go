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
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

// PublishOption defines options to publish a strategy
type PublishOption struct {
	BizID            uint32                 `json:"biz_id"`
	AppID            uint32                 `json:"app_id"`
	ReleaseID        uint32                 `json:"release_id"`
	Memo             string                 `json:"memo"`
	All              bool                   `json:"all"`
	Default          bool                   `json:"default"`
	Groups           []uint32               `json:"groups"`
	Revision         *table.CreatedRevision `json:"revision"`
	PublishType      table.PublishType      `json:"publish_type"`
	PublishTime      string                 `json:"publish_time"`
	PublishStatus    table.PublishStatus    `json:"publish_status"`
	RejectReason     string                 `json:"reject_reason"`
	Approver         string                 `json:"approver"`
	ApproverProgress string                 `json:"approver_progress"`
	PubState         string                 `json:"pub_state"`
}

// Validate options is valid or not.
func (ps *PublishOption) Validate() error {
	if ps.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "biz_id is invalid")
	}

	if ps.AppID <= 0 {
		return errf.New(errf.InvalidParameter, "app_id is invalid")
	}

	if ps.ReleaseID <= 0 {
		return errf.New(errf.InvalidParameter, "release_id is invalid")
	}

	if !ps.All && len(ps.Groups) == 0 {
		return errf.New(errf.InvalidParameter, "groups can not be empty when gray publish")
	}

	if ps.All && len(ps.Groups) > 0 {
		return errf.New(errf.InvalidParameter, "groups can not be set when full publish")
	}

	if ps.Revision == nil {
		return errf.New(errf.InvalidParameter, "revision is not set")
	}

	if err := ps.Revision.Validate(); err != nil {
		return errf.New(errf.InvalidParameter, fmt.Sprintf("invalid revision %v", err))
	}

	return nil
}

// TicketInfo defines fields to ticket info
type TicketInfo struct {
	Status   string
	Operater string
	Message  string
	Over     bool
}
