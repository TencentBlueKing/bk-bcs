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
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
)

// DeleteCloudNodeGroupTask 删除节点池 - delete cloud node group task
func DeleteCloudNodeGroupTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		return nil
	}
	// extract parameter
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	// check validate
	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 {
		blog.Errorf("DeleteCloudNodeGroupTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("DeleteCloudNodeGroupTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("DeleteCloudNodeGroupTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("DeleteCloudNodeGroupTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if len(dependInfo.NodeGroup.CloudNodeGroupID) == 0 {
		blog.Errorf("DeleteCloudNodeGroupTask[%s]: nodegroup %s in task %s step %s has no autoscaling group",
			taskID, nodeGroupID, taskID, stepName)
		_ = state.UpdateStepSucc(start, stepName)
		return nil
	}

	// delete agentPool
	if err = deleteAgentPool(ctx, dependInfo); err != nil {
		blog.Errorf("DeleteCloudNodeGroupTask[%s]: deleteAgentPool[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call deleteAgentPool[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		return errors.Wrapf(err, "DeleteCloudNodeGroupTask[%s] task %s %s update to storage fatal", taskID,
			taskID, stepName)
	}
	return nil
}

// deleteAgentPool 删除节点池
func deleteAgentPool(rootCtx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	var (
		group       = info.NodeGroup
		cluster     = info.Cluster
		taskID      = cloudprovider.GetTaskIDFromContext(rootCtx)
		ctx, cancel = context.WithTimeout(rootCtx, 30*time.Second)
	)
	defer cancel()

	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption) // new Azure client
	if err != nil {
		return errors.Wrapf(err, "call NewAgentPoolClientWithOpt[%s] falied", taskID)
	}

	if _, err = client.GetPoolAndReturn(ctx, cloudprovider.GetClusterResourceGroup(info.Cluster),
		cluster.SystemID, group.CloudNodeGroupID); err != nil {
		if !(strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "not found")) {
			return errors.Wrapf(err, "deleteAgentPool[%s]: call GetPoolAndReturn[%s][%s] failed", taskID,
				cluster.SystemID, group.CloudNodeGroupID)
		}
		blog.Warnf("DeleteCloudNodeGroupTask[%s]: nodegroup[%s/%s] not found, skip delete",
			taskID, group.CloudNodeGroupID, group.CloudNodeGroupID)

		return nil
	}

	ctx, cancel = context.WithTimeout(rootCtx, 20*time.Minute)
	defer cancel()
	if err = client.DeletePool(ctx, info, cloudprovider.GetClusterResourceGroup(info.Cluster)); err != nil {
		return errors.Wrapf(err, "deleteAgentPool[%s]: call DeletePool[%s][%s] failed", taskID,
			cluster.SystemID, group.CloudNodeGroupID)
	}
	blog.Infof("DeleteCloudNodeGroupTask[%s]: call DeleteAgentPool successful", taskID)

	return nil
}
