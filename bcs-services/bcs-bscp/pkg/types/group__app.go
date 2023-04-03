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

// ListGroupAppsOption defines the option for list group apps.
type ListGroupAppsOption struct {
	BizID  uint32             `json:"biz_id"`
	Filter *filter.Expression `json:"filter"`
}

// Validate the list group apps options
func (opt *ListGroupAppsOption) Validate() error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is nil")
	}

	exprOpt := &filter.ExprOption{
		// remove biz_id because it's a required field in the option.
		RuleFields: table.GroupAppColumns.WithoutColumn("biz_id"),
	}
	if err := opt.Filter.Validate(exprOpt); err != nil {
		return err
	}

	return nil
}
