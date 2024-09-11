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

package client

import (
	"context"
	"crypto/tls"
	"math/big"

	bkcmdbkube "configcenter/src/kube/types" // nolint
	pmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	bcsregistry "github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
)

// field result
const (
	FieldBS2NameID = "bs2_name_id"
)

// condition result
const (
	ConditionBkBizID = "bk_biz_id"
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

// ListBizHostRequest list biz host request
type ListBizHostRequest struct {
	Page        Page     `json:"page"`
	BKBizID     int      `json:"bk_biz_id"`
	BKSetIDs    []int    `json:"bk_set_ids"`
	BKModuleIDs []int    `json:"bk_module_ids"`
	Fields      []string `json:"fields"`
}

// ListBizHostsResponse list biz host response
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

// HostData host data
type HostData struct {
	HostInnerIP        string `json:"bk_host_innerip"`
	SvrTypeName        string `json:"svr_type_name"`
	BkSvrDeviceClsName string `json:"bk_svr_device_cls_name"`
	BkServiceArr       string `json:"bk_service_arr"`
	BkSvcIdArr         string `json:"bk_svc_id_arr"`
	IdcCityId          string `json:"idc_city_id"`
	IdcCityName        string `json:"idc_city_name"`
	BkHostId           int64  `json:"bk_host_id"`
}

const (
	// KeyBizID xxx
	KeyBizID       = "BsiId"
	methodBusiness = "Business"
	methodServer   = "Server"
	// MethodBusinessRaw xxx
	MethodBusinessRaw = "BusinessRaw"
)

var (
	// ReqColumns xxx
	ReqColumns = []string{"BsiId", "BsipId", "BsiProductName", "BsiProductId", "BsiName", "BsiL1", "BsiL2"}
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
	BsiL1          int    `json:"BsiL1"`
	BsiL2          int    `json:"BsiL2"`
}

// BizInfo business id info
type BizInfo struct {
	BizID int64 `json:"bizID"`
}

// ClusterManagerClientWithHeader client for cluster manager
type ClusterManagerClientWithHeader struct {
	Cli cmp.ClusterManagerClient
	Ctx context.Context
}

// ProjectManagerClientWithHeader client for project manager
type ProjectManagerClientWithHeader struct {
	Cli pmp.BCSProjectClient
	Ctx context.Context
}

// Config for bcsapi
type Config struct {
	// bcsapi host, available like 127.0.0.1:8080
	Hosts []string
	// tls configuratio
	TLSConfig *tls.Config
	// AuthToken for permission verification
	AuthToken string
	// UserName for permission verification
	Username string
	// clusterID for Kubernetes/Mesos operation
	ClusterID string
	// proxy flag for go through bcs-api-gateway
	Gateway bool
	// etcd registry config for bcs modules
	Etcd bcsregistry.CMDOptions
}

// GetBcsClusterRequest defines the request structure for getting BCS cluster information.
type GetBcsClusterRequest struct {
	CommonRequest
}

// GetBcsClusterResponse defines the response structure for getting BCS cluster information.
type GetBcsClusterResponse struct {
	CommonResponse
	Data GetBcsClusterResponseData `json:"data"`
}

// GetBcsClusterResponseData defines the data structure for getting BCS cluster information.
type GetBcsClusterResponseData struct {
	Count int64                `json:"count"`
	Info  []bkcmdbkube.Cluster `json:"info"`
}

// CommonRequest defines the common request structure for BCS cluster operations.
type CommonRequest struct {
	BKBizID int64           `json:"bk_biz_id"`
	Page    Page            `json:"page"`
	Fields  []string        `json:"fields"`
	Filter  *PropertyFilter `json:"filter"`
}

// CommonResponse defines the common response structure for BCS cluster operations.
type CommonResponse struct {
	Code      int64  `json:"code"`
	Result    bool   `json:"result"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// CreateBcsClusterRequest defines the request structure for creating a BCS cluster.
type CreateBcsClusterRequest struct {
	BKBizID          *int64    `json:"bk_biz_id"`
	Name             *string   `json:"name"`
	SchedulingEngine *string   `json:"scheduling_engine"`
	UID              *string   `json:"uid"`
	XID              *string   `json:"xid"`
	Version          *string   `json:"version"`
	NetworkType      *string   `json:"network_type"`
	Region           *string   `json:"region"`
	Vpc              *string   `json:"vpc"`
	Network          *[]string `json:"network"`
	Type             *string   `json:"type"`
	Environment      *string   `json:"environment"`
	// BKProjectID      *string   `json:"bk_project_id"`
	// BKProjectName    *string   `json:"bk_project_name"`
	// BKProjectCode    *string   `json:"bk_project_code"`
}

// CreateBcsClusterResponse defines the response structure for creating a BCS cluster.
type CreateBcsClusterResponse struct {
	CommonResponse
	Data *CreateBcsClusterResponseData `json:"data"`
}

// CreateBcsClusterResponseData defines the data structure for creating a BCS cluster.
type CreateBcsClusterResponseData struct {
	ID int64 `json:"id"`
}

// UpdateBcsClusterRequest defines the request structure for updating a BCS cluster.
type UpdateBcsClusterRequest struct {
	BKBizID *int64                       `json:"bk_biz_id"`
	IDs     *[]int64                     `json:"ids"`
	Data    *UpdateBcsClusterRequestData `json:"data"`
}

// UpdateBcsClusterResponse represents the response for updating a BCS cluster
type UpdateBcsClusterResponse struct {
	CommonResponse
	Data interface{} `json:"data"`
}

// UpdateBcsClusterTypeRequest represents the request for updating a BCS cluster type
type UpdateBcsClusterTypeRequest struct {
	BKBizID *int64  `json:"bk_biz_id"`
	ID      *int64  `json:"id"`
	Type    *string `json:"type"`
}

// UpdateBcsClusterTypeResponse represents the response for updating a BCS cluster type
type UpdateBcsClusterTypeResponse struct {
	CommonResponse
	Data *CreateBcsWorkloadResponseData `json:"data"`
}

// UpdateBcsNamespaceRequest represents the request for updating a BCS namespace
type UpdateBcsNamespaceRequest struct {
	BKBizID *int64                         `json:"bk_biz_id"`
	IDs     *[]int64                       `json:"ids"`
	Data    *UpdateBcsNamespaceRequestData `json:"data"`
}

// UpdateBcsNamespaceRequestData represents the data for updating a BCS namespace
type UpdateBcsNamespaceRequestData struct {
	Labels         *map[string]string          `json:"labels"`
	ResourceQuotas *[]bkcmdbkube.ResourceQuota `json:"resource_quotas"`
}

// UpdateBcsNamespaceResponse represents the response for updating a BCS namespace
type UpdateBcsNamespaceResponse struct {
	CommonResponse
	Data interface{} `json:"data"`
}

// UpdateBcsClusterRequestData represents the data for updating a BCS cluster
type UpdateBcsClusterRequestData struct {
	Name             *string   `json:"name"`
	XID              *string   `json:"xid"`
	SchedulingEngine *string   `json:"scheduling_engine"`
	VPC              *string   `json:"vpc"`
	Version          *string   `json:"version"`
	NetworkType      *string   `json:"network_type"`
	Region           *string   `json:"region"`
	Type             *string   `json:"type"`
	Network          *[]string `json:"network"`
	Environment      *string   `json:"environment"`
}

// UpdateBcsClusterRequestDataCluster represents the data for updating a BCS cluster
type UpdateBcsClusterRequestDataCluster struct {
	XID              *string   `json:"xid"`
	SchedulingEngine *string   `json:"scheduling_engine"`
	VPC              *string   `json:"vpc"`
	Version          *string   `json:"version"`
	NetworkType      *string   `json:"network_type"`
	Region           *string   `json:"region"`
	Type             *string   `json:"type"`
	Network          *[]string `json:"network"`
}

// DeleteBcsClusterRequest represents the request for deleting a BCS cluster
type DeleteBcsClusterRequest struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsClusterResponse represents the response for deleting a BCS cluster
type DeleteBcsClusterResponse struct {
	CommonResponse
	Data interface{} `json:"data"`
}

// DeleteBcsNamespaceRequest represents the request for deleting a BCS namespace
type DeleteBcsNamespaceRequest struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsNamespaceResponse represents the response for deleting a BCS namespace
type DeleteBcsNamespaceResponse struct {
	CommonResponse
	Data interface{} `json:"data"`
}

// GetBcsNamespaceRequest represents the request for getting a BCS namespace
type GetBcsNamespaceRequest struct {
	CommonRequest
}

// GetBcsNamespaceResponse represents the response for getting a BCS namespace
type GetBcsNamespaceResponse struct {
	CommonResponse
	Data *GetBcsNamespaceResponseData `json:"data"`
}

// GetBcsNamespaceResponseData represents the data for getting a BCS namespace
type GetBcsNamespaceResponseData struct {
	Count int64                   `json:"count"`
	Info  *[]bkcmdbkube.Namespace `json:"info"`
}

// CreateBcsNamespaceRequest represents the request for creating a BCS namespace
type CreateBcsNamespaceRequest struct {
	BKBizID *int64 `json:"bk_biz_id"`
	// Data    []CreateBcsNamespaceRequestData `json:"data"`
	Data *[]bkcmdbkube.Namespace `json:"data"`
}

// CreateBcsNamespaceResponse represents the response for creating a BCS namespace
type CreateBcsNamespaceResponse struct {
	CommonResponse
	Data *CreateBcsNamespaceResponseData `json:"data"`
}

// CreateBcsNamespaceResponseData represents the data for creating a BCS namespace
type CreateBcsNamespaceResponseData struct {
	IDs []int64 `json:"ids"`
}

// GetBcsWorkloadRequest represents a request for getting BCS workload
type GetBcsWorkloadRequest struct {
	CommonRequest
	Kind string `json:"kind"`
}

// GetBcsWorkloadResponse represents a response for getting BCS workload
type GetBcsWorkloadResponse struct {
	CommonResponse
	Data *GetBcsWorkloadResponseData `json:"data"`
}

// GetBcsWorkloadResponseData represents the data structure of the response for getting BCS workload
type GetBcsWorkloadResponseData struct {
	Count int64         `json:"count"`
	Info  []interface{} `json:"info"`
}

// CreateBcsWorkloadRequest represents a request for creating BCS workload
type CreateBcsWorkloadRequest struct {
	BKBizID *int64                          `json:"bk_biz_id"`
	Kind    *string                         `json:"kind"`
	Data    *[]CreateBcsWorkloadRequestData `json:"data"`
}

// CreateBcsWorkloadRequestData defines the structure of the request data for creating a BCS workload.
type CreateBcsWorkloadRequestData struct {
	NamespaceID           *int64                    `json:"bk_namespace_id,omitempty" bson:"bk_namespace_id"`
	Name                  *string                   `json:"name,omitempty" bson:"name"`
	Labels                *map[string]string        `json:"labels,omitempty" bson:"labels"`
	Selector              *bkcmdbkube.LabelSelector `json:"selector,omitempty" bson:"selector"`
	Replicas              *int64                    `json:"replicas,omitempty" bson:"replicas"`
	MinReadySeconds       *int64                    `json:"min_ready_seconds,omitempty" bson:"min_ready_seconds"`
	StrategyType          *string                   `json:"strategy_type,omitempty" bson:"strategy_type"`
	RollingUpdateStrategy *map[string]interface{}   `json:"rolling_update_strategy,omitempty" bson:"rolling_update_strategy"` // nolint
}

// CreateBcsWorkloadResponse defines the structure of the response for creating a BCS workload.
type CreateBcsWorkloadResponse struct {
	CommonResponse
	Data *CreateBcsWorkloadResponseData `json:"data"`
}

// CreateBcsWorkloadResponseData defines the structure of the response data for creating a BCS workload.
type CreateBcsWorkloadResponseData struct {
	IDs []int64 `json:"ids"`
}

// UpdateBcsWorkloadRequest defines the structure of the request for updating a BCS workload.
type UpdateBcsWorkloadRequest struct {
	BKBizID *int64                        `json:"bk_biz_id"`
	Kind    *string                       `json:"kind"`
	IDs     *[]int64                      `json:"ids"`
	Data    *UpdateBcsWorkloadRequestData `json:"data"`
}

// UpdateBcsWorkloadRequestData defines the structure of the request data for updating a BCS workload.
type UpdateBcsWorkloadRequestData struct {
	Labels                *map[string]string        `json:"labels,omitempty" bson:"labels"`
	Selector              *bkcmdbkube.LabelSelector `json:"selector,omitempty" bson:"selector"`
	Replicas              *int64                    `json:"replicas,omitempty" bson:"replicas"`
	MinReadySeconds       *int64                    `json:"min_ready_seconds,omitempty" bson:"min_ready_seconds"`
	StrategyType          *string                   `json:"strategy_type,omitempty" bson:"strategy_type"`
	RollingUpdateStrategy *map[string]interface{}   `json:"rolling_update_strategy,omitempty" bson:"rolling_update_strategy"` // nolint
}

// UpdateBcsWorkloadResponse defines the structure of the response for updating a BCS workload.
type UpdateBcsWorkloadResponse struct {
	CommonResponse
	Data interface{} `json:"data"`
}

// DeleteBcsWorkloadRequest defines the structure of the request for deleting a BCS workload.
type DeleteBcsWorkloadRequest struct {
	BKBizID *int64   `json:"bk_biz_id"`
	Kind    *string  `json:"kind"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsWorkloadResponse defines the structure of the response for deleting a BCS workload.
type DeleteBcsWorkloadResponse struct {
	CommonResponse
	Data interface{} `json:"data"`
}

// GetBcsNodeRequest defines the structure of the request for getting BCS nodes.
type GetBcsNodeRequest struct {
	CommonRequest
}

// GetBcsNodeResponse defines the structure of the response for getting BCS nodes.
type GetBcsNodeResponse struct {
	CommonResponse
	Data *GetBcsNodeResponseData `json:"data"`
}

// GetBcsNodeResponseData defines the structure of the response data for getting BCS nodes.
type GetBcsNodeResponseData struct {
	Count int64              `json:"count"`
	Info  *[]bkcmdbkube.Node `json:"info"`
}

// CreateBcsNodeRequest defines the structure of the request for creating BCS nodes.
type CreateBcsNodeRequest struct {
	BKBizID *int64                      `json:"bk_biz_id"`
	Data    *[]CreateBcsNodeRequestData `json:"data"`
}

// CreateBcsNodeRequestData defines the structure of the request data for creating BCS nodes.
type CreateBcsNodeRequestData struct {
	HostID           *int64             `json:"bk_host_id,omitempty" bson:"bk_host_id"`
	ClusterID        *int64             `json:"bk_cluster_id,omitempty" bson:"bk_cluster_id"`
	Name             *string            `json:"name,omitempty" bson:"name"`
	Labels           *map[string]string `json:"labels,omitempty" bson:"labels"`
	Taints           *map[string]string `json:"taints,omitempty" bson:"taints"`
	Unschedulable    *bool              `json:"unschedulable,omitempty" bson:"unschedulable"`
	InternalIP       *[]string          `json:"internal_ip,omitempty" bson:"internal_ip"`
	ExternalIP       *[]string          `json:"external_ip,omitempty" bson:"external_ip"`
	HostName         *string            `json:"host_name,omitempty" bson:"host_name"`
	RuntimeComponent *string            `json:"runtime_component,omitempty" bson:"runtime_component"`
	KubeProxyMode    *string            `json:"kube_proxy_mode,omitempty" bson:"kube_proxy_mode"`
	PodCidr          *string            `json:"pod_cidr,omitempty" bson:"pod_cidr"`
}

// CreateBcsNodeResponse defines the structure of the response for creating BCS nodes.
type CreateBcsNodeResponse struct {
	CommonResponse
	Data *CreateBcsNodeResponseData `json:"data"`
}

// CreateBcsNodeResponseData defines the structure of the response data for creating BCS nodes.
type CreateBcsNodeResponseData struct {
	IDs []int64 `json:"ids"`
}

// UpdateBcsNodeRequest defines the structure of the request for updating BCS nodes.
type UpdateBcsNodeRequest struct {
	BKBizID *int64                    `json:"bk_biz_id"`
	IDs     *[]int64                  `json:"ids"`
	Data    *UpdateBcsNodeRequestData `json:"data"`
}

// UpdateBcsNodeRequestData defines the structure of the request data for updating BCS nodes.
type UpdateBcsNodeRequestData struct {
	Labels           *map[string]string `json:"labels,omitempty" bson:"labels"`
	Taints           *map[string]string `json:"taints,omitempty" bson:"taints"`
	Unschedulable    *bool              `json:"unschedulable,omitempty" bson:"unschedulable"`
	Hostname         *string            `json:"hostname,omitempty" bson:"hostname"`
	RuntimeComponent *string            `json:"runtime_component,omitempty" bson:"runtime_component"`
	KubeProxyMode    *string            `json:"kube_proxy_mode,omitempty" bson:"kube_proxy_mode"`
	PodCidr          *string            `json:"pod_cidr,omitempty" bson:"pod_cidr"`
}

// UpdateBcsNodeResponse defines the structure of the response for updating BCS nodes.
type UpdateBcsNodeResponse struct {
	CommonResponse
	Data interface{} `json:"data"`
}

// DeleteBcsNodeRequest defines the structure of the request for deleting BCS nodes.
type DeleteBcsNodeRequest struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsNodeResponse defines the structure of the response for deleting BCS nodes.
type DeleteBcsNodeResponse struct {
	CommonResponse
	Data interface{} `json:"data"`
}

// GetBcsPodRequest defines the structure of the request for getting BCS pods.
type GetBcsPodRequest struct {
	CommonRequest
}

// GetBcsPodResponse defines the structure of the response for getting BCS pods.
type GetBcsPodResponse struct {
	CommonResponse
	Data *GetBcsPodResponseData `json:"data"`
}

// GetBcsPodResponseData defines the structure of the response data for getting BCS pods.
type GetBcsPodResponseData struct {
	Count int               `json:"count"`
	Info  *[]bkcmdbkube.Pod `json:"info"`
}

// GetBcsContainerRequest 表示获取BCS容器请求的结构体
type GetBcsContainerRequest struct {
	CommonRequest       // 嵌入通用请求结构体
	BkPodID       int64 `json:"bk_pod_id"` // BkPodID定了要查询的Pod的ID
}

// GetBcsContainerResponse 表示获取BCS容器响应的结构体
type GetBcsContainerResponse struct {
	CommonResponse                              // 嵌入通用响应结
	Data           *GetBcsContainerResponseData `json:"data"` // 包含响应数据的指针
}

// GetBcsContainerResponseData 包含实际的响应数据
type GetBcsContainerResponseData struct {
	Count int          `json:"count"` // 返回的容器数量
	Info  *[]Container `json:"info"`  // 容器的详细信息
}

// CreateBcsPodRequest defines the structure of the request for creating BCS pods.
type CreateBcsPodRequest struct {
	Data *[]CreateBcsPodRequestData `json:"data"`
}

// CreateBcsPodRequestData defines the structure of the request data for creating a BCS pod.
type CreateBcsPodRequestData struct {
	BizID *int64                        `json:"bk_biz_id" bson:"bk_biz_id"`
	Pods  *[]CreateBcsPodRequestDataPod `json:"pods" bson:"pods"`
}

// CreateBcsPodRequestDataPod defines the structure of a pod in the CreateBcsPodRequestData.
type CreateBcsPodRequestDataPod struct {
	Spec       *CreateBcsPodRequestPodSpec       `json:"spec" bson:"spec"`
	Name       *string                           `json:"name" bson:"name"`
	HostID     *int64                            `json:"bk_host_id,omitempty" bson:"bk_host_id"`
	Operator   *[]string                         `json:"operator,omitempty" bson:"operator"`
	Priority   *int32                            `json:"priority,omitempty" bson:"priority"`
	Labels     *map[string]string                `json:"labels,omitempty" bson:"labels"`
	IP         *string                           `json:"ip,omitempty" bson:"ip"`
	IPs        *[]bkcmdbkube.PodIP               `json:"ips,omitempty"  bson:"ips"`
	Containers *[]bkcmdbkube.ContainerBaseFields `json:"containers,omitempty" bson:"containers"`
}

// CreateBcsPodRequestPodSpec defines the structure of the pod spec in the CreateBcsPodRequestData.
type CreateBcsPodRequestPodSpec struct {
	ClusterID    *int64                `json:"bk_cluster_id,omitempty" bson:"bk_cluster_id"`
	NameSpaceID  *int64                `json:"bk_namespace_id,omitempty" bson:"bk_namespace_id"`
	WorkloadKind *string               `json:"workload_kind,omitempty" bson:"workload_kind"`
	WorkloadID   *int64                `json:"workload_id,omitempty" bson:"workload_id"`
	NodeID       *int64                `json:"bk_node_id,omitempty" bson:"bk_node_id"`
	Ref          *bkcmdbkube.Reference `json:"ref,omitempty" bson:"ref"`
}

// CreateBcsPodResponse defines the response structure for creating a BCS pod.
type CreateBcsPodResponse struct {
	CommonResponse
	Data *CreateBcsPodResponseData `json:"data"`
}

// CreateBcsPodResponseData defines the data structure in the CreateBcsPodResponse.
type CreateBcsPodResponseData struct {
	IDs []int64 `json:"ids"`
}

// UpdateBcsPodRequest defines the request structure for updating a BCS pod.
type UpdateBcsPodRequest struct {
}

// UpdateBcsPodResponse defines the response structure for updating a BCS pod.
type UpdateBcsPodResponse struct {
	CommonResponse
	Data interface{} `json:"data"`
}

// DeleteBcsPodRequest defines the request structure for deleting a BCS pod.
type DeleteBcsPodRequest struct {
	Data *[]DeleteBcsPodRequestData `json:"data"`
}

// DeleteBcsPodRequestData defines the data structure in the DeleteBcsPodRequest.
type DeleteBcsPodRequestData struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// DeleteBcsPodResponse defines the response structure for deleting a BCS pod.
type DeleteBcsPodResponse struct {
	CommonResponse
	Data interface{} `json:"data"`
}

// ListHostsWithoutBizRequest list host without biz request
type ListHostsWithoutBizRequest struct {
	Page               Page           `json:"page"`
	Fields             []string       `json:"fields"`
	HostPropertyFilter PropertyFilter `json:"host_property_filter"`
}

// ListHostsByBizRequest 结构体用于请求按业务ID列出主机信息
type ListHostsByBizRequest struct {
	// Page 分页信息
	Page Page `json:"page"`
	// BkBizID 业务ID
	BkBizID int64 `json:"bk_biz_id"`
	// Fields 需要返回的字段列表
	Fields []string `json:"fields"`
	// HostPropertyFilter 主机属性过滤条件
	HostPropertyFilter PropertyFilter `json:"host_property_filter"`
}

// ListHostsWithoutBizResponse list host without biz response
type ListHostsWithoutBizResponse struct {
	Code      int      `json:"code"`
	Result    bool     `json:"result"`
	Message   string   `json:"message"`
	RequestID string   `json:"request_id"`
	Data      HostResp `json:"data"`
}

// PropertyFilter property filter
type PropertyFilter struct {
	Condition string `json:"condition"`
	Rules     []Rule `json:"rules"`
}

// Rule rule
type Rule struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// DeleteBcsClusterAllRequest represents the request for deleting a BCS cluster
type DeleteBcsClusterAllRequest struct {
	BKBizID *int64   `json:"bk_biz_id"`
	IDs     *[]int64 `json:"ids"`
}

// Container container details
type Container struct {
	// cc的自增主键
	ID    int64 `json:"id,omitempty" bson:"id"`
	PodID int64 `json:"bk_pod_id,omitempty" bson:"bk_pod_id"`
	// ClusterID cluster id in cc
	ClusterID           int64 `json:"bk_cluster_id,omitempty" bson:"bk_cluster_id"`
	ContainerBaseFields `json:",inline" bson:",inline"`
	// Revision record this app's revision information
	Revision `json:",inline" bson:",inline"`
}

// Revision resource revision information.
type Revision struct {
	Creator    string `json:"creator,omitempty" bson:"creator"`
	Modifier   string `json:"modifier,omitempty" bson:"modifier"`
	CreateTime int64  `json:"create_time,omitempty" bson:"create_time"`
	LastTime   int64  `json:"last_time,omitempty" bson:"last_time"`
}

// ContainerBaseFields container core details
type ContainerBaseFields struct {
	Name            *string          `json:"name,omitempty" bson:"name"`
	ContainerID     *string          `json:"container_uid,omitempty" bson:"container_uid"`
	Image           *string          `json:"image,omitempty" bson:"image"`
	Ports           *[]ContainerPort `json:"ports,omitempty" bson:"ports"`
	HostPorts       *[]ContainerPort `json:"host_ports,omitempty" bson:"host_ports"`
	Args            *[]string        `json:"args,omitempty" bson:"args"`
	Started         *int64           `json:"started,omitempty" bson:"started"`
	Limits          *ResourceList    `json:"limits,omitempty" bson:"limits"`
	ReqSysSpecuests *ResourceList    `json:"requests,omitempty" bson:"requests"`
	Liveness        *Probe           `json:"liveness,omitempty" bson:"liveness"`
	Environment     *[]EnvVar        `json:"environment,omitempty" bson:"environment"`
	Mounts          *[]VolumeMount   `json:"mounts,omitempty" bson:"mounts"`
}

// ContainerPort represents a network port in a single container.
type ContainerPort struct {
	// If specified, this must be an IANA_SVC_NAME and unique within the pod. Each
	// named port in a pod must have a unique name. Name for the port that can be
	// referred to by services.
	// +optional
	Name string `json:"name,omitempty" bson:"name"`
	// Number of port to expose on the host.
	// If specified, this must be a valid port number, 0 < x < 65536.
	// If HostNetwork is specified, this must match ContainerPort.
	// Most containers do not need this.
	// +optional
	HostPort int32 `json:"hostPort,omitempty" bson:"hostPort"`
	// Number of port to expose on the pod's IP address.
	// This must be a valid port number, 0 < x < 65536.
	ContainerPort int32 `json:"containerPort" bson:"containerPort"`
	// Protocol for port. Must be UDP, TCP, or SCTP.
	// Defaults to "TCP".
	// +optional
	// +default="TCP"
	Protocol Protocol `json:"protocol,omitempty" bson:"protocol"`
	// What host IP to bind the external port to.
	// +optional
	HostIP string `json:"hostIP,omitempty" bson:"hostIP"`
}

// Protocol defines network protocols supported for things like container ports.
// +enum
type Protocol string

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[ResourceName]Quantity

// ResourceName is the name identifying various resources in a ResourceList.
type ResourceName string

// Quantity 结构体表示一个数量，可能用于库存管理、订单处理等场景。
// 该结构体可以包含如数量值和单位等信息。
type Quantity struct {
	// i is the quantity in int64 scaled form, if d.Dec == nil
	i int64Amount //nolint:unused
	// d is the quantity in inf.Dec form if d.Dec != nil
	d infDecAmount //nolint:unused
	// s is the generated value of this quantity to avoid recalculation
	s string //nolint:unused

	// Change Format at will. See the comment for Canonicalize for
	// more details.
	Format
}

// int64Amount represents a fixed precision numerator and arbitrary scale exponent. It is faster
// than operations on inf.Dec for values that can be represented as int64.
// +k8s:openapi-gen=true
type int64Amount struct { //nolint:unused
	value int64
	scale Scale
}

// infDecAmount implements common operations over an inf.Dec that are specific to the quantity
// representation.
type infDecAmount struct { //nolint:unused
	*Dec
}

// Dec 结构体表示一个十进制数，包含未缩放的整数部分和缩放比例。
type Dec struct {
	unscaled big.Int //nolint:unused
	scale    Scale   //nolint:unused
}

// Scale is used for getting and setting the base-10 scaled value.
// Base-2 scales are omitted for mathematical simplicity.
// See Quantity.ScaledValue for more details.
type Scale int32

// Format lists the three possible formattings of a quantity.
type Format string

// Probe describes a health check to be performed against a container to determine whether it is
// alive or ready to receive traffic.
type Probe struct {
	// The action taken to determine the health of a container
	ProbeHandler `json:",inline" bson:",inline"`
	// Number of seconds after the container has started before liveness probes are initiated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty" bson:"initialDelaySeconds"`
	// Number of seconds after which the probe times out.
	// Defaults to 1 second. Minimum value is 1.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty" bson:"timeoutSeconds"`
	// How often (in seconds) to perform the probe.
	// Default to 10 seconds. Minimum value is 1.
	// +optional
	PeriodSeconds int32 `json:"periodSeconds,omitempty" bson:"periodSeconds"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	// Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.
	// +optional
	SuccessThreshold int32 `json:"successThreshold,omitempty" bson:"successThreshold"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	// Defaults to 3. Minimum value is 1.
	// +optional
	FailureThreshold int32 `json:"failureThreshold,omitempty" bson:"failureThreshold"`
	// Optional duration in seconds the pod needs to terminate gracefully upon probe failure.
	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with a kill signal.
	// Set this value longer than the expected cleanup time for your process.
	// If this value is nil, the pod's terminationGracePeriodSeconds will be used. Otherwise, this
	// value overrides the value provided by the pod spec.
	// Value must be non-negative integer. The value zero indicates stop immediately via
	// the kill signal (no opportunity to shut down).
	// This is a beta field and requires enabling ProbeTerminationGracePeriod feature gate.
	// Minimum value is 1. spec.terminationGracePeriodSeconds is used if unset.
	// +optional
	// NOCC:tosa/linelength(忽略长度)
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty" bson:"terminationGracePeriodSeconds"` // nolint
}

// ProbeHandler defines a specific action that should be taken in a probe.
// One and only one of the fields must be specified.
type ProbeHandler struct {
	// Exec specifies the action to take.
	// +optional
	Exec *ExecAction `json:"exec,omitempty" bson:"exec"`
	// HTTPGet specifies the http request to perform.
	// +optional
	HTTPGet *HTTPGetAction `json:"httpGet,omitempty" bson:"httpGet"`
	// TCPSocket specifies an action involving a TCP port.
	// +optional
	TCPSocket *TCPSocketAction `json:"tcpSocket,omitempty" bson:"tcpSocket"`

	// GRPC specifies an action involving a GRPC port.
	// This is a beta field and requires enabling GRPCContainerProbe feature gate.
	// +featureGate=GRPCContainerProbe
	// +optional
	GRPC *GRPCAction `json:"grpc,omitempty" bson:"grpc"`
}

// ExecAction describes a "run in container" action.
type ExecAction struct {
	// Command is the command line to execute inside the container, the working directory for the
	// command  is root ('/') in the container's filesystem. The command is simply exec'd, it is
	// not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use
	// a shell, you need to explicitly call out to that shell.
	// Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
	// +optional
	Command []string `json:"command,omitempty" bson:"command"`
}

// HTTPGetAction describes an action based on HTTP Get requests.
type HTTPGetAction struct {
	// Path to access on the HTTP server.
	// +optional
	Path string `json:"path,omitempty" bson:"path"`
	// Name or number of the port to access on the container.
	// Number must be in the range 1 to 65535.
	// Name must be an IANA_SVC_NAME.
	Port IntOrString `json:"port" bson:"port"`
	// Host name to connect to, defaults to the pod IP. You probably want to set
	// "Host" in httpHeaders instead.
	// +optional
	Host string `json:"host,omitempty" bson:"host"`
	// Scheme to use for connecting to the host.
	// Defaults to HTTP.
	// +optional
	Scheme URIScheme `json:"scheme,omitempty" bson:"scheme"`
	// Custom headers to set in the request. HTTP allows repeated headers.
	// +optional
	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty" bson:"httpHeaders"`
}

// TCPSocketAction describes an action based on opening a socket
type TCPSocketAction struct {
	// Number or name of the port to access on the container.
	// Number must be in the range 1 to 65535.
	// Name must be an IANA_SVC_NAME.
	Port IntOrString `json:"port" bson:"port"`
	// Optional: Host name to connect to, defaults to the pod IP.
	// +optional
	Host string `json:"host,omitempty" bson:"host"`
}

// GRPCAction grpc service
type GRPCAction struct {
	// Port number of the gRPC service. Number must be in the range 1 to 65535.
	Port int32 `json:"port" bson:"port"`

	// Service is the name of the service to place in the gRPC HealthCheckRequest
	// (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md).
	//
	// If this is not specified, the default behavior is defined by gRPC.
	// +optional
	// +default=""
	Service *string `json:"service" bson:"service"`
}

// IntOrString is a type that can hold an int32 or a string.
type IntOrString struct {
	Type   Type   `json:"type" bson:"type"`
	IntVal int32  `json:"int_val" bson:"int_val"`
	StrVal string `json:"str_val" bson:"str_val"`
}

// Type represents the stored type of IntOrString.
type Type int64

// URIScheme identifies the scheme used for connection to a host for Get actions
// +enum
type URIScheme string

// HTTPHeader describes a custom header to be used in HTTP probes
type HTTPHeader struct {
	// The header field name
	Name string `json:"name" bson:"name"`
	// The header field value
	Value string `json:"value" bson:"value"`
}

// EnvVar represents an environment variable present in a Container.
type EnvVar struct {
	// Name of the environment variable. Must be a C_IDENTIFIER.
	Name string `json:"name" bson:"name"`

	// Optional: no more than one of the following may be specified.

	// Variable references $(VAR_NAME) are expanded
	// using the previously defined environment variables in the container and
	// any service environment variables. If a variable cannot be resolved,
	// the reference in the input string will be unchanged. Double $$ are reduced
	// to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e.
	// "$$(VAR_NAME)" will produce the string literal "$(VAR_NAME)".
	// Escaped references will never be expanded, regardless of whether the variable
	// exists or not.
	// Defaults to "".
	// +optional
	Value string `json:"value,omitempty" bson:"value"`
	// Source for the environment variable's value. Cannot be used if value is not empty.
	// +optional
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty" bson:"valueFrom"`
}

// EnvVarSource represents a source for the value of an EnvVar.
type EnvVarSource struct {
	// Selects a field of the pod: supports metadata.name, metadata.namespace, `metadata.labels['<KEY>']`,
	// `metadata.annotations['<KEY>']`, spec.nodeName, spec.serviceAccountName, status.hostIP, status.podIP,
	// status.podIPs.
	// +optional
	FieldRef *ObjectFieldSelector `json:"fieldRef,omitempty" bson:"fieldRef"`
	// Selects a resource of the container: only resources limits and requests
	// (limits.cpu, limits.memory, limits.ephemeral-storage, requests.cpu, requests.memory and
	// requests.ephemeral-storage) are currently supported.
	// +optional
	// NOCC:tosa/linelength(忽略长度)
	ResourceFieldRef *ResourceFieldSelector `json:"resourceFieldRef,omitempty" bson:"resourceFieldRef"`
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *ConfigMapKeySelector `json:"configMapKeyRef,omitempty" bson:"configMapKeyRef"`
	// Selects a key of a secret in the pod's namespace
	// +optional
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty" bson:"secretKeyRef"`
}

// ObjectFieldSelector selects an APIVersioned field of an object.
// +structType=atomic
type ObjectFieldSelector struct {
	// Version of the schema the FieldPath is written in terms of, defaults to "v1".
	// +optional
	APIVersion string `json:"apiVersion,omitempty" bson:"apiVersion"`
	// Path of the field to select in the specified API version.
	FieldPath string `json:"fieldPath" bson:"fieldPath"`
}

// ResourceFieldSelector represents container resources (cpu, memory) and their output format
// +structType=atomic
type ResourceFieldSelector struct {
	// Container name: required for volumes, optional for env vars
	// +optional
	ContainerName string `json:"containerName,omitempty" bson:"containerName"`
	// Required: resource to select
	Resource string `json:"resource" bson:"resource"`
	// Specifies the output format of the exposed resources, defaults to "1"
	// +optional
	Divisor Quantity `json:"divisor,omitempty" bson:"divisor"`
}

// ConfigMapKeySelector selects a key of a ConfigMap.
// +structType=atomic
type ConfigMapKeySelector struct {
	// The ConfigMap to select from.
	LocalObjectReference `json:",inline" bson:",inline"`
	// The key to select.
	Key string `json:"key" bson:"key"`
	// Specify whether the ConfigMap or its key must be defined
	// +optional
	Optional *bool `json:"optional,omitempty" bson:"optional"`
}

// LocalObjectReference contains enough information to let you locate the
// referenced object inside the same namespace.
// +structType=atomic
type LocalObjectReference struct {
	// Name of the referent.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// TODO: Add other useful fields. apiVersion, kind, uid?
	// +optional
	Name string `json:"name,omitempty" bson:"name"`
}

// SecretKeySelector selects a key of a Secret.
// +structType=atomic
type SecretKeySelector struct {
	// The name of the secret in the pod's namespace to select from.
	LocalObjectReference `json:",inline" bson:",inline"`
	// The key of the secret to select from.  Must be a valid secret key.
	Key string `json:"key" bson:"key"`
	// Specify whether the Secret or its key must be defined
	// +optional
	Optional *bool `json:"optional,omitempty" bson:"optional"`
}

// VolumeMount describes a mounting of a Volume within a container.
type VolumeMount struct {
	// This must match the Name of a Volume.
	Name string `json:"name" bson:"name"`
	// Mounted read-only if true, read-write otherwise (false or unspecified).
	// Defaults to false.
	// +optional
	ReadOnly bool `json:"readOnly,omitempty" bson:"readOnly"`
	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	MountPath string `json:"mountPath" bson:"mountPath"`
	// Path within the volume from which the container's volume should be mounted.
	// Defaults to "" (volume's root).
	// +optional
	SubPath string `json:"subPath,omitempty" bson:"subPath"`
	// mountPropagation determines how mounts are propagated from the host
	// to container and the other way around.
	// When not set, MountPropagationNone is used.
	// This field is beta in 1.10.
	// +optional
	// NOCC:tosa/linelength(忽略长度)
	MountPropagation *MountPropagationMode `json:"mountPropagation,omitempty" bson:"mountPropagation"`
	// Expanded path within the volume from which the container's volume should be mounted.
	// Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the
	// container's environment. Defaults to "" (volume's root). SubPathExpr and SubPath are mutually exclusive.
	// +optional
	SubPathExpr string `json:"subPathExpr,omitempty" bson:"subPathExpr"`
}

// MountPropagationMode describes mount propagation.
// +enum
type MountPropagationMode string
