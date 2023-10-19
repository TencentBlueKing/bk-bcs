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

package common

import (
	bcscommon "github.com/Tencent/bk-bcs/bcs-common/common"
)

// ResourceType resource type
type ResourceType string

// String ()
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

const (
	// MasterRole label
	MasterRole = "node-role.kubernetes.io/master"
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
)

// DefaultClusterConfig cluster default service config
var DefaultClusterConfig = map[string]string{
	Etcd: "node-data-dir=/data/bcs/lib/etcd;",
}

const (
	// InitClusterID initClusterID
	InitClusterID = "BCS-K8S-00000"
	// ClusterManager xxx
	ClusterManager = "bcs-cluster-manager"

	// Prod prod env
	Prod = "prod"
	// Debug debug env
	Debug = "debug"

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

	// DockerGraphPath docker path
	DockerGraphPath = "/data/bcs/service/docker"
	// MountTarget default mount path
	MountTarget = "/data"

	// DefaultImageName default image name
	DefaultImageName = "Tencent Linux Release 2.2 (Final)"
	// DefaultECKImageName default ECK image name
	DefaultECKImageName = "CentOS-7.6-BITS64"

	// DockerContainerRuntime runtime
	DockerContainerRuntime = "docker"
	// DockerRuntimeVersion runtime version
	DockerRuntimeVersion = "19.3"

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

	// TkeCidrStatusAvailable available tke cidr status
	TkeCidrStatusAvailable = "available"
	// TkeCidrStatusUsed used tke cidr status
	TkeCidrStatusUsed = "used"
	// TkeCidrStatusReserved reserved tke cidr status
	TkeCidrStatusReserved = "reserved"

	// StatusInitialization node/cluster/nodegroup status
	StatusInitialization = "INITIALIZATION"
	// StatusCreateClusterFailed status create failed
	StatusCreateClusterFailed = "CREATE-FAILURE"
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
	// StatusCreateNodeGroupFailed xxx
	StatusCreateNodeGroupFailed = "CREATE-FAILURE"
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
	// BcsErrClusterManagerCheckPermErr cloud config error
	BcsErrClusterManagerCheckPermErr = bcscommon.BCSErrClusterManager + 33
	// BcsErrClusterManagerGetPermErr cloud config error
	BcsErrClusterManagerGetPermErr = bcscommon.BCSErrClusterManager + 34
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

// Develop dev env
var Develop = "dev"

// StagClusterENV stag env
var StagClusterENV = "stag"

// ImageProvider
const (
	// 公共镜像
	PublicImageProvider = "PUBLIC_IMAGE"
	// 市场镜像
	MarketImageProvider = "MARKET_IMAGE"
)

// ContentType
const (
	MIMEJSONOrigin = "application/json+origin"
)
