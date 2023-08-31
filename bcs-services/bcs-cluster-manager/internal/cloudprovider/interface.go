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

package cloudprovider

import (
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

var (
	lock sync.RWMutex
	once sync.Once

	clusterMgrs       map[string]ClusterManager
	cloudInfoMgrs     map[string]CloudInfoManager
	cloudValidateMgrs map[string]CloudValidateManager
	nodeGroupMgrs     map[string]NodeGroupManager
	nodeMgrs          map[string]NodeManager
	taskMgrs          map[string]TaskManager
	vpcMgrs           map[string]VPCManager
	storage           store.ClusterManagerModel
)

func init() {
	once.Do(func() {
		clusterMgrs = make(map[string]ClusterManager)
		nodeGroupMgrs = make(map[string]NodeGroupManager)
		nodeMgrs = make(map[string]NodeManager)
		taskMgrs = make(map[string]TaskManager)

		cloudInfoMgrs = make(map[string]CloudInfoManager)
		cloudValidateMgrs = make(map[string]CloudValidateManager)
		vpcMgrs = make(map[string]VPCManager)
	})
}

// InitStorageModel for cluster manager storage tools
func InitStorageModel(model store.ClusterManagerModel) {
	lock.Lock()
	defer lock.Unlock()
	storage = model
}

// GetStorageModel for cluster manager storage tools
func GetStorageModel() store.ClusterManagerModel {
	return storage
}

// InitTaskManager for cluster manager initialization
func InitTaskManager(provider string, t TaskManager) {
	lock.Lock()
	defer lock.Unlock()
	taskMgrs[provider] = t
}

// GetTaskManager for nodegroup manager initialization
func GetTaskManager(provider string) (TaskManager, error) {
	lock.RLock()
	defer lock.RUnlock()
	mgr, ok := taskMgrs[provider]
	if !ok {
		return nil, ErrCloudNoProvider
	}
	return mgr, nil
}

// GetAllTaskManager for all task manager
func GetAllTaskManager() []TaskManager {
	lock.RLock()
	defer lock.RUnlock()
	var mgrs []TaskManager
	for _, mgr := range taskMgrs {
		mgrs = append(mgrs, mgr)
	}
	return mgrs
}

// InitClusterManager for cluster manager initialization
func InitClusterManager(provider string, cls ClusterManager) {
	lock.Lock()
	defer lock.Unlock()
	clusterMgrs[provider] = cls
}

// InitNodeGroupManager for nodegroup manager initialization
func InitNodeGroupManager(provider string, group NodeGroupManager) {
	lock.Lock()
	defer lock.Unlock()
	nodeGroupMgrs[provider] = group
}

// InitNodeManager for nodegroup manager initialization
func InitNodeManager(provider string, nodeMgr NodeManager) {
	lock.Lock()
	defer lock.Unlock()
	nodeMgrs[provider] = nodeMgr
}

// InitCloudInfoManager for cloudInfo manager initialization
func InitCloudInfoManager(provider string, cloudInfoMgr CloudInfoManager) {
	lock.Lock()
	defer lock.Unlock()
	cloudInfoMgrs[provider] = cloudInfoMgr
}

// InitCloudValidateManager for cloud validate manager check
func InitCloudValidateManager(provider string, cloudValidateMgr CloudValidateManager) {
	lock.Lock()
	defer lock.Unlock()
	cloudValidateMgrs[provider] = cloudValidateMgr
}

// InitVPCManager for vpc manager check
func InitVPCManager(provider string, vpcMgr VPCManager) {
	lock.Lock()
	defer lock.Unlock()
	vpcMgrs[provider] = vpcMgr
}

// GetClusterMgr get cluster manager implementation according cloud provider
func GetClusterMgr(provider string) (ClusterManager, error) {
	lock.RLock()
	defer lock.RUnlock()
	cls, ok := clusterMgrs[provider]
	if !ok {
		return nil, ErrCloudNoProvider
	}
	return cls, nil
}

// GetNodeGroupMgr get NodeGroup implementation according cloud provider
func GetNodeGroupMgr(provider string) (NodeGroupManager, error) {
	lock.RLock()
	defer lock.RUnlock()
	group, ok := nodeGroupMgrs[provider]
	if !ok {
		return nil, ErrCloudNoProvider
	}
	return group, nil
}

// GetNodeMgr get node implementation according cloud provider
func GetNodeMgr(provider string) (NodeManager, error) {
	lock.RLock()
	defer lock.RUnlock()
	nodeMgr, ok := nodeMgrs[provider]
	if !ok {
		return nil, ErrCloudNoProvider
	}
	return nodeMgr, nil
}

// GetCloudInfoMgr get cloudInfo according cloud provider
func GetCloudInfoMgr(provider string) (CloudInfoManager, error) {
	lock.RLock()
	defer lock.RUnlock()
	cloudInfo, ok := cloudInfoMgrs[provider]
	if !ok {
		return nil, ErrCloudNoProvider
	}
	return cloudInfo, nil
}

// GetCloudValidateMgr get cloudValidate according cloud provider
func GetCloudValidateMgr(provider string) (CloudValidateManager, error) {
	lock.RLock()
	defer lock.RUnlock()

	cloudValidate, ok := cloudValidateMgrs[provider]
	if !ok {
		return nil, ErrCloudNoProvider
	}

	return cloudValidate, nil
}

// GetVPCMgr get vpc according cloud provider
func GetVPCMgr(provider string) (VPCManager, error) {
	lock.RLock()
	defer lock.RUnlock()

	vpcmgr, ok := vpcMgrs[provider]
	if !ok {
		return nil, ErrCloudNoProvider
	}

	return vpcmgr, nil
}

// CloudInfoManager cloud interface for basic config info(region or no region)
type CloudInfoManager interface {
	// InitCloudClusterDefaultInfo init cloud cluster default configInfo
	InitCloudClusterDefaultInfo(cls *proto.Cluster, opt *InitClusterConfigOption) error
	// SyncClusterCloudInfo sync cluster metadata
	SyncClusterCloudInfo(cls *proto.Cluster, opt *SyncClusterCloudInfoOption) error
}

// NodeManager cloud interface for cvm management
type NodeManager interface {
	// GetNodeByIP get specified Node by innerIP address
	GetNodeByIP(ip string, opt *GetNodeOption) (*proto.Node, error)
	// ListNodesByIP list node by IP set
	ListNodesByIP(ips []string, opt *ListNodesOption) ([]*proto.Node, error)
	// GetExternalNodeByIP get specified Node by innerIP address
	GetExternalNodeByIP(ip string, opt *GetNodeOption) (*proto.Node, error)
	// ListExternalNodesByIP list node by IP set
	ListExternalNodesByIP(ips []string, opt *ListNodesOption) ([]*proto.Node, error)
	// GetCVMImageIDByImageName get imageID by imageName
	GetCVMImageIDByImageName(imageName string, opt *CommonOption) (string, error)
	// GetCloudRegions get cloud regions
	GetCloudRegions(opt *CommonOption) ([]*proto.RegionInfo, error)
	// GetZoneList get zoneList by region
	GetZoneList(opt *CommonOption) ([]*proto.ZoneInfo, error)
	// ListNodeInstanceType get node instance type list
	ListNodeInstanceType(info InstanceInfo, opt *CommonOption) ([]*proto.InstanceType, error)
	// ListOsImage get osimage list
	ListOsImage(provider string, opt *CommonOption) ([]*proto.OsImage, error)
	// ListKeyPairs list ssh keyPairs
	ListKeyPairs(opt *CommonOption) ([]*proto.KeyPair, error)
}

// CloudValidateManager validate interface for check cloud resourceInfo
type CloudValidateManager interface {
	// ImportClusterValidate import cluster validate
	ImportClusterValidate(req *proto.ImportClusterReq, opt *CommonOption) error
	// AddNodesToClusterValidate validate
	AddNodesToClusterValidate(req *proto.AddNodesRequest, opt *CommonOption) error
	// DeleteNodesFromClusterValidate validate
	DeleteNodesFromClusterValidate(req *proto.DeleteNodesRequest, opt *CommonOption) error
	// ImportCloudAccountValidate import cloud account validate
	ImportCloudAccountValidate(account *proto.Account) error
	// GetCloudRegionZonesValidate get cloud region zones validate
	GetCloudRegionZonesValidate(req *proto.GetCloudRegionZonesRequest, account *proto.Account) error
	// ListCloudRegionClusterValidate get cloud region zones validate
	ListCloudRegionClusterValidate(req *proto.ListCloudRegionClusterRequest, account *proto.Account) error
	// ListCloudSubnetsValidate list subnets validate
	ListCloudSubnetsValidate(req *proto.ListCloudSubnetsRequest, account *proto.Account) error
	// ListSecurityGroupsValidate list SecurityGroups validate
	ListSecurityGroupsValidate(req *proto.ListCloudSecurityGroupsRequest, account *proto.Account) error
	// ListKeyPairsValidate list key pairs validate
	ListKeyPairsValidate(req *proto.ListKeyPairsRequest, account *proto.Account) error
	// ListInstanceTypeValidate list instance type validate
	ListInstanceTypeValidate(req *proto.ListCloudInstanceTypeRequest, account *proto.Account) error
	// ListCloudOsImageValidate list tke image os validate
	ListCloudOsImageValidate(req *proto.ListCloudOsImageRequest, account *proto.Account) error
	// CreateNodeGroupValidate create node group validate
	CreateNodeGroupValidate(req *proto.CreateNodeGroupRequest, opt *CommonOption) error
	// ListInstancesValidate ListInstanceTypeValidate list instance type validate
	ListInstancesValidate(req *proto.ListCloudInstancesRequest, account *proto.Account) error
}

// ClusterManager cloud interface for kubernetes cluster management
type ClusterManager interface {
	// CreateCluster create kubernetes cluster according cloudprovider
	CreateCluster(cls *proto.Cluster, opt *CreateClusterOption) (*proto.Task, error)
	// CreateVirtualCluster create virtual cluster by hostCluster namespace
	CreateVirtualCluster(cls *proto.Cluster, opt *CreateVirtualClusterOption) (*proto.Task, error)
	// ImportCluster import different cluster by provider
	ImportCluster(cls *proto.Cluster, opt *ImportClusterOption) (*proto.Task, error)
	// DeleteCluster delete kubernetes cluster according cloudprovider
	DeleteCluster(cls *proto.Cluster, opt *DeleteClusterOption) (*proto.Task, error)
	// DeleteVirtualCluster delete virtual cluster in hostCluster according cloudprovider
	DeleteVirtualCluster(cls *proto.Cluster, opt *DeleteVirtualClusterOption) (*proto.Task, error)
	// GetCluster get kubernetes cluster detail information according cloudprovider
	GetCluster(cloudID string, opt *GetClusterOption) (*proto.Cluster, error)
	// ListCluster get cloud cluster list by region
	ListCluster(opt *ListClusterOption) ([]*proto.CloudClusterInfo, error)
	// CheckClusterCidrAvailable check cluster cidr if meet to add nodes
	CheckClusterCidrAvailable(cls *proto.Cluster, opt *CheckClusterCIDROption) (bool, error)
	// GetNodesInCluster get all nodes belong to cluster according cloudprovider
	GetNodesInCluster(cls *proto.Cluster, opt *GetNodesOption) ([]*proto.Node, error)
	// AddNodesToCluster add new node to cluster according cloudprovider
	AddNodesToCluster(cls *proto.Cluster, nodes []*proto.Node, opt *AddNodesOption) (*proto.Task, error)
	// DeleteNodesFromCluster delete specified nodes from cluster according cloudprovider
	DeleteNodesFromCluster(cls *proto.Cluster, nodes []*proto.Node, opt *DeleteNodesOption) (*proto.Task, error)
	// EnableExternalNodeSupport enable cluster support external node
	EnableExternalNodeSupport(cls *proto.Cluster, opt *EnableExternalNodeOption) error
	// ListOsImage get osimage list
	ListOsImage(provider string, opt *CommonOption) ([]*proto.OsImage, error)
	// CheckClusterEndpointStatus check cluster endpoint status
	CheckClusterEndpointStatus(clusterID string, isExtranet bool, opt *CheckEndpointStatusOption) (bool, error)
}

// NodeGroupManager cloud interface for nodegroup management
type NodeGroupManager interface {
	// CreateNodeGroup create nodegroup by cloudprovider api, only create NodeGroup entity
	CreateNodeGroup(group *proto.NodeGroup, opt *CreateNodeGroupOption) (*proto.Task, error)
	// DeleteNodeGroup delete nodegroup by cloudprovider api, all nodes belong to NodeGroup
	// will be released. Task is backgroup automatic task
	DeleteNodeGroup(group *proto.NodeGroup, nodes []*proto.Node, opt *DeleteNodeGroupOption) (*proto.Task, error)
	// UpdateNodeGroup update specified nodegroup configuration
	UpdateNodeGroup(group *proto.NodeGroup, opt *UpdateNodeGroupOption) (*proto.Task, error)
	// GetNodesInGroup get all nodes belong to NodeGroup
	GetNodesInGroup(group *proto.NodeGroup, opt *CommonOption) ([]*proto.Node, error)
	// GetNodesInGroupV2 get all nodes belong to NodeGroup
	GetNodesInGroupV2(group *proto.NodeGroup, opt *CommonOption) ([]*proto.NodeGroupNode, error)
	// MoveNodesToGroup add cluster nodes to NodeGroup
	MoveNodesToGroup(nodes []*proto.Node, group *proto.NodeGroup, opt *MoveNodesOption) (*proto.Task, error)

	// RemoveNodesFromGroup remove nodes from NodeGroup, nodes are still in cluster
	RemoveNodesFromGroup(nodes []*proto.Node, group *proto.NodeGroup, opt *RemoveNodesOption) error
	// CleanNodesInGroup clean specified nodes in NodeGroup,
	CleanNodesInGroup(nodes []*proto.Node, group *proto.NodeGroup, opt *CleanNodesOption) (*proto.Task, error)
	// UpdateDesiredNodes update nodegroup desired node
	UpdateDesiredNodes(desired uint32, group *proto.NodeGroup, opt *UpdateDesiredNodeOption) (*ScalingResponse, error)
	// SwitchNodeGroupAutoScaling switch nodegroup auto scale
	SwitchNodeGroupAutoScaling(group *proto.NodeGroup, enable bool, opt *SwitchNodeGroupAutoScalingOption) (*proto.Task, error)

	// CreateAutoScalingOption create cluster autoscaling option, cloudprovider will
	// deploy cluster-autoscaler in backgroup according cloudprovider implementation
	CreateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption, opt *CreateScalingOption) (*proto.Task, error)
	// DeleteAutoScalingOption delete cluster autoscaling, cloudprovider will clean
	// cluster-autoscaler in backgroup according cloudprovider implementation
	DeleteAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption, opt *DeleteScalingOption) (*proto.Task, error)
	// UpdateAutoScalingOption update cluster autoscaling option, cloudprovider will update
	// cluster-autoscaler configuration in backgroup according cloudprovider implementation.
	// Implementation is optional.
	UpdateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption, opt *UpdateScalingOption) (*proto.Task, error)
	// SwitchAutoScalingOptionStatus switch cluster autoscaling option enable auto scaling status
	SwitchAutoScalingOptionStatus(scalingOption *proto.ClusterAutoScalingOption, enable bool, opt *CommonOption) (*proto.Task, error)

	//ExternalNode support external node operation

	// AddExternalNodeToCluster add external to cluster
	AddExternalNodeToCluster(group *proto.NodeGroup, nodes []*proto.Node, opt *AddExternalNodesOption) (*proto.Task, error)
	// DeleteExternalNodeFromCluster remove external node from cluster
	DeleteExternalNodeFromCluster(group *proto.NodeGroup, nodes []*proto.Node, opt *DeleteExternalNodesOption) (*proto.Task, error)
	// GetExternalNodeScript get external node script from cluster nodeGroup
	GetExternalNodeScript(group *proto.NodeGroup) (string, error)
}

// VPCManager cloud interface for vpc management
type VPCManager interface {
	// ListSubnets list vpc's subnets
	ListSubnets(vpcID string, opt *CommonOption) ([]*proto.Subnet, error)
	// ListSecurityGroups list security groups
	ListSecurityGroups(opt *CommonOption) ([]*proto.SecurityGroup, error)
	// GetCloudNetworkAccountType get cloud account type
	GetCloudNetworkAccountType(opt *CommonOption) (*proto.CloudAccountType, error)
}

// TaskManager backgroup back management
type TaskManager interface {
	Name() string
	// GetAllTask get all register task for worker running
	GetAllTask() map[string]interface{}

	// specific cloud different implement

	// NodeGroup taskList

	// BuildCreateNodeGroupTask build create nodegroup task
	BuildCreateNodeGroupTask(group *proto.NodeGroup, opt *CreateNodeGroupOption) (*proto.Task, error)
	// BuildDeleteNodeGroupTask when delete nodegroup, we need to create background
	// task to clean all nodes in nodegroup, release all resource in cloudprovider,
	// finally delete nodes information in local storage.
	BuildDeleteNodeGroupTask(group *proto.NodeGroup, nodes []*proto.Node, opt *DeleteNodeGroupOption) (*proto.Task, error)
	// BuildMoveNodesToGroupTask when move nodes to nodegroup, we need to create background task
	BuildMoveNodesToGroupTask(nodes []*proto.Node, group *proto.NodeGroup, opt *MoveNodesOption) (*proto.Task, error)
	// BuildCleanNodesInGroupTask clean specified nodes in NodeGroup
	BuildCleanNodesInGroupTask(nodes []*proto.Node, group *proto.NodeGroup, opt *CleanNodesOption) (*proto.Task, error)
	// BuildUpdateDesiredNodesTask update nodegroup desired node
	BuildUpdateDesiredNodesTask(desired uint32, group *proto.NodeGroup, opt *UpdateDesiredNodeOption) (*proto.Task, error)

	// BuildSwitchNodeGroupAutoScalingTask switch nodegroup autoscaling
	BuildSwitchNodeGroupAutoScalingTask(group *proto.NodeGroup, enable bool, opt *SwitchNodeGroupAutoScalingOption) (*proto.Task, error)
	// BuildUpdateAutoScalingOptionTask update cluster autoscaling option
	BuildUpdateAutoScalingOptionTask(scalingOption *proto.ClusterAutoScalingOption, opt *UpdateScalingOption) (*proto.Task, error)
	// BuildSwitchAsOptionStatusTask switch cluster autoscaling option enable auto scaling status
	BuildSwitchAsOptionStatusTask(scalingOption *proto.ClusterAutoScalingOption, enable bool, opt *CommonOption) (*proto.Task, error)
	// BuildUpdateNodeGroupTask update nodeGroup data and update autoScalingOption
	BuildUpdateNodeGroupTask(group *proto.NodeGroup, opt *CommonOption) (*proto.Task, error)

	// ClusterManager taskList

	// BuildCreateVirtualClusterTask create virtual cluster by different cloud provider
	BuildCreateVirtualClusterTask(cls *proto.Cluster, opt *CreateVirtualClusterOption) (*proto.Task, error)
	// BuildDeleteVirtualClusterTask delete virtual cluster by different cloud provider
	BuildDeleteVirtualClusterTask(cls *proto.Cluster, opt *DeleteVirtualClusterOption) (*proto.Task, error)
	// BuildImportClusterTask create cluster by different cloud provider
	BuildImportClusterTask(cls *proto.Cluster, opt *ImportClusterOption) (*proto.Task, error)
	// BuildCreateClusterTask create cluster by different cloud provider
	BuildCreateClusterTask(cls *proto.Cluster, opt *CreateClusterOption) (*proto.Task, error)
	// BuildDeleteClusterTask delete cluster by different cloud provider
	BuildDeleteClusterTask(cls *proto.Cluster, opt *DeleteClusterOption) (*proto.Task, error)
	// BuildAddNodesToClusterTask add instances to cluster
	BuildAddNodesToClusterTask(cls *proto.Cluster, nodes []*proto.Node, opt *AddNodesOption) (*proto.Task, error)
	// BuildRemoveNodesFromClusterTask remove instances from cluster
	BuildRemoveNodesFromClusterTask(cls *proto.Cluster, nodes []*proto.Node, opt *DeleteNodesOption) (*proto.Task, error)

	// BuildAddExternalNodeToCluster add external to cluster
	BuildAddExternalNodeToCluster(group *proto.NodeGroup, nodes []*proto.Node, opt *AddExternalNodesOption) (*proto.Task, error)
	// BuildDeleteExternalNodeFromCluster remove external node from cluster
	BuildDeleteExternalNodeFromCluster(group *proto.NodeGroup, nodes []*proto.Node, opt *DeleteExternalNodesOption) (*proto.Task, error)
}
