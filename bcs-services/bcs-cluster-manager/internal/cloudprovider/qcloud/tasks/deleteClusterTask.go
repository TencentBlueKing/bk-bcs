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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// DeleteTKEClusterTask delete cluster task
func DeleteTKEClusterTask(taskID string, stepName string) error { // nolint
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start delete tke cluster")
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
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
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	deleteMode := step.Params[cloudprovider.DeleteModeKey.String()]
	clusterStatus := step.Params[cloudprovider.LastClusterStatus.String()]

	// only support retain mode
	if deleteMode != cloudprovider.Retain.String() {
		deleteMode = cloudprovider.Retain.String()
	}

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("DeleteTKEClusterTask[%s]: GetClusterDependBasicInfo for cluster %s "+
			"in task %s step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// need to clean cluster nodes when cluster create or delete failed
	if (clusterStatus == icommon.StatusCreateClusterFailed ||
		clusterStatus == icommon.StatusDeleteClusterFailed) && dependInfo.Cluster.GetSystemID() != "" {
		_, workerNodes, errLocal := getClusterInstancesByClusterID(dependInfo)
		if errLocal == nil && len(workerNodes) > 0 {
			nodeIds := make([]string, 0)
			for i := range workerNodes {
				nodeIds = append(nodeIds, workerNodes[i].InstanceId)
			}

			// nolint
			ids, err := business.DeleteClusterInstance(ctx, dependInfo, nodeIds, false)
			if err != nil {
				blog.Errorf("DeleteTKEClusterTask[%s] DeleteClusterInstance failed: %v", taskID, err)
			} else {
				blog.Infof("DeleteTKEClusterTask[%s] DeleteClusterInstance success: %v", taskID, ids)
			}

			err = business.CheckClusterDeletedNodes(ctx, dependInfo, nodeIds)
			if err != nil {
				blog.Errorf("DeleteTKEClusterTask[%s] CheckClusterDeletedNodes failed: %v", taskID, err)
			}
			// this need to wait nodes to deleted status because of tke bug
			time.Sleep(time.Second * 60)
		}
	}

	err = business.DeleteTkeClusterByClusterId(ctx, dependInfo.CmOption, dependInfo.Cluster.SystemID, deleteMode)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("delete tke cluster failed [%s]", err))
		blog.Errorf("DeleteTKEClusterTask[%s]: task[%s] step[%s] call qcloud DeleteTKECluster failed: %v",
			taskID, taskID, stepName, err)
		retErr := fmt.Errorf("call qcloud DeleteTKECluster failed: %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("DeleteTKEClusterTask[%s]: task %s DeleteTKECluster[%s] successful",
		taskID, taskID, dependInfo.Cluster.SystemID)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"delete tke cluster successful")

	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("DeleteTKEClusterTask[%s]: task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// CleanClusterDBInfoTask clean cluster DB info
func CleanClusterDBInfoTask(taskID string, stepName string) error { // nolint
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start clean cluster db info")
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
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
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// delete cluster autoscalingOption
	err = cloudprovider.GetStorageModel().DeleteAutoScalingOption(context.Background(), cluster.ClusterID)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("delete cluste auto scaling option failed [%s]", err))
		blog.Errorf("CleanClusterDBInfoTask[%s]: clean cluster[%s] "+
			"autoscalingOption failed: %v", taskID, cluster.ClusterID, err)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"delete cluste autos caling option successful")

	// delete nodes
	err = cloudprovider.GetStorageModel().DeleteNodesByClusterID(context.Background(), cluster.ClusterID)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("delete nodes failed [%s]", err))
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete nodes for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete node for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete nodes for cluster[%s] in DB successful", taskID, clusterID)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"delete nodes successful")

	// delete nodeGroup
	err = cloudprovider.GetStorageModel().DeleteNodeGroupByClusterID(context.Background(), cluster.ClusterID)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("delete nodegroups failed [%s]", err))
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete nodeGroups for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete nodeGroups for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete nodeGroups for cluster[%s] in DB successful",
		taskID, clusterID)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"delete nodegroups successful")

	// delete CIDR and only print logInfo
	err = releaseClusterCIDR(cluster)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("release cluster cidr failed [%s]", err))
		blog.Errorf("CleanClusterDBInfoTask[%s]: releaseClusterCIDR[%s] cidr failed", taskID, clusterID)
	} else {
		cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
			"release cluster cidr successful")
		blog.Infof("CleanClusterDBInfoTask[%s]: releaseClusterCIDR[%s] cidr successful", taskID, clusterID)
	}

	// delete cluster
	cluster.Status = icommon.StatusDeleting
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("update cluster db info failed [%s]", err))
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete cluster for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete cluster[%s] in DB successful", taskID, clusterID)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"update cluster db info successful")

	utils.SyncDeletePassCCCluster(taskID, cluster)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"sync delete pass-cc cluster successful")

	_ = utils.DeleteClusterCredentialInfo(cluster.ClusterID)

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"delete cluster credential successful")

	// virtual cluster need to clean cluster token
	if cluster.ClusterType == icommon.ClusterTypeVirtual {
		_ = utils.DeleteBcsAgentToken(clusterID)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"clean cluster db info successful")

	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}
