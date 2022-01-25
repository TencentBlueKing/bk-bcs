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

var clusterMgrs map[string]ClusterManager
var nodeGroupMgrs map[string]NodeGroupManager
var nodeMgrs map[string]NodeManager
var taskMgrs map[string]TaskManager
var storage store.ClusterManagerModel

var lock sync.RWMutex
var once sync.Once

func init() {
	once.Do(func() {
		clusterMgrs = make(map[string]ClusterManager)
		nodeGroupMgrs = make(map[string]NodeGroupManager)
		nodeMgrs = make(map[string]NodeManager)
		taskMgrs = make(map[string]TaskManager)
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

// CloudInfoManager cloud interface for basic config info(region or no region)
type CloudInfoManager interface {
	// GetRegionsInfo get regions info by cloud(no region)
	GetRegionsInfo(opt *CommonOption) (map[string]string, error)
	// GetRegionZonesInfo get zones info by region
	GetRegionZonesInfo(opt *CommonOption) ([]*proto.ZoneInfo, error)
	// GetClusterK8sVersionInfo get k8s version info
	GetClusterK8sVersionInfo(opt *CommonOption) ([]string, error)
	// GetClusterImages get cluster images by region
	GetClusterImages(opt *CommonOption) ([]*proto.ImageInfo, error)
	// GetImageIDByImageName get image by image name
	GetImageIDByImageName(name string, opt *CommonOption) (string, error)
}

// NodeManager cloud interface for cvm management
type NodeManager interface {
	// GetNodeByIP get specified Node by innerIP address
	GetNodeByIP(ip string, opt *GetNodeOption) (*proto.Node, error)
	// ListNodesByIP list node by IP set
	ListNodesByIP(ips []string, opt *ListNodesOption) ([]*proto.Node, error)
	// GetCVMImageIDByImageName
	GetCVMImageIDByImageName(imageName string, opt *CommonOption) (string, error)
}

// ClusterManager cloud interface for kubernetes cluster management
type ClusterManager interface {
	// CreateCluster create kubernetes cluster according cloudprovider
	CreateCluster(cls *proto.Cluster, opt *CreateClusterOption) (*proto.Task, error)
	// DeleteCluster delete kubernetes cluster according cloudprovider
	DeleteCluster(cls *proto.Cluster, opt *DeleteClusterOption) (*proto.Task, error)
	// GetCluster get kubernetes cluster detail information according cloudprovider
	GetCluster(cloudID string, opt *GetClusterOption) (*proto.Cluster, error)
	// CheckClusterCidrAvailable check cluster cidr if meet to add nodes
	CheckClusterCidrAvailable(cls *proto.Cluster, opt *CheckClusterCIDROption) (bool, error)
	// GetNodesInCluster get all nodes belong to cluster according cloudprovider
	GetNodesInCluster(cls *proto.Cluster, opt *GetNodesOption) ([]*proto.Node, error)
	// AddNodesToCluster add new node to cluster according cloudprovider
	AddNodesToCluster(cls *proto.Cluster, nodes []*proto.Node, opt *AddNodesOption) (*proto.Task, error)
	// DeleteNodesFromCluster delete specified nodes from cluster according cloudprovider
	DeleteNodesFromCluster(cls *proto.Cluster, nodes []*proto.Node, opt *DeleteNodesOption) (*proto.Task, error)
}

// NodeGroupManager cloud interface for nodegroup management
type NodeGroupManager interface {
	// CreateNodeGroup create nodegroup by cloudprovider api, only create NodeGroup entity
	CreateNodeGroup(group *proto.NodeGroup, opt *CreateNodeGroupOption) error
	// DeleteNodeGroup delete nodegroup by cloudprovider api, all nodes belong to NodeGroup
	// will be released. Task is backgroup automatic task
	DeleteNodeGroup(group *proto.NodeGroup, nodes []*proto.Node, opt *DeleteNodeGroupOption) (*proto.Task, error)
	// UpdateNodeGroup update specified nodegroup configuration
	UpdateNodeGroup(group *proto.NodeGroup, opt *CommonOption) error
	// GetNodesInGroup get all nodes belong to NodeGroup
	GetNodesInGroup(group *proto.NodeGroup, opt *CommonOption) ([]*proto.Node, error)
	// MoveNodesToGroup add cluster nodes to NodeGroup
	MoveNodesToGroup(nodes []*proto.Node, group *proto.NodeGroup, opt *MoveNodesOption) error

	// RemoveNodesFromGroup remove nodes from NodeGroup, nodes are still in cluster
	RemoveNodesFromGroup(nodes []*proto.Node, group *proto.NodeGroup, opt *RemoveNodesOption) error
	// CleanNodesInGroup clean specified nodes in NodeGroup,
	CleanNodesInGroup(nodes []*proto.Node, group *proto.NodeGroup, opt *CleanNodesOption) (*CleanNodesResponse, error)
	// UpdateDesiredNodes update nodegroup desired node
	UpdateDesiredNodes(desired uint32, group *proto.NodeGroup, opt *UpdateDesiredNodeOption) (*ScalingResponse, error)

	// CreateAutoScalingOption create cluster autoscaling option, cloudprovider will
	// deploy cluster-autoscaler in backgroup according cloudprovider implementation
	CreateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption, opt *CreateScalingOption) (*proto.Task, error)
	// DeleteAutoScalingOption delete cluster autoscaling, cloudprovider will clean
	// cluster-autoscaler in backgroup according cloudprovider implementation
	DeleteAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption, opt *DeleteScalingOption) (*proto.Task, error)
	// UpdateAutoScalingOption update cluster autoscaling option, cloudprovider will update
	// cluster-autoscaler configuration in backgroup according cloudprovider implementation.
	// Implementation is optional.
	UpdateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption, opt *DeleteScalingOption) (*proto.Task, error)
}

// TaskManager backgroup back management
type TaskManager interface {
	Name() string
	// GetAllTask get all register task for worker running
	GetAllTask() map[string]interface{}

	// specific cloud different implement

	// NodeGroup taskList

	// BuildCleanNodesInGroupTask clean specified nodes in NodeGroup
	BuildCleanNodesInGroupTask(nodes []*proto.Node, group *proto.NodeGroup, opt *TaskOptions) (*proto.Task, error)
	// BuildScalingNodesTask when scaling nodes, we need to create background task to verify scaling status and update new nodes to local storage.
	BuildScalingNodesTask(scaling uint32, group *proto.NodeGroup, opt *TaskOptions) (*proto.Task, error)
	// BuildDeleteNodeGroupTask when delete nodegroup, we need to create background
	// task to clean all nodes in nodegroup, release all resource in cloudprovider,
	// finally delete nodes information in local storage.
	BuildDeleteNodeGroupTask(group *proto.NodeGroup, nodes []*proto.Node, opt *DeleteNodeGroupOption) (*proto.Task, error)

	// ClusterManager taskList

	// BuildCreateClusterTask create cluster by different cloud provider
	BuildCreateClusterTask(cls *proto.Cluster, opt *CreateClusterOption) (*proto.Task, error)
	// BuildDeleteClusterTask delete cluster by different cloud provider
	BuildDeleteClusterTask(cls *proto.Cluster, opt *DeleteClusterOption) (*proto.Task, error)
	// BuildAddNodesToClusterTask add instances to cluster
	BuildAddNodesToClusterTask(cls *proto.Cluster, nodes []*proto.Node, opt *AddNodesOption) (*proto.Task, error)
	// BuildRemoveNodesFromClusterTask remove instances from cluster
	BuildRemoveNodesFromClusterTask(cls *proto.Cluster, nodes []*proto.Node, opt *DeleteNodesOption) (*proto.Task, error)
}
