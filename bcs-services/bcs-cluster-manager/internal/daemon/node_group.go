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

// Package daemon for daemon
package daemon

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// GetNodeGroups get cluster node groups && check yunti groups
func GetNodeGroups(model store.ClusterManagerModel) ([]cmproto.NodeGroup,
	[]cmproto.NodeGroup, []cmproto.NodeGroup, error) {
	condGroup := operator.NewLeafCondition(operator.Eq, operator.M{
		"status": common.StatusRunning,
	})
	groupList, err := model.ListNodeGroup(context.Background(), condGroup, &storeopt.ListOption{All: true})
	if err != nil {
		blog.Errorf("GetNodeGroups ListNodeGroup failed: %v", err)
		return nil, nil, nil, err
	}

	var normalYunti, normalSelf, crSelf, external = make([]cmproto.NodeGroup, 0), make([]cmproto.NodeGroup, 0),
		make([]cmproto.NodeGroup, 0), make([]cmproto.NodeGroup, 0)

	for _, group := range groupList {
		switch group.NodeGroupType {
		// 普通节点池: yunti  资源池； 自建资源池
		case common.Normal.String(), "":
			pool, ok := group.ExtraInfo[resource.ResourcePoolType]
			if ok {
				switch pool {
				case resource.SelfPool:
					normalSelf = append(normalSelf, group)
					continue
				case resource.CrPool:
					crSelf = append(crSelf, group) // nolint
					continue
				case resource.BcsResourcePool:
					continue
				default:
				}
			}
			normalYunti = append(normalYunti, group)
		// 第三方节点池
		case common.External.String():
			external = append(external, group)
		default:
			blog.Infof("GetNodeGroups not supported group[%s]", group.NodeGroupID)
		}
	}

	// 记录yunti资源池是否是 yuntiProvider
	for _, group := range normalYunti {
		consumer, errLocal := tresource.GetResourceManagerClient().GetDeviceConsumer(
			context.Background(), group.ConsumerID)
		if errLocal != nil {
			blog.Errorf("GetNodeGroups GetDeviceConsumer failed: %v", errLocal)
			continue
		}

		if consumer.GetProvider() != resource.YunTiPool {
			blog.Infof("resourceProvider %s cluster %s nodeGroup %s resourceType %s", consumer.GetProvider(),
				group.ClusterID, group.NodeGroupID, group.ExtraInfo[resource.ResourcePoolType])
		}
	}

	return normalYunti, normalSelf, external, nil
}

// FilterGroupsByRegionInsType filter region instanceType nodeGroup
func FilterGroupsByRegionInsType(model store.ClusterManagerModel, region,
	instanceType string) ([]cmproto.NodeGroup, error) {
	normalYunti, _, _, err := GetNodeGroups(model)
	if err != nil {
		return nil, err
	}

	filterGroups := make([]cmproto.NodeGroup, 0)
	for _, group := range normalYunti {
		if group.Region == region && group.GetLaunchTemplate().GetInstanceType() == instanceType {
			filterGroups = append(filterGroups, group)
		}
	}

	return filterGroups, nil
}

// GetNodeGroupAndNodes get node group and nodes
func GetNodeGroupAndNodes(model store.ClusterManagerModel,
	groupId string) (*cmproto.NodeGroup, []*cmproto.Node, error) {
	group, err := model.GetNodeGroup(context.Background(), groupId)
	if err != nil {
		blog.Errorf("GetNodeGroupAndNodes %s failed: %s", group, err.Error())
		return nil, nil, err
	}

	condM := make(operator.M)
	condM["nodegroupid"] = groupId
	cond := operator.NewLeafCondition(operator.Eq, condM)
	nodes, err := model.ListNode(context.Background(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Errorf("GetNodeGroupAndNodes %s ListNode failed, %s", group.NodeGroupID, err.Error())
		return nil, nil, err
	}

	groupNodes := make([]*cmproto.Node, 0)

	// filter not running
	for i := range nodes {
		if nodes[i].Status == common.StatusRunning {
			groupNodes = append(groupNodes, nodes[i])
		}
	}

	return group, groupNodes, nil
}

// GetGroupCurAndPredictNodes 获取当前节点池 节点分布情况，包括quota预测的分配情况
func GetGroupCurAndPredictNodes(model store.ClusterManagerModel, groupId string,
	resourceZones []string) (map[string]int, map[string]int, map[string]int, error) {
	// 获取节点池 可用区维度已分配的节点数
	curZoneNodes := make(map[string]int, 0)
	preZoneNodes := make(map[string]int, 0)

	// desiredSize 已分配的节点数目
	group, nodes, errLocal := GetNodeGroupAndNodes(model, groupId)
	if errLocal != nil {
		return nil, nil, nil, errLocal
	}
	for _, n := range nodes {
		_, ok := curZoneNodes[n.GetZoneID()]
		if ok {
			curZoneNodes[n.GetZoneID()]++
		} else {
			curZoneNodes[n.GetZoneID()] = 1
		}
	}

	// max - desired = quota, quota分配算法. 需要进行预测分配
	if group.GetAutoScaling().DesiredSize < group.GetAutoScaling().MaxSize {
		quota := group.GetAutoScaling().MaxSize - group.GetAutoScaling().DesiredSize
		zoneNums := AllocateZoneResource(groupId, group.GetRegion(), group.GetLaunchTemplate().GetInstanceType(),
			group.GetAutoScaling().GetZones(), resourceZones, int(quota))

		for zone, num := range zoneNums {
			_, ok := preZoneNodes[zone]
			if ok {
				preZoneNodes[zone] += num
				continue
			}

			preZoneNodes[zone] = num
		}
	}
	zoneNodes := utils.MergeStringIntMaps(curZoneNodes, preZoneNodes)

	blog.Infof("GetGroupCurAndPredictNodes[%s] cur[%v] max[%v] currentNodes[%v] preNodes[%v] totalNodes[%v]",
		groupId, group.GetAutoScaling().GetDesiredSize(), group.GetAutoScaling().GetMaxSize(),
		curZoneNodes, preZoneNodes, zoneNodes)

	return zoneNodes, curZoneNodes, preZoneNodes, nil
}

// AllocateZoneResource 可用区资源分配算法，目前采用均分方式. (单次资源申请个数相关、资源管理模块资源分配策略相关)
func AllocateZoneResource(group, region, insType string, groupZones []string, resourceZones []string,
	needResource int) map[string]int {
	var zoneResource = make(map[string]int, 0)

	// 节点池支持任意可用区, 可将所需资源随机分配至资源池
	if len(groupZones) == 0 || (len(groupZones) == 1 && groupZones[0] == "") {
		allocateResources := utils.DistributeMachines(needResource, resourceZones)

		for i := range resourceZones {
			zoneResource[resourceZones[i]] = allocateResources[i]
		}

		return zoneResource
	}

	// 节点池配置的可用区域 和 资源池可用区域取交集, 只取资源池的可用区
	zones := make([]string, 0)
	for i := range groupZones {
		if utils.StringInSlice(groupZones[i], resourceZones) {
			zones = append(zones, groupZones[i])
		} else {
			// 表示节点池可用区配置了资源池不存在的可用区, 可做日志告警
			blog.Errorf("region[%s] instanceType[%s] 在可用区[%s]不存在资源池", region, insType, groupZones[i])
		}
	}

	// 没有交集 说明group配置的可用区 在资源池中不存在
	if len(zones) == 0 {
		blog.Errorf("group[%s] region[%s] instanceType[%s] 配置的可用区不存在资源池", group, region, insType)
		return zoneResource
	}

	// 指定可用区资源分配
	allocateResources := utils.DistributeMachines(needResource, zones)

	for i := range zones {
		zoneResource[zones[i]] = allocateResources[i]
	}

	return zoneResource
}
