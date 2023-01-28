/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package types

import (
	"fmt"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/runtime/filter"
)

// GetCPSMaxPageLimit NOTES
const GetCPSMaxPageLimit = 100

// PublishStrategyOption defines options to publish a strategy
type PublishStrategyOption struct {
	BizID      uint32                 `json:"biz_id"`
	AppID      uint32                 `json:"app_id"`
	StrategyID uint32                 `json:"strategy_id"`
	Revision   *table.CreatedRevision `json:"revision"`
}

// Validate options is valid or not.
func (ps PublishStrategyOption) Validate() error {
	if ps.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "biz_id is invalid")
	}

	if ps.AppID <= 0 {
		return errf.New(errf.InvalidParameter, "app_id is invalid")
	}

	if ps.StrategyID <= 0 {
		return errf.New(errf.InvalidParameter, "strategy_id is invalid")
	}

	if ps.Revision == nil {
		return errf.New(errf.InvalidParameter, "revision is not set")
	}

	if err := ps.Revision.Validate(); err != nil {
		return errf.New(errf.InvalidParameter, fmt.Sprintf("invalid revision %v", err))
	}

	return nil
}

// ListPSHistoriesOption defines options to list published strategy history.
type ListPSHistoriesOption struct {
	BizID  uint32             `json:"biz_id"`
	AppID  uint32             `json:"app_id"`
	Filter *filter.Expression `json:"filter"`
	Page   *BasePage          `json:"page"`
}

// Validate the list published strategy history options
func (opt *ListPSHistoriesOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.AppID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid app id, should >= 1")
	}

	if opt.Filter == nil {
		return errf.New(errf.InvalidParameter, "invalid filter, is nil")
	}

	exprOpt := &filter.ExprOption{
		// remove biz_id,app_id because it's a required field in the option.
		RuleFields: table.StrategyColumns.WithoutColumn("biz_id", "app_id"),
	}
	if err := opt.Filter.Validate(exprOpt); err != nil {
		return err
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page not set")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// ListPSHistoryDetails defines the response details of requested ListPSHistoriesOption.
type ListPSHistoryDetails struct {
	Count   uint32                            `json:"count"`
	Details []*table.PublishedStrategyHistory `json:"details"`
}

// FinishPublishOption defines options to finish a publish process.
type FinishPublishOption struct {
	BizID      uint32 `json:"biz_id"`
	AppID      uint32 `json:"app_id"`
	StrategyID uint32 `json:"strategy_id"`
}

// Validate the finish publish option is valid or not.
func (fp FinishPublishOption) Validate() error {
	if fp.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if fp.AppID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid app id, should >= 1")
	}

	if fp.StrategyID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid strategy id, should >= 1")
	}

	return nil
}

// ListAppCPStrategiesOptions defines options to list app's current
// published strategies
type ListAppCPStrategiesOptions struct {
	BizID     uint32 `json:"biz_id"`
	AppID     uint32 `json:"app_id"`
	Namespace string `json:"namespace"`
}

// GetAppCPSOption defines options to get app's current published strategies.
type GetAppCPSOption struct {
	BizID uint32    `json:"biz_id"`
	AppID uint32    `json:"app_id"`
	Page  *BasePage `json:"page"`
}

// Validate the get current published strategies options.
func (opt *GetAppCPSOption) Validate(po *PageOption) error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.AppID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid app id, should >= 1")
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}

// GetAppCpsIDOption defines options to get app's current published strategy ids.
type GetAppCpsIDOption struct {
	BizID     uint32 `json:"biz_id"`
	AppID     uint32 `json:"app_id"`
	Namespace string `json:"namespace"`
}

// Validate the get current published strategies options.
func (opt *GetAppCpsIDOption) Validate() error {
	if opt.BizID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid biz id, should >= 1")
	}

	if opt.AppID <= 0 {
		return errf.New(errf.InvalidParameter, "invalid app id, should >= 1")
	}

	return nil
}
