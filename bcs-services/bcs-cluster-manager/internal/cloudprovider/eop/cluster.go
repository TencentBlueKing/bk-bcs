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

package eop

import (
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var clsMgr sync.Once

func init() {
	clsMgr.Do(func() {
		//init Node
		cloudprovider.InitClusterManager(cloudName, &Cluster{})
	})
}

// Cluster tke management implementation
type Cluster struct {
}

// CreateCluster create kubenretes cluster according cloudprovider
func (c Cluster) CreateCluster(cls *cmproto.Cluster, opt *cloudprovider.CreateClusterOption) (*cmproto.Task, error) {
	// call eopCloud interface to create cluster
	if cls == nil {
		return nil, fmt.Errorf("eopCloud CreateCluster cluster is empty")
	}

	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("eopCloud CreateCluster cluster opt or cloud is empty")
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("eopCloud CreateCluster opt lost valid crendential info")
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

// ImportCluster import cluster according cloudprovider
func (c Cluster) ImportCluster(cls *cmproto.Cluster, opt *cloudprovider.ImportClusterOption) (*cmproto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteCluster delete kubenretes cluster according cloudprovider
func (c Cluster) DeleteCluster(cls *cmproto.Cluster, opt *cloudprovider.DeleteClusterOption) (*cmproto.Task, error) {
	if cls == nil {
		return nil, fmt.Errorf("eopCloud DeleteCluster cluster is empty")
	}

	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 {
		return nil, fmt.Errorf("eopCloud DeleteCluster cluster lost oprion")
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
func (c Cluster) GetCluster(cloudID string, opt *cloudprovider.GetClusterOption) (*cmproto.Cluster, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListCluster get cloud cluster list by region
func (c Cluster) ListCluster(opt *cloudprovider.ListClusterOption) ([]*cmproto.CloudClusterInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterCidrAvailable check cluster CIDR nodesNum when add nodes
func (c Cluster) CheckClusterCidrAvailable(cls *cmproto.Cluster, opt *cloudprovider.CheckClusterCIDROption) (bool, error) {
	return false, cloudprovider.ErrCloudNotImplemented
}

// GetNodesInCluster get all nodes belong to cluster according cloudprovider
func (c Cluster) GetNodesInCluster(cls *cmproto.Cluster, opt *cloudprovider.GetNodesOption) ([]*cmproto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// AddNodesToCluster add new node to cluster according cloudprovider
func (c Cluster) AddNodesToCluster(cls *cmproto.Cluster, nodes []*cmproto.Node, opt *cloudprovider.AddNodesOption) (*cmproto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteNodesFromCluster delete specified nodes from cluster according cloudprovider
func (c Cluster) DeleteNodesFromCluster(cls *cmproto.Cluster, nodes []*cmproto.Node, opt *cloudprovider.DeleteNodesOption) (*cmproto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListOsImage list image os
func (c Cluster) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*cmproto.OsImage, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
