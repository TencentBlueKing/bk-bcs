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

// idcZoneMap idc zone
var idcZoneMap = map[string]string{
	"nanjing":   "南京",
	"guangzhou": "广州",
	"shanghai":  "上海",
	"beijing":   "北京",
	"tianjin":   "天津",
	"shenzhen":  "深圳",
}

// GetCityZoneByCityName trans cityName to region
func GetCityZoneByCityName(name string) string {
	for region, city := range idcZoneMap {
		if name == city {
			return region
		}
	}

	return ""
}

// field result
const (
	fieldBS2NameID = "bs2_name_id"
)

const (
	// host field info
	fieldCloudID = "bk_cloud_id"
	// FieldHostIP field host ip
	FieldHostIP      = "bk_host_innerip"
	fieldHostIPv6    = "bk_host_innerip_v6"
	fieldHostOutIP   = "bk_host_outerip"
	fieldHostOutIPV6 = "bk_host_outerip_v6"
	fieldHostID      = "bk_host_id"
	fieldHostName    = "bk_host_name" // 主机名称
	fieldOsType      = "bk_os_type"   // 操作系统类型
	fieldOsName      = "bk_os_name"   // 操作系统名称
	// FieldAssetId 固资号ID
	FieldAssetId = "bk_asset_id"

	fieldDeviceType  = "bk_svr_device_cls_name"
	fieldIDCCityName = "idc_city_name"
	fieldIDCCityID   = "idc_city_id"
	fieldDeviceClass = "svr_device_class"
	fieldRack        = "rack"
	fieldIDCName     = "idc_name"
	fieldAgentId     = "bk_agent_id"

	fieldSubZoneID = "sub_zone_id"    // 子ZoneID
	fieldSubZone   = "sub_zone"       // 子Zone
	fieldIDCAreaID = "bk_idc_area_id" // 区域ID
	fieldIDCArea   = "bk_idc_area"    // 区域
	fieldIspName   = "bk_isp_name"    // 所属运营商
	fieldCpuModule = "bk_cpu_module"  // cpu型号

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
	fieldHostDetailInfo = []string{fieldCloudID, FieldHostIP, fieldHostIPv6, fieldHostOutIP, fieldHostOutIPV6,
		fieldHostID, fieldDeviceType, fieldIDCCityName, fieldIDCCityID, fieldDeviceClass, fieldHostCPU, fieldCpuModule,
		fieldHostMem, fieldHostDisk, fieldOperator, fieldBakOperator, fieldRack, fieldIDCName, fieldSubZoneID,
		fieldIspName, fieldAgentId, FieldAssetId}

	fieldHostIPSelectorInfo = []string{FieldHostIP, fieldHostIPv6, fieldCloudID, fieldHostName, fieldOsType,
		fieldOsName, fieldHostID, fieldOperator, fieldBakOperator, fieldAgentId}
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
	and Condition = "AND" // nolint
	or  Condition = "OR"  // nolint
	in  Condition = "IN"  // nolint
)

// Page page
type Page struct {
	Start int    `json:"start"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort"`
}

// BaseResponse baseResp
type BaseResponse struct {
	Code      int    `json:"code"`
	Result    bool   `json:"result"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
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

// ListBizHostRequest list biz host request
type ListBizHostRequest struct {
	Page        Page     `json:"page"`
	BKBizID     int      `json:"bk_biz_id"`
	BKSetIDs    []int    `json:"bk_set_ids"`
	BKModuleIDs []int    `json:"bk_module_ids"`
	Fields      []string `json:"fields"`
}

// ListBizHostsResponse list biz host resp
type ListBizHostsResponse struct {
	Code      int      `json:"code"`
	Result    bool     `json:"result"`
	Message   string   `json:"message"`
	RequestID string   `json:"request_id"`
	Data      HostResp `json:"data"`
}

// HostResp host data
type HostResp struct {
	Count int        `json:"count"`
	Info  []HostData `json:"info"`
}

// HostData host info
type HostData struct {
	BKHostInnerIP   string `json:"bk_host_innerip"`
	BKHostInnerIPV6 string `json:"bk_host_innerip_v6"`
	BKHostCloudID   int    `json:"bk_cloud_id"`
	BKHostID        int64  `json:"bk_host_id"`
	BKHostName      string `json:"bk_host_name"`
	BKHostOsType    string `json:"bk_os_type"`
	BKHostOsName    string `json:"bk_os_name"`
	Operator        string `json:"operator"`
	BKBakOperator   string `json:"bk_bak_operator"`
	BkAgentID       string `json:"bk_agent_id"`
	BkAssetID       string `json:"bk_asset_id"`
}

// HostTopoRelationReq request
type HostTopoRelationReq struct {
	Page        Page  `json:"page"`
	BkBizID     int   `json:"bk_biz_id"`
	BkSetIDs    []int `json:"bk_set_ids"`
	BkModuleIDs []int `json:"bk_module_ids"`
	BkHostIDs   []int `json:"bk_host_ids"`
}

// HostTopoRelationResp host topo
type HostTopoRelationResp struct {
	BaseResponse
	Data HostTopoData `json:"data"`
}

// HostTopoData topo data
type HostTopoData struct {
	Count int                `json:"count"`
	Data  []HostTopoRelation `json:"data"`
}

// HostTopoRelation host topo relation
type HostTopoRelation struct {
	BkBizID           int    `json:"bk_biz_id"`
	BkSetID           int    `json:"bk_set_id"`
	BkModuleID        int    `json:"bk_module_id"`
	BkHostID          int    `json:"bk_host_id"`
	BkSupplierAccount string `json:"bk_supplier_account"`
}

// ListHostsWithoutBizRequest list hosts request
type ListHostsWithoutBizRequest struct {
	Page               Page                `json:"page"`
	BKBizID            int                 `json:"bk_biz_id,omitempty"`
	HostPropertyFilter *HostPropertyFilter `json:"host_property_filter"`
	Fields             []string            `json:"fields"`
}

func buildFilterConditionByStrValues(field string, values []string) *HostPropertyFilter {
	return &HostPropertyFilter{
		Condition: and.String(),
		Rules: []Rule{
			{
				Field:    field,
				Operator: "in",
				Value:    values,
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
	BkCloudID        int64  `json:"bk_cloud_id"`
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
	SubZoneID        string `json:"sub_zone_id"`
	CpuModule        string `json:"bk_cpu_module"`
}

const (
	keyBizID          = "BsiId"       // nolint
	keySvrIP          = "SvrIp"       // nolint
	methodBusiness    = "Business"    // nolint
	methodServer      = "Server"      // nolint
	methodBusinessRaw = "BusinessRaw" // nolint
)

// BizInfo business id info
type BizInfo struct {
	BizID int64 `json:"bizID"`
}

// FindHostBizRelationsRequest xxx
type FindHostBizRelationsRequest struct {
	BkHostID []int `json:"bk_host_id"`
	BkBizID  int   `json:"bk_biz_id,omitempty"`
}

// FindHostBizRelationsResponse xxx
type FindHostBizRelationsResponse struct {
	BaseResponse
	Data []HostBizRelations `json:"data"`
}

// HostBizRelations xxx
type HostBizRelations struct {
	BkBizID           int    `json:"bk_biz_id"`
	BkHostID          int    `json:"bk_host_id"`
	BkModuleID        int    `json:"bk_module_id"`
	BkSetID           int    `json:"bk_set_id"`
	BkSupplierAccount string `json:"bk_supplier_account"`
}

// TransHostToERecycleModuleRequest xxx
type TransHostToERecycleModuleRequest struct {
	BkBizID    int   `json:"bk_biz_id"`
	BkSetID    int   `json:"bk_set_id,omitempty"`
	BkModuleID int   `json:"bk_module_id,omitempty"`
	BkHostID   []int `json:"bk_host_id"`
}

// TransHostToERecycleModuleResponse xxx
type TransHostToERecycleModuleResponse struct {
	BaseResponse
}

// QueryBizInternalModuleRequest xxx
type QueryBizInternalModuleRequest struct {
	BizID int `json:"bk_biz_id"`
}

// QueryBizInternalModuleResponse xxx
type QueryBizInternalModuleResponse struct {
	BaseResponse
	Data BizInternalModuleData `json:"data"`
}

// BizInternalModuleData xxx
type BizInternalModuleData struct {
	SetID      int      `json:"bk_set_id"`
	SetName    string   `json:"bk_set_name"`
	ModuleInfo []Module `json:"module"`
}

// Module module info
type Module struct {
	ModuleID        int    `json:"bk_module_id"`
	ModuleName      string `json:"bk_module_name"`
	Default         int    `json:"default"`
	HostApplyEnable bool   `json:"host_apply_enabled"`
}

// moduleNameMaps map module name
var moduleNameMaps = map[string]string{
	"idle pool":    "空闲机池",
	"idle host":    "空闲机",
	"fault host":   "故障机",
	"recycle host": "待回收",
}

// ReplaceName replace module name
func (g *BizInternalModuleData) ReplaceName() {
	if v, ok := moduleNameMaps[g.SetName]; ok {
		g.SetName = v
	}
	for i := range g.ModuleInfo {
		if v, ok := moduleNameMaps[g.ModuleInfo[i].ModuleName]; ok {
			g.ModuleInfo[i].ModuleName = v
		}
	}
}

// TransHostAcrossBizInfo across biz
type TransHostAcrossBizInfo struct {
	SrcBizID    int
	HostID      []int
	DstBizID    int
	DstModuleID int
}

// TransferHostAcrossBizRequest xxx
type TransferHostAcrossBizRequest struct {
	SrcBizID   int   `json:"src_bk_biz_id"`
	BkHostID   []int `json:"bk_host_id"`
	DstBizID   int   `json:"dst_bk_biz_id"`
	BkModuleID int   `json:"bk_module_id"` // 主机要转移到的模块ID，该模块ID必须为下空闲机池set下的模块ID。
}

// TransferHostAcrossBizResponse xxx
type TransferHostAcrossBizResponse struct {
	BaseResponse
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

// BuildCloudAreaCondition build cloudID condition
func BuildCloudAreaCondition(cloudID int) map[string]interface{} {
	return map[string]interface{}{
		fieldCloudID: cloudID,
	}
}

// SearchCloudAreaRequest search area request
type SearchCloudAreaRequest struct {
	Condition map[string]interface{} `json:"condition"`
	Page      Page                   `json:"page"`
}

// SearchCloudAreaResp search cloud area
type SearchCloudAreaResp struct {
	BaseResponse
	Data SearchCloudAreaData `json:"data"`
}

// SearchCloudAreaData search cloud data
type SearchCloudAreaData struct {
	Count int                    `json:"count"`
	Info  []*SearchCloudAreaInfo `json:"info"`
}

// SearchCloudAreaInfo search cloud info
type SearchCloudAreaInfo struct {
	CloudID         int    `json:"bk_cloud_id"`
	CloudName       string `json:"bk_cloud_name"`
	SupplierAccount string `json:"bk_supplier_account"`
	CreateTime      string `json:"create_time"`
	LastTime        string `json:"last_time"`
}

// AddHostFromCmpyReq add host from cmpy
type AddHostFromCmpyReq struct {
	SvrIds   []string `json:"svr_ids,omitempty"`
	AssetIds []string `json:"asset_ids,omitempty"`
	InnerIps []string `json:"inner_ips,omitempty"`
}

// AddHostFromCmpyResp resp
type AddHostFromCmpyResp struct {
	BaseResponse
}

// SyncHostInfoFromCmpyReq sync host info from cmpy
type SyncHostInfoFromCmpyReq struct {
	BkHostIds []int64 `json:"bk_host_ids"`
	BkCloudId int     `json:"bk_cloud_id"`
}

// SyncHostInfoFromCmpyResp resp
type SyncHostInfoFromCmpyResp struct {
	BaseResponse
}
