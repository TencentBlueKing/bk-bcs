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

package cmdb

import (
	"encoding/json"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/types"
)

// SearchBizParams is esb search cmdb business parameter.
type esbSearchBizParams struct {
	*SearchBizParams
}

// SearchBizParams is cmdb search business parameter.
type SearchBizParams struct {
	Condition         map[string]string `json:"condition,omitempty"`
	Fields            []string          `json:"fields"`
	Page              BasePage          `json:"page"`
	BizPropertyFilter *QueryFilter      `json:"biz_property_filter,omitempty"`
}

// BizIDField NOTES
const BizIDField = "bk_biz_id"

// QueryFilter is cmdb common query filter.
type QueryFilter struct {
	Rule `json:",inline"`
}

// Rule is cmdb common query rule type.
type Rule interface {
	GetDeep() int
}

// CombinedRule is cmdb query rule that is combined by multiple AtomRule.
type CombinedRule struct {
	Condition Condition `json:"condition"`
	Rules     []Rule    `json:"rules"`
}

// Condition NOTES
type Condition string

const (
	// ConditionAnd NOTES
	ConditionAnd = Condition("AND")
)

// GetDeep get query rule depth.
func (r CombinedRule) GetDeep() int {
	maxChildDeep := 1
	for _, child := range r.Rules {
		childDeep := child.GetDeep()
		if childDeep > maxChildDeep {
			maxChildDeep = childDeep
		}
	}
	return maxChildDeep + 1
}

// AtomRule is cmdb atomic query rule.
type AtomRule struct {
	Field    string      `json:"field"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
}

// Operator https://github.com/TencentBlueKing/bk-cmdb/tree/master/src/common/querybuilder 规范文档
type Operator string

var (
	// OperatorEqual 等于
	OperatorEqual = Operator("equal")
	// OperatorIn 包含
	OperatorIn = Operator("in")
)

// GetDeep get query rule depth.
func (r AtomRule) GetDeep() int {
	return 1
}

// MarshalJSON marshal QueryFilter to json.
func (qf *QueryFilter) MarshalJSON() ([]byte, error) {
	if qf.Rule != nil {
		return json.Marshal(qf.Rule)
	}
	return make([]byte, 0), nil
}

// BasePage is cmdb paging parameter.
type BasePage struct {
	Sort        string `json:"sort,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Start       int    `json:"start"`
	EnableCount bool   `json:"enable_count,omitempty"`
}

// SearchBizResp is cmdb search business response.
type SearchBizResp struct {
	types.BaseResponse
	SearchBizResult `json:"data"`
}

// SearchBizResult is cmdb search business response.
type SearchBizResult struct {
	Count int64 `json:"count"`
	Info  []Biz `json:"info"`
}

// Biz is cmdb biz info.
type Biz struct {
	BizID         int64  `json:"bk_biz_id"`
	BizName       string `json:"bk_biz_name"`
	BizMaintainer string `json:"bk_biz_maintainer"`
}
