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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// EnsureAutoScalerTask ensure auto scaler task, if not exist, create it, if exist, update it
func EnsureAutoScalerTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	state, step, err := getStateAndStep(taskID, "EnsureAutoScalerTask", stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	// TODO ensure auto scaler with helm api

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("EnsureAutoScalerTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// DeleteAutoScalerTask delete auto scaler task
func DeleteAutoScalerTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	state, step, err := getStateAndStep(taskID, "DeleteAutoScalerTask", stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	// TODO delete auto scaler with helm api

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("DeleteAutoScalerTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// UpdateNodeGroupAutoScalingDBTask delete auto scaler task
func UpdateNodeGroupAutoScalingDBTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	state, step, err := getStateAndStep(taskID, "UpdateNodeGroupAutoScalingDBTask", stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	nodeGroupID := step.Params["NodeGroupID"]

	np, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		blog.Errorf("UpdateNodeGroupAutoScalingDBTask[%s]: get cluster for %s failed", taskID, nodeGroupID)
		retErr := fmt.Errorf("get nodegroup information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	np.Status = icommon.StatusRunning

	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), np)
	if err != nil {
		blog.Errorf("UpdateNodeGroupAutoScalingDBTask[%s]: update nodegroup status for %s failed", taskID, np.Status)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateNodeGroupAutoScalingDBTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
