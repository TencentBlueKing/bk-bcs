/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

// CreateClusterRequest for create ECK cluster
type CreateClusterRequest struct {
	// 自定义证书SAN，多个IP或域名以英文逗号（,）分隔
	CertExtraSans string `json:"certExtraSans,omitempty"`
	// 集群创建过程需要安装的组件，可选值通过组件管理平台支持安装的组件获取
	Components []*Component `json:"components,omitempty"`
	// 支持的容器运行时
	ContainerRuntime *ContainerRuntime `json:"containerRuntime"`
	// 集群自定义名称，命名规则：由数字、汉字、英文字符、短划线(-)组成，长度范围1~63个字符
	CustomName string `json:"customName"`
	// 集群的描述，要求长度小于100
	Description string `json:"description,omitempty"`
	// 是否开启集群删除保护，取值：
	// `true`：开启删除保护
	// `false`：不开启删除保护
	// 默认：`false`
	EnableDeleteProtection bool `json:"enableDeleteProtection,omitempty"`
	// Pod IP数量，结合PodCidr的子网掩码计算，需要保证集群可以容纳的节点数大于16
	IPNum uint32 `json:"ipNum"`
	// k8s扩展类型，取值：
	// `K8SEXTENSION_NATIVE`：标准版k8s集群
	// `K8SEXTENSION_EDGE`：云边协同版k8s集群
	// 默认：`K8SEXTENSION_EDGE`
	K8SExtension string `json:"k8sExtension,omitempty"`
	// 集群k8s版本，可选值通过调用集群管理可创建集群k8s版本获取。默认值：`1.19.16`
	K8sVersion string `json:"k8sVersion,omitempty"`
	// 集群的KubeProxy模式，取值:
	// `KUBEPROXYMODE_IPTABLES`：iptables
	// `KUBEPROXYMODE_IPVS`：ipvs
	KubeProxyMode string `json:"kubeProxyMode"`
	// 集群的标签，标签的key值不能重复
	Labels []*Label `json:"labels,omitempty"`
	// 集群控制平面节点信息
	MasterNodes *MasterNode `json:"masterNodes"`
	// 集群Pod网段信息，不能与VPC和Service网段冲突
	PodCidr string `json:"podCidr"`
	// 集群Service网段信息，不能与Pod和VPC网段冲突
	ServiceCidr string `json:"serviceCidr"`
	// 集群负载均衡器网络设置
	SlbConfig *SlbConfig `json:"slbConfig"`
	// 集群工作节点信息
	WorkerNodes *WorkerNode `json:"workerNodes"`
}

// SlbConfig configures LB network
type SlbConfig struct {
	// 是否使用EIP。
	// `true`：自动分配EIP
	// `false`：使用已有IP
	// 默认`false`
	AllocEip bool `json:"allocEip"`
	// 当allocEip=false时，通过本参数指定已有eip
	ApiServerEipId string `json:"apiServerEipId"`
	// 自动分配的EIP限制带宽大小（单位Mbps）
	// allocEip=false时不生效，带宽大小由所选eip自身带宽大小决定
	BwSize uint32 `json:"bwSize"`
}

// MasterNode configures master nodes
type MasterNode struct {
	// 数据盘信息
	DataDisks []*Disk `json:"dataDisks,omitempty"`
	// 镜像名称，通过调用基础信息管理的可用镜像接口获取
	ImageName string `json:"imageName"`
	// 是否将容器运行时挂载最后一块数据盘。
	// `true`：表示挂载
	// `false`：表示不挂载
	// 默认值：`false`
	MountLastDisk bool `json:"mountLastDisk,omitempty"`
	// VPC网络信息
	NetworkInfo *NetworkInfo `json:"networkInfo"`
	// 节点的ECX集群编码，通过调用基础信息管理的可用集群获取
	NodeCode string `json:"nodeCode"`
	// master节点数量，可选值：`1`，`3`
	Num uint32 `json:"num"`
	// 登录密码，要求：
	// 密码长度为8~26个字符
	// 密码需为字母（区分大小写）、数字和特殊字符的组合，不含空格和中文符号
	// 密码不能包含与账号相关的信息，不建议包含账号完整字符串、大小写变为或形似变换的字符串
	// 密码不能使用连续3个及以上键位排序字符，如123，Qwe
	// 密码不能使用常用的具有特殊含义的字符串或形似变换的字符串
	Password string `json:"password"`
	// 系统盘信息，若虚机规格是宿主共享型则不填，否则必填
	SystemDisk *Disk `json:"systemDisk,omitempty"`
	// 虚机规格名称
	VmInstanceName string `json:"vmInstanceName"`
}

// WorkerNode configures worker nodes
type WorkerNode struct {
	DataDisks     []*Disk      `json:"dataDisks,omitempty"`
	ImageName     string       `json:"imageName"`
	MountLastDisk bool         `json:"mountLastDisk,omitempty"`
	NetworkInfo   *NetworkInfo `json:"networkInfo"`
	NodeCode      string       `json:"nodeCode"`
	// 所属节点池名称，校验规则同集群自定义名称
	NodePoolName   string `json:"nodePoolName"`
	Num            uint32 `json:"num"`
	Password       string `json:"password"`
	SystemDisk     *Disk  `json:"systemDisk,omitempty"`
	VmInstanceName string `json:"vmInstanceName"`
}

// NetworkInfo configures network info
type NetworkInfo struct {
	// VPC子网ID，通过调用基础信息管理中集群可用VPC接口获取
	SubnetId uint32 `json:"subnetId"`
	// VPC ID，通过调用基础信息管理中集群可用VPC接口获取
	VpcId uint32 `json:"vpcId"`
}

// Disk configures nodes disks
type Disk struct {
	// 磁盘数量
	Count uint32 `json:"count"`
	// 磁盘IO类型，取值：
	// `DISK_IO_TYPE_NORMAL`：高IO
	// `DISK_IO_TYPE_HIGH`：通用型SSD
	// `DISK_IO_TYPE_ULTRA`：超高IO
	IOType string `json:"ioType"`
	// 磁盘大小，单位：GB
	Size uint32 `json:"size"`
	// 磁盘类型，取值：
	// `DISK_TYPE_CLOUD_DISK`：云盘
	// `DISK_TYPE_LVM_DISK`：lvm本地盘
	// `DISK_TYPE_ZFS`：zfs本地盘
	DiskType string `json:"type"`
}

// Component defines the component to be installed when creating ECK
type Component struct {
	// 组件名称，可以安装的组件通过查询平台支持安装的组件接口获取
	Name string `json:"name"`
}

// Label cluster label
type Label struct {
	// 标签的key，校验规则同k8s标签校验规则
	Key string `json:"key"`
	// 标签的value，校验规则同k8s标签校验规则
	Value string `json:"value"`
}

// ContainerRuntime cluster container runtime
type ContainerRuntime struct {
	// 容器运行时名称, 取值：
	// `docker`
	// `containerd`
	// k8s集群版本大于1.24后，不支持docker。
	Name string `json:"name"`
	// 容器运行时版本，最新版本：
	// `docker`：`20.10.18`
	// `containerd`：`1.6.9`
	Version string `json:"version"`
}

// CommonResponse common response
type CommonResponse struct {
	StatusCode string `json:"statusCode,omitempty"`
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ListRegionResponse response for ListRegion
type ListRegionResponse struct {
	Message    string         `json:"message"`
	StatusCode string         `json:"statusCode"`
	ReturnObj  *ListRegionObj `json:"returnObj"`
}

// ListRegionObj xxx
type ListRegionObj struct {
	Regions   []*Region `json:"regions"`
	RequestId string    `json:"requestId"`
}

// Region region info
type Region struct {
	Name  string  `json:"name"`
	Zones []*Zone `json:"regionClusters"`
}

// Zone zone info
type Zone struct {
	Name     string `json:"name"`
	NodeCode string `json:"nodeCode"`
}

// GetClusterResponse response for GetCluster
type GetClusterResponse struct {
	CommonResponse
	ReturnObj *GetClusterReObj `json:"returnObj,omitempty"`
}

// GetClusterReObj xxx
type GetClusterReObj struct {
	RequestId string   `json:"requestId"`
	Cluster   *Cluster `json:"cluster"`
}

// Cluster ECK cluster
type Cluster struct {
	ClusterId             string      `json:"clusterId"`
	CustomName            string      `json:"customName"`
	K8sVersion            string      `json:"k8sVersion"`
	State                 string      `json:"state"`
	VpcInfo               *VpcInfo    `json:"vpcInfo"`
	NodeCode              string      `json:"nodeCode"`
	RegionName            string      `json:"regionName"`
	ClusterName           string      `json:"clusterName"`
	PodCidr               string      `json:"podCidr"`
	ServiceCidr           string      `json:"serviceCidr"`
	IpNum                 uint32      `json:"ipNum"`
	InternalApiServerAddr string      `json:"internalApiServerAddr"`
	ExternalApiServerAddr string      `json:"externalApiServerAddr"`
	Description           string      `json:"description"`
	Labels                []*Label    `json:"labels,omitempty"`
	InternalKubeConfig    *KubeConfig `json:"internalKubeConfig,omitempty"`
	ExternalKubeConfig    *KubeConfig `json:"externalKubeConfig,omitempty"`
	K8sExtension          string      `json:"k8sExtension"`
	DeleteProtection      bool        `json:"deleteProtection"`
}

// VpcInfo vpc info
type VpcInfo struct {
	VpcName    string `json:"vpcName"`
	StartIp    string `json:"startIp"`
	EndIp      string `json:"endIp"`
	SubnetMask uint32 `json:"subnetMask"`
	CidrBlock  string `json:"cidrBlock"`
	VpcId      string `json:"vpcId"`
}

// KubeConfig kube config
type KubeConfig struct {
	KubeConfigType string `json:"type"`
	Content        string `json:"content"`
	ExpireTime     string `json:"expireTime"`
}

// CreateClusterResponse response for CreateCluster
type CreateClusterResponse struct {
	CommonResponse
	ReturnObj *CreateClusterReObj `json:"returnObj,omitempty"`
}

// CreateClusterReObj xxx
type CreateClusterReObj struct {
	RequestId string `json:"requestId"`
	ClusterId string `json:"clusterId"`
	TaskId    string `json:"taskId"`
}

// DeleteClusterReq request info for DeleteCluster
type DeleteClusterReq struct {
	ClusterId      string   `json:"clusterId"`
	ReservedLbIds  []string `json:"reservedLbIds"`
	ReservedNatIds []string `json:"reservedNatIds"`
	ReservedSgId   string   `json:"reservedSgId"`
}

// DeleteClusterResponse response for DeleteCluster
type DeleteClusterResponse struct {
	CommonResponse
	ReturnObj *DeleteClusterReObj `json:"returnObj,omitempty"`
}

// DeleteClusterReObj xxx
type DeleteClusterReObj struct {
	RequestId string `json:"requestId"`
	TaskId    string `json:"taskId"`
}

// GetKubeConfigResponse response for GetKubeConfig
type GetKubeConfigResponse struct {
	CommonResponse
	ReturnObj *GetKubeConfigReObj `json:"returnObj,omitempty"`
}

// GetKubeConfigReObj xxx
type GetKubeConfigReObj struct {
	RequestId          string      `json:"requestId"`
	InternalKubeConfig *KubeConfig `json:"internalKubeConfig"`
	ExternalKubeConfig *KubeConfig `json:"externalKubeConfig"`
}

// ListNodeReq request for ListNode
type ListNodeReq struct {
	ClusterID  string `json:"clusterID"`
	NodeNames  string `json:"nodeNames,omitempty"`
	NodePoolId string `json:"nodePoolId,omitempty"`
	Page       uint32 `json:"page,omitempty"`
	PerPage    uint32 `json:"perPage,omitempty"`
}

// ListNodeResponse response for ListNode
type ListNodeResponse struct {
	CommonResponse
	ReturnObj *ListNodeReObj `json:"returnObj,omitempty"`
}

// ListNodeReObj xxx
type ListNodeReObj struct {
	RequestId string  `json:"requestId"`
	Nodes     []*Node `json:"nodes"`
	Paging    *Paging `json:"paging"`
}

// Node ECK node info
type Node struct {
	NodePoolId     string `json:"nodePoolId"`
	NodeName       string `json:"nodeName"`
	InstanceId     string `json:"instanceId"`
	InnerIp        string `json:"innerIp"`
	State          string `json:"state"`
	Role           string `json:"role"`
	SsExternalNode bool   `json:"isExternalNode"`
	PriceType      string `json:"priceType"`
	CycleCnt       uint32 `json:"cycleCnt"`
	CycleType      string `json:"cycleType"`
	CreateTime     string `json:"createTime"`
}

// GetNodePoolResponse response for GetNodePool
type GetNodePoolResponse struct {
	CommonResponse
	ReturnObj *GetNodePoolReObj `json:"returnObj"`
}

// GetNodePoolReObj xxx
type GetNodePoolReObj struct {
	RequestId string    `json:"requestId"`
	NodePool  *NodePool `json:"nodePool"`
}

// NodePool ECK GetNodePool info
type NodePool struct {
	ClusterId         string       `json:"clusterId"`
	NodePoolId        string       `json:"nodePoolId"`
	EnableSnat        bool         `json:"enableSnat"`
	NodePoolConf      NodePoolConf `json:"nodePoolConf"`
	Labels            []*Label     `json:"labels"`
	EnableAutoScaling string       `json:"enableAutoScaling"`
	MaximumCapacity   uint32       `json:"maximumCapacity"`
	MinimalCapacity   uint32       `json:"minimalCapacity"`
	ContainerRuntime  string       `json:"containerRuntime"`
}

// NodePoolConf ECK nodepool config
type NodePoolConf struct {
	NodeCode       string       `json:"nodeCode"`
	ImageName      string       `json:"imageName"`
	NodePoolName   string       `json:"nodePoolName"`
	MountLastDisk  bool         `json:"mountLastDisk"`
	NetworkInfo    *NetworkInfo `json:"networkInfo"`
	VmInstanceName string       `json:"vmInstanceName"`
	SystemDisk     *Disk        `json:"systemDisk"`
	DataDisks      []*Disk      `json:"dataDisks"`
}

// Paging info for paging
type Paging struct {
	TotalPage   uint32 `json:"totalPage"`
	Page        uint32 `json:"page"`
	PerPage     uint32 `json:"perPage"`
	TotalRecord uint32 `json:"totalRecord"`
}

// ListNodePoolReq request for ListNodePool
type ListNodePoolReq struct {
	ClusterID            string `json:"clusterID"`
	EnableAutoScaling    string `json:"enableAutoScaling,omitempty"`
	NodePoolName         string `json:"nodePoolName,omitempty"`
	Page                 uint32 `json:"page,omitempty"`
	PerPage              uint32 `json:"perPage,omitempty"`
	RetainSystemNodePool bool   `json:"retainSystemNodePool,omitempty"`
}

// ListNodePoolResponse response for ListNodePool
type ListNodePoolResponse struct {
	CommonResponse
	ReturnObj *ListNodePoolReObj `json:"returnObj"`
}

// ListNodePoolReObj xxx
type ListNodePoolReObj struct {
	RequestId string        `json:"requestId"`
	Paging    *Paging       `json:"paging"`
	NodePools []*NodePoolV2 `json:"nodePools"`
}

// NodePoolV2 ECK ListNodePool info
type NodePoolV2 struct {
	NodePoolId        string   `json:"nodePoolId"`
	Name              string   `json:"name"`
	NodePooltype      string   `json:"type"`
	State             string   `json:"state"`
	VmInfo            *VmInfo  `json:"vmInfo"`
	NodeNum           *NodeNum `json:"nodeNum"`
	RegionName        string   `json:"regionName"`
	ClusterName       string   `json:"clusterName"`
	VmImageName       string   `json:"vmImageName"`
	UpdateTime        string   `json:"updateTime"`
	EnableAutoScaling string   `json:"enableAutoScaling"`
	MaximumCapacity   uint32   `json:"maximumCapacity"`
	MinimalCapacity   uint32   `json:"minimalCapacity"`
	NodeCode          string   `json:"nodeCode"`
	PriceType         string   `json:"priceType"`
	CycleCnt          uint32   `json:"cycleCnt"`
	CycleType         string   `json:"cycleType"`
}

// VmInfo nodepool VM info
type VmInfo struct {
	InstanceName string `json:"instanceName"`
	Mem          uint32 `json:"mem"`
	VCpu         uint32 `json:"vCpu"`
}

// NodeNum node num for different node states
type NodeNum struct {
	All    uint32 `json:"all"`
	Normal uint32 `json:"normal"`
}

// ListVpcResponse response for ListVpc
type ListVpcResponse struct {
	CommonResponse
	ReturnObj *ListVpcReObj `json:"returnObj,omitempty"`
}

// ListVpcReObj xxx
type ListVpcReObj struct {
	RequestId string `json:"requestId"`
	Vpcs      []*Vpc `json:"vpcs"`
}

// Vpc info for Vpc
type Vpc struct {
	VpcId     uint32    `json:"vpcId"`
	Name      string    `json:"name"`
	CidrBlock string    `json:"cidrBlock"`
	Subnets   []*Subnet `json:"subnets"`
}

// Subnet info for Subnet
type Subnet struct {
	SubnetId         uint32 `json:"subnetId"`
	Name             string `json:"name"`
	CidrBlock        string `json:"cidrBlock"`
	AvailableIpCount uint32 `json:"availableIpCount"`
}
