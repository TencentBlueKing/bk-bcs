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

package azure

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var clusterMgr sync.Once
var defaultTimeout = 10 * time.Second

func init() {
	clusterMgr.Do(func() {
		cloudprovider.InitClusterManager(cloudName, &Cluster{})
	})
}

// Cluster kubernetes cluster management implementation
type Cluster struct {
}

// CreateCluster create kubernetes cluster according cloudprovider
func (c *Cluster) CreateCluster(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
	// call azure interface to create cluster
	if cls == nil {
		return nil, fmt.Errorf("%s CreateCluster cluster is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("%s CreateCluster cluster opt or cloud is empty", cloudName)
	}

	if opt.Account == nil || len(opt.Account.SubscriptionID) == 0 || len(opt.Account.TenantID) == 0 ||
		len(opt.Account.ClientID) == 0 || len(opt.Account.ClientSecret) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("%s CreateCluster opt lost valid crendential info", cloudName)
	}

	// GetTaskManager for nodegroup manager initialization
	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CreateCluster %s failed, %s",
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

// CreateVirtualCluster create virtual cluster by cloud provider
func (c *Cluster) CreateVirtualCluster(cls *proto.Cluster,
	opt *cloudprovider.CreateVirtualClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteVirtualCluster delete virtual cluster
func (c *Cluster) DeleteVirtualCluster(cls *proto.Cluster,
	opt *cloudprovider.DeleteVirtualClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ImportCluster import cluster according cloudprovider
func (c *Cluster) ImportCluster(cls *proto.Cluster, opt *cloudprovider.ImportClusterOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when ImportCluster %s failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
	}

	if opt.Account == nil || len(opt.Account.SubscriptionID) == 0 || len(opt.Account.TenantID) == 0 ||
		len(opt.Account.ClientID) == 0 || len(opt.Account.ClientSecret) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("%s ImportCluster opt lost valid crendential info", cloudName)
	}

	// build import cluster task
	task, err := mgr.BuildImportClusterTask(cls, opt)
	if err != nil {
		blog.Errorf("build ImportCluster task for cluster %s with cloudprovider %s failed, %s",
			cls.ClusterName, cls.Provider, err.Error(),
		)
		return nil, err
	}

	return task, nil
}

// DeleteCluster delete kubenretes cluster according cloudprovider
func (c *Cluster) DeleteCluster(cls *proto.Cluster, opt *cloudprovider.DeleteClusterOption) (*proto.Task, error) {
	if cls == nil {
		return nil, fmt.Errorf("%s DeleteCluster cluster is empty", cloudName)
	}

	if opt.Account == nil || len(opt.Account.SubscriptionID) == 0 || len(opt.Account.TenantID) == 0 ||
		len(opt.Account.ClientID) == 0 || len(opt.Account.ClientSecret) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("%s DeleteCluster opt lost valid crendential info", cloudName)
	}

	// GetTaskManager for nodegroup manager initialization
	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteCluster %s failed, %s",
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
	return opt.Cluster, nil
}

// ListCluster get cloud cluster list by region
func (c *Cluster) ListCluster(opt *cloudprovider.ListClusterOption) ([]*proto.CloudClusterInfo, error) {
	client, err := api.NewAksServiceImplWithCommonOption(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("create azure client failed, %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	clusters, err := client.ListClusterByResourceGroupName(ctx, opt.Region, opt.ResourceGroupName)
	if err != nil {
		return nil, fmt.Errorf("list azure cluster failed, %v", err)
	}
	result := make([]*proto.CloudClusterInfo, 0)
	for _, v := range clusters {
		info := &proto.CloudClusterInfo{
			ClusterID:      *v.Name,
			ClusterName:    *v.Name,
			ClusterVersion: *v.Properties.CurrentKubernetesVersion,
			ClusterType:    *v.Type,
			ClusterStatus:  string(*v.Properties.PowerState.Code),
			Location:       *v.Location,
		}
		if len(v.Properties.AgentPoolProfiles) > 0 {
			p := v.Properties.AgentPoolProfiles
			info.ClusterOS = string(*p[0].OSSKU)
		}
		result = append(result, info)
	}
	return result, nil
}

// GetNodesInCluster get all nodes belong to cluster according cloudprovider
func (c *Cluster) GetNodesInCluster(cls *proto.Cluster, opt *cloudprovider.GetNodesOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// AddNodesToCluster add new node to cluster according cloudprovider
func (c *Cluster) AddNodesToCluster(
	cls *proto.Cluster, nodes []*proto.Node, opt *cloudprovider.AddNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteNodesFromCluster delete specified nodes from cluster according cloudprovider
func (c *Cluster) DeleteNodesFromCluster(
	cls *proto.Cluster, nodes []*proto.Node, opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterCidrAvailable check cluster CIDR nodesNum when add nodes
func (c *Cluster) CheckClusterCidrAvailable(
	cls *proto.Cluster, opt *cloudprovider.CheckClusterCIDROption) (bool, error) {
	return true, nil
}

// EnableExternalNodeSupport enable cluster support external node
func (c *Cluster) EnableExternalNodeSupport(cls *proto.Cluster, opt *cloudprovider.EnableExternalNodeOption) error {
	return nil
}

// ListOsImage list image os
func (c *Cluster) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	if opt == nil || opt.Account == nil {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	account := opt.Account
	if len(account.SubscriptionID) == 0 || len(account.TenantID) == 0 ||
		len(account.ClientID) == 0 || len(account.ClientSecret) == 0 {
		return nil, fmt.Errorf("azure ListOsImage lost authoration")
	}

	return utils.AKSImageOsList, nil
}

// AddSubnetsToCluster add subnets to cluster
func (c *Cluster) AddSubnetsToCluster(ctx context.Context, subnet *proto.SubnetSource,
	opt *cloudprovider.AddSubnetsToClusterOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// GetMasterSuggestedMachines get master suggested machines
func (c *Cluster) GetMasterSuggestedMachines(level, vpcId string,
	opt *cloudprovider.GetMasterSuggestedMachinesOption) ([]*proto.InstanceTemplateConfig, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListProjects list cloud projects
func (c *Cluster) ListProjects(opt *cloudprovider.CommonOption) ([]*proto.CloudProject, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterEndpointStatus check cluster endpoint status
func (c *Cluster) CheckClusterEndpointStatus(clusterID string, isExtranet bool,
	opt *cloudprovider.CheckEndpointStatusOption) (bool, error) {
	cli, err := api.NewAksServiceImplWithCommonOption(&opt.CommonOption)
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus create aks client failed, %v", err)
	}

	credentials, err := cli.GetClusterAdminCredentialsWithName(context.Background(), opt.ResourceGroupName, clusterID)
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus GetClusterAdminCredentialsWithName failed, %v", err)
	}
	if len(credentials) == 0 {
		return false, fmt.Errorf("credentials not found")
	}

	_, err = cloudprovider.GetCRDByKubeConfig(string(credentials[0].Value))
	if err != nil {
		return false, err
	}

	return true, nil
}

// AppendCloudNodeInfo append cloud node detailed info
func (c *Cluster) AppendCloudNodeInfo(ctx context.Context,
	nodes []*proto.ClusterNode, opt *cloudprovider.CommonOption) error {
	return nil
}

// CheckIfGetNodesFromCluster check cluster if can get nodes from k8s
func (c *Cluster) CheckIfGetNodesFromCluster(ctx context.Context, cluster *proto.Cluster,
	nodes []*proto.ClusterNode) bool {
	return true
}

// SwitchClusterNetwork switch cluster network mode
func (c *Cluster) SwitchClusterNetwork(
	cls *proto.Cluster, subnet *proto.SubnetSource, opt *cloudprovider.SwitchClusterNetworkOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterNetworkStatus get cluster network
func (c *Cluster) CheckClusterNetworkStatus(cloudID string,
	opt *cloudprovider.CheckClusterNetworkStatusOption) (bool, error) {
	return false, cloudprovider.ErrCloudNotImplemented
}
