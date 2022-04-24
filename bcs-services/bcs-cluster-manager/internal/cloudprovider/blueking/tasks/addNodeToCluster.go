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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// UpdateAddNodeDBInfoTask update nodes DB info
func UpdateAddNodeDBInfoTask(taskID string, stepName string) error {
	start := time.Now()

	// get task form database
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("UpdateAddNodeDBInfoTask[%s] task %s get detail task information from storage failed: %s, task retry", taskID, taskID, err.Error())
		return err
	}

	// task state check
	state := &cloudprovider.TaskState{
		Task:      task,
		JobResult: cloudprovider.NewJobSyncResult(task),
	}
	// check task already terminated
	if state.IsTerminated() {
		blog.Errorf("UpdateAddNodeDBInfoTask[%s] task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	// workflow switch current step to stepName when previous task exec successful
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("UpdateAddNodeDBInfoTask[%s] task %s not turn ro run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateAddNodeDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}

	blog.Infof("UpdateAddNodeDBInfoTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	nodeIPs := strings.Split(step.Params["NodeIPs"], ",")

	// update nodes status in DB
	for i := range nodeIPs {
		node, err := cloudprovider.GetStorageModel().GetNodeByIP(context.Background(), nodeIPs[i])
		if err != nil {
			continue
		}
		node.Status = common.StatusInitialization
		err = cloudprovider.GetStorageModel().UpdateNode(context.Background(), node)
		if err != nil {
			continue
		}
	}
	blog.Infof("UpdateAddNodeDBInfoTask[%s] step %s successful", taskID, stepName)

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateNodeDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
