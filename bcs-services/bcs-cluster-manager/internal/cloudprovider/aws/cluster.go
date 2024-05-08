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

// Package aws xxx
package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

func init() {
	cloudprovider.InitClusterManager("aws", &Cluster{})
}

// Cluster aws kubernetes cluster management implementation
type Cluster struct {
}

// CreateCluster create kubenretes cluster according cloudprovider
func (c *Cluster) CreateCluster(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
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
	// call aws interface to create cluster
	if cls == nil {
		return nil, fmt.Errorf("awsCloud ImportCluster cluster is empty")
	}

	if opt == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("awsCloud ImportCluster cluster opt or cloud is empty")
	}

	if len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("awsCloud CreateCluster opt lost valid crendential info")
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
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetCluster get kubenretes cluster detail information according cloudprovider
func (c *Cluster) GetCluster(cloudID string, opt *cloudprovider.GetClusterOption) (*proto.Cluster, error) {
	return opt.Cluster, nil
	runtimeInfo, err := business.GetRuntimeInfo(opt.Cluster.ClusterID)
	if err != nil {
		return nil, err
	}

	if opt.Cluster.ClusterAdvanceSettings == nil {
		opt.Cluster.ClusterAdvanceSettings = &proto.ClusterAdvanceSetting{}
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
		return nil, fmt.Errorf("%s ListCluster cluster lost operation", cloudName)
	}

	cli, err := api.NewEksClient(&opt.CommonOption)
	if err != nil {
		return nil, err
	}

	clusters, err := cli.ListEksCluster()
	if err != nil {
		return nil, err
	}

	cloudClusterList := make([]*proto.CloudClusterInfo, 0)
	for _, v := range clusters {
		cluster, err := cli.GetEksCluster(*v)
		if err != nil {
			return nil, err
		}

		cloudClusterList = append(cloudClusterList, &proto.CloudClusterInfo{
			ClusterID:      *v,
			ClusterName:    *v,
			ClusterStatus:  *cluster.Status,
			ClusterVersion: *cluster.Version,
			Location:       opt.CommonOption.Region,
		})
	}

	return cloudClusterList, nil
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

// CheckClusterCidrAvailable check cluster CIDR nodesNum when add nodes
func (c *Cluster) CheckClusterCidrAvailable(cls *proto.Cluster,
	opt *cloudprovider.CheckClusterCIDROption) (bool, error) {
	if cls == nil || opt == nil {
		return true, nil
	}

	return true, nil
}

// EnableExternalNodeSupport enable cluster support external node
func (c *Cluster) EnableExternalNodeSupport(cls *proto.Cluster, opt *cloudprovider.EnableExternalNodeOption) error {
	return nil
}

// ListOsImage get osimage list
func (c *Cluster) ListOsImage(provider string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 ||
		len(opt.Account.SecretKey) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("awsCloud ListOsImage lost authoration")
	}

	return utils.EKSImageOsList, nil
}

// ListProjects list cloud projects
func (c *Cluster) ListProjects(opt *cloudprovider.CommonOption) ([]*proto.CloudProject, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterEndpointStatus check cluster endpoint status
func (c *Cluster) CheckClusterEndpointStatus(clusterID string, isExtranet bool,
	opt *cloudprovider.CheckEndpointStatusOption) (bool, error) {
	client, err := api.NewEksClient(&opt.CommonOption)
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus get NewEksClient failed: %v", err)
	}

	cluster, err := client.GetEksCluster(clusterID)
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus GetEksCluster failed: %v", err)
	}

	restConfig, err := api.GenerateAwsRestConf(&opt.CommonOption, cluster)
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus GenerateAwsRestConf failed: %v", err)
	}

	clientSet, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus get clientset failed: %v", err)
	}

	// 获取 CRD
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	_, err = clientSet.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus failed: %v", err)
	}

	return true, nil
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
