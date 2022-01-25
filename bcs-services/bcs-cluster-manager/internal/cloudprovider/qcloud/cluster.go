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

package qcloud

import (
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var clsMgr sync.Once

func init() {
	clsMgr.Do(func() {
		//init Node
		cloudprovider.InitClusterManager("qcloud", &Cluster{})
	})
}

// Cluster tke management implementation
type Cluster struct {
}

// build task or handle data

// CreateCluster create kubenretes cluster according cloudprovider
func (c *Cluster) CreateCluster(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
	// call qcloud interface to create cluster
	if cls == nil {
		return nil, fmt.Errorf("qcloud CreateCluster cluster is empty")
	}

	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("qcloud CreateCluster cluster opt or cloud is empty")
	}

	if len(opt.Key) == 0 || len(opt.Secret) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud CreateCluster opt lost valid crendential info")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CreateCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build create cluster task
	task, err := mgr.BuildCreateClusterTask(cls, opt)
	if err != nil {
		blog.Errorf("build CreateCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// DeleteCluster delete kubenretes cluster according cloudprovider
func (c *Cluster) DeleteCluster(cls *proto.Cluster, opt *cloudprovider.DeleteClusterOption) (*proto.Task, error) {
	if cls == nil {
		return nil, fmt.Errorf("qcloud DeleteCluster cluster is empty")
	}

	if opt == nil || len(opt.Key) == 0 || len(opt.Secret) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud DeleteCluster cluster lost oprion")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build delete cluster task
	task, err := mgr.BuildDeleteClusterTask(cls, opt)
	if err != nil {
		blog.Errorf("build DeleteCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// GetCluster get kubernetes cluster detail information according cloudprovider
func (c *Cluster) GetCluster(cloudID string, opt *cloudprovider.GetClusterOption) (*proto.Cluster, error) {
	//qcloud.GetClusterClient(opt)

	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetNodesInCluster get all nodes belong to cluster according cloudprovider
func (c *Cluster) GetNodesInCluster(cls *proto.Cluster, opt *cloudprovider.GetNodesOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// AddNodesToCluster add new node to cluster according cloudprovider
func (c *Cluster) AddNodesToCluster(cls *proto.Cluster, nodes []*proto.Node, opt *cloudprovider.AddNodesOption) (*proto.Task, error) {
	if cls == nil {
		return nil, fmt.Errorf("qcloud AddNodesToCluster cluster is empty")
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("qcloud AddNodesToCluster nodes is empty")
	}

	if opt == nil || len(opt.Key) == 0 || len(opt.Secret) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud AddNodesToCluster cluster lost operation")
	}

	if opt.Operator == "" || opt.Cloud == nil {
		return nil, fmt.Errorf("qcloud AddNodesToCluster cluster lost operator|cloud")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when AddNodesToCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build add nodes to cluster task
	task, err := mgr.BuildAddNodesToClusterTask(cls, nodes, opt)
	if err != nil {
		blog.Errorf("build AddNodesToCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// DeleteNodesFromCluster delete specified nodes from cluster according cloudprovider
func (c *Cluster) DeleteNodesFromCluster(cls *proto.Cluster, nodes []*proto.Node, opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
	if cls == nil {
		return nil, fmt.Errorf("qcloud DeleteNodesFromCluster cluster is empty")
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("qcloud DeleteNodesFromCluster nodes is empty")
	}

	if opt == nil || len(opt.Key) == 0 || len(opt.Secret) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("qcloud DeleteNodesFromCluster cluster lost operation")
	}

	if opt.Operator == "" || opt.Cloud == nil {
		return nil, fmt.Errorf("qcloud DeleteNodesFromCluster cluster lost operator|cloud")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteNodesFromCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	// build delete nodes from cluster task
	task, err := mgr.BuildRemoveNodesFromClusterTask(cls, nodes, opt)
	if err != nil {
		blog.Errorf("build DeleteNodesFromCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// CheckClusterCidrAvailable check cluster CIDR nodesNum when add nodes
func (c *Cluster) CheckClusterCidrAvailable(cls *proto.Cluster, opt *cloudprovider.CheckClusterCIDROption) (bool, error) {
	if cls == nil || opt == nil {
		return true, nil
	}
	clusterCIDR := cls.NetworkSettings.GetClusterIPv4CIDR()
	if len(clusterCIDR) == 0 {
		return false, fmt.Errorf("cluster[%s] CIDR is empty", cls.ClusterID)
	}

	cidr, err := utils.ParseCIDR(clusterCIDR)
	if err != nil {
		return true, err
	}

	ipCount, err := strconv.ParseUint(cidr.IPCount().String(), 10, 64)
	if err != nil {
		return true, err
	}

	if cls.NetworkSettings.MaxNodePodNum <= 0 || cls.NetworkSettings.MaxServiceNum <= 0 ||
		ipCount <= uint64(cls.NetworkSettings.MaxServiceNum) {
		return true, nil
	}

	// CIDR IP 数量 - 集群内 Service 数量上限）/ 单节点 Pod 数量上限
	clusterTotalNodes := uint64(math.Floor(float64((ipCount - uint64(cls.NetworkSettings.MaxServiceNum)) / uint64(cls.NetworkSettings.MaxNodePodNum))))

	blog.Infof("cluster[%s] cloud[%s] CheckClusterCidrAvailable ipCount[%v] totalNodesCnt[%v] currentNodes[%v] masterCnt[%v]"+
		"addNodeCnt[%v]", cls.ClusterID, cloudName, ipCount, clusterTotalNodes, opt.CurrentNodeCnt, len(cls.Master), opt.IncomingNodeCnt)

	availableNodesCnt := clusterTotalNodes - uint64(len(cls.Master)) - opt.CurrentNodeCnt
	if availableNodesCnt-opt.IncomingNodeCnt < 0 {
		return false, fmt.Errorf("cluster[%s] cloud[%s] availableIPCnt[%v]", cls.ClusterID, cloudName, availableNodesCnt)
	}

	return true, nil
}
