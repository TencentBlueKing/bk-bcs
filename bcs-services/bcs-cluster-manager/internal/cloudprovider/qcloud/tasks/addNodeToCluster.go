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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
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
	vpcID := step.Params[cloudprovider.VpcKey.String()]

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
	if vpcID == "" {
		vpcID = dependInfo.Cluster.GetVpcID()
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDAndStepNameForContext(context.Background(), taskID, stepName)

	err = business.ModifyInstancesVpcAttribute(ctx, vpcID, nodeIds, dependInfo.CmOption)
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
func CheckInstanceStateTask(taskID string, stepName string) error { // nolint
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
	successIds, failedIds := handleClusterWorkerNodesData(ctx, clusterID, instanceList)

	blog.Infof("CheckCvmInstanceState[%s] nodeIds[%v] success[%v] failed[%v]",
		taskID, nodeIds, successIds, failedIds)

	// handle task nodes
	handleTaskData(state, failedIds)

	// step partFailure
	partFailure := false
	if len(failedIds) > 0 {
		partFailure = true
		state.PartFailure = partFailure
		state.Message = fmt.Sprintf("node[%s] trans vpc failed", strings.Join(failedIds, ","))
	}

	// update step
	if partFailure {
		cloudprovider.GetStorageModel().CreateTaskStepLogWarn(context.Background(), taskID, stepName,
			"check instance operation status part failure")

		retErr := fmt.Errorf("CheckInstanceStateTask partfailure failedNodes: [%s]",
			strings.Join(failedIds, ","))
		if err := state.UpdateStepPartFailure(start, stepName, retErr); err != nil {
			blog.Errorf("CheckCvmInstanceState[%s] update to storage fatal: %s", taskID, stepName)
			return err
		}

		return nil
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

// handleTaskData handle task data
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
	state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(availableNodeIps, ",")
	state.Task.CommonParams[cloudprovider.FailedTransVpcNodeIDsKey.String()] = strings.Join(failedIds, ",")
}

// handleClusterMasterNodesData handle master nodes data
func handleClusterMasterNodesData(ctx context.Context, clusterID string, nodes *business.InstanceList) error {
	taskId := cloudprovider.GetTaskIDFromContext(ctx)

	cls, err := cloudprovider.GetStorageModel().GetCluster(ctx, clusterID)
	if err != nil {
		blog.Errorf("handleClusterMasterNodesData[%s] get cluster[%s] failed: %v", taskId, clusterID, err)
		return err
	}

	// if cluster is managed, do not need to update master nodes
	if len(cls.Master) == 0 || cls.ManageType == common.ClusterManageTypeManaged {
		return nil
	}

	var (
		masterIDs      = make([]string, 0)
		masterIDToNode = make(map[string]*proto.Node)
	)
	for i := range cls.GetMaster() {
		masterIDs = append(masterIDs, cls.GetMaster()[i].NodeID)
		masterIDToNode[cls.GetMaster()[i].NodeID] = cls.GetMaster()[i]
	}

	nodeIDToNode := make(map[string]business.InstanceInfo)
	for _, n := range nodes.SuccessNodes {
		nodeIDToNode[n.NodeId] = n
	}

	// update master nodes ip
	masterNodes := make(map[string]*proto.Node)
	for _, id := range masterIDs {
		dbNode := masterIDToNode[id]

		ins, ok := nodeIDToNode[id]
		if ok {
			dbNode.InnerIP = ins.NodeIp
			dbNode.VPC = ins.VpcId
		}
		masterNodes[dbNode.InnerIP] = dbNode
	}
	cls.Master = masterNodes

	return cloudprovider.GetStorageModel().UpdateCluster(ctx, cls)
}

// handleClusterWorkerNodesData handle nodes data
func handleClusterWorkerNodesData(ctx context.Context, clusterID string,
	nodes *business.InstanceList) ([]string, []string) {
	var (
		failedNodeIds  = make([]string, 0)
		successNodeIds = make([]string, 0)
	)

	// get taskID
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// update success nodes ip
	for i := range nodes.SuccessNodes {
		err := updateNodeIPByNodeID(ctx, clusterID, nodes.SuccessNodes[i])
		if err != nil {
			blog.Errorf("handleAddNodesData[%s] updateNodeIPByNodeID[%s][%s] failed: %v",
				taskID, nodes.SuccessNodes[i].NodeId, nodes.SuccessNodes[i].NodeIp, err)
			continue
		}
		successNodeIds = append(successNodeIds, nodes.SuccessNodes[i].NodeId)

		blog.Infof("handleAddNodesData[%s] updateNodeIPByNodeID[%s][%s] successful",
			taskID, nodes.SuccessNodes[i].NodeId, nodes.SuccessNodes[i].NodeIp)
	}

	// update failed nodes status
	for i := range nodes.FailedNodes {
		failedNodeIds = append(failedNodeIds, nodes.FailedNodes[i].NodeId)
	}
	reason := "node trans vpc failed"
	_ = updateNodeStatusByNodeID(failedNodeIds, common.StatusAddNodesFailed, reason)

	return successNodeIds, failedNodeIds
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
	// get node advance info
	advancedInfo := &proto.NodeAdvancedInfo{}
	advance, exist := step.Params[cloudprovider.NodeAdvanceKey.String()]
	if exist {
		_ = json.Unmarshal([]byte(advance), advancedInfo)
	}

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

	blog.Infof("AddNodesToClusterTask[%s] GetNodeIdToIpMapByNodeIds %v", taskID, idToIPMap)

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

	// handle instance list
	existedInstance, notExistedInstance, err := business.FilterClusterInstanceFromNodesIDs(ctx, dependInfo, idList)
	if err != nil {
		blog.Errorf("AddNodesToClusterTask[%s]: FilterClusterInstanceFromNodesIDs for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("FilterClusterInstanceFromNodesIDs err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	existedInstanceIps := make([]string, 0)
	notExistedInstanceIps := make([]string, 0)
	for _, ins := range existedInstance {
		existedInstanceIps = append(existedInstanceIps, idToIPMap[ins])
	}
	for _, ins := range notExistedInstance {
		notExistedInstanceIps = append(notExistedInstanceIps, idToIPMap[ins])
	}
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		fmt.Sprintf("AddNodesToClusterTask existedInstance id[%v] ip[%v], notExistedInstance id[%v] ip[%v]",
			existedInstance, existedInstanceIps, notExistedInstance, notExistedInstanceIps))
	blog.Infof("AddNodesToClusterTask[%s] existedInstance id[%v] ip[%v], notExistedInstance id[%v] ip[%v]",
		taskID, existedInstance, existedInstanceIps, notExistedInstance, notExistedInstanceIps)

	// record success and failed node ids
	var (
		successNodeIds, failedNodeIds []string
	)
	successNodeIds = append(successNodeIds, existedInstance...)

	// notExistedInstance handle
	if len(notExistedInstance) > 0 {
		// if node template exists, set user script for new node
		result, err := business.AddNodesToCluster(ctx, dependInfo, &business.NodeAdvancedOptions{ // nolint
			NodeScheduler:         schedule,
			SetPreStartUserScript: true,
			Advance:               advancedInfo,
		}, notExistedInstance, initPasswd, false, idToIPMap, operator)
		if err != nil {
			cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
				fmt.Sprintf("add nodes to cluster failed [%s]", err))
			blog.Errorf("AddNodesToClusterTask[%s] AddNodesToCluster failed: %v", taskID, err)
			retErr := fmt.Errorf("AddNodesToCluster err, %s", err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		// record success and failed nodes
		for i := range result.SuccessNodeInfos {
			successNodeIds = append(successNodeIds, result.SuccessNodeInfos[i].NodeId)
		}
		for i := range result.FailedNodeInfos {
			failedNodeIds = append(failedNodeIds, result.FailedNodeInfos[i].NodeId)
		}
	}

	blog.Infof("AddNodesToClusterTask[%s] cluster[%s] success [%v] failed[%v]",
		taskID, clusterID, successNodeIds, failedNodeIds)
	if len(successNodeIds) == 0 {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			"success nodes empty")
		blog.Errorf("AddNodesToClusterTask[%s] AddNodesToCluster failed: %v", taskID, err)
		retErr := fmt.Errorf("AddNodesToCluster err, %s", "successNodes empty")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// success node log
	if len(idList) == len(successNodeIds) {
		blog.Infof("AddNodesToClusterTask[%s] cluster[%s] successful", taskID, clusterID)
	}

	// handle task data
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	state.Task.CommonParams[cloudprovider.SuccessNodeIDsKey.String()] = strings.Join(successNodeIds, ",")
	state.Task.CommonParams[cloudprovider.FailedNodeIDsKey.String()] = strings.Join(failedNodeIds, ",")

	// set failed node status
	if len(failedNodeIds) > 0 {
		reason := "call tke addNode failed"
		_ = updateNodeStatusByNodeID(failedNodeIds, common.StatusAddNodesFailed, reason)
		cloudprovider.GetStorageModel().CreateTaskStepLogWarn(context.Background(), taskID, stepName,
			fmt.Sprintf("add nodes to cluster part failed, failedNodes: [%s]", strings.Join(failedNodeIds, ",")))
		retErr := fmt.Errorf("AddNodesToClusterTask partfailure failedNodes: [%v]", strings.Join(failedNodeIds, ","))
		if err := state.UpdateStepPartFailure(start, stepName, retErr); err != nil {
			blog.Errorf("AddNodesToClusterTask[%s %s] update to storage fatal: %s", taskID, stepName, err)
			return err
		}
		return nil
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
func CheckAddNodesStatusTask(taskID string, stepName string) error { // nolint
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

	// check add node status
	addSuccessNodeIds, addFailureNodeIds, err := business.CheckClusterInstanceStatus(ctx, dependInfo, successNodes)
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
		taskID, addSuccessNodeIds, addFailureNodeIds)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	// if successNodes empty
	if len(addSuccessNodeIds) == 0 {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			"add success nodes empty")
		blog.Errorf("CheckAddNodesStatusTask[%s] AddSuccessNodes empty", taskID)
		retErr := fmt.Errorf("上架节点超时/失败, 请联系管理员")
		// update step
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// step partFailure
	partFailure := false
	if len(addFailureNodeIds) > 0 {
		insInfos, reason, _ := business.GetFailedNodesReason(ctx, dependInfo, addFailureNodeIds)
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(addFailureNodeIds, ",")
		state.Task.CommonParams[cloudprovider.FailedClusterNodeReasonKey.String()] = reason

		partFailure = true
		state.PartFailure = partFailure
		state.Message = reason
		blog.Errorf("CheckAddNodesStatusTask[%s] failedNodes[%+v] reason[%s]", taskID, insInfos, reason)
		_ = updateFailedNodeStatusByNodeID(ctx, insInfos, common.StatusAddNodesFailed)
	}

	// update info to task common params
	addSuccessNodeIps := cloudprovider.GetInstanceIPsByID(ctx, addSuccessNodeIds)
	addFailureNodeIps := cloudprovider.GetInstanceIPsByID(ctx, addFailureNodeIds)
	state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(addSuccessNodeIds, ",")
	state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(addSuccessNodeIps, ",")
	state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(addSuccessNodeIps, ",")

	blog.Infof("CheckAddNodesStatusTask[%s] successNodeIds[%v] successNodeIps[%v] failureNodeIds[%v] "+
		"failureNodeIps[%v]", taskID, addSuccessNodeIds, addSuccessNodeIps, addFailureNodeIds, addFailureNodeIps)

	// update step
	if partFailure {
		cloudprovider.GetStorageModel().CreateTaskStepLogWarn(context.Background(), taskID, stepName,
			fmt.Sprintf("add nodes to cluster part failed, failedNodesIds: [%s], failedNodesIps: [%s]",
				strings.Join(addFailureNodeIds, ","), strings.Join(addFailureNodeIps, ",")))
		retErr := fmt.Errorf("CheckAddNodesStatusTask partfailure:clusterFailedNodes: [%s]",
			strings.Join(addFailureNodeIds, ","))
		if err := state.UpdateStepPartFailure(start, stepName, retErr); err != nil {
			blog.Errorf("CheckAddNodesStatusTask[%s] update to storage fatal: %s", taskID, stepName)
			return err
		}
		return nil
	}

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
