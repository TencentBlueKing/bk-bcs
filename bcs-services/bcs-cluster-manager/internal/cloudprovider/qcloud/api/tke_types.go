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

package api

import "fmt"

// EndpointStatus endpoint status
type EndpointStatus string

// Created status
func (es EndpointStatus) Created() bool {
	return es == Created
}

// Creating status
func (es EndpointStatus) Creating() bool {
	return es == Creating
}

// NotFound status
func (es EndpointStatus) NotFound() bool {
	return es == NotFound
}

var (
	// Created status
	Created EndpointStatus = "Created"
	// Creating status
	Creating EndpointStatus = "Creating"
	// NotFound status
	NotFound EndpointStatus = "NotFound"
)

const (
	// DockerGraphPath default docker graphPath
	DockerGraphPath = "/data/bcs/service/docker"
	// MountTarget default mountTarget
	MountTarget = "/data"
)

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
)

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
	ForceDelete bool       `json:"forceDelete, omitempty"`
}

// DeleteInstancesResult xxx
type DeleteInstancesResult struct {
	Success  []string `json:"success"`
	Failure  []string `json:"failure"`
	NotFound []string `json:"notFound"`
}

func (dir *DeleteInstancesRequest) validateDeleteClusterInstanceRequest() error {
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

// AddExistedInstanceReq xxx
type AddExistedInstanceReq struct {
	ClusterID       string                    `json:"clusterID"`
	InstanceIDs     []string                  `json:"instanceIDs"`
	AdvancedSetting *InstanceAdvancedSettings `json:"advancedSetting"`
	// SecurityGroupIds instance security group set, only support single group; if null, use default group
	SecurityGroupIds []string `json:"securityGroupIds"`
	// NodePool nodePool conf
	NodePool        *NodePoolOption  `json:"nodePool"`
	EnhancedSetting *EnhancedService `json:"enhancedSetting"`
	LoginSetting    *LoginSettings   `json:"loginSetting"`
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

// InstanceAdvancedSettings instance advanced setting
type InstanceAdvancedSettings struct {
	// MountTarget data disk mountPoint
	MountTarget string `json:"mountTarget"`
	// DockerGraphPath dockerd --graph
	DockerGraphPath string `json:"dockerGraphPath"`
	// Unschedulable involved scheduler
	Unschedulable *int64 `json:"unschedulable"`
	// Labels instance labels
	Labels []*KeyValue `json:"labels"`
	// DataDisks many disk mount info
	DataDisks []DataDetailDisk `json:"dataDisks"`
	// ExtraArgs component start parameter
	ExtraArgs *InstanceExtraArgs `json:"extraArgs"`
	// UserScript  base64 编码的用户脚本, 此脚本会在 k8s 组件运行后执行, 需要用户保证脚本的可重入及重试逻辑
	UserScript string `json:"userScript"`
}

// KeyValue struct(name/value)
type KeyValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// InstanceExtraArgs kubelet startup parameter
type InstanceExtraArgs struct {
	// Kubelet user-defined parameter，["k1=v1", "k1=v2"]
	// for example: ["root-dir=/var/lib/kubelet","feature-gates=PodShareProcessNamespace=true,DynamicKubeletConfig=true"]
	Kubelet []string `json:"kubelet"`
}

// DataDetailDisk data disk
type DataDetailDisk struct {
	// DiskType type
	DiskType string `json:"diskType"`
	// DiskSize size
	DiskSize int64 `json:"diskSize"`
	// MountTarget mount point
	MountTarget string `json:"mountTarget"`
	// FileSystem file system
	FileSystem string `json:"fileSystem"`
	// AutoFormatAndMount auto format and mount
	AutoFormatAndMount bool `json:"autoFormatAndMount"`
}

// EnhancedService instance enhanced service
type EnhancedService struct {
	// SecurityService cloud security
	SecurityService bool `json:"securityService"`
	// MonitorService cloud monitor
	MonitorService bool `json:"monitorService"`
}

// LoginSettings reset passwd
type LoginSettings struct {
	// Password reset instance passwd
	Password string `json:"password"`
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
	//tke clusterId, required
	ClusterID string `json:"clusterId"`
	//VPCID, required
	VPCID string `json:"vpcId"`
	//subnet, required
	SubNetID string `json:"subnetId"`
	//available zone, required
	Zone string `json:"zone"`
	//cvm number, required
	ApplyNum uint32 `json:"applyNum"`
	// cvm instance type, required
	InstanceType string `json:"instanceType"`
	// required
	SystemDiskType string `json:"systemDiskType"`
	// required
	SystemDiskSize uint32 `json:"systemDiskSize"`
	// dataDisk, optional
	DataDisks []*DataDisk `json:"dataDisk"`
	//image information for system, required
	Image *ImageInfo `json:"image"`
	//security group, optional
	Security *SecurityGroup `json:"security"`
	//setup security service, optional, default 0
	IsSecurityService uint32 `json:"isSecurityService,omitempty"`
	//cloud monitor, optional, default 1
	IsMonitorService uint32 `json:"isMonitorService"`
	//cvm instance name, optional
	InstanceName string `json:"instanceName,omitempty"`
	// instance login setting
	Login LoginSettings `json:"login"`
	// required
	Operator string `json:"operator"`
}

// ImageInfo for system
type ImageInfo struct {
	ID   string `json:"imageId"`           //required
	Name string `json:"imageName"`         //required
	OS   string `json:"imageOs,omitempty"` //optional
	Type string `json:"imageType"`         //optional
}

//SecurityGroup sg information
type SecurityGroup struct {
	ID   string `json:"securityGroupId"`             //required
	Name string `json:"securityGroupName,omitempty"` //optional
	Desc string `json:"securityGroupDesc,omitempty"` //optional
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
	// DeletionProtection cluster delete protection
	DeletionProtection bool `json:"deletionProtection"`
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
	// CVM创建透传参数，json化字符串格式，详见[CVM创建实例](https://cloud.tencent.com/document/product/213/15730)接口，传入公共参数外的其他参数即可，其中ImageId会替换为TKE集群OS对应的镜像。
	RunInstancesPara []*string `json:"runInstancesPara"`
	// InstanceAdvancedSettingsOverrides node advanced setting(上边的RunInstancesPara按照顺序一一对应（当前只对节点自定义参数ExtraArgs生效）
	InstanceAdvancedSettingsOverrides []*InstanceAdvancedSettings `json:"InstanceAdvancedSettingsOverrides,omitempty" name:"InstanceAdvancedSettingsOverrides"`
}

// ClusterExtraArgs cluster extra args
type ClusterExtraArgs struct {
	// KubeAPIServer xxx
	// kube-apiserver自定义参数，参数格式为["k1=v1", "k1=v2"]， 例如["max-requests-inflight=500","feature-gates=PodShareProcessNamespace=true,DynamicKubeletConfig=true"]
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
	InstanceID    string
	InstanceIP    string
	InstanceRole  string
	InstanceState string
}

// EnableVpcCniInput xxx
type EnableVpcCniInput struct {
	TkeClusterID string
	// 开启vpc-cni的模式，tke-route-eni开启的是策略路由模式，tke-direct-eni开启的是独立网卡模式
	VpcCniType string
	SubnetsIDs []string

	EnableStaticIP bool
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
