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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// allowTaskRetry 判断任务是否允许重试
func allowTaskRetry(task *proto.Task) bool {
	allowRetry := true
	// 检查条件：
	// 1. 任务类型属于自动扩缩容相关类型（UpdateNodeGroupDesiredNode: 更新节点组期望节点数；CleanNodeGroupNodes: 清理节点组节点）
	// 2. 任务公共参数中ManualKey（手动触发标识）不为"true"（即非手动触发的自动任务）
	// 满足以上条件时，禁止重试
	if utils.SliceContainInString([]string{cloudprovider.UpdateNodeGroupDesiredNode.String(),
		cloudprovider.CleanNodeGroupNodes.String()}, task.TaskType) && (task.GetCommonParams() != nil &&
		task.GetCommonParams()[cloudprovider.ManualKey.String()] != common.True) {
		allowRetry = false
	}
	return allowRetry
}

// updateTaskDataStatus update task status
// updateTaskDataStatus 根据任务类型更新关联资源（集群/节点/节点组）的状态
// 参数：
//
//	model - 存储模型接口，用于操作数据库
//	task - 当前需要处理状态更新的任务对象
//
// 返回值：操作过程中产生的错误（无错误返回nil）
func updateTaskDataStatus(model store.ClusterManagerModel, task *proto.Task) error {
	blog.Infof("updateTaskDataStatus[%s] taskType[%s]", task.TaskID, task.TaskType)

	var err error
	switch {
	// 集群创建/导入/虚拟集群创建类任务：更新集群状态为初始化中
	case strings.Contains(task.TaskType, cloudprovider.CreateCluster.String()),
		strings.Contains(task.TaskType, cloudprovider.ImportCluster.String()),
		strings.Contains(task.TaskType, cloudprovider.CreateVirtualCluster.String()):
		err = updateClusterStatus(model, task.ClusterID, common.StatusInitialization)
	// 集群删除/虚拟集群删除类任务：更新集群状态为删除中
	case strings.Contains(task.TaskType, cloudprovider.DeleteCluster.String()),
		strings.Contains(task.TaskType, cloudprovider.DeleteVirtualCluster.String()):
		err = updateClusterStatus(model, task.ClusterID, common.StatusDeleting)
	// 节点添加类任务：从任务公共参数中解析节点IP列表，更新节点状态为初始化中
	case strings.Contains(task.TaskType, cloudprovider.AddNodesToCluster.String()):
		// 原注释：err = updateNodeStatus(ua.model, ua.task.NodeIPList, common.StatusInitialization)
		err = updateNodeStatus(model, cloudprovider.ParseNodeIpOrIdFromCommonMap(task.CommonParams,
			cloudprovider.NodeIPsKey.String(), ","), common.StatusInitialization)
	// 节点移除类任务：直接使用任务中的节点IP列表，更新节点状态为删除中
	case strings.Contains(task.TaskType, cloudprovider.RemoveNodesFromCluster.String()):
		err = updateNodeStatus(model, task.NodeIPList, common.StatusDeleting)
	// 节点组创建类任务：更新节点组状态为"创建中"
	case strings.Contains(task.TaskType, cloudprovider.CreateNodeGroup.String()):
		err = updateNodeGroupStatus(model, task.NodeGroupID, common.StatusCreateNodeGroupCreating)
	// 节点组删除类任务：更新节点组状态为"删除中"
	case strings.Contains(task.TaskType, cloudprovider.DeleteNodeGroup.String()):
		err = updateNodeGroupStatus(model, task.NodeGroupID, common.StatusDeleteNodeGroupDeleting)
	// 节点组更新类任务（后缀匹配）：更新节点组状态为"更新中"
	case strings.HasSuffix(task.TaskType, cloudprovider.UpdateNodeGroup.String()):
		err = updateNodeGroupStatus(model, task.NodeGroupID, common.StatusUpdateNodeGroupUpdating)
	// 节点组期望节点数更新类任务（后缀匹配）：解析节点IP列表，更新节点状态为初始化中
	case strings.HasSuffix(task.TaskType, cloudprovider.UpdateNodeGroupDesiredNode.String()):
		err = updateNodeStatus(model, cloudprovider.ParseNodeIpOrIdFromCommonMap(task.CommonParams,
			cloudprovider.NodeIPsKey.String(), ","), common.StatusInitialization)
	// 节点组节点清理类任务：解析节点IP列表，更新节点状态为删除中
	case strings.Contains(task.TaskType, cloudprovider.CleanNodeGroupNodes.String()):
		err = updateNodeStatus(model, cloudprovider.ParseNodeIpOrIdFromCommonMap(task.CommonParams,
			cloudprovider.NodeIPsKey.String(), ","), common.StatusDeleting)
	// 未匹配到支持的任务类型时记录警告日志
	default:
		blog.Warnf("updateTaskDataStatus[%s] not support taskType[%s]", task.TaskID, task.TaskType)
	}
	if err != nil {
		return err
	}

	return nil
}

// retryPartFailureTask retry part failure task
func retryPartFailureTask(model store.ClusterManagerModel, cluster *proto.Cluster,
	task *proto.Task) (*proto.Task, error) {
	switch {
	case strings.Contains(task.TaskType, cloudprovider.AddNodesToCluster.String()):
		return addNodesToClusterTask(model, cluster, task)
	default:
	}

	return nil, fmt.Errorf("retryPartFailureTask[%s] not support taskType[%s]", task.TaskID, task.TaskType)
}

// addNodesToClusterTask add nodes task
func addNodesToClusterTask(model store.ClusterManagerModel, cluster *proto.Cluster,
	task *proto.Task) (*proto.Task, error) {
	allFailedNodeIds, err := getAllFailedNodeIds(task)
	if err != nil {
		return nil, err
	}
	nodes := cloudprovider.GetNodesByInstanceIDs(allFailedNodeIds)
	_, nodeIps := getNodeIdsAndIps(nodes)

	// check node in cluster
	err = checkNodeInCluster(task, nodeIps)
	if err != nil {
		return nil, err
	}

	// get add nodes task params
	params, err := getAddNodesTaskParams(task)
	if err != nil {
		return nil, err
	}

	cloud, err := actions.GetCloudByCloudID(model, cluster.Provider)
	if err != nil {
		blog.Errorf("get Cluster %s provider %s failed, %s", cluster.ClusterID, cluster.Provider, err.Error())
		return nil, err
	}
	var nodeTemplate *proto.NodeTemplate
	if len(params.nodeTemplateID) > 0 {
		templateInfo, errGet := actions.GetNodeTemplateByTemplateID(model, params.nodeTemplateID)
		if errGet != nil {
			blog.Errorf("get Cluster %s getNodeTemplateByTemplateID %s failed, %s",
				cluster.ClusterID, params.nodeTemplateID, errGet.Error())
			return nil, errGet
		}
		nodeTemplate = templateInfo
	}

	clusterMgr, err := cloudprovider.GetClusterMgr(cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s ClusterManager for add nodes %v to Cluster %s failed, %s",
			cloud.CloudProvider, nodes, cluster.ClusterID, err.Error(),
		)
		return nil, err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s when add nodes %s to cluster %s failed, %s",
			cloud.CloudID, cloud.CloudProvider, nodes, cluster.ClusterID, err.Error(),
		)
		return nil, err
	}
	cmOption.Region = cluster.Region

	// default reinstall system when add node to cluster
	newTasks, err := clusterMgr.AddNodesToCluster(cluster, nodes, &cloudprovider.AddNodesOption{
		CommonOption: *cmOption,
		Cloud:        cloud,
		NodeTemplate: nodeTemplate,
		Operator:     params.operator,
		NodeSchedule: params.nodeSchedule,
		Advance:      params.advance,
	})
	if err != nil {
		blog.Errorf("cloudprovider %s addNodes %v to Cluster %s failed, %s",
			cloud.CloudProvider, nodes, cluster.ClusterID, err.Error(),
		)
		return nil, err
	}

	if len(newTasks) == 0 {
		retErr := fmt.Errorf("cluster[%s] add nodes task return empty task list", cluster.ClusterID)
		return nil, retErr
	}

	// update node status
	_ = updateNodeStatus(model, nodeIps, common.StatusInitialization)

	return newTasks[0], nil
}

// getNodeIdsAndIps get node ids and ips
func getNodeIdsAndIps(nodes []*proto.Node) ([]string, []string) {
	nodeIds := make([]string, 0, len(nodes))
	nodeIps := make([]string, 0, len(nodes))

	for _, node := range nodes {
		nodeIds = append(nodeIds, node.NodeID)
		nodeIps = append(nodeIps, node.InnerIP)
	}

	return nodeIds, nodeIps
}

// getAllFailedNodeIds get all failed node ids
func getAllFailedNodeIds(task *proto.Task) ([]string, error) { // nolint
	var allFailedNodeIds []string

	if task == nil || len(task.CommonParams) == 0 {
		return allFailedNodeIds, nil
	}

	for k := range task.CommonParams {
		switch k {
		case cloudprovider.FailedNodeIDsKey.String():
			failedNodeIds := cloudprovider.ParseNodeIpOrIdFromCommonMap(
				task.CommonParams, cloudprovider.FailedNodeIDsKey.String(), ",")
			allFailedNodeIds = append(allFailedNodeIds, failedNodeIds...)
		case cloudprovider.FailedClusterNodeIDsKey.String():
			addFailedNodeIds := cloudprovider.ParseNodeIpOrIdFromCommonMap(
				task.CommonParams, cloudprovider.FailedClusterNodeIDsKey.String(), ",")
			allFailedNodeIds = append(allFailedNodeIds, addFailedNodeIds...)
		case cloudprovider.FailedTransVpcNodeIDsKey.String():
			transVpcFailedNodeIds := cloudprovider.ParseNodeIpOrIdFromCommonMap(
				task.CommonParams, cloudprovider.FailedTransVpcNodeIDsKey.String(), ",")
			allFailedNodeIds = append(allFailedNodeIds, transVpcFailedNodeIds...)
		}
	}

	return allFailedNodeIds, nil
}

// checkNodeInCluster check node id in cluster
func checkNodeInCluster(task *proto.Task, nodeIPs []string) error {
	// check if nodes are already in cluster
	nodeStatus := []string{common.StatusRunning, common.StatusInitialization,
		common.StatusDeleting}
	clusterCond := operator.NewLeafCondition(operator.Eq, operator.M{"clusterid": task.ClusterID})
	statusCond := operator.NewLeafCondition(operator.In, operator.M{"status": nodeStatus})
	cond := operator.NewBranchCondition(operator.And, clusterCond, statusCond)

	nodes, err := cloudprovider.GetStorageModel().ListNode(context.Background(), cond, &options.ListOption{})
	if err != nil {
		blog.Errorf("get Cluster %s all Nodes failed when checkNodeInCluster, %s", task.ClusterID, err.Error())
		return err
	}
	newNodeIP := make(map[string]string)
	for _, nodeIp := range nodeIPs {
		newNodeIP[nodeIp] = nodeIp
	}
	for _, node := range nodes {
		if _, ok := newNodeIP[node.InnerIP]; ok {
			return fmt.Errorf("node %s is already in Cluster", node.InnerIP)
		}
	}

	return nil
}

// taskParams add nodes task params
type taskParams struct {
	operator       string
	nodeTemplateID string
	nodeSchedule   bool
	advance        *proto.NodeAdvancedInfo
}

// getTaskParams get add nodes task params
func getAddNodesTaskParams(task *proto.Task) (*taskParams, error) { // nolint
	params := &taskParams{}

	if task == nil || len(task.Steps) == 0 {
		return params, nil
	}

	for _, step := range task.Steps {
		for k := range step.Params {
			switch k {
			case cloudprovider.OperatorKey.String():
				params.operator = task.CommonParams[cloudprovider.OperatorKey.String()]
			case cloudprovider.NodeTemplateIDKey.String():
				params.nodeTemplateID = step.Params[cloudprovider.NodeTemplateIDKey.String()]
			case cloudprovider.NodeSchedule.String():
				params.nodeSchedule, _ = strconv.ParseBool(step.Params[cloudprovider.NodeSchedule.String()])
			case cloudprovider.NodeAdvanceKey.String():
				advanceBytes := step.Params[cloudprovider.NodeAdvanceKey.String()]
				_ = json.Unmarshal([]byte(advanceBytes), &params.advance)
			}
		}
	}

	return params, nil
}

// updateNodeGroupStatus update node group status
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

// updateClusterStatus update cluster status
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

// updateNodeStatus update node status
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
