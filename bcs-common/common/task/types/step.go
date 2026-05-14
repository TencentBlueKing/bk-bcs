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
	MaxRetries          uint32
	SkipFailed          bool
	MaxExecutionSeconds uint32
}

// StepOption xxx
type StepOption func(opt *StepOptions)

// WithMaxRetries xxx
func WithMaxRetries(count uint32) StepOption {
	return func(opt *StepOptions) {
		opt.MaxRetries = count
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
func NewStep(name string, executor string, opts ...StepOption) *Step {
	defaultOptions := &StepOptions{MaxRetries: 0}
	for _, opt := range opts {
		opt(defaultOptions)
	}

	return &Step{
		Name:                name,
		Executor:            executor,
		Params:              map[string]string{},
		Payload:             DefaultPayloadContent,
		Status:              TaskStatusNotStarted,
		Message:             "",
		RetryCount:          0,
		SkipOnFailed:        defaultOptions.SkipFailed,
		MaxRetries:          defaultOptions.MaxRetries,
		MaxExecutionSeconds: defaultOptions.MaxExecutionSeconds,
	}
}

// GetName return task name
func (s *Step) GetName() string {
	return s.Name
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
	s.paramsLock.Lock()
	defer s.paramsLock.Unlock()
	if value, ok := s.Params[key]; ok {
		return value, true
	}
	return "", false
}

// AddParam set step param by key,value
func (s *Step) AddParam(key, value string) *Step {
	s.paramsLock.Lock()
	defer s.paramsLock.Unlock()
	if s.Params == nil {
		s.Params = make(map[string]string, 0)
	}
	s.Params[key] = value
	return s
}

// GetParams return all step params
func (s *Step) GetParams() map[string]string {
	s.paramsLock.Lock()
	defer s.paramsLock.Unlock()
	if s.Params == nil {
		s.Params = make(map[string]string, 0)
	}
	return s.Params
}

// SetParams set step params by map
func (s *Step) SetParams(params map[string]string) {
	s.paramsLock.Lock()
	defer s.paramsLock.Unlock()
	if s.Params == nil {
		s.Params = make(map[string]string, 0)
	}
	for key, value := range params {
		s.Params[key] = value
	}
}

// SetNewParams replace all params by new params
func (s *Step) SetNewParams(params map[string]string) *Step {
	s.paramsLock.Lock()
	defer s.paramsLock.Unlock()
	s.Params = params
	return s
}

// GetPayload unmarshal step payload to struct obj
func (s *Step) GetPayload(obj any) error {
	if len(s.Payload) == 0 {
		s.Payload = DefaultPayloadContent
	}
	return json.Unmarshal([]byte(s.Payload), obj)
}

// SetPayload marshal struct obj to step payload
func (s *Step) SetPayload(obj any) error {
	result, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	s.Payload = string(result)
	return nil
}

// GetStatus return step status
func (s *Step) GetStatus() string {
	return s.Status
}

// IsCompleted return step completed or not
func (s *Step) IsCompleted() bool {
	// 已经完成
	if s.Status == TaskStatusSuccess {
		return true
	}

	// 失败需要看重试次数
	if s.Status == TaskStatusFailure {
		// 还有重试次数
		if s.MaxRetries > 0 && s.RetryCount < s.MaxRetries {
			return false
		}
		return true
	}

	return false
}

// SetStatus set status
func (s *Step) SetStatus(stat string) *Step {
	s.Status = stat
	return s
}

// GetMessage get step message
func (s *Step) GetMessage() string {
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

// SetMaxTries set step max retry count
func (s *Step) SetMaxTries(count uint32) *Step {
	s.MaxRetries = count
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

// SetCountdown step eta with countdown(seconds)
func (s *Step) SetCountdown(c int) *Step {
	// 默认就是立即执行, 0值忽略
	if c <= 0 {
		return s
	}

	t := time.Now().Add(time.Duration(c) * time.Second)
	s.ETA = &t
	return s
}

// SetETA step estimated time of arrival
func (s *Step) SetETA(t time.Time) *Step {
	if t.Before(time.Now()) {
		return s
	}

	s.ETA = &t
	return s
}

// GetStartTime get start time
func (s *Step) GetStartTime() time.Time {
	return s.Start
}

// SetStartTime update start time
func (s *Step) SetStartTime(t time.Time) *Step {
	s.Start = t
	return s
}

// GetEndTime get end time
func (s *Step) GetEndTime() time.Time {
	return s.End
}

// SetEndTime set end time
func (s *Step) SetEndTime(t time.Time) *Step {
	// set end time
	s.End = t
	return s
}

// GetExecutionTime set execution time
func (s *Step) GetExecutionTime() time.Duration {
	return time.Duration(s.ExecutionTime) * time.Millisecond
}

// SetExecutionTime set execution time
func (s *Step) SetExecutionTime(start time.Time, end time.Time) *Step {
	s.ExecutionTime = uint32(end.Sub(start).Milliseconds())
	return s
}

// GetMaxExecution get max execution seconds
func (s *Step) GetMaxExecution() time.Duration {
	return time.Duration(s.MaxExecutionSeconds) * time.Second
}

// SetMaxExecution set max execution seconds
func (s *Step) SetMaxExecution(duration time.Duration) *Step {
	s.MaxExecutionSeconds = uint32(duration.Seconds())
	return s
}

// GetLastUpdate get last update time
func (s *Step) GetLastUpdate() time.Time {
	return s.LastUpdate
}

// SetLastUpdate set last update time
func (s *Step) SetLastUpdate(t time.Time) *Step {
	s.LastUpdate = t
	return s
}
