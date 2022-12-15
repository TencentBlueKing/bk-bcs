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

package types

// GetNodeReq 查询节点信息request
type GetNodeReq struct {
	InnerIP string `json:"innerIP"`
}

// UpdateNodeReq 更新节点信息request
type UpdateNodeReq struct {
	InnerIPs    []string `json:"innerIPs"`
	Status      string   `json:"status"`
	NodeGroupID string   `json:"nodeGroupID"`
	ClusterID   string   `json:"clusterID"`
}

// CheckNodeInClusterReq 检查节点是否存在bcs集群中request
type CheckNodeInClusterReq struct {
	InnerIPs []string `json:"innerIPs"`
}

// CordonNodeReq 节点设置不可调度状态request
type CordonNodeReq struct {
	InnerIPs  []string `json:"innerIPs"`
	ClusterID string
}

// UnCordonNodeReq 节点设置可调度状态request
type UnCordonNodeReq struct {
	InnerIPs  []string `json:"innerIPs"`
	ClusterID string   `json:"clusterID"`
}

// DrainNodeReq 节点pod迁移request
type DrainNodeReq struct {
	InnerIPs  []string `json:"innerIPs"`
	ClusterID string   `json:"clusterID"`
}

// GetNodeResp 查询节点信息response
type GetNodeResp struct {
	Data []*Node `json:"data"`
}

// CheckNodeInClusterResp 检查节点是否存在bcs集群中response
type CheckNodeInClusterResp struct {
	Data map[string]NodeResult `json:"data"`
}

// CordonNodeResp 节点设置不可调度状态response
type CordonNodeResp struct {
	Data []string `json:"data"`
}

// UnCordonNodeResp 节点设置可调度状态response
type UnCordonNodeResp struct {
	Data []string `json:"data"`
}

// DrainNodeResp 节点pod迁移response
type DrainNodeResp struct {
	Data []string `json:"data"`
}

// UpdateNodeResp 更新节点response
type UpdateNodeResp struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
	Result  bool   `json:"result"`
}

// Node 节点信息
type Node struct {
	NodeID       string `json:"nodeID"`
	InnerIP      string `json:"innerIP"`
	InstanceType string `json:"InstanceType"`
	CPU          uint32 `json:"cpu"`
	Mem          uint32 `json:"mem"`
	GPU          uint32 `json:"gpu"`
	Status       string `json:"status"`
	ZoneID       string `json:"zoneID"`
	NodeGroupID  string `json:"nodeGroupID"`
	ClusterID    string `json:"clusterID"`
	VPC          string `json:"vpc"`
	Region       string `json:"region"`
	Passwd       string `json:"passwd"`
	Zone         uint32 `json:"zone"`
	DeviceID     string `json:"deviceID"`
}

// NodeResult 节点是否存在集群中
type NodeResult struct {
	IsExist     bool   `json:"isExist"`
	ClusterID   string `json:"clusterID"`
	ClusterName string `json:"clusterName"`
}

// NodeOperationStatus 节点操作状态
type NodeOperationStatus struct {
	Fail    []NodeOperationStatusInfo `json:"fail"`
	Success []NodeOperationStatusInfo `json:"success"`
}

// NodeOperationStatusInfo 节点操作状态信息
type NodeOperationStatusInfo struct {
	NodeName string `json:"nodeName"`
	Message  string `json:"message"`
}

// NodeMgr 节点管理接口
type NodeMgr interface {
	// Get 查询指定InnerIP的节点信息
	Get(GetNodeReq) (GetNodeResp, error)
	// Update 更新node信息
	Update(UpdateNodeReq) error
	// CheckNodeInCluster 检查节点是否存在bcs集群中
	CheckNodeInCluster(CheckNodeInClusterReq) (CheckNodeInClusterResp, error)
	// Cordon 节点设置不可调度状态
	Cordon(CordonNodeReq) (CordonNodeResp, error)
	// UnCordon 节点设置可调度状态
	UnCordon(UnCordonNodeReq) (UnCordonNodeResp, error)
	// Drain 节点pod迁移,将节点上的Pod驱逐
	Drain(DrainNodeReq) (DrainNodeResp, error)
}
