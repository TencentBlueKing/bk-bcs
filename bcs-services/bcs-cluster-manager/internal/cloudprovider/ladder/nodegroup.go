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

// Package ladder xxx
package ladder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/daemon"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var groupMgr sync.Once

func init() {
	groupMgr.Do(func() {
		cloudprovider.InitNodeGroupManager(cloudName, &NodeGroup{})
	})
}

//! this part can not open source to github
//! write an random ip address for banning commit: 10.1.1.1

// NodeGroup nodegroup management for yunti resource pool solution.
// yunti has no api implementation for nodegroup management.
// it offers only three features: resource application, addExistedNodeToCluster and
// leaveClusterAndReturnCVM. We need implement all NodeGroup features
// then offering apis to cluster-autoscaler.
type NodeGroup struct {
	// internal authentication information
}

// CreateNodeGroup create nodegroup by cloudprovider api, build createNodeGroup task
func (ng *NodeGroup) CreateNodeGroup(
	group *proto.NodeGroup, opt *cloudprovider.CreateNodeGroupOption) (*proto.Task, error) {
	if opt.OnlyData {
		return nil, nil
	}

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
func (ng *NodeGroup) DeleteNodeGroup(
	group *proto.NodeGroup, nodes []*proto.Node, opt *cloudprovider.DeleteNodeGroupOption) (*proto.Task, error) {
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
		blog.Errorf("get cloud %s TaskManager when DeleteNodeGroup in NodeGroup %s failed, %s",
			opt.Cloud.CloudProvider, group.NodeGroupID, err.Error(),
		)
		return nil, err
	}
	task, err := mgr.BuildDeleteNodeGroupTask(group, nodes, opt)
	if err != nil {
		blog.Errorf("BuildDeleteNodeGroupTask failed: %v", err)
		return nil, err
	}

	return task, nil
}

// UpdateNodeGroup update specified nodegroup configuration
func (ng *NodeGroup) UpdateNodeGroup(
	group *proto.NodeGroup, opt *cloudprovider.UpdateNodeGroupOption) (*proto.Task, error) {
	// nothing need to be updated, yunti not entity nodeGroup
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
		blog.Errorf("nodegroup id or cluster id is empty")
		return nil, fmt.Errorf("nodegroup id or cluster id is empty")
	}

	err := cloudprovider.UpdateNodeGroupCloudAndModuleInfo(group.NodeGroupID, group.ConsumerID,
		true, opt.Cluster.BusinessID)
	if err != nil {
		blog.Errorf("UpdateNodeGroup[%s] UpdateNodeGroupCloudAndModuleInfo failed: %v", cloudName, err)
		return nil, err
	}

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

// RecommendNodeGroupConf recommends nodegroup configs
func (ng *NodeGroup) RecommendNodeGroupConf(
	ctx context.Context, opt *cloudprovider.CommonOption) ([]*proto.RecommendNodeGroupConf, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// GetNodesInGroup get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroup(group *proto.NodeGroup, opt *cloudprovider.CommonOption) ([]*proto.Node, error) {
	// just get from cluster-manager storage no more implementation
	// already done in action part
	return nil, nil
}

// MoveNodesToGroup add cluster nodes to NodeGroup
func (ng *NodeGroup) MoveNodesToGroup(
	nodes []*proto.Node, group *proto.NodeGroup, opt *cloudprovider.MoveNodesOption) (*proto.Task, error) {
	// just update cluster-manager nodes belong to NodeGroup in local storage
	// already done in action part
	return nil, nil
}

// RemoveNodesFromGroup remove nodes from NodeGroup, nodes are still in cluster
func (ng *NodeGroup) RemoveNodesFromGroup(
	nodes []*proto.Node, group *proto.NodeGroup, opt *cloudprovider.RemoveNodesOption) error {
	// just remove nodes that belong to NodeGroup to cluster-manager in local storage
	// but nodes still are under controlled by cluster, so no other operation needed
	// already done in action part
	return nil
}

// CleanNodesInGroup clean specified nodes in NodeGroup and destroy machine in yunti cloud-provider
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

// CreateAutoScalingOption create cluster autoscaling option, cloudprovider will
// deploy cluster-autoscaler in backgroup according cloudprovider implementation
func (ng *NodeGroup) CreateAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.CreateScalingOption) (*proto.Task, error) {

	return nil, nil
	// return nil, cloudprovider.ErrCloudNotImplemented
}

// DeleteAutoScalingOption delete cluster autoscaling, cloudprovider will clean
// cluster-autoscaler in backgroup according cloudprovider implementation
func (ng *NodeGroup) DeleteAutoScalingOption(scalingOption *proto.ClusterAutoScalingOption,
	opt *cloudprovider.DeleteScalingOption) (*proto.Task, error) {

	return nil, nil
	// return nil, cloudprovider.ErrCloudNotImplemented
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

// GetNodesInGroupV2 get all nodes belong to NodeGroup
func (ng *NodeGroup) GetNodesInGroupV2(group *proto.NodeGroup, opt *cloudprovider.CommonOption) (
	[]*proto.NodeGroupNode, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// SwitchAutoScalingOptionStatus switch autoscaler component
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

// GetProjectCaResourceQuota get project ca resource quota
func (ng *NodeGroup) GetProjectCaResourceQuota(groups []*proto.NodeGroup, // nolint
	opt *cloudprovider.CommonOption) ([]*proto.ProjectAutoscalerQuota, error) {

	// 仅统计CA云梯资源 & 获取项目下所有节点池的资源使用情况 & 资源quota情况

	filterGroups := make([]*proto.NodeGroup, 0)
	// filter yunti ca nodeGroup
	for i := range groups {
		if !utils.StringInSlice(groups[i].GetNodeGroupType(), []string{common.Normal.String(), ""}) {
			continue
		}

		if groups[i].GetExtraInfo() != nil {
			rt, ok := groups[i].ExtraInfo[resource.ResourcePoolType]
			if ok &&
				utils.StringContainInSlice(rt, []string{resource.SelfPool, resource.CrPool}) {
				continue
			}
		}

		if groups[i].GetRegion() == "" || groups[i].GetLaunchTemplate().GetInstanceType() == "" {
			continue
		}

		filterGroups = append(filterGroups, groups[i])
	}

	var (
		lock         = sync.Mutex{}
		insZoneQuota = make(map[string]map[string]*proto.ProjectAutoscalerQuota, 0)
	)

	barrier := utils.NewRoutinePool(20)
	defer barrier.Close()

	for i := range filterGroups {
		barrier.Add(1)

		go func(group *proto.NodeGroup) {
			defer barrier.Done()

			// 地域-机型 维度的 资源池 和 可用区列表
			zonePools, resourceZones, err := tresource.GetResourceManagerClient().ListRegionZonePools(
				context.Background(), resource.YunTiPool,
				group.GetRegion(), group.GetLaunchTemplate().GetInstanceType())
			if err != nil {
				blog.Errorf("GetProjectCaResourceQuota[%s:%s] ListRegionZonePools failed: %v",
					group.ProjectID, group.NodeGroupID, err)
				return
			}
			if len(resourceZones) == 0 || len(zonePools) == 0 {
				blog.Errorf("GetProjectCaResourceQuota[%s:%s] region[%s] instanceType[%s] 无可用区机型",
					group.ProjectID, group.NodeGroupID, group.GetRegion(), group.GetLaunchTemplate().GetInstanceType())
				return
			}

			total, cur, pre, err := daemon.GetGroupCurAndPredictNodes(cloudprovider.GetStorageModel(),
				group.GetNodeGroupID(), resourceZones)
			if err != nil {
				blog.Errorf("GetProjectCaResourceQuota[%s:%s] failed: %v", group.GetProjectID(), group.GetNodeGroupID(), err)
				return
			}

			blog.Infof("GetProjectCaResourceQuota[%s] cur[%v] max[%v] currentNodes[%v] preNodes[%v] totalNodes[%v]",
				group.GetNodeGroupID(), group.GetAutoScaling().GetDesiredSize(), group.GetAutoScaling().GetMaxSize(),
				cur, pre, total)

			lock.Lock()

			if insZoneQuota[group.GetLaunchTemplate().GetInstanceType()] == nil {
				insZoneQuota[group.GetLaunchTemplate().GetInstanceType()] = make(map[string]*proto.ProjectAutoscalerQuota, 0)
			}

			// 统计已使用
			for zone, num := range cur {
				_, ok := insZoneQuota[group.GetLaunchTemplate().GetInstanceType()][zone]
				if !ok {
					insZoneQuota[group.GetLaunchTemplate().GetInstanceType()][zone] = &proto.ProjectAutoscalerQuota{
						InstanceType: group.GetLaunchTemplate().GetInstanceType(),
						Region:       group.GetRegion(),
						Zone:         zone,
						Used:         uint32(num),
						GroupIds:     []string{group.GetNodeGroupID()},
					}
				} else {
					insZoneQuota[group.GetLaunchTemplate().GetInstanceType()][zone].Used += uint32(num)
					insZoneQuota[group.GetLaunchTemplate().GetInstanceType()][zone].GroupIds = append(
						insZoneQuota[group.GetLaunchTemplate().GetInstanceType()][zone].GroupIds,
						group.GetNodeGroupID())
				}
			}

			// 统计quota总量
			for zone, num := range total {
				_, ok := insZoneQuota[group.GetLaunchTemplate().GetInstanceType()][zone]
				if !ok {
					insZoneQuota[group.GetLaunchTemplate().GetInstanceType()][zone] = &proto.ProjectAutoscalerQuota{
						InstanceType: group.GetLaunchTemplate().GetInstanceType(),
						Region:       group.GetRegion(),
						Zone:         zone,
						Total:        uint32(num),
					}
				} else {
					insZoneQuota[group.GetLaunchTemplate().GetInstanceType()][zone].Total += uint32(num)
				}
			}
			lock.Unlock()
		}(filterGroups[i])
	}
	barrier.Wait()

	projectQuotas := make([]*proto.ProjectAutoscalerQuota, 0)
	for _, zoneQuota := range insZoneQuota {
		for _, quota := range zoneQuota {
			projectQuotas = append(projectQuotas, quota)
		}
	}

	for _, projectQuota := range projectQuotas {
		matchProjectAutoscalerQuotaByGroup(filterGroups, projectQuota)
	}

	return projectQuotas, nil
}

func matchProjectAutoscalerQuotaByGroup(groups []*proto.NodeGroup, projectQuota *proto.ProjectAutoscalerQuota) {
	if projectQuota.GetTotalGroupIds() == nil {
		projectQuota.TotalGroupIds = make([]string, 0)
	}

	for _, group := range groups {
		if group.Region != projectQuota.Region {
			continue
		}
		if group.GetLaunchTemplate().GetInstanceType() != projectQuota.InstanceType {
			continue
		}

		// 任意可用区 && 指定可用区
		if group.GetAutoScaling().GetZones() == nil || len(group.GetAutoScaling().GetZones()) == 0 ||
			(len(group.GetAutoScaling().Zones) == 1 && group.GetAutoScaling().Zones[0] == "") ||
			utils.StringInSlice(projectQuota.Zone, group.GetAutoScaling().GetZones()) {
			projectQuota.TotalGroupIds = append(projectQuota.TotalGroupIds, group.NodeGroupID)
			continue
		}
	}
}

// CheckResourcePoolQuota check resource pool quota when revise group limit
func (ng *NodeGroup) CheckResourcePoolQuota(
	ctx context.Context, group *proto.NodeGroup, operation string, scaleUpNum uint32) error { // nolint
	cloud, err := cloudprovider.GetCloudByProvider(cloudName)
	if err == nil && cloud.GetConfInfo().GetDisableCheckGroupResource() {
		return nil
	}

	if !utils.StringInSlice(group.GetNodeGroupType(), []string{common.Normal.String(), ""}) {
		return nil
	}

	if group.GetExtraInfo() != nil && utils.StringContainInSlice(group.ExtraInfo[resource.ResourcePoolType],
		[]string{resource.SelfPool, resource.CrPool}) {
		return nil
	}

	if group.GetRegion() == "" || group.GetLaunchTemplate().GetInstanceType() == "" || scaleUpNum <= 0 {
		return nil
	}

	// 节点池配置任意可用区
	anyZone := func() bool {
		if group.GetAutoScaling().GetZones() == nil {
			return true
		}

		if len(group.GetAutoScaling().GetZones()) == 1 && group.GetAutoScaling().GetZones()[0] == "" {
			return true
		}

		return false
	}()

	quotaGrayValue, err := project.GetProjectManagerClient().CheckProjectQuotaGrayLabel(ctx, group.GetProjectID())
	if err != nil {
		blog.Errorf("GetProjectManagerClient GetProjectQuotaGrayLabel[%s] failed: %v",
			group.GetProjectID(), err)
		return err
	}

	switch quotaGrayValue {
	case project.QuotaGrayOverMode:
		if operation != "" {
			scaleUpNum = group.GetAutoScaling().GetMaxSize() + scaleUpNum
		}
		return ng.checkRpqByPjModeOversold(group, scaleUpNum, anyZone)
	case project.QuotaGrayNormalMode:
		return ng.checkRpqByPjModeNormal(group, scaleUpNum, anyZone)
	default:
	}

	return ng.checkRpqByRm(group, scaleUpNum, anyZone)
}

func (ng *NodeGroup) checkRpqByPjModeOversold(group *proto.NodeGroup, scaleUpNum uint32, anyZone bool) error {
	pools, err := ng.getResourcePoolByProjectQuotaList(group.GetProjectID(), group.GetRegion(),
		group.GetLaunchTemplate().GetInstanceType())
	if err != nil {
		blog.Errorf("GetProjectManagerClient GetResourcePoolByProjectQuotaList[%s:%s] failed: %v",
			group.GetProjectID(), group.GetNodeGroupID(), err)
		return err
	}
	var (
		poolTotal int32
	)

	// 任意可用区
	if anyZone {
		for i := range pools {
			poolTotal += pools[i].Total
		}

		blog.Infof("cloud[%s] checkResourcePoolQuotaByPjModeOversold region[%s] zone[%s] instanceType[%s] "+
			"poolTotal[%v] scaleUpNum[%v]", cloudName, group.Region, "anyZone",
			group.GetLaunchTemplate().GetInstanceType(), poolTotal, scaleUpNum)

		if int32(scaleUpNum) > poolTotal {
			return errors.New(fmt.Sprintf("anyZone region[%s] instanceType[%s] ",
				group.GetRegion(), group.GetLaunchTemplate().GetInstanceType()) + poolInsufficientQuotaMessage.Error())
		}

		return nil
	}

	// 指定可用区
	selectedZones := group.GetAutoScaling().GetZones()
	for i := range pools {
		if utils.StringInSlice(pools[i].Zone, selectedZones) {
			poolTotal += pools[i].Total
		}
	}

	blog.Infof("cloud[%s] checkResourcePoolQuotaByPjModeOversold[%s] zones[%+v] poolTotal[%v] scaleUpNum[%v]",
		cloudName, group.GetNodeGroupID(), selectedZones, poolTotal, scaleUpNum)

	if int32(scaleUpNum) > poolTotal {
		return errors.New(fmt.Sprintf("region[%s] zone[%s] instanceType[%s]",
			group.GetRegion(), strings.Join(group.GetAutoScaling().GetZones(), ","),
			group.GetLaunchTemplate().GetInstanceType()) + poolInsufficientQuotaMessage.Error())
	}

	return nil
}

func (ng *NodeGroup) checkRpqByPjModeNormal(group *proto.NodeGroup, scaleUpNum uint32, anyZone bool) error {

	// 项目维度的 地域-机型 资源池配置
	resourcePoolZones, resourceZones, err := ng.getPoolsZonesByProjectQuotaList(group.GetProjectID(), group.GetRegion(),
		group.GetLaunchTemplate().GetInstanceType())
	if err != nil {
		blog.Errorf("checkResourcePoolQuotaByPjModeNormal GetResourcePoolByProjectQuotaList[%s:%s] failed: %v",
			group.GetProjectID(), group.GetNodeGroupID(), err)
		return err
	}

	// 项目下所有节点池的资源使用情况
	pools, err := ng.getProjectRegionDevicePoolDetail(group.GetProjectID(), group.GetRegion(),
		group.GetLaunchTemplate().GetInstanceType(), resourcePoolZones, resourceZones)
	if err != nil {
		blog.Errorf("checkResourcePoolQuotaByPjModeNormal getProjectRegionDevicePoolDetail[%s:%s:%s] failed: %v",
			group.GetProjectID(), group.GetRegion(), group.GetLaunchTemplate().GetInstanceType(), err)
		return err
	}

	// nodegroup config any zone
	if anyZone {
		var (
			poolTotal  int32
			groupQuota int32
		)
		for i := range pools {
			poolTotal += pools[i].Total
			groupQuota += int32(pools[i].GroupQuota)
		}

		blog.Infof("cloud[%s] checkResourcePoolQuotaByPjModeNormal[%s] anyZone poolTotal[%v] "+
			"groupQuota[%v] scaleUpNum[%v]", cloudName, group.GetNodeGroupID(), poolTotal, groupQuota, scaleUpNum)

		if groupQuota+int32(scaleUpNum) > poolTotal {
			return errors.New(fmt.Sprintf("anyZone region[%s] instanceType[%s] ",
				group.GetRegion(), group.GetLaunchTemplate().GetInstanceType()) + poolInsufficientQuotaMessage.Error())
		}
		return nil
	}

	// 指定可用区域时进行预测分配, 每个资源域分配资源多少
	zoneNum := daemon.AllocateZoneResource(group.GetNodeGroupID(), group.GetRegion(),
		group.GetLaunchTemplate().GetInstanceType(), group.GetAutoScaling().GetZones(), resourceZones, int(scaleUpNum))

	blog.Infof("cloud[%s] checkResourcePoolQuotaByPjModeNormal[%s] zoneNum[%+v]",
		cloudName, group.GetNodeGroupID(), zoneNum)

	mulErrors := utils.NewMultiError()
	// 检验配额是否充足
	for i := range pools {
		num, ok := zoneNum[pools[i].Zone]
		if ok && num > 0 && (pools[i].GroupQuota+num) > int(pools[i].Total) {
			mulErrors.Append(fmt.Errorf("region[%s] zone[%s] instanceType[%s]",
				group.GetRegion(), pools[i].Zone, group.GetLaunchTemplate().GetInstanceType()))
		}
	}

	if mulErrors.HasErrors() {
		mulErrors.Append(poolInsufficientQuotaMessage)
		return mulErrors
	}

	return nil
}

func (ng *NodeGroup) checkRpqByRm(group *proto.NodeGroup, scaleUpNum uint32, anyZone bool) error {
	// 获取当前资源池的使用情况 & 超卖情况
	pools, err := daemon.GetRegionDevicePoolDetail(cloudprovider.GetStorageModel(), group.GetRegion(),
		group.GetLaunchTemplate().GetInstanceType(), nil)
	if err != nil {
		return fmt.Errorf("get region %s instanceType %s device pool detail failed, %s",
			group.GetRegion(), group.GetLaunchTemplate().GetInstanceType(), err.Error())
	}

	// 机型资源所在的可用区
	resourceZones := make([]string, 0)
	for _, pool := range pools {
		blog.Infof("cloud[%s] checkResourcePoolQuotaByRm pool[%s] region[%s] zone[%s] instanceType[%s] "+
			"poolTotal[%v] poolAvailable[%v] poolOversoldTotal[%v] poolOversoldAvailable[%v] groupQuota[%v] "+
			"groupUsed[%v]", cloudName, pool.PoolId, pool.Region, pool.Zone, pool.InstanceType,
			pool.Total, pool.Available, pool.OversoldTotal, pool.OversoldAvailable, pool.GroupQuota, pool.GroupUsed)

		resourceZones = append(resourceZones, pool.Zone)
	}

	// nodegroup config any zone
	if anyZone {
		var (
			poolTotal  int32
			groupQuota int32
		)
		for i := range pools {
			poolTotal += pools[i].OversoldTotal
			groupQuota += int32(pools[i].GroupQuota)
		}

		blog.Infof("cloud[%s] checkResourcePoolQuotaByRm[%s] anyZone poolTotal[%v] groupQuota[%v] scaleUpNum[%v]",
			cloudName, group.GetNodeGroupID(), poolTotal, groupQuota, scaleUpNum)

		if groupQuota+int32(scaleUpNum) > poolTotal {
			return errors.New(fmt.Sprintf("anyZone region[%s] instanceType[%s] ",
				group.GetRegion(), group.GetLaunchTemplate().GetInstanceType()) + poolInsufficientQuotaMessage.Error())
		}
		return nil
	}

	// 指定可用区域时进行预测分配, 每个资源域分配资源多少
	zoneNum := daemon.AllocateZoneResource(group.GetNodeGroupID(), group.GetRegion(),
		group.GetLaunchTemplate().GetInstanceType(), group.GetAutoScaling().GetZones(), resourceZones, int(scaleUpNum))

	blog.Infof("cloud[%s] checkResourcePoolQuotaByRm[%s] zoneNum[%+v]", cloudName, group.GetNodeGroupID(), zoneNum)

	mulErrors := utils.NewMultiError()
	// 检验配额是否充足
	for i := range pools {
		num, ok := zoneNum[pools[i].Zone]
		if ok && num > 0 && (pools[i].GroupQuota+num) > int(pools[i].OversoldTotal) {
			mulErrors.Append(fmt.Errorf("region[%s] zone[%s] instanceType[%s]",
				group.GetRegion(), pools[i].Zone, group.GetLaunchTemplate().GetInstanceType()))
		}
	}

	if mulErrors.HasErrors() {
		mulErrors.Append(poolInsufficientQuotaMessage)
		return mulErrors
	}

	return nil
}

// getResourcePoolByProjectQuotaList get resource quota from zoneResource by project quota list info
func (ng *NodeGroup) getResourcePoolByProjectQuotaList(projectId, region, instanceType string) (
	[]*resource.DevicePoolInfo, error) {

	listProjectQuotasData, err := project.GetProjectManagerClient().ListProjectQuotas(projectId,
		project.ProjectQuotaHostType, project.ProjectQuotaProvider)
	if err != nil {
		blog.Errorf("GetProjectManagerClient ListProjectQuotas[%s:%s:%s] failed: %v",
			projectId, project.ProjectQuotaHostType, project.ProjectQuotaProvider, err)
		return nil, err
	}

	resourcePools := make([]*resource.DevicePoolInfo, 0)
	projectQuotaLists := listProjectQuotasData.GetResults()

	for _, projectQuota := range projectQuotaLists {
		zoneResources := projectQuota.GetQuota().GetZoneResources()

		if (region != "" && zoneResources.GetRegion() != region) ||
			(instanceType != "" && zoneResources.GetInstanceType() != instanceType) {
			continue
		}

		if projectQuota.GetStatus() != common.StatusRunning {
			continue
		}

		zonePool := &resource.DevicePoolInfo{
			PoolId:       projectQuota.GetQuotaId(),
			PoolName:     projectQuota.GetQuotaName(),
			Region:       zoneResources.GetRegion(),
			Zone:         zoneResources.GetZoneName(),
			InstanceType: zoneResources.GetInstanceType(),
			Total:        int32(zoneResources.GetQuotaNum()),
			// 每个可用区实际使用的资源数
			Used:      int32(zoneResources.GetQuotaUsed()),
			Available: int32(zoneResources.GetQuotaNum() - zoneResources.GetQuotaUsed()),

			Status: projectQuota.GetStatus(),
		}

		resourcePools = append(resourcePools, zonePool)
	}
	return resourcePools, nil
}

func (ng *NodeGroup) getPoolsZonesByProjectQuotaList(projectId, region, instanceType string) (
	map[string]*resource.DevicePoolInfo, []string, error) {

	// 项目维度的 地域-机型 资源池配置
	pools, err := ng.getResourcePoolByProjectQuotaList(projectId, region, instanceType)
	if err != nil {
		blog.Errorf("getPoolsZonesByProjectQuotaList[%s:%s:%s] failed: %v",
			projectId, region, instanceType, err)
		return nil, nil, err
	}

	var (
		// zone级别资源池
		resourcePoolZones = make(map[string]*resource.DevicePoolInfo, 0)
		// 资源所在可用区
		resourceZones = make([]string, 0)
	)
	for _, pool := range pools {
		blog.Infof("getPoolsZonesByProjectQuotaList pool[%s] region[%s] zone[%s] instanceType[%s] "+
			"poolTotal[%v] poolAvailable[%v] poolUsed[%v]", pool.PoolId, pool.Region, pool.Zone,
			pool.InstanceType, pool.Total, pool.Available, pool.Used)

		resourcePoolZones[pool.Zone] = pool
		resourceZones = append(resourceZones, pool.Zone)
	}

	return resourcePoolZones, resourceZones, nil
}

// filterProjectGroupsByRegionInsType filter region instanceType nodeGroup in project
func filterProjectGroupsByRegionInsType(model store.ClusterManagerModel, projectId, region,
	instanceType string) ([]*proto.NodeGroup, error) {
	normalYunti, _, _, err := daemon.GetNodeGroups(model)
	if err != nil {
		return nil, err
	}

	filterGroups := make([]*proto.NodeGroup, 0)
	for _, group := range normalYunti {
		if group.ProjectID == projectId && group.Region == region &&
			group.GetLaunchTemplate().GetInstanceType() == instanceType {
			filterGroups = append(filterGroups, group)
		}
	}

	return filterGroups, nil
}

// getProjectRegionDevicePoolDetail get region device pool detail
func (ng *NodeGroup) getProjectRegionDevicePoolDetail(projectId, region string, instanceType string,
	zonePools map[string]*resource.DevicePoolInfo, resourceZones []string) ([]*resource.DevicePoolInfo, error) {

	// 过滤项目下的yunti节点池列表
	filterGroups, err := filterProjectGroupsByRegionInsType(cloudprovider.GetStorageModel(),
		projectId, region, instanceType)
	if err != nil {
		blog.Errorf("GetProjectRegionDevicePoolDetail[%s:%s] FilterGroupsByRegionInsType failed: %v",
			region, instanceType, err)
		return nil, err
	}

	// 项目-地域-机型 维度的 资源池 和 资源可用区列表
	if len(resourceZones) == 0 || len(zonePools) == 0 {
		blog.Errorf("GetProjectRegionDevicePoolDetail region[%s] instanceType[%s] 无可用区机型",
			region, instanceType)
		return nil, fmt.Errorf("GetProjectRegionDevicePoolDetail region[%s] instanceType[%s] 无可用区机型",
			region, instanceType)
	}

	// 当前需要如何分配 (只要机器足够即可，机器不够的情况，简单按照平均即可)
	for _, group := range filterGroups {
		// 获取当前已存在节点池资源分布情况 (包括当前已经分配 & 预测分配)
		nodesDistribution, curDistribution, _, errLocal := daemon.GetGroupCurAndPredictNodes(
			cloudprovider.GetStorageModel(), group.NodeGroupID, resourceZones)
		if errLocal != nil {
			blog.Errorf("nodeGroup[%s] GetProjectRegionDevicePoolDetail[%s:%s] GetGroupCurAndPredictNodes "+
				"failed: %v", group.GetNodeGroupID(), region, instanceType, errLocal)
			continue
		}

		blog.Infof("GetProjectRegionDevicePoolDetail GetGroupCurAndPredictNodes[%v] %+v %+v", group.NodeGroupID,
			nodesDistribution, curDistribution)

		for zone := range nodesDistribution {
			_, ok := zonePools[zone]
			if ok {
				zonePools[zone].GroupQuota += nodesDistribution[zone]
			}
		}

		for zone := range curDistribution {
			_, ok := zonePools[zone]
			if ok {
				zonePools[zone].GroupUsed += curDistribution[zone]
			}
		}
	}

	pools := make([]*resource.DevicePoolInfo, 0)

	for i := range zonePools {
		blog.Infof("GetProjectRegionDevicePoolDetail region[%s] zone[%s] instanceType[%s] pool[%s] "+
			"poolTotal[%v] poolAvailable[%v] poolOversoldTotal[%v] poolOversoldAvailable[%v] poolUsed[%v] "+
			"groupQuota[%v] groupUsed[%v]", region, zonePools[i].Zone, instanceType, zonePools[i].PoolId,
			zonePools[i].Total, zonePools[i].Available, zonePools[i].OversoldTotal, zonePools[i].OversoldAvailable,
			zonePools[i].Used, zonePools[i].GroupQuota, zonePools[i].GroupUsed)
		pools = append(pools, zonePools[i])
	}

	return pools, nil
}
