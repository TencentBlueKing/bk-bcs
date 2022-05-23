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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// DeleteTKEClusterTask delete cluster task
func DeleteTKEClusterTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("DeleteTKEClusterTask[%s]: task %s get detail task information from storage failed, %s. task retry", taskID, taskID, err.Error())
		return err
	}

	state := &cloudprovider.TaskState{Task: task, JobResult: cloudprovider.NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("DeleteTKEClusterTask[%s]: task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("DeleteTKEClusterTask[%s]: task %s not turn to run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("DeleteTKEClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("DeleteTKEClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params["ClusterID"]
	cloudID := step.Params["CloudID"]
	deleteMode := step.Params["DeleteMode"]

	// only support retain mode
	if deleteMode != cloudprovider.Retain.String() {
		deleteMode = cloudprovider.Retain.String()
	}

	cloud, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), cloudID, clusterID)
	if err != nil {
		blog.Errorf("DeleteTKEClusterTask[%s]: get cloud/project for cluster %s in task %s step %s failed, %s",
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
		blog.Errorf("DeleteTKEClusterTask[%s]: get credential for cluster %s in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = cluster.Region

	// get qcloud client
	cli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("DeleteTKEClusterTask[%s]: get tke client for cluster[%s] in task %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if cluster.SystemID != "" {
		err = cli.DeleteTKECluster(cluster.SystemID, api.DeleteMode(deleteMode))
		if err != nil {
			blog.Errorf("DeleteTKEClusterTask[%s]: task[%s] step[%s] call qcloud DeleteTKECluster failed: %v",
				taskID, taskID, stepName, err)
			retErr := fmt.Errorf("call qcloud DeleteTKECluster failed: %s", err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		_ = updateClusterSystemID(clusterID, "")
		blog.Infof("DeleteTKEClusterTask[%s]: task %s DeleteTKECluster[%s] successful", taskID, taskID, cluster.SystemID)
	} else {
		blog.Infof("DeleteTKEClusterTask[%s]: task %s DeleteTKECluster skip current step because SystemID empty", taskID, taskID)
	}

	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("DeleteTKEClusterTask[%s]: task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// CleanClusterDBInfoTask clean cluster DB info
func CleanClusterDBInfoTask(taskID string, stepName string) error {
	// delete node && nodeGroup && cluster
	// get relative nodes by clusterID
	start := time.Now()
	//get task information and validate
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: task %s get detail task information from storage failed, %s. task retry", taskID, taskID, err.Error())
		return err
	}

	state := &cloudprovider.TaskState{Task: task, JobResult: cloudprovider.NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("CleanClusterDBInfoTask[%s]: task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: task %s not turn to run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CleanClusterDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params["ClusterID"]
	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// delete nodes
	err = cloudprovider.GetStorageModel().DeleteNodesByClusterID(context.Background(), cluster.ClusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete nodes for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete node for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete nodes for cluster[%s] in DB successful", taskID, clusterID)

	// delete nodeGroup
	err = cloudprovider.GetStorageModel().DeleteNodeGroupByClusterID(context.Background(), cluster.ClusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete nodeGroups for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete nodeGroups for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete nodeGroups for cluster[%s] in DB successful", taskID, clusterID)

	// delete CIDR and only print logInfo
	err = releaseClusterCIDR(cluster)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: releaseClusterCIDR[%s] cidr failed", taskID, clusterID)
	} else {
		blog.Infof("CleanClusterDBInfoTask[%s]: releaseClusterCIDR[%s] cidr successful", taskID, clusterID)
	}

	// delete cluster
	cluster.Status = icommon.StatusDeleting
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete cluster for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete cluster[%s] in DB successful", taskID, clusterID)

	utils.SyncDeletePassCCCluster(taskID, cluster)
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}
