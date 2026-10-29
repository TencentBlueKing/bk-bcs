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

package ibm

import (
	"context"
	"sync"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var clusterMgr sync.Once

func init() {
	clusterMgr.Do(func() {
		cloudprovider.InitClusterManager(cloudName, &ClusterManager{})
	})
}

// ClusterManager cluster manager for ibm
type ClusterManager struct{}

// CreateCluster create kubernetes cluster
func (c *ClusterManager) CreateCluster(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (
	*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CreateVirtualCluster create virtual cluster by hostCluster namespace
func (c *ClusterManager) CreateVirtualCluster(cls *proto.Cluster,
	opt *cloudprovider.CreateVirtualClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ImportCluster import different cluster by provider
func (c *ClusterManager) ImportCluster(cls *proto.Cluster,
	opt *cloudprovider.ImportClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteCluster delete kubernetes cluster
func (c *ClusterManager) DeleteCluster(cls *proto.Cluster,
	opt *cloudprovider.DeleteClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteVirtualCluster delete virtual cluster in hostCluster
func (c *ClusterManager) DeleteVirtualCluster(cls *proto.Cluster,
	opt *cloudprovider.DeleteVirtualClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetCluster get kubernetes cluster detail information
func (c *ClusterManager) GetCluster(cloudID string, opt *cloudprovider.GetClusterOption) (
	*proto.Cluster, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListCluster get cloud cluster list by region
func (c *ClusterManager) ListCluster(opt *cloudprovider.ListClusterOption) (
	[]*proto.CloudClusterInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterCidrAvailable check cluster cidr if meet to add nodes
func (c *ClusterManager) CheckClusterCidrAvailable(cls *proto.Cluster,
	opt *cloudprovider.CheckClusterCIDROption) (bool, error) {
	return false, cloudprovider.ErrCloudNotImplemented
}

// GetNodesInCluster get all nodes belong to cluster
func (c *ClusterManager) GetNodesInCluster(cls *proto.Cluster,
	opt *cloudprovider.GetNodesOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// AddNodesToCluster add new node to cluster
func (c *ClusterManager) AddNodesToCluster(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.AddNodesOption) ([]*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteNodesFromCluster delete specified nodes from cluster
func (c *ClusterManager) DeleteNodesFromCluster(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// EnableExternalNodeSupport enable cluster support external node
func (c *ClusterManager) EnableExternalNodeSupport(cls *proto.Cluster,
	opt *cloudprovider.EnableExternalNodeOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// ListOsImage get osimage list
func (c *ClusterManager) ListOsImage(provider, clusterID string,
	opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListProjects get cloud projects list
func (c *ClusterManager) ListProjects(opt *cloudprovider.CommonOption) (
	[]*proto.CloudProject, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterEndpointStatus check cluster endpoint status
func (c *ClusterManager) CheckClusterEndpointStatus(clusterID string, isExtranet bool,
	opt *cloudprovider.CheckEndpointStatusOption) (bool, error) {
	return false, cloudprovider.ErrCloudNotImplemented
}

// UpdateCloudKubeConfig update cluster kubeconfig to clustercredential
func (c *ClusterManager) UpdateCloudKubeConfig(kubeconfig string,
	opt *cloudprovider.UpdateCloudKubeConfigOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// GetMasterSuggestedMachines get master suggested machines
func (c *ClusterManager) GetMasterSuggestedMachines(level, vpcId string,
	opt *cloudprovider.GetMasterSuggestedMachinesOption) ([]*proto.InstanceTemplateConfig, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// AddSubnetsToCluster cluster add subnet
func (c *ClusterManager) AddSubnetsToCluster(ctx context.Context, subnet *proto.SubnetSource,
	opt *cloudprovider.AddSubnetsToClusterOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// AppendCloudNodeInfo append cloud node detailed info
func (c *ClusterManager) AppendCloudNodeInfo(ctx context.Context, nodes []*proto.ClusterNode,
	opt *cloudprovider.CommonOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// CheckIfGetNodesFromCluster check cluster if can get nodes from k8s
func (c *ClusterManager) CheckIfGetNodesFromCluster(ctx context.Context, cluster *proto.Cluster,
	nodes []*proto.ClusterNode) bool {
	return false
}

// SwitchClusterNetwork switch cluster network mode
func (c *ClusterManager) SwitchClusterNetwork(cls *proto.Cluster, subnet *proto.SubnetSource,
	opt *cloudprovider.SwitchClusterNetworkOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterNetworkStatus check cluster network status
func (c *ClusterManager) CheckClusterNetworkStatus(cloudID string,
	opt *cloudprovider.CheckClusterNetworkStatusOption) (bool, error) {
	return false, cloudprovider.ErrCloudNotImplemented
}

// CheckHighAvailabilityMasterNodes check master nodes high availability
func (c *ClusterManager) CheckHighAvailabilityMasterNodes(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.CheckHaMasterNodesOption) error {
	return cloudprovider.ErrCloudNotImplemented
}
