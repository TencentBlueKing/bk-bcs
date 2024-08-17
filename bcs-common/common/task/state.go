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
	"fmt"
	"time"

	"github.com/RichardKnop/machinery/v2/log"

	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// getTaskStateAndCurrentStep get task state and current step
func (m *TaskManager) getTaskStateAndCurrentStep(taskId, stepName string) (*State, *types.Step, error) {
	task, err := GetGlobalStorage().GetTask(context.Background(), taskId)
	if err != nil {
		return nil, nil, fmt.Errorf("get task %s information failed, %s", taskId, err.Error())
	}

	if task.CommonParams == nil {
		task.CommonParams = make(map[string]string, 0)
	}

	state := NewState(task, stepName)
	if state.isTaskTerminated() {
		return nil, nil, fmt.Errorf("task %s is terminated, step %s skip", taskId, stepName)
	}
	step, err := state.isReadyToStep(stepName)
	if err != nil {
		return nil, nil, fmt.Errorf("task %s step %s is not ready, %w", taskId, stepName, err)
	}

	if step == nil {
		// step successful and skip
		log.INFO.Printf("task %s step %s already execute successful", taskId, stepName)
		return state, nil, nil
	}

	// inject call back func
	if state.task.GetCallback() != "" && len(m.callbackExecutors) > 0 {
		name := istep.CallbackName(state.task.GetCallback())
		if cbExecutor, ok := m.callbackExecutors[name]; ok {
			state.cbExecutor = cbExecutor
		} else {
			log.WARNING.Println("task %s callback %s not registered, just ignore", taskId, name)
		}
	}

	return state, step, nil
}

// State is a struct for task state
type State struct {
	task        *types.Task
	currentStep string
	cbExecutor  istep.CallbackExecutor
}

// NewState return state relative to task
func NewState(task *types.Task, currentStep string) *State {
	return &State{
		task:        task,
		currentStep: currentStep,
	}
}

// isTaskTerminated is terminated
func (s *State) isTaskTerminated() bool {
	status := s.task.GetStatus()
	if status == types.TaskStatusFailure || status == types.TaskStatusForceTerminate ||
		status == types.TaskStatusTimeout || status == types.TaskStatusSuccess {
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
	case types.TaskStatusForceTerminate:
		return nil, fmt.Errorf("task %s state for terminate", s.task.GetTaskID())
	default:
		return nil, fmt.Errorf("task %s is not running, state is %s", s.task.GetTaskID(), s.task.GetStatus())
	}

	// validate step existence
	curStep, ok := s.task.GetStep(stepName)
	if !ok {
		return nil, fmt.Errorf("step %s is not exist", stepName)
	}

	// return nil & nil means  step had been executed
	if curStep.GetStatus() == types.TaskStatusSuccess {
		// step is success, skip
		return nil, nil
	}

	defer func() {
		// update Task in storage
		if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
			log.ERROR.Printf("task %s update step %s failed: %s", s.task.TaskID, curStep.GetName(), err.Error())
		}
	}()

	// not first time to execute current step
	if curStep.GetStatus() == types.TaskStatusFailure {
		curStep.AddRetryCount(1)
	}

	curStep = curStep.SetStartTime(nowTime).
		SetStatus(types.TaskStatusRunning).
		SetMessage("step ready to run").
		SetLastUpdate(nowTime)

	s.task.SetCurrentStep(stepName).SetStatus(types.TaskStatusRunning).SetMessage("task running")
	return curStep, nil
}

// updateStepSuccess update step status to success
func (s *State) updateStepSuccess(start time.Time, stepName string) error {
	step, ok := s.task.GetStep(stepName)
	if !ok {
		return fmt.Errorf("step %s is not exist", stepName)
	}

	endTime := time.Now()

	defer func() {
		// update Task in storage
		if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
			log.ERROR.Printf("task %s update step %s to success failed: %s", s.task.TaskID, step.GetName(), err.Error())
		}
	}()

	step.SetEndTime(endTime).
		SetExecutionTime(start, endTime).
		SetStatus(types.TaskStatusSuccess).
		SetMessage(fmt.Sprintf("step %s running successfully", step.Name)).
		SetLastUpdate(endTime)

	taskStartTime := s.task.GetStartTime()
	s.task.SetStatus(types.TaskStatusRunning).
		SetExecutionTime(taskStartTime, endTime).
		SetMessage(fmt.Sprintf("step %s running successfully", step.Name)).
		SetLastUpdate(endTime)

	// last step
	if s.isLastStep(step) {
		s.task.SetEndTime(endTime).
			SetStatus(types.TaskStatusSuccess).
			SetMessage("task finished successfully")

			// callback
		if s.cbExecutor != nil {
			c := istep.NewContext(context.Background(), GetGlobalStorage(), s.task, step)
			s.cbExecutor.Callback(c, nil)
		}
	}

	return nil
}

// updateStepFailure update step status to failure
func (s *State) updateStepFailure(start time.Time, name string, stepErr error, taskTimeOut bool) error {
	step, ok := s.task.GetStep(name)
	if !ok {
		return fmt.Errorf("step %s is not exist", name)
	}

	defer func() {
		// update Task in storage
		if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
			log.ERROR.Printf("task %s update step %s to failure failed: %s", s.task.TaskID, step.GetName(), err.Error())
		}
	}()

	endTime := time.Now()

	stepMsg := fmt.Sprintf("running failed, err=%s", stepErr)
	if step.MaxRetries > 0 {
		stepMsg = fmt.Sprintf("running failed, err=%s, retried=%d", stepErr, step.GetRetryCount())
	}
	step.SetEndTime(endTime).
		SetExecutionTime(start, endTime).
		SetStatus(types.TaskStatusFailure).
		SetMessage(stepMsg).
		SetLastUpdate(endTime)

	taskStartTime := s.task.GetStartTime()
	s.task.SetExecutionTime(taskStartTime, endTime).
		SetLastUpdate(endTime)

	// 任务超时, 整体结束
	if taskTimeOut {
		s.task.SetStatus(types.TaskStatusTimeout).SetMessage("task timeout")

		// callback
		if s.cbExecutor != nil {
			c := istep.NewContext(context.Background(), GetGlobalStorage(), s.task, step)
			s.cbExecutor.Callback(c, stepErr)
		}
		return nil
	}

	// last step failed and skipOnFailed is true, update task status to success
	if s.isLastStep(step) {
		if step.GetSkipOnFailed() {
			s.task.SetStatus(types.TaskStatusSuccess).SetMessage("task finished successfully")
		} else {
			s.task.SetStatus(types.TaskStatusFailure).SetMessage(fmt.Sprintf("step %s running failed", step.Name))
		}

		// callback
		if s.cbExecutor != nil {
			c := istep.NewContext(context.Background(), GetGlobalStorage(), s.task, step)
			s.cbExecutor.Callback(c, stepErr)
		}
		return nil
	}

	// 忽略错误
	if step.GetSkipOnFailed() {
		msg := fmt.Sprintf("step %s running failed, with skip on failed", step.Name)
		s.task.SetStatus(types.TaskStatusRunning).SetMessage(msg)
		return nil
	}

	// 重试流程中
	if step.MaxRetries > 0 {
		msg := fmt.Sprintf("step %s running failed, with retried=%d, maxRetries=%d",
			step.Name, step.GetRetryCount(), step.MaxRetries)

		if step.GetRetryCount() < step.MaxRetries {
			s.task.SetStatus(types.TaskStatusRunning).SetMessage(msg)
		} else {
			// 重试次数用完
			s.task.SetStatus(types.TaskStatusFailure).SetMessage(msg)
		}
		return nil
	}

	return nil
}

func (s *State) isLastStep(step *types.Step) bool {
	count := len(s.task.Steps)
	// 没有step默认返回false
	if count == 0 {
		return false
	}

	// 最后一步
	if step.GetName() != s.task.Steps[count-1].Name {
		return false
	}

	// 最后一步, 但是还有重试次数
	if step.MaxRetries > 0 && step.GetRetryCount() < step.MaxRetries {
		return false
	}

	return true
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

// GetCurrentStep get current step
func (s *State) GetCurrentStep() (*types.Step, bool) {
	return s.task.GetStep(s.currentStep)
}

// GetStep get step by stepName
func (s *State) GetStep(stepName string) (*types.Step, bool) {
	return s.task.GetStep(stepName)
}
