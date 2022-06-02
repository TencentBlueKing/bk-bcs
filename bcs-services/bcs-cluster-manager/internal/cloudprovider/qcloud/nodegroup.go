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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	intercommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
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
// will be released. Task is background automatic task
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

	// modify node pool
	// 节点池必须在 asg 前更新，因为 nodePool NodeOs 参数会覆盖 asg imageID
	if err := tkeCli.ModifyClusterNodePool(ng.generateModifyClusterNodePoolInput(group, cluster.SystemID)); err != nil {
		return err
	}

	// modify asg
	if err := asCli.ModifyAutoScalingGroup(ng.generateModifyAutoScalingGroupInput(group)); err != nil {
		return err
	}

	// modify launch config
	if err := asCli.UpgradeLaunchConfiguration(ng.generateUpgradeLaunchConfigurationInput(group)); err != nil {
		return err
	}
	return nil
}

func (ng *NodeGroup) generateModifyClusterNodePoolInput(group *proto.NodeGroup, clusterID string) *tke.ModifyClusterNodePoolRequest {
	// modify nodegroup
	req := tke.NewModifyClusterNodePoolRequest()
	req.ClusterId = &clusterID
	req.NodePoolId = &group.CloudNodeGroupID
	req.Name = &group.Name
	req.Labels = api.MapToCloudLabels(group.Labels)
	req.Taints = api.MapToCloudTaints(group.Taints)
	req.Tags = api.MapToCloudTags(group.Tags)
	req.EnableAutoscale = common.BoolPtr(false)
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
	if group.LaunchTemplate.DataDisks != nil {
		for i := range group.LaunchTemplate.DataDisks {
			diskSize, _ := strconv.Atoi(group.LaunchTemplate.DataDisks[i].DiskSize)
			req.DataDisks = append(req.DataDisks, &as.DataDisk{
				DiskType: common.StringPtr(group.LaunchTemplate.DataDisks[i].DiskType),
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
	diskSize, _ := strconv.Atoi(group.LaunchTemplate.SystemDisk.GetDiskSize())
	req.SystemDisk = &as.SystemDisk{
		DiskType: common.StringPtr(group.LaunchTemplate.SystemDisk.GetDiskType()),
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
	_, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("GetCloudAndCluster failed, err: %s", err.Error())
		return nil, err
	}
	tkecli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("create tke client failed, err: %s", err.Error())
		return nil, err
	}
	nodePool, err := tkecli.DescribeClusterNodePoolDetail(cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		blog.Errorf("DescribeClusterNodePoolDetail failed, err: %s", err.Error())
		return nil, err
	}
	if nodePool == nil || nodePool.AutoscalingGroupId == nil {
		err = fmt.Errorf("GetNodesInGroup failed, node pool is empty")
		blog.Errorf("%s", err.Error())
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
	nodes, err := nm.ListNodesByInstanceID(insIDs, &cloudprovider.ListNodesOption{
		Common:       opt,
		ClusterVPCID: group.AutoScaling.VpcID,
	})
	if err != nil {
		blog.Errorf("DescribeInstances failed, err: %s", err.Error())
		return nil, err
	}

	groupNodes := make([]*proto.Node, 0)
	for i := range nodes {
		nodes[i].Status = intercommon.StatusRunning
		nodes[i].ClusterID = group.ClusterID
		nodes[i].NodeGroupID = group.NodeGroupID
		groupNodes = append(groupNodes, nodes[i])
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
	_, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("GetCloudAndCluster failed, err: %s", err.Error())
		return err
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
	return tkeCli.RemoveNodeFromNodePool(cluster.SystemID, group.NodeGroupID, ids)
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

	// scaling nodes with desired, first get all node for status filtering
	// check if nodes are already in cluster
	goodNodes, err := cloudprovider.ListNodesInClusterNodePool(opt.Cluster.ClusterID, group.NodeGroupID)
	if err != nil {
		blog.Errorf("cloudprovider yunti get NodeGroup %s all Nodes failed, %s", group.NodeGroupID, err.Error())
		return nil, err
	}

	// check incoming nodes
	inComingNodes, err := cloudprovider.GetNodesNumWhenApplyInstanceTask(opt.Cluster.ClusterID, group.NodeGroupID,
		cloudprovider.GetTaskType(opt.Cloud.CloudProvider, cloudprovider.UpdateNodeGroupDesiredNode), cloudprovider.TaskStatusRunning,
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
	blog.Infof("NodeGroup %s has total nodes %d, current capable nodes %d, current incoming nodes %d, details %v",
		group.NodeGroupID, len(goodNodes), current, inComingNodes, nodeNames)

	if current >= int(desired) {
		blog.Infof("NodeGroup %s current capable nodes %d larger than desired %d nodes, nothing to do",
			group.NodeGroupID, current, desired)
		return &cloudprovider.ScalingResponse{
			ScalingUp:    0,
			CapableNodes: nodeNames,
		}, fmt.Errorf("NodeGroup %s UpdateDesiredNodes nodes %d larger than desired %d nodes", group.NodeGroupID, current, desired)
	}

	// current scale nodeNum
	scalingUp := int(desired) - current

	return &cloudprovider.ScalingResponse{
		ScalingUp:    uint32(scalingUp),
		CapableNodes: nodeNames,
	}, nil
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
