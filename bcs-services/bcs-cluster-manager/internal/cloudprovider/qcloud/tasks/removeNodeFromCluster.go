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

// RemoveNodesFromClusterTask remove node from cluster
func RemoveNodesFromClusterTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("RemoveNodesFromClusterTask[%s]: task %s get detail task information from storage failed, %s. task retry", taskID, taskID, err.Error())
		return err
	}

	state := &cloudprovider.TaskState{Task: task, JobResult: cloudprovider.NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("RemoveNodesFromClusterTask[%s]: task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("RemoveNodesFromClusterTask[%s]: task %s not turn to run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RemoveNodesFromClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}

	blog.Infof("RemoveNodesFromClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// get data info
	clusterID := step.Params["ClusterID"]
	cloudID := step.Params["CloudID"]
	deleteMode := step.Params["DeleteMode"]

	// get nodes IDs and IPs
	ipList := strings.Split(step.Params["NodeIPs"], ",")
	idList := strings.Split(step.Params["NodeIDs"], ",")
	if len(idList) != len(ipList) {
		blog.Errorf("RemoveNodesFromClusterTask[%s] [inner fatal] task %s step %s NodeID %d is not equal to InnerIP %d, fatal", taskID, taskID, stepName,
			len(idList), len(ipList))
		_ = state.UpdateStepFailure(start, stepName, fmt.Errorf("NodeID & InnerIP params err"))
		return fmt.Errorf("task %s parameter err", taskID)
	}

	// step login started here
	cloud, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), cloudID, clusterID)
	if err != nil {
		blog.Errorf("RemoveNodesFromClusterTask[%s]: get cloud/project for cluster %s in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("RemoveNodesFromClusterTask[%s]: get credential for cluster %s in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = cluster.Region

	// get qcloud client
	cli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("RemoveNodesFromClusterTask[%s]: get tke client for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	allClusterInstance, err := cli.QueryTkeClusterAllInstances(cluster.SystemID)
	if err != nil {
		blog.Errorf("RemoveNodesFromClusterTask[%s]: QueryTkeClusterAllInstances for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("QueryTkeClusterAllInstances err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	instanceIDMap := make(map[string]*api.InstanceInfo)
	for i := range allClusterInstance {
		instanceIDMap[allClusterInstance[i].InstanceID] = allClusterInstance[i]
	}

	var (
		success, failed, notFound []string
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

	blog.Infof("RemoveNodesFromClusterTask[%s] task[%s] existedInstance[%v] notExistedInstance[%v]",
		taskID, taskID, existedInstance, notExistedInstance)

	if len(notExistedInstance) > 0 {
		success = append(success, notExistedInstance...)
	}

	if len(existedInstance) > 0 {
		req := &api.DeleteInstancesRequest{
			ClusterID:   cluster.SystemID,
			Instances:   existedInstance,
			DeleteMode:  api.DeleteMode(deleteMode),
			ForceDelete: true,
		}
		deleteResult, err := cli.DeleteTkeClusterInstance(req)
		if err != nil {
			blog.Errorf("RemoveNodesFromClusterTask[%s] DeleteTkeClusterInstance [task:%s step:%s] failed: %v",
				taskID, taskID, stepName, err)
			retErr := fmt.Errorf("DeleteTkeClusterInstance err, %s", err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		success = append(success, deleteResult.Success...)
		failed = append(failed, deleteResult.Failure...)
		notFound = append(notFound, deleteResult.NotFound...)
	}

	blog.Infof("RemoveNodesFromClusterTask[%s] DeleteTkeClusterInstance result, success[%v] failed[%v] notFound[%v]",
		taskID, success, failed, notFound)

	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	state.Task.CommonParams["successNodes"] = strings.Join(success, ",")
	state.Task.CommonParams["failedNodes"] = strings.Join(failed, ",")
	state.Task.CommonParams["notFoundNodes"] = strings.Join(notFound, ",")
	// set failed node status
	_ = updateNodeStatusByNodeID(failed, common.StatusRemoveNodesFailed)
	_ = updateNodeStatusByNodeID(notFound, common.StatusRemoveNodesFailed)

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RemoveNodesFromClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// UpdateRemoveNodeDBInfoTask update remove node DB info
func UpdateRemoveNodeDBInfoTask(taskID string, stepName string) error {
	start := time.Now()

	// get task form database
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("UpdateRemoveNodeDBInfoTask[%s] task %s get detail task information from storage failed: %s, task retry", taskID, taskID, err.Error())
		return err
	}

	// task state check
	state := &cloudprovider.TaskState{
		Task:      task,
		JobResult: cloudprovider.NewJobSyncResult(task),
	}
	// check task already terminated
	if state.IsTerminated() {
		blog.Errorf("UpdateRemoveNodeDBInfoTask[%s] task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	// workflow switch current step to stepName when previous task exec successful
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("UpdateRemoveNodeDBInfoTask[%s] task %s not turn ro run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateRemoveNodeDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}

	blog.Infof("UpdateRemoveNodeDBInfoTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	success := strings.Split(state.Task.CommonParams["successNodes"], ",")

	if len(success) > 0 {
		for i := range success {
			err = cloudprovider.GetStorageModel().DeleteNode(context.Background(), success[i])
			if err != nil {
				blog.Errorf("UpdateRemoveNodeDBInfoTask[%s] task %s deleteNodeByNodeID failed: %v", taskID, taskID, err)
			}
		}
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateNodeDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
