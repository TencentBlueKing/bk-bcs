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
	tcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ApplyCVMFromResourcePoolTask apply instance from resource module
func ApplyCVMFromResourcePoolTask(taskID, stepName string) error { // nolint
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("ApplyCVMFromResourcePoolTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("ApplyCVMFromResourcePoolTask[%s] run current step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract valid parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	desiredNodes := step.Params[cloudprovider.ScalingNodesNumKey.String()]
	operator := step.Params[cloudprovider.OperatorKey.String()]
	passwd := state.Task.CommonParams[cloudprovider.PasswordKey.String()]
	manual := state.Task.CommonParams[cloudprovider.ManualKey.String()]
	scalingNum, _ := strconv.Atoi(desiredNodes)

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("ApplyCVMFromResourcePoolTask[%s] GetClusterDependBasicInfo for NodeGroup %s to clean Node in task %s "+
			"step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		blog.Infof("ApplyCVMFromResourceManagerTask[%s] begin DeleteVirtualNodes", taskID)
		_ = cloudprovider.DeleteVirtualNodes(clusterID, nodeGroupID, taskID)
		_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, scalingNum, true)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	// get task old orderId
	oldOrderId := state.Task.CommonParams[cloudprovider.DeviceRecordIDKey.String()]

	recordInstanceList, orderID, err := applyInstanceFromResourcePool(ctx, dependInfo, state,
		oldOrderId, scalingNum, operator)
	if err != nil {
		blog.Errorf("ApplyCVMFromResourcePoolTask[%s] applyInstanceFromResourcePool for group %s orderID %s failed, %s",
			taskID, nodeGroupID, orderID, err.Error())
		state.Task.CommonParams[cloudprovider.DeviceRecordIDKey.String()] = orderID
		retErr := fmt.Errorf("ApplyCVMFromResourcePoolTask failed: %s", err.Error())
		if manual == common.True {
			_ = cloudprovider.UpdateVirtualNodeStatus(clusterID, nodeGroupID, taskID)
		} else {
			_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, scalingNum, true)
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("ApplyCVMFromResourcePoolTask[%s] applyInstanceFromResourcePool desiredNum[%v] applyNum[%v]",
		taskID, scalingNum, len(recordInstanceList.InstanceIDList))

	if len(recordInstanceList.InstanceIPList) > 0 && len(recordInstanceList.DeviceIDList) > 0 &&
		len(recordInstanceList.InstanceIDList) > 0 {
		// Job Result parameter
		state.Task.NodeIPList = recordInstanceList.InstanceIPList
		state.Task.CommonParams[cloudprovider.DeviceRecordIDKey.String()] = orderID
		state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(recordInstanceList.InstanceIPList,
			",")
		state.Task.CommonParams[cloudprovider.OriginNodeIPsKey.String()] = strings.Join(recordInstanceList.InstanceIPList, // nolint
			",")
		state.Task.CommonParams[cloudprovider.NodeIDsKey.String()] = strings.Join(recordInstanceList.InstanceIDList,
			",")
		state.Task.CommonParams[cloudprovider.DeviceIDsKey.String()] = strings.Join(recordInstanceList.DeviceIDList,
			",")

		blog.Infof("ApplyCVMFromResourcePoolTask[%s] instanceIP[%+v], instanceID[%+v] deviceID[%+v]", taskID,
			recordInstanceList.InstanceIPList, recordInstanceList.InstanceIDList, recordInstanceList.DeviceIDList)
	}

	balance := scalingNum - len(recordInstanceList.InstanceIDList)
	if balance > 0 {
		blog.Infof("ApplyCVMFromResourcePoolTask[%s] applyInstanceFromResourcePool updateDesiredSize[%v:%v:%v]",
			taskID, scalingNum, len(recordInstanceList.InstanceIDList), balance)
		err = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, balance, true)
		if err != nil {
			blog.Errorf("ApplyCVMFromResourcePoolTask[%s] applyInstanceFromResourcePool "+
				"update balanced failed: %v", taskID, err)
		}
	}

	// save nodes to clusterManager db
	err = recordClusterCVMInfoToDB(ctx, dependInfo, &RecordInstanceToDBOption{
		Password:    passwd,
		InstanceIDs: recordInstanceList.InstanceIDList,
		DeviceIDs:   recordInstanceList.DeviceIDList,
	})
	if err != nil {
		blog.Errorf("ApplyCVMFromResourcePoolTask[%s] applyInstanceFromResourcePool for NodeGroup %s "+
			"step %s failed, %s",
			taskID, nodeGroupID, stepName, err.Error())
		retErr := fmt.Errorf("applyInstanceFromResourcePool[%s] failed, %s", orderID, err.Error())
		if manual == common.True {
			_ = cloudprovider.UpdateVirtualNodeStatus(clusterID, nodeGroupID, taskID)
		} else {
			destroyDeviceList(ctx, dependInfo, recordInstanceList.DeviceIDList, operator) // nolint
			_ = cloudprovider.UpdateNodeGroupDesiredSize(nodeGroupID, scalingNum, true)
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// destroy virtual nodes
	if manual == common.True {
		blog.Infof("ApplyCVMFromResourceManagerTask[%s] begin DeleteVirtualNodes", taskID)
		_ = cloudprovider.DeleteVirtualNodes(clusterID, nodeGroupID, taskID)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("ApplyCVMFromResourceManagerTask[%s] task %s %s update to storage fatal", taskID, taskID,
			stepName)
		return err
	}

	return nil
}

// AddNodesToClusterTask add instance to cluster
func AddNodesToClusterTask(taskID, stepName string) error { // nolint
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

	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	desiredNodes := step.Params[cloudprovider.ScalingNodesNumKey.String()]
	nodeNum, _ := strconv.Atoi(desiredNodes)
	operator := step.Params[cloudprovider.OperatorKey.String()]
	manual := state.Task.CommonParams[cloudprovider.ManualKey.String()]

	passwd := state.Task.CommonParams[cloudprovider.PasswordKey.String()]
	if passwd == "" {
		passwd = utils.BuildInstancePwd()
	}

	// parse instanceIP/instanceID
	instanceIPs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIPsKey.String(), ",")
	instanceIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIDsKey.String(), ",")

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// cluster basic depend info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("AddNodesToClusterTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("AddNodesToClusterTask GetClusterDependBasicInfo failed")
		_ = returnDevicesToRMAndCleanNodes(ctx, dependInfo, instanceIDs, false, operator)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	if len(instanceIDs) == 0 || len(instanceIDs) != len(instanceIPs) {
		blog.Errorf("AddNodesToClusterTask[%s] strconv desiredNodes failed: %v", taskID, err)
		retErr := fmt.Errorf("strconv desiredNodes failed: %v", err.Error())
		if manual != common.True {
			_ = returnDevicesToRMAndCleanNodes(ctx, dependInfo, instanceIDs, false, operator)
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}
	idToIPMap := cloudprovider.GetIDToIPMap(instanceIDs, instanceIPs)

	// shield nodes alarm
	cloudprovider.ShieldHostAlarm(ctx, dependInfo.Cluster.BusinessID, instanceIPs) // nolint

	// tke cluster add nodes to cluster
	result, err := business.AddNodesToCluster(ctx, dependInfo, nil, instanceIDs,
		passwd, true, idToIPMap, operator)
	if err != nil {
		// rollback nodes when addNodes failed
		if manual != common.True {
			errMsg := returnDevicesToRMAndCleanNodes(ctx, dependInfo, instanceIDs, true, operator)
			if errMsg != nil {
				blog.Errorf("AddNodesToClusterTask[%s] returnDevicesToRMAndCleanNodes failed: %v", taskID,
					errMsg)
			}
		}
		blog.Errorf("AddNodesToClusterTask[%s] nodegroupId[%s] failed: %v", taskID, nodeGroupID, err)
		retErr := fmt.Errorf("AddNodesToClusterTask err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// rollback failed nodes
	if len(result.FailedNodes) > 0 {
		// return devices and clean DB
		blog.Errorf("AddNodesToClusterTask[%s] handle failedNodes[%v]", taskID, result.FailedNodes)
		errMsg := returnDevicesToRMAndCleanNodes(ctx, dependInfo, result.FailedNodes, true, operator)
		if errMsg != nil {
			blog.Errorf("AddNodesToClusterTask[%s] handle failedNodes[%v] %v", taskID, result.FailedNodes,
				errMsg)
		}
	}
	if len(result.SuccessNodes) == 0 {
		blog.Errorf("AddNodesToClusterTask[%s] AddNodesToCluster failed: succeedNodes empty", taskID)
		if manual != common.True {
			errMsg := returnDevicesToRMAndCleanNodes(ctx, dependInfo, instanceIDs, true, operator)
			if errMsg != nil {
				blog.Errorf("AddNodesToClusterTask[%s] returnDevicesToRMAndCleanNodes failed: %v", taskID,
					errMsg)
			}
		}
		retErr := fmt.Errorf("AddNodesToCluster err, %s", "successNodes empty")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if nodeNum == len(result.SuccessNodes) {
		blog.Infof("AddNodesToClusterTask[%s] all instanceIDs successful", taskID)
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	state.Task.CommonParams[cloudprovider.SuccessAddClusterNodeIDsKey.String()] =
		strings.Join(result.SuccessNodes, ",")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("AddNodesToClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// CheckClusterNodeStatusTask check cluster nodes status
func CheckClusterNodeStatusTask(taskID, stepName string) error { // nolint
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckClusterNodeStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckClusterNodeStatusTask[%s] task %s run current step %s, "+
		"system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	operator := step.Params[cloudprovider.OperatorKey.String()]
	manual := state.Task.CommonParams[cloudprovider.ManualKey.String()]

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	// cluster basic depend info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CheckClusterNodeStatusTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterNodeStatusTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// parse add nodes instanceID
	successNodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessAddClusterNodeIDsKey.String(), ",")
	if len(successNodes) == 0 {
		blog.Errorf("CheckClusterNodeStatusTask[%s] failed: succeedNodes empty", taskID)
		retErr := fmt.Errorf("CheckClusterNodeStatusTask succeedNodes empty")
		if manual != common.True {
			_ = returnDevicesToRMAndCleanNodes(ctx, dependInfo, successNodes, true, operator)
		}
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// check cluster nodes status
	success, failed, err := business.CheckClusterInstanceStatus(ctx, dependInfo, successNodes)
	if err != nil || len(success) == 0 {
		// rollback failed nodes
		if manual != common.True {
			errMsg := returnDevicesToRMAndCleanNodes(ctx, dependInfo, successNodes, true, operator)
			if errMsg != nil {
				blog.Errorf("CheckClusterNodeStatusTask[%s] nodes[%+v] failed: %v",
					taskID, successNodes, errMsg)
			}
		}
		blog.Errorf("CheckClusterNodeStatusTask[%s] nodegroupId[%s] failed: %v", taskID, nodeGroupID, err)
		retErr := fmt.Errorf("CheckClusterNodeStatusTask err: %v", "add nodes timeOut")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	// rollback abnormal nodes
	if len(failed) > 0 {
		blog.Errorf("CheckClusterNodeStatusTask[%s] handle failedNodes[%v]", taskID, failed)
		errMsg := returnDevicesToRMAndCleanNodes(ctx, dependInfo, failed, true, operator)
		if errMsg != nil {
			blog.Errorf("CheckClusterNodeStatusTask[%s] returnDevicesToRMAndCleanNodes failed %v", taskID,
				errMsg)
		}
	}
	blog.Infof("CheckClusterNodeStatusTask[%s] delivery succeed[%d] instances[%v] failed[%d] instances[%v]",
		taskID, len(success), success, len(failed), failed)

	// trans instanceIDs to ipList
	ipList := cloudprovider.GetInstanceIPsByID(ctx, success)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(success) > 0 {
		state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(success, ",")
	}
	if len(failed) > 0 {
		_, reason, _ := business.GetFailedNodesReason(ctx, dependInfo, failed)
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(failed, ",")
		state.Task.CommonParams[cloudprovider.FailedClusterNodeReasonKey.String()] = reason
	}

	// success ip list
	if len(ipList) > 0 {
		state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(ipList, ",")
		state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(ipList, ",")
		state.Task.NodeIPList = ipList
	}

	blog.Infof("CheckClusterNodeStatusTask[%s] instanceIP[%+v], instanceID[%+v]", taskID,
		ipList, successNodes)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckClusterNodeStatusTask[%s] task %s %s update to storage fatal", taskID, taskID,
			stepName)
		return err
	}

	return nil
}

// CheckClusterNodesInCMDBTask check nodes exist in cmdb task
func CheckClusterNodesInCMDBTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	operator := step.Params[cloudprovider.OperatorKey.String()]
	manual := state.Task.CommonParams[cloudprovider.ManualKey.String()]

	// cluster basic depend info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckClusterNodesInCMDBTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterNodesInCMDBTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get nodeIPs
	nodeIPs := state.Task.CommonParams[cloudprovider.NodeIPsKey.String()]
	ips := strings.Split(nodeIPs, ",")
	nodeIds := state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()]
	ids := strings.Split(nodeIds, ",")

	if len(ips) == 0 {
		blog.Infof("CheckNodeIpsInCMDBTask[%s] nodeIPs empty", taskID)
		return nil
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	err = tcommon.CheckIPsInCmdb(ctx, ips)
	if err != nil {
		blog.Errorf("CheckNodeIpsInCMDBTask[%s] failed: %v", taskID, err)
		if manual != common.True {
			_ = returnDevicesToRMAndCleanNodes(ctx, dependInfo, ids, true, operator)
		}
		_ = state.UpdateStepFailure(start, stepName, err)
		return err
	}
	blog.Infof("CheckNodeIpsInCMDBTask %s successful", taskID)

	// update step
	_ = state.UpdateStepSucc(start, stepName)
	return nil
}
