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

package task

import (
	"context"
	"errors"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

func allowTaskRetry(task *proto.Task) bool {
	allowRetry := true
	// attention: 开启CA节点自动扩缩容的任务不允许手动重试
	if utils.SliceContainInString([]string{cloudprovider.UpdateNodeGroupDesiredNode.String(),
		cloudprovider.CleanNodeGroupNodes.String()}, task.TaskType) && (task.GetCommonParams() != nil &&
		task.GetCommonParams()[cloudprovider.ManualKey.String()] != common.True) {
		allowRetry = false
	}
	return allowRetry
}

func updateTaskDataStatus(model store.ClusterManagerModel, task *proto.Task) error {
	blog.Infof("updateTaskDataStatus[%s] taskType[%s]", task.TaskID, task.TaskType)

	var err error
	switch {
	case strings.Contains(task.TaskType, cloudprovider.CreateCluster.String()),
		strings.Contains(task.TaskType, cloudprovider.ImportCluster.String()),
		strings.Contains(task.TaskType, cloudprovider.CreateVirtualCluster.String()):
		err = updateClusterStatus(model, task.ClusterID, common.StatusInitialization)
	case strings.Contains(task.TaskType, cloudprovider.DeleteCluster.String()),
		strings.Contains(task.TaskType, cloudprovider.DeleteVirtualCluster.String()):
		err = updateClusterStatus(model, task.ClusterID, common.StatusDeleting)
	case strings.Contains(task.TaskType, cloudprovider.AddNodesToCluster.String()):
		// err = updateNodeStatus(ua.model, ua.task.NodeIPList, common.StatusInitialization)
		err = updateNodeStatus(model, cloudprovider.ParseNodeIpOrIdFromCommonMap(task.CommonParams,
			cloudprovider.NodeIPsKey.String(), ","), common.StatusInitialization)
	case strings.Contains(task.TaskType, cloudprovider.RemoveNodesFromCluster.String()):
		err = updateNodeStatus(model, task.NodeIPList, common.StatusDeleting)
	case strings.Contains(task.TaskType, cloudprovider.CreateNodeGroup.String()):
		err = updateNodeGroupStatus(model, task.NodeGroupID, common.StatusCreateNodeGroupCreating)
	case strings.Contains(task.TaskType, cloudprovider.DeleteNodeGroup.String()):
		err = updateNodeGroupStatus(model, task.NodeGroupID, common.StatusDeleteNodeGroupDeleting)
	case strings.HasSuffix(task.TaskType, cloudprovider.UpdateNodeGroup.String()):
		err = updateNodeGroupStatus(model, task.NodeGroupID, common.StatusUpdateNodeGroupUpdating)
	case strings.HasSuffix(task.TaskType, cloudprovider.UpdateNodeGroupDesiredNode.String()):
		err = updateNodeStatus(model, cloudprovider.ParseNodeIpOrIdFromCommonMap(task.CommonParams,
			cloudprovider.NodeIPsKey.String(), ","), common.StatusInitialization)
	case strings.Contains(task.TaskType, cloudprovider.CleanNodeGroupNodes.String()):
		err = updateNodeStatus(model, cloudprovider.ParseNodeIpOrIdFromCommonMap(task.CommonParams,
			cloudprovider.NodeIPsKey.String(), ","), common.StatusInitialization)
	default:
		blog.Warnf("updateTaskDataStatus[%s] not support taskType[%s]", task.TaskID, task.TaskType)
	}
	if err != nil {
		return err
	}

	return nil
}

func updateNodeGroupStatus(model store.ClusterManagerModel, nodeGroupID, status string) error {
	group, err := model.GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("updateNodeGroupStatus[%s] GetNodeGroup failed: %v", nodeGroupID, err)
		return err
	}
	if group == nil {
		blog.Warnf("updateNodeGroupStatus[%s] not found nodegroup", nodeGroupID)
		return nil
	}

	group.Status = status
	err = model.UpdateNodeGroup(context.Background(), group)
	if err != nil {
		blog.Errorf("updateNodeGroupStatus[%s] UpdateNodeGroup failed: %v", nodeGroupID, err)
		return err
	}

	return nil
}

func updateClusterStatus(model store.ClusterManagerModel, clusterID, status string) error {
	cluster, err := model.GetCluster(context.Background(), clusterID)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("updateClusterStatus[%s] GetCluster failed: %v", clusterID, err)
		return err
	}
	if cluster == nil {
		blog.Warnf("updateClusterStatus[%s] not found cluster", clusterID)
		return nil
	}

	cluster.Status = status
	cluster.Message = ""
	err = model.UpdateCluster(context.Background(), cluster)
	if err != nil {
		blog.Errorf("updateClusterStatus[%s] UpdateCluster failed: %v", clusterID, err)
		return err
	}

	return nil
}

func updateNodeStatus(model store.ClusterManagerModel, nodeList []string, status string) error {
	for i := range nodeList {
		node, err := model.GetNodeByIP(context.Background(), nodeList[i])
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("updateNodeStatus[%s] GetNodeByIP failed: %v", nodeList[i], err)
			return err
		}
		if node == nil {
			blog.Warnf("updateNodeStatus[%s] not found node", nodeList[i])
			continue
		}

		node.Status = status
		err = model.UpdateNode(context.Background(), node)
		if err != nil {
			blog.Errorf("updateNodeStatus[%s] UpdateNode failed: %v", nodeList[i], err)
			return err
		}
	}

	return nil
}
