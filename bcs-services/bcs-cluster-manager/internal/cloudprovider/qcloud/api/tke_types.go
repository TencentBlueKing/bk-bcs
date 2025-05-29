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

package api

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
)

// EndpointStatus endpoint status
type EndpointStatus string

// Created status
func (es EndpointStatus) Created() bool {
	return es == Created
}

// CreateFailed status
func (es EndpointStatus) CreateFailed() bool {
	return es == CreateFailed
}

// Creating status
func (es EndpointStatus) Creating() bool {
	return es == Creating
}

// NotFound status
func (es EndpointStatus) NotFound() bool {
	return es == NotFound
}

// Deleted status
func (es EndpointStatus) Deleted() bool {
	return es == Deleted
}

// Status return es
func (es EndpointStatus) Status() string {
	return string(es)
}

var (
	// Created status
	Created EndpointStatus = "Created"
	// CreateFailed status
	CreateFailed EndpointStatus = "CreateFailed"
	// Creating status
	Creating EndpointStatus = "Creating"
	// NotFound status
	NotFound EndpointStatus = "NotFound"
	// Deleted status
	Deleted EndpointStatus = "Deleted"
)

// ActivityStatus activity status
type ActivityStatus string

// String return es
func (as ActivityStatus) String() string {
	return string(as)
}

var (
	// InitActivity ActivityStatus xxx
	InitActivity ActivityStatus = "INIT"
	// RunningActivity ActivityStatus xxx
	RunningActivity ActivityStatus = "RUNNING"
	// SuccessfulActivity ActivityStatus xxx
	SuccessfulActivity ActivityStatus = "SUCCESSFUL"
	// SuccessfulPartActivity ActivityStatus xxx
	SuccessfulPartActivity ActivityStatus = "PARTIALLY_SUCCESSFUL"
	// FailedActivity ActivityStatus  xxx
	FailedActivity ActivityStatus = "FAILED"
	// CancelledActivity ActivityStatus xxx
	CancelledActivity ActivityStatus = "CANCELLED" // nolint
)

// InstanceAsStatus 实例在伸缩活动中的状态
type InstanceAsStatus string

// String return is
func (is InstanceAsStatus) String() string {
	return string(is)
}

var (
	// InitInstanceAS xxx
	InitInstanceAS InstanceAsStatus = "INIT"
	// RunningInstanceAS xxx
	RunningInstanceAS InstanceAsStatus = "RUNNING"
	// SuccessfulInstanceAS xxx
	SuccessfulInstanceAS InstanceAsStatus = "SUCCESSFUL"
	// FailedInstanceAS xxx
	FailedInstanceAS InstanceAsStatus = "FAILED"
)

// InstanceTkeStatus tke集群中实例状态
type InstanceTkeStatus string

// String return is
func (is InstanceTkeStatus) String() string {
	return string(is)
}

var (
	// InitInstanceTke xxx
	InitInstanceTke InstanceAsStatus = "initializing"
	// RunningInstanceTke xxx
	RunningInstanceTke InstanceAsStatus = "running"
	// FailedInstanceTke xxx
	FailedInstanceTke InstanceAsStatus = "failed"
)

const (
	// TkeClusterType tke
	TkeClusterType = "tke"
	// EksClusterType eks
	EksClusterType = "eks"
)

const (
	// DockerGraphPath default docker graphPath
	DockerGraphPath = "/data/bcs/service/docker"
	// MountTarget default mountTarget
	MountTarget = "/data"
)

// filter key
const (
	// NodePoolIDKey poolID
	NodePoolIDKey = "nodepool-id"
	// NodePoolInstanceTypeKey instanceType
	NodePoolInstanceTypeKey = "nodepool-instance-type"

	// NodePoolInstanceManually manuallyAdd
	NodePoolInstanceManually = "MANUALLY_ADDED"
	// NodePoolInstanceAuto autoAdd
	NodePoolInstanceAuto = "AUTOSCALING_ADDED"
	// NodePoolInstanceAll instanceAll
	NodePoolInstanceAll = "ALL"

	// TKERouteEni router-eni
	TKERouteEni = "tke-route-eni"
	// TKEDirectEni direct-eni
	TKEDirectEni = "tke-direct-eni"
	// DefaultRegion for default setting
	DefaultRegion = regions.Shanghai
)

// QueryFilter xxx
type QueryFilter interface {
	// BuildFilters build filters
	BuildFilters() []*tke.Filter
}

// QueryClusterInstanceFilter xxx
type QueryClusterInstanceFilter struct {
	NodePoolID           string
	NodePoolInstanceType string
}

// BuildFilters build filter
func (filter QueryClusterInstanceFilter) BuildFilters() []*tke.Filter {
	filters := make([]*tke.Filter, 0)

	if len(filter.NodePoolID) > 0 {
		filters = append(filters, &tke.Filter{
			Name:   common.StringPtr(NodePoolIDKey),
			Values: []*string{common.StringPtr(filter.NodePoolID)},
		})
	}

	if len(filter.NodePoolInstanceType) > 0 {
		filters = append(filters, &tke.Filter{
			Name:   common.StringPtr(NodePoolInstanceTypeKey),
			Values: []*string{common.StringPtr(filter.NodePoolInstanceType)},
		})
	}

	return filters
}

var (
	// MASTER role
	MASTER InstanceRole = "MASTER"
	// WORKER role
	WORKER InstanceRole = "WORKER"
	// ETCD role
	ETCD InstanceRole = "ETCD"
	// MASTER_ETCD role
	MASTER_ETCD InstanceRole = "MASTER_ETCD"
	// ALL role
	ALL InstanceRole = "ALL"
)

// InstanceRole for instanceType
type InstanceRole string

// String to string
func (ir InstanceRole) String() string {
	return string(ir)
}

// DeleteMode instance deletedMode
type DeleteMode string

var (
	// Terminate mode
	Terminate DeleteMode = "terminate"
	// Retain mode
	Retain DeleteMode = "retain"
)

// String to string
func (dm DeleteMode) String() string {
	return string(dm)
}

// Versions cluster k8s version
type Versions struct {
	Name    string
	Version string
}

// Images cluster image
type Images struct {
	OsName  string
	ImageID string
}

// DescribeClusterInstances cluster instances request
type DescribeClusterInstances struct {
	// ClusterID cluster unique id
	ClusterID string `json:"clusterID"`
	// InstanceIDs instance ids
	InstanceIDs []string `json:"instanceIDs"`
	// InstanceRole master/node &&
	InstanceRole InstanceRole `json:"instanceRole"`
	// Offset offset
	Offset int64 `json:"offset"`
	// Limit query num
	Limit int64 `json:"limit"`
}

// DeleteInstancesRequest xxx
type DeleteInstancesRequest struct {
	ClusterID   string     `json:"clusterID"`
	Instances   []string   `json:"instances"`
	DeleteMode  DeleteMode `json:"deleteMode"`
	ForceDelete bool       `json:"forceDelete,omitempty"`
}

// DeleteInstancesResult xxx
type DeleteInstancesResult struct {
	Success  []string `json:"success"`
	Failure  []string `json:"failure"`
	NotFound []string `json:"notFound"`
}

func (dir *DeleteInstancesRequest) validate() error {
	if len(dir.ClusterID) == 0 {
		return fmt.Errorf("DeleteTkeClusterInstance failed: clusterID is empty")
	}

	if len(dir.Instances) == 0 {
		return fmt.Errorf("DeleteTkeClusterInstance failed: InstancesList is empty")
	}

	if dir.DeleteMode != Terminate && dir.DeleteMode != Retain && dir.DeleteMode != "" {
		return fmt.Errorf("DeleteTkeClusterInstance[%s] invalid deleteMode[%s]", dir.ClusterID, dir.DeleteMode)
	}

	return nil
}

var (
	// GlobalRouteCIDRCheck global router
	GlobalRouteCIDRCheck = "GlobalRouteCIDRCheck"
	// VpcCniCIDRCheck vpc-cni subnet resource
	VpcCniCIDRCheck = "VpcCniCIDRCheck"
)

// AddExistedInstanceReq xxx
type AddExistedInstanceReq struct {
	ClusterID       string                    `json:"clusterID"`
	InstanceIDs     []string                  `json:"instanceIDs"`
	AdvancedSetting *InstanceAdvancedSettings `json:"advancedSetting"`
	// 实例所属安全组. 该参数可以通过调用 DescribeSecurityGroups 的返回值中的sgId字段来获取。若不指定该参数，则绑定默认安全组。
	SecurityGroupIds []string `json:"securityGroupIds"`
	// NodePool nodePool conf
	NodePool        *NodePoolOption  `json:"nodePool"`
	EnhancedSetting *EnhancedService `json:"enhancedSetting"`
	LoginSetting    *LoginSettings   `json:"loginSetting"`
	// SkipValidateOptions 校验规则相关选项，可配置跳过某些校验规则。目前支持GlobalRouteCIDRCheck(跳过GlobalRouter的相关校验),
	// VpcCniCIDRCheck（跳过VpcCni相关校验）
	SkipValidateOptions []string `json:"skipValidateOptions"`
	// InstanceAdvancedSettingsOverrides 参数InstanceAdvancedSettingsOverride数组的长度应与InstanceIds数组一致；
	// 当长度大于InstanceIds数组长度时将报错；当长度小于InstanceIds数组时，没有对应配置的instace将使用默认配置。
	InstanceAdvancedSettingsOverrides []*InstanceAdvancedSettings `json:"instanceAdvancedSettingsOverrides"`
	// ImageId 节点镜像ID
	ImageId string `json:"imageId"`
}

func (aei *AddExistedInstanceReq) validate() error {
	if aei == nil {
		return fmt.Errorf("AddExistedInstanceReq is nil")
	}

	if aei.ClusterID == "" {
		return fmt.Errorf("AddExistedInstancesToCluster failed: clusterID is empty")
	}

	if len(aei.InstanceIDs) == 0 {
		return fmt.Errorf("AddExistedInstancesToCluster failed: instanceIDs is empty")
	}

	return nil
}

// AddExistedInstanceRsp add existed instances result
type AddExistedInstanceRsp struct {
	FailedInstanceIDs  []string `json:"failedInstanceIDs"`
	FailedReasons      []string `json:"failedReasons"`
	SuccessInstanceIDs []string `json:"successInstanceIDs"`
	TimeoutInstanceIDs []string `json:"timeoutInstanceIDs"`
}

const (
	// Ext4 fs type
	Ext4 = "ext4"
)

// GetDefaultDataDisk set disk ext4 type data when cvm has many disks
func GetDefaultDataDisk(fsType string) DataDetailDisk {
	return DataDetailDisk{
		FileSystem:         fsType,
		MountTarget:        MountTarget,
		AutoFormatAndMount: true,
		DiskPartition:      "/dev/vdb",
	}
}

// DefaultDiskPartition default disk partition
var DefaultDiskPartition = []string{"/dev/vdb", "/dev/vdc", "/dev/vdd", "/dev/vde", "/dev/vdf"}

// InstanceAdvancedSettings instance advanced setting
type InstanceAdvancedSettings struct {
	// 数据盘挂载点, 默认不挂载数据盘. 已格式化的 ext3，ext4，xfs 文件系统的数据盘将直接挂载，
	// 其他文件系统或未格式化的数据盘将自动格式化为ext4 (tlinux系统格式化成xfs)并挂载，请注意备份数据!
	// 无数据盘或有多块数据盘的云主机此设置不生效。
	// 注意，多盘场景请使用下方的DataDisks数据结构，设置对应的云盘类型、云盘大小、挂载路径、是否格式化等信息。
	// MountTarget data disk mountPoint
	MountTarget string `json:"mountTarget"`
	// DockerGraphPath dockerd --graph
	DockerGraphPath string `json:"dockerGraphPath"`
	// UserScript  base64 编码的用户脚本, 此脚本会在 k8s 组件运行后执行, 需要用户保证脚本的可重入及重试逻辑
	UserScript string `json:"userScript"`
	// Unschedulable involved scheduler, 默认值是0 表示参与调度
	Unschedulable *int64 `json:"unschedulable"`
	// Labels instance labels
	Labels []*KeyValue `json:"labels"`
	// DataDisks many disk mount info
	DataDisks []DataDetailDisk `json:"dataDisks"`
	// ExtraArgs component start parameter
	ExtraArgs *InstanceExtraArgs `json:"extraArgs"`
	// PreStartUserScript base64 编码的用户脚本，在初始化节点之前执行，目前只对添加已有节点生效
	PreStartUserScript string `json:"preStartUserScript"`
	// TaintList 节点污点
	TaintList []*Taint `json:"taintList"`
	// GPUArgs GPU参数信息
	GPUArgs *GPUArgs `json:"GPUArgs"`
}

// KeyValue struct(name/value)
type KeyValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// InstanceExtraArgs kubelet startup parameter
type InstanceExtraArgs struct {
	// Kubelet user-defined parameter，["k1=v1", "k1=v2"]
	// for example: ["root-dir=/var/lib/kubelet","feature-gates=PodShareProcessNamespace=true,
	// DynamicKubeletConfig=true"]
	Kubelet []string `json:"kubelet"`
}

// DataDetailDisk data disk
type DataDetailDisk struct {
	// DiskType type
	DiskType string `json:"diskType"`
	// DiskSize size
	DiskSize int64 `json:"diskSize"`
	// FileSystem file system 文件系统(ext3/ext4/xfs)
	FileSystem string `json:"fileSystem"`
	// MountTarget mount point
	MountTarget string `json:"mountTarget"`
	// AutoFormatAndMount auto format and mount
	AutoFormatAndMount bool `json:"autoFormatAndMount"`
	// DiskPartition 挂载设备名或分区名，当且仅当添加已有节点时需要
	DiskPartition string `json:"diskPartition"`
}

// EnhancedService tke cluster enhanced service
type EnhancedService struct {
	// 开启云安全服务。若不指定该参数，则默认开启云安全服务。
	SecurityService *RunSecurityServiceEnabled `json:"SecurityService,omitempty" name:"SecurityService"`

	// 开启云监控服务。若不指定该参数，则默认开启云监控服务。
	MonitorService *RunMonitorServiceEnabled `json:"MonitorService,omitempty" name:"MonitorService"`
}

// RunMonitorServiceEnabled tke cluster monitor service
type RunMonitorServiceEnabled struct {
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`
}

// RunSecurityServiceEnabled 开启云安全服务。若不指定该参数，则默认开启云安全服务。
type RunSecurityServiceEnabled struct {
	Enabled *bool `json:"Enabled,omitempty" name:"Enabled"`
}

// LoginSettings reset passwd
type LoginSettings struct {
	// Password reset instance passwd
	Password string `json:"Password,omitempty"`
	// KeyIds secret
	KeyIds []string `json:"KeyIds,omitempty"`
}

// NodePoolOption nodePool options
type NodePoolOption struct {
	// AddToNodePool add node to pool
	AddToNodePool bool `json:"addToNodePool"`
	// NodePoolId nodePool ID
	NodePoolID string `json:"nodePoolID"`
	// InheritConfigurationFromNodePool conf from nodePoll
	InheritConfigurationFromNodePool bool
}

// CreateCVMRequest create cluster cvm request
type CreateCVMRequest struct {
	// tke clusterId, required
	ClusterID string `json:"clusterId"`
	// VPCID, required
	VPCID string `json:"vpcId"`
	// subnet, required
	SubNetID string `json:"subnetId"`
	// available zone, required
	Zone string `json:"zone"`
	// cvm number, required
	ApplyNum uint32 `json:"applyNum"`
	// cvm instance type, required
	InstanceType string `json:"instanceType"`
	// required
	SystemDiskType string `json:"systemDiskType"`
	// required
	SystemDiskSize uint32 `json:"systemDiskSize"`
	// dataDisk, optional
	DataDisks []*DataDisk `json:"dataDisk"`
	// image information for system, required
	Image *ImageInfo `json:"image"`
	// security group, optional
	Security *SecurityGroup `json:"security"`
	// setup security service, optional, default 0
	IsSecurityService uint32 `json:"isSecurityService,omitempty"`
	// cloud monitor, optional, default 1
	IsMonitorService uint32 `json:"isMonitorService"`
	// cvm instance name, optional
	InstanceName string `json:"instanceName,omitempty"`
	// instance login setting
	Login LoginSettings `json:"login"`
	// required
	Operator string `json:"operator"`
}

// ImageInfo for system
type ImageInfo struct {
	ID   string `json:"imageId"`           // required
	Name string `json:"imageName"`         // required
	OS   string `json:"imageOs,omitempty"` // optional
	Type string `json:"imageType"`         // optional
}

// SecurityGroup sg information
type SecurityGroup struct {
	ID   string `json:"securityGroupId"`             // required
	Name string `json:"securityGroupName,omitempty"` // optional
	Desc string `json:"securityGroupDesc,omitempty"` // optional
}

// DataDisk for CVMOrder
type DataDisk struct {
	DiskType string `json:"dataDiskType"`
	DiskSize uint32 `json:"dataDiskSize"`
}

// CreateClusterResponse xxx
type CreateClusterResponse struct {
	// ClusterID cloud clusterID
	ClusterID string
}

// CreateClusterRequest xxx
type CreateClusterRequest struct {
	// AddNodeMode add node by existed nodes or run instances(true: runInstances false: addExistedNodes)
	AddNodeMode bool `json:"addNodeMode"`
	// Region regionInfo, required
	Region string `json:"region"`
	// ClusterType  托管集群：MANAGED_CLUSTER，独立集群：INDEPENDENT_CLUSTER; required
	ClusterType string `json:"clusterType"`
	// ClusterCIDRSettings cluster network config info
	ClusterCIDR *ClusterCIDRSettings `json:"clusterCIDR"`
	// ClusterBasic cluster basic config info
	ClusterBasic *ClusterBasicSettings `json:"clusterBasic"`
	// ClusterAdvanced cluster advanced config info
	ClusterAdvanced *ClusterAdvancedSettings `json:"clusterAdvanced"`
	// InstanceAdvanced instance advanced config
	InstanceAdvanced *InstanceAdvancedSettings `json:"instanceAdvanced"`
	// ExistedInstancesForNode use existed instance when create cluster or add node(must belong to )
	ExistedInstancesForNode []*ExistedInstancesForNode `json:"existedInstancesForNode"`
	// RunInstancesForNode run instance by CVM create request
	// CVM创建透传参数，json化字符串格式，详见[CVM创建实例](https://cloud.tencent.com/document/product/213/15730)接口。
	// 总机型(包括地域)数量不超过10个，相同机型(地域)购买多台机器可以通过设置参数中RunInstances中InstanceCount来实现。
	RunInstancesForNode []*RunInstancesForNode `json:"RunInstancesForNode,omitempty" name:"RunInstancesForNode"`
	// InstanceDataDiskMountSettings instance dataDisk mount setting
	InstanceDataDiskMountSettings []*InstanceDataDiskMountSetting `json:"instanceDataDiskMountSettings"`
	// ExtensionAddon createCluster deploy extensionAddon
	Addons []ExtensionAddon `json:"addons"`
}

// ExtensionAddon addon parameters
type ExtensionAddon struct {
	AddonName  string
	AddonParam string
}

// ClusterCIDRSettings cluster network config
type ClusterCIDRSettings struct {
	// ClusterCIDR allocate cluster pod and service ip, cidr must be internal segment.
	// for example:10.1.0.0/14, 192.168.0.1/18,172.16.0.0/16
	ClusterCIDR string `json:"clusterCIDR,omitempty"`
	// MaxNodePodNum cluster max pods per node
	MaxNodePodNum uint64 `json:"maxNodePodNum,omitempty"`
	// MaxClusterServiceNum cluster max services number
	MaxClusterServiceNum uint64 `json:"maxClusterServiceNum,omitempty"`
	// ServiceCIDR cluster service CIDR
	ServiceCIDR string `json:"serviceCIDR,omitempty"`
	// EniSubnetIds VPC-CNI网络模式下，弹性网卡的子网Id
	EniSubnetIds []string `json:"eniSubnetIds,omitempty"`
	// ClaimExpiredSeconds VPC-CNI网络模式下，弹性网卡IP的回收时间，取值范围[300,15768000)
	ClaimExpiredSeconds uint32 `json:"claimExpiredSeconds,omitempty"`
}

// ClusterBasicSettings cluster basic config
type ClusterBasicSettings struct {
	// ClusterOS
	ClusterOS string `json:"clusterOS,omitempty"`
	// ClusterVersion(kubernetes version)
	ClusterVersion string `json:"clusterVersion"`
	// ClusterName(kubernetes cluster name)
	ClusterName string `json:"clusterName"`
	// ClusterDescription cluster description
	ClusterDescription string `json:"ClusterDescription,omitempty"`
	// VpcID cluster private network id, for example: vpc-xxx
	VpcID string `json:"vpcID"`
	// ProjectID
	ProjectID int64 `json:"projectID,omitempty"`
	// TagSpecification describe cluster tag list, resourceType only support "cluster"
	TagSpecification []*TagSpecification `json:"tagSpecification,omitempty"`
	// SubnetID 当选择Cilium Overlay网络插件时，TKE会从该子网获取2个IP用来创建内网负载均衡
	SubnetID string `json:"subnetID,omitempty"`
	// ClusterLevel 集群等级，针对托管集群生效
	ClusterLevel string `json:"clusterLevel,omitempty"`
	// IsAutoUpgrade 集群是否自动升级，针对托管集群生效
	IsAutoUpgradeClusterLevel bool `json:"isAutoUpgrade,omitempty"`
}

// ClusterAdvancedSettings cluster advanced setting
type ClusterAdvancedSettings struct {
	// IPVS enable or disable
	IPVS bool `json:"ipvs"`
	// ContainerRuntime runtime type(docker/containerd) default: docker
	ContainerRuntime string `json:"containerRuntime"`
	// RuntimeVersion runtime version
	RuntimeVersion string `json:"runtimeVersion"`
	// NetworkType cluster networkType(GR/VPC-CNI) default GR
	NetworkType string `json:"networkType,omitempty" name:"NetworkType"`
	// cluster extraArgs (cluster component start parameter)
	ExtraArgs *ClusterExtraArgs `json:"extraArgs"`
	// IsNonStaticIpMode 集群VPC-CNI模式是否为非固定IP，默认: FALSE 固定IP
	IsNonStaticIpMode bool `json:"isNonStaticIpMode"`
	// VpcCniType 共享网卡多IP模式和独立网卡模式，共享网卡多 IP 模式"tke-route-eni"，独立网卡模式"tke-direct-eni"，默认为共享网卡模式
	VpcCniType string `json:"vpcCniType"`
	// DeletionProtection cluster delete protection
	DeletionProtection bool `json:"deletionProtection"`
	// AuditEnabled cluster audit
	AuditEnabled bool `json:"auditEnabled"`
}

// ExistedInstancesForNode use existed nodes for create cluster or add node
type ExistedInstancesForNode struct {
	// NodeRole node role (MASTER_ETCD, WORKER)
	NodeRole string `json:"nodeRole"`
	// ExistedInstancesPara existed instance reinstall parameter
	ExistedInstancesPara *ExistedInstancesPara `json:"existedInstancesPara"`
	// InstanceAdvancedSettingsOverride(节点高级设置), will override cluster instance advanced setting
	InstanceAdvancedSettingsOverride *InstanceAdvancedSettings `json:"instanceAdvancedSettingsOverride"`
}

// ExistedInstancesPara existed instance para
type ExistedInstancesPara struct {
	// InstanceIDs instance ids
	InstanceIDs []string `json:"instanceIDs"`
	// InstanceAdvancedSettings instance advanced setting para(实例额外需要设置参数信息)
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"instanceAdvancedSettings"`
	// EnhancedService enhanced service, start cloudSecurity/cloudMonitor. default on
	EnhancedService *EnhancedService `json:"enhancedService"`
	// LoginSettings instance loginInfo(passwd reset)
	LoginSettings *LoginSettings `json:"loginSettings"`
	// SecurityGroupIds instance security group(实例所属安全组), use default security group when nil
	SecurityGroupIds []*string `json:"securityGroupIds"`
}

// RunInstancesForNode apply instance and set cluster master/node
type RunInstancesForNode struct {
	// NodeRole node role (MASTER_ETCD, WORKER)
	NodeRole string `json:"nodeRole"`
	// CVM创建透传参数，json化字符串格式，详见[CVM创建实例](https://cloud.tencent.com/document/product/213/15730)接口，
	// 传入公共参数外的其他参数即可，其中ImageId会替换为TKE集群OS对应的镜像。
	RunInstancesPara []*string `json:"runInstancesPara"`
	// InstanceAdvancedSettingsOverrides node advanced setting(上边的RunInstancesPara按照顺序一一对应
	// （当前只对节点自定义参数ExtraArgs生效）
	InstanceAdvancedSettingsOverrides []*InstanceAdvancedSettings `json:"InstanceAdvancedSettingsOverrides,omitempty" name:"InstanceAdvancedSettingsOverrides"` // nolint
}

// ClusterExtraArgs cluster extra args
type ClusterExtraArgs struct {
	// KubeAPIServer xxx
	// kube-apiserver自定义参数，参数格式为["k1=v1", "k1=v2"]， 例如["max-requests-inflight=500",
	// "feature-gates=PodShareProcessNamespace=true,DynamicKubeletConfig=true"]
	KubeAPIServer []*string `json:"KubeAPIServer"`
	// KubeControllerManager kube-controller-manager自定义参数
	KubeControllerManager []*string `json:"KubeControllerManager"`
	// KubeScheduler kube-scheduler自定义参数
	KubeScheduler []*string `json:"KubeScheduler"`
	// Etcd etcd自定义参数，只支持独立集群
	Etcd []*string `json:"etcd"`
}

// TagSpecification tag label
type TagSpecification struct {
	// ResourceType 标签绑定的资源类型，当前支持类型："cluster"
	ResourceType string `json:"resourceType"`
	// Tags 标签对列表
	Tags []*Tag `json:"tags"`
}

// Tag key/value
type Tag struct {
	// Key
	Key *string `json:"key"`
	// Value
	Value *string `json:"value"`
}

// InstanceDataDiskMountSetting instance data disk mount setting
type InstanceDataDiskMountSetting struct {
	// InstanceType instance type
	InstanceType *string `json:"instanceType"`
	// DataDisks data disk
	DataDisks []*DataDetailDisk `json:"dataDisks"`
	// Zone instance zone
	Zone *string `json:"zone"`
}

// RegionInfo regionInfo
type RegionInfo struct {
	Region      string `json:"region"`
	RegionName  string `json:"regionName"`
	RegionState string `json:"regionState"`
}

// InstanceInfo instanceInfo
type InstanceInfo struct {
	// 实例ID
	InstanceID string
	// 节点内网IP
	InstanceIP string
	// 节点角色, MASTER, WORKER, ETCD, MASTER_ETCD,ALL, 默认为WORKER
	InstanceRole string
	// 实例的状态（running 运行中，initializing 初始化中，failed 异常）
	InstanceState string
	// 资源池ID
	NodePoolId string
	// 自动伸缩组ID
	AutoscalingGroupId string
}

// EnableVpcCniInput xxx
type EnableVpcCniInput struct {
	TkeClusterID string
	// 开启vpc-cni的模式，tke-route-eni开启的是策略路由模式，tke-direct-eni开启的是独立网卡模式
	VpcCniType string
	SubnetsIDs []string

	EnableStaticIp bool
	ExpiredSeconds int
}

// GetEnableVpcCniProgressOutput xxx
type GetEnableVpcCniProgressOutput struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	RequestID string `json:"requestID"`
}

// VpcCniStatus enable vpc-cni status
type VpcCniStatus string

var (
	// Running status
	Running VpcCniStatus = "Running"
	// Succeed success
	Succeed VpcCniStatus = "Succeed"
	// Failed failure
	Failed VpcCniStatus = "Failed"
)

// AddVpcCniSubnetsInput vpccni subnet
type AddVpcCniSubnetsInput struct {
	// ClusterID clusterID
	ClusterID string
	// VpcID vpcID
	VpcID string
	// SubnetIDs subnetIDs
	SubnetIDs []string
}

// CreateNodePoolInput create node pool input
type CreateNodePoolInput struct {
	// cluster id
	ClusterID *string `json:"ClusterID,omitempty" name:"ClusterID"`

	// AutoScalingGroupPara AS组参数
	AutoScalingGroupPara *AutoScalingGroup `json:"AutoScalingGroupPara,omitempty" name:"AutoScalingGroupPara"`

	// LaunchConfigurePara 运行参数
	LaunchConfigurePara *LaunchConfiguration `json:"LaunchConfigurePara,omitempty" name:"LaunchConfigurePara"`

	// InstanceAdvancedSettings 实例参数
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"InstanceAdvancedSettings,omitempty" name:"InstanceAdvancedSettings"` // nolint

	// 是否启用自动伸缩
	EnableAutoscale *bool `json:"EnableAutoscale,omitempty" name:"EnableAutoscale"`

	// 节点池名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// Labels标签
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// Taints互斥
	Taints []*Taint `json:"Taints,omitempty" name:"Taints"`

	// 节点池纬度运行时类型及版本
	ContainerRuntime *string `json:"ContainerRuntime,omitempty" name:"ContainerRuntime"`

	// 运行时版本
	RuntimeVersion *string `json:"RuntimeVersion,omitempty" name:"RuntimeVersion"`

	// 节点池os
	NodePoolOs *string `json:"NodePoolOs,omitempty" name:"NodePoolOs"`

	// 容器的镜像版本，"DOCKER_CUSTOMIZE"(容器定制版),"GENERAL"(普通版本，默认值)
	OsCustomizeType *string `json:"OsCustomizeType,omitempty" name:"OsCustomizeType"`

	// 资源标签
	Tags []*Tag `json:"Tags,omitempty" name:"Tags"`
}

// Label key/value struct
type Label struct {

	// map表中的Name
	Name *string `json:"Name,omitempty" name:"Name"`

	// map表中的Value
	Value *string `json:"Value,omitempty" name:"Value"`
}

// Taint key/value struct
type Taint struct {

	// Key
	Key *string `json:"Key,omitempty" name:"Key"`

	// Value
	Value *string `json:"Value,omitempty" name:"Value"`

	// Effect
	Effect *string `json:"Effect,omitempty" name:"Effect"`
}

// AutoScalingGroup asg参数
type AutoScalingGroup struct {
	// 伸缩组名称，在您账号中必须唯一。名称仅支持中文、英文、数字、下划线、分隔符"-"、小数点，最大长度不能超55个字节。
	AutoScalingGroupName *string `json:"AutoScalingGroupName,omitempty" name:"AutoScalingGroupName"`

	// 启动配置ID
	LaunchConfigurationID *string `json:"LaunchConfigurationId,omitempty" name:"LaunchConfigurationId"`

	// 最大实例数，取值范围为0-2000。
	MaxSize *uint64 `json:"MaxSize,omitempty" name:"MaxSize"`

	// 最小实例数，取值范围为0-2000。
	MinSize *uint64 `json:"MinSize,omitempty" name:"MinSize"`

	// VPC ID，基础网络则填空字符串
	VpcID *string `json:"VpcId,omitempty" name:"VpcId"`

	// 默认冷却时间，单位秒，默认值为300
	DefaultCooldown *uint64 `json:"DefaultCooldown,omitempty" name:"DefaultCooldown"`

	// 期望实例数，大小介于最小实例数和最大实例数之间
	DesiredCapacity *uint64 `json:"DesiredCapacity,omitempty" name:"DesiredCapacity"`

	// 传统负载均衡器ID列表，目前长度上限为20，LoadBalancerIds 和 ForwardLoadBalancers 二者同时最多只能指定一个
	LoadBalancerIds []*string `json:"LoadBalancerIds,omitempty" name:"LoadBalancerIds"`

	// 伸缩组内实例所属项目ID。不填为默认项目。
	ProjectID *uint64 `json:"ProjectId,omitempty" name:"ProjectId"`

	// 应用型负载均衡器列表，目前长度上限为50，LoadBalancerIds 和 ForwardLoadBalancers 二者同时最多只能指定一个
	ForwardLoadBalancers []*ForwardLoadBalancer `json:"ForwardLoadBalancers,omitempty" name:"ForwardLoadBalancers"`

	// 子网ID列表，VPC场景下必须指定子网。多个子网以填写顺序为优先级，依次进行尝试，直至可以成功创建实例。
	SubnetIds []*string `json:"SubnetIds,omitempty" name:"SubnetIds"`

	// 销毁策略，目前长度上限为1。取值包括 OLDEST_INSTANCE 和 NEWEST_INSTANCE，默认取值为 OLDEST_INSTANCE。
	// <br><li> OLDEST_INSTANCE 优先销毁伸缩组中最旧的实例。
	// <br><li> NEWEST_INSTANCE，优先销毁伸缩组中最新的实例。
	TerminationPolicies []*string `json:"TerminationPolicies,omitempty" name:"TerminationPolicies"`

	// 可用区列表，基础网络场景下必须指定可用区。多个可用区以填写顺序为优先级，依次进行尝试，直至可以成功创建实例。
	Zones []*string `json:"Zones,omitempty" name:"Zones"`

	// 重试策略，取值包括 IMMEDIATE_RETRY、 INCREMENTAL_INTERVALS、NO_RETRY，默认取值为 IMMEDIATE_RETRY。
	// <br><li> IMMEDIATE_RETRY，立即重试，在较短时间内快速重试，连续失败超过一定次数（5次）后不再重试。
	// <br><li> INCREMENTAL_INTERVALS，间隔递增重试，随着连续失败次数的增加，重试间隔逐渐增大，重试间隔从秒级到1天不等。
	// <br><li> NO_RETRY，不进行重试，直到再次收到用户调用或者告警信息后才会重试。
	RetryPolicy *string `json:"RetryPolicy,omitempty" name:"RetryPolicy"`

	// 可用区校验策略，取值包括 ALL 和 ANY，默认取值为ANY。
	// <br><li> ALL，所有可用区（Zone）或子网（SubnetId）都可用则通过校验，否则校验报错。
	// <br><li> ANY，存在任何一个可用区（Zone）或子网（SubnetId）可用则通过校验，否则校验报错。
	//
	// 可用区或子网不可用的常见原因包括该可用区CVM实例类型售罄、该可用区CBS云盘售罄、该可用区配额不足、该子网IP不足等。
	// 如果 Zones/SubnetIds 中可用区或者子网不存在，则无论 ZonesCheckPolicy 采用何种取值，都会校验报错。
	ZonesCheckPolicy *string `json:"ZonesCheckPolicy,omitempty" name:"ZonesCheckPolicy"`

	// 标签描述列表。通过指定该参数可以支持绑定标签到伸缩组。同时绑定标签到相应的资源实例。每个伸缩组最多支持30个标签。
	Tags []*Tag `json:"Tags,omitempty" name:"Tags"`

	// 服务设置，包括云监控不健康替换等服务设置。
	ServiceSettings *ServiceSettings `json:"ServiceSettings,omitempty" name:"ServiceSettings"`

	// 实例具有IPv6地址数量的配置，取值包括 0、1，默认值为0。
	Ipv6AddressCount *int64 `json:"Ipv6AddressCount,omitempty" name:"Ipv6AddressCount"`

	// 多可用区/子网策略，取值包括 PRIORITY 和 EQUALITY，默认为 PRIORITY。
	// <br><li> PRIORITY，按照可用区/子网列表的顺序，作为优先级来尝试创建实例，如果优先级最高的可用区/子网可以创建成功，
	// 则总在该可用区/子网创建。
	// <br><li> EQUALITY：扩容出的实例会打散到多个可用区/子网，保证扩容后的各个可用区/子网实例数相对均衡。
	//
	// 与本策略相关的注意点：
	// <br><li> 当伸缩组为基础网络时，本策略适用于多可用区；当伸缩组为VPC网络时，本策略适用于多子网，此时不再考虑可用区因素，
	// 例如四个子网ABCD，其中ABC处于可用区1，D处于可用区2，此时考虑子网ABCD进行排序，而不考虑可用区1、2。
	// <br><li> 本策略适用于多可用区/子网，不适用于启动配置的多机型。多机型按照优先级策略进行选择。
	// <br><li> 按照 PRIORITY 策略创建实例时，先保证多机型的策略，后保证多可用区/子网的策略。例如多机型A、B，
	// 多子网1、2、3，会按照A1、A2、A3、B1、B2、B3 进行尝试，如果A1售罄，会尝试A2（而非B1）。
	MultiZoneSubnetPolicy *string `json:"MultiZoneSubnetPolicy,omitempty" name:"MultiZoneSubnetPolicy"`

	// 伸缩组实例健康检查类型
	HealthCheckType *string `json:"HealthCheckType,omitempty" name:"HealthCheckType"`

	// CLB健康检查宽限期，当扩容的实例进入`IN_SERVICE`后，在宽限期时间范围内将不会被标记为不健康`CLB_UNHEALTHY`。<br>默认值：0。
	// 取值范围[0, 7200]，单位：秒。
	LoadBalancerHealthCheckGracePeriod *uint64 `json:"LoadBalancerHealthCheckGracePeriod,omitempty" name:"LoadBalancerHealthCheckGracePeriod"` // nolint

	// 实例分配策略，取值包括 LAUNCH_CONFIGURATION 和 SPOT_MIXED，默认取 LAUNCH_CONFIGURATION。
	// <br><li> LAUNCH_CONFIGURATION，代表传统的按照启动配置模式。
	// <br><li> SPOT_MIXED，代表竞价混合模式。目前仅支持启动配置为按量计费模式时使用混合模式，混合模式下，伸缩组将根据设定扩容按量或竞价机型。
	// 使用混合模式时，关联的启动配置的计费类型不可被修改。
	InstanceAllocationPolicy *string `json:"InstanceAllocationPolicy,omitempty" name:"InstanceAllocationPolicy"`

	// 竞价混合模式下，各计费类型实例的分配策略。
	// 仅当 InstanceAllocationPolicy 取 SPOT_MIXED 时可用。
	SpotMixedAllocationPolicy *SpotMixedAllocationPolicy `json:"SpotMixedAllocationPolicy,omitempty" name:"SpotMixedAllocationPolicy"` // nolint

	// 容量重平衡功能，仅对伸缩组内的竞价实例有效。取值范围：
	// <br><li> TRUE，开启该功能，当伸缩组内的竞价实例即将被竞价实例服务自动回收前，AS 主动发起竞价实例销毁流程，如果有配置过缩容 hook，
	// 则销毁前 hook 会生效。销毁流程启动后，AS 会异步开启一个扩容活动，用于补齐期望实例数。
	// <br><li> FALSE，不开启该功能，则 AS 等待竞价实例被销毁后才会去扩容补齐伸缩组期望实例数。
	//
	// 默认取 FALSE。
	CapacityRebalance *bool `json:"CapacityRebalance,omitempty" name:"CapacityRebalance"`
}

// ForwardLoadBalancer 转发负载均衡器。
type ForwardLoadBalancer struct {

	// 负载均衡器ID
	LoadBalancerID *string `json:"LoadBalancerId,omitempty" name:"LoadBalancerId"`

	// 应用型负载均衡监听器 ID
	ListenerID *string `json:"ListenerId,omitempty" name:"ListenerId"`

	// 目标规则属性列表
	TargetAttributes []*TargetAttribute `json:"TargetAttributes,omitempty" name:"TargetAttributes"`

	// 转发规则ID，注意：针对七层监听器此参数必填
	LocationID *string `json:"LocationId,omitempty" name:"LocationId"`

	// 负载均衡实例所属地域，默认取AS服务所在地域。格式与公共参数Region相同，如："ap-guangzhou"。
	Region *string `json:"Region,omitempty" name:"Region"`
}

// TargetAttribute 目标规则属性
type TargetAttribute struct {

	// 端口
	Port *uint64 `json:"Port,omitempty" name:"Port"`

	// 权重
	Weight *uint64 `json:"Weight,omitempty" name:"Weight"`
}

// ServiceSettings 设置服务的相关参数。
type ServiceSettings struct {

	// 开启监控不健康替换服务。若开启则对于云监控标记为不健康的实例，弹性伸缩服务会进行替换。若不指定该参数，则默认为 False。
	ReplaceMonitorUnhealthy *bool `json:"ReplaceMonitorUnhealthy,omitempty" name:"ReplaceMonitorUnhealthy"`

	// 取值范围：
	// CLASSIC_SCALING：经典方式，使用创建、销毁实例来实现扩缩容；
	// WAKE_UP_STOPPED_SCALING：扩容优先开机。扩容时优先对已关机的实例执行开机操作，若开机后实例数仍低于期望实例数，则创建实例，
	// 缩容仍采用销毁实例的方式。用户可以使用StopAutoScalingInstances接口来关闭伸缩组内的实例。监控告警触发的扩容仍将创建实例
	// 默认取值：CLASSIC_SCALING
	ScalingMode *string `json:"ScalingMode,omitempty" name:"ScalingMode"`

	// 开启负载均衡不健康替换服务。若开启则对于负载均衡健康检查判断不健康的实例，弹性伸缩服务会进行替换。若不指定该参数，则默认为 False。
	ReplaceLoadBalancerUnhealthy *bool `json:"ReplaceLoadBalancerUnhealthy,omitempty" name:"ReplaceLoadBalancerUnhealthy"`
}

// SpotMixedAllocationPolicy 竞价混合模式下，各计费类型实例的分配策略。
type SpotMixedAllocationPolicy struct {

	// 混合模式下，基础容量的大小，基础容量部分固定为按量计费实例。默认值 0，最大不可超过伸缩组的最大实例数。
	// 注意：此字段可能返回 null，表示取不到有效值。
	BaseCapacity *uint64 `json:"BaseCapacity,omitempty" name:"BaseCapacity"`

	// 超出基础容量部分，按量计费实例所占的比例。取值范围 [0, 100]，0 代表超出基础容量的部分仅生产竞价实例，100 代表仅生产按量实例，
	// 默认值为 70。按百分比计算按量实例数时，向上取整。
	// 比如，总期望实例数取 3，基础容量取 1，超基础部分按量百分比取 1，则最终按量 2 台（1 台来自基础容量，1 台按百分比向上取整得到），
	// 竞价 1台。
	// 注意：此字段可能返回 null，表示取不到有效值。
	OnDemandPercentageAboveBaseCapacity *uint64 `json:"OnDemandPercentageAboveBaseCapacity,omitempty" name:"OnDemandPercentageAboveBaseCapacity"` // nolint

	// 混合模式下，竞价实例的分配策略。取值包括 COST_OPTIMIZED 和 CAPACITY_OPTIMIZED，默认取 COST_OPTIMIZED。
	// <br><li> COST_OPTIMIZED，成本优化策略。对于启动配置内的所有机型，按照各机型在各可用区的每核单价由小到大依次尝试。
	// 优先尝试购买每核单价最便宜的，如果购买失败则尝试购买次便宜的，以此类推。
	// <br><li> CAPACITY_OPTIMIZED，容量优化策略。对于启动配置内的所有机型，按照各机型在各可用区的库存情况由大到小依次尝试。
	// 优先尝试购买剩余库存最大的机型，这样可尽量降低竞价实例被动回收的发生概率。
	// 注意：此字段可能返回 null，表示取不到有效值。
	SpotAllocationStrategy *string `json:"SpotAllocationStrategy,omitempty" name:"SpotAllocationStrategy"`

	// 按量实例替补功能。取值范围：
	// <br><li> TRUE，开启该功能，当所有竞价机型因库存不足等原因全部购买失败后，尝试购买按量实例。
	// <br><li> FALSE，不开启该功能，伸缩组在需要扩容竞价实例时仅尝试所配置的竞价机型。
	//
	// 默认取值： TRUE。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CompensateWithBaseInstance *bool `json:"CompensateWithBaseInstance,omitempty" name:"CompensateWithBaseInstance"`
}

// LaunchConfiguration is the configuration of launch configuration.
type LaunchConfiguration struct {
	// 启动配置显示名称。名称仅支持中文、英文、数字、下划线、分隔符"-"、小数点，最大长度不能超60个字节。
	LaunchConfigurationName *string `json:"LaunchConfigurationName,omitempty" name:"LaunchConfigurationName"`

	// 指定有效的[镜像]
	ImageID *string `json:"ImageId,omitempty" name:"ImageId"`

	// 启动配置所属项目ID。不填为默认项目。
	// 注意：伸缩组内实例所属项目ID取伸缩组项目ID，与这里取值无关。
	ProjectID *uint64 `json:"ProjectId,omitempty" name:"ProjectId"`

	// 实例机型。不同实例机型指定了不同的资源规格
	InstanceType *string `json:"InstanceType,omitempty" name:"InstanceType"`

	// 实例系统盘配置信息。若不指定该参数，则按照系统默认值进行分配。
	SystemDisk *SystemDisk `json:"SystemDisk,omitempty" name:"SystemDisk"`

	// 实例数据盘配置信息。若不指定该参数，则默认不购买数据盘，最多支持指定11块数据盘。
	DataDisks []*LaunchConfigureDataDisk `json:"DataDisks,omitempty" name:"DataDisks"`

	// 公网带宽相关信息设置。若不指定该参数，则默认公网带宽为0Mbps。
	InternetAccessible *InternetAccessible `json:"InternetAccessible,omitempty" name:"InternetAccessible"`

	// 实例登录设置。通过该参数可以设置实例的登录方式密码、密钥或保持镜像的原始登录设置。默认情况下会随机生成密码，并以站内信方式知会到用户。
	LoginSettings *LoginSettings `json:"LoginSettings,omitempty" name:"LoginSettings"`

	// 实例所属安全组。该参数可以通过调用 [DescribeSecurityGroups](https://cloud.tencent.com/document/api/215/15808)
	// 的返回值中的`SecurityGroupId`字段来获取。若不指定该参数，则默认不绑定安全组。
	SecurityGroupIds []*string `json:"SecurityGroupIds,omitempty" name:"SecurityGroupIds"`

	// 增强服务。通过该参数可以指定是否开启云安全、云监控等服务。若不指定该参数，则默认开启云监控、云安全服务。
	EnhancedService *EnhancedService `json:"EnhancedService,omitempty" name:"EnhancedService"`

	// 经过 Base64 编码后的自定义数据，最大长度不超过16KB。
	UserData *string `json:"UserData,omitempty" name:"UserData"`

	// 实例计费类型，CVM默认值按照POSTPAID_BY_HOUR处理。
	// <br><li>POSTPAID_BY_HOUR：按小时后付费
	// <br><li>SPOTPAID：竞价付费
	// <br><li>PREPAID：预付费，即包年包月
	InstanceChargeType *string `json:"InstanceChargeType,omitempty" name:"InstanceChargeType"`

	// 实例的市场相关选项，如竞价实例相关参数，若指定实例的付费模式为竞价付费则该参数必传。
	InstanceMarketOptions *InstanceMarketOptionsRequest `json:"InstanceMarketOptions,omitempty" name:"InstanceMarketOptions"` // nolint

	// 实例机型列表，不同实例机型指定了不同的资源规格，最多支持10种实例机型。
	// `InstanceType`和`InstanceTypes`参数互斥，二者必填一个且只能填写一个。
	InstanceTypes []*string `json:"InstanceTypes,omitempty" name:"InstanceTypes"`

	// 实例类型校验策略，取值包括 ALL 和 ANY，默认取值为ANY。
	// <br><li> ALL，所有实例类型（InstanceType）都可用则通过校验，否则校验报错。
	// <br><li> ANY，存在任何一个实例类型（InstanceType）可用则通过校验，否则校验报错。
	//
	// 实例类型不可用的常见原因包括该实例类型售罄、对应云盘售罄等。
	// 如果 InstanceTypes 中一款机型不存在或者已下线，则无论 InstanceTypesCheckPolicy 采用何种取值，都会校验报错。
	InstanceTypesCheckPolicy *string `json:"InstanceTypesCheckPolicy,omitempty" name:"InstanceTypesCheckPolicy"`

	// 标签列表。通过指定该参数，可以为扩容的实例绑定标签。最多支持指定10个标签。
	InstanceTags []*InstanceTag `json:"InstanceTags,omitempty" name:"InstanceTags"`

	// CAM角色名称。可通过DescribeRoleList接口返回值中的roleName获取。
	CamRoleName *string `json:"CamRoleName,omitempty" name:"CamRoleName"`

	// 云服务器主机名（HostName）的相关设置。
	HostNameSettings *HostNameSettings `json:"HostNameSettings,omitempty" name:"HostNameSettings"`

	// 云服务器实例名（InstanceName）的相关设置。
	// 如果用户在启动配置中设置此字段，则伸缩组创建出的实例 InstanceName 参照此字段进行设置，并传递给 CVM；如果用户未在启动配置中设置此字段，
	// 则伸缩组创建出的实例 InstanceName 按照“as-{{ 伸缩组AutoScalingGroupName }}”进行设置，并传递给 CVM。
	InstanceNameSettings *InstanceNameSettings `json:"InstanceNameSettings,omitempty" name:"InstanceNameSettings"`

	// 预付费模式，即包年包月相关参数设置。通过该参数可以指定包年包月实例的购买时长、是否设置自动续费等属性。
	// 若指定实例的付费模式为预付费则该参数必传。
	InstanceChargePrepaid *InstanceChargePrepaid `json:"InstanceChargePrepaid,omitempty" name:"InstanceChargePrepaid"`

	// 云盘类型选择策略，默认取值 ORIGINAL，取值范围：
	// <br><li>ORIGINAL：使用设置的云盘类型
	// <br><li>AUTOMATIC：自动选择当前可用的云盘类型
	DiskTypePolicy *string `json:"DiskTypePolicy,omitempty" name:"DiskTypePolicy"`
}

// SystemDisk 系统盘配置信息。
type SystemDisk struct {

	// 系统盘类型
	DiskType *string `json:"DiskType,omitempty" name:"DiskType"`

	// 系统盘大小，单位：GB。默认值为 50
	// 注意：此字段可能返回 null，表示取不到有效值。
	DiskSize *uint64 `json:"DiskSize,omitempty" name:"DiskSize"`
}

// LaunchConfigureDataDisk 数据盘配置信息。
type LaunchConfigureDataDisk struct {

	// 数据盘类型
	DiskType *string `json:"DiskType,omitempty" name:"DiskType"`

	// 数据盘大小，单位：GB。最小调整步长为10G，不同数据盘类型取值范围不同
	DiskSize *uint64 `json:"DiskSize,omitempty" name:"DiskSize"`
}

// InternetAccessible 网络计费类型。
type InternetAccessible struct {

	// 网络计费类型
	InternetChargeType *string `json:"InternetChargeType,omitempty" name:"InternetChargeType"`

	// 公网出带宽上限，单位：Mbps。默认值：0Mbps。不同机型带宽上限范围不一致，
	// 具体限制详见[购买网络带宽](https://cloud.tencent.com/document/product/213/509)。
	// 注意：此字段可能返回 null，表示取不到有效值。
	InternetMaxBandwidthOut *uint64 `json:"InternetMaxBandwidthOut,omitempty" name:"InternetMaxBandwidthOut"`

	// 是否分配公网IP。取值范围：<br><li>TRUE：表示分配公网IP<br><li>FALSE：表示不分配公网IP<br><br>当公网带宽大于0Mbps时，
	// 可自由选择开通与否，默认开通公网IP；当公网带宽为0，则不允许分配公网IP。
	// 注意：此字段可能返回 null，表示取不到有效值。
	PublicIPAssigned *bool `json:"PublicIpAssigned,omitempty" name:"PublicIpAssigned"`

	// 带宽包ID。可通过[DescribeBandwidthPackages]
	// (https://cloud.tencent.com/document/api/215/19209)接口返回值中的`BandwidthPackageId`获取。
	// 注意：此字段可能返回 null，表示取不到有效值。
	BandwidthPackageID *string `json:"BandwidthPackageId,omitempty" name:"BandwidthPackageId"`
}

// InstanceTag 实例标签
type InstanceTag struct {

	// 标签键
	Key *string `json:"Key,omitempty" name:"Key"`

	// 标签值
	Value *string `json:"Value,omitempty" name:"Value"`
}

// HostNameSettings 主机名配置。
type HostNameSettings struct {

	// 云服务器的主机名。
	// <br><li> 点号（.）和短横线（-）不能作为 HostName 的首尾字符，不能连续使用。
	// <br><li> 不支持 Windows 实例。
	// <br><li> 其他类型（Linux 等）实例：字符长度为[2, 40]，允许支持多个点号，点之间为一段，每段允许字母（不限制大小写）、数字和短横线（-）组成。不允许为纯数字。
	// 注意：此字段可能返回 null，表示取不到有效值。
	HostName *string `json:"HostName,omitempty" name:"HostName"`

	// 云服务器主机名的风格，取值范围包括 ORIGINAL 和  UNIQUE，默认为 ORIGINAL。
	// <br><li> ORIGINAL，AS 直接将入参中所填的 HostName 传递给 CVM，CVM 可能会对 HostName 追加序列号，伸缩组中实例的 HostName 会出现冲突的情况。
	// <br><li> UNIQUE，入参所填的 HostName 相当于主机名前缀，AS 和 CVM 会对其进行拓展，伸缩组中实例的 HostName 可以保证唯一。
	// 注意：此字段可能返回 null，表示取不到有效值。
	HostNameStyle *string `json:"HostNameStyle,omitempty" name:"HostNameStyle"`
}

// InstanceNameSettings 实例名称配置。
type InstanceNameSettings struct {

	// 云服务器的实例名。
	//
	// 点号（.）和短横线（-）不能作为 InstanceName 的首尾字符，不能连续使用。
	// 字符长度为[2, 40]，允许支持多个点号，点之间为一段，每段允许字母（不限制大小写）、数字和短横线（-）组成。不允许为纯数字。
	InstanceName *string `json:"InstanceName,omitempty" name:"InstanceName"`

	// 云服务器实例名的风格，取值范围包括 ORIGINAL 和 UNIQUE，默认为 ORIGINAL。
	//
	// ORIGINAL，AS 直接将入参中所填的 InstanceName 传递给 CVM，CVM 可能会对 InstanceName 追加序列号，伸缩组中实例的 InstanceName 会出现冲突的情况。
	//
	// UNIQUE，入参所填的 InstanceName 相当于实例名前缀，AS 和 CVM 会对其进行拓展，伸缩组中实例的 InstanceName 可以保证唯一。
	InstanceNameStyle *string `json:"InstanceNameStyle,omitempty" name:"InstanceNameStyle"`
}

// InstanceChargePrepaid 计费配置。
type InstanceChargePrepaid struct {

	// 购买实例的时长，单位：月。取值范围：1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36。
	Period *int64 `json:"Period,omitempty" name:"Period"`

	// 自动续费标识
	RenewFlag *string `json:"RenewFlag,omitempty" name:"RenewFlag"`
}

// InstanceMarketOptionsRequest 设置市场竞价相关选项
type InstanceMarketOptionsRequest struct {

	// 竞价相关选项
	SpotOptions *SpotMarketOptions `json:"SpotOptions,omitempty" name:"SpotOptions"`

	// 市场选项类型，当前只支持取值：spot
	// 注意：此字段可能返回 null，表示取不到有效值。
	MarketType *string `json:"MarketType,omitempty" name:"MarketType"`
}

// SpotMarketOptions 市场竞价相关选项
type SpotMarketOptions struct {

	// 竞价出价，例如“1.05”
	MaxPrice *string `json:"MaxPrice,omitempty" name:"MaxPrice"`

	// 竞价请求类型，当前仅支持类型：one-time，默认值为one-time
	// 注意：此字段可能返回 null，表示取不到有效值。
	SpotInstanceType *string `json:"SpotInstanceType,omitempty" name:"SpotInstanceType"`
}

// ModifyClusterNodePoolInput is a struct for modify cluster node pool
type ModifyClusterNodePoolInput struct {
	// 集群ID
	ClusterID *string `json:"ClusterId,omitempty" name:"ClusterId"`

	// 节点池ID
	NodePoolID *string `json:"NodePoolId,omitempty" name:"NodePoolId"`

	// 名称
	Name *string `json:"Name,omitempty" name:"Name"`

	// 最大节点数
	MaxNodesNum *int64 `json:"MaxNodesNum,omitempty" name:"MaxNodesNum"`

	// 最小节点数
	MinNodesNum *int64 `json:"MinNodesNum,omitempty" name:"MinNodesNum"`

	// 标签
	Labels []*Label `json:"Labels,omitempty" name:"Labels"`

	// 污点
	Taints []*Taint `json:"Taints,omitempty" name:"Taints"`

	// 是否开启伸缩
	EnableAutoscale *bool `json:"EnableAutoscale,omitempty" name:"EnableAutoscale"`

	// 操作系统名称
	OsName *string `json:"OsName,omitempty" name:"OsName"`

	// 镜像版本，"DOCKER_CUSTOMIZE"(容器定制版),"GENERAL"(普通版本，默认值)
	OsCustomizeType *string `json:"OsCustomizeType,omitempty" name:"OsCustomizeType"`

	// 节点自定义参数
	ExtraArgs *InstanceExtraArgs `json:"ExtraArgs,omitempty" name:"ExtraArgs"`

	// 资源标签
	Tags []*Tag `json:"Tags,omitempty" name:"Tags"`

	// 设置加入的节点是否参与调度，默认值为0，表示参与调度；非0表示不参与调度, 待节点初始化完成之后, 可执行kubectl uncordon nodename使node加入调度.
	Unschedulable *int64 `json:"Unschedulable,omitempty" name:"Unschedulable"`
}

// Filter is a struct for filter
type Filter struct {

	// 需要过滤的字段。
	Name string `json:"Name,omitempty" name:"Name"`

	// 字段的过滤值。
	Values []string `json:"Values,omitempty" name:"Values"`
}

// InstanceTypeConfig 实例机型
type InstanceTypeConfig struct {

	// 可用区。
	Zone *string `json:"Zone,omitempty" name:"Zone"`

	// 实例机型。
	InstanceType *string `json:"InstanceType,omitempty" name:"InstanceType"`

	// 实例机型系列。
	InstanceFamily *string `json:"InstanceFamily,omitempty" name:"InstanceFamily"`

	// GPU核数，单位：核。
	GPU *int64 `json:"GPU,omitempty" name:"GPU"`

	// CPU核数，单位：核。
	CPU *int64 `json:"CPU,omitempty" name:"CPU"`

	// 内存容量，单位：`GB`。
	Memory *int64 `json:"Memory,omitempty" name:"Memory"`

	// FPGA核数，单位：核。
	FPGA *int64 `json:"FPGA,omitempty" name:"FPGA"`
}

// Subnet 子网
type Subnet struct {

	// `VPC`实例`ID`。
	VpcID *string `json:"VpcId,omitempty" name:"VpcId"`

	// 子网实例`ID`，例如：subnet-bthucmmy。
	SubnetID *string `json:"SubnetId,omitempty" name:"SubnetId"`

	// 子网名称。
	SubnetName *string `json:"SubnetName,omitempty" name:"SubnetName"`

	// 子网的 `IPv4` `CIDR`。
	CidrBlock *string `json:"CidrBlock,omitempty" name:"CidrBlock"`

	// 是否默认子网。
	IsDefault *bool `json:"IsDefault,omitempty" name:"IsDefault"`

	// 是否开启广播。
	EnableBroadcast *bool `json:"EnableBroadcast,omitempty" name:"EnableBroadcast"`

	// 可用区。
	Zone *string `json:"Zone,omitempty" name:"Zone"`

	// 路由表实例ID，例如：rtb-l2h8d7c2。
	RouteTableID *string `json:"RouteTableId,omitempty" name:"RouteTableId"`

	// 创建时间。
	CreatedTime *string `json:"CreatedTime,omitempty" name:"CreatedTime"`

	// 可用`IPv4`数。
	AvailableIPAddressCount *uint64 `json:"AvailableIpAddressCount,omitempty" name:"AvailableIpAddressCount"`

	// 子网的 `IPv6` `CIDR`。
	Ipv6CidrBlock *string `json:"Ipv6CidrBlock,omitempty" name:"Ipv6CidrBlock"`

	// 关联`ACL`ID
	NetworkACLID *string `json:"NetworkAclId,omitempty" name:"NetworkAclId"`

	// 是否为 `SNAT` 地址池子网。
	IsRemoteVpcSnat *bool `json:"IsRemoteVpcSnat,omitempty" name:"IsRemoteVpcSnat"`

	// 子网`IPv4`总数。
	TotalIPAddressCount *uint64 `json:"TotalIpAddressCount,omitempty" name:"TotalIpAddressCount"`

	// 标签键值对。
	TagSet []*Tag `json:"TagSet,omitempty" name:"TagSet"`

	// CDC实例ID。
	// 注意：此字段可能返回 null，表示取不到有效值。
	CdcID *string `json:"CdcId,omitempty" name:"CdcId"`

	// 是否是CDC所属子网。0:否 1:是
	// 注意：此字段可能返回 null，表示取不到有效值。
	IsCdcSubnet *int64 `json:"IsCdcSubnet,omitempty" name:"IsCdcSubnet"`
}

// AutoScalingInstances 伸缩组实例信息
type AutoScalingInstances struct {

	// 实例ID
	InstanceID *string `json:"InstanceId,omitempty" name:"InstanceId"`

	// 伸缩组ID
	AutoScalingGroupID *string `json:"AutoScalingGroupId,omitempty" name:"AutoScalingGroupId"`

	// 启动配置ID
	LaunchConfigurationID *string `json:"LaunchConfigurationId,omitempty" name:"LaunchConfigurationId"`

	// 启动配置名称
	LaunchConfigurationName *string `json:"LaunchConfigurationName,omitempty" name:"LaunchConfigurationName"`

	// 生命周期状态，取值如下：<br>
	// <li>IN_SERVICE：运行中
	// <li>CREATING：创建中
	// <li>CREATION_FAILED：创建失败
	// <li>TERMINATING：中止中
	// <li>TERMINATION_FAILED：中止失败
	// <li>ATTACHING：绑定中
	// <li>DETACHING：解绑中
	// <li>ATTACHING_LB：绑定LB中<li>DETACHING_LB：解绑LB中
	// <li>STARTING：开机中
	// <li>START_FAILED：开机失败
	// <li>STOPPING：关机中
	// <li>STOP_FAILED：关机失败
	// <li>STOPPED：已关机
	LifeCycleState *string `json:"LifeCycleState,omitempty" name:"LifeCycleState"`

	// 健康状态，取值包括HEALTHY和UNHEALTHY
	HealthStatus *string `json:"HealthStatus,omitempty" name:"HealthStatus"`

	// 是否加入缩容保护
	ProtectedFromScaleIn *bool `json:"ProtectedFromScaleIn,omitempty" name:"ProtectedFromScaleIn"`

	// 可用区
	Zone *string `json:"Zone,omitempty" name:"Zone"`

	// 创建类型，取值包括AUTO_CREATION, MANUAL_ATTACHING。
	CreationType *string `json:"CreationType,omitempty" name:"CreationType"`

	// 实例加入时间
	AddTime *string `json:"AddTime,omitempty" name:"AddTime"`

	// 实例类型
	InstanceType *string `json:"InstanceType,omitempty" name:"InstanceType"`

	// 版本号
	VersionNumber *int64 `json:"VersionNumber,omitempty" name:"VersionNumber"`

	// 伸缩组名称
	AutoScalingGroupName *string `json:"AutoScalingGroupName,omitempty" name:"AutoScalingGroupName"`
}

// NetWorkType xxx
type NetWorkType string

var (
	// Flannel xxx
	Flannel NetWorkType = "Flannel"
	// CiliumBGP xxx
	CiliumBGP NetWorkType = "CiliumBGP"
	// CiliumVXLan xxx
	CiliumVXLan NetWorkType = "CiliumVXLan"
)

var netWorkTypeMap = map[NetWorkType]struct{}{
	Flannel:     {},
	CiliumBGP:   {},
	CiliumVXLan: {},
}

// EnableExternalNodeConfig enable externalNode config
type EnableExternalNodeConfig struct {
	// NetworkType 集群网络插件类型，支持：Flannel、CiliumBGP、CiliumVXLan
	NetworkType string `json:"networkType,omitempty"`
	// ClusterCIDR Pod CIDR
	ClusterCIDR string `json:"clusterCIDR,omitempty"`
	// SubnetId 子网ID
	SubnetId string `json:"subnetId,omitempty"`
	// 是否开启第三方节点池支持
	Enabled bool `json:"enabled,omitempty"`
}

func (cfg EnableExternalNodeConfig) validate() error {
	if _, ok := netWorkTypeMap[NetWorkType(cfg.NetworkType)]; !ok {
		return fmt.Errorf("EnableExternalNodeConfig not support networkType[%s]", cfg.NetworkType)
	}

	if cfg.ClusterCIDR == "" || cfg.SubnetId == "" {
		return fmt.Errorf("EnableExternalNodeConfig clusterCIDR&SubnetID not empty")
	}

	return nil
}

// DescribeExternalNodeScriptConfig xxx
type DescribeExternalNodeScriptConfig struct {
	// 节点池ID
	NodePoolId string `json:"NodePoolId,omitempty"`
	// 网卡名
	Interface string `json:"Interface,omitempty"`
	// 节点名称
	Name string `json:"Name,omitempty"`
	// 内外网脚本
	Internal bool `json:"Internal,omitempty"`
}

func (cfg DescribeExternalNodeScriptConfig) validate() error {
	if cfg.NodePoolId == "" {
		return fmt.Errorf("DescribeExternalNodeScriptConfig NodePoolId is empty")
	}

	return nil
}

// DeleteExternalNodeConfig xxx
type DeleteExternalNodeConfig struct {
	// Names 第三方节点列表
	Names []string `json:"names,omitempty"`
	// Force 是否强制删除：如果第三方节点上有运行中Pod，则非强制删除状态下不会进行删除
	Force bool `json:"force,omitempty"`
}

func (cfg DeleteExternalNodeConfig) validate() error {
	if len(cfg.Names) == 0 {
		return fmt.Errorf("DeleteExternalNodeConfig Names is empty")
	}

	return nil
}

// DeleteExternalNodePoolConfig xxx
type DeleteExternalNodePoolConfig struct {
	// NodePoolIds 第三方节点池ID列表
	NodePoolIds []string `json:"nodePoolIds,omitempty"`
	// Force 是否强制删除，在第三方节点上有pod的情况下，如果选择非强制删除，则删除会失败
	Force bool `json:"force,omitempty"`
}

func (cfg DeleteExternalNodePoolConfig) validate() error {
	if len(cfg.NodePoolIds) == 0 {
		return fmt.Errorf("DeleteExternalNodePoolConfig NodePoolIds is empty")
	}

	return nil
}

// DescribeExternalNodeConfigInfoResponse xxx
type DescribeExternalNodeConfigInfoResponse struct {
	// ClusterCIDR 用于分配集群容器和服务 IP 的 CIDR，不得与 VPC CIDR 冲突，也不得与同 VPC 内其他集群 CIDR 冲突。
	// 且网段范围必须在内网网段内，例如:10.1.0.0/14, 192.168.0.1/18,172.16.0.0/16
	ClusterCIDR string `json:"clusterCIDR,omitempty"`
	// NetworkType 集群网络插件类型，支持：CiliumBGP、CiliumVXLan
	NetworkType string `json:"networkType,omitempty"`
	// SubnetId 子网ID
	SubnetId string `json:"subnetId,omitempty"`
	// Enabled 是否开启第三方节点支持
	Enabled bool `json:"enabled,omitempty"`
	// AS 节点所属交换机的BGP AS 号
	AS string `json:"aS,omitempty"`
	// SwitchIP 节点所属交换机的交换机 IP
	SwitchIP string `json:"switchIP,omitempty"`
	// Status 开启第三方节电池状态
	Status string `json:"status,omitempty"`
	// FailedReason 如果开启失败原因
	FailedReason string `json:"failedReason,omitempty"`
	// Master 内网访问地址
	Master string `json:"master,omitempty"`
	// Proxy 镜像仓库代理地址
	Proxy string `json:"Proxy,omitempty"`
}

// CreateExternalNodePoolConfig xxx
type CreateExternalNodePoolConfig struct {
	// Name 节点池名称
	Name string `json:"name,omitempty"`
	// ContainerRuntime 运行时
	ContainerRuntime string `json:"containerRuntime,omitempty"`
	// RuntimeVersion 运行时版本
	RuntimeVersion string `json:"runtimeVersion,omitempty"`
	// Labels 第三方节点label
	Labels []*Label `json:"labels,omitempty"`
	// 第三方节点taint
	Taints []*Taint `json:"taints,omitempty"`
	// 第三方节点高级设置
	InstanceAdvancedSettings *InstanceAdvancedSettings `json:"instanceAdvancedSettings,omitempty"`
}

func (cfg CreateExternalNodePoolConfig) validate() error {
	if cfg.Name == "" || cfg.ContainerRuntime == "" || cfg.RuntimeVersion == "" {
		return fmt.Errorf("CreateExternalNodePoolConfig must paras empty")
	}

	return nil
}

func (cfg CreateExternalNodePoolConfig) transToTkeExternalNodeConfig(clusterID string) *CreateExternalNodePoolRequest {
	req := NewCreateExternalNodePoolRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.Name = common.StringPtr(cfg.Name)
	req.ContainerRuntime = common.StringPtr(cfg.ContainerRuntime)
	req.RuntimeVersion = common.StringPtr(cfg.RuntimeVersion)

	if len(cfg.Labels) > 0 {
		req.Labels = cfg.Labels
	}
	if len(cfg.Taints) > 0 {
		req.Taints = cfg.Taints
	}

	req.InstanceAdvancedSettings = generateInstanceAdvancedSetting(cfg.InstanceAdvancedSettings)

	return req
}

// ModifyExternalNodePoolConfig xxx
type ModifyExternalNodePoolConfig struct {
	// NodePoolId 节点池ID
	NodePoolId string `json:"nodePoolId,omitempty"`
	// Name 节点池名称
	Name string `json:"name,omitempty"`
	// Labels 第三方节点label
	Labels []*Label `json:"Labels,omitempty"`
	// 第三方节点taint
	Taints []*Taint `json:"Taints,omitempty"`
}

func (cfg ModifyExternalNodePoolConfig) validate() error {
	if cfg.NodePoolId == "" {
		return fmt.Errorf("ModifyExternalNodePoolConfig NodePoolId empty")
	}

	return nil
}

// DescribeExternalNodeConfig xxx
type DescribeExternalNodeConfig struct {
	// NodePoolId 节点池ID
	NodePoolId string `json:"nodePoolId,omitempty"`
	// Names 节点名称
	Names []string `json:"names,omitempty"`
}

func (cfg DescribeExternalNodeConfig) validate() error {
	if cfg.NodePoolId == "" {
		return fmt.Errorf("DescribeExternalNodeConfig NodePoolId empty")
	}

	return nil
}

// ExternalNodeInfo 第三方节点信息
type ExternalNodeInfo struct {
	// Name 第三方节点名称
	Name string `json:"Name,omitempty"`
	// NodePoolId 第三方节点所属节点池
	NodePoolId string `json:"NodePoolId,omitempty"`
	// IP 第三方IP地址
	IP string `json:"ip,omitempty"`
	// Location 第三方地域
	Location string `json:"location,omitempty"`
	// Status 第三方节点状态
	Status string `json:"Status,omitempty"`
	// CreatedTime 创建时间
	CreatedTime string `json:"createdTime,omitempty"`
	// Reason 异常原因
	Reason string `json:"reason,omitempty"`
	// Unschedulable 是否封锁。true表示已封锁，false表示未封锁
	Unschedulable bool `json:"unschedulable,omitempty"`
}

// DescribeVpcCniPodLimitsOut xxx
type DescribeVpcCniPodLimitsOut struct {
	Zone         string
	InstanceType string
	Limits       PodLimits
}

// PodLimits 某机型可支持的最大 VPC-CNI 模式的 Pod 数量
type PodLimits struct {
	RouterEniNonStaticIP int64
	RouterEniStaticIP    int64
	directEni            int64
}

// ClusterEndpointConfig xxx
type ClusterEndpointConfig struct {
	// 是否为外网访问（TRUE 外网访问 FALSE 内网访问，默认值： FALSE）
	IsExtranet bool
	// 集群端口所在的子网ID  (仅在开启非外网访问时需要填，必须为集群所在VPC内的子网)
	SubnetId string
	// 设置域名
	Domain string
	// 使用的安全组，只有外网访问需要传递（开启外网访问时必传）
	SecurityGroup string
	// 创建lb参数，只有外网访问需要设置
	ExtensiveParameters string
}

func (cef ClusterEndpointConfig) getEndpointConfig(clusterID string) *tke.CreateClusterEndpointRequest {
	req := tke.NewCreateClusterEndpointRequest()
	req.ClusterId = common.StringPtr(clusterID)
	req.IsExtranet = common.BoolPtr(cef.IsExtranet)

	if cef.IsExtranet {
		req.SecurityGroup = common.StringPtr(cef.SecurityGroup)
		req.ExtensiveParameters = common.StringPtr(cef.ExtensiveParameters)

		return req
	}

	req.SubnetId = common.StringPtr(cef.SubnetId)
	req.Domain = common.StringPtr(cef.Domain)
	return req
}

// ClusterEndpointInfo endpointInfo
type ClusterEndpointInfo struct {
	// 集群APIServer的CA证书
	CertClusterAuthority string
	// 集群APIServer的外网访问地址
	ClusterExternalEndpoint string
	// 集群APIServer的内网访问地址
	ClusterIntranetEndpoint string
	// 外网域名
	ClusterExternalDomain string
	// 内网域名
	ClusterIntranetDomain string
	// 集群APIServer的域名
	ClusterDomain string
	// 外网安全组
	SecurityGroup string
	// 集群APIServer的外网访问ACL列表
	ClusterExternalACL []string
}

// ZoneInfo zone info
type ZoneInfo struct {
	ZoneID    string
	Zone      string
	ZoneName  string
	ZoneState string
}

// GPUArgs gpu 参数
type GPUArgs struct {
	// 是否启用MIG特性
	MIGEnable bool `json:"MIGEnable,omitempty" name:"MIGEnable"`
	// GPU驱动版本信息
	Driver *DriverVersion `json:"Driver,omitempty" name:"Driver"`
	// CUDA版本信息
	CUDA *DriverVersion `json:"CUDA,omitempty" name:"CUDA"`
	// cuDNN版本信息
	CUDNN *CUDNN `json:"CUDNN,omitempty" name:"CUDNN"`
	// 自定义GPU驱动信息
	CustomDriver *CustomDriver `json:"CustomDriver,omitempty" name:"CustomDriver"`
}

// DriverVersion driver version
type DriverVersion struct {
	// GPU驱动或者CUDA的版本
	Version string `json:"Version,omitempty" name:"Version"`
	// GPU驱动或者CUDA的名字
	Name string `json:"Name,omitempty" name:"Name"`
}

// CUDNN cudnn
type CUDNN struct {
	// cuDNN的版本
	Version string `json:"Version,omitempty" name:"Version"`
	// cuDNN的名字
	Name string `json:"Name,omitempty" name:"Name"`
	// cuDNN的Doc名字
	DocName string `json:"DocName,omitempty" name:"DocName"`
	// cuDNN的Dev名字
	DevName string `json:"DevName,omitempty" name:"DevName"`
}

// CustomDriver custom driver
type CustomDriver struct {
	// 自定义GPU驱动地址链接
	Address string `json:"Address,omitempty" name:"Address"`
}
