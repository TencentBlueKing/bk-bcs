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

// Package task is a package for task management
package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/RichardKnop/machinery/v2/log"

	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// taskEndStatus task结束状态,处理超时和revoke
type taskEndStatus struct {
	status   string
	messsage string
}

// getTaskStateAndCurrentStep get task state and current step
func (m *TaskManager) getTaskState(taskId, stepName string) (*State, error) {
	task, err := GetGlobalStorage().GetTask(context.Background(), taskId)
	if err != nil {
		return nil, fmt.Errorf("get task %s information failed, %s", taskId, err.Error())
	}

	if task.CommonParams == nil {
		task.CommonParams = make(map[string]string, 0)
	}

	state := NewState(task, stepName)
	if state.isTaskTerminated() {
		return nil, fmt.Errorf("task %s is terminated, step %s skip", taskId, stepName)
	}
	step, err := state.isReadyToStep(stepName)
	if err != nil {
		return nil, fmt.Errorf("task %s step %s is not ready, %w", taskId, stepName, err)
	}

	if step == nil {
		// step successful and skip
		log.INFO.Printf("task %s step %s already execute successful", taskId, stepName)
		return state, nil
	}
	state.step = step

	// inject call back func
	if state.task.GetCallback() != "" && len(m.callbackExecutors) > 0 {
		name := istep.CallbackName(state.task.GetCallback())
		if cbExecutor, ok := m.callbackExecutors[name]; ok {
			state.cbExecutor = cbExecutor
		} else {
			log.WARNING.Println("task %s callback %s not registered, just ignore", taskId, name)
		}
	}

	return state, nil
}

// State is a struct for task state
type State struct {
	task       *types.Task
	step       *types.Step
	stepName   string
	cbExecutor istep.CallbackExecutor
}

// NewState return state relative to task
func NewState(task *types.Task, stepName string) *State {
	return &State{
		task:     task,
		stepName: stepName,
	}
}

// isTaskTerminated is terminated
func (s *State) isTaskTerminated() bool {
	status := s.task.GetStatus()
	if status == types.TaskStatusFailure ||
		status == types.TaskStatusSuccess ||
		status == types.TaskStatusRevoked ||
		status == types.TaskStatusTimeout {
		return true
	}
	return false
}

// isReadyToStep check if step is ready to step
func (s *State) isReadyToStep(stepName string) (*types.Step, error) {
	nowTime := time.Now()

	switch s.task.GetStatus() {
	case types.TaskStatusInit:
		s.task.SetStartTime(nowTime)
	case types.TaskStatusRunning:
	default:
		return nil, fmt.Errorf("task %s is not running, state is %s", s.task.GetTaskID(), s.task.GetStatus())
	}

	// validate step existence
	curStep, ok := s.task.GetStep(stepName)
	if !ok {
		return nil, fmt.Errorf("step %s is not exist", stepName)
	}
	s.task.SetCurrentStep(stepName).SetLastUpdate(nowTime)

	defer func() {
		// update Task in storage
		if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
			log.ERROR.Printf("task %s update step %s failed: %s", s.task.TaskID, curStep.GetName(), err.Error())
		}
	}()

	// return nil & nil means step had been executed
	// if task retring and so on, shoud update task status and ignore callback because task actually not execute
	if curStep.IsCompleted() {
		// task success
		taskStartTime := s.task.GetStartTime()
		if curStep.GetStatus() == types.TaskStatusSuccess {
			if s.isLastStep(curStep) {
				s.task.SetEndTime(nowTime).
					SetExecutionTime(taskStartTime, nowTime).
					SetStatus(types.TaskStatusSuccess).
					SetMessage("task finished successfully")
			}
			// step is success, skip
			return nil, nil
		}

		// task failed
		failMsg := fmt.Sprintf("step %s running failed", curStep.Name)
		if s.isLastStep(curStep) {
			if curStep.GetSkipOnFailed() {
				s.task.SetEndTime(nowTime).
					SetExecutionTime(taskStartTime, nowTime).
					SetStatus(types.TaskStatusSuccess).
					SetMessage("task finished successfully")
				return nil, nil
			}

			s.task.SetEndTime(nowTime).
				SetExecutionTime(taskStartTime, nowTime).
				SetStatus(types.TaskStatusFailure).
				SetMessage(failMsg)
			return nil, fmt.Errorf(failMsg)
		}

		if curStep.GetSkipOnFailed() {
			return nil, nil
		}

		s.task.SetEndTime(nowTime).
			SetExecutionTime(taskStartTime, nowTime).
			SetStatus(types.TaskStatusFailure).
			SetMessage(failMsg)
		return nil, fmt.Errorf(failMsg)
	}

	// not first time to execute current step
	if curStep.GetStatus() == types.TaskStatusFailure {
		curStep.AddRetryCount(1)
	}

	curStep = curStep.SetStartTime(nowTime).
		SetStatus(types.TaskStatusRunning).
		SetMessage("step ready to run").
		SetLastUpdate(nowTime)

	s.task.SetStatus(types.TaskStatusRunning).SetMessage("task running")
	return curStep, nil
}

// updateStepSuccess update step status to success
func (s *State) updateStepSuccess(start time.Time) {
	endTime := time.Now()

	defer func() {
		// update Task in storage
		if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
			log.ERROR.Printf("task %s update step %s to success failed: %s", s.task.TaskID, s.step.GetName(), err.Error())
		}
	}()

	s.step.SetEndTime(endTime).
		SetExecutionTime(start, endTime).
		SetStatus(types.TaskStatusSuccess).
		SetMessage(fmt.Sprintf("step %s running successfully", s.step.Name)).
		SetLastUpdate(endTime)

	taskStartTime := s.task.GetStartTime()
	s.task.SetStatus(types.TaskStatusRunning).
		SetExecutionTime(taskStartTime, endTime).
		SetMessage(fmt.Sprintf("step %s running successfully", s.step.Name)).
		SetLastUpdate(endTime)

	// last step
	if s.isLastStep(s.step) {
		s.task.SetEndTime(endTime).
			SetStatus(types.TaskStatusSuccess).
			SetMessage("task finished successfully")

			// callback
		if s.cbExecutor != nil {
			c := istep.NewContext(context.Background(), GetGlobalStorage(), s.task, s.step)
			s.cbExecutor.Callback(c, nil)
		}
	}
}

// updateStepFailure update step status to failure
func (s *State) updateStepFailure(start time.Time, stepErr error, taskStatus *taskEndStatus) {
	defer func() {
		// update Task in storage
		if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
			log.ERROR.Printf("task %s update step %s to failure failed: %s", s.task.TaskID, s.step.GetName(), err.Error())
		}
	}()

	endTime := time.Now()

	stepFailMsg := fmt.Sprintf("running failed, err=%s", stepErr)
	taskFailMsg := fmt.Sprintf("step %s running failed, err=%s", s.step.Name, stepErr)
	if s.step.MaxRetries > 0 {
		stepFailMsg = fmt.Sprintf("running failed, err=%s, retried=%d, maxRetries=%d",
			stepErr, s.step.GetRetryCount(), s.step.MaxRetries)
		taskFailMsg = fmt.Sprintf("step %s running failed, err=%s, retried=%d, maxRetries=%d",
			s.step.Name, stepErr, s.step.GetRetryCount(), s.step.MaxRetries)
	}

	s.step.SetEndTime(endTime).
		SetExecutionTime(start, endTime).
		SetStatus(types.TaskStatusFailure).
		SetMessage(stepFailMsg).
		SetLastUpdate(endTime)

	taskStartTime := s.task.GetStartTime()
	s.task.SetExecutionTime(taskStartTime, endTime).
		SetLastUpdate(endTime)

	// 任务超时, 整体结束
	if taskStatus != nil {
		if taskStatus.messsage != "" {
			taskFailMsg = taskStatus.messsage
		}
		s.task.SetEndTime(endTime).
			SetStatus(taskStatus.status).
			SetMessage(taskFailMsg)

		// callback
		if s.cbExecutor != nil {
			c := istep.NewContext(context.Background(), GetGlobalStorage(), s.task, s.step)
			s.cbExecutor.Callback(c, stepErr)
		}
		return
	}

	// last step failed and skipOnFailed is true, update task status to success
	if s.isLastStep(s.step) {
		if s.step.GetSkipOnFailed() {
			// ignore error
			stepErr = nil
			s.task.SetEndTime(endTime).
				SetStatus(types.TaskStatusSuccess).
				SetMessage("task finished successfully")
		} else {
			s.task.SetEndTime(endTime).
				SetStatus(types.TaskStatusFailure).
				SetMessage(taskFailMsg)
		}

		// callback
		if s.cbExecutor != nil {
			c := istep.NewContext(context.Background(), GetGlobalStorage(), s.task, s.step)
			s.cbExecutor.Callback(c, stepErr)
		}
		return
	}

	// 重试流程中
	if !errors.Is(stepErr, istep.ErrRevoked) && s.step.GetRetryCount() < s.step.MaxRetries {
		s.task.SetStatus(types.TaskStatusRunning).SetMessage(taskFailMsg)
		return
	}

	// 忽略错误
	if s.step.GetSkipOnFailed() {
		msg := fmt.Sprintf("step %s running failed, with skip on failed", s.step.Name)
		s.task.SetStatus(types.TaskStatusRunning).SetMessage(msg)
		return
	}

	// 重试次数用完且没有忽略错误
	s.task.SetEndTime(endTime).
		SetStatus(types.TaskStatusFailure).
		SetMessage(taskFailMsg)
}

func (s *State) isLastStep(step *types.Step) bool {
	count := len(s.task.Steps)
	// 没有step也就没有后续流程, 返回true
	if count == 0 {
		return true
	}

	// 非最后一步
	if step.GetName() != s.task.Steps[count-1].Name {
		return false
	}

	// 最后一步还需要看重试次数
	return step.IsCompleted()
}

// GetCommonParam get common params by key
func (s *State) GetCommonParam(key string) (string, bool) {
	return s.task.GetCommonParam(key)
}

// AddCommonParam add common params
func (s *State) AddCommonParam(key, value string) *State {
	s.task.AddCommonParam(key, value)
	return s
}

// GetCommonPayload get extra params by obj
func (s *State) GetCommonPayload(obj interface{}) error {
	return s.task.GetCommonPayload(obj)
}

// SetCommonPayload set extra params by obj
func (s *State) SetCommonPayload(obj interface{}) error {
	return s.task.SetCommonPayload(obj)
}

// GetStepParam get step params by key
func (s *State) GetStepParam(stepName, key string) (string, bool) {
	return s.task.GetStepParam(stepName, key)
}

// AddStepParams add step params
func (s *State) AddStepParams(stepName, key, value string) error {
	return s.task.AddStepParams(stepName, key, value)
}

// GetTask get task
func (s *State) GetTask() *types.Task {
	return s.task
}

// GetStep get step by stepName
func (s *State) GetStep(stepName string) (*types.Step, bool) {
	return s.task.GetStep(stepName)
}
