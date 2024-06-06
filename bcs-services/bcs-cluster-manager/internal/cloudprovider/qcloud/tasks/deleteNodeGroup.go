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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// DeleteCloudNodeGroupTask delete cloud node group task
func DeleteCloudNodeGroupTask(taskID string, stepName string) error { // nolint
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start delete cloud nodegroup")
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CheckCleanNodeGroupNodesStatusTask[%s]: GetClusterDependBasicInfo "+
			"for nodegroup %s in task %s step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	keepInstance := false
	if step.Params[cloudprovider.KeepInstanceKey.String()] == common.True {
		keepInstance = true
	}

	// create node group
	tkeCli, err := api.NewTkeClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("DeleteCloudNodeGroupTask[%s]: get tke client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}
	found := true
	if dependInfo.NodeGroup.CloudNodeGroupID != "" {
		_, errPool := tkeCli.DescribeClusterNodePoolDetail(dependInfo.Cluster.SystemID,
			dependInfo.NodeGroup.CloudNodeGroupID)
		if errPool != nil {
			if !(strings.Contains(errPool.Error(), "NotFound") ||
				strings.Contains(errPool.Error(), "not found")) {
				cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
					fmt.Sprintf("describe cluster nodepool detail failed [%s]", errPool))
				blog.Errorf("DeleteCloudNodeGroupTask[%s]: call DescribeClusterNodePoolDetail[%s] "+
					"api in task %s step %s failed, %s",
					taskID, nodeGroupID, taskID, stepName, errPool.Error())
				retErr := fmt.Errorf("call DescribeClusterNodePoolDetail[%s] api err, %s", nodeGroupID,
					errPool.Error())
				_ = state.UpdateStepFailure(start, stepName, retErr)
				return retErr
			}
			blog.Warnf("DeleteCloudNodeGroupTask[%s]: nodegroup[%s/%s] in task %s step %s not found, skip delete",
				taskID, nodeGroupID, dependInfo.NodeGroup.CloudNodeGroupID, stepName, stepName)
			found = false
		}
	}
	if found && dependInfo.NodeGroup.CloudNodeGroupID != "" {
		err = tkeCli.DeleteClusterNodePool(dependInfo.Cluster.SystemID, []string{dependInfo.NodeGroup.CloudNodeGroupID},
			keepInstance)
		if err != nil {
			cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
				fmt.Sprintf("delete cluster nodepool failed [%s]", err))
			blog.Errorf("DeleteCloudNodeGroupTask[%s]: call DeleteClusterNodePool[%s] api in "+
				"task %s step %s failed, %s",
				taskID, nodeGroupID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("call DeleteClusterNodePool[%s] api err, %s", nodeGroupID, err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
	}
	blog.Infof("DeleteCloudNodeGroupTask[%s]: call DeleteClusterNodePool successful", taskID)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"delete cloud nodegroup successful")

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("DeleteCloudNodeGroupTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}
