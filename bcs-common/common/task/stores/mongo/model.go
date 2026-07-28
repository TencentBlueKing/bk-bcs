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

package mongo

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// Task mongodb task model
type Task struct {
	TaskIndex           string            `bson:"taskIndex"`
	TaskIndexType       string            `bson:"taskIndexType"`
	TaskID              string            `bson:"taskID"`
	TaskType            string            `bson:"taskType"`
	TaskName            string            `bson:"taskName"`
	CurrentStep         string            `bson:"currentStep"`
	Steps               []*Step           `bson:"steps"`
	CallbackName        string            `bson:"callbackName"`
	CallbackResult      string            `bson:"callbackResult"`
	CallbackMessage     string            `bson:"callbackMessage"`
	CommonParams        map[string]string `bson:"commonParams"`
	CommonPayload       string            `bson:"commonPayload"`
	Status              string            `bson:"status"`
	Message             string            `bson:"message"`
	ExecutionTime       uint32            `bson:"executionTime"`
	MaxExecutionSeconds uint32            `bson:"maxExecutionSeconds"`
	Creator             string            `bson:"creator"`
	Updater             string            `bson:"updater"`
	Start               time.Time         `bson:"start"`
	End                 time.Time         `bson:"end"`
	CreatedAt           time.Time         `bson:"createdAt"`
	LastUpdate          time.Time         `bson:"lastUpdate"`
}

// Step mongodb step model
type Step struct {
	Name                string            `bson:"name"`
	Alias               string            `bson:"alias"`
	Executor            string            `bson:"executor"`
	Params              map[string]string `bson:"params"`
	Payload             string            `bson:"payload"`
	Status              string            `bson:"status"`
	Message             string            `bson:"message"`
	ETA                 *time.Time        `bson:"eta"` // 延迟执行时间(Estimated Time of Arrival)
	SkipOnFailed        bool              `bson:"skipOnFailed"`
	RetryCount          uint32            `bson:"retryCount"`
	MaxRetries          uint32            `bson:"maxRetries"`
	ExecutionTime       uint32            `bson:"executionTime"`
	MaxExecutionSeconds uint32            `bson:"maxExecutionSeconds"`
	Start               time.Time         `bson:"start"`
	End                 time.Time         `bson:"end"`
	LastUpdate          time.Time         `bson:"lastUpdate"`
}

// toMongoTask 将业务层 Task 转换为存储层 Task
func toMongoTask(task *types.Task) *Task {
	if task == nil {
		return nil
	}

	record := &Task{
		TaskID:              task.GetTaskID(),
		TaskType:            task.GetTaskType(),
		TaskName:            task.GetTaskName(),
		TaskIndex:           task.GetTaskIndex(),
		TaskIndexType:       task.GetTaskIndexType(),
		CurrentStep:         task.GetCurrentStep(),
		CallbackName:        task.GetCallback(),
		CallbackResult:      task.GetCallbackResult(),
		CallbackMessage:     task.GetCallbackMessage(),
		CommonParams:        task.CommonParams,
		CommonPayload:       task.CommonPayload,
		Status:              task.GetStatus(),
		Message:             task.GetMessage(),
		ExecutionTime:       task.ExecutionTime,
		MaxExecutionSeconds: task.MaxExecutionSeconds,
		Creator:             task.GetCreator(),
		Updater:             task.GetUpdater(),
		Start:               task.GetStartTime(),
		End:                 task.GetEndTime(),
		CreatedAt:           task.CreatedAt,
		LastUpdate:          task.LastUpdate,
		Steps:               make([]*Step, 0, len(task.Steps)),
	}

	for _, step := range task.Steps {
		record.Steps = append(record.Steps, toMongoStep(step))
	}

	return record
}

// toMongoStep 将业务层 Step 转换为存储层 Step
func toMongoStep(step *types.Step) *Step {
	if step == nil {
		return nil
	}

	return &Step{
		Name:                step.Name,
		Alias:               step.Alias,
		Executor:            step.Executor,
		Params:              step.Params,
		Payload:             step.Payload,
		Status:              step.Status,
		Message:             step.Message,
		ETA:                 step.ETA,
		SkipOnFailed:        step.SkipOnFailed,
		RetryCount:          step.RetryCount,
		MaxRetries:          step.MaxRetries,
		ExecutionTime:       step.ExecutionTime,
		MaxExecutionSeconds: step.MaxExecutionSeconds,
		Start:               step.Start,
		End:                 step.End,
		LastUpdate:          step.LastUpdate,
	}
}

// toTask 将存储层 Task 转换为业务层 Task
func toTask(record *Task) *types.Task {
	if record == nil {
		return nil
	}

	task := &types.Task{
		TaskID:              record.TaskID,
		TaskType:            record.TaskType,
		TaskName:            record.TaskName,
		TaskIndex:           record.TaskIndex,
		TaskIndexType:       record.TaskIndexType,
		CurrentStep:         record.CurrentStep,
		CallbackName:        record.CallbackName,
		CallbackResult:      record.CallbackResult,
		CallbackMessage:     record.CallbackMessage,
		CommonParams:        record.CommonParams,
		CommonPayload:       record.CommonPayload,
		Status:              record.Status,
		Message:             record.Message,
		ExecutionTime:       record.ExecutionTime,
		MaxExecutionSeconds: record.MaxExecutionSeconds,
		Creator:             record.Creator,
		Updater:             record.Updater,
		Start:               record.Start,
		End:                 record.End,
		CreatedAt:           record.CreatedAt,
		LastUpdate:          record.LastUpdate,
		Steps:               make([]*types.Step, 0, len(record.Steps)),
	}

	for _, step := range record.Steps {
		task.Steps = append(task.Steps, toStep(step))
	}

	return task
}

// toStep 将存储层 Step 转换为业务层 Step
func toStep(record *Step) *types.Step {
	if record == nil {
		return nil
	}

	return &types.Step{
		Name:                record.Name,
		Alias:               record.Alias,
		Executor:            record.Executor,
		Params:              record.Params,
		Payload:             record.Payload,
		Status:              record.Status,
		Message:             record.Message,
		ETA:                 record.ETA,
		SkipOnFailed:        record.SkipOnFailed,
		RetryCount:          record.RetryCount,
		MaxRetries:          record.MaxRetries,
		ExecutionTime:       record.ExecutionTime,
		MaxExecutionSeconds: record.MaxExecutionSeconds,
		Start:               record.Start,
		End:                 record.End,
		LastUpdate:          record.LastUpdate,
	}
}
