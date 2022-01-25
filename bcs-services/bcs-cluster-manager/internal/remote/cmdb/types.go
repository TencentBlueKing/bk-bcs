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
 *
 */

package cmdb

// field result
const (
	fieldBS2NameID = "bs2_name_id"
)

const (
	fieldHostIP      = "bk_host_innerip"
	fieldHostID      = "bk_host_id"
	fieldOperator    = "operator"
	fieldBakOperator = "bk_bak_operator"

	// StartAt offset
	StartAt = 0
	// MaxLimits limit
	MaxLimits = 500
)

// condition result
const (
	conditionBkBizID = "bk_biz_id"
)

// Page page
type Page struct {
	Start int    `json:"start"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort"`
}

// SearchBusinessRequest search business request
type SearchBusinessRequest struct {
	Fields    []string               `json:"fields"`
	Condition map[string]interface{} `json:"condition"`
	Page      Page                   `json:"page"`
	UserName  string                 `json:"bk_username"`
	Operator  string                 `json:"operator"`
}

// SearchBusinessResponse search business resp
type SearchBusinessResponse struct {
	Code      int          `json:"code"`
	Result    bool         `json:"result"`
	Message   string       `json:"message"`
	RequestID string       `json:"request_id"`
	Data      BusinessResp `json:"data"`
}

// BusinessResp resp
type BusinessResp struct {
	Count int            `json:"count"`
	Info  []BusinessData `json:"info"`
}

// BusinessData data
type BusinessData struct {
	BS2NameID       int    `json:"bs2_name_id"`
	Default         int    `json:"default"`
	BKBizID         int64  `json:"bk_biz_id"`
	BKBizName       string `json:"bk_biz_name"`
	BKBizMaintainer string `json:"bk_biz_maintainer"`
}

// ListBizHostRequest xxx
type ListBizHostRequest struct {
	Page        Page     `json:"page"`
	BKBizID     int      `json:"bk_biz_id"`
	BKSetIDs    []int    `json:"bk_set_ids"`
	BKModuleIDs []int    `json:"bk_module_ids"`
	Fields      []string `json:"fields"`
}

// ListBizHostsResponse xxx
type ListBizHostsResponse struct {
	Code      int      `json:"code"`
	Result    bool     `json:"result"`
	Message   string   `json:"message"`
	RequestID string   `json:"request_id"`
	Data      HostResp `json:"data"`
}

// HostResp host resp
type HostResp struct {
	Count int        `json:"count"`
	Info  []HostData `json:"info"`
}

// HostData info
type HostData struct {
	BKHostID      int64  `json:"bk_host_id"`
	BKHostInnerIP string `json:"bk_host_innerip"`
	Operator      string `json:"operator"`
	BKBakOperator string `json:"bk_bak_operator"`
}

const (
	keyBizID          = "BsiId"
	methodBusiness    = "Business"
	methodServer      = "Server"
	methodBusinessRaw = "BusinessRaw"
)

var (
	reqColumns = []string{"BsiId", "BsipId", "BsiProductName", "BsiProductId", "BsiName"}
)

// QueryBusinessInfoReq query business request
type QueryBusinessInfoReq struct {
	Method    string                 `json:"method"`
	ReqColumn []string               `json:"req_column"`
	KeyValues map[string]interface{} `json:"key_values"`
}

// QueryBusinessInfoResp query business resp
type QueryBusinessInfoResp struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Result    bool         `json:"result"`
	RequestID string       `json:"request_id"`
	Data      BusinessInfo `json:"data"`
}

// BusinessInfo business resp
type BusinessInfo struct {
	Data []Business `json:"data"`
}

// Business business info
type Business struct {
	BsiID          int    `json:"BsiId"`
	BsiProductName string `json:"BsiProductName"`
	BsipID         int    `json:"BsipId"`
	BsiName        string `json:"BsiName"`
	BsiProductId   int    `json:"BsiProductId"`
}

// BizInfo business id info
type BizInfo struct {
	BizID int64 `json:"bizID"`
}
