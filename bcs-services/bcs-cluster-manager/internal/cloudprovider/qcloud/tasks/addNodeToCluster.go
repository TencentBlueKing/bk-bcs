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

package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// AddNodesShieldAlarmTask shield nodes alarm
func AddNodesShieldAlarmTask(taskID string, stepName string) error {
	start := time.Now()

	// get task form database
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("AddNodesShieldAlarmTask[%s] task %s get detail task information from storage failed: %s, task retry", taskID, taskID, err.Error())
		return err
	}

	// task state check
	state := &cloudprovider.TaskState{
		Task:      task,
		JobResult: cloudprovider.NewJobSyncResult(task),
	}
	// check task already terminated
	if state.IsTerminated() {
		blog.Errorf("AddNodesShieldAlarmTask[%s] task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	// workflow switch current step to stepName when previous task exec successful
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("AddNodesShieldAlarmTask[%s] task %s not turn ro run step %s, err %s", taskID, taskID, stepName, err.Error())
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
	clusterID := step.Params["ClusterID"]
	_ = step.Params["NodeIPs"]
	ipList := strings.Split(step.Params["NodeIPs"], ",")
	_, err = cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("AddNodesShieldAlarmTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if len(ipList) == 0 || len(clusterID) == 0 {
		blog.Errorf("AddNodesShieldAlarmTask[%s]: get cluster IPList/clusterID empty", taskID)
		retErr := fmt.Errorf("AddNodesShieldAlarmTask: get cluster IPList/clusterID empty")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// attention: call client to shield alarm
	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("AddNodesShieldAlarmTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// AddNodesToClusterTask add node to cluster
func AddNodesToClusterTask(taskID string, stepName string) error {
	start := time.Now()

	// get task form database
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("AddNodesToClusterTask[%s] task %s get detail task information from storage failed: %s, task retry", taskID, taskID, err.Error())
		return err
	}

	// task state check
	state := &cloudprovider.TaskState{
		Task:      task,
		JobResult: cloudprovider.NewJobSyncResult(task),
	}
	// check task already terminated
	if state.IsTerminated() {
		blog.Errorf("AddNodesToClusterTask[%s] task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	// workflow switch current step to stepName when previous task exec successful
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("AddNodesToClusterTask[%s] task %s not turn ro run step %s, err %s", taskID, taskID, stepName, err.Error())
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
	clusterID := step.Params["ClusterID"]
	nodeGroupID := step.Params["NodeGroupID"] // nodeGroup may be empty
	cloudID := step.Params["CloudID"]
	initPasswd := step.Params["InitPasswd"] // initPasswd may be empty
	if len(initPasswd) == 0 {
		initPasswd = task.CommonParams["Password"]
	}

	// get nodes IDs and IPs
	ipList := strings.Split(step.Params["NodeIPs"], ",")
	idList := strings.Split(step.Params["NodeIDs"], ",")
	if len(idList) != len(ipList) {
		blog.Errorf("AddNodesToClusterTask[%s] [inner fatal] task %s step %s NodeID %d is not equal to InnerIP %d, fatal", taskID, taskID, stepName,
			len(idList), len(ipList))
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("NodeID & InnerIP params err"))
		return fmt.Errorf("task %s parameter err", taskID)
	}

	// get cloudInfo bu cloudID & get projectInfo by ProjectID
	cloud, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), cloudID, clusterID)
	if err != nil {
		blog.Errorf("AddNodesToClusterTask[%s] get cloud/project for NodeGroup %s to clean Node in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get cloud api_interface dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("AddNodesToClusterTask[%s] get credential for NodeGroup %s to clean Node in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = cluster.Region

	cli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("AddNodesToClusterTask[%s]: get tke client for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// 实例的状态（running 运行中，initializing 初始化中，failed 异常）
	allClusterInstance, err := cli.QueryTkeClusterAllInstances(cluster.SystemID)
	if err != nil {
		blog.Errorf("AddNodesToClusterTask[%s]: QueryTkeClusterAllInstances for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("QueryTkeClusterAllInstances err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	instanceIDMap := make(map[string]*api.InstanceInfo, 0)
	for i := range allClusterInstance {
		instanceIDMap[allClusterInstance[i].InstanceID] = allClusterInstance[i]
	}

	var (
		successNodes, failedNodes, timeoutNodes []string
	)
	// check current nodes if exist cluster
	existedInstance := make([]string, 0)
	notExistedInstance := make([]string, 0)
	for i := range idList {
		_, ok := instanceIDMap[idList[i]]
		if !ok {
			notExistedInstance = append(notExistedInstance, idList[i])
			continue
		}
		existedInstance = append(existedInstance, idList[i])
	}

	blog.Infof("AddNodesToClusterTask[%s] task[%s] existedInstance[%v] notExistedInstance[%v]",
		taskID, taskID, existedInstance, notExistedInstance)
	successNodes = append(successNodes, existedInstance...)

	// add instance to cluster
	if len(notExistedInstance) > 0 {
		req := &api.AddExistedInstanceReq{
			ClusterID:       cluster.SystemID,
			InstanceIDs:     notExistedInstance,
			AdvancedSetting: generateInstanceAdvanceInfo(cluster),
			NodePool:        nil,
			LoginSetting:    &api.LoginSettings{Password: initPasswd},
		}
		if nodeGroupID != "" {
			req.NodePool = &api.NodePoolOption{
				AddToNodePool: true,
				NodePoolID:    nodeGroupID,
			}
		}
		addNodesResp, err := cli.AddExistedInstancesToCluster(req)
		if err != nil {
			blog.Errorf("AddNodesToClusterTask[%s] AddExistedInstancesToCluster [task:%s step:%s] failed: %v",
				taskID, taskID, stepName, err)
			retErr := fmt.Errorf("AddExistedInstancesToCluster err, %s", err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}

		successNodes = append(successNodes, addNodesResp.SuccessInstanceIDs...)
		failedNodes = append(failedNodes, addNodesResp.FailedInstanceIDs...)
		timeoutNodes = append(timeoutNodes, addNodesResp.TimeoutInstanceIDs...)
	}

	blog.Infof("AddNodesToClusterTask[%s] task[%s] success[%v] failed[%v] timeout[%v]",
		taskID, taskID, successNodes, failedNodes, timeoutNodes)

	if len(idList) == len(successNodes) {
		blog.Infof("AddNodesToClusterTask[%s] task[%s] successful", taskID, taskID)
	}

	if task.CommonParams == nil {
		task.CommonParams = make(map[string]string)
	}

	task.CommonParams["Passwd"] = initPasswd
	task.CommonParams["successNodes"] = strings.Join(successNodes, ",")
	task.CommonParams["failedNodes"] = strings.Join(failedNodes, ",")
	task.CommonParams["timeoutNodes"] = strings.Join(timeoutNodes, ",")
	// set failed node status
	_ = updateNodeStatusByNodeID(failedNodes, common.StatusAddNodesFailed)
	_ = updateNodeStatusByNodeID(timeoutNodes, common.StatusAddNodesFailed)

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

	// get task form database
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("CheckAddNodesStatusTask[%s] task %s get detail task information from storage failed: %s, task retry", taskID, taskID, err.Error())
		return err
	}

	// task state check
	state := &cloudprovider.TaskState{
		Task:      task,
		JobResult: cloudprovider.NewJobSyncResult(task),
	}
	// check task already terminated
	if state.IsTerminated() {
		blog.Errorf("CheckAddNodesStatusTask[%s] task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	// workflow switch current step to stepName when previous task exec successful
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("CheckAddNodesStatusTask[%s] task %s not turn ro run step %s, err %s", taskID, taskID, stepName, err.Error())
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
	clusterID := step.Params["ClusterID"]
	nodeGroupID := step.Params["NodeGroupID"] // nodeGroup may be empty
	cloudID := step.Params["CloudID"]

	// get previous step paras
	successNodes := strings.Split(task.CommonParams["successNodes"], ",")

	// get nodes IDs and IPs
	ipList := strings.Split(step.Params["NodeIPs"], ",")
	idList := strings.Split(step.Params["NodeIDs"], ",")
	if len(idList) != len(ipList) {
		blog.Errorf("CheckAddNodesStatusTask[%s] [inner fatal] task %s step %s NodeID %d is not equal to InnerIP %d, fatal", taskID, taskID, stepName,
			len(idList), len(ipList))
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("NodeID & InnerIP params err"))
		return fmt.Errorf("task %s parameter err", taskID)
	}

	// handler logic
	cloud, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), cloudID, clusterID)
	if err != nil {
		blog.Errorf("CheckAddNodesStatusTask[%s] get cloud/project for NodeGroup %s to clean Node in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get cloud api_interface dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("CheckAddNodesStatusTask[%s] get credential for NodeGroup %s to clean Node in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = cluster.Region

	cli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("CheckAddNodesStatusTask[%s]: get tke client for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	describeInstance := &api.DescribeClusterInstances{
		ClusterID:   cluster.SystemID,
		InstanceIDs: successNodes,
	}

	// get add success or failed nodes
	addSuccessNodes := make([]string, 0)
	addFailureNodes := make([]string, 0)
	timeOut := false
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			blog.Infof("CheckAddNodesStatusTask[%s] QueryTkeClusterInstances timeout", taskID)
			timeOut = true
		}

		// timeOut quit
		if timeOut {
			break
		}

		instanceList, err := cli.QueryTkeClusterInstances(describeInstance)
		if err != nil {
			continue
		}

		index := 0
		running := make([]string, 0)
		failure := make([]string, 0)

		for i := range instanceList {
			blog.Infof("CheckAddNodesStatusTask[%s] instance[%s] status[%s]",
				taskID, *instanceList[i].InstanceId, *instanceList[i].InstanceState)
			switch *instanceList[i].InstanceState {
			case "running":
				running = append(running, *instanceList[i].InstanceId)
				index++
			case "failed":
				failure = append(failure, *instanceList[i].InstanceId)
				index++
			default:
			}
		}
		if index == len(instanceList) {
			addSuccessNodes = running
			addFailureNodes = failure
			break
		}
	}
	if timeOut {
		instanceList, err := cli.QueryTkeClusterInstances(describeInstance)
		if err != nil {
			blog.Errorf("CheckAddNodesStatusTask[%s] QueryTkeClusterInstances timeOut [taskId(%s): step(%s)] failed: %v",
				taskID, taskID, step, err)
			retErr := fmt.Errorf("QueryTkeClusterInstances err, %v", err)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}

		running := make([]string, 0)
		exception := make([]string, 0)

		for i := range instanceList {
			if *instanceList[i].InstanceState == "running" {
				running = append(running, *instanceList[i].InstanceId)
			} else {
				exception = append(exception, *instanceList[i].InstanceId)
			}
		}

		addSuccessNodes = running
		addFailureNodes = exception
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	if len(addSuccessNodes) > 0 {
		state.Task.CommonParams["addSuccessNodes"] = strings.Join(addSuccessNodes, ",")
	}
	if len(addFailureNodes) > 0 {
		state.Task.CommonParams["addFailedNodes"] = strings.Join(addFailureNodes, ",")
		_ = updateNodeStatusByNodeID(addFailureNodes, common.StatusAddNodesFailed)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckAddNodesStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// UpdateNodeDBInfoTask update node DB info
func UpdateNodeDBInfoTask(taskID string, stepName string) error {
	start := time.Now()

	// get task form database
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("UpdateNodeDBInfoTask[%s] task %s get detail task information from storage failed: %s, task retry", taskID, taskID, err.Error())
		return err
	}

	// task state check
	state := &cloudprovider.TaskState{
		Task:      task,
		JobResult: cloudprovider.NewJobSyncResult(task),
	}
	// check task already terminated
	if state.IsTerminated() {
		blog.Errorf("UpdateNodeDBInfoTask[%s] task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	// workflow switch current step to stepName when previous task exec successful
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("UpdateNodeDBInfoTask[%s] task %s not turn ro run step %s, err %s", taskID, taskID, stepName, err.Error())
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
	passwd := task.CommonParams["Passwd"]

	addSuccessNodes := strings.Split(state.Task.CommonParams["addSuccessNodes"], ",")
	addFailedNodes := strings.Split(state.Task.CommonParams["addFailedNodes"], ",")
	failedNodes := strings.Split(task.CommonParams["failedNodes"], ",")
	timeoutNodes := strings.Split(task.CommonParams["timeoutNodes"], ",")

	// get nodes IDs and IPs
	ipList := strings.Split(step.Params["NodeIPs"], ",")
	idList := strings.Split(step.Params["NodeIDs"], ",")
	if len(idList) != len(ipList) {
		blog.Errorf("UpdateNodeDBInfoTask[%s] [inner fatal] task %s step %s NodeID %d is not equal to InnerIP %d, fatal", taskID, taskID, stepName,
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
	if len(timeoutNodes) > 0 {
		failedInstances = append(failedInstances, timeoutNodes...)
	}

	instanceIPs := make([]string, 0)
	for _, instanceID := range successInstances {
		if ip, ok := nodeIDToIPMap[instanceID]; ok {
			instanceIPs = append(instanceIPs, ip)
		}
	}

	// update nodes status in DB
	for i := range successInstances {
		node, err := cloudprovider.GetStorageModel().GetNode(context.Background(), successInstances[i])
		if err != nil {
			continue
		}
		node.Passwd = passwd
		node.Status = common.StatusInitialization
		err = cloudprovider.GetStorageModel().UpdateNode(context.Background(), node)
		if err != nil {
			continue
		}
	}
	blog.Infof("UpdateNodeDBInfoTask[%s] step %s successful", taskID, stepName)

	// update failed nodes status in DB
	for i := range failedInstances {
		node, err := cloudprovider.GetStorageModel().GetNode(context.Background(), failedInstances[i])
		if err != nil {
			continue
		}
		node.Passwd = passwd
		node.Status = common.StatusAddNodesFailed
		err = cloudprovider.GetStorageModel().UpdateNode(context.Background(), node)
		if err != nil {
			continue
		}
	}

	// save common ips
	if state.Task.CommonParams == nil {
		task.CommonParams = make(map[string]string)
	}
	if len(instanceIPs) > 0 {
		state.Task.CommonParams["nodeIPList"] = strings.Join(instanceIPs, ",")
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateNodeDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
