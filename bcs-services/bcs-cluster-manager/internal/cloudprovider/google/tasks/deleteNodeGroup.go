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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
)

// DeleteCloudNodeGroupTask delete cloud node group task
func DeleteCloudNodeGroupTask(taskID string, stepName string) error {
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
		blog.Errorf("DeleteCloudNodeGroupTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption := dependInfo.CmOption
	cluster := dependInfo.Cluster
	group := dependInfo.NodeGroup

	containerCli, err := api.NewContainerServiceClient(cmOption)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: get gke container client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud gke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	found := true
	if group.CloudNodeGroupID != "" {
		_, err = containerCli.GetClusterNodePool(context.Background(), cluster.SystemID, group.CloudNodeGroupID)
		if err != nil {
			if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "not found") {
				blog.Warnf("DeleteCloudNodeGroupTask[%s]: nodegroup[%s/%s] in task %s step %s not found, skip delete",
					taskID, nodeGroupID, group.CloudNodeGroupID, stepName, stepName)
				found = false
			} else {
				blog.Errorf(
					"DeleteCloudNodeGroupTask[%s]: call GetClusterNodePool[%s] api in task %s step %s failed, %s",
					taskID, nodeGroupID, taskID, stepName, err.Error())
				retErr := fmt.Errorf("call GetClusterNodePool[%s] api err, %s", nodeGroupID, err.Error())
				_ = state.UpdateStepFailure(start, stepName, retErr)
				return retErr
			}
		}
	}
	if found && group.CloudNodeGroupID != "" {
		_, err = containerCli.DeleteClusterNodePool(context.Background(), cluster.SystemID, group.CloudNodeGroupID)
		if err != nil {
			blog.Errorf("DeleteCloudNodeGroupTask[%s]: call DeleteClusterNodePool[%s] api in task %s step %s failed, %s",
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

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("DeleteCloudNodeGroupTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}
