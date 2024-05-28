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
	"time"
)

// Step step definition
type Step struct {
	Name         string            `json:"name" bson:"name"`
	TaskName     string            `json:"taskname" bson:"taskname"`
	Params       map[string]string `json:"params" bson:"params"`
	Extras       string            `json:"extras" bson:"extras"`
	Status       string            `json:"status" bson:"status"`
	Message      string            `json:"message" bson:"message"`
	SkipOnFailed bool              `json:"skipOnFailed" bson:"skipOnFailed"`
	RetryCount   uint32            `json:"retryCount" bson:"retryCount"`

	Start               string `json:"start" bson:"start"`
	End                 string `json:"end" bson:"end"`
	ExecutionTime       uint32 `json:"executionTime" bson:"executionTime"`
	MaxExecutionSeconds uint32 `json:"maxExecutionSeconds" bson:"maxExecutionSeconds"`
	LastUpdate          string `json:"lastUpdate" bson:"lastUpdate"`
}

// NewStep return a new step by default params
func NewStep(stepName string, taskName string) *Step {
	return &Step{
		Name:                stepName,
		TaskName:            taskName,
		Params:              map[string]string{},
		Extras:              DefaultJsonExtrasContent,
		Status:              TaskStatusNotStarted,
		Message:             "",
		SkipOnFailed:        false,
		RetryCount:          0,
		MaxExecutionSeconds: DefaultMaxExecuteTimeSeconds,
	}
}

// GetStepName return step name
func (s *Step) GetStepName() string {
	return s.Name
}

// SetStepName set step name
func (s *Step) SetStepName(name string) *Step {
	s.Name = name
	return s
}

// GetTaskName return task name
func (s *Step) GetTaskName() string {
	return s.TaskName
}

// SetTaskName set task name
func (s *Step) SetTaskName(taskName string) *Step {
	s.TaskName = taskName
	return s
}

// GetParam return step param by key
func (s *Step) GetParam(key string) (string, bool) {
	if value, ok := s.Params[key]; ok {
		return value, true
	}
	return "", false
}

// AddParam set step param by key,value
func (s *Step) AddParam(key, value string) *Step {
	if s.Params == nil {
		s.Params = make(map[string]string, 0)
	}
	s.Params[key] = value
	return s
}

// GetParamsAll return all step params
func (s *Step) GetParamsAll() map[string]string {
	if s.Params == nil {
		s.Params = make(map[string]string, 0)
	}
	return s.Params
}

// SetParamMulti set step params by map
func (s *Step) SetParamMulti(params map[string]string) {
	if s.Params == nil {
		s.Params = make(map[string]string, 0)
	}
	for key, value := range params {
		s.Params[key] = value
	}
}

// SetNewParams replace all params by new params
func (s *Step) SetNewParams(params map[string]string) *Step {
	s.Params = params
	return s
}

// GetExtras return unmarshal step extras
func (s *Step) GetExtras(obj interface{}) error {
	if s.Extras == "" {
		s.Extras = DefaultJsonExtrasContent
	}
	return json.Unmarshal([]byte(s.Extras), obj)
}

// SetExtrasAll set step extras by json string
func (s *Step) SetExtrasAll(obj interface{}) error {
	result, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	s.Extras = string(result)
	return nil
}

// GetStatus return step status
func (s *Step) GetStatus() string {
	return s.Status
}

// SetStatus set status
func (s *Step) SetStatus(stat string) *Step {
	s.Status = stat
	return s
}

// GetMessage get step message
func (s *Step) GetMessage() string {
	if s.Message == "" {
		return ""
	}
	return s.Message
}

// SetMessage set step message
func (s *Step) SetMessage(msg string) *Step {
	s.Message = msg
	return s
}

// GetSkipOnFailed get step skipOnFailed
func (s *Step) GetSkipOnFailed() bool {
	return s.SkipOnFailed
}

// SetSkipOnFailed set step skipOnFailed
func (s *Step) SetSkipOnFailed(skipOnFailed bool) *Step {
	s.SkipOnFailed = skipOnFailed
	return s
}

// GetRetryCount get step retry count
func (s *Step) GetRetryCount() uint32 {
	return s.RetryCount
}

// AddRetryCount add step retry count
func (s *Step) AddRetryCount(count uint32) *Step {
	s.RetryCount += count
	return s
}

// GetStartTime get start time
func (s *Step) GetStartTime() (time.Time, error) {
	return time.Parse(TaskTimeFormat, s.Start)
}

// SetStartTime update start time
func (s *Step) SetStartTime(t time.Time) *Step {
	s.Start = t.Format(TaskTimeFormat)
	return s
}

// GetEndTime get end time
func (s *Step) GetEndTime() (time.Time, error) {
	return time.Parse(TaskTimeFormat, s.End)
}

// SetEndTime set end time
func (s *Step) SetEndTime(t time.Time) *Step {
	// set end time
	s.End = t.Format(TaskTimeFormat)
	return s
}

// GetExecutionTime set execution time
func (s *Step) GetExecutionTime() time.Duration {
	return time.Duration(time.Duration(s.ExecutionTime) * time.Millisecond)
}

// SetExecutionTime set execution time
func (s *Step) SetExecutionTime(start time.Time, end time.Time) *Step {
	s.ExecutionTime = uint32(end.Sub(start).Milliseconds())
	return s
}

// GetMaxExecutionSeconds get max execution seconds
func (s *Step) GetMaxExecutionSeconds() time.Duration {
	return time.Duration(time.Duration(s.MaxExecutionSeconds) * time.Second)
}

// SetMaxExecutionSeconds set max execution seconds
func (s *Step) SetMaxExecutionSeconds(maxExecutionSeconds time.Duration) *Step {
	s.MaxExecutionSeconds = uint32(maxExecutionSeconds.Seconds())
	return s
}

// GetLastUpdate get last update time
func (s *Step) GetLastUpdate() (time.Time, error) {
	return time.Parse(TaskTimeFormat, s.LastUpdate)
}

// SetLastUpdate set last update time
func (s *Step) SetLastUpdate(t time.Time) *Step {
	s.LastUpdate = t.Format(TaskTimeFormat)
	return s
}
