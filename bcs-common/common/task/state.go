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

package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RichardKnop/machinery/v2/log"

	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// defaultIgnoredTaskMessage 被忽略的步骤兜底描述
const defaultIgnoredTaskMessage = "task finished with ignored steps"

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
		return nil, fmt.Errorf("task %s step %s is not ready, err %s", taskId, stepName, err.Error())
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
	return types.IsTaskTerminal(s.task.GetStatus())
}

// succeededTerminal 返回任务成功收尾时应写入的终态与描述。
// 任意步骤被标记为幂等忽略时, 任务终态收敛为 IGNORED 而非 SUCCESS,
// 描述取各被忽略步骤通过 istep.Context.MarkIgnored 传入的原因, 多个原因用 "; " 连接,
// 调用方可直接把 task.Message 展示给用户。
func (s *State) succeededTerminal() (string, string) {
	ignored := false
	reasons := make([]string, 0, len(s.task.Steps))
	for _, step := range s.task.Steps {
		if step.GetStatus() != types.TaskStatusIgnored {
			continue
		}
		ignored = true
		if message := step.GetMessage(); message != "" {
			reasons = append(reasons, message)
		}
	}

	switch {
	case !ignored:
		return types.TaskStatusSuccess, "task finished successfully"
	case len(reasons) == 0:
		return types.TaskStatusIgnored, defaultIgnoredTaskMessage
	default:
		return types.TaskStatusIgnored, strings.Join(reasons, "; ")
	}
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
		if types.IsTaskSucceeded(curStep.GetStatus()) {
			if s.isLastStep(curStep) {
				status, message := s.succeededTerminal()
				s.task.SetEndTime(nowTime).
					SetExecutionTime(taskStartTime, nowTime).
					SetStatus(status).
					SetMessage(message)
			}
			// step is success, skip
			return nil, nil
		}

		// task failed
		failMsg := fmt.Sprintf("step %s running failed", curStep.Name)
		if s.isLastStep(curStep) {
			if curStep.GetSkipOnFailed() {
				status, message := s.succeededTerminal()
				s.task.SetEndTime(nowTime).
					SetExecutionTime(taskStartTime, nowTime).
					SetStatus(status).
					SetMessage(message)
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

// tryCallback
func (s *State) tryCallback(stepErr error) {
	// callback
	if s.cbExecutor != nil {
		c := istep.NewContext(context.Background(), GetGlobalStorage(), s.task, s.step)
		if err := s.cbExecutor.Callback(c, stepErr); err != nil {
			// 如果callback失败，则把整个任务置为失败
			errMsg := fmt.Sprintf("callbackName[%s] of step[%s] failed, %s",
				s.task.CallbackName, s.step.GetName(), err.Error())
			s.task.SetCallbackResult(types.CallbackResultFailure).
				SetCallbackMessage(errMsg)
			return
		}
		s.task.SetCallbackResult(types.CallbackResultSuccess)
		return
	}
}

// updateStepSuccess update step status to success
func (s *State) updateStepSuccess(start time.Time) {
	s.settleStepSucceeded(start, types.TaskStatusSuccess,
		fmt.Sprintf("step %s running successfully", s.step.Name))
}

// updateStepIgnored 步骤被业务侧标记为幂等忽略, 按成功收尾但状态记为 IGNORED
func (s *State) updateStepIgnored(start time.Time, message string) {
	if message == "" {
		message = fmt.Sprintf("step %s ignored", s.step.Name)
	}
	s.settleStepSucceeded(start, types.TaskStatusIgnored, message)
}

// settleStepSucceeded 以视同成功的终态收尾当前步骤, 并在最后一步时收敛任务终态
func (s *State) settleStepSucceeded(start time.Time, stepStatus, message string) {
	endTime := time.Now()
	s.step.SetEndTime(endTime).
		SetExecutionTime(start, endTime).
		SetStatus(stepStatus).
		SetMessage(message).
		SetLastUpdate(endTime)

	taskStartTime := s.task.GetStartTime()
	s.task.SetStatus(types.TaskStatusRunning).
		SetExecutionTime(taskStartTime, endTime).
		SetMessage(message).
		SetLastUpdate(endTime)

	if s.isLastStep(s.step) {
		status, taskMsg := s.succeededTerminal()
		s.task.SetEndTime(endTime).
			SetStatus(status).
			SetMessage(taskMsg)
	}
}

func (s *State) saveTaskState() error {
	// update Task in storage
	if err := GetGlobalStorage().UpdateTask(context.Background(), s.task); err != nil {
		return fmt.Errorf(
			"task %s update step %s to failure failed: %s", s.task.TaskID, s.step.GetName(), err.Error())
	}
	return nil
}

// updateStepFailure update step status to failure
func (s *State) updateStepFailure(start time.Time, stepErr error, taskStatus *taskEndStatus) {
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
		return
	}

	// last step failed and skipOnFailed is true, update task status to success
	if s.isLastStep(s.step) {
		if s.step.GetSkipOnFailed() {
			status, message := s.succeededTerminal()
			s.task.SetEndTime(endTime).
				SetStatus(status).
				SetMessage(message)
		} else {
			s.task.SetEndTime(endTime).
				SetStatus(types.TaskStatusFailure).
				SetMessage(taskFailMsg)
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

// GetTask get task
func (s *State) GetTask() *types.Task {
	return s.task
}
