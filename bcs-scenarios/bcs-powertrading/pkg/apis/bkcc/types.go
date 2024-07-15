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

package bkcc

// CCHostInfo defines the host info of node in bkcc
type CCHostInfo struct {
	BKHostID      int64  `json:"bk_host_id"`
	BKCloudID     int32  `json:"bk_cloud_id"`
	BKHostInnerIP string `json:"bk_host_innerip"`
	BKAssetID     string `json:"bk_asset_id"`
	IDCName       string `json:"idc_name"`
	IDCCityName   string `json:"idc_city_name"`
	IDCUnitName   string `json:"idc_unit_name"`

	// Region is calculated by idc info
	Region string `json:"-"`
}

type appInfo struct {
	AppCode   string `json:"bk_app_code"`   // app code for api
	AppSecret string `json:"bk_app_secret"` // app secret for api
	Operator  string `json:"bk_username"`   // rtx name
}

type listBizHostsRequest struct {
	appInfo            `json:",inline"`
	Fields             []string            `json:"fields"`
	Page               *page               `json:"page"`
	BkBizID            int64               `json:"bk_biz_id"`
	HostPropertyFilter *hostPropertyFilter `json:"host_property_filter,omitempty"`
}

type hostPropertyFilter struct {
	Condition string       `json:"condition,omitempty"`
	Rules     []*queryRule `json:"rules"`
}

type queryRule struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type page struct {
	Start int    `json:"start"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort"`
}

type listHostsWithoutBizResponse struct {
	commonResponse `json:",inline"`
	Data           *listHostsWithoutBizResponseData `json:"data"`
}

type listHostsWithoutBizResponseData struct {
	Count int64        `json:"count"`
	Info  []CCHostInfo `json:"info"`
}

type commonResponse struct {
	Code      int64  `json:"code"`
	Result    bool   `json:"result"`
	RequestId string `json:"request_id"`
	Message   string `json:"message"`
}
