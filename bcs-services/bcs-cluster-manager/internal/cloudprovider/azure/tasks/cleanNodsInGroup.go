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
	"github.com/avast/retry-go"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
)

// CleanNodeGroupNodesTask 缩容，不保留节点 - clean node group nodes task
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
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
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

	if err = removeVMSSsInstances(ctx, dependInfo, nodeIDs); err != nil {
		blog.Errorf("CleanNodeGroupNodesTask[%s] nodegroup %s removeVMSSsInstances failed: %v",
			taskID, nodeGroupID, err)
		retErr := fmt.Errorf("removeVMSSsInstances err, %v", err)
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

// removeVMSSsInstances 移除实例
func removeVMSSsInstances(rootCtx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeIDs []string) error {
	group := info.NodeGroup
	taskID := cloudprovider.GetTaskIDFromContext(rootCtx)
	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		return errors.Wrapf(err, "removeVMSSsInstances[%s] get cloud aks client err", taskID)
	}
	// check instances if exist
	instanceIDMap := make(map[string]bool)
	validateInstances := make([]string, 0)
	ctx, cancel := context.WithTimeout(rootCtx, 30*time.Second)
	defer cancel()

	// fetch node list
	vmList, err := client.ListInstanceAndReturn(ctx, group.AutoScaling.AutoScalingName, group.AutoScaling.AutoScalingID)
	if err != nil {
		return errors.Wrapf(err, "removeVMSSsInstances[%s] ListInstanceAndReturn[%s][%s] failed",
			taskID, group.CloudNodeGroupID, group.Name)
	}

	for _, ins := range vmList {
		instanceIDMap[*ins.InstanceID] = true
	}
	for _, nodeID := range nodeIDs {
		id := nodeID
		start := strings.IndexByte(id, '/')
		end := strings.LastIndexByte(id, '/')
		instanceID := id[start+1 : end]
		if _, ok := instanceIDMap[instanceID]; ok {
			validateInstances = append(validateInstances, instanceID)
		}
	}
	if len(validateInstances) == 0 {
		blog.Infof("removeVMSSsInstances[%s] validateInstances is empty", taskID)
		return nil
	}

	blog.Infof("removeVMSSsInstances[%s] validateInstances[%v]", taskID, validateInstances)

	err = retry.Do(func() error {
		ctxLocal, cancelLocal := context.WithTimeout(rootCtx, 10*time.Minute)
		defer cancelLocal()

		if errLocal := client.BatchDeleteVMs(ctxLocal, info, validateInstances); errLocal != nil {
			return errors.Wrapf(errLocal, "removeVMSSsInstances[%s] BatchDeleteVMs failed", taskID)
		}
		return nil
	}, retry.Attempts(10), retry.DelayType(retry.FixedDelay), retry.Delay(time.Second))
	if err != nil {
		return errors.Wrapf(err, "removeVMSSsInstances failed")
	}

	blog.Infof("removeVMSSsInstances[%s] BatchDeleteVMs[%v] successful[%s]", taskID, nodeIDs, group.Name)

	return nil
}
