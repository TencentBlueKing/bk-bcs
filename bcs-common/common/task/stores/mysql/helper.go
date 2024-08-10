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

package mysql

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

func getStepRecords(t *types.Task) []*StepRecords {
	records := make([]*StepRecords, 0, len(t.Steps))
	for _, step := range t.Steps {
		record := &StepRecords{
			TaskID:              t.TaskID,
			Name:                step.Name,
			Alias:               step.Alias,
			Extras:              step.Extras,
			Status:              step.Status,
			Message:             step.Message,
			SkipOnFailed:        step.SkipOnFailed,
			RetryCount:          step.RetryCount,
			Params:              step.Params,
			Start:               step.Start,
			End:                 step.End,
			ExecutionTime:       step.ExecutionTime,
			MaxExecutionSeconds: step.MaxExecutionSeconds,
		}
		records = append(records, record)
	}

	return records
}
func getTaskRecord(t *types.Task) *TaskRecords {
	stepSequence := make([]string, 0, len(t.Steps))
	for i := range t.Steps {
		stepSequence = append(stepSequence, t.Steps[i].Name)
	}

	record := &TaskRecords{
		TaskID:              t.TaskID,
		TaskType:            t.TaskType,
		TaskName:            t.TaskName,
		CurrentStep:         t.CurrentStep,
		StepSequence:        stepSequence,
		CallbackName:        t.CallbackName,
		CommonParams:        t.CommonParams,
		ExtraJson:           t.ExtraJson,
		Status:              t.Status,
		Message:             t.Message,
		ForceTerminate:      t.ForceTerminate,
		Start:               t.Start,
		End:                 t.End,
		ExecutionTime:       t.ExecutionTime,
		MaxExecutionSeconds: t.MaxExecutionSeconds,
		Creator:             t.Creator,
		Updater:             t.Updater,
	}
	return record
}

func toTask(task *TaskRecords, steps []*StepRecords) *types.Task {
	t := &types.Task{
		TaskID:              task.TaskID,
		TaskType:            task.TaskType,
		TaskName:            task.TaskName,
		CurrentStep:         task.CurrentStep,
		CallbackName:        task.CallbackName,
		CommonParams:        task.CommonParams,
		ExtraJson:           task.ExtraJson,
		Status:              task.Status,
		Message:             task.Message,
		ForceTerminate:      task.ForceTerminate,
		Start:               task.Start,
		End:                 task.End,
		ExecutionTime:       task.ExecutionTime,
		MaxExecutionSeconds: task.MaxExecutionSeconds,
		Creator:             task.Creator,
		LastUpdate:          task.UpdatedAt,
		Updater:             task.Updater,
	}

	t.Steps = make([]*types.Step, 0, len(steps))
	for _, step := range steps {
		t.Steps = append(t.Steps, step.ToStep())
	}
	return t
}

func getUpdateTaskRecord(t *types.Task) *TaskRecords {
	record := &TaskRecords{
		CurrentStep:   t.CurrentStep,
		CommonParams:  t.CommonParams,
		Status:        t.Status,
		Message:       t.Message,
		Start:         t.Start,
		End:           t.End,
		ExecutionTime: t.ExecutionTime,
		Updater:       t.Updater,
	}
	return record
}

func getUpdateStepRecord(t *types.Step) *StepRecords {
	record := &StepRecords{
		Params:              t.Params,
		Extras:              t.Extras,
		Status:              t.Status,
		Message:             t.Message,
		Start:               t.Start,
		End:                 t.End,
		ExecutionTime:       t.ExecutionTime,
		RetryCount:          t.RetryCount,
		MaxExecutionSeconds: t.MaxExecutionSeconds,
	}
	return record
}
