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

package aws

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
)

func init() {
	cloudprovider.InitNodeGroupManager("aws", &NodeGroup{})
}

// NodeGroup nodegroup management for blueking resource pool solution
type NodeGroup struct {
}

// CreateNodeGroup create nodegroup by cloudprovider api, only create NodeGroup entity
func (ng *NodeGroup) CreateNodeGroup(group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (*proto.Task,
	error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CreateNodeGroup %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildCreateNodeGroupTask(group, opt)
	if err != nil {
		blog.Errorf("build CreateNodeGroup task for cluster %s with cloudprovider %s failed, %s",
			group.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// DeleteNodeGroup delete nodegroup by cloudprovider api, all nodes belong to NodeGroup
// will be released. Task is backgroup automatic task
func (ng *NodeGroup) DeleteNodeGroup(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteNodeGroupOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteNodeGroup %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildDeleteNodeGroupTask(group, nodes, opt)
	if err != nil {
		blog.Errorf("build DeleteNodeGroup task for cluster %s with cloudprovider %s failed, %s",
			group.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// UpdateNodeGroup update specified nodegroup configuration
func (ng *NodeGroup) UpdateNodeGroup(group *proto.NodeGroup, opt *cloudprovider.CommonOption) error {
	_, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("get cluster %s failed, %s", group.ClusterID, err.Error())
		return err
	}
	eksCli, err := api.NewEksClient(opt)
	if err != nil {
		blog.Errorf("create eks client failed, err: %s", err.Error())
		return err
	}
	if group.NodeGroupID == "" || group.ClusterID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return fmt.Errorf("nodegroup id or cluster id is empty")
	}
	_, err = eksCli.UpdateNodegroupConfig(ng.generateUpdateNodegroupConfigInput(group, cluster.SystemID))
	if err != nil {
		return err
	}
	return nil
}

func (ng *NodeGroup) generateUpdateNodegroupConfigInput(group *proto.NodeGroup,
	cluster string) *eks.UpdateNodegroupConfigInput {
	input := &eks.UpdateNodegroupConfigInput{
		ClusterName:   &cluster,
		NodegroupName: &group.CloudNodeGroupID,
		Labels: &eks.UpdateLabelsPayload{
			AddOrUpdateLabels: aws.StringMap(group.Labels),
		},
	}
	if group.AutoScaling != nil {
		input.ScalingConfig = &eks.NodegroupScalingConfig{
			MaxSize: aws.Int64(int64(group.AutoScaling.MaxSize)),
			MinSize: aws.Int64(int64(group.AutoScaling.MinSize)),
		}
	}
	if group.NodeTemplate != nil && group.NodeTemplate.Taints != nil {
		input.Taints = &eks.UpdateTaintsPayload{
			AddOrUpdateTaints: api.MapToAwsTaints(group.NodeTemplate.Taints),
		}
	}

	return input
}

// GetNodesInGroup get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroup(group *proto.NodeGroup, opt *cloudprovider.CommonOption) ([]*proto.NodeGroupNode,
	error) {
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
	return nil, cloudprovider.ErrCloudNotImplemented
}

// UpdateDesiredNodes update nodegroup desired node
func (ng *NodeGroup) UpdateDesiredNodes(desiredNode uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (*cloudprovider.ScalingResponse, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// SwitchNodeGroupAutoScaling switch nodegroup autoscaling
func (ng *NodeGroup) SwitchNodeGroupAutoScaling(group *proto.NodeGroup, enable bool,
	opt *cloudprovider.SwitchNodeGroupAutoScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
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
	return nil, cloudprovider.ErrCloudNotImplemented
}

// SwitchAutoScalingOptionStatus switch cluster autoscaling option status
func (ng *NodeGroup) SwitchAutoScalingOptionStatus(scalingOption *proto.ClusterAutoScalingOption, enable bool,
	opt *cloudprovider.CommonOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
