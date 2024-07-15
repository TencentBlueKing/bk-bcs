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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// getTaskStateAndCurrentStep get task state and current step
func getTaskStateAndCurrentStep(taskId, stepName string,
	callBackFuncs map[string]CallbackInterface) (*State, *types.Step, error) {
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
		return nil, nil, fmt.Errorf("task %s step %s is not ready, %s", taskId, stepName, err.Error())
	}

	if step == nil {
		// step successful and skip
		blog.Infof("task %s step %s already execute successful", taskId, stepName)
		return state, nil, nil
	}

	// inject call back func
	if state.task.GetCallback() != "" && len(callBackFuncs) > 0 {
		if callback, ok := callBackFuncs[state.task.GetCallback()]; ok {
			state.callBack = callback.Callback
		}
	}

	return state, step, nil
}

// State is a struct for task state
type State struct {
	task        *types.Task
	currentStep string
	callBack    func(isSuccess bool, task *types.Task)
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
	switch s.task.GetStatus() {
	case types.TaskStatusRunning, types.TaskStatusInit:
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

	// not first time to execute current step
	if stepName == s.task.GetCurrentStep() {
		if curStep.GetStatus() == types.TaskStatusFailure {
			curStep.AddRetryCount(1)
		}

		nowTime := time.Now()
		curStep = curStep.SetStartTime(nowTime).
			SetStatus(types.TaskStatusRunning).
			SetMessage("step ready to run").
			SetLastUpdate(nowTime)

		// update Task in storage
		if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
			return nil, err
		}
		return curStep, nil
	}

	// first time to execute step
	for _, name := range s.task.StepSequence {
		step, ok := s.task.GetStep(name)
		if !ok {
			return nil, fmt.Errorf("step %s is not exist", stepName)
		}

		// find current step
		if name == stepName {
			// step already success
			if step.GetStatus() == types.TaskStatusSuccess {
				return nil, fmt.Errorf("task %s step %s already success", s.task.GetTaskID(), stepName)
			}
			// set current step
			nowTime := time.Now()
			s.task.SetCurrentStep(stepName)
			step = step.SetStartTime(nowTime).
				SetStatus(types.TaskStatusRunning).
				SetMessage("step ready to run").
				SetLastUpdate(nowTime)

			// update Task in storage
			if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
				return nil, fmt.Errorf("update task %s step %s status error", s.task.GetTaskID(), stepName)
			}
			return step, nil
		}

		// skip step if step allow skipOnFailed
		if step.SkipOnFailed {
			continue
		}
		// previous step execute failure
		if step.GetStatus() != types.TaskStatusSuccess {
			break
		}
	}
	// previous step execute failure
	return nil, fmt.Errorf("step %s is not ready", stepName)
}

// updateStepSuccess update step status to success
func (s *State) updateStepSuccess(start time.Time, stepName string) error {
	step, ok := s.task.GetStep(stepName)
	if !ok {
		return fmt.Errorf("step %s is not exist", stepName)
	}
	endTime := time.Now()

	step.SetStartTime(start).
		SetEndTime(endTime).
		SetExecutionTime(start, endTime).
		SetStatus(types.TaskStatusSuccess).
		SetMessage(fmt.Sprintf("step %s running successfully", step.Name)).
		SetLastUpdate(endTime)

	s.task.SetStatus(types.TaskStatusRunning).
		SetMessage(fmt.Sprintf("step %s running successfully", step.Name)).
		SetLastUpdate(endTime)

	// last step
	if s.isLastStep(stepName) {
		taskStartTime, err := s.task.GetStartTime()
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("get task %s start time error", s.task.GetTaskID()))
		}
		s.task.SetEndTime(endTime).
			SetExecutionTime(taskStartTime, endTime).
			SetStatus(types.TaskStatusSuccess).
			SetMessage("task finished successfully")

		// callback
		if s.callBack != nil {
			s.callBack(true, s.task)
		}
	}

	// update Task in storage
	if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
		return fmt.Errorf("update task %s status error", s.task.GetTaskID())
	}

	return nil
}

// updateStepFailure update step status to failure
func (s *State) updateStepFailure(start time.Time, name string, stepErr error, taskTimeOut bool) error {
	step, ok := s.task.GetStep(name)
	if !ok {
		return fmt.Errorf("step %s is not exist", name)
	}
	endTime := time.Now()

	step.SetStartTime(start).
		SetEndTime(endTime).
		SetExecutionTime(start, endTime).
		SetStatus(types.TaskStatusFailure).
		SetMessage(fmt.Sprintf("running failed, %s", stepErr.Error())).
		SetLastUpdate(endTime)

	// if step SkipOnFailed, update task status to running
	if !taskTimeOut && step.GetSkipOnFailed() {
		// skip, set task running or success
		s.task.SetStatus(types.TaskStatusRunning).
			SetMessage(fmt.Sprintf("step %s running failed", step.Name)).
			SetLastUpdate(endTime)

		// last step failed and skipOnFailed is true, update task status to success
		if s.isLastStep(name) {
			// last step
			taskStartTime, err := s.task.GetStartTime()
			if err != nil {
				return fmt.Errorf(fmt.Sprintf("get task %s start time error", s.task.GetTaskID()))
			}
			s.task.SetStatus(types.TaskStatusSuccess).
				SetMessage("task finished successfully").
				SetLastUpdate(endTime).
				SetEndTime(endTime).
				SetExecutionTime(taskStartTime, endTime)

			// callback
			if s.callBack != nil {
				s.callBack(true, s.task)
			}
		}
	} else {
		// not skip, set task failed
		s.task.SetStatus(types.TaskStatusFailure).
			SetMessage(fmt.Sprintf("step %s running failed", step.Name)).
			SetLastUpdate(endTime).
			SetEndTime(endTime).
			SetExecutionTime(start, endTime)

		if taskTimeOut {
			s.task.SetStatus(types.TaskStatusTimeout).SetMessage("task timeout")
		}

		// callback
		if s.callBack != nil {
			s.callBack(false, s.task)
		}
	}

	// update Task in storage
	if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
		return fmt.Errorf("update task %s status error", s.task.GetTaskID())
	}

	return nil
}

func (s *State) isLastStep(stepName string) bool {
	return stepName == s.task.StepSequence[len(s.task.StepSequence)-1]
}

// GetCommonParams get common params by key
func (s *State) GetCommonParams(key string) (string, bool) {
	return s.task.GetCommonParams(key)
}

// AddCommonParams add common params
func (s *State) AddCommonParams(key, value string) *State {
	s.task.AddCommonParams(key, value)
	return s
}

// GetExtraParams get extra params by obj
func (s *State) GetExtraParams(obj interface{}) error {
	return s.task.GetExtra(obj)
}

// SetExtraAll set extra params by obj
func (s *State) SetExtraAll(obj interface{}) error {
	return s.task.SetExtraAll(obj)
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
