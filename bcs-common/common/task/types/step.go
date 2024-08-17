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

// StepOptions xxx
type StepOptions struct {
	Retry               uint32
	SkipFailed          bool
	MaxExecutionSeconds uint32
}

// StepOption xxx
type StepOption func(opt *StepOptions)

// WithStepRetry xxx
func WithStepRetry(retry uint32) StepOption {
	return func(opt *StepOptions) {
		opt.Retry = retry
	}
}

// WithStepSkipFailed xxx
func WithStepSkipFailed(skip bool) StepOption {
	return func(opt *StepOptions) {
		opt.SkipFailed = skip
	}
}

// WithMaxExecutionSeconds xxx
func WithMaxExecutionSeconds(execSecs uint32) StepOption {
	return func(opt *StepOptions) {
		opt.MaxExecutionSeconds = execSecs
	}
}

// NewStep return a new step by default params
func NewStep(name string, alias string, opts ...StepOption) *Step {
	defaultOptions := &StepOptions{Retry: 0}
	for _, opt := range opts {
		opt(defaultOptions)
	}

	return &Step{
		Name:                name,
		Alias:               alias,
		Params:              map[string]string{},
		Payload:             DefaultPayloadContent,
		Status:              TaskStatusNotStarted,
		Message:             "",
		SkipOnFailed:        defaultOptions.SkipFailed,
		RetryCount:          defaultOptions.Retry,
		Start:               time.Now(),
		MaxExecutionSeconds: defaultOptions.MaxExecutionSeconds,
	}
}

// GetName return task name
func (s *Step) GetName() string {
	return s.Name
}

// SetName set task method
func (s *Step) SetName(name string) *Step {
	s.Name = name
	return s
}

// GetAlias return task alias
func (s *Step) GetAlias() string {
	return s.Alias
}

// SetAlias set task alias
func (s *Step) SetAlias(alias string) *Step {
	s.Alias = alias
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

// GetParams return all step params
func (s *Step) GetParams() map[string]string {
	if s.Params == nil {
		s.Params = make(map[string]string, 0)
	}
	return s.Params
}

// SetParams set step params by map
func (s *Step) SetParams(params map[string]string) {
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

// GetPayload return unmarshal step extras
func (s *Step) GetPayload(obj interface{}) error {
	if len(s.Payload) == 0 {
		s.Payload = DefaultPayloadContent
	}
	return json.Unmarshal(s.Payload, obj)
}

// SetPayload set step extras by json string
func (s *Step) SetPayload(obj interface{}) error {
	result, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	s.Payload = result
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
	return s.Start, nil
}

// SetStartTime update start time
func (s *Step) SetStartTime(t time.Time) *Step {
	s.Start = t
	return s
}

// GetEndTime get end time
func (s *Step) GetEndTime() (time.Time, error) {
	return s.End, nil
}

// SetEndTime set end time
func (s *Step) SetEndTime(t time.Time) *Step {
	// set end time
	s.End = t
	return s
}

// GetExecutionTime set execution time
func (s *Step) GetExecutionTime() time.Duration {
	return time.Duration(s.ExecutionTime)
}

// SetExecutionTime set execution time
func (s *Step) SetExecutionTime(start time.Time, end time.Time) *Step {
	s.ExecutionTime = uint32(end.Sub(start).Milliseconds())
	return s
}

// GetMaxExecutionSeconds get max execution seconds
func (s *Step) GetMaxExecutionSeconds() time.Duration {
	return time.Duration(s.MaxExecutionSeconds) * time.Second
}

// SetMaxExecutionSeconds set max execution seconds
func (s *Step) SetMaxExecutionSeconds(maxExecutionSeconds time.Duration) *Step {
	s.MaxExecutionSeconds = uint32(maxExecutionSeconds.Seconds())
	return s
}

// GetLastUpdate get last update time
func (s *Step) GetLastUpdate() (time.Time, error) {
	return s.LastUpdate, nil
}

// SetLastUpdate set last update time
func (s *Step) SetLastUpdate(t time.Time) *Step {
	s.LastUpdate = t
	return s
}
