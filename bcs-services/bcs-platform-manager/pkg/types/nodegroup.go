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

// Package types pod types
package types

// ListNodeGroupReq list nodegroup request
type ListNodeGroupReq struct {
	Name      string `json:"name" in:"query=name"`
	ClusterID string `json:"clusterID" in:"query=clusterID"`
	Region    string `json:"region" in:"query=region"`
	ProjectID string `json:"projectID" in:"query=projectID"`
	Limit     uint32 `json:"limit" in:"query=limit"`
	Page      uint32 `json:"page" in:"query=page"`
}

// ListNodeGroupResp list nodegroup response
type ListNodeGroupResp struct {
	Total   uint32               `json:"total"`
	Results []*ListNodeGroupData `json:"results"`
}

// ListNodeGroupData nodegroup data
type ListNodeGroupData struct {
	NodeGroupID    string               `json:"nodeGroupID"`
	Name           string               `json:"name"`
	ClusterID      string               `json:"clusterID"`
	AutoScaling    *AutoScalingGroup    `json:"autoScaling"`
	LaunchTemplate *LaunchConfiguration `json:"launchTemplate"`
	Status         string               `json:"status"`
}

// NodeGroup pool for kubernetes cluster-autoscaling
type NodeGroup struct {
	NodeGroupID      string               `json:"nodeGroupID"`
	Name             string               `json:"name"`
	ClusterID        string               `json:"clusterID"`
	Region           string               `json:"region"`
	EnableAutoscale  bool                 `json:"enableAutoscale"`
	AutoScaling      *AutoScalingGroup    `json:"autoScaling"`
	LaunchTemplate   *LaunchConfiguration `json:"launchTemplate"`
	Labels           map[string]string    `json:"labels"`
	Taints           map[string]string    `json:"taints"`
	NodeOS           string               `json:"nodeOS"`
	Creator          string               `json:"creator"`
	Updater          string               `json:"updater"`
	CreateTime       string               `json:"createTime"`
	UpdateTime       string               `json:"updateTime"`
	ProjectID        string               `json:"projectID"`
	Provider         string               `json:"provider"`
	Status           string               `json:"status"`
	ConsumerID       string               `json:"consumerID"`
	NodeTemplate     *NodeTemplate        `json:"nodeTemplate"`
	CloudNodeGroupID string               `json:"cloudNodeGroupID"`
	Tags             map[string]string    `json:"tags"`
	NodeGroupType    string               `json:"nodeGroupType"`
	Area             *CloudArea           `json:"area"`
	ExtraInfo        map[string]string    `json:"extraInfo"`
}

// AutoScalingGroup define auto scaling information
type AutoScalingGroup struct {
	AutoScalingID         string       `json:"autoScalingID"`
	AutoScalingName       string       `json:"autoScalingName"`
	MinSize               uint32       `json:"minSize"`
	MaxSize               uint32       `json:"maxSize"`
	DesiredSize           uint32       `json:"desiredSize"`
	VpcID                 string       `json:"vpcID"`
	DefaultCooldown       uint32       `json:"defaultCooldown"`
	SubnetIDs             []string     `json:"subnetIDs"`
	Zones                 []string     `json:"zones"`
	RetryPolicy           string       `json:"retryPolicy"`
	MultiZoneSubnetPolicy string       `json:"multiZoneSubnetPolicy"`
	ReplaceUnhealthy      bool         `json:"replaceUnhealthy"`
	ScalingMode           string       `json:"scalingMode"`
	TimeRanges            []*TimeRange `json:"timeRanges"`
	AutoUpgrade           bool         `json:"autoUpgrade"`
	ServiceRole           string       `json:"serviceRole"`
}

// TimeRange define time range for auto scaling
type TimeRange struct {
	Name       string `json:"name"`
	Schedule   string `json:"schedule"`
	Zone       string `json:"zone"`
	DesiredNum uint32 `json:"desiredNum"`
}

// LaunchConfiguration template for scaling node
type LaunchConfiguration struct {
	LaunchConfigurationID string                 `json:"launchConfigurationID"`
	LaunchConfigureName   string                 `json:"launchConfigureName"`
	ProjectID             string                 `json:"projectID"`
	CPU                   uint32                 `json:"CPU"`
	Mem                   uint32                 `json:"Mem"`
	GPU                   uint32                 `json:"GPU"`
	InstanceType          string                 `json:"instanceType"`
	InstanceChargeType    string                 `json:"instanceChargeType"`
	SystemDisk            *DataDisk              `json:"systemDisk"`
	DataDisks             []*DataDisk            `json:"dataDisks"`
	InternetAccess        *InternetAccessible    `json:"internetAccess"`
	InitLoginPassword     string                 `json:"initLoginPassword"`
	SecurityGroupIDs      []string               `json:"securityGroupIDs"`
	ImageInfo             *ImageInfo             `json:"imageInfo"`
	IsSecurityService     bool                   `json:"isSecurityService"`
	IsMonitorService      bool                   `json:"isMonitorService"`
	UserData              string                 `json:"userData"`
	InitLoginUsername     string                 `json:"initLoginUsername"`
	Selector              map[string]string      `json:"selector"`
	KeyPair               *KeyInfo               `json:"keyPair"`
	Charge                *InstanceChargePrepaid `json:"charge"`
	NetworkTag            []string               `json:"networkTag"`
}

// DataDisk 数据盘定义
type DataDisk struct {
	DiskType string `json:"diskType"`
	DiskSize string `json:"diskSize"`
}

// InternetAccessible 公网带宽设置
type InternetAccessible struct {
	InternetChargeType   string   `json:"internetChargeType"`
	InternetMaxBandwidth string   `json:"internetMaxBandwidth"`
	PublicIPAssigned     bool     `json:"publicIPAssigned"`
	BandwidthPackageId   string   `json:"bandwidthPackageId"`
	PublicIP             string   `json:"publicIP"`
	PublicAccessCidrs    []string `json:"publicAccessCidrs"`
	NodePublicIPPrefixID string   `json:"nodePublicIPPrefixID"`
}

// ImageInfo 创建cvm实例的镜像信息
type ImageInfo struct {
	ImageID   string `json:"imageID"`
	ImageName string `json:"imageName"`
	ImageType string `json:"imageType"`
	ImageOs   string `json:"imageOs"`
}

// KeyInfo key pair information
type KeyInfo struct {
	KeyID     string `json:"keyID"`
	KeySecret string `json:"keySecret"`
	KeyPublic string `json:"keyPublic"`
}

// InstanceChargePrepaid instance charge prepaid
type InstanceChargePrepaid struct {
	Period    uint32 `json:"period"`
	RenewFlag string `json:"renewFlag"`
}

// NodeTemplate for kubernetes cluster node common setting
type NodeTemplate struct {
	NodeTemplateID              string            `json:"nodeTemplateID"`
	Name                        string            `json:"name"`
	ProjectID                   string            `json:"projectID"`
	Labels                      map[string]string `json:"labels"`
	Taints                      []*Taint          `json:"taints"`
	DockerGraphPath             string            `json:"dockerGraphPath"`
	MountTarget                 string            `json:"mountTarget"`
	UserScript                  string            `json:"userScript"`
	UnSchedulable               uint32            `json:"unSchedulable"`
	DataDisks                   []*CloudDataDisk  `json:"dataDisks"`
	ExtraArgs                   map[string]string `json:"extraArgs"`
	PreStartUserScript          string            `json:"preStartUserScript"`
	BcsScaleOutAddons           *Action           `json:"bcsScaleOutAddons"`
	BcsScaleInAddons            *Action           `json:"bcsScaleInAddons"`
	ScaleOutExtraAddons         *Action           `json:"scaleOutExtraAddons"`
	ScaleInExtraAddons          *Action           `json:"scaleInExtraAddons"`
	NodeOS                      string            `json:"nodeOS"`
	Creator                     string            `json:"creator"`
	Updater                     string            `json:"updater"`
	CreateTime                  string            `json:"createTime"`
	UpdateTime                  string            `json:"updateTime"`
	Desc                        string            `json:"desc"`
	Runtime                     *RunTimeInfo      `json:"runtime"`
	Module                      *ModuleInfo       `json:"module"`
	ScaleInPreScript            string            `json:"scaleInPreScript"`
	ScaleInPostScript           string            `json:"scaleInPostScript"`
	Annotations                 map[string]string `json:"annotations"`
	MaxPodsPerNode              uint32            `json:"maxPodsPerNode"`
	SkipSystemInit              bool              `json:"skipSystemInit"`
	AllowSkipScaleOutWhenFailed bool              `json:"allowSkipScaleOutWhenFailed"`
	AllowSkipScaleInWhenFailed  bool              `json:"allowSkipScaleInWhenFailed"`
	Image                       *ImageInfo        `json:"image"`
	GpuArgs                     *GPUArgs          `json:"gpuArgs"`
	ExtraInfo                   map[string]string `json:"extraInfo"`
}

// Taint for node taints
type Taint struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

// CloudDataDisk 云磁盘格式化数据, 对应CVM数据盘。应用于节点模版
// 主要用于 CA 自动扩容节点并上架节点时重装系统，多块数据盘 mountTarget 不能重复
// 上架已存在节点时, 用户需指定diskPartition参数区分设备
type CloudDataDisk struct {
	DiskType           string `json:"diskType"`
	DiskSize           string `json:"diskSize"`
	FileSystem         string `json:"fileSystem"`
	AutoFormatAndMount bool   `json:"autoFormatAndMount"`
	MountTarget        string `json:"mountTarget"`
	DiskPartition      string `json:"diskPartition"`
}

// Action 节点操作信息
type Action struct {
	PreActions  []string                `json:"preActions"`
	PostActions []string                `json:"postActions"`
	Plugins     map[string]*BKOpsPlugin `json:"plugins"`
}

// BKOpsPlugin 插件信息
type BKOpsPlugin struct {
	System              string            `json:"system"`
	Link                string            `json:"link"`
	Params              map[string]string `json:"params"`
	AllowSkipWhenFailed bool              `json:"allowSkipWhenFailed"`
}

// RunTimeInfo cluster runTime info
type RunTimeInfo struct {
	ContainerRuntime string `json:"containerRuntime"`
	RuntimeVersion   string `json:"runtimeVersion"`
}

// ModuleInfo 业务模块信息,主要涉及到节点模块转移
type ModuleInfo struct {
	ScaleOutModuleID   string `json:"scaleOutModuleID"`
	ScaleInModuleID    string `json:"scaleInModuleID"`
	ScaleOutBizID      string `json:"scaleOutBizID"`
	ScaleInBizID       string `json:"scaleInBizID"`
	ScaleOutModuleName string `json:"scaleOutModuleName"`
	ScaleInModuleName  string `json:"scaleInModuleName"`
}

// GPUArgs GPU参数
type GPUArgs struct {
	MigEnable    bool           `json:"migEnable"`
	Driver       *DriverVersion `json:"driver"`
	Cuda         *DriverVersion `json:"cuda"`
	Cudnn        *CUDNN         `json:"cudnn"`
	CustomDriver *CustomDriver  `json:"customDriver"`
}

// DriverVersion driver version info
type DriverVersion struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}

// CUDNN cudnn driver info
type CUDNN struct {
	Version string `json:"version"`
	Name    string `json:"name"`
	DocName string `json:"docName"`
	DevName string `json:"devName"`
}

// CustomDriver 自定义驱动
type CustomDriver struct {
	Address string `json:"address"`
}

// CloudArea 云区域信息
type CloudArea struct {
	BkCloudID   uint32 `json:"bkCloudID"`
	BkCloudName string `json:"bkCloudName"`
}

// GetNodeGroupReq get nodegroup request
type GetNodeGroupReq struct {
	NodeGroupID string `json:"nodeGroupID" in:"path=nodeGroupID"`
}

// EnableNodeGroupAutoScaleReq enable nodegroup auto scale request
type EnableNodeGroupAutoScaleReq struct {
	NodeGroupID string `json:"nodeGroupID" in:"path=nodeGroupID"`
}

// DisableNodeGroupAutoScaleReq disable nodegroup auto scale request
type DisableNodeGroupAutoScaleReq struct {
	NodeGroupID string `json:"nodeGroupID" in:"path=nodeGroupID"`
}

// UpdateNodeGroupReq update nodegroup request
type UpdateNodeGroupReq struct {
	NodeGroupID     string               `json:"nodeGroupID" in:"path=nodeGroupID"`
	ClusterID       string               `json:"clusterID"`
	Name            string               `json:"name"`
	Region          string               `json:"region"`
	EnableAutoscale *bool                `json:"enableAutoscale"`
	AutoScaling     *AutoScalingGroup    `json:"autoScaling"`
	LaunchTemplate  *LaunchConfiguration `json:"launchTemplate"`
	NodeTemplate    *NodeTemplate        `json:"nodeTemplate"`
	Labels          map[string]string    `json:"labels"`
	Taints          map[string]string    `json:"taints"`
	Tags            map[string]string    `json:"tags"`
	NodeOS          string               `json:"nodeOS"`
	Updater         string               `json:"updater"`
	Provider        string               `json:"provider"`
	ConsumerID      string               `json:"consumerID"`
	Desc            string               `json:"desc"`
	BkCloudID       *uint32              `json:"bkCloudID"`
	CloudAreaName   *string              `json:"cloudAreaName"`
	OnlyUpdateInfo  bool                 `json:"onlyUpdateInfo"`
	ExtraInfo       map[string]string    `json:"extraInfo"`
}

// UpdateGroupMinMaxSizeReq update nodegroup min max size request
type UpdateGroupMinMaxSizeReq struct {
	NodeGroupID string `json:"nodeGroupID" in:"path=nodeGroupID"`
	MinSize     uint32 `json:"minSize" in:"query=minSize"`
	MaxSize     uint32 `json:"maxSize" in:"query=maxSize"`
	Operator    string `json:"operator" in:"query=operator"`
}
