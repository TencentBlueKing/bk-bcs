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

// Task task definition
type Task struct {
	// index for task, client should set this field
	Index    string `json:"index" bson:"index"`
	TaskID   string `json:"taskId" bson:"taskId"`
	TaskType string `json:"taskType" bson:"taskType"`
	TaskName string `json:"taskName" bson:"taskName"`
	// steps and params
	CurrentStep      string            `json:"currentStep" bson:"currentStep"`
	StepSequence     []string          `json:"stepSequence" bson:"stepSequence"`
	Steps            map[string]*Step  `json:"steps" bson:"steps"`
	CallBackFuncName string            `json:"callBackFuncName" bson:"callBackFuncName"`
	CommonParams     map[string]string `json:"commonParams" bson:"commonParams"`
	ExtraJson        string            `json:"extraJson" bson:"extraJson"`

	Status              string `json:"status" bson:"status"`
	Message             string `json:"message" bson:"message"`
	ForceTerminate      bool   `json:"forceTerminate" bson:"forceTerminate"`
	Start               string `json:"start" bson:"start"`
	End                 string `json:"end" bson:"end"`
	ExecutionTime       uint32 `json:"executionTime" bson:"executionTime"`
	MaxExecutionSeconds uint32 `json:"maxExecutionSeconds" bson:"maxExecutionSeconds"`
	Creator             string `json:"creator" bson:"creator"`
	LastUpdate          string `json:"lastUpdate" bson:"lastUpdate"`
	Updater             string `json:"updater" bson:"updater"`
}

// TaskOptions task options definition
type TaskOptions struct {
	TaskIndex        string
	TaskType         string
	TaskName         string
	Creator          string
	CallBackFuncName string
}

// NewTask create new task by default
func NewTask(o *TaskOptions) *Task {
	nowTime := time.Now().Format(TaskTimeFormat)
	return &Task{
		Index:            o.TaskIndex,
		TaskID:           uuid.NewString(),
		TaskType:         o.TaskType,
		TaskName:         o.TaskName,
		Status:           TaskStatusInit,
		ForceTerminate:   false,
		Start:            nowTime,
		Steps:            make(map[string]*Step, 0),
		StepSequence:     make([]string, 0),
		Creator:          o.Creator,
		Updater:          o.Creator,
		LastUpdate:       nowTime,
		CommonParams:     make(map[string]string, 0),
		ExtraJson:        DefaultJsonExtrasContent,
		CallBackFuncName: o.CallBackFuncName,
	}
}

// GetTaskID get task id
func (t *Task) GetTaskID() string {
	return t.TaskID
}

// GetIndex get task id
func (t *Task) GetIndex() string {
	return t.Index
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
	if _, ok := t.Steps[stepName]; !ok {
		return nil, false
	}
	return t.Steps[stepName], true
}

// AddStep add step to task
func (t *Task) AddStep(step *Step) *Task {
	if step == nil {
		return t
	}

	if t.StepSequence == nil {
		t.StepSequence = make([]string, 0)
	}
	t.StepSequence = append(t.StepSequence, step.GetStepName())
	t.Steps[step.GetStepName()] = step
	return t
}

// GetCommonParams get common params
func (t *Task) GetCommonParams(key string) (string, bool) {
	if t.CommonParams == nil {
		t.CommonParams = make(map[string]string, 0)
		return "", false
	}
	if value, ok := t.CommonParams[key]; ok {
		return value, true
	}
	return "", false
}

// AddCommonParams add common params
func (t *Task) AddCommonParams(k, v string) *Task {
	if t.CommonParams == nil {
		t.CommonParams = make(map[string]string, 0)
	}
	t.CommonParams[k] = v
	return t
}

// GetCallback set callback function name
func (t *Task) GetCallback() string {
	return t.CallBackFuncName
}

// SetCallback set callback function name
func (t *Task) SetCallback(callBackFuncName string) *Task {
	t.CallBackFuncName = callBackFuncName
	return t
}

// GetExtra get extra json
func (t *Task) GetExtra(obj interface{}) error {
	if t.ExtraJson == "" {
		t.ExtraJson = DefaultJsonExtrasContent
	}
	return json.Unmarshal([]byte(t.ExtraJson), obj)
}

// SetExtraAll set extra json
func (t *Task) SetExtraAll(obj interface{}) error {
	result, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	t.ExtraJson = string(result)
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
func (t *Task) GetMessage(msg string) string {
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
func (t *Task) GetStartTime() (time.Time, error) {
	return time.Parse(TaskTimeFormat, t.Start)
}

// SetStartTime set start time
func (t *Task) SetStartTime(time time.Time) *Task {
	t.Start = time.Format(TaskTimeFormat)
	return t
}

// GetEndTime get end time
func (t *Task) GetEndTime() (time.Time, error) {
	return time.Parse(TaskTimeFormat, t.End)
}

// SetEndTime set end time
func (t *Task) SetEndTime(time time.Time) *Task {
	t.End = time.Format(TaskTimeFormat)
	return t
}

// GetExecutionTime get execution time
func (t *Task) GetExecutionTime() time.Duration {
	return time.Duration(time.Duration(t.ExecutionTime) * time.Millisecond)
}

// SetExecutionTime set execution time
func (t *Task) SetExecutionTime(start time.Time, end time.Time) *Task {
	t.ExecutionTime = uint32(end.Sub(start).Milliseconds())
	return t
}

// GetMaxExecutionSeconds get max execution seconds
func (t *Task) GetMaxExecutionSeconds() time.Duration {
	return time.Duration(time.Duration(t.MaxExecutionSeconds) * time.Second)
}

// SetMaxExecutionSeconds set max execution seconds
func (t *Task) SetMaxExecutionSeconds(maxExecutionSeconds time.Duration) *Task {
	t.MaxExecutionSeconds = uint32(maxExecutionSeconds.Seconds())
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
	return time.Parse(TaskTimeFormat, t.LastUpdate)
}

// SetLastUpdate set last update time
func (t *Task) SetLastUpdate(lastUpdate time.Time) *Task {
	t.LastUpdate = lastUpdate.Format(TaskTimeFormat)
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
	if _, ok := t.Steps[stepName]; !ok {
		return fmt.Errorf("step %s not exist", stepName)
	}
	for k, v := range params {
		t.Steps[stepName].AddParam(k, v)
	}
	return nil
}
