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

package tasks

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ModifyInstancesVpcTask modify nodes vpc task
func ModifyInstancesVpcTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start modify nodes vpc")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("ModifyInstancesVpcTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("ModifyInstancesVpcTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeIds := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIDsKey.String(), ",")
	vpcId := step.Params[cloudprovider.VpcKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("ModifyInstancesVpcTask[%s] GetClusterDependBasicInfo task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	if vpcId == "" {
		vpcId = dependInfo.Cluster.GetVpcID()
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	err = business.ModifyInstancesVpcAttribute(ctx, vpcId, nodeIds, dependInfo.CmOption)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("modify nodes vpc failed [%s]", err))
		blog.Errorf("ModifyInstancesVpcTask[%s]: ModifyInstancesVpcAttribute for nodes[%v] failed, %s",
			taskID, nodeIds, err.Error())
		retErr := fmt.Errorf("ModifyInstancesVpcAttribute err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"modify nodes vpc successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("ModifyInstancesVpcTask[%s] update to storage fatal: %s", taskID, stepName)
		return err
	}
	return nil
}

// CheckInstanceStateTask check instance operation state
func CheckInstanceStateTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start check instance operation status")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckInstanceStateTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckInstanceStateTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeIds := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIDsKey.String(), ",")

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckInstanceStateTask[%s] GetClusterDependBasicInfo task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	instanceList, err := business.CheckCvmInstanceState(ctx, nodeIds,
		&cloudprovider.ListNodesOption{Common: dependInfo.CmOption})
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("check cvm instance state failed [%s]", err))
		blog.Errorf("CheckCvmInstanceState[%s]: CheckCvmInstanceState for nodes[%v] failed, %s",
			taskID, nodeIds, err.Error())
		retErr := fmt.Errorf("CheckCvmInstanceState err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	if len(instanceList.SuccessNodes) == 0 {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			"success nodes empty")
		blog.Errorf("CheckCvmInstanceState[%s] failed[%+v], successNodes empty",
			taskID, nodeIds)
		retErr := fmt.Errorf("CheckCvmInstanceState failed: successNodes empty")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	successIds, failedIds := handleAddNodesData(ctx, clusterID, instanceList)

	blog.Infof("CheckCvmInstanceState[%s] nodeIds[%v] success[%v] failed[%v]",
		taskID, nodeIds, successIds, failedIds)

	// handle task nodes
	handleTaskData(state, failedIds)
	if len(failedIds) > 0 {
		state.PartFailure = true
		state.Message = fmt.Sprintf("node[%s] trans vpc failed", strings.Join(failedIds, ","))
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"check instance operation status successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCvmInstanceState[%s] update to storage fatal: %s", taskID, stepName)
		return err
	}
	return nil
}

func handleTaskData(state *cloudprovider.TaskState, failedIds []string) {
	// again inject nodeIds/nodeIps
	nodeIds := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIDsKey.String(), ",")

	var (
		availableNodeIds = make([]string, 0)
		availableNodeIps = make([]string, 0)
	)

	for i := range nodeIds {
		if utils.StringInSlice(nodeIds[i], failedIds) {
			continue
		}
		availableNodeIds = append(availableNodeIds, nodeIds[i])
	}

	nodes := cloudprovider.GetNodesByInstanceIDs(availableNodeIds)
	for i := range nodes {
		availableNodeIps = append(availableNodeIps, nodes[i].GetInnerIP())
	}

	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	state.Task.NodeIPList = availableNodeIps
	state.Task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(availableNodeIds, ",")
	state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(availableNodeIps, ",")
	state.Task.CommonParams[cloudprovider.FailedTransVpcNodeIDsKey.String()] = strings.Join(failedIds, ",")
}

func handleAddNodesData(ctx context.Context, clusterId string, nodes *business.InstanceList) ([]string, []string) {
	var (
		failedNodeIds  = make([]string, 0)
		successNodeIds = make([]string, 0)
	)

	taskId := cloudprovider.GetTaskIDFromContext(ctx)

	// update success nodes ip
	for i := range nodes.SuccessNodes {
		successNodeIds = append(successNodeIds, nodes.SuccessNodes[i].NodeId)
		err := updateNodeIPByNodeID(ctx, clusterId, nodes.SuccessNodes[i])
		if err != nil {
			blog.Errorf("handleAddNodesData[%s] updateNodeIPByNodeID[%s][%s] failed: %v",
				taskId, nodes.SuccessNodes[i].NodeId, nodes.SuccessNodes[i].NodeIp, err)
			continue
		}

		blog.Infof("handleAddNodesData[%s] updateNodeIPByNodeID[%s][%s] successful",
			taskId, nodes.SuccessNodes[i].NodeId, nodes.SuccessNodes[i].NodeIp)
	}

	// update failed nodes status
	for i := range nodes.FailedNodes {
		failedNodeIds = append(failedNodeIds, nodes.FailedNodes[i].NodeId)
	}
	reason := "node trans vpc failed"
	_ = updateNodeStatusByNodeID(failedNodeIds, common.StatusAddNodesFailed, reason)

	return successNodeIds, failedNodeIds
}

// AddNodesShieldAlarmTask shield nodes alarm
func AddNodesShieldAlarmTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start shield nodes alarm")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("AddNodesShieldAlarmTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("AddNodesShieldAlarmTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]

	ipList := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.GetCommonParams(),
		cloudprovider.NodeIPsKey.String(), ",")
	if len(ipList) == 0 {
		blog.Errorf("AddNodesShieldAlarmTask[%s]: get cluster IPList/clusterID empty", taskID)
		retErr := fmt.Errorf("AddNodesShieldAlarmTask: get cluster IPList/clusterID empty")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("AddNodesShieldAlarmTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)
	err = cloudprovider.ShieldHostAlarm(ctx, cluster.BusinessID, ipList)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("shield nodes alarm failed [%s]", err))
		blog.Errorf("AddNodesShieldAlarmTask[%s] ShieldHostAlarmConfig failed: %v", taskID, err)
	} else {
		blog.Infof("AddNodesShieldAlarmTask[%s] ShieldHostAlarmConfig success", taskID)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"shield nodes alarm successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("AddNodesShieldAlarmTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// AddNodesToClusterTask add node to cluster
func AddNodesToClusterTask(taskID string, stepName string) error { // nolint
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start add nodes to cluster")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("AddNodesToClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("AddNodesToClusterTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	initPasswd := step.Params[cloudprovider.PasswordKey.String()]
	if len(initPasswd) == 0 {
		initPasswd = utils.BuildInstancePwd()
	}
	operator := step.Params[cloudprovider.OperatorKey.String()]
	nodeTemplateID := step.Params[cloudprovider.NodeTemplateIDKey.String()]
	scheduleStr := step.Params[cloudprovider.NodeSchedule.String()]

	// parse node schedule status
	schedule, _ := strconv.ParseBool(scheduleStr)

	// get nodes IDs and IPs
	ipList := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.GetCommonParams(),
		cloudprovider.NodeIPsKey.String(), ",")
	idList := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.GetCommonParams(),
		cloudprovider.NodeIDsKey.String(), ",")
	if len(idList) != len(ipList) {
		blog.Errorf("AddNodesToClusterTask[%s] [inner fatal] task %s step %s NodeID %d is not equal to "+
			"InnerIP %d, fatal", taskID, taskID, stepName, // nolint
			len(idList), len(ipList))
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("NodeID & InnerIP params err"))
		return fmt.Errorf("task %s parameter err", taskID)
	}
	idToIPMap := cloudprovider.GetNodeIdToIpMapByNodeIds(idList)

	// cluster/cloud/nodeGroup/cloudCredential Info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:      clusterID,
		CloudID:        cloudID,
		NodeTemplateID: nodeTemplateID,
	})
	if err != nil {
		blog.Errorf("AddNodesToClusterTask[%s] GetClusterDependBasicInfo task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	existedInstance, notExistedInstance, err := business.FilterClusterInstanceFromNodesIDs(ctx, dependInfo, idList)
	if err != nil {
		blog.Errorf("AddNodesToClusterTask[%s]: FilterClusterInstanceFromNodesIDs for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("FilterClusterInstanceFromNodesIDs err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("AddNodesToClusterTask[%s] existedInstance[%v] notExistedInstance[%v]",
		taskID, existedInstance, notExistedInstance)

	var (
		successNodes, failedNodes []string
	)
	successNodes = append(successNodes, existedInstance...)

	if len(notExistedInstance) > 0 {
		result, err := business.AddNodesToCluster(ctx, dependInfo, &business.NodeAdvancedOptions{NodeScheduler: schedule}, // nolint
			notExistedInstance, initPasswd, false, idToIPMap, operator)
		if err != nil {
			cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
				fmt.Sprintf("add nodes to cluster failed [%s]", err))
			blog.Errorf("AddNodesToClusterTask[%s] AddNodesToCluster failed: %v", taskID, err)
			retErr := fmt.Errorf("AddNodesToCluster err, %s", err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		successNodes = append(successNodes, result.SuccessNodes...)
		failedNodes = append(failedNodes, result.FailedNodes...)
	}
	blog.Infof("AddNodesToClusterTask[%s] cluster[%s] success[%v] failed[%v]",
		taskID, clusterID, successNodes, failedNodes)
	if len(successNodes) == 0 {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			"success nodes empty")
		blog.Errorf("AddNodesToClusterTask[%s] AddNodesToCluster failed: %v", taskID, err)
		retErr := fmt.Errorf("AddNodesToCluster err, %s", "successNodes empty")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if len(idList) == len(successNodes) {
		blog.Infof("AddNodesToClusterTask[%s] cluster[%s] successful", taskID, clusterID)
	}

	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()] = strings.Join(successNodes, ",")
	state.Task.CommonParams[cloudprovider.FailedNodeIDsKey.String()] = strings.Join(failedNodes, ",")
	// set failed node status
	if len(failedNodes) > 0 {
		reason := "call tke addNode failed"
		_ = updateNodeStatusByNodeID(failedNodes, common.StatusAddNodesFailed, reason)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"add nodes to cluster successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("AddNodesToClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// CheckAddNodesStatusTask check add node status
func CheckAddNodesStatusTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start check add nodes status")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckAddNodesStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckAddNodesStatusTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	// get previous step paras
	successNodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessNodeIDsKey.String(), ",")

	// handler logic
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckAddNodesStatusTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	addSuccessNodes, addFailureNodes, err := business.CheckClusterInstanceStatus(ctx, dependInfo, successNodes)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("check cluster instance status failed [%s]", err))
		blog.Errorf("CheckAddNodesStatusTask[%s] CheckClusterInstanceStatus failed, %s",
			taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterInstanceStatus failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CheckAddNodesStatusTask[%s] addSuccessNodes[%v] addFailureNodes[%v]",
		taskID, addSuccessNodes, addFailureNodes)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(addFailureNodes) > 0 {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			"add success nodes empty")
		insInfos, reason, _ := business.GetFailedNodesReason(ctx, dependInfo, addFailureNodes)
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(addFailureNodes, ",")
		state.Task.CommonParams[cloudprovider.FailedClusterNodeReasonKey.String()] = reason

		state.PartFailure = true
		state.Message = reason
		blog.Errorf("CheckAddNodesStatusTask[%s] failedNodes[%+v] reason[%s]", taskID, insInfos, reason)
		_ = updateFailedNodeStatusByNodeID(ctx, insInfos, common.StatusAddNodesFailed)
	}

	if len(addSuccessNodes) == 0 {
		blog.Errorf("CheckAddNodesStatusTask[%s] AddSuccessNodes empty", taskID)
		retErr := fmt.Errorf("上架节点超时/失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	nodeIPs := cloudprovider.GetInstanceIPsByID(ctx, addSuccessNodes)
	state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(addSuccessNodes, ",")
	state.Task.NodeIPList = nodeIPs
	state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(nodeIPs, ",")
	blog.Infof("CheckAddNodesStatusTask[%s] successNodeIds[%v] successNodeIps[%v]",
		taskID, addSuccessNodes, nodeIPs)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"check added nodes status successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckAddNodesStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// UpdateNodeDBInfoTask update node DB info
func UpdateNodeDBInfoTask(taskID string, stepName string) error { // nolint
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start update node db info")
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateNodeDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateNodeDBInfoTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	passwd := state.Task.CommonParams[cloudprovider.PasswordKey.String()]
	addSuccessNodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessClusterNodeIDsKey.String(), ",")

	// update nodes status in DB
	for i := range addSuccessNodes {
		node, errGet := cloudprovider.GetStorageModel().GetNode(context.Background(), addSuccessNodes[i])
		if errGet != nil {
			continue
		}
		node.Passwd = passwd
		node.Status = common.StatusInitialization
		errGet = cloudprovider.GetStorageModel().UpdateNode(context.Background(), node)
		if errGet != nil {
			continue
		}
	}
	blog.Infof("UpdateNodeDBInfoTask[%s] step %s successful", taskID, stepName)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"update node db info successful")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateNodeDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
