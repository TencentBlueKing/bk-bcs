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
	"errors"
	"time"
)

const (
	// TaskTimeFormat task time format, e.g. 2006-01-02T15:04:05Z07:00
	TaskTimeFormat = time.RFC3339

	// DefaultMaxExecuteTimeSeconds default max execute time for 1 hour
	DefaultMaxExecuteTimeSeconds = 3600
	// DefaultTaskMessage default message
	DefaultTaskMessage = "task initializing"
)

const (
	// TaskStatusInit INIT task status
	TaskStatusInit = "INITIALIZING"
	// TaskStatusRunning running task status
	TaskStatusRunning = "RUNNING"
	// TaskStatusSuccess task success
	TaskStatusSuccess = "SUCCESS"
	// TaskStatusFailure task failed
	TaskStatusFailure = "FAILURE"
	// TaskStatusTimeout task run timeout
	TaskStatusTimeout = "TIMEOUT"
	// TaskStatusRevoked task has been revoked
	TaskStatusRevoked = "REVOKED"
	// TaskStatusForceTerminate force task terminate
	TaskStatusForceTerminate = "FORCETERMINATE"
	// TaskStatusNotStarted force task terminate
	TaskStatusNotStarted = "NOTSTARTED"
)

var (
	// DefaultPayloadContent default json extras content
	DefaultPayloadContent = []byte("{}")
	// ErrNotImplemented not implemented error
	ErrNotImplemented = errors.New("not implemented")
)

// Task task definition
type Task struct {
	TaskIndex           string            `json:"taskIndex"`
	TaskIndexType       string            `json:"taskIndexType"`
	TaskID              string            `json:"taskId"`
	TaskType            string            `json:"taskType"`
	TaskName            string            `json:"taskName"`
	CurrentStep         string            `json:"currentStep"`
	Steps               []*Step           `json:"steps"`
	CallbackName        string            `json:"callbackName"`
	CommonParams        map[string]string `json:"commonParams"`
	CommonPayload       []byte            `json:"commonPayload"`
	Status              string            `json:"status"`
	Message             string            `json:"message"`
	ForceTerminate      bool              `json:"forceTerminate"`
	ExecutionTime       uint32            `json:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds"`
	Creator             string            `json:"creator"`
	Updater             string            `json:"updater"`
	Start               *time.Time        `json:"start"`
	End                 *time.Time        `json:"end"`
	LastUpdate          time.Time         `json:"lastUpdate"`
}

// Step step definition
type Step struct {
	Name                string            `json:"name"`
	Alias               string            `json:"alias"`
	Executor            string            `json:"executor"`
	Params              map[string]string `json:"params"`
	Payload             []byte            `json:"payload"`
	Status              string            `json:"status"`
	Message             string            `json:"message"`
	ETA                 *time.Time        `json:"eta"` // 延迟执行时间(Estimated Time of Arrival)
	SkipOnFailed        bool              `json:"skipOnFailed"`
	RetryCount          uint32            `json:"retryCount"`
	MaxRetries          uint32            `json:"maxRetries"`
	ExecutionTime       uint32            `json:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds"`
	Start               *time.Time        `json:"start"`
	End                 *time.Time        `json:"end"`
	LastUpdate          time.Time         `json:"lastUpdate"`
}

// TaskType taskType
type TaskType string // nolint

// String toString
func (tt TaskType) String() string {
	return string(tt)
}

// TaskName xxx
type TaskName string // nolint

// String xxx
func (tn TaskName) String() string {
	return string(tn)
}
