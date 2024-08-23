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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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

type commonResponse struct {
	Code      int64  `json:"code"`
	Result    bool   `json:"result"`
	RequestId string `json:"request_id"`
	Message   string `json:"message"`
}

type queryRule struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type searchBusinessRequest struct {
	Fields            []string           `json:"fields"`
	BizPropertyFilter *bizPropertyFilter `json:"biz_property_filter,omitempty"`
}

type bizPropertyFilter struct {
	Condition string       `json:"condition"`
	Rules     []*queryRule `json:"rules"`
}

type searchBusinessResponse struct {
	commonResponse `json:",inline"`
	Data           *searchBusinessResponseData `json:"data"`
}

type searchBusinessResponseData struct {
	Count int64        `json:"count"`
	Info  []CCBusiness `json:"info"`
}

// SearchBusiness search businesses with bk_biz_ids
func (h *handler) SearchBusiness(bkBizIds []int64) ([]CCBusiness, error) {
	req := &searchBusinessRequest{
		Fields: []string{"bk_biz_id", "bk_biz_name", "bk_biz_maintainer"},
	}
	if len(bkBizIds) != 0 {
		req.BizPropertyFilter = &bizPropertyFilter{
			Condition: "AND",
			Rules: []*queryRule{
				{
					Field:    "bk_biz_id",
					Operator: "in",
					Value:    bkBizIds,
				},
			},
		}
	}
	respBytes, err := h.query(req, searchBusinessApi)
	if err != nil {
		return nil, errors.Wrapf(err, "CC search business query failed")
	}
	resp := new(searchBusinessResponse)
	if err := json.Unmarshal(respBytes, resp); err != nil {
		return nil, errors.Wrapf(err, "CC search business unmarshal failed")
	}
	if resp.Code != 0 {
		return nil, errors.Errorf("response code not 0, errMsg: %s", resp.Message)
	}
	if resp.Data != nil {
		return resp.Data.Info, nil
	}
	return nil, nil
}

var (
	bkAuthFormat = `{"bk_app_code": "%s", "bk_app_secret": "%s", "bk_token": "%s"}`
)

func (h *handler) query(request interface{}, uri string) (resp []byte, err error) {
	data, err := json.Marshal(request)
	if err != nil {
		return nil, errors.Wrapf(err, "marshal failed")
	}
	httpRequest, err := http.NewRequest("POST", h.op.BKCCUrl+uri,
		bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrapf(err, "create request failed")
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	httpRequest.Header.Set("X-Bkapi-Authorization", fmt.Sprintf(bkAuthFormat, h.op.Auth.AppCode,
		h.op.Auth.AppSecret, "admin"))
	c := &http.Client{
		Timeout: time.Second * 20,
	}
	httpResponse, err := c.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()
	respBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read http response failed")
	}
	return respBytes, nil
}
