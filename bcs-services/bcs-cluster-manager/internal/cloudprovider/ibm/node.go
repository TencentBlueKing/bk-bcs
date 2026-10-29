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
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

var nodeMgr sync.Once

func init() {
	nodeMgr.Do(func() {
		cloudprovider.InitNodeManager(cloudName, &NodeManager{})
	})
}

// NodeManager CVM relative API management
type NodeManager struct{}

// GetNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (
	*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListNodesByIP list node by IP set
func (nm *NodeManager) ListNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) (
	[]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetExternalNodeByIP get specified Node by innerIP address
func (nm *NodeManager) GetExternalNodeByIP(ip string, opt *cloudprovider.GetNodeOption) (
	*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListExternalNodesByIP list node by IP set
func (nm *NodeManager) ListExternalNodesByIP(ips []string, opt *cloudprovider.ListNodesOption) (
	[]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetCVMImageIDByImageName get imageID by imageName
func (nm *NodeManager) GetCVMImageIDByImageName(imageName string,
	opt *cloudprovider.CommonOption) (string, error) {
	return "", cloudprovider.ErrCloudNotImplemented
}

// GetCloudRegions get cloud regions
func (nm *NodeManager) GetCloudRegions(opt *cloudprovider.CommonOption) (
	[]*proto.RegionInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetZoneList get zoneList by region
func (nm *NodeManager) GetZoneList(opt *cloudprovider.GetZoneListOption) (
	[]*proto.ZoneInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListNodeInstanceType get node instance type list
func (nm *NodeManager) ListNodeInstanceType(ctx context.Context, info cloudprovider.InstanceInfo,
	opt *cloudprovider.CommonOption) ([]*proto.InstanceType, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListDiskTypes get disk type list
func (nm *NodeManager) ListDiskTypes(instanceTypes []string, zones []string, diskChargeType string,
	cpu, memory uint64, opt *cloudprovider.CommonOption) ([]*proto.DiskConfigSet, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListOsImage get osimage list
func (nm *NodeManager) ListOsImage(provider string,
	opt *cloudprovider.CommonOption) ([]*proto.OsImage, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListKeyPairs list ssh keyPairs
func (nm *NodeManager) ListKeyPairs(opt *cloudprovider.ListNetworksOption) (
	[]*proto.KeyPair, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetResourceGroups resource groups list
func (nm *NodeManager) GetResourceGroups(opt *cloudprovider.CommonOption) (
	[]*proto.ResourceGroupInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListRuntimeInfo get runtime info list
func (nm *NodeManager) ListRuntimeInfo(opt *cloudprovider.ListRuntimeInfoOption) (
	map[string][]string, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetServiceRoles service roles list
func (nm *NodeManager) GetServiceRoles(opt *cloudprovider.CommonOption, roleType string) (
	[]*proto.ServiceRoleInfo, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListNodePublicPrefixs list node public prefix
func (nm *NodeManager) ListNodePublicPrefixs(opt *cloudprovider.ListNodePublicPrefixesOption) (
	[]*proto.NodePublicPrefix, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListImportClusterNodes 根据节点IP列表从k8s节点信息获取导入集群的节点信息
// IBM 通过k8s node label获取zone/region/instance-type/nodeID等信息
func (nm *NodeManager) ListImportClusterNodes(ips []string, opt *cloudprovider.ListNodesOption) (
	[]*proto.Node, error) {
	if len(ips) == 0 {
		return nil, nil
	}

	// 从k8s集群获取节点列表
	clusterID := opt.ClusterID
	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	k8sNodes, err := k8sOperator.ListClusterNodes(context.Background(), clusterID)
	if err != nil {
		return nil, fmt.Errorf("ListClusterNodes[%s] failed: %v", clusterID, err)
	}

	if len(k8sNodes) == 0 {
		return nil, fmt.Errorf("no k8s nodes found in cluster[%s]", clusterID)
	}

	// 使用IBM VPC专属label配置
	labelCfg := getLabelConfig()

	// IBM VPC providerID格式不统一，优先从label ibm-cloud.kubernetes.io/vpc-instance-id获取实例ID
	nodes := cloudprovider.K8sNodesToImportNodes(k8sNodes, ips, labelCfg,
		parseIBMInstanceID)

	// 补充NodeID：优先从IBM专属label获取vpc-instance-id，其次尝试worker-id
	for i, k8sNode := range k8sNodes {
		if i >= len(nodes) {
			break
		}
		if nodes[i].NodeID != "" {
			continue
		}
		// 优先使用vpc-instance-id
		if vpcID, ok := k8sNode.Labels[IBMVpcInstanceIDLabel]; ok && vpcID != "" {
			nodes[i].NodeID = vpcID
			continue
		}
		// 回退到worker-id
		if workerID, ok := k8sNode.Labels[IBMWorkerIDLabel]; ok && workerID != "" {
			nodes[i].NodeID = workerID
		}
	}

	blog.Infof("ListImportClusterNodes from k8s labels, ips[%d] matched[%d]", len(ips), len(nodes))
	return nodes, nil
}
