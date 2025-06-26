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

package blueking

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
)

func init() {
	cloudprovider.InitNodeGroupManager(cloudName, &NodeGroup{})
}

// NodeGroup nodegroup management for blueking resource pool solution
type NodeGroup struct {
}

// CreateNodeGroup create nodegroup by cloudprovider api, only create NodeGroup entity
func (ng *NodeGroup) CreateNodeGroup(group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (
	*proto.Task, error) {
	if opt.OnlyData {
		return nil, nil
	}

	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		return nil, err
	}
	task, err := mgr.BuildCreateNodeGroupTask(group, opt)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// DeleteNodeGroup delete nodegroup by cloudprovider api, all nodes belong to NodeGroup
// will be released. Task is backgroup automatic task
func (ng *NodeGroup) DeleteNodeGroup(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodeGroupOption) (*proto.Task, error) {
	// validate request
	if group == nil {
		return nil, fmt.Errorf("lost clean nodes or group")
	}

	if opt == nil || len(opt.Region) == 0 || opt.Account == nil ||
		len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 || opt.Cloud == nil {
		return nil, fmt.Errorf("lost connect cloud_provider auth information")
	}
	if opt.OnlyData {
		return nil, nil
	}

	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		return nil, err
	}
	task, err := mgr.BuildDeleteNodeGroupTask(group, nodes, opt)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// UpdateNodeGroup update specified nodegroup configuration
func (ng *NodeGroup) UpdateNodeGroup(
	group *proto.NodeGroup, opt *cloudprovider.UpdateNodeGroupOption) (*proto.Task, error) {
	if group == nil || opt == nil {
		return nil, fmt.Errorf("UpdateNodeGroup group or opt is nil")
	}
	if opt.Cluster == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("UpdateNodeGroup lost validate data")
	}
	// only update nodegroup data, not build task
	if opt.OnlyData {
		return nil, nil
	}

	if group.NodeGroupID == "" || group.ClusterID == "" {
		return nil, fmt.Errorf("nodegroup id or cluster id is empty")
	}

	ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(), tenant.ResourceMetaData{
		ProjectId: group.ProjectID,
	})
	if err != nil {
		return nil, err
	}

	err = cloudprovider.UpdateNodeGroupCloudAndModuleInfo(ctx, group.NodeGroupID, group.ConsumerID,
		true, opt.Cluster.BusinessID)
	if err != nil {
		return nil, err
	}

	// build task
	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		return nil, err
	}
	task, err := mgr.BuildUpdateNodeGroupTask(group, &opt.CommonOption)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// RecommendNodeGroupConf recommends nodegroup configs
func (ng *NodeGroup) RecommendNodeGroupConf(
	ctx context.Context, opt *cloudprovider.CommonOption) ([]*proto.RecommendNodeGroupConf, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetNodesInGroup get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroup(group *proto.NodeGroup, opt *cloudprovider.CommonOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetNodesInGroupV2 get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroupV2(group *proto.NodeGroup,
	opt *cloudprovider.CommonOption) ([]*proto.NodeGroupNode, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// MoveNodesToGroup add cluster nodes to NodeGroup
func (ng *NodeGroup) MoveNodesToGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// RemoveNodesFromGroup remove nodes from NodeGroup, nodes are still in cluster
func (ng *NodeGroup) RemoveNodesFromGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.RemoveNodesOption) error {
	return cloudprovider.ErrCloudNotImplemented
}

// CleanNodesInGroup clean specified nodes in NodeGroup,
func (ng *NodeGroup) CleanNodesInGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.CleanNodesOption) (*proto.Task, error) {
	// validate request
	if len(nodes) == 0 || group == nil {
		return nil, fmt.Errorf("lost clean nodes or group")
	}
	if opt == nil || opt.Cluster == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("lost cluster or cloud information")
	}

	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CleanNodesInGroup %s failed, %s",
			cloudName, group.Name, err.Error())
		return nil, err
	}
	task, err := mgr.BuildCleanNodesInGroupTask(nodes, group, opt)
	if err != nil {
		blog.Errorf("build CleanNodesInGroup task for cluster %s with cloudprovider %s failed, %s",
			group.ClusterID, cloudName, err.Error())
		return nil, err
	}
	return task, nil
}

// UpdateDesiredNodes update nodegroup desired node
func (ng *NodeGroup) UpdateDesiredNodes(desired uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (*cloudprovider.ScalingResponse, error) {
	if group == nil || opt == nil || opt.Cluster == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("invalid request")
	}

	// scaling nodes with desired, first get all node for status filtering
	// check if nodes are already in cluster
	goodNodes, err := cloudprovider.ListNodesInClusterNodePool(opt.Cluster.ClusterID, group.NodeGroupID)
	if err != nil {
		blog.Errorf("cloudprovider qcloud get NodeGroup %s all Nodes failed, %s", group.NodeGroupID, err.Error())
		return nil, err
	}

	// check incoming nodes
	inComingNodes, err := cloudprovider.GetNodesNumWhenApplyInstanceTask(opt.Cluster.ClusterID, group.NodeGroupID,
		cloudprovider.GetTaskType(opt.Cloud.CloudProvider, cloudprovider.UpdateNodeGroupDesiredNode),
		cloudprovider.TaskStatusRunning,
		[]string{cloudprovider.GetTaskType(opt.Cloud.CloudProvider, cloudprovider.ApplyInstanceMachinesTask)})
	if err != nil {
		blog.Errorf("UpdateDesiredNodes GetNodesNumWhenApplyInstanceTask failed: %v", err)
		return nil, err
	}

	// cluster current node
	current := len(goodNodes) + inComingNodes

	nodeNames := make([]string, 0)
	for _, node := range goodNodes {
		nodeNames = append(nodeNames, node.InnerIP)
	}
	blog.Infof("NodeGroup %s has total nodes %d, current capable nodes %d, current incoming nodes %d, "+
		"desired nodes %d, details %v", group.NodeGroupID, len(goodNodes), current, inComingNodes, desired, nodeNames)

	if current >= int(desired) {
		blog.Infof("NodeGroup %s current capable nodes %d larger than desired %d nodes, nothing to do",
			group.NodeGroupID, current, desired)
		return &cloudprovider.ScalingResponse{
				ScalingUp:    0,
				CapableNodes: nodeNames,
			}, fmt.Errorf("NodeGroup %s UpdateDesiredNodes nodes %d larger than desired %d nodes",
				group.NodeGroupID, current, desired)
	}

	// current scale nodeNum
	scalingUp := int(desired) - current

	return &cloudprovider.ScalingResponse{
		ScalingUp:    uint32(scalingUp),
		CapableNodes: nodeNames,
	}, nil
}

// SwitchNodeGroupAutoScaling switch nodegroup autoscaling
func (ng *NodeGroup) SwitchNodeGroupAutoScaling(group *proto.NodeGroup, enable bool,
	opt *cloudprovider.SwitchNodeGroupAutoScalingOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when SwitchNodeGroupAutoScaling %s failed, %s",
			cloudName, group.NodeGroupID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildSwitchNodeGroupAutoScalingTask(group, enable, opt)
	if err != nil {
		blog.Errorf("build SwitchNodeGroupAutoScaling task for nodeGroup %s with cloudprovider %s failed, %s",
			group.NodeGroupID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// CreateAutoScalingOption create cluster autoscaling option, cloudprovider will
// deploy cluster-autoscaler in backgroup according cloudprovider implementation
func (ng *NodeGroup) CreateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.CreateScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteAutoScalingOption delete cluster autoscaling, cloudprovider will clean
// cluster-autoscaler in backgroup according cloudprovider implementation
func (ng *NodeGroup) DeleteAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.DeleteScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// UpdateAutoScalingOption update cluster autoscaling option, cloudprovider will update
// cluster-autoscaler configuration in backgroup according cloudprovider implementation.
// Implementation is optional.
func (ng *NodeGroup) UpdateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.UpdateScalingOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		return nil, err
	}

	ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(), tenant.ResourceMetaData{
		ProjectId: scalingOption.ProjectID,
	})
	if err != nil {
		return nil, err
	}

	err = cloudprovider.UpdateAutoScalingOptionModuleInfo(ctx, scalingOption.ClusterID)
	if err != nil {
		return nil, err
	}

	task, err := mgr.BuildUpdateAutoScalingOptionTask(scalingOption, opt)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// SwitchAutoScalingOptionStatus switch cluster autoscaling option status
func (ng *NodeGroup) SwitchAutoScalingOptionStatus(scalingOption *proto.ClusterAutoScalingOption, enable bool,
	opt *cloudprovider.CommonOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when SwitchAutoScalingOptionStatus %s failed, %s",
			cloudName, scalingOption.ClusterID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildSwitchAsOptionStatusTask(scalingOption, enable, opt)
	if err != nil {
		blog.Errorf("build SwitchAutoScalingOptionStatus task for cluster %s with cloudprovider %s failed, %s",
			scalingOption.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// AddExternalNodeToCluster add external to cluster
func (ng *NodeGroup) AddExternalNodeToCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.AddExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteExternalNodeFromCluster remove external node from cluster
func (ng *NodeGroup) DeleteExternalNodeFromCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteExternalNodesOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetExternalNodeScript get nodegroup external node script
func (ng *NodeGroup) GetExternalNodeScript(group *proto.NodeGroup, internal bool) (string, error) {
	return "", cloudprovider.ErrCloudNotImplemented
}

// CheckResourcePoolQuota check resource pool quota when revise group limit
func (ng *NodeGroup) CheckResourcePoolQuota(
	ctx context.Context, group *proto.NodeGroup, operation string, scaleUpNum uint32) error {
	return nil
}

// GetProjectCaResourceQuota get project ca resource quota
func (ng *NodeGroup) GetProjectCaResourceQuota(groups []*proto.NodeGroup,
	opt *cloudprovider.CommonOption) ([]*proto.ProjectAutoscalerQuota, error) {
	return nil, nil
}
