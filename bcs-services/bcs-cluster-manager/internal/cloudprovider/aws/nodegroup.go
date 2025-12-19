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

package aws

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/eks"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

func init() {
	cloudprovider.InitNodeGroupManager("aws", &NodeGroup{})
}

// NodeGroup nodegroup management for aws resource pool solution
type NodeGroup struct {
}

// CreateNodeGroup create nodegroup by cloudprovider api, only create NodeGroup entity
func (ng *NodeGroup) CreateNodeGroup(group *proto.NodeGroup,
	opt *cloudprovider.CreateNodeGroupOption) (*proto.Task, error) {
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
func (ng *NodeGroup) UpdateNodeGroup(
	group *proto.NodeGroup, opt *cloudprovider.UpdateNodeGroupOption) (*proto.Task, error) {
	_, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("get cluster %s failed, %s", group.ClusterID, err.Error())
		return nil, err
	}

	eksCli, err := api.NewEksClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create eks client failed, err: %s", err.Error())
		return nil, err
	}

	if group.NodeGroupID == "" || group.ClusterID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return nil, fmt.Errorf("nodegroup id or cluster id is empty")
	}

	cloudNg, err := eksCli.DescribeNodegroup(&group.CloudNodeGroupID, &cluster.SystemID)
	if err != nil {
		blog.Errorf("get cloud nodegroup failed, err: %s", err.Error())
		return nil, err
	}

	_, err = eksCli.UpdateNodegroupConfig(ng.generateUpdateNodegroupConfigInput(group, cloudNg, cluster.SystemID))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (ng *NodeGroup) generateUpdateNodegroupConfigInput(group *proto.NodeGroup, cloudNg *eks.Nodegroup, // nolint
	cluster string) *eks.UpdateNodegroupConfigInput {
	input := &eks.UpdateNodegroupConfigInput{
		ClusterName:   &cluster,
		NodegroupName: &group.CloudNodeGroupID,
	}

	// labels 和taints 使用bcs的任务添加，不用节点池自带功能
	// if len(group.GetNodeTemplate().GetLabels()) > 0 {
	// 	input.Labels = &eks.UpdateLabelsPayload{
	// 		AddOrUpdateLabels: aws.StringMap(group.NodeTemplate.Labels),
	// 	}

	// 	for k := range cloudNg.Labels {
	// 		if _, ok := group.NodeTemplate.Labels[k]; !ok {
	// 			input.Labels.RemoveLabels = append(input.Labels.RemoveLabels, aws.String(k))
	// 		}
	// 	}
	// } else if len(cloudNg.Labels) > 0 {
	// 	input.Labels = &eks.UpdateLabelsPayload{}
	// 	for k := range cloudNg.Labels {
	// 		input.Labels.RemoveLabels = append(input.Labels.RemoveLabels, aws.String(k))
	// 	}
	// }

	// if len(group.GetNodeTemplate().GetTaints()) > 0 {
	// 	input.Taints = &eks.UpdateTaintsPayload{
	// 		AddOrUpdateTaints: api.MapToAwsTaints(group.NodeTemplate.Taints),
	// 	}

	// 	for _, v := range cloudNg.Taints {
	// 		exit := false
	// 		for _, y := range group.NodeTemplate.Taints {
	// 			if *v.Key == y.Key {
	// 				exit = true
	// 				continue
	// 			}
	// 		}

	// 		if !exit {
	// 			input.Taints.RemoveTaints = append(input.Taints.RemoveTaints, &eks.Taint{
	// 				Key:    v.Key,
	// 				Value:  v.Value,
	// 				Effect: v.Effect,
	// 			})
	// 		}
	// 	}
	// } else if len(cloudNg.Taints) > 0 {
	// 	input.Taints = &eks.UpdateTaintsPayload{}
	// 	input.Taints.RemoveTaints = append(input.Taints.RemoveTaints, cloudNg.Taints...)
	// }

	if group.AutoScaling != nil {
		input.ScalingConfig = &eks.NodegroupScalingConfig{
			MaxSize: aws.Int64(int64(group.AutoScaling.MaxSize)),
			MinSize: aws.Int64(int64(group.AutoScaling.MinSize)),
		}
	}

	return input
}

// RecommendNodeGroupConf recommends nodegroup configs
func (ng *NodeGroup) RecommendNodeGroupConf(
	ctx context.Context, opt *cloudprovider.CommonOption) ([]*proto.RecommendNodeGroupConf, error) {
	if opt == nil {
		return nil, fmt.Errorf("invalid request")
	}

	configs := make([]*proto.RecommendNodeGroupConf, 0)
	config := generateNodeGroupConf()

	mgr := api.NodeManager{}
	serviceRoles, err := mgr.GetServiceRoles(opt, "nodeGroup")
	if err != nil {
		blog.Errorf("RecommendNodeGroupConf GetServiceRoles failed, %s", err.Error())
		return nil, err
	}
	if len(serviceRoles) == 0 {
		return nil, fmt.Errorf("RecommendNodeGroupConf GetServiceRoles failed, no valid EKS-Node-Role")
	}
	config.ServiceRoleName = serviceRoles[0].RoleName

	insTypes, err := mgr.ListNodeInstanceType(ctx, cloudprovider.InstanceInfo{
		Region: opt.Region,
		CPU:    8,
		Memory: 16,
	}, opt)
	if err != nil {
		return nil, fmt.Errorf("list node instance type failed, %s", err.Error())
	}
	if len(insTypes) == 0 {
		return nil, fmt.Errorf("RecommendNodeGroupConf no valid instanceType for 8c16g")
	}
	config.InstanceProfile.InstanceType = insTypes[0].NodeType
	configs = append(configs, config)

	return configs, nil
}

func generateNodeGroupConf() *proto.RecommendNodeGroupConf {
	return &proto.RecommendNodeGroupConf{
		Name: "default",
		InstanceProfile: &proto.InstanceProfile{
			NodeOS:             "AL_X86_64",
			InstanceChargeType: "POSTPAID_BY_HOUR",
		},
		HardwareProfile: &proto.HardwareProfile{
			CPU: 8,
			Mem: 16,
			SystemDisk: &proto.DataDisk{
				DiskType: "gp2",
				DiskSize: "100",
			},
			DataDisks: []*proto.DataDisk{
				{
					DiskType: "gp2",
					DiskSize: "100",
				},
			},
		},
		NetworkProfile: &proto.NetworkProfile{
			PublicIPAssigned: false,
		},
		ScalingProfile: &proto.ScalingProfile{
			DesiredSize: 2,
			MaxSize:     10,
			// 释放模式
			ScalingMode: "Delete",
		},
	}
}

// GetNodesInGroup get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroup(group *proto.NodeGroup, opt *cloudprovider.CommonOption) ([]*proto.Node, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetNodesInGroupV2 get nodeGroup nodes by v2 version
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
	if group.ClusterID == "" || group.NodeGroupID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return fmt.Errorf("nodegroup id or cluster id is empty")
	}
	client, err := api.NewAutoScalingClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create gce client failed, err: %s", err.Error())
		return err
	}
	ids := make([]string, 0)
	for _, v := range nodes {
		ids = append(ids, v.NodeID)
	}
	_, err = client.DetachInstances(&autoscaling.DetachInstancesInput{
		AutoScalingGroupName: aws.String(group.AutoScaling.AutoScalingName),
		InstanceIds:          aws.StringSlice(ids),
	})
	return err
}

// CleanNodesInGroup clean specified nodes in NodeGroup,
func (ng *NodeGroup) CleanNodesInGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.CleanNodesOption) (*proto.Task, error) {
	if len(nodes) == 0 || opt == nil || opt.Cluster == nil || opt.Cloud == nil {
		return nil, fmt.Errorf("invalid request")
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

	taskType := cloudprovider.GetTaskType(opt.Cloud.CloudProvider, cloudprovider.UpdateNodeGroupDesiredNode)

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"clusterid":   opt.Cluster.ClusterID,
		"tasktype":    taskType,
		"nodegroupid": group.NodeGroupID,
		"status":      cloudprovider.TaskStatusRunning,
	})
	taskList, err := cloudprovider.GetStorageModel().ListTask(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("UpdateDesiredNodes failed: %v", err)
		return nil, err
	}
	if len(taskList) != 0 {
		return nil, fmt.Errorf("eks task(%d) %s is still running", len(taskList), taskType)
	}

	needScaleOutNodes := desired - group.GetAutoScaling().GetDesiredSize()

	blog.Infof("cluster[%s] nodeGroup[%s] current nodes[%d] desired nodes[%d] needNodes[%s]",
		group.ClusterID, group.NodeGroupID, group.GetAutoScaling().GetDesiredSize(), desired, needScaleOutNodes)

	if desired <= group.GetAutoScaling().GetDesiredSize() {
		return nil, fmt.Errorf("NodeGroup %s current nodes %d larger than or equel to desired %d nodes",
			group.Name, group.GetAutoScaling().GetDesiredSize(), desired)
	}

	return &cloudprovider.ScalingResponse{
		ScalingUp: needScaleOutNodes,
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
		blog.Errorf("get cloud %s TaskManager when UpdateAutoScalingOption %s failed, %s",
			cloudName, scalingOption.ClusterID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildUpdateAutoScalingOptionTask(scalingOption, opt)
	if err != nil {
		blog.Errorf("build UpdateAutoScalingOption task for cluster %s with cloudprovider %s failed, %s",
			scalingOption.ClusterID, cloudName, err.Error(),
		)
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
