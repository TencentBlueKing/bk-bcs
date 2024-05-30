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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

// CleanNodeGroupNodesTask clean node group nodes task
func CleanNodeGroupNodesTask(taskID string, stepName string) error {
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

	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeIDs := strings.Split(step.Params[cloudprovider.NodeIDsKey.String()], ",")

	if len(clusterID) == 0 || len(nodeGroupID) == 0 || len(cloudID) == 0 || len(nodeIDs) == 0 {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("CleanNodeGroupNodesTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CleanNodeGroupNodesTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cluster := dependInfo.Cluster
	cmOption := dependInfo.CmOption

	client, err := api.NewCceClient(cmOption)
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: nodegroup %s in task %s step %s create as client error",
			taskID, nodeGroupID, taskID, stepName)
		retErr := fmt.Errorf("create as client err, %v", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("CleanNodeGroupNodesTask[%s]: nodeIds: %+v", taskID, nodeIDs)

	err = client.CleanNodePoolNodes(cluster.SystemID, nodeIDs)
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: nodegroup %s in task %s step %s clean nodepool error",
			taskID, nodeGroupID, taskID, stepName)
		retErr := fmt.Errorf("clean nodepool err, %v", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("CleanNodeGroupNodesTask[%s]: successfully", taskID)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// CheckClusterCleanNodsTask check cluster clean nodes task
func CheckClusterCleanNodsTask(taskID string, stepName string) error {
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

	// extract parameter && check validate
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIDsKey.String(), ",")

	if len(clusterID) == 0 || len(cloudID) == 0 || len(nodeIDs) == 0 {
		blog.Errorf("CheckClusterCleanNodsTask[%s]: check parameter validate failed", taskID)
		retErr := fmt.Errorf("CheckClusterCleanNodsTask check parameters failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CheckClusterCleanNodsTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterCleanNodsTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// wait check delete component status
	timeContext, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	err = loop.LoopDoFunc(timeContext, func() error {
		exist, notExist, err := business.FilterClusterInstanceFromNodesIDs(timeContext, dependInfo, nodeIDs) // nolint
		if err != nil {
			blog.Errorf("CheckClusterCleanNodsTask[%s] FilterClusterInstanceFromNodesIDs failed: %v",
				taskID, err)
			return nil
		}

		blog.Infof("CheckClusterCleanNodsTask[%s] nodeIDs[%v] exist[%v] notExist[%v]",
			taskID, nodeIDs, exist, notExist)

		if len(exist) == 0 {
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(30*time.Second))
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("CheckClusterCleanNodsTask[%s] cluster[%s] failed: %v", taskID, clusterID, err)
	}

	// timeout error
	if errors.Is(err, context.DeadlineExceeded) {
		blog.Infof("CheckClusterCleanNodsTask[%s] cluster[%s] timeout failed: %v", taskID, clusterID, err)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckClusterCleanNodsTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}
