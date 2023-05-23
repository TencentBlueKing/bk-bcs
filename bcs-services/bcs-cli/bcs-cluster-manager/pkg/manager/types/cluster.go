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

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// CreateClusterReq 创建集群request
type CreateClusterReq struct {
	ProjectID            string               `json:"projectID"`
	BusinessID           string               `json:"businessID"`
	EngineType           string               `json:"engineType"`
	IsExclusive          bool                 `json:"isExclusive"`
	ClusterType          string               `json:"clusterType"`
	Creator              string               `json:"creator"`
	ManageType           string               `json:"manageType"`
	ClusterName          string               `json:"clusterName"`
	Environment          string               `json:"environment"`
	Provider             string               `json:"provider"`
	Description          string               `json:"description"`
	ClusterBasicSettings ClusterBasicSettings `json:"clusterBasicSettings"`
	NetworkType          string               `json:"networkType"`
	Region               string               `json:"region"`
	VpcID                string               `json:"vpcID"`
	NetworkSettings      NetworkSettings      `json:"networkSettings"`
	Master               []string             `json:"master"`
}

// CreateVirtualClusterReq 创建虚拟集群request
type CreateVirtualClusterReq struct {
	ClusterID          string            `json:"clusterID"`
	ClusterName        string            `json:"clusterName"`
	Provider           string            `json:"provider"`
	Region             string            `json:"region"`
	VpcID              string            `json:"vpcID"`
	ProjectID          string            `json:"projectID"`
	BusinessID         string            `json:"businessID"`
	Environment        string            `json:"environment"`
	EngineType         string            `json:"engineType"`
	IsExclusive        bool              `json:"isExclusive"`
	ClusterType        string            `json:"clusterType"`
	HostClusterID      string            `json:"hostClusterID"`
	HostClusterNetwork string            `json:"hostClusterNetwork"`
	Labels             map[string]string `json:"labels"`
	Creator            string            `json:"creator"`
	// 集群创建的控制信息
	OnlyCreateInfo         bool                   `json:"onlyCreateInfo"`
	Master                 []string               `json:"master"`
	NetworkSettings        *NetworkSettings       `json:"networkSettings"`
	ClusterBasicSettings   *ClusterBasicSettings  `json:"clusterBasicSettings"`
	ClusterAdvanceSettings *ClusterAdvanceSetting `json:"clusterAdvanceSettings"`
	NodeSettings           *NodeSetting           `json:"nodeSettings"`
	ExtraInfo              map[string]string      `json:"extraInfo"`
	Description            string                 `json:"description"`
	Ns                     *NamespaceInfo         `json:"ns"`
}

// ClusterAdvanceSetting 集群高级设置
type ClusterAdvanceSetting struct {
	IPVS             bool              `json:"IPVS"`
	ContainerRuntime string            `json:"containerRuntime"`
	RuntimeVersion   string            `json:"runtimeVersion"`
	ExtraArgs        map[string]string `json:"extraArgs"`
}

// NodeSetting 节点设置
type NodeSetting struct {
	DockerGraphPath string            `json:"dockerGraphPath"`
	MountTarget     string            `json:"mountTarget"`
	UnSchedulable   uint32            `json:"unSchedulable"`
	Labels          map[string]string `json:"labels"`
	ExtraArgs       map[string]string `json:"extraArgs"`
}

// NamespaceInfo 命名空间信息
type NamespaceInfo struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" `
	Annotations map[string]string `json:"annotations,omitempty"`
}

// CreateVirtualClusterResp 创建虚拟集群response
type CreateVirtualClusterResp struct {
	Code    uint32                  `json:"code,omitempty"`
	Message string                  `json:"message,omitempty"`
	Result  bool                    `json:"result,omitempty"`
	Data    *clustermanager.Cluster `json:"data,omitempty"`
	Task    *clustermanager.Task    `json:"task,omitempty"`
}

// DeleteVirtualClusterReq 删除虚拟集群request
type DeleteVirtualClusterReq struct {
	ClusterID      string `json:"clusterID,omitempty"`
	OnlyDeleteInfo bool   `json:"onlyDeleteInfo,omitempty"`
	Operator       string `json:"operator,omitempty"`
}

// DeleteVirtualClusterResp 删除虚拟集群response
type DeleteVirtualClusterResp struct {
	Code    uint32                  `json:"code,omitempty"`
	Message string                  `json:"message,omitempty"`
	Result  bool                    `json:"result,omitempty"`
	Data    *clustermanager.Cluster `json:"data,omitempty"`
	Task    *clustermanager.Task    `json:"task,omitempty"`
}

// CreateClusterResp 创建集群response
type CreateClusterResp struct {
	ClusterID string `json:"clusterID"`
	TaskID    string `json:"taskID"`
}

// DeleteClusterReq 删除集群request
type DeleteClusterReq struct {
	ClusterID string `json:"clusterID"`
}

// DeleteClusterResp 删除集群response
type DeleteClusterResp struct {
	ClusterID string `json:"clusterID"`
	TaskID    string `json:"taskID"`
}

// UpdateClusterReq 更新集群request
type UpdateClusterReq struct {
	ClusterID            string               `json:"clusterID"`
	ProjectID            string               `json:"projectID"`
	BusinessID           string               `json:"businessID"`
	EngineType           string               `json:"engineType"`
	IsExclusive          bool                 `json:"isExclusive"`
	ClusterType          string               `json:"clusterType"`
	Updater              string               `json:"updater"`
	ManageType           string               `json:"manageType"`
	ClusterName          string               `json:"clusterName"`
	Environment          string               `json:"environment"`
	Provider             string               `json:"provider"`
	Description          string               `json:"description"`
	ClusterBasicSettings ClusterBasicSettings `json:"clusterBasicSettings"`
	NetworkType          string               `json:"networkType"`
	Region               string               `json:"region"`
	VpcID                string               `json:"vpcID"`
	NetworkSettings      NetworkSettings      `json:"networkSettings"`
	Master               []string             `json:"master"`
}

// ListClusterReq 查询集群列表request
type ListClusterReq struct {
	Offset uint32 `json:"offset"`
	Limit  uint32 `json:"limit"`
}

// GetClusterReq 查询集群request
type GetClusterReq struct {
	ClusterID string `json:"clusterID"`
}

// RetryCreateClusterReq 重试创建集群request
type RetryCreateClusterReq struct {
	ClusterID string `json:"clusterID"`
}

// AddNodesClusterReq 添加集群节点request
type AddNodesClusterReq struct {
	ClusterID    string   `json:"clusterID"`
	Nodes        []string `json:"nodes"`
	InitPassword string   `json:"initPassword"`
}

// DeleteNodesClusterReq 删除集群节点request
type DeleteNodesClusterReq struct {
	ClusterID string   `json:"clusterID"`
	Nodes     []string `json:"nodes"`
}

// CheckCloudKubeConfigReq kubeConfig连接集群可用性检测request
type CheckCloudKubeConfigReq struct {
	Kubeconfig string `json:"kubeconfig"`
}

// ImportClusterReq 导入集群request
type ImportClusterReq struct {
	ClusterName string `json:"clusterName"`
	Provider    string `json:"provider"`
	ProjectID   string `json:"projectID"`
	BusinessID  string `json:"businessID"`
	Environment string `json:"environment"`
	EngineType  string `json:"engineType"`
	IsExclusive bool   `json:"isExclusive"`
	ClusterType string `json:"clusterType"`
}

// ListClusterNodesReq 查询集群节点列表request
type ListClusterNodesReq struct {
	ClusterID string `json:"clusterID"`
	Offset    uint32 `json:"offset"`
	Limit     uint32 `json:"limit"`
}

// GetClusterResp 查询集群response
type GetClusterResp struct {
	Data Cluster `json:"data"`
}

// ListClusterResp 查询集群列表response
type ListClusterResp struct {
	Data []*Cluster `json:"data"`
}

// RetryCreateClusterResp 重试创建集群response
type RetryCreateClusterResp struct {
	ClusterID string `json:"clusterID"`
	TaskID    string `json:"taskID"`
}

// AddNodesClusterResp 添加集群节点response
type AddNodesClusterResp struct {
	TaskID string `json:"taskID"`
}

// DeleteNodesClusterResp 删除集群节点response
type DeleteNodesClusterResp struct {
	TaskID string `json:"taskID"`
}

// ListClusterNodesResp 查询集群节点列表response
type ListClusterNodesResp struct {
	Data []*ClusterNode `json:"data"`
}

// ListCommonClusterResp 查询公共集群response
type ListCommonClusterResp struct {
	Data []*Cluster `json:"data"`
}

// Cluster 集群信息
type Cluster struct {
	ClusterID            string               `json:"clusterID"`
	ProjectID            string               `json:"projectID"`
	BusinessID           string               `json:"businessID"`
	EngineType           string               `json:"engineType"`
	IsExclusive          bool                 `json:"isExclusive"`
	ClusterType          string               `json:"clusterType"`
	Creator              string               `json:"creator"`
	Updater              string               `json:"updater"`
	ManageType           string               `json:"manageType"`
	ClusterName          string               `json:"clusterName"`
	Environment          string               `json:"environment"`
	Provider             string               `json:"provider"`
	Description          string               `json:"description"`
	ClusterBasicSettings ClusterBasicSettings `json:"clusterBasicSettings"`
	NetworkType          string               `json:"networkType"`
	Region               string               `json:"region"`
	VpcID                string               `json:"vpcID"`
	NetworkSettings      NetworkSettings      `json:"networkSettings"`
	Master               []string             `json:"master"`
}

// ClusterBasicSettings 集群基础设置
type ClusterBasicSettings struct {
	Version string `json:"version"`
}

// NetworkSettings 网络设置
type NetworkSettings struct {
	CidrStep      uint32 `json:"cidrStep"`
	MaxNodePodNum uint32 `json:"maxNodePodNum"`
	MaxServiceNum uint32 `json:"maxServiceNum"`
}

// ImportCloudMode 导入云模式
type ImportCloudMode struct {
	CloudID    string `json:"cloudID"`
	KubeConfig string `json:"kubeConfig"`
}

// ClusterNode 集群节点信息
type ClusterNode struct {
	NodeID  string `json:"nodeID"`
	InnerIP string `json:"innerIP"`
}

// ClusterMgr 集群管理接口
type ClusterMgr interface {
	// Create 创建集群
	Create(CreateClusterReq) (CreateClusterResp, error)
	// Delete 删除集群
	Delete(DeleteClusterReq) (DeleteClusterResp, error)
	// Update 更新集群
	Update(UpdateClusterReq) error
	// Get 获取集群
	Get(GetClusterReq) (GetClusterResp, error)
	// List 获取集群列表
	List(ListClusterReq) (ListClusterResp, error)
	// Retry 重试创建集群
	RetryCreate(RetryCreateClusterReq) (RetryCreateClusterResp, error)
	// AddNodes 添加节点到集群
	AddNodes(AddNodesClusterReq) (AddNodesClusterResp, error)
	// DeleteNodes 从集群中删除节点
	DeleteNodes(DeleteNodesClusterReq) (DeleteNodesClusterResp, error)
	// CheckCloudKubeConfig kubeConfig连接集群可用性检测
	CheckCloudKubeConfig(CheckCloudKubeConfigReq) error
	// Import 导入用户集群(支持多云集群导入功能: 集群ID/kubeConfig)
	Import(ImportClusterReq) error
	// ListNodes 查询集群下所有节点列表
	ListNodes(ListClusterNodesReq) (ListClusterNodesResp, error)
	// ListCommon 查询公共集群及公共集群所属权限
	ListCommon() (ListCommonClusterResp, error)
}
