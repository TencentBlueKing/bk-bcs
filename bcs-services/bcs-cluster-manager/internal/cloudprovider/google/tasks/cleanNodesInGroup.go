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

// Package tasks xxx
package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	compute "google.golang.org/api/compute/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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
	nodeIDs := strings.Split(state.Task.CommonParams[cloudprovider.NodeIDsKey.String()], ",")

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

	if dependInfo.NodeGroup.AutoScaling == nil || dependInfo.NodeGroup.AutoScaling.AutoScalingID == "" {
		blog.Errorf("CleanNodeGroupNodesTask[%s]: nodegroup %s in task %s step %s has no autoscaling group",
			taskID, nodeGroupID, taskID, stepName)
		retErr := fmt.Errorf("get autoScalingID err, %v", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	err = deleteIgmInstances(ctx, dependInfo, nodeIDs)
	if err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s] nodegroup %s removeAsgInstances failed: %v",
			taskID, nodeGroupID, err)
		retErr := fmt.Errorf("removeAsgInstances err, %v", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func deleteIgmInstances(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeNames []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	igmInfo, err := api.GetGCEResourceInfo(info.NodeGroup.AutoScaling.AutoScalingID)
	if err != nil {
		return fmt.Errorf("deleteIgmInstances[%s] get igm info failed: %v", taskID, err)
	}

	client, err := api.NewComputeServiceClient(info.CmOption)
	if err != nil {
		blog.Errorf("deleteIgmInstances[%s] get gce client failed: %v", taskID, err.Error())
		return err
	}

	// check instances if exist
	var (
		instanceNameList, validateInstances = make([]string, 0), make([]string, 0)
	)
	igmInstances, err := client.ListInstanceGroupsInstances(ctx, igmInfo[3], igmInfo[(len(igmInfo)-1)])
	if err != nil {
		blog.Errorf("deleteIgmInstances[%s] ListInstanceGroupsInstances[%s] failed: %v", taskID,
			igmInfo[(len(igmInfo)-1)], err.Error())
		return err
	}
	for _, ins := range igmInstances {
		insInfo, errInfo := api.GetGCEResourceInfo(ins.Instance)
		if errInfo != nil {
			return err
		}
		instanceNameList = append(instanceNameList, insInfo[len(insInfo)-1])
	}
	for _, id := range nodeNames {
		if utils.StringInSlice(id, instanceNameList) {
			validateInstances = append(validateInstances, id)
		}
	}
	if len(validateInstances) == 0 {
		blog.Infof("deleteIgmInstances[%s] validateInstances is empty", taskID)
		return nil
	}
	blog.Infof("deleteIgmInstances[%s] validateInstances[%v]", taskID, validateInstances)

	var (
		operation *compute.Operation
		zone      = info.NodeGroup.Region
	)

	zones := info.NodeGroup.GetAutoScaling().GetZones()
	if len(zones) > 0 {
		zone = zones[0]
	}

	err = retry.Do(func() error {
		var errLocal error
		operation, errLocal = client.DeleteMigInstances(ctx, zone, igmInfo[len(igmInfo)-1], validateInstances)
		if errLocal != nil {
			blog.Errorf("deleteIgmInstances[%s] DeleteInstancesInMIG failed: %v", taskID, errLocal)
			return errLocal
		}
		blog.Infof("deleteIgmInstances[%s] DeleteInstancesInMIG[%v] successful", taskID, validateInstances)

		return nil
	}, retry.Attempts(3))
	if err != nil {
		return err
	}

	return checkOperationStatus(client, operation.SelfLink, taskID, time.Second*5)
}
