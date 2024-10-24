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

package qcloud

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	as "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	cutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	intercommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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

	// handler external nodes
	if group.NodeGroupType == intercommon.External.String() {
		nodePoolID, err := createExternalNodePool(group, opt)
		if err != nil {
			blog.Errorf("CreateNodeGroup createExternalNodePool failed: %v", err)
			return nil, err
		}
		group.CloudNodeGroupID = nodePoolID

		return nil, nil
	}

	// normal nodePool
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when CreateNodeGroup %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	// build createNodeGroup task
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
	// external nodePool
	if group.NodeGroupType == intercommon.External.String() {
		err := deleteExternalNodePool(group, opt)
		if err != nil {
			blog.Errorf("DeleteNodeGroup deleteExternalNodePool failed: %v", err)
			return nil, err
		}

		blog.Infof("DeleteNodeGroup[%s:%s] deleteExternalNodePool successful",
			group.NodeGroupID, group.CloudNodeGroupID)
		return nil, nil
	}

	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteNodeGroup %s failed, %s",
			cloudName, group.Name, err.Error(),
		)
		return nil, err
	}
	// build Delete nodeGroup task
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
func (ng *NodeGroup) UpdateNodeGroup(group *proto.NodeGroup, opt *cloudprovider.UpdateNodeGroupOption) (
	*proto.Task, error) {
	if group == nil || opt == nil {
		return nil, fmt.Errorf("UpdateNodeGroup group or opt is nil")
	}
	if group.NodeGroupID == "" || group.ClusterID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return nil, fmt.Errorf("nodegroup id or cluster id is empty")
	}

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: group.ClusterID,
		CloudID:   group.Provider,
	})
	if err != nil {
		blog.Errorf("get cluster %s failed, %s", group.ClusterID, err.Error())
		return nil, err
	}

	// external nodePool
	if group.NodeGroupType == intercommon.External.String() {
		err = ng.updateExternalNodePool(dependInfo.Cluster.SystemID, group, &opt.CommonOption)
		if err != nil {
			blog.Errorf("UpdateNodeGroup[%s] updateExternalNodePool failed: %v", group.NodeGroupID, err)
			return nil, err
		}
		blog.Infof("UpdateNodeGroup[%s] updateExternalNodePool successful", group.NodeGroupID)

		return nil, nil
	}

	// update normal nodePool
	err = ng.updateNormalNodePool(dependInfo.Cluster.SystemID, group, &opt.CommonOption)
	if err != nil {
		blog.Errorf("UpdateNodeGroup[%s] updateNormalNodePool failed: %v", group.NodeGroupID, err)
		return nil, err
	}
	blog.Infof("UpdateNodeGroup[%s] updateNormalNodePool successful", group.NodeGroupID)

	// build task
	mgr, err := cloudprovider.GetTaskManager(opt.Cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when BuildUpdateNodeGroupTask in NodeGroup %s failed, %s",
			opt.Cloud.CloudProvider, group.NodeGroupID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildUpdateNodeGroupTask(group, &opt.CommonOption)
	if err != nil {
		blog.Errorf("BuildUpdateNodeGroupTask failed: %v", err)
		return nil, err
	}

	return task, nil
}

// updateExternalNodePool update external nodePool
func (ng *NodeGroup) updateExternalNodePool(
	systemID string, group *proto.NodeGroup, opt *cloudprovider.CommonOption) error {
	tkeCli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("updateExternalNodePool NewTkeClient failed, err: %s", err.Error())
		return err
	}

	// tke external only support modify nodePoolName / labels / taints
	err = retry.Do(func() error {
		errModify := tkeCli.ModifyExternalNodePool(systemID, api.ModifyExternalNodePoolConfig{
			NodePoolId: group.GetCloudNodeGroupID(),
			Name:       group.Name,
			Labels: func() []*api.Label {
				if group.NodeTemplate == nil {
					return nil
				}
				return api.MapToLabels(group.NodeTemplate.Labels)
			}(),
			Taints: func() []*api.Taint {
				if group.NodeTemplate == nil {
					return nil
				}
				return api.MapToTaints(group.NodeTemplate.Taints)
			}(),
		})
		if errModify != nil {
			blog.Errorf("updateExternalNodePool[%s] ModifyExternalNodePool failed: %v",
				group.CloudNodeGroupID, errModify)
			return errModify
		}

		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("updateExternalNodePool[%s] ModifyExternalNodePool failed: %v", group.CloudNodeGroupID, err)
		return err
	}

	blog.Errorf("updateExternalNodePool[%s] ModifyExternalNodePool successful", group.CloudNodeGroupID)
	return nil
}

// updateNormalNodePool update normal nodePool
func (ng *NodeGroup) updateNormalNodePool(
	systemID string, group *proto.NodeGroup, opt *cloudprovider.CommonOption) error {
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: group.ClusterID,
		CloudID:   group.Provider,
	})
	if err != nil {
		blog.Errorf("get cluster %s failed, %s", group.ClusterID, err.Error())
		return err
	}
	tkeCli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("updateNormalNodePool NewTkeClient failed, err: %s", err.Error())
		return err
	}

	// modify node pool
	if err = tkeCli.ModifyClusterNodePool(ng.generateModifyClusterNodePoolInput(group, systemID)); err != nil {
		return err
	}
	/*
		// as client
		asCli, err := api.NewASClient(opt)
		if err != nil {
			blog.Errorf("updateNormalNodePool NewASClient failed, err: %s", err.Error())
			return err
		}
		// modify asg
		if err = asCli.ModifyAutoScalingGroup(ng.generateModifyAutoScalingGroupInput(group)); err != nil {
			return err
		}

		// modify launch config
		if err = asCli.UpgradeLaunchConfiguration(ng.generateUpgradeLaunchConfInput(group)); err != nil {
			return err
		}
	*/
	// update bkCloudName
	if group.Area == nil {
		group.Area = &proto.CloudArea{}
	}
	group.Area.BkCloudName = cloudprovider.GetBKCloudName(int(group.Area.BkCloudID))

	// module info
	if group.NodeTemplate != nil && group.NodeTemplate.Module != nil &&
		len(group.NodeTemplate.Module.ScaleOutModuleID) != 0 {
		bkBizID, _ := strconv.Atoi(dependInfo.Cluster.BusinessID)
		bkModuleID, _ := strconv.Atoi(group.NodeTemplate.Module.ScaleOutModuleID)
		group.NodeTemplate.Module.ScaleOutModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	}

	// update imageName
	if err = ng.updateImageInfo(group); err != nil {
		return err
	}

	return nil
}

// update image info
func (ng *NodeGroup) updateImageInfo(group *proto.NodeGroup) error {
	if group.LaunchTemplate == nil || group.LaunchTemplate.ImageInfo == nil {
		return nil
	}

	// image info
	imageName := group.LaunchTemplate.ImageInfo.ImageName
	for _, v := range utils.ImageOsList {
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

// generateModifyClusterNodePoolInput modify nodePool info
func (ng *NodeGroup) generateModifyClusterNodePoolInput(
	group *proto.NodeGroup, clusterID string) *tke.ModifyClusterNodePoolRequest {
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

	if group.AutoScaling != nil {
		req.MinNodesNum = common.Int64Ptr(int64(group.AutoScaling.MinSize))
		req.MaxNodesNum = common.Int64Ptr(int64(group.AutoScaling.MaxSize))
	}

	if group.NodeTemplate.PreStartUserScript != "" {
		req.UserScript = common.StringPtr(group.NodeTemplate.PreStartUserScript)
	}

	// MaxNodesNum/MinNodesNum 通过 asg 修改，这里不用修改
	if group.NodeTemplate != nil {
		req.Unschedulable = common.Int64Ptr(int64(group.NodeTemplate.UnSchedulable))
	}

	// 节点池Os 当为自定义镜像时，传镜像id；否则为公共镜像的osName; 若为空复用集群级别
	// 示例值：ubuntu18.04.1x86_64
	if group.NodeTemplate != nil && group.NodeTemplate.NodeOS != "" {
		req.OsName = &group.NodeTemplate.NodeOS
	}

	kubeletParas := cutils.GetKubeletParas(group.NodeTemplate)
	if paras, ok := kubeletParas[intercommon.Kubelet]; ok {
		if req.ExtraArgs == nil {
			req.ExtraArgs = &tke.InstanceExtraArgs{}
		}
		req.ExtraArgs.Kubelet = common.StringPtrs(utils.FilterEmptyString(strings.Split(paras, ";")))
	}

	return req
}

// 根据需要修改 asg
func (ng *NodeGroup) generateModifyAutoScalingGroupInput(group *proto.NodeGroup) *as.ModifyAutoScalingGroupRequest { // nolint
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

// generateUpgradeLaunchConfInput upgrade launch config
func (ng *NodeGroup) generateUpgradeLaunchConfInput( // nolint
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
	// req.UserData = common.StringPtr(group.LaunchTemplate.UserData)
	return req
}

// GetNodesInGroupV2 get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroupV2(
	group *proto.NodeGroup, opt *cloudprovider.CommonOption) ([]*proto.NodeGroupNode, error) {
	if group.ClusterID == "" || group.NodeGroupID == "" {
		blog.Errorf("nodegroup id or cluster id is empty")
		return nil, fmt.Errorf("nodegroup id or cluster id is empty")
	}
	_, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), group.Provider, group.ClusterID)
	if err != nil {
		blog.Errorf("GetCloudAndCluster failed, err: %s", err.Error())
		return nil, err
	}
	tkeCli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("create tke client failed, err: %s", err.Error())
		return nil, err
	}

	// get group nodes
	nodes, err := tkeCli.GetNodeGroupInstances(cluster.SystemID, group.CloudNodeGroupID)
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

// transTkeNodeToNode trans node status
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

// RecommendNodeGroupConf recommends nodegroup configs
func (ng *NodeGroup) RecommendNodeGroupConf(opt *cloudprovider.CommonOption) ([]*proto.RecommendNodeGroupConf, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
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
	tkeCli, err := api.NewTkeClient(opt)
	if err != nil {
		blog.Errorf("create tke client failed, err: %s", err.Error())
		return nil, err
	}

	// cloud node pool info
	nodePool, err := tkeCli.DescribeClusterNodePoolDetail(cluster.SystemID, group.CloudNodeGroupID)
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

	// nodePool instances
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
	nm := NodeManager{}
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

	return groupNodes, nil
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

	// build move nodes to group task
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
		blog.Errorf("cloudprovider qcloud get NodeGroup %s all Nodes failed, %s", group.NodeGroupID, err.Error())
		return nil, err
	}

	// check incoming nodes
	inComingNodes, err := cloudprovider.GetNodesNumWhenApplyInstanceTask(opt.Cluster.ClusterID, group.NodeGroupID,
		cloudprovider.GetTaskType(opt.Cloud.CloudProvider, cloudprovider.UpdateNodeGroupDesiredNode),
		cloudprovider.TaskStatusRunning,
		[]string{cloudprovider.GetTaskType(opt.Cloud.CloudProvider, cloudprovider.ApplyInstanceMachinesTask),
			cloudprovider.GetTaskType(opt.Cloud.CloudProvider, cloudprovider.ApplyExternalNodeMachinesTask)})
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
	return nil, nil
}

// DeleteAutoScalingOption delete cluster autoscaling, cloudprovider will clean
// cluster-autoscaler in backgroup according cloudprovider implementation
func (ng *NodeGroup) DeleteAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.DeleteScalingOption) (*proto.Task, error) {
	return nil, nil
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

	err = cloudprovider.UpdateAutoScalingOptionModuleInfo(scalingOption.ClusterID)
	if err != nil {
		blog.Errorf("UpdateAutoScalingOption update asOption moduleInfo failed: %v", err)
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
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when AddExternalNodeToCluster %s failed, %s",
			cloudName, group.NodeGroupID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildAddExternalNodeToCluster(group, nodes, opt)
	if err != nil {
		blog.Errorf("build AddExternalNodeToCluster task for nodeGroup %s with cloudprovider %s failed, %s",
			group.NodeGroupID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

// DeleteExternalNodeFromCluster remove external node from cluster
func (ng *NodeGroup) DeleteExternalNodeFromCluster(group *proto.NodeGroup, nodes []*proto.Node,
	opt *cloudprovider.DeleteExternalNodesOption) (*proto.Task, error) {
	mgr, err := cloudprovider.GetTaskManager(cloudName)
	if err != nil {
		blog.Errorf("get cloud %s TaskManager when DeleteExternalNodeFromCluster %s failed, %s",
			cloudName, group.NodeGroupID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildDeleteExternalNodeFromCluster(group, nodes, opt)
	if err != nil {
		blog.Errorf("build DeleteExternalNodeFromCluster task for nodeGroup %s with cloudprovider %s failed, %s",
			group.NodeGroupID, cloudName, err.Error(),
		)
		return nil, err
	}
	return task, nil
}

func createExternalNodePool(group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (string, error) {
	tkeCli, err := api.NewTkeClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("createExternalNodePool[%s] failed: %v", cloudName, err)
		return "", err
	}

	var (
		nodePoolID string
	)
	err = retry.Do(func() error {
		nodePoolID, err = tkeCli.CreateExternalNodePool(opt.Cluster.GetSystemID(), api.CreateExternalNodePoolConfig{
			Name: group.Name,
			ContainerRuntime: func() string {
				if group != nil && group.NodeTemplate != nil && group.NodeTemplate.GetRuntime() != nil {
					return group.NodeTemplate.GetRuntime().ContainerRuntime
				}

				return intercommon.DockerContainerRuntime
			}(),
			RuntimeVersion: func() string {
				if group != nil && group.NodeTemplate != nil && group.NodeTemplate.GetRuntime() != nil {
					return group.NodeTemplate.GetRuntime().RuntimeVersion
				}
				return intercommon.DockerRuntimeVersion
			}(),
			Labels: func() []*api.Label {
				if group.NodeTemplate == nil {
					return nil
				}
				return api.MapToLabels(group.NodeTemplate.Labels)
			}(),
			Taints: func() []*api.Taint {
				if group.NodeTemplate == nil {
					return nil
				}
				return api.MapToTaints(group.NodeTemplate.Taints)
			}(),
			InstanceAdvancedSettings: business.GenerateClsAdvancedInsSettingFromNT(&cloudprovider.CloudDependBasicInfo{
				Cluster:      opt.Cluster,
				NodeTemplate: group.NodeTemplate,
			}, template.RenderVars{Render: false}, nil),
		})
		if err != nil {
			blog.Errorf("createExternalNodePool[%s] failed: %v", cloudName, err)
			return err
		}
		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("createExternalNodePool[%s] failed: %v", cloudName, err)
		return nodePoolID, err
	}

	blog.Infof("cloud[%s] createExternalNodePool successful[%s]", cloudName, nodePoolID)

	return nodePoolID, nil
}

func deleteExternalNodePool(group *proto.NodeGroup, opt *cloudprovider.DeleteNodeGroupOption) error {
	tkeCli, err := api.NewTkeClient(&opt.CommonOption)
	if err != nil {
		blog.Errorf("deleteExternalNodePool[%s] failed: %v", cloudName, err)
		return err
	}

	err = retry.Do(func() error {
		err = tkeCli.DeleteExternalNodePool(opt.Cluster.GetSystemID(), api.DeleteExternalNodePoolConfig{
			NodePoolIds: []string{group.CloudNodeGroupID},
			Force:       false,
		})

		if err != nil {
			blog.Errorf("deleteExternalNodePool[%s] failed: %v", cloudName, err)
			return err
		}
		return nil
	}, retry.Attempts(3))
	if err != nil {
		blog.Errorf("deleteExternalNodePool[%s] failed: %v", cloudName, err)
		return err
	}

	blog.Infof("cloud[%s] deleteExternalNodePool successful[%s:%s]", cloudName, group.ClusterID, group.CloudNodeGroupID)

	return nil
}

// GetExternalNodeScript get nodegroup external node script
func (ng *NodeGroup) GetExternalNodeScript(group *proto.NodeGroup, internal bool) (string, error) {
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   group.ClusterID,
		CloudID:     group.Provider,
		NodeGroupID: group.NodeGroupID,
	})
	if err != nil {
		errMsg := fmt.Errorf("GetExternalNodeScript[%s] GetClusterDependBasicInfo failed, %s",
			group.NodeGroupID, err.Error())
		return "", errMsg
	}

	script, err := business.GetClusterExternalNodeScript(context.Background(), dependInfo, internal)
	if err != nil {
		blog.Errorf("GetExternalNodeScript[%s] GetClusterExternalNodeScript failed: %v", group.NodeGroupID, err)
		return "", err
	}

	blog.Infof("GetExternalNodeScript[%s] successful", group.NodeGroupID)
	return script, nil
}

// CheckResourcePoolQuota check resource pool quota when revise group limit
func (ng *NodeGroup) CheckResourcePoolQuota(group *proto.NodeGroup, scaleUpNum uint32) error {
	return nil
}
