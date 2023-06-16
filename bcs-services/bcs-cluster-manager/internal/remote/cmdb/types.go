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
	// host field info
	fieldCloudID     = "bk_cloud_id"
	fieldHostIP      = "bk_host_innerip"
	fieldHostIPv6    = "bk_host_innerip_v6"
	fieldHostOutIP   = "bk_host_outerip"
	fieldHostOutIPV6 = "bk_host_outerip_v6"
	fieldHostID      = "bk_host_id"
	fieldDeviceType  = "bk_svr_device_cls_name"
	fieldIDCCityName = "idc_city_name"
	fieldIDCCityID   = "idc_city_id"
	fieldDeviceClass = "svr_device_class"
	fieldRack        = "rack"
	fieldIDCName     = "idc_name"

	fieldHostCPU  = "bk_cpu"
	fieldHostMem  = "bk_mem"
	fieldHostDisk = "bk_disk"

	fieldOperator    = "operator"
	fieldBakOperator = "bk_bak_operator"

	// StartAt offset
	StartAt = 0
	// MaxLimits limit
	MaxLimits = 500
)

var (
	fieldHostDetailInfo = []string{fieldHostIP, fieldHostIPv6, fieldHostOutIP, fieldHostOutIPV6, fieldHostID,
		fieldDeviceType, fieldIDCCityName, fieldIDCCityID, fieldDeviceClass, fieldHostCPU,
		fieldHostMem, fieldHostDisk, fieldOperator, fieldBakOperator, fieldRack, fieldIDCName}
)

// condition result
const (
	conditionBkBizID = "bk_biz_id"
)

// Condition xxx
type Condition string

// String to string
func (c Condition) String() string {
	return string(c)
}

var (
	and Condition = "AND"
	or  Condition = "OR"
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

// ListHostsWithoutBizRequest list hosts request
type ListHostsWithoutBizRequest struct {
	Page               Page                `json:"page"`
	BKBizID            int                 `json:"bk_biz_id,omitempty"`
	HostPropertyFilter *HostPropertyFilter `json:"host_property_filter"`
	Fields             []string            `json:"fields"`
}

func buildFilterConditionByInnerIP(ips []string) *HostPropertyFilter {
	return &HostPropertyFilter{
		Condition: and.String(),
		Rules: []Rule{
			{
				Field:    fieldHostIP,
				Operator: "in",
				Value:    ips,
			},
		},
	}
}

// HostPropertyFilter filter confition
type HostPropertyFilter struct {
	// Condition AND OR
	Condition string `json:"condition"`
	// Rules
	Rules []Rule `json:"rules"`
}

// Rule filter rule
type Rule struct {
	Field string `json:"field"`
	// Operator equal,not_equal,in,not_in,less,less_or_equal,greater,greater_or_equal,between,not_between
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// ListHostsWithoutBizResponse resp
type ListHostsWithoutBizResponse struct {
	Code      int    `json:"code"`
	Result    bool   `json:"result"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      struct {
		Count int              `json:"count"`
		Info  []HostDetailData `json:"info"`
	} `json:"data"`
}

// HostDetailData host detailed info
type HostDetailData struct {
	HostData
	BkHostInnerIPV6  string `json:"bk_host_innerip_v6"`
	BkHostOutIP      string `json:"bk_host_outerip"`
	BkHostOutIPV6    string `json:"bk_host_outerip_v6"`
	IDCName          string `json:"idc_name"`
	IDCCityName      string `json:"idc_city_name"`
	IDCCityID        string `json:"idc_city_id"`
	NormalDeviceType string `json:"bk_svr_device_cls_name"`
	SCMDeviceType    string `json:"svr_device_class"`
	HostCpu          int64  `json:"bk_cpu"`
	HostMem          int64  `json:"bk_mem"`
	HostDisk         int64  `json:"bk_disk"`
	Rack             string `json:"rack"`
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
	BsiProductID   int    `json:"BsiProductId"`
}

// BizInfo business id info
type BizInfo struct {
	BizID int64 `json:"bizID"`
}

// BaseResponse baseResp
type BaseResponse struct {
	Code      int    `json:"code"`
	Result    bool   `json:"result"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// TransferHostToIdleModuleRequest transfer host to idle module request
type TransferHostToIdleModuleRequest struct {
	BkBizID  int   `json:"bk_biz_id"`
	BkHostID []int `json:"bk_host_id"`
}

// TransferHostToIdleModuleResponse transfer host to idle module response
type TransferHostToIdleModuleResponse struct {
	BaseResponse
}

// TransferHostToResourceModuleRequest transfer host to resource module request
type TransferHostToResourceModuleRequest struct {
	BkBizID  int   `json:"bk_biz_id"`
	BkHostID []int `json:"bk_host_id"`
}

// TransferHostToResourceModuleResponse transfer host to resource module response
type TransferHostToResourceModuleResponse struct {
	BaseResponse
}

// DeleteHostRequest delete host request
type DeleteHostRequest struct {
	BkHostID string `json:"bk_host_id"`
}

// DeleteHostResponse delete host response
type DeleteHostResponse struct {
	BaseResponse
}

// GetBizInternalModuleResponse get biz internal module response
type GetBizInternalModuleResponse struct {
	BaseResponse
	Data GetBizInternalModuleData `json:"data"`
}

// GetBizInternalModuleData get biz internal module data
type GetBizInternalModuleData struct {
	BKSetID   int                        `json:"bk_set_id"`
	BKSetName string                     `json:"bk_set_name"`
	Modules   []GetBizInternalModuleInfo `json:"module"`
}

// GetBizInternalModuleInfo get biz internal module info
type GetBizInternalModuleInfo struct {
	BKModuleID       int    `json:"bk_module_id"`
	BKModuleName     string `json:"bk_module_name"`
	Default          int    `json:"default"`
	HostApplyEnabled bool   `json:"host_apply_enabled"`
}

// moduleNameMaps map module name
var moduleNameMaps = map[string]string{
	"idle pool":    "空闲机池",
	"idle host":    "空闲机",
	"fault host":   "故障机",
	"recycle host": "待回收",
}

// ReplaceName replace module name
func (g *GetBizInternalModuleData) ReplaceName() {
	if v, ok := moduleNameMaps[g.BKSetName]; ok {
		g.BKSetName = v
	}
	for i := range g.Modules {
		if v, ok := moduleNameMaps[g.Modules[i].BKModuleName]; ok {
			g.Modules[i].BKModuleName = v
		}
	}
}

// SearchBizInstTopoResponse search biz inst topo response
type SearchBizInstTopoResponse struct {
	BaseResponse
	Data []SearchBizInstTopoData `json:"data"`
}

// SearchBizInstTopoData search biz inst topo data
type SearchBizInstTopoData struct {
	BKInstID   int                     `json:"bk_inst_id"`
	BKInstName string                  `json:"bk_inst_name"`
	BKObjID    string                  `json:"bk_obj_id"`
	BKObjName  string                  `json:"bk_obj_name"`
	Default    int                     `json:"default"`
	Child      []SearchBizInstTopoData `json:"child"`
}

// TransferHostModuleRequest transfer host module
type TransferHostModuleRequest struct {
	BKBizID     int   `json:"bk_biz_id"`
	BKHostID    []int `json:"bk_host_id"`
	BKModuleID  []int `json:"bk_module_id"`
	IsIncrement bool  `json:"is_increment"`
}
