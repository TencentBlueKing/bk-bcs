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

package google

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var clusterMgr sync.Once

func init() {
	clusterMgr.Do(func() {
		cloudprovider.InitClusterManager(cloudName, &Cluster{})
	})
}

// Cluster kubernetes cluster management implementation
type Cluster struct {
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

// CreateCluster create kubenretes cluster according cloudprovider
func (c *Cluster) CreateCluster(cls *proto.Cluster, opt *cloudprovider.CreateClusterOption) (*proto.Task, error) {
	// call google interface to create cluster
	if cls == nil {
		return nil, fmt.Errorf("%s CreateCluster cluster is empty", cloudName)
	}

	if opt == nil || opt.Account == nil || len(opt.Account.ServiceAccountSecret) == 0 ||
		len(opt.Account.GkeProjectID) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("google CreateCluster lost authoration")
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

	if opt == nil || opt.Account == nil || len(opt.Account.ServiceAccountSecret) == 0 ||
		len(opt.Account.GkeProjectID) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("google DeleteCluster lost authoration")
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

	if opt.Cluster.ClusterAdvanceSettings.NetworkType == "" {
		opt.Cluster.ClusterAdvanceSettings.NetworkType = api.NetworkPolicyProviderCalico
	}

	if opt.Cluster.NetworkType == "" {
		opt.Cluster.NetworkType = common.ClusterOverlayNetwork
	}

	return opt.Cluster, nil
}

// ListCluster get cloud cluster list by region
func (c *Cluster) ListCluster(opt *cloudprovider.ListClusterOption) ([]*proto.CloudClusterInfo, error) {
	client, err := api.NewContainerServiceClient(&opt.CommonOption)
	if err != nil {
		return nil, fmt.Errorf("create google client failed, err %s", err.Error())
	}
	clusters, err := client.ListCluster(context.Background(), "-")
	if err != nil {
		return nil, fmt.Errorf("list google cluster failed, err %s", err.Error())
	}

	result := make([]*proto.CloudClusterInfo, 0)
	for _, v := range clusters {
		if strings.Contains(v.Location, opt.Region) {
			info := &proto.CloudClusterInfo{
				ClusterID:      v.Name,
				ClusterName:    v.Name,
				ClusterVersion: v.CurrentMasterVersion,
				ClusterStatus:  v.Status,
				ClusterType:    api.Standard,
				Location:       v.Location,
			}
			if v.NodeConfig != nil {
				info.ClusterOS = v.NodeConfig.ImageType
			}
			if v.Autopilot != nil && v.Autopilot.Enabled {
				info.ClusterType = api.Autopilot
			}
			if len(strings.Split(v.Location, "-")) == 2 {
				info.ClusterLevel = api.RegionLevel
			} else {
				info.ClusterLevel = api.ZoneLevel
			}

			result = append(result, info)
		}
	}

	return result, nil
}

// GetNodesInCluster get all nodes belong to cluster according cloudprovider
func (c *Cluster) GetNodesInCluster(cls *proto.Cluster, opt *cloudprovider.GetNodesOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// AddNodesToCluster add new node to cluster according cloudprovider
func (c *Cluster) AddNodesToCluster(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.AddNodesOption) ([]*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteNodesFromCluster delete specified nodes from cluster according cloudprovider
func (c *Cluster) DeleteNodesFromCluster(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckClusterCidrAvailable check cluster CIDR nodesNum when add nodes
func (c *Cluster) CheckClusterCidrAvailable(cls *proto.Cluster, opt *cloudprovider.CheckClusterCIDROption) (bool,
	error) {
	return true, nil
}

// EnableExternalNodeSupport enable cluster support external node
func (c *Cluster) EnableExternalNodeSupport(cls *proto.Cluster, opt *cloudprovider.EnableExternalNodeOption) error {
	return nil
}

// ListOsImage get osi  mage list
func (c *Cluster) ListOsImage(provider, clusterID string, opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.ServiceAccountSecret) == 0 ||
		len(opt.Account.GkeProjectID) == 0 || len(opt.Region) == 0 {
		return nil, fmt.Errorf("google ListOsImage lost authoration")
	}

	return utils.GkeImageOsList, nil
}

// ListProjects list cloud projects
func (c *Cluster) ListProjects(opt *cloudprovider.CommonOption) ([]*proto.CloudProject, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
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

// CheckClusterEndpointStatus check cluster endpoint status
func (c *Cluster) CheckClusterEndpointStatus(clusterID string, isExtranet bool,
	opt *cloudprovider.CheckEndpointStatusOption) (bool, error) {

	gkeCli, err := api.NewContainerServiceClient(&opt.CommonOption)
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus get gke client failed, %v", err)
	}

	gkeCluster, err := gkeCli.GetCluster(context.Background(), clusterID)
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus get cluster failed, %v", err)
	}

	cert, err := base64.StdEncoding.DecodeString(gkeCluster.MasterAuth.ClusterCaCertificate)
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus get cluster certificate failed, %v", err)
	}

	restConfig := &rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			CAData: cert,
		},
		Host: "https://" + gkeCluster.Endpoint,
		AuthProvider: &clientcmdapi.AuthProviderConfig{
			Name: api.GoogleAuthPlugin,
			Config: map[string]string{
				"scopes":      "https://www.googleapis.com/auth/cloud-platform",
				"credentials": opt.CommonOption.Account.ServiceAccountSecret,
			},
		},
	}
	cs, err := clientset.NewForConfig(restConfig)
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus create clientset failed: %v", err)
	}

	// 获取 CRD
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	_, err = cs.ApiextensionsV1().CustomResourceDefinitions().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("CheckClusterEndpointStatus failed: %v", err)
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

// UpdateCloudKubeConfig update cluster kubeconfig to clustercredential
func (c *Cluster) UpdateCloudKubeConfig(kubeConfig string,
	opt *cloudprovider.UpdateCloudKubeConfigOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// CheckHighAvailabilityMasterNodes check master nodes high availability
func (c *Cluster) CheckHighAvailabilityMasterNodes(cls *proto.Cluster, nodes []*proto.Node,
	opt *cloudprovider.CheckHaMasterNodesOption) error {
	return nil
}
