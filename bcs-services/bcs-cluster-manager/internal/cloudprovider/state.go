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

package cloudprovider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
)

// GetTaskStateAndCurrentStep get task and task current step
func GetTaskStateAndCurrentStep(taskID, stepName string) (*TaskState, *proto.Step, error) {
	task, err := GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("task[%s] get detail task information from storage failed, %s. task retry",
			taskID, err.Error())
		return nil, nil, err
	}
	if task.CommonParams == nil {
		task.CommonParams = make(map[string]string)
	}

	state := &TaskState{Task: task, JobResult: NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("task[%s] is terminated, step %s skip", taskID, stepName)
		return nil, nil, fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("task[%s] not turn to run step %s, err %s", taskID, stepName, err.Error())
		return nil, nil, err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("task[%s]: current step[%s] successful and skip", taskID, stepName)
		return state, nil, nil
	}
	blog.Infof("task[%s]: run step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	return state, step, nil
}

// SetTaskStepParas set task step key:value
func SetTaskStepParas(taskID, stepName string, key, value string) error {
	task, err := GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("task[%s] get task information from storage failed, %v", taskID, err)
		return err
	}

	step, ok := task.Steps[stepName]
	if !ok {
		errMsg := fmt.Errorf("task[%s] not exist step[%s]", taskID, stepName)
		blog.Errorf(errMsg.Error())
		return errMsg
	}
	if step.Params == nil {
		step.Params = make(map[string]string)
	}
	step.Params[key] = value

	err = GetStorageModel().UpdateTask(context.Background(), task)
	if err != nil {
		return err
	}

	return nil
}

// TaskState handle task state
type TaskState struct {
	Task      *proto.Task
	JobResult *SyncJobResult
	// inject task url for bksops
	TaskUrl string
	// Manual CA scale nodes
	Manual bool
	// PartFailure step part failure state
	PartFailure bool
	// Message error log
	Message string
}

// IsTerminated check task already terminated
func (stat *TaskState) IsTerminated() bool {
	if stat.Task.Status == TaskStatusFailure || stat.Task.Status == TaskStatusForceTerminate ||
		stat.Task.Status == TaskStatusTimeout || stat.Task.Status == TaskStatusSuccess {
		return true
	}
	return false
}

// IsReadyToStep check current step or switch step
func (stat *TaskState) IsReadyToStep(stepName string) (*proto.Step, error) {
	switch stat.Task.Status {
	case TaskStatusRunning, TaskStatusInit:
	case TaskStatusForceTerminate:
		return nil, fmt.Errorf("task %s state for terminate", stat.Task.TaskID)
	default:
		return nil, fmt.Errorf("task %s is not running, state is %s", stat.Task.TaskID, stat.Task.Status)
	}

	// validate step existence
	curStep, exist := stat.Task.Steps[stepName]
	if !exist {
		return nil, fmt.Errorf("lost step")
	}

	// previous step success when retry task or skip scene
	switch curStep.Status {
	case TaskStatusSuccess, TaskStatusPartFailure, TaskStatusSkip:
		blog.Infof("task[%s] current step[%s] status[%s]", stat.Task.TaskID, stepName, curStep.Status)
		return nil, nil
	default:
		blog.Infof("task[%s] current step[%s] status[%s]", stat.Task.TaskID, stepName, curStep.Status)
	}

	// check turn to step
	if stepName != stat.Task.CurrentStep {
		// check if pre steps are all ok, then we can set sthi step running
		for _, name := range stat.Task.StepSequence {
			step, found := stat.Task.Steps[name]
			if !found {
				return nil, fmt.Errorf("task %s fatal, lost step %s in definition", stat.Task.TaskID, name)
			}

			// found current step
			if name == stepName {
				if step.Status == TaskStatusSuccess ||
					step.Status == TaskStatusPartFailure || step.Status == TaskStatusSkip {
					return nil, fmt.Errorf("task %s step %s already success", stat.Task.TaskID, stepName)
				}
				stat.Task.CurrentStep = stepName
				step.Start = time.Now().Format(time.RFC3339)
				step.Status = TaskStatusRunning
				step.Message = "step ready to run"
				step.LastUpdate = time.Now().Format(time.RFC3339)
				stat.Task.Steps[name] = step
				_ = GetStorageModel().UpdateTask(context.Background(), stat.Task)
				return step, nil
			}
			// skip step if failed
			if step.SkipOnFailed {
				continue
			}
			// check this step is ok
			if step.Status != TaskStatusSuccess &&
				step.Status != TaskStatusPartFailure && step.Status != TaskStatusSkip {
				// ok = false
				break
			}
		}
		return nil, fmt.Errorf("step %s don't turn to run, task already failed", stepName)
	}

	// refresh step status & task status
	if curStep.Status == TaskStatusFailure {
		curStep.Retry++
	}

	curStep.Start = time.Now().Format(time.RFC3339)
	curStep.Status = TaskStatusRunning
	curStep.Message = "step ready to run"
	curStep.LastUpdate = time.Now().Format(time.RFC3339)

	stat.Task.Status = TaskStatusRunning
	stat.Task.Message = fmt.Sprintf("step %s is running", stepName)
	stat.Task.LastUpdate = curStep.LastUpdate

	// update state in storage
	if err := GetStorageModel().UpdateTask(context.Background(), stat.Task); err != nil {
		blog.Errorf("task %s fatal, update task status failed, %s. required admin intervetion",
			stat.Task.TaskID, err.Error())
		return nil, err
	}
	blog.Infof("task %s step %s turn to running", stat.Task.TaskID, stepName)
	return curStep, nil
}

// UpdateStepSucc update step to success
func (stat *TaskState) UpdateStepSucc(start time.Time, stepName string) error {
	step := stat.Task.Steps[stepName]
	end := time.Now()
	step.ExecutionTime = uint32(end.Unix() - start.Unix())
	step.Start = start.Format(time.RFC3339)
	step.End = end.Format(time.RFC3339)
	step.Status = TaskStatusSuccess
	/*
		if stat.PartFailure {
			step.Status = TaskStatusPartFailure
		}
	*/
	step.LastUpdate = step.End
	step.Message = "running successfully"
	if stat.Message != "" {
		step.Message = stat.Message
	}
	if stat.TaskUrl != "" {
		step.Params[BkSopsTaskUrlKey.String()] = stat.TaskUrl
	}
	stat.Task.Status = TaskStatusRunning
	stat.Task.Message = fmt.Sprintf("step %s running successfully", step.Name)
	stat.Task.LastUpdate = step.End

	if stepName == stat.Task.StepSequence[len(stat.Task.StepSequence)-1] {
		// last step in task, just make whole task success
		taskStart, _ := time.Parse(time.RFC3339, stat.Task.Start)
		stat.Task.End = end.Format(time.RFC3339)
		stat.Task.ExecutionTime = uint32(end.Unix() - taskStart.Unix())
		stat.Task.Status = TaskStatusSuccess
		stat.Task.Message = "whole task is done"

		metrics.ReportMasterTaskMetric(stat.Task.TaskType, stat.Task.Status, "", taskStart)

		if stat.JobResult != nil {
			err := stat.JobResult.UpdateJobResultStatus(true)
			if err != nil {
				blog.Errorf("task[%s] stepName[%s] UpdateJobResultStatus failed: %v", stat.Task.TaskID, stepName, err)
			} else {
				blog.Infof("task[%s] stepName[%s] UpdateJobResultStatus successful", stat.Task.TaskID, stepName)
			}
		}
	} else {
		stepStart, _ := time.Parse(time.RFC3339, step.Start)
		metricName := step.Name
		if strings.HasPrefix(step.Name, BKSOPTask) {
			metricName = BKSOPTask
		}
		metrics.ReportMasterTaskMetric(stat.Task.TaskType, step.Status, metricName, stepStart)
	}

	if err := GetStorageModel().UpdateTask(context.Background(), stat.Task); err != nil {
		blog.Errorf("task %s fatal, update task success status failed, %s. required admin intervetion",
			stat.Task.TaskID, err.Error())
		return err
	}
	blog.Infof("task %s step %s running successfully", stat.Task.TaskID, stepName)
	return nil
}

// UpdateStepFailure update step failure
func (stat *TaskState) UpdateStepFailure(start time.Time, stepName string, err error) error {
	step := stat.Task.Steps[stepName]
	end := time.Now()
	step.ExecutionTime = uint32(end.Unix() - start.Unix())
	step.Start = start.Format(time.RFC3339)
	step.End = end.Format(time.RFC3339)
	step.Status = TaskStatusFailure
	step.LastUpdate = step.End
	step.Message = fmt.Sprintf("running failed, %s", err.Error())
	if stat.TaskUrl != "" {
		step.Params[BkSopsTaskUrlKey.String()] = stat.TaskUrl
	}

	taskStart, _ := time.Parse(time.RFC3339, stat.Task.Start)
	stat.Task.End = end.Format(time.RFC3339)
	stat.Task.ExecutionTime = uint32(end.Unix() - taskStart.Unix())
	stat.Task.Status = TaskStatusFailure
	stat.Task.Message = fmt.Sprintf("step %s running failed, %s", step.Name, err.Error())
	stat.Task.LastUpdate = step.End
	if err = GetStorageModel().UpdateTask(context.Background(), stat.Task); err != nil {
		blog.Errorf("task %s fatal, update task step %s failure status failed, %s. required admin intervetion",
			stat.Task.TaskID, stepName, err.Error())
		return err
	}

	metricName := step.Name
	if strings.HasPrefix(step.Name, BKSOPTask) {
		metricName = BKSOPTask
	}
	metrics.ReportMasterTaskMetric(stat.Task.TaskType, stat.Task.Status, metricName, taskStart)

	stepStart, _ := time.Parse(time.RFC3339, step.Start)
	metrics.ReportMasterTaskMetric(stat.Task.TaskType, step.Status, metricName, stepStart)

	if stat.JobResult != nil {
		err = stat.JobResult.UpdateJobResultStatus(false)
		if err != nil {
			blog.Errorf("task[%s] stepName[%s] UpdateJobResultStatus failed: %v", stat.Task.TaskID, stepName, err)
		} else {
			blog.Infof("task[%s] stepName[%s] UpdateJobResultStatus successful", stat.Task.TaskID, stepName)
		}
	}

	blog.Infof("task %s step %s running failure", stat.Task.TaskID, stepName)
	return nil
}

// SkipFailure step if need to skip failure status
func (stat *TaskState) SkipFailure(start time.Time, stepName string, err error) error {
	step := stat.Task.Steps[stepName]
	end := time.Now()
	step.ExecutionTime = uint32(end.Unix() - start.Unix())
	step.Start = start.Format(time.RFC3339)
	step.End = end.Format(time.RFC3339)
	step.Status = TaskStatusFailure
	step.LastUpdate = step.End
	step.Message = fmt.Sprintf("running failed, %s", err.Error())
	if stat.TaskUrl != "" {
		step.Params[BkSopsTaskUrlKey.String()] = stat.TaskUrl
	}

	stat.Task.Status = TaskStatusRunning
	stat.Task.Message = fmt.Sprintf("step %s running failed", step.Name)
	stat.Task.LastUpdate = step.End

	if stepName == stat.Task.StepSequence[len(stat.Task.StepSequence)-1] {
		// last step in task, just make whole task success
		taskStart, _ := time.Parse(time.RFC3339, stat.Task.Start)
		stat.Task.End = end.Format(time.RFC3339)
		stat.Task.ExecutionTime = uint32(end.Unix() - taskStart.Unix())
		stat.Task.Status = TaskStatusSuccess
		stat.Task.Message = "whole task is done"

		metrics.ReportMasterTaskMetric(stat.Task.TaskType, stat.Task.Status, "", taskStart)

		if stat.JobResult != nil {
			err := stat.JobResult.UpdateJobResultStatus(true)
			if err != nil {
				blog.Errorf("task[%s] stepName[%s] UpdateJobResultStatus failed: %v", stat.Task.TaskID, stepName, err)
			} else {
				blog.Infof("task[%s] stepName[%s] UpdateJobResultStatus successful", stat.Task.TaskID, stepName)
			}
		}
	} else {
		stepStart, _ := time.Parse(time.RFC3339, step.Start)
		metricName := step.Name
		if strings.HasPrefix(step.Name, BKSOPTask) {
			metricName = BKSOPTask
		}
		metrics.ReportMasterTaskMetric(stat.Task.TaskType, step.Status, metricName, stepStart)
	}

	if err := GetStorageModel().UpdateTask(context.Background(), stat.Task); err != nil {
		blog.Errorf("task %s fatal, update task success status failed, %s. required admin intervetion",
			stat.Task.TaskID, err.Error())
		return err
	}
	blog.Infof("task %s step %s running failed, but skip it", stat.Task.TaskID, stepName)
	return nil
}
