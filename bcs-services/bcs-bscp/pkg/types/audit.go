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

// Package types NOTES
package types

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/filter"
)

// ListAuditsOption defines options to list audit.
type ListAuditsOption struct {
	BizID  uint32             `json:"biz_id"`
	Filter *filter.Expression `json:"filter"`
	Page   *BasePage          `json:"page"`
}

// Validate the list audit options
func (lao *ListAuditsOption) Validate(po *PageOption) error {
	if lao.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if lao.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is nil")
	}

	exprOpt := &filter.ExprOption{
		// remove biz_id because it's a required field in the option.
		RuleFields: table.AuditColumns.WithoutColumn("biz_id"),
	}
	if err := lao.Filter.Validate(exprOpt); err != nil {
		return err
	}

	if lao.Page == nil {
		return errf.New(errf.InvalidParameter, "page is null")
	}

	if err := lao.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// ListAuditDetails defines the response details of requested ListAuditsOption
type ListAuditDetails struct {
	Count   uint32         `json:"count"`
	Details []*table.Audit `json:"details"`
}
