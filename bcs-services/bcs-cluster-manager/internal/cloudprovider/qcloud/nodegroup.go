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
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
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
		// init Node
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

	// update bkCloudName
	group.BkCloudName = cloudprovider.GetBKCloudName(int(group.BkCloudID))
	if group.NodeTemplate != nil && group.NodeTemplate.Module != nil &&
		len(group.NodeTemplate.Module.ScaleOutModuleID) != 0 {
		bkBizID, _ := strconv.Atoi(cluster.BusinessID)
		bkModuleID, _ := strconv.Atoi(group.NodeTemplate.Module.ScaleOutModuleID)
		group.NodeTemplate.Module.ScaleOutModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	}

	// update imageName
	if err := ng.updateImageInfo(group); err != nil {
		return err
	}

	return nil
}

func (ng *NodeGroup) generateModifyClusterNodePoolInput(group *proto.NodeGroup,
	clusterID string) *tke.ModifyClusterNodePoolRequest {
	// modify nodegroup
	req := tke.NewModifyClusterNodePoolRequest()
	req.ClusterId = &clusterID
	req.NodePoolId = &group.CloudNodeGroupID
	req.Name = &group.Name
	if group.NodeTemplate != nil {
		req.Taints = api.MapToCloudTaints(group.NodeTemplate.Taints)
		req.Labels = api.MapToCloudLabels(group.NodeTemplate.Labels)
	}
	req.Tags = api.MapToCloudTags(group.Tags)
	req.EnableAutoscale = common.BoolPtr(false)
	// MaxNodesNum/MinNodesNum 通过 asg 修改，这里不用修改
	if group.NodeTemplate != nil {
		req.Unschedulable = common.Int64Ptr(int64(group.NodeTemplate.UnSchedulable))
	}
	if group.LaunchTemplate != nil && group.LaunchTemplate.ImageInfo != nil && group.LaunchTemplate.ImageInfo.ImageID !=
		"" {
		req.OsName = &group.LaunchTemplate.ImageInfo.ImageID
	}
	return req
}

// generateModifyAutoScalingGroupInput 根据需要修改 asg
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

func (ng *NodeGroup) generateUpgradeLaunchConfigurationInput(
	group *proto.NodeGroup) *as.UpgradeLaunchConfigurationRequest {
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
	bw, _ := strconv.Atoi(group.LaunchTemplate.InternetAccess.InternetMaxBandwidth)
	internetChargeType := group.LaunchTemplate.InternetAccess.InternetChargeType
	if len(internetChargeType) == 0 {
		internetChargeType = api.InternetChargeTypeTrafficPostpaidByHour
	}
	req.InternetAccessible = &as.InternetAccessible{
		InternetChargeType:      common.StringPtr(internetChargeType),
		InternetMaxBandwidthOut: common.Uint64Ptr(uint64(bw)),
		PublicIpAssigned:        common.BoolPtr(group.LaunchTemplate.InternetAccess.PublicIPAssigned),
	}
	req.LoginSettings = &as.LoginSettings{Password: common.StringPtr(group.LaunchTemplate.InitLoginPassword)}

	projectID, err := strconv.Atoi(group.LaunchTemplate.ProjectID)
	if err == nil {
		req.ProjectId = common.Int64Ptr(int64(projectID))
	}
	req.SecurityGroupIds = common.StringPtrs(group.LaunchTemplate.SecurityGroupIDs)
	diskSize, _ := strconv.Atoi(group.LaunchTemplate.SystemDisk.GetDiskSize())
	req.SystemDisk = &as.SystemDisk{
		DiskType: common.StringPtr(group.LaunchTemplate.SystemDisk.GetDiskType()),
		DiskSize: common.Uint64Ptr(uint64(diskSize)),
	}
	req.UserData = common.StringPtr(group.LaunchTemplate.UserData)
	return req
}

func (ng *NodeGroup) updateImageInfo(group *proto.NodeGroup) error {
	if group.LaunchTemplate == nil || group.LaunchTemplate.ImageInfo == nil {
		return nil
	}
	imageName := group.LaunchTemplate.ImageInfo.ImageName
	for _, v := range api.ImageOsList {
		if v.ImageID == group.LaunchTemplate.ImageInfo.ImageID {
			imageName = v.Alias
			break
		}
	}
	if imageName == group.LaunchTemplate.ImageInfo.ImageName {
		return nil
	}
	group.LaunchTemplate.ImageInfo.ImageName = imageName
	return cloudprovider.GetStorageModel().UpdateNodeGroup(context.TODO(), group)
}

// GetNodesInGroup get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroup(group *proto.NodeGroup, opt *cloudprovider.CommonOption) ([]*proto.NodeGroupNode,
	error) {
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
	nodes, err := tkecli.GetNodeGroupInstances(cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		blog.Errorf("GetNodeGroupInstances failed, err: %s", err.Error())
		return nil, err
	}
	groupNodes := make([]*proto.NodeGroupNode, 0)
	for _, v := range nodes {
		if v.InstanceId == nil {
			continue
		}
		node := transTkeNodeToNode(v)
		node.NodeGroupID = group.NodeGroupID
		node.ClusterID = group.ClusterID
		groupNodes = append(groupNodes, node)
	}
	return groupNodes, nil
}

func transTkeNodeToNode(node *tke.Instance) *proto.NodeGroupNode {
	n := &proto.NodeGroupNode{NodeID: *node.InstanceId}
	if node.InstanceRole != nil {
		n.InstanceRole = *node.InstanceRole
	}
	if node.InstanceState != nil {
		switch *node.InstanceState {
		case "running":
			n.Status = "RUNNING"
		case "initializing":
			n.Status = "INITIALIZATION"
		case "failed":
			n.Status = "FAILED"
		default:
			n.Status = *node.InstanceState
		}
	}
	if node.LanIP != nil {
		n.InnerIP = *node.LanIP
	}
	return n
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
	blog.Infof("NodeGroup %s has total nodes %d, current capable nodes %d, current incoming nodes %d, details %v",
		group.NodeGroupID, len(goodNodes), current, inComingNodes, nodeNames)

	if current >= int(desired) {
		blog.Infof("NodeGroup %s current capable nodes %d larger than desired %d nodes, nothing to do",
			group.NodeGroupID, current, desired)
		return &cloudprovider.ScalingResponse{
				ScalingUp:    0,
				CapableNodes: nodeNames,
			}, fmt.Errorf("NodeGroup %s UpdateDesiredNodes nodes %d larger than desired %d nodes", group.NodeGroupID, current,
				desired)
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
	task, err := mgr.BuildSwitchAutoScalingOptionStatusTask(scalingOption, enable, opt)
	if err != nil {
		blog.Errorf("build SwitchAutoScalingOptionStatus task for cluster %s with cloudprovider %s failed, %s",
			scalingOption.ClusterID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}
