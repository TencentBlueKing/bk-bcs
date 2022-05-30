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

package qcloud

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	as "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
)

var groupMgr sync.Once

func init() {
	groupMgr.Do(func() {
		//init Node
		cloudprovider.InitNodeGroupManager(cloudName, &NodeGroup{})
	})
}

// NodeGroup nodegroup management in qcloud
type NodeGroup struct {
}

// CreateNodeGroup create nodegroup by cloudprovider api, only create NodeGroup entity
func (ng *NodeGroup) CreateNodeGroup(group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (
	*proto.Task, error) {
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
	tkeCli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("create tke client failed, err: %s", err.Error())
		return err
	}
	asCli, err := api.NewASClient(opt)
	if err != nil {
		blog.Errorf("create as client failed, err: %s", err.Error())
		return err
	}
	if group.NodeGroupID == "" || group.ClusterID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return fmt.Errorf("nodegroup id or cluster id is empty")
	}

	// modify asg
	if err := asCli.ModifyAutoScalingGroup(ng.generateModifyAutoScalingGroupInput(group)); err != nil {
		return err
	}

	// modify launch config
	if err := asCli.UpgradeLaunchConfiguration(ng.generateUpgradeLaunchConfigurationInput(group)); err != nil {
		return err
	}

	// modify node pool
	if err := tkeCli.ModifyClusterNodePool(ng.generateModifyClusterNodePoolInput(group)); err != nil {
		return err
	}
	return nil
}

func (ng *NodeGroup) generateModifyClusterNodePoolInput(group *proto.NodeGroup) *tke.ModifyClusterNodePoolRequest {
	// modify nodegroup
	req := tke.NewModifyClusterNodePoolRequest()
	req.ClusterId = &group.ClusterID
	req.NodePoolId = &group.NodeGroupID
	req.Name = &group.Name
	req.Labels = api.MapToCloudLabels(group.Labels)
	req.Taints = api.MapToCloudTaints(group.Taints)
	req.Tags = api.MapToCloudTags(group.Tags)
	req.EnableAutoscale = common.BoolPtr(false)
	req.OsName = &group.NodeOS
	// MaxNodesNum/MinNodesNum 通过 asg 修改，这里不用修改
	if group.NodeTemplate != nil {
		req.Unschedulable = common.Int64Ptr(int64(group.NodeTemplate.UnSchedulable))
	}
	return req
}

// 根据需要修改 asg
func (ng *NodeGroup) generateModifyAutoScalingGroupInput(group *proto.NodeGroup) *as.ModifyAutoScalingGroupRequest {
	req := as.NewModifyAutoScalingGroupRequest()
	if group.AutoScaling == nil {
		return nil
	}
	req.AutoScalingGroupId = &group.AutoScaling.AutoScalingID
	req.AutoScalingGroupName = &group.AutoScaling.AutoScalingName
	req.MaxSize = common.Uint64Ptr(uint64(group.AutoScaling.MaxSize))
	req.MinSize = common.Uint64Ptr(uint64(group.AutoScaling.MinSize))
	req.DefaultCooldown = common.Uint64Ptr(uint64(group.AutoScaling.DefaultCooldown))
	req.SubnetIds = common.StringPtrs(group.AutoScaling.SubnetIDs)
	req.RetryPolicy = common.StringPtr(group.AutoScaling.RetryPolicy)
	req.MultiZoneSubnetPolicy = common.StringPtr(group.AutoScaling.MultiZoneSubnetPolicy)
	req.ServiceSettings = &as.ServiceSettings{ScalingMode: &group.AutoScaling.ScalingMode,
		ReplaceMonitorUnhealthy: common.BoolPtr(group.AutoScaling.ReplaceUnhealthy)}
	return req
}

func (ng *NodeGroup) generateUpgradeLaunchConfigurationInput(group *proto.NodeGroup) *as.UpgradeLaunchConfigurationRequest {
	req := as.NewUpgradeLaunchConfigurationRequest()
	if group.LaunchTemplate == nil || group.LaunchTemplate.InternetAccess == nil {
		blog.Warnf("group launch template is nil, %v", utils.ToJSONString(group))
		return nil
	}
	req.LaunchConfigurationId = common.StringPtr(group.LaunchTemplate.LaunchConfigurationID)
	req.LaunchConfigurationName = common.StringPtr(group.LaunchTemplate.LaunchConfigureName)
	if group.LaunchTemplate.ImageInfo != nil && group.LaunchTemplate.ImageInfo.ImageID != "" {
		req.ImageId = common.StringPtr(group.LaunchTemplate.ImageInfo.ImageID)
	}
	req.InstanceTypes = common.StringPtrs([]string{group.LaunchTemplate.InstanceType})
	if group.NodeTemplate.DataDisks != nil {
		for i := range group.NodeTemplate.DataDisks {
			diskSize, _ := strconv.Atoi(group.NodeTemplate.DataDisks[i].DiskSize)
			req.DataDisks = append(req.DataDisks, &as.DataDisk{
				DiskType: common.StringPtr(group.NodeTemplate.DataDisks[i].DiskType),
				DiskSize: common.Uint64Ptr(uint64(diskSize)),
			})
		}
	}
	req.EnhancedService = &as.EnhancedService{
		SecurityService: &as.RunSecurityServiceEnabled{Enabled: common.BoolPtr(group.LaunchTemplate.IsSecurityService)},
		MonitorService:  &as.RunMonitorServiceEnabled{Enabled: common.BoolPtr(group.LaunchTemplate.IsMonitorService)},
	}
	req.InstanceChargeType = common.StringPtr(group.LaunchTemplate.InstanceChargeType)
	req.InternetAccessible = &as.InternetAccessible{
		InternetChargeType:      common.StringPtr(group.LaunchTemplate.InternetAccess.InternetChargeType),
		InternetMaxBandwidthOut: common.Uint64Ptr(uint64(group.LaunchTemplate.InternetAccess.InternetMaxBandwidth)),
		PublicIpAssigned:        common.BoolPtr(group.LaunchTemplate.InternetAccess.PublicIPAssigned),
	}
	req.LoginSettings = &as.LoginSettings{Password: common.StringPtr(group.LaunchTemplate.InitLoginPassword)}
	req.ProjectId = common.Int64Ptr(int64(group.LaunchTemplate.ProjectID))
	req.SecurityGroupIds = common.StringPtrs(group.LaunchTemplate.SecurityGroupIDs)
	diskSize, _ := strconv.Atoi(group.NodeTemplate.SystemDisk.DiskSize)
	req.SystemDisk = &as.SystemDisk{
		DiskType: common.StringPtr(group.NodeTemplate.SystemDisk.DiskType),
		DiskSize: common.Uint64Ptr(uint64(diskSize)),
	}
	req.UserData = common.StringPtr(group.LaunchTemplate.UserData)
	return req
}

// GetNodesInGroup get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroup(group *proto.NodeGroup, opt *cloudprovider.CommonOption) ([]*proto.Node, error) {
	if group.ClusterID == "" || group.NodeGroupID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return nil, fmt.Errorf("nodegroup id or cluster id is empty")
	}
	tkecli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("create tke client failed, err: %s", err.Error())
		return nil, err
	}
	nodePool, err := tkecli.DescribeClusterNodePoolDetail(group.ClusterID, group.NodeGroupID)
	if err != nil {
		blog.Errorf("DescribeClusterNodePoolDetail failed, err: %s", err.Error())
		return nil, err
	}
	asCli, err := api.NewASClient(opt)
	if err != nil {
		blog.Errorf("create as client failed, err: %s", err.Error())
		return nil, err
	}
	ins, err := asCli.DescribeAutoScalingInstances(*nodePool.AutoscalingGroupId)
	if err != nil {
		blog.Errorf("DescribeAutoScalingInstances failed, err: %s", err.Error())
		return nil, err
	}
	insIDs := make([]string, 0)
	for _, v := range ins {
		insIDs = append(insIDs, *v.InstanceID)
	}
	if len(insIDs) == 0 {
		return nil, nil
	}
	nm := api.NodeManager{}
	nodes, err := nm.DescribeInstances(insIDs, nil, opt)
	if err != nil {
		blog.Errorf("DescribeInstances failed, err: %s", err.Error())
		return nil, err
	}
	return nodes, nil
}

// MoveNodesToGroup add cluster nodes to NodeGroup
func (ng *NodeGroup) MoveNodesToGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when MoveNodesToGroup %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildMoveNodesToGroupTask(nodes, group, opt)
	if err != nil {
		blog.Errorf("build MoveNodesToGroup task for cluster %s with cloudprovider %s failed, %s",
			group.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// RemoveNodesFromGroup remove nodes from NodeGroup, nodes are still in cluster
func (ng *NodeGroup) RemoveNodesFromGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.RemoveNodesOption) error {
	if group.ClusterID == "" || group.NodeGroupID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return fmt.Errorf("nodegroup id or cluster id is empty")
	}
	tkeCli, err := api.NewTkeClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("create tke client failed, err: %s", err.Error())
		return err
	}
	ids := make([]string, 0)
	for _, v := range nodes {
		ids = append(ids, v.NodeID)
	}
	return tkeCli.RemoveNodeFromNodePool(group.ClusterID, group.NodeGroupID, ids)
}

// CleanNodesInGroup clean specified nodes in NodeGroup,
func (ng *NodeGroup) CleanNodesInGroup(nodes []*proto.Node, group *proto.NodeGroup,
	opt *cloudprovider.CleanNodesOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CleanNodesInGroup %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildCleanNodesInGroupTask(nodes, group, opt)
	if err != nil {
		blog.Errorf("build CleanNodesInGroup task for cluster %s with cloudprovider %s failed, %s",
			group.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// UpdateDesiredNodes update nodegroup desired node
func (ng *NodeGroup) UpdateDesiredNodes(desired uint32, group *proto.NodeGroup,
	opt *cloudprovider.UpdateDesiredNodeOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when UpdateDesiredNodes %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildUpdateDesiredNodesTask(desired, group, opt)
	if err != nil {
		blog.Errorf("build UpdateDesiredNodes task for cluster %s with cloudprovider %s failed, %s",
			group.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// SwitchNodeGroupAutoScaling switch nodegroup auto scaling
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
	opt *cloudprovider.DeleteScalingOption) (*proto.Task, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
