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

// Package common xxx
package common

import (
	bcscommon "github.com/Tencent/bk-bcs/bcs-common/common"
)

// ResourceType resource type
type ResourceType string

// String xxx
func (rt ResourceType) String() string {
	return string(rt)
}

var (
	// Cluster type
	Cluster ResourceType = "cluster"
	// AutoScalingOption type
	AutoScalingOption ResourceType = "autoscalingoption"
	// Cloud type
	Cloud ResourceType = "cloud"
	// CloudVPC type
	CloudVPC ResourceType = "cloudvpc"
	// ClusterCredential type
	ClusterCredential ResourceType = "clustercredential"
	// NameSpace type
	NameSpace ResourceType = "namespace"
	// NameSpaceQuota type
	NameSpaceQuota ResourceType = "namespacequota"
	// NodeGroup type
	NodeGroup ResourceType = "nodegroup"
	// Project type
	Project ResourceType = "project"
	// Task type
	Task ResourceType = "task"
)

// NodeType node type
type NodeType string

// String xxx
func (nt NodeType) String() string {
	return string(nt)
}

var (
	// CVM cloud instance
	CVM NodeType = "CVM"
	// IDC instance
	IDC NodeType = "IDC"
)

// NodeGroupType group type
type NodeGroupType string

// String xxx
func (nt NodeGroupType) String() string {
	return string(nt)
}

// NodeGroupTypeMap nodePool type
var NodeGroupTypeMap = map[NodeGroupType]struct{}{
	Normal:   {},
	External: {},
}

var (
	// Normal 普通云实例节点池
	Normal NodeGroupType = "normal"
	// External 第三方节点池
	External NodeGroupType = "external"
)

// ScriptType external nodeGroup script type
type ScriptType string

// String xxx
func (st ScriptType) String() string {
	return string(st)
}

// IsInterType inter
func (st ScriptType) IsInterType() bool {
	return st == ScriptInterType
}

// IsExtraType extra
func (st ScriptType) IsExtraType() bool {
	return st == ScriptExtraType
}

// external nodeGroup script type
var (
	// ScriptInterType inter type
	ScriptInterType ScriptType = "inter"
	// ScriptExtraType extra type
	ScriptExtraType ScriptType = "extra"
)

const (
	// MasterRole label
	MasterRole = "node-role.kubernetes.io/master"
	// ControlPlanRole label
	ControlPlanRole = "node-role.kubernetes.io/control-plane"
)

const (
	// KubeAPIServer cluster apiserver key
	KubeAPIServer = "KubeAPIServer"
	// KubeController cluster controller key
	KubeController = "KubeController"
	// KubeScheduler cluster scheduler key
	KubeScheduler = "KubeScheduler"
	// Etcd cluster etcd key
	Etcd = "Etcd"
	// Kubelet cluster kubelet key
	Kubelet = "kubelet"
	// RootDir kubelet root-dir para
	RootDir = "root-dir"
	// RootDirValue kubelet root-dir value
	RootDirValue = "/data/bcs/service/kubelet"
)

var (
	// Unschedulable set node unSchedulable
	Unschedulable int64 = 1
)

// DefaultClusterConfig cluster default service config
var DefaultClusterConfig = map[string]string{
	Etcd: "node-data-dir=/data/bcs/lib/etcd;",
}

// DefaultNodeConfig default node config
var DefaultNodeConfig = map[string]string{
	Kubelet: "root-dir=/data/bcs/service/kubelet;",
}

var (
	// DefaultDockerRuntime xxx
	DefaultDockerRuntime = &RunTimeInfo{
		Runtime: DockerContainerRuntime,
		Version: DockerRuntimeVersion,
	}

	// DefaultContainerdRuntime xxx
	DefaultContainerdRuntime = &RunTimeInfo{
		Runtime: ContainerdRuntime,
		Version: ContainerdRuntimeVersion,
	}
)

// RunTimeInfo runtime
type RunTimeInfo struct {
	Runtime string
	Version string
}

// IsDockerRuntime docker
func IsDockerRuntime(runtime string) bool {
	return runtime == DockerContainerRuntime
}

// IsContainerdRuntime containerd
func IsContainerdRuntime(runtime string) bool {
	return runtime == ContainerdRuntime
}

const (
	// Iptables iptables mode
	Iptables = "iptables"
	// Ipvs ipvs mode
	Ipvs = "ipvs"

	// InitClusterID initClusterID
	InitClusterID = "BCS-K8S-00000"
	// RuntimeFlag xxx
	RuntimeFlag = "runtime"

	// ClusterApiServer cluster api server
	ClusterApiServer = "apiServer"
	// RegionName regionName
	RegionName = "regionName"

	// ShowSharedCluster flag show shared cluster
	ShowSharedCluster = "showSharedCluster"
	// VClusterNetworkKey xxx
	VClusterNetworkKey = "vclusterNetwork"
	// VClusterNamespaceInfo xxx
	VClusterNamespaceInfo = "namespaceInfo"
	// VclusterNetworkMode xxx
	VclusterNetworkMode = "vclusterMode"

	// ClusterManager xxx
	ClusterManager = "bcs-cluster-manager"

	// Biz business
	Biz = "biz"
	// BizSet business set
	BizSet = "biz_set"

	// Prod prod env
	Prod = "prod"
	// Debug debug env
	Debug = "debug"

	// Regions gke region cluster
	Regions = "regions"
	// Zones gke zone cluster
	Zones = "zones"

	// CloudClusterTypeKey cloud cluster type
	CloudClusterTypeKey = "CloudClusterType"
	// CloudClusterTypeEdge cloud cluster type for ECK EDGE
	CloudClusterTypeEdge = "K8SEXTENSION_EDGE"
	// CloudClusterTypeNative cloud cluster type for ECK native
	CloudClusterTypeNative = "K8SEXTENSION_NATIVE"

	// NodeRoleMaster node role master
	NodeRoleMaster = "MASTER_ETCD"
	// NodeRoleWorker node role worker
	NodeRoleWorker = "WORKER"

	// ClusterAddNodesLimit cluster addNodes limit
	ClusterAddNodesLimit = 100
	// ClusterManagerServiceDomain domain name for service
	ClusterManagerServiceDomain = "clustermanager.bkbcs.tencent.com"
	// ResourceManagerServiceDomain domain name for service
	ResourceManagerServiceDomain = "resourcemanager.bkbcs.tencent.com"

	// ClusterOverlayNetwork overlay
	ClusterOverlayNetwork = "overlay"
	// ClusterUnderlayNetwork underlay
	ClusterUnderlayNetwork = "underlay"

	// Flannel network plugin
	Flannel = "flannel"
	// GlobalRouter gr plugin
	GlobalRouter = "GR"
	// VpcCni vpc-cni plugin
	VpcCni = "VPC-CNI"
	// CiliumOverlay cilium plugin
	CiliumOverlay = "CiliumOverlay"

	// KubeletRootDirPath root-dir default path
	KubeletRootDirPath = "/data/bcs/service/kubelet"

	// DockerGraphPath docker path
	DockerGraphPath = "/data/bcs/service/docker"
	// MountTarget default mount path
	MountTarget = "/data"

	// DefaultImageName default image name
	DefaultImageName = "TencentOS Server 2.6 (TK4)"
	// DefaultECKImageName default ECK image name
	DefaultECKImageName = "CentOS-7.6-BITS64"

	// DockerContainerRuntime runtime
	DockerContainerRuntime = "docker"
	// DockerRuntimeVersion runtime version
	DockerRuntimeVersion = "19.3"
	// DockerRuntimeSelfVersion runtime version
	DockerRuntimeSelfVersion = "19.03.9"

	// ContainerdRuntime runtime
	ContainerdRuntime = "containerd"
	// ContainerdRuntimeVersion runtime version
	ContainerdRuntimeVersion = "1.4.3"

	// ClusterEngineTypeMesos mesos cluster
	ClusterEngineTypeMesos = "mesos"
	// ClusterEngineTypeK8s k8s cluster
	ClusterEngineTypeK8s = "k8s"

	// ClusterTypeFederation federation cluster
	ClusterTypeFederation = "federation"
	// ClusterTypeSingle single cluster
	ClusterTypeSingle = "single"
	// ClusterTypeVirtual virtual cluster
	ClusterTypeVirtual = "virtual"

	// MicroMetaKeyHTTPPort http port in micro service meta
	MicroMetaKeyHTTPPort = "httpport"

	// ClusterManageTypeManaged cloud manage cluster
	ClusterManageTypeManaged = "MANAGED_CLUSTER"
	// ClusterManageTypeIndependent BCS manage cluster
	ClusterManageTypeIndependent = "INDEPENDENT_CLUSTER"

	// TkeCidrStatusAvailable available tke cidr status
	TkeCidrStatusAvailable = "available"
	// TkeCidrStatusUsed used tke cidr status
	TkeCidrStatusUsed = "used"
	// TkeCidrStatusReserved reserved tke cidr status
	TkeCidrStatusReserved = "reserved"

	// Builder self builder cluster
	Builder = "builder"
	// Importer import external cluster
	Importer = "importer"
	// KubeConfigImport import
	KubeConfigImport = "kubeConfig"
	// CloudImport import
	CloudImport = "cloud"
	// ImportType cloud import type
	ImportType = "importType"
	// ClusterResourceGroup cluster resource group
	ClusterResourceGroup = "clusterResourceGroup"
	// NodeResourceGroup xxx
	NodeResourceGroup = "nodeResourceGroup"

	// CloudProjectId cloud project id
	CloudProjectId = "cloudProjectId"
	// TagClusterResourceKey resource tag key
	TagClusterResourceKey = "cluster"

	// StatusInitialization node/cluster/nodegroup/task status
	StatusInitialization = "INITIALIZATION"
	// StatusCreateClusterFailed status create failed
	StatusCreateClusterFailed = "CREATE-FAILURE"
	// StatusConnectClusterFailed status connect failed
	StatusConnectClusterFailed = "CONNECT-FAILURE"
	// StatusImportClusterFailed status import failed
	StatusImportClusterFailed = "IMPORT-FAILURE"
	// StatusRunning status running
	StatusRunning = "RUNNING"
	// StatusDeleting status deleting for scaling down
	StatusDeleting = "DELETING"
	// StatusDeleted status deleted
	StatusDeleted = "DELETED"
	// StatusDeleteClusterFailed status delete failed
	StatusDeleteClusterFailed = "DELETE-FAILURE"
	// StatusAddNodesFailed status add nodes failed
	StatusAddNodesFailed = "ADD-FAILURE"
	// StatusRemoveNodesFailed status remove nodes failed
	StatusRemoveNodesFailed = "REMOVE-FAILURE"
	// StatusNodeRemovable node is removable
	StatusNodeRemovable = "REMOVABLE"
	// StatusNodeUnknown node status is unknown
	StatusNodeUnknown = "UNKNOWN"
	// StatusNodeNotReady node not ready
	StatusNodeNotReady = "NOTREADY"

	// StatusDeleteNodeGroupFailed xxx
	StatusDeleteNodeGroupFailed = "DELETE-FAILURE"
	// StatusCreateNodeGroupCreating xxx
	StatusCreateNodeGroupCreating = "CREATING"
	// StatusDeleteNodeGroupDeleting xxx
	StatusDeleteNodeGroupDeleting = "DELETING"
	// StatusUpdateNodeGroupUpdating xxx
	StatusUpdateNodeGroupUpdating = "UPDATING"
	// StatusCreateNodeGroupFailed xxx
	StatusCreateNodeGroupFailed = "CREATE-FAILURE"

	// StatusAddCANodesFailed status add CA nodes failed
	StatusAddCANodesFailed = "ADD-CA-FAILURE"
	// StatusRemoveCANodesFailed delete CA nodes failure
	StatusRemoveCANodesFailed = "REMOVE-CA-FAILURE"

	// StatusResourceApplying 申请资源状态
	StatusResourceApplying = "APPLYING"
	// StatusResourceApplyFailed 申请资源失败状态
	StatusResourceApplyFailed = "APPLY-FAILURE"

	// StatusNodeGroupUpdating xxx
	StatusNodeGroupUpdating = "UPDATING"

	// StatusNodeGroupUpdateFailed xxx
	StatusNodeGroupUpdateFailed = "UPDATE-FAILURE"

	// StatusAutoScalingOptionNormal normal status
	StatusAutoScalingOptionNormal = "NORMAL"
	// StatusAutoScalingOptionUpdating updating status
	StatusAutoScalingOptionUpdating = "UPDATING"
	// StatusAutoScalingOptionUpdateFailed update failed status
	StatusAutoScalingOptionUpdateFailed = "UPDATE-FAILURE"
	// StatusAutoScalingOptionStopped stopped status
	StatusAutoScalingOptionStopped = "STOPPED"

	// TaskStatusSuccess task success
	TaskStatusSuccess = "SUCCESS"
	// TaskStatusFailure task failed
	TaskStatusFailure = "FAILURE"
	// TaskStatusTimeout task run timeout
	TaskStatusTimeout = "TIMEOUT"
)

const (
	// BcsErrClusterManagerSuccess success code
	BcsErrClusterManagerSuccess = 0
	// BcsErrClusterManagerSuccessStr success string
	BcsErrClusterManagerSuccessStr = "success"
	// BcsErrClusterManagerInvalidParameter invalid request parameter
	BcsErrClusterManagerInvalidParameter = bcscommon.BCSErrClusterManager + 1
	// BcsErrClusterManagerStoreOperationFailed invalid request parameter
	BcsErrClusterManagerStoreOperationFailed = bcscommon.BCSErrClusterManager + 2
	// BcsErrClusterManagerUnknown unknown error
	BcsErrClusterManagerUnknown = bcscommon.BCSErrClusterManager + 3
	// BcsErrClusterManagerUnknownStr unknown error msg
	BcsErrClusterManagerUnknownStr = "unknown error"

	// BcsErrClusterManagerDatabaseRecordNotFound database record not found
	BcsErrClusterManagerDatabaseRecordNotFound = bcscommon.BCSErrClusterManager + 4
	// BcsErrClusterManagerDatabaseRecordDuplicateKey database index key is duplicate
	BcsErrClusterManagerDatabaseRecordDuplicateKey = bcscommon.BCSErrClusterManager + 5
	// 6~19 is reserved error for database

	// BcsErrClusterManagerDBOperation db operation error
	BcsErrClusterManagerDBOperation = bcscommon.BCSErrClusterManager + 20

	// BcsErrClusterManagerAllocateClusterInCreateQuota allocate cluster error
	BcsErrClusterManagerAllocateClusterInCreateQuota = bcscommon.BCSErrClusterManager + 21
	// BcsErrClusterManagerK8SOpsFailed k8s operation failed
	BcsErrClusterManagerK8SOpsFailed = bcscommon.BCSErrClusterManager + 22
	// BcsErrClusterManagerResourceDuplicated resource duplicated
	BcsErrClusterManagerResourceDuplicated = bcscommon.BCSErrClusterManager + 23
	// BcsErrClusterManagerCommonErr common error
	BcsErrClusterManagerCommonErr = bcscommon.BCSErrClusterManager + 24
	// BcsErrClusterManagerTaskErr Task error
	BcsErrClusterManagerTaskErr = bcscommon.BCSErrClusterManager + 25
	// BcsErrClusterManagerCloudProviderErr cloudprovider error
	BcsErrClusterManagerCloudProviderErr = bcscommon.BCSErrClusterManager + 26
	// BcsErrClusterManagerDataEmptyErr request data empty error
	BcsErrClusterManagerDataEmptyErr = bcscommon.BCSErrClusterManager + 27
	// BcsErrClusterManagerClusterIDBuildErr build clusterID error
	BcsErrClusterManagerClusterIDBuildErr = bcscommon.BCSErrClusterManager + 28
	// BcsErrClusterManagerNodeManagerErr build clusterID error
	BcsErrClusterManagerNodeManagerErr = bcscommon.BCSErrClusterManager + 29
	// BcsErrClusterManagerTaskDoneErr build task doing or done error
	BcsErrClusterManagerTaskDoneErr = bcscommon.BCSErrClusterManager + 30
	// BcsErrClusterManagerSyncCloudErr cloud config error
	BcsErrClusterManagerSyncCloudErr = bcscommon.BCSErrClusterManager + 31
	// BcsErrClusterManagerCheckKubeErr cloud config error
	BcsErrClusterManagerCheckKubeErr = bcscommon.BCSErrClusterManager + 32
	// BcsErrClusterManagerCheckCloudClusterResourceErr cloud/cluster resource error
	BcsErrClusterManagerCheckCloudClusterResourceErr = bcscommon.BCSErrClusterManager + 33
	// BcsErrClusterManagerBkSopsInterfaceErr cloud/cluster resource error
	BcsErrClusterManagerBkSopsInterfaceErr = bcscommon.BCSErrClusterManager + 34
	// BcsErrClusterManagerDecodeBase64ScriptErr base64 error
	BcsErrClusterManagerDecodeBase64ScriptErr = bcscommon.BCSErrClusterManager + 35
	// BcsErrClusterManagerDecodeActionErr decode action error
	BcsErrClusterManagerDecodeActionErr = bcscommon.BCSErrClusterManager + 36
	// BcsErrClusterManagerExternalNodeScriptErr get external script action error
	BcsErrClusterManagerExternalNodeScriptErr = bcscommon.BCSErrClusterManager + 37
	// BcsErrClusterManagerCheckPermErr cloud config error
	BcsErrClusterManagerCheckPermErr = bcscommon.BCSErrClusterManager + 38
	// BcsErrClusterManagerGetPermErr cloud config error
	BcsErrClusterManagerGetPermErr = bcscommon.BCSErrClusterManager + 39
	// BcsErrClusterManagerCACleanNodesEmptyErr nodegroup clean nodes empty error
	BcsErrClusterManagerCACleanNodesEmptyErr = bcscommon.BCSErrClusterManager + 40
	// BcsErrClusterManagerCheckKubeConnErr cloud config error
	BcsErrClusterManagerCheckKubeConnErr = bcscommon.BCSErrClusterManager + 41
	// BcsErrClusterManagerClsMgrCloudErr cloud config error
	BcsErrClusterManagerClsMgrCloudErr = bcscommon.BCSErrClusterManager + 42
)

// ClusterIDRange for generate clusterID range
var ClusterIDRange = map[string][]int{
	"mesos-stag":  {10000, 15000},
	"mesos-debug": {20000, 25000},
	"mesos-prod":  {30000, 399999},
	"k8s-stag":    {15001, 19999},
	"k8s-debug":   {25001, 29999},
	"k8s-prod":    {40000, 1000000},
}

// Develop run environment
var Develop = "dev"

// StagClusterENV stag env
var StagClusterENV = "stag"

// ImageProvider
const (
	// ImageProvider 镜像提供方
	ImageProvider = "IMAGE_PROVIDER"
	// PublicImageProvider 公共镜像
	PublicImageProvider = "PUBLIC_IMAGE"
	// PrivateImageProvider 私有镜像
	PrivateImageProvider = "PRIVATE_IMAGE"
	// MarketImageProvider 市场镜像
	MarketImageProvider = "MARKET_IMAGE"
	// AllImageProvider 所有镜像
	AllImageProvider = "ALL"
)

// Instance sell status
const (
	// InstanceSell SELL status
	InstanceSell = "SELL"
	// InstanceSoldOut SOLD_OUT status
	InstanceSoldOut = "SOLD_OUT"
)

const (
	// True xxx
	True = "true"
	// False xxx
	False = "false"
	// Limit xxx
	Limit = 100
	// MaxFilterValues xxx
	MaxFilterValues = 5
)

const (
	// MetadataCookiesKey 在 GoMicro Metadata 中，Cookie 的键名
	MetadataCookiesKey = "Grpcgateway-Cookie"
	// LangCookieName 语言版本 Cookie 名称
	LangCookieName = "blueking_language"
)

// HeaderKey string
const (
	// ForwardedForHeaderKey is the header name of X-Forwarded-For.
	ForwardedForHeaderKey = "X-Forwarded-For"
	// UserAgentHeaderKey is the header name of User-Agent.
	UserAgentHeaderKey = "Grpcgateway-User-Agent"
)

const (
	// InterImport cloud inter import
	InterImport = "internal"
	// ExternalImport cloud external import
	ExternalImport = "external"
)

// InstanceChargeType
const (
	// PREPAID xxx
	PREPAID = "PREPAID" // 预付费，即包年包月
	// POSTPAIDBYHOUR xxx
	POSTPAIDBYHOUR = "POSTPAID_BY_HOUR" // 按小时后付费
	// NOTIFYANDAUTORENEW // 自动续期
	NOTIFYANDAUTORENEW = "NOTIFY_AND_AUTO_RENEW"
)
