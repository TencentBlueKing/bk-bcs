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
	BkBizProductor  string `json:"bk_biz_productor"`
	BkBizTester     string `json:"bk_biz_tester"`
	BkBizDeveloper  string `json:"bk_biz_developer"`
	Operator        string `json:"operator"`
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

// GetBcsPodReq defines the structure of the request for getting BCS pods.
type GetBcsPodReq struct {
	BKBizID int64               `json:"bk_biz_id"`
	Page    Page                `json:"page"`
	Fields  []string            `json:"fields"`
	Filter  *HostPropertyFilter `json:"filter"`
}

// GetBcsPodResp defines the structure of the response for getting BCS pods.
type GetBcsPodResp struct {
	BaseResponse
	Data *GetBcsPodRespData `json:"data"`
}

// GetBcsPodRespData defines the structure of the response data for getting BCS pods.
type GetBcsPodRespData struct {
	Count int    `json:"count"`
	Info  *[]Pod `json:"info"`
}

// Pod pod details
type Pod struct {
	// cc的自增主键
	ID            int64              `json:"id"`
	Name          *string            `json:"name"`
	Priority      *int32             `json:"priority"`
	Labels        *map[string]string `json:"labels"`
	IP            *string            `json:"ip"`
	NodeSelectors *map[string]string `json:"node_selectors"`
	Operator      *[]string          `json:"operator"`
}

// DeleteBcsPodReq defines the request structure for deleting a BCS pod.
type DeleteBcsPodReq struct {
	Data *[]DeleteBcsPodReqData `json:"data"`
}

// DeleteBcsPodReqData defines the data structure in the DeleteBcsPodRequest.
type DeleteBcsPodReqData struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsPodResp defines the response structure for deleting a BCS pod.
type DeleteBcsPodResp struct {
	BaseResponse
	Data interface{} `json:"data"`
}

// GetBcsWorkloadReq represents a request for getting BCS workload
type GetBcsWorkloadReq struct {
	BKBizID int64               `json:"bk_biz_id"`
	Page    Page                `json:"page"`
	Fields  []string            `json:"fields"`
	Filter  *HostPropertyFilter `json:"filter"`
	Kind    string              `json:"kind"`
}

// GetBcsWorkloadResp represents a response for getting BCS workload
type GetBcsWorkloadResp struct {
	BaseResponse
	Data *GetBcsWorkloadRespData `json:"data"`
}

// GetBcsWorkloadRespData represents the data structure of the response for getting BCS workload
type GetBcsWorkloadRespData struct {
	Count int64         `json:"count"`
	Info  []interface{} `json:"info"`
}

// DeleteBcsWorkloadReq defines the structure of the request for deleting a BCS workload.
type DeleteBcsWorkloadReq struct {
	BKBizID *int64   `json:"bk_biz_id"`
	Kind    *string  `json:"kind"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsWorkloadResp defines the structure of the response for deleting a BCS workload.
type DeleteBcsWorkloadResp struct {
	BaseResponse
	Data interface{} `json:"data"`
}

// GetBcsNamespaceReq represents the request for getting a BCS namespace
type GetBcsNamespaceReq struct {
	BKBizID int64               `json:"bk_biz_id"`
	Page    Page                `json:"page"`
	Fields  []string            `json:"fields"`
	Filter  *HostPropertyFilter `json:"filter"`
}

// Namespace define the namespace struct.
type Namespace struct {
	ID              int64              `json:"id"`
	Name            string             `json:"name"`
	Labels          *map[string]string `json:"labels"`
	SupplierAccount string             `json:"bk_supplier_account"`
}

// GetBcsNamespaceResp represents the response for getting a BCS namespace
type GetBcsNamespaceResp struct {
	BaseResponse
	Data *GetBcsNamespaceRespData `json:"data"`
}

// GetBcsNamespaceRespData represents the data for getting a BCS namespace
type GetBcsNamespaceRespData struct {
	Count int64        `json:"count"`
	Info  *[]Namespace `json:"info"`
}

// DeleteBcsNamespaceReq represents the request for deleting a BCS namespace
type DeleteBcsNamespaceReq struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsNamespaceResp represents the response for deleting a BCS namespace
type DeleteBcsNamespaceResp struct {
	BaseResponse
	Data interface{} `json:"data"`
}

// GetBcsNodeReq defines the structure of the request for getting BCS nodes.
type GetBcsNodeReq struct {
	BKBizID int64               `json:"bk_biz_id"`
	Page    Page                `json:"page"`
	Fields  []string            `json:"fields"`
	Filter  *HostPropertyFilter `json:"filter"`
}

// Node node structural description.
type Node struct {
	// ID cluster auto-increment ID in cc
	ID int64 `json:"id,omitempty" bson:"id"`
	// BizID the business ID to which the cluster belongs
	BizID int64 `json:"bk_biz_id,omitempty" bson:"bk_biz_id"`
	// SupplierAccount the supplier account that this resource belongs to.
	SupplierAccount string `json:"bk_supplier_account,omitempty" bson:"bk_supplier_account"`
	// HostID the node ID to which the host belongs
	HostID int64 `json:"bk_host_id,omitempty" bson:"bk_host_id"`
	// ClusterID the node ID to which the cluster belongs
	ClusterID int64 `json:"bk_cluster_id,omitempty" bson:"bk_cluster_id"`
	// ClusterUID the node ID to which the cluster belongs
	ClusterUID string `json:"cluster_uid,omitempty" bson:"cluster_uid"`
	// HasPod this field indicates whether there is a pod in the node.
	// if there is a pod, this field is true. If there is no pod, this
	// field is false. this field is false when node is created by default.
	HasPod           *bool     `json:"has_pod,omitempty" bson:"has_pod"`
	Name             *string   `json:"name,omitempty" bson:"name"`
	Roles            *string   `json:"roles,omitempty" bson:"roles"`
	Unschedulable    *bool     `json:"unschedulable,omitempty" bson:"unschedulable"`
	InternalIP       *[]string `json:"internal_ip,omitempty" bson:"internal_ip"`
	ExternalIP       *[]string `json:"external_ip,omitempty" bson:"external_ip"`
	HostName         *string   `json:"hostname,omitempty" bson:"hostname"`
	RuntimeComponent *string   `json:"runtime_component,omitempty" bson:"runtime_component"`
	KubeProxyMode    *string   `json:"kube_proxy_mode,omitempty" bson:"kube_proxy_mode"`
	PodCidr          *string   `json:"pod_cidr,omitempty" bson:"pod_cidr"`
}

// GetBcsNodeResp defines the structure of the response for getting BCS nodes.
type GetBcsNodeResp struct {
	BaseResponse
	Data *GetBcsNodeRespData `json:"data"`
}

// GetBcsNodeRespData defines the structure of the response data for getting BCS nodes.
type GetBcsNodeRespData struct {
	Count int64   `json:"count"`
	Info  *[]Node `json:"info"`
}

// DeleteBcsNodeReq defines the structure of the request for deleting BCS nodes.
type DeleteBcsNodeReq struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsNodeResp defines the structure of the response for deleting BCS nodes.
type DeleteBcsNodeResp struct {
	BaseResponse
	Data interface{} `json:"data"`
}

// GetBcsClusterReq defines the request structure for getting BCS cluster information.
type GetBcsClusterReq struct {
	BKBizID int64               `json:"bk_biz_id"`
	Page    Page                `json:"page"`
	Fields  []string            `json:"fields"`
	Filter  *HostPropertyFilter `json:"filter"`
}

// Cluster container cluster table structure
type Cluster struct {
	// ID cluster auto-increment ID in cc
	ID int64 `json:"id" bson:"id"`
	// BizID the business ID to which the cluster belongs
	BizID int64 `json:"bk_biz_id" bson:"bk_biz_id"`
	// SupplierAccount the supplier account that this resource belongs to.
	SupplierAccount string `json:"bk_supplier_account" bson:"bk_supplier_account"`
	// Name cluster name.
	Name *string `json:"name,omitempty" bson:"name"`
	// SchedulingEngine scheduling engines, such as k8s, tke, etc.
	SchedulingEngine *string `json:"scheduling_engine,omitempty" bson:"scheduling_engine"`
	// Uid ID of the cluster itself
	Uid *string `json:"uid,omitempty" bson:"uid"`
	// Xid The underlying cluster ID it depends on
	Xid *string `json:"xid,omitempty" bson:"xid"`
	// Version cluster version
	Version *string `json:"version,omitempty" bson:"version"`
	// NetworkType network type, such as overlay or underlay
	NetworkType *string `json:"network_type,omitempty" bson:"network_type"`
	// Region the region where the cluster is located
	Region *string `json:"region,omitempty" bson:"region"`
	// Vpc vpc network
	Vpc *string `json:"vpc,omitempty" bson:"vpc"`
	// Environment cluster environment
	Environment *string `json:"environment,omitempty" bson:"environment"`
	// NetWork global routing network address (container overlay network) For example: ["1.1.1.0/21"]
	NetWork *[]string `json:"network,omitempty" bson:"network"`
}

// GetBcsClusterResp defines the response structure for getting BCS cluster information.
type GetBcsClusterResp struct {
	BaseResponse
	Data GetBcsClusterRespData `json:"data"`
}

// GetBcsClusterRespData defines the data structure for getting BCS cluster information.
type GetBcsClusterRespData struct {
	Count int64     `json:"count"`
	Info  []Cluster `json:"info"`
}

// DeleteBcsClusterReq represents the request for deleting a BCS cluster
type DeleteBcsClusterReq struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsClusterResp represents the response for deleting a BCS cluster
type DeleteBcsClusterResp struct {
	BaseResponse
	Data interface{} `json:"data"`
}
