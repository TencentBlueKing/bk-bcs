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

package common

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
)

//* here are common tasks that for backgroup running
//* backgroup task running depends on machinery framework

// RunBKsopsJob running bksops job and wait for results
func RunBKsopsJob(taskID string, stepName string) error {
	// step1: get BKops url and para by taskID
	// step2: create bkops task
	// step3: start task & query status

	start := time.Now()

	// get task form database
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] task %s get detail task information from storage failed: %s, task retry",
			taskID, taskID, err.Error())
		return err
	}

	// task state check
	state := &cloudprovider.TaskState{
		Task:      task,
		JobResult: cloudprovider.NewJobSyncResult(task),
	}
	// check task already terminated
	if state.IsTerminated() {
		blog.Errorf("RunBKsopsJob[%s] task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	// workflow switch current step to stepName when previous task exec successful
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] task %s not turn ro run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RunBKsopsJob[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}

	blog.Infof("RunBKsopsJob[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// get common parameter
	url := step.Params["url"]
	bizID := step.Params["template_biz_id"]
	templateID := step.Params["template_id"]
	operator := step.Params["template_user"]
	constants := step.Params["constants"]

	taskName := task.CommonParams["taskName"]

	if url == "" || bizID == "" || operator == "" || templateID == "" || taskName == "" || constants == "" {
		errMsg := fmt.Sprintf("RunBKsopsJob[%s] validateParameter task[%s] step[%s] failed", taskID, taskID, stepName)
		blog.Errorf(errMsg)
		retErr := fmt.Errorf("RunBKsopsJob err, %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// extract constants parameter & inject dynamic value
	consMap := map[string]string{}
	err = json.Unmarshal([]byte(constants), &consMap)
	if err != nil {
		errMsg := fmt.Sprintf("RunBKsopsJob[%s] unmarshal constants failed[%v]", taskID, err)
		blog.Errorf(errMsg)

		retErr := fmt.Errorf("RunBKsopsJob err, %s", errMsg)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject dynamic parameter
	for ck, cv := range consMap {
		if value, ok := template.DynamicParameterInject[cv]; ok {
			consMap[ck] = task.CommonParams[value]
		}
	}

	// create task
	pathParas := &CreateTaskPathParas{
		BkBizID:    bizID,
		TemplateID: templateID,
		Operator:   operator,
	}
	createTaskReq := &CreateTaskRequest{
		TaskName:  taskName,
		Constants: consMap,
	}
	taskRes, err := BKOpsClient.CreateBkOpsTask(url, pathParas, createTaskReq)
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] CreateBkOpsTask task[%s] step[%s] failed; %v",
			taskID, task.TaskName, stepName, err)
		retErr := fmt.Errorf("CreateBkOpsTask err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// start task
	startTaskReq := &TaskPathParas{
		BkBizID:  bizID,
		TaskID:   fmt.Sprintf("%d", taskRes.Data.TaskID),
		Operator: operator,
	}
	_, err = BKOpsClient.StartBkOpsTask("", startTaskReq, &StartTaskRequest{})
	if err != nil {
		blog.Errorf("RunBKsopsJob[%s] StartBkOpsTask task[%s] step[%s] failed; %v", taskID, taskID, stepName, err)
		retErr := fmt.Errorf("StartBkOpsTask err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			blog.Errorf("RunBKsopsJob[%s] GetTaskStatus task[%s] step[%s] failed; %v", taskID, taskID, stepName, ctx.Err())
			retErr := fmt.Errorf("GetTaskStatus %s %s err, %s", startTaskReq.TaskID, "timeOut", ctx.Err())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return nil
		case <-ticker.C:
		}

		data, err := BKOpsClient.GetTaskStatus("", startTaskReq, &StartTaskRequest{})
		if err != nil {
			blog.Errorf("RunBKsopsJob[%s] GetTaskStatus failed: %v", taskID, err)
			continue
		}

		blog.Infof("RunBKsopsJob[%s] GetTaskStatus %s status %s", taskID, startTaskReq.TaskID, data.Data.State)
		if data.Data.State == FINISHED.String() {
			// update step
			_ = state.UpdateStepSucc(start, stepName)
			break
		}
		if data.Data.State == FAILED.String() || data.Data.State == REVOKED.String() || data.Data.State == SUSPENDED.String() {
			blog.Errorf("RunBKsopsJob[%s] GetTaskStatus task[%s] step[%s] failed: %v", taskID, taskID, stepName, err)
			retErr := fmt.Errorf("GetTaskStatus %s %s err, %v", startTaskReq.TaskID, data.Data.State, err)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
	}

	return nil
}
