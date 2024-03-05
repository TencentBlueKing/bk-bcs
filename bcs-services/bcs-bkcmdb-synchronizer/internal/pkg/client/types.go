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
