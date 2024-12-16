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

// Package huawei xxx
package huawei

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

func init() {
	cloudprovider.InitClusterManager(cloudName, &Cluster{})
}

// Cluster blueking kubernetes cluster management implementation
type Cluster struct {
}

// CreateCluster create kubenretes cluster according cloudprovider
func (c *Cluster) CreateCluster(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
	// call qcloud interface to create cluster
	if cls == nil {
		return nil, fmt.Errorf("%s CreateCluster cluster is empty", cloudName)
	}

	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("%s CreateCluster cluster opt or cloud is empty", cloudName)
	}

	if opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("%s CreateCluster opt lost valid crendential info", cloudName)
	}

	// GetTaskManager for nodegroup manager initialization
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
func (c *Cluster) ImportCluster(cls *proto.Cluster, opt *cloudprovider.ImportClusterOption) (*proto.Task, error) {
	// call cce interface to create cluster
	if cls == nil {
		return nil, fmt.Errorf("huawei ImportCluster cluster is empty")
	}

	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("huawei ImportCluster cluster opt or cloud is empty")
	}

	if len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("huawei CreateCluster opt lost valid crendential info")
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when ImportCluster %d failed, %s",
			opt.Cloud.CloudID, cls.ClusterName, err.Error(),
		)
		return nil, err
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

	if opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
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

// GetCluster get kubenretes cluster detail information according cloudprovider
func (c *Cluster) GetCluster(cloudID string, opt *cloudprovider.GetClusterOption) (*proto.Cluster, error) {
	runtimeInfo, err := business.GetRuntimeInfo(opt.Cluster.ClusterID)
	if err != nil {
		return nil, err
	}

	if v, ok := runtimeInfo[common.ContainerdRuntime]; ok {
		opt.Cluster.ClusterAdvanceSettings.ContainerRuntime = common.ContainerdRuntime
		if len(v) > 0 {
			opt.Cluster.ClusterAdvanceSettings.RuntimeVersion = v[0]
		}
	} else if v, ok := runtimeInfo[common.DockerContainerRuntime]; ok {
		opt.Cluster.ClusterAdvanceSettings.ContainerRuntime = common.DockerContainerRuntime
		if len(v) > 0 {
			opt.Cluster.ClusterAdvanceSettings.RuntimeVersion = v[0]
		}
	}

	return opt.Cluster, nil
}

// ListCluster get cloud cluster list by region
func (c *Cluster) ListCluster(opt *cloudprovider.ListClusterOption) ([]*proto.CloudClusterInfo, error) {
	if opt == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("huawei ListCluster cluster lost operation")
	}

	cli, err := api.NewCceClient(&opt.CommonOption)
	if err != nil {
		return nil, err
	}

	clusters, err := cli.ListCceCluster(nil)
	if err != nil {
		return nil, err
	}

	cloudClusterList := make([]*proto.CloudClusterInfo, 0)
	for _, v := range *clusters {
		cloudClusterList = append(cloudClusterList, &proto.CloudClusterInfo{
			ClusterID:          *v.Metadata.Uid,
			ClusterName:        v.Metadata.Name,
			ClusterDescription: *v.Metadata.Alias,
			ClusterVersion:     *v.Spec.Version,
			ClusterOS:          v.Spec.Type.Value(),
			ClusterStatus:      *v.Status.Phase,
			Location:           opt.Region,
		})
	}

	return cloudClusterList, nil
}

// AddSubnetsToCluster cluster add subnet
func (c *Cluster) AddSubnetsToCluster(ctx context.Context, subnet *proto.SubnetSource,
	opt *cloudprovider.AddSubnetsToClusterOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// AppendCloudNodeInfo append cloud node detailed info
func (c *Cluster) AppendCloudNodeInfo(ctx context.Context,
	nodes []*proto.ClusterNode, opt *cloudprovider.CommonOption) error {

	// 获取语言
	lang := i18n.LanguageFromCtx(ctx)
	if lang == utils.ZH {
		for i, node := range nodes {
			if node.ZoneID != "" {
				nodes[i].ZoneName = fmt.Sprintf("可用区%d", business.GetZoneNameByZoneId(opt.Region, node.ZoneID))
			}
		}
	}

	return nil
}

// CheckClusterEndpointStatus check cluster endpoint status
func (c *Cluster) CheckClusterEndpointStatus(clusterID string, isExtranet bool,
	opt *cloudprovider.CheckEndpointStatusOption) (bool, error) {
	if opt == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return false, fmt.Errorf("qcloud CheckClusterEndpointStatus lost authoration")
	}

	cli, err := api.NewCceClient(&opt.CommonOption)
	if err != nil {
		return false, err
	}

	cluster, err := cli.GetCceCluster(clusterID)
	if err != nil {
		return false, err
	}
	if *cluster.Status.Phase != api.Available {
		return false, fmt.Errorf("cluster status is not available")
	}

	kubeConfig, err := cli.GetClusterKubeConfig(clusterID, isExtranet)
	if err != nil {
		return false, err
	}

	_, err = cloudprovider.GetCRDByKubeConfig(kubeConfig)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetNodesInCluster get all nodes belong to cluster according cloudprovider
func (c *Cluster) GetNodesInCluster(cls *proto.Cluster, opt *cloudprovider.GetNodesOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// AddNodesToCluster add new node to cluster according cloudprovider
func (c *Cluster) AddNodesToCluster(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.AddNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteNodesFromCluster delete specified nodes from cluster according cloudprovider
func (c *Cluster) DeleteNodesFromCluster(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckIfGetNodesFromCluster check cluster if can get nodes from k8s
func (c *Cluster) CheckIfGetNodesFromCluster(ctx context.Context, cluster *proto.Cluster,
	nodes []*proto.ClusterNode) bool {
	return true
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

// EnableExternalNodeSupport enable cluster support external node
func (c *Cluster) EnableExternalNodeSupport(cls *proto.Cluster, opt *cloudprovider.EnableExternalNodeOption) error {
	return nil
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

// CheckClusterCidrAvailable check cluster CIDR nodesNum when add nodes
func (c *Cluster) CheckClusterCidrAvailable(cls *proto.Cluster, opt *cloudprovider.CheckClusterCIDROption) (bool,
	error) {
	return true, nil
}

// ListOsImage list image os
func (c *Cluster) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("huawei ListOsImage lost authoration")
	}

	return utils.CCEImageOsList, nil
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

// UpdateCloudKubeConfig update cluster kubeconfig to clustercredential
func (c *Cluster) UpdateCloudKubeConfig(kubeConfig string,
	opt *cloudprovider.UpdateCloudKubeConfigOption) error {
	return cloudprovider.ErrCloudNotImplemented
}
