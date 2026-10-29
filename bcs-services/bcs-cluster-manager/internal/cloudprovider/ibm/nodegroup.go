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

var nodeGroupMgr sync.Once

func init() {
	nodeGroupMgr.Do(func() {
		cloudprovider.InitNodeGroupManager(cloudName, &NodeGroup{})
	})
}

// NodeGroup nodegroup manager for ibm
type NodeGroup struct{}

// CreateNodeGroup create nodegroup by cloudprovider api
func (n *NodeGroup) CreateNodeGroup(group *proto.NodeGroup,
	opt *cloudprovider.CreateNodeGroupOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteNodeGroup delete nodegroup by cloudprovider api
func (n *NodeGroup) DeleteNodeGroup(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodeGroupOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// UpdateNodeGroup update specified nodegroup configuration
func (n *NodeGroup) UpdateNodeGroup(group *proto.NodeGroup,
	opt *cloudprovider.UpdateNodeGroupOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// RecommendNodeGroupConf recommends nodegroup configs
func (n *NodeGroup) RecommendNodeGroupConf(ctx context.Context,
	opt *cloudprovider.CommonOption) ([]*proto.RecommendNodeGroupConf, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetNodesInGroup get all nodes belong to NodeGroup
func (n *NodeGroup) GetNodesInGroup(group *proto.NodeGroup,
	opt *cloudprovider.CommonOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetNodesInGroupV2 get all nodes belong to NodeGroup
func (n *NodeGroup) GetNodesInGroupV2(group *proto.NodeGroup,
	opt *cloudprovider.CommonOption) ([]*proto.NodeGroupNode, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// MoveNodesToGroup add cluster nodes to NodeGroup
func (n *NodeGroup) MoveNodesToGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CheckResourcePoolQuota check resource pool quota
func (n *NodeGroup) CheckResourcePoolQuota(ctx context.Context, group *proto.NodeGroup,
	operation string, scaleUpNum uint32) error {
	return cloudprovider.ErrCloudNotImplemented
}

// RemoveNodesFromGroup remove nodes from NodeGroup
func (n *NodeGroup) RemoveNodesFromGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.RemoveNodesOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// CleanNodesInGroup clean specified nodes in NodeGroup
func (n *NodeGroup) CleanNodesInGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.CleanNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// UpdateDesiredNodes update nodegroup desired node
func (n *NodeGroup) UpdateDesiredNodes(desired uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (*cloudprovider.ScalingResponse, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// SwitchNodeGroupAutoScaling switch nodegroup auto scale
func (n *NodeGroup) SwitchNodeGroupAutoScaling(group *proto.NodeGroup, enable bool,
	opt *cloudprovider.SwitchNodeGroupAutoScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// CreateAutoScalingOption create cluster autoscaling option
func (n *NodeGroup) CreateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.CreateScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteAutoScalingOption delete cluster autoscaling
func (n *NodeGroup) DeleteAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.DeleteScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// UpdateAutoScalingOption update cluster autoscaling option
func (n *NodeGroup) UpdateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.UpdateScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// SwitchAutoScalingOptionStatus switch cluster autoscaling option enable auto scaling status
func (n *NodeGroup) SwitchAutoScalingOptionStatus(scalingOption *proto.ClusterAutoScalingOption,
	enable bool, opt *cloudprovider.CommonOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// AddExternalNodeToCluster add external to cluster
func (n *NodeGroup) AddExternalNodeToCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.AddExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteExternalNodeFromCluster remove external node from cluster
func (n *NodeGroup) DeleteExternalNodeFromCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetExternalNodeScript get external node script from cluster nodeGroup
func (n *NodeGroup) GetExternalNodeScript(group *proto.NodeGroup, internal bool) (string, error) {
	return "", cloudprovider.ErrCloudNotImplemented
}

// GetProjectResourceQuota get project autoscaler quota
func (n *NodeGroup) GetProjectResourceQuota(groups []*proto.NodeGroup, resourcePoolType string,
	opt *cloudprovider.CommonOption) ([]*proto.ProjectAutoscalerQuota, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
