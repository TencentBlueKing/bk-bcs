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

// Package types for task
package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TaskOptions xxx
type TaskOptions struct {
	CallbackName        string
	MaxExecutionSeconds uint32
}

// TaskOption xxx
type TaskOption func(opt *TaskOptions)

// WithTaskCallback xxx
func WithTaskCallback(callbackName string) TaskOption {
	return func(opt *TaskOptions) {
		opt.CallbackName = callbackName
	}
}

// WithTaskMaxExecutionSeconds xxx
func WithTaskMaxExecutionSeconds(timeout uint32) TaskOption {
	return func(opt *TaskOptions) {
		opt.MaxExecutionSeconds = timeout
	}
}

// TaskInfo task basic info definition
type TaskInfo struct {
	TaskType      string
	TaskName      string
	TaskIndex     string // TaskIndex for resource index
	TaskIndexType string
	Creator       string
}

// NewTask create new task by default
func NewTask(o TaskInfo, opts ...TaskOption) *Task {
	defaultOptions := &TaskOptions{CallbackName: "", MaxExecutionSeconds: 0}
	for _, opt := range opts {
		opt(defaultOptions)
	}

	now := time.Now()
	return &Task{
		TaskID:              uuid.NewString(),
		TaskType:            o.TaskType,
		TaskName:            o.TaskName,
		TaskIndex:           o.TaskIndex,
		TaskIndexType:       o.TaskIndexType,
		Status:              TaskStatusInit,
		ForceTerminate:      false,
		Steps:               make([]*Step, 0),
		Creator:             o.Creator,
		Updater:             o.Creator,
		LastUpdate:          now,
		CommonParams:        make(map[string]string, 0),
		CommonPayload:       DefaultPayloadContent,
		CallbackName:        defaultOptions.CallbackName,
		Message:             DefaultTaskMessage,
		MaxExecutionSeconds: defaultOptions.MaxExecutionSeconds,
	}
}

// GetTaskID get task id
func (t *Task) GetTaskID() string {
	return t.TaskID
}

// GetTaskType get task type
func (t *Task) GetTaskType() string {
	return t.TaskType
}

// GetTaskName get task name
func (t *Task) GetTaskName() string {
	return t.TaskName
}

// GetStep get step by name
func (t *Task) GetStep(stepName string) (*Step, bool) {
	for _, step := range t.Steps {
		if step.Name == stepName {
			return step, true
		}
	}
	return nil, false
}

// AddStep add step to task
func (t *Task) AddStep(step *Step) *Task {
	if step == nil {
		t.Steps = make([]*Step, 0)
	}

	t.Steps = append(t.Steps, step)
	return t
}

// GetCommonParam get common params
func (t *Task) GetCommonParam(key string) (string, bool) {
	if t.CommonParams == nil {
		t.CommonParams = make(map[string]string, 0)
		return "", false
	}
	if value, ok := t.CommonParams[key]; ok {
		return value, true
	}
	return "", false
}

// AddCommonParam add common params
func (t *Task) AddCommonParam(k, v string) *Task {
	if t.CommonParams == nil {
		t.CommonParams = make(map[string]string, 0)
	}
	t.CommonParams[k] = v
	return t
}

// GetCallback set callback function name
func (t *Task) GetCallback() string {
	return t.CallbackName
}

// SetCallback set callback function name
func (t *Task) SetCallback(callBackName string) *Task {
	t.CallbackName = callBackName
	return t
}

// GetCommonPayload get extra json
func (t *Task) GetCommonPayload(obj interface{}) error {
	if len(t.CommonPayload) == 0 {
		t.CommonPayload = DefaultPayloadContent
	}
	return json.Unmarshal(t.CommonPayload, obj)
}

// SetCommonPayload set extra json
func (t *Task) SetCommonPayload(obj interface{}) error {
	result, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	t.CommonPayload = result
	return nil
}

// GetStatus get status
func (t *Task) GetStatus() string {
	return t.Status
}

// SetStatus set status
func (t *Task) SetStatus(status string) *Task {
	t.Status = status
	return t
}

// GetMessage set message
func (t *Task) GetMessage() string {
	return t.Message
}

// SetMessage set message
func (t *Task) SetMessage(msg string) *Task {
	t.Message = msg
	return t
}

// GetForceTerminate get force terminate
func (t *Task) GetForceTerminate() bool {
	return t.ForceTerminate
}

// SetForceTerminate set force terminate
func (t *Task) SetForceTerminate(f bool) *Task {
	t.ForceTerminate = f
	return t
}

// GetStartTime get start time
func (t *Task) GetStartTime() time.Time {
	return t.Start
}

// SetStartTime set start time
func (t *Task) SetStartTime(time time.Time) *Task {
	t.Start = time
	return t
}

// GetEndTime get end time
func (t *Task) GetEndTime() time.Time {
	return t.End
}

// SetEndTime set end time
func (t *Task) SetEndTime(time time.Time) *Task {
	t.End = time
	return t
}

// GetExecutionTime get execution time
func (t *Task) GetExecutionTime() time.Duration {
	return time.Duration(t.ExecutionTime)
}

// SetExecutionTime set execution time
func (t *Task) SetExecutionTime(start time.Time, end time.Time) *Task {
	t.ExecutionTime = uint32(end.Sub(start).Milliseconds())
	return t
}

// GetMaxExecution get max execution seconds
func (t *Task) GetMaxExecution() time.Duration {
	return time.Duration(t.MaxExecutionSeconds) * time.Second
}

// SetMaxExecution set max execution seconds
func (t *Task) SetMaxExecution(duration time.Duration) *Task {
	t.MaxExecutionSeconds = uint32(duration.Seconds())
	return t
}

// GetCreator get creator
func (t *Task) GetCreator() string {
	return t.Creator
}

// SetCreator set creator
func (t *Task) SetCreator(creator string) *Task {
	t.Creator = creator
	return t
}

// GetUpdater get updater
func (t *Task) GetUpdater() string {
	return t.Updater
}

// SetUpdater set updater
func (t *Task) SetUpdater(updater string) *Task {
	t.Updater = updater
	return t
}

// GetLastUpdate get last update time
func (t *Task) GetLastUpdate() (time.Time, error) {
	return t.LastUpdate, nil
}

// SetLastUpdate set last update time
func (t *Task) SetLastUpdate(lastUpdate time.Time) *Task {
	t.LastUpdate = lastUpdate
	return t
}

// GetCurrentStep get current step
func (t *Task) GetCurrentStep() string {
	return t.CurrentStep
}

// SetCurrentStep set current step
func (t *Task) SetCurrentStep(stepName string) *Task {
	t.CurrentStep = stepName
	return t
}

// GetStepParam get step params
func (t *Task) GetStepParam(stepName, key string) (string, bool) {
	step, ok := t.GetStep(stepName)
	if !ok {
		return "", false
	}
	return step.GetParam(key)
}

// AddStepParams add step params
func (t *Task) AddStepParams(stepName string, k, v string) error {
	step, ok := t.GetStep(stepName)
	if !ok {
		return fmt.Errorf("step %s not exist", stepName)
	}
	step.AddParam(k, v)
	return nil
}

// AddStepParamsBatch add step params batch
func (t *Task) AddStepParamsBatch(stepName string, params map[string]string) error {
	step, ok := t.GetStep(stepName)
	if !ok {
		return fmt.Errorf("step %s not exist", stepName)
	}

	for k, v := range params {
		step.AddParam(k, v)
	}
	return nil
}

// Validate 校验 task
func (t *Task) Validate() error {
	if t.TaskName == "" {
		return fmt.Errorf("task name is required")
	}

	if len(t.Steps) == 0 {
		return fmt.Errorf("task steps empty")
	}

	uniq := map[string]struct{}{}
	for _, s := range t.Steps {
		if s.Name == "" {
			return fmt.Errorf("step name is required")
		}

		if _, ok := uniq[s.Name]; ok {
			return fmt.Errorf("step name %s is not unique", s.Name)
		}
		uniq[s.Name] = struct{}{}
	}

	return nil
}
