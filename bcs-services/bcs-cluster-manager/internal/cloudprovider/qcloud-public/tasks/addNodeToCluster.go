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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud-public/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// AddNodesToClusterTask add node to cluster
func AddNodesToClusterTask(taskID string, stepName string) error { // nolint
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
	operator := step.Params[cloudprovider.OperatorKey.String()]
	nodeTemplateID := step.Params[cloudprovider.NodeTemplateIDKey.String()]
	scheduleStr := step.Params[cloudprovider.NodeSchedule.String()]
	loginStr := step.Params[cloudprovider.NodeLoginKey.String()]

	var login = &proto.NodeLoginInfo{}
	err = json.Unmarshal([]byte(loginStr), login)
	if err != nil {
		_ = state.UpdateStepFailure(start, stepName, err)
		return fmt.Errorf("task %s parameter err: %v", taskID, err)
	}

	// parse node schedule status
	schedule, _ := strconv.ParseBool(scheduleStr)

	// get nodes IDs and IPs
	ipList := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIPsKey.String(), ",")
	idList := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIDsKey.String(), ",")
	if len(idList) != len(ipList) {
		blog.Errorf("AddNodesToClusterTask[%s] [inner fatal] task %s step %s NodeID %d is not equal to "+
			"InnerIP %d, fatal", taskID, taskID, stepName, // nolint
			len(idList), len(ipList))
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("NodeID & InnerIP params err"))
		return fmt.Errorf("task %s parameter err", taskID)
	}
	idToIPMap := cloudprovider.GetIDToIPMap(idList, ipList)

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
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

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
			notExistedInstance, login, false, idToIPMap, operator)
		if err != nil {
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
	_ = updateNodeStatusByNodeID(failedNodes, common.StatusAddNodesFailed)

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("AddNodesToClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// CheckAddNodesStatusTask check add node status
func CheckAddNodesStatusTask(taskID string, stepName string) error {
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
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	addSuccessNodes, addFailureNodes, err := business.CheckClusterInstanceStatus(ctx, dependInfo, successNodes)
	if err != nil {
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
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(addFailureNodes, ",")
	}
	if len(addSuccessNodes) == 0 {
		blog.Errorf("CheckAddNodesStatusTask[%s] AddSuccessNodes empty", taskID)
		retErr := fmt.Errorf("上架节点超时/失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(addSuccessNodes, ",")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckAddNodesStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// UpdateNodeDBInfoTask update node DB info
func UpdateNodeDBInfoTask(taskID string, stepName string) error { // nolint
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
	addFailedNodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.FailedClusterNodeIDsKey.String(), ",")
	failedNodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.FailedNodeIDsKey.String(), ",")

	// get nodes IDs and IPs
	ipList := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIPsKey.String(), ",")
	idList := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIDsKey.String(), ",")
	if len(idList) != len(ipList) {
		blog.Errorf("UpdateNodeDBInfoTask[%s] [inner fatal] task %s step %s NodeID %d is not equal to "+
			"InnerIP %d, fatal", taskID, taskID, stepName,
			len(idList), len(ipList))
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("NodeID & InnerIP params err"))
		return fmt.Errorf("task %s parameter err", taskID)
	}
	nodeIDToIPMap := make(map[string]string, 0)
	for i := 0; i < len(ipList); i++ {
		nodeIDToIPMap[idList[i]] = ipList[i]
	}

	successInstances := addSuccessNodes
	failedInstances := make([]string, 0)
	if len(addFailedNodes) > 0 {
		failedInstances = append(failedInstances, addFailedNodes...)
	}
	if len(failedNodes) > 0 {
		failedInstances = append(failedInstances, failedNodes...)
	}

	// instanceIPs := make([]string, 0)
	// for _, instanceID := range successInstances {
	//	if ip, ok := nodeIDToIPMap[instanceID]; ok {
	//		instanceIPs = append(instanceIPs, ip)
	//	}
	// }

	// update nodes status in DB
	for i := range successInstances {
		node, errGet := cloudprovider.GetStorageModel().GetNode(context.Background(), successInstances[i])
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

	// update failed nodes status in DB
	for i := range failedInstances {
		node, errGet := cloudprovider.GetStorageModel().GetNode(context.Background(), failedInstances[i])
		if errGet != nil {
			continue
		}
		node.Passwd = passwd
		node.Status = common.StatusAddNodesFailed
		errGet = cloudprovider.GetStorageModel().UpdateNode(context.Background(), node)
		if errGet != nil {
			continue
		}
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateNodeDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
