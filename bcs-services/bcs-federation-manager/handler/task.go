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

// Package handler task service
package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	taskstore "github.com/Tencent/bk-bcs/bcs-common/common/task/store"
	tasktypes "github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	fedtasks "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/tasks"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"
	"golang.org/x/exp/slices"
)

const (
	// default offset and length for list task
	DefaultTaskListOffset = 0
	DefaultTaskListLength = 1000
	DefaultTokenMask      = "********"
)

// GetTask get task
func (f *FederationManager) GetTask(ctx context.Context,
	req *federationmgr.GetTaskRequest, resp *federationmgr.GetTaskResponse) error {

	blog.Infof("Receive GetTask request, taskId: %s", req.GetTaskId())

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate GetTask request failed, err: %s", err.Error()))
	}

	// get task
	task, err := f.taskmanager.GetTaskWithID(ctx, req.GetTaskId())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("get task %s failed, err: %s", req.GetTaskId(), err.Error()))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = transferTask(task)

	return nil
}

// GetTaskRecord get task record
func (f *FederationManager) GetTaskRecord(ctx context.Context,
	req *federationmgr.GetTaskRecordRequest, resp *federationmgr.GetTaskRecordResponse) error {

	blog.Infof("Receive GetTaskRecord request, taskId: %s", req.GetTaskId())

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate GetTaskRecord request failed, err: %s", err.Error()))
	}

	// get task
	task, err := f.taskmanager.GetTaskWithID(ctx, req.GetTaskId())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("get task %s failed, err: %s", req.GetTaskId(), err.Error()))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = transferTaskRecord(task)

	return nil
}

// ListTasks list tasks
func (f *FederationManager) ListTasks(ctx context.Context,
	req *federationmgr.ListTasksRequest, resp *federationmgr.ListTasksResponse) error {

	blog.Infof("Receive ListTasks request, taskType: %s, taskIndex: %s", req.GetTaskType(), req.GetTaskIndex())

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate ListTasks request failed, err: %s", err.Error()))
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"taskType":  req.GetTaskType(),
		"taskIndex": req.GetTaskIndex(),
	})

	opt := &taskstore.ListOption{
		Offset: DefaultTaskListOffset,
		Limit:  DefaultTaskListLength,
		Sort: map[string]int{
			"start": -1,
		},
	}

	if req.GetOffset() != 0 {
		opt.Offset = int64(req.GetOffset())
	}

	if req.GetLimit() != 0 {
		opt.Limit = int64(req.GetLimit())
	}

	// list task
	tasks, err := f.taskmanager.ListTask(ctx, cond, opt)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("ListTasks error, err %s", err.Error()))
	}

	data := make([]*federationmgr.Task, 0)
	for _, task := range tasks {
		data = append(data, transferTask(&task))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = data

	return nil
}

// RetryTask retry task
func (f *FederationManager) RetryTask(ctx context.Context,
	req *federationmgr.RetryTaskRequest, resp *federationmgr.RetryTaskResponse) error {

	blog.Infof("Receive RetryTask request, taskId: %s, beginStep: %s", req.GetTaskId(), req.GetBeginStepName())

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate RetryTask request failed, err: %s", err.Error()))
	}

	taskId, beginStep := req.GetTaskId(), req.GetBeginStepName()
	if taskId == "" {
		return ErrReturn(resp, "taskId is empty")
	}

	task, err := f.taskmanager.GetTaskWithID(ctx, taskId)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("get task %s failed, err %s", taskId, err.Error()))
	}

	// only task status is failure, can retry
	if task.GetStatus() != tasktypes.TaskStatusFailure {
		return ErrReturn(resp, fmt.Sprintf("task %s is not finished", taskId))
	}

	// reset status when retry task
	if task.GetTaskType() == fedtasks.InstallFederationTaskName.Type {
		// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
		if err := fedtasks.BeforeRetryInstallFederation(ctx, task); err != nil {
			return ErrReturn(resp, fmt.Sprintf("before retry install federation failed, err %s", err.Error()))
		}
	} else if task.GetTaskType() == fedtasks.RegisterSubclusterTaskName.Type {
		// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
		if err := fedtasks.BeforeRetryRegisterSubCluster(ctx, task); err != nil {
			return ErrReturn(resp, fmt.Sprintf("before retry register subcluster failed, err %s", err.Error()))
		}
	}

	if beginStep == "" {
		err = f.taskmanager.RetryAll(task)
	} else {
		// begin step must exist in task step sequence
		if !slices.Contains(task.StepSequence, beginStep) {
			return ErrReturn(resp, "beginStep is not in step sequence")
		}
		err = f.taskmanager.RetryAt(task, beginStep)
	}

	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("retry task %s failed from begin step: %s, err %s", taskId, beginStep, err.Error()))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = &federationmgr.TaskDistributeResponseData{
		TaskId: taskId,
	}

	return nil
}

func transferTask(task *tasktypes.Task) *federationmgr.Task {
	newTask := &federationmgr.Task{
		TaskIndex:           task.TaskIndex,
		TaskId:              task.TaskID,
		TaskType:            task.TaskType,
		TaskName:            task.TaskName,
		CurrentStep:         task.CurrentStep,
		StepSequence:        task.StepSequence,
		CallBackFuncName:    task.CallBackFuncName,
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
		LastUpdate:          task.LastUpdate,
		Updater:             task.Updater,
		Steps:               make(map[string]*federationmgr.Step),
	}

	// mask token info
	for k := range newTask.CommonParams {
		if k == steps.UserTokenKey {
			newTask.CommonParams[k] = DefaultTokenMask
		}
	}

	for key, step := range task.Steps {
		newStep := &federationmgr.Step{
			Name:                step.Name,
			Alias:               step.Alias,
			Params:              step.Params,
			Extras:              step.Extras,
			Status:              step.Status,
			Message:             step.Message,
			SkipOnFailed:        step.SkipOnFailed,
			RetryCount:          step.RetryCount,
			Start:               step.Start,
			End:                 step.End,
			ExecutionTime:       step.ExecutionTime,
			MaxExecutionSeconds: step.MaxExecutionSeconds,
			LastUpdate:          step.LastUpdate,
		}
		// mask token info
		for k := range newStep.Params {
			if k == steps.UserTokenKey {
				newStep.Params[k] = DefaultTokenMask
			}
		}
		newTask.Steps[key] = newStep

	}
	return newTask
}

func transferTaskRecord(task *tasktypes.Task) *federationmgr.TaskRecord {
	record := &federationmgr.TaskRecord{
		Status: task.Status,
		Step:   make([]*federationmgr.TaskRecordStep, 0, len(task.StepSequence)),
	}

	for _, stepName := range task.StepSequence {
		step, ok := task.GetStep(stepName)
		if !ok {
			blog.Errorf("get step %s failed", stepName)
			continue
		}

		startTime := transferRFC3339Time(step.Start)
		endTime := transferRFC3339Time(step.End)
		lastUpdateTime := transferRFC3339Time(step.LastUpdate)

		messageLevel := "INFO"
		if step.GetStatus() != tasktypes.TaskStatusSuccess {
			messageLevel = "ERROR"
		}

		recordStep := &federationmgr.TaskRecordStep{
			Name:       step.GetAlias(),
			Status:     step.GetStatus(),
			StartTime:  startTime,
			EndTime:    endTime,
			AllowSkip:  &wrappers.BoolValue{Value: step.GetSkipOnFailed()},
			AllowRetry: &wrappers.BoolValue{Value: true},
			Data: []*federationmgr.TaskRecordStepData{
				{
					Log:       fmt.Sprintf("name: %s", step.GetName()),
					Timestamp: lastUpdateTime,
					Level:     "INFO",
				},
				{
					Log:       fmt.Sprintf("message: %s", step.GetMessage()),
					Timestamp: lastUpdateTime,
					Level:     messageLevel,
				},
				{
					Log:       fmt.Sprintf("execute time: %dms", step.ExecutionTime),
					Timestamp: lastUpdateTime,
					Level:     "INFO",
				},
			},
		}
		record.Step = append(record.Step, recordStep)
	}

	return record
}

func transferRFC3339Time(rfc3339 string) int64 {
	if rfc3339 == "" {
		return 0
	}

	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		blog.Errorf("parse time %s failed, err %s", rfc3339, err.Error())
		return 0
	}

	return t.UnixNano() / int64(time.Millisecond)
}
