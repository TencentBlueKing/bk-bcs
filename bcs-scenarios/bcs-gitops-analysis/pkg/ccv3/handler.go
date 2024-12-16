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

package ccv3

import (
	"net/http"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/options"
)

// Handler the struct that used to query cc
type handler struct {
	op *options.AnalysisOptions
}

// NewHandler create cc query handler
func NewHandler() Interface {
	return &handler{
		op: options.GlobalOptions(),
	}
}

type queryRule struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type searchBusinessRequest struct {
	Fields            []string           `json:"fields"`
	BizPropertyFilter *bizPropertyFilter `json:"biz_property_filter,omitempty"`
	Page              *page              `json:"page"`
}

type bizPropertyFilter struct {
	Condition string       `json:"condition"`
	Rules     []*queryRule `json:"rules"`
}

type page struct {
	Start int `json:"start"`
	Limit int `json:"limit"`
}

type searchBusinessResponse struct {
	Code      int64                       `json:"code"`
	Result    bool                        `json:"result"`
	RequestId string                      `json:"request_id"`
	Message   string                      `json:"message"`
	Data      *searchBusinessResponseData `json:"data"`
}

type searchBusinessResponseData struct {
	Count int64        `json:"count"`
	Info  []CCBusiness `json:"info"`
}

// CCBusiness business for cc
type CCBusiness struct {
	BkBizId      int64  `json:"bk_biz_id"`
	BkBizName    string `json:"bk_biz_name"`
	BkMaintainer string `json:"bk_biz_maintainer"`
	GroupName    string `json:"bk_oper_grp_name"`
	GroupID      int64  `json:"bk_oper_grp_name_id"`
	LocalBizID   string `json:"local_biz_id,omitempty"`
}

// SearchBusiness search businesses with bk_biz_ids
func (h *handler) SearchBusiness(bkBizIds []int64) ([]CCBusiness, error) {
	req := &searchBusinessRequest{
		Fields: []string{"bk_biz_id", "bk_biz_name", "bk_biz_maintainer", "bk_oper_grp_name", "bk_oper_grp_name_id",
			"local_biz_id"},
		BizPropertyFilter: &bizPropertyFilter{
			Condition: "AND",
			Rules: []*queryRule{
				{
					Field:    "bk_biz_id",
					Operator: "in",
					Value:    bkBizIds,
				},
			},
		},
		Page: &page{
			Start: 0,
			Limit: 1000,
		},
	}
	resp := new(searchBusinessResponse)
	if err := h.query(h.op.BKCCUrl+searchBusinessApi, http.MethodPost, req, resp); err != nil {
		return nil, errors.Wrapf(err, "CC search business query failed")
	}
	if resp.Code != 0 {
		return nil, errors.Errorf("response code not 0, errMsg: %s", resp.Message)
	}
	if resp.Data == nil {
		return nil, nil
	}
	for i := range resp.Data.Info {
		bizInfo := resp.Data.Info[i]
		if bizInfo.LocalBizID != "" {
			v, err := strconv.ParseInt(bizInfo.LocalBizID, 0, 64)
			if err != nil {
				blog.Errorf("parse local_biz_id '%s' failed", bizInfo.LocalBizID)
				continue
			}
			resp.Data.Info[i].BkBizId = int64(v)
		}
	}
	return resp.Data.Info, nil
}
