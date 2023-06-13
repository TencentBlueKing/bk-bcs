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

// CreateNodeGroupReq 创建节点池request
type CreateNodeGroupReq struct {
	Name            string              `json:"name"`
	AutoScaling     AutoScaling         `json:"autoScaling"`
	EnableAutoscale bool                `json:"enableAutoscale"`
	NodeTemplate    NodeTemplate        `json:"nodeTemplate"`
	LaunchTemplate  LaunchConfiguration `json:"launchTemplate"`
	ClusterID       string              `json:"clusterID"`
	Region          string              `json:"region"`
	Labels          map[string]string   `json:"labels"`
	Taints          map[string]string   `json:"taints"`
	Tags            map[string]string   `json:"tags"`
	NodeOS          string              `json:"nodeOS"`
	Creator         string              `json:"creator"`
}

// UpdateNodeGroupReq 更新节点池request
type UpdateNodeGroupReq struct {
	NodeGroupID     string              `json:"nodeGroupID"`
	Name            string              `json:"name"`
	AutoScaling     AutoScaling         `json:"autoScaling"`
	EnableAutoscale bool                `json:"enableAutoscale"`
	NodeTemplate    NodeTemplate        `json:"nodeTemplate"`
	LaunchTemplate  LaunchConfiguration `json:"launchTemplate"`
	ClusterID       string              `json:"clusterID"`
	Region          string              `json:"region"`
	Labels          map[string]string   `json:"labels"`
	Taints          map[string]string   `json:"taints"`
	Tags            map[string]string   `json:"tags"`
	NodeOS          string              `json:"nodeOS"`
	Updater         string              `json:"updater"`
}

// DeleteNodeGroupReq 删除节点池request
type DeleteNodeGroupReq struct {
	NodeGroupID string `json:"nodeGroupID"`
}

// GetNodeGroupReq 查询节点池信息request
type GetNodeGroupReq struct {
	NodeGroupID string `json:"nodeGroupID"`
}

// ListNodeGroupReq 查询节点池列表request
type ListNodeGroupReq struct {
}

// MoveNodesToGroupReq 移动节点到节点池request
type MoveNodesToGroupReq struct {
	ClusterID   string   `json:"clusterID"`
	NodeGroupID string   `json:"nodeGroupID"`
	Nodes       []string `json:"nodes"`
}

// RemoveNodesFromGroupReq 从节点池移除节点request
type RemoveNodesFromGroupReq struct {
	ClusterID   string   `json:"clusterID"`
	NodeGroupID string   `json:"nodeGroupID"`
	Nodes       []string `json:"nodes"`
}

// CleanNodeGroupReq 从节点池移除节点并清理资源request
type CleanNodeGroupReq struct {
	ClusterID   string   `json:"clusterID"`
	NodeGroupID string   `json:"nodeGroupID"`
	Nodes       []string `json:"nodes"`
}

// CleanNodeGroupV2Req 从节点池移除节点并清理资源request
type CleanNodeGroupV2Req struct {
	ClusterID   string   `json:"clusterID"`
	NodeGroupID string   `json:"nodeGroupID"`
	Nodes       []string `json:"nodes"`
}

// ListNodesInGroupReq 查询节点池的节点列表request
type ListNodesInGroupReq struct {
	NodeGroupID string `json:"nodeGroupID"`
}

// UpdateGroupDesiredNodeReq 更新节点池DesiredNode request
type UpdateGroupDesiredNodeReq struct {
	NodeGroupID string `json:"nodeGroupID"`
	DesiredNode uint32 `json:"desiredNode"`
}

// UpdateGroupDesiredSizeReq 更新节点池DesiredSize request
type UpdateGroupDesiredSizeReq struct {
	NodeGroupID string `json:"nodeGroupID"`
	DesiredSize uint32 `json:"desiredSize"`
}

// EnableNodeGroupAutoScaleReq 开启节点池自动伸缩功能
type EnableNodeGroupAutoScaleReq struct {
	NodeGroupID string `json:"nodeGroupID"`
}

// DisableNodeGroupAutoScaleReq 关闭节点池自动伸缩功能
type DisableNodeGroupAutoScaleReq struct {
	NodeGroupID string `json:"nodeGroupID"`
}

// CreateNodeGroupResp 创建节点池response
type CreateNodeGroupResp struct {
	NodeGroupID string `json:"nodeGroupID"`
	TaskID      string `json:"taskID"`
}

// ListNodeGroupResp 查询节点池列表response
type ListNodeGroupResp struct {
	Data []NodeGroup `json:"data"`
}

// DeleteNodeGroupResp 删除节点池response
type DeleteNodeGroupResp struct {
	TaskID string `json:"taskID"`
}

// GetNodeGroupResp 查询节点池信息response
type GetNodeGroupResp struct {
	Data NodeGroup `json:"data"`
}

// MoveNodesToGroupResp 移动节点到节点池response
type MoveNodesToGroupResp struct {
	TaskID string `json:"taskID"`
}

// RemoveNodesFromGroupResp 移除节点池的节点response
type RemoveNodesFromGroupResp struct {
	TaskID string `json:"taskID"`
}

// CleanNodeGroupResp 从节点池移除节点并清理资源response
type CleanNodeGroupResp struct {
	TaskID string `json:"taskID"`
}

// CleanNodeGroupV2Resp 从节点池移除节点并清理资源response
type CleanNodeGroupV2Resp struct {
	TaskID string `json:"taskID"`
}

// ListNodesInGroupResp 查询节点池的节点列表response
type ListNodesInGroupResp struct {
	Data []NodeGroupNode `json:"data"`
}

// UpdateGroupDesiredNodeResp 更新节点池DesiredNode response
type UpdateGroupDesiredNodeResp struct {
	TaskID string `json:"taskID"`
}

// NodeGroup 节点池信息
type NodeGroup struct {
	NodeGroupID      string              `json:"nodeGroupID"`
	Name             string              `json:"name"`
	ClusterID        string              `json:"clusterID"`
	Region           string              `json:"region"`
	EnableAutoscale  bool                `json:"enableAutoscale"`
	AutoScaling      AutoScaling         `json:"autoScaling"`
	LaunchTemplate   LaunchConfiguration `json:"launchTemplate"`
	Labels           map[string]string   `json:"labels,omitempty"`
	Taints           map[string]string   `json:"taints,omitempty"`
	NodeOS           string              `json:"nodeOS"`
	Creator          string              `json:"creator"`
	Updater          string              `json:"updater"`
	CreateTime       string              `json:"createTime"`
	UpdateTime       string              `json:"updateTime"`
	ProjectID        string              `json:"projectID"`
	Provider         string              `json:"provider"`
	Status           string              `json:"status"`
	ConsumerID       string              `json:"consumerID,omitempty"`
	NodeTemplate     *NodeTemplate       `json:"nodeTemplate,omitempty"`
	CloudNodeGroupID string              `json:"cloudNodeGroupID,omitempty"`
	Tags             map[string]string   `json:"tags,omitempty"`
	BkCloudID        uint32              `json:"bkCloudID,omitempty"`
	BkCloudName      string              `json:"bkCloudName,omitempty"`
}

// AutoScaling 自动伸缩信息
type AutoScaling struct {
	AutoScalingID         string      `json:"autoScalingID,omitempty"`
	AutoScalingName       string      `json:"autoScalingName,omitempty"`
	MinSize               uint32      `json:"minSize"`
	MaxSize               uint32      `json:"maxSize"`
	DesiredSize           uint32      `json:"desiredSize,omitempty"`
	VpcID                 string      `json:"vpcID,omitempty"`
	DefaultCooldown       uint32      `json:"defaultCooldown,omitempty"`
	SubnetIDs             []string    `json:"subnetIDs,omitempty"`
	Zones                 []string    `json:"zones,omitempty"`
	RetryPolicy           string      `json:"retryPolicy,omitempty"`
	MultiZoneSubnetPolicy string      `json:"multiZoneSubnetPolicy,omitempty"`
	ReplaceUnhealthy      bool        `json:"replaceUnhealthy,omitempty"`
	ScalingMode           string      `json:"scalingMode,omitempty"`
	TimeRanges            []TimeRange `json:"timeRanges,omitempty"`
}

// TimeRange 时间范围信息
type TimeRange struct {
	Name       string `json:"name"`
	Schedule   string `json:"schedule"`
	Zone       string `json:"zone"`
	DesiredNum uint32 `json:"desiredNum"`
}

// LaunchConfiguration 节点Launch配置模板
type LaunchConfiguration struct {
	LaunchConfigurationID string              `json:"launchConfigurationID,omitempty"`
	LaunchConfigureName   string              `json:"launchConfigureName,omitempty"`
	ProjectID             string              `json:"projectID,omitempty"`
	CPU                   uint32              `json:"CPU,omitempty"`
	Mem                   uint32              `json:"Mem,omitempty"`
	GPU                   uint32              `json:"GPU,omitempty"`
	InstanceType          string              `json:"instanceType"`
	InstanceChargeType    string              `json:"instanceChargeType,omitempty"`
	SystemDisk            *DataDisk           `json:"systemDisk,omitempty"`
	DataDisks             []DataDisk          `json:"dataDisks,omitempty"`
	InternetAccess        *InternetAccessible `json:"internetAccess,omitempty"`
	InitLoginPassword     string              `json:"initLoginPassword,omitempty"`
	SecurityGroupIDs      []string            `json:"securityGroupIDs,omitempty"`
	ImageInfo             ImageInfo           `json:"imageInfo"`
	IsSecurityService     bool                `json:"isSecurityService,omitempty"`
	IsMonitorService      bool                `json:"isMonitorService,omitempty"`
	UserData              string              `json:"userData,omitempty"`
}

// DataDisk 数据盘定义
type DataDisk struct {
	DiskType           string `json:"diskType"`
	DiskSize           string `json:"diskSize"`
	FileSystem         string `json:"fileSystem"`
	AutoFormatAndMount bool   `json:"autoFormatAndMount"`
	MountTarget        string `json:"mountTarget"`
}

// InternetAccessible 公网带宽设置
type InternetAccessible struct {
	InternetChargeType   string `json:"internetChargeType"`
	InternetMaxBandwidth string `json:"internetMaxBandwidth"`
	PublicIPAssigned     bool   `json:"publicIPAssigned"`
}

// ImageInfo cvm实例的镜像信息
type ImageInfo struct {
	ImageID   string `json:"imageID"`
	ImageName string `json:"imageName"`
}

// NodeTemplate 节点模版信息
type NodeTemplate struct {
	NodeTemplateID      string            `json:"nodeTemplateID"`
	Name                string            `json:"name"`
	ProjectID           string            `json:"projectID"`
	Labels              map[string]string `json:"labels"`
	Taints              []Taint           `json:"taints"`
	DockerGraphPath     string            `json:"dockerGraphPath"`
	MountTarget         string            `json:"mountTarget"`
	UserScript          string            `json:"userScript"`
	UnSchedulable       uint32            `json:"unSchedulable"`
	DataDisks           []DataDisk        `json:"dataDisks"`
	ExtraArgs           map[string]string `json:"extraArgs"`
	PreStartUserScript  string            `json:"preStartUserScript"`
	BcsScaleOutAddons   Action            `json:"bcsScaleOutAddons"`
	BcsScaleInAddons    Action            `json:"bcsScaleInAddons"`
	ScaleOutExtraAddons Action            `json:"scaleOutExtraAddons"`
	ScaleInExtraAddons  Action            `json:"scaleInExtraAddons"`
	NodeOS              string            `json:"nodeOS"`
	ModuleID            string            `json:"moduleID"`
	Creator             string            `json:"creator"`
	Updater             string            `json:"updater"`
	CreateTime          string            `json:"createTime"`
	UpdateTime          string            `json:"updateTime"`
	Desc                string            `json:"desc"`
	Runtime             RunTimeInfo       `json:"runtime"`
	Module              ModuleInfo        `json:"module"`
}

// Action 自动化行为模板
type Action struct {
	PreActions  []string               `json:"preActions"`
	PostActions []string               `json:"postActions"`
	Plugins     map[string]BKOpsPlugin `json:"plugins"`
}

// BKOpsPlugin 标准运维模板信息记录
type BKOpsPlugin struct {
	System string            `json:"system"`
	Link   string            `json:"link"`
	Params map[string]string `json:"params"`
}

// Taint 污点
type Taint struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

// RunTimeInfo 容器运行时
type RunTimeInfo struct {
	ContainerRuntime string `json:"containerRuntime"`
	RuntimeVersion   string `json:"runtimeVersion"`
}

// ModuleInfo 业务模块信息
type ModuleInfo struct {
	ScaleOutModuleID   string `json:"scaleOutModuleID"`
	ScaleInModuleID    string `json:"scaleInModuleID"`
	ScaleOutModuleName string `json:"scaleOutModuleName"`
	ScaleInModuleName  string `json:"scaleInModuleName"`
}

// NodeGroupNode 节点池节点信息
type NodeGroupNode struct {
	NodeID        string `json:"nodeID"`
	InnerIP       string `json:"innerIP"`
	InstanceType  string `json:"instanceType"`
	CPU           uint32 `json:"CPU"`
	Mem           uint32 `json:"mem"`
	GPU           uint32 `json:"GPU"`
	Status        string `json:"status"`
	ZoneID        string `json:"zoneID"`
	NodeGroupID   string `json:"nodeGroupID"`
	ClusterID     string `json:"clusterID"`
	VPC           string `json:"VPC"`
	Region        string `json:"region"`
	Passwd        string `json:"passwd"`
	Zone          uint32 `json:"zone"`
	DeviceID      string `json:"deviceID"`
	InstanceRole  string `json:"instanceRole"`
	UnSchedulable uint32 `json:"unSchedulable"`
}

// NodeGroupMgr 节点池管理接口
type NodeGroupMgr interface {
	// Create 创建节点池
	Create(CreateNodeGroupReq) (CreateNodeGroupResp, error)
	// Update 更新节点池
	Update(UpdateNodeGroupReq) error
	// Delete 删除节点池
	Delete(DeleteNodeGroupReq) (DeleteNodeGroupResp, error)
	// Get 查询节点池信息
	Get(GetNodeGroupReq) (GetNodeGroupResp, error)
	// List 查询节点池列表
	List(ListNodeGroupReq) (ListNodeGroupResp, error)
	// MoveNodes 移动节点到节点池
	MoveNodes(MoveNodesToGroupReq) (MoveNodesToGroupResp, error)
	// RemoveNodes 从节点池移除节点
	RemoveNodes(RemoveNodesFromGroupReq) (RemoveNodesFromGroupResp, error)
	// CleanNodes 从节点池移除节点并清理资源回收节点
	CleanNodes(CleanNodeGroupReq) (CleanNodeGroupResp, error)
	// CleanNodesV2 从节点池移除节点并清理资源回收节点
	CleanNodesV2(CleanNodeGroupV2Req) (CleanNodeGroupV2Resp, error)
	// ListNodes 查询节点池的节点列表
	ListNodes(ListNodesInGroupReq) (ListNodesInGroupResp, error)
	// UpdateDesiredNode 更新节点池DesiredNode信息
	UpdateDesiredNode(UpdateGroupDesiredNodeReq) (UpdateGroupDesiredNodeResp, error)
	// UpdateDesiredSize 更新节点池DesiredSize信息
	UpdateDesiredSize(UpdateGroupDesiredSizeReq) error
	// EnableAutoScale 开启节点池自动伸缩功能
	EnableAutoScale(EnableNodeGroupAutoScaleReq) error
	// DisableAutoScale 关闭节点池自动伸缩功能
	DisableAutoScale(DisableNodeGroupAutoScaleReq) error
}
