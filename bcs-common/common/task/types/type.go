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
	// DefaultJsonExtrasContent default json extras content
	DefaultJsonExtrasContent = "{}"
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
	// TaskStatusForceTerminate force task terminate
	TaskStatusForceTerminate = "FORCETERMINATE"
	// TaskStatusNotStarted force task terminate
	TaskStatusNotStarted = "NOTSTARTED"
)

var (
	// ErrNotImplemented not implemented error
	ErrNotImplemented = errors.New("not implemented")
)

// Task task definition
type Task struct {
	TaskID              string            `json:"taskId"`
	TaskType            string            `json:"taskType"`
	TaskName            string            `json:"taskName"`
	CurrentStep         string            `json:"currentStep"`
	Steps               []*Step           `json:"steps"`
	CallBackFuncName    string            `json:"callBackFuncName"`
	CommonParams        map[string]string `json:"commonParams"`
	ExtraJson           string            `json:"extraJson"`
	Status              string            `json:"status"`
	Message             string            `json:"message"`
	ForceTerminate      bool              `json:"forceTerminate"`
	Start               time.Time         `json:"start"`
	End                 time.Time         `json:"end"`
	ExecutionTime       uint32            `json:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds"`
	Creator             string            `json:"creator"`
	LastUpdate          time.Time         `json:"lastUpdate"`
	Updater             string            `json:"updater"`
}

// Step step definition
type Step struct {
	Name                string            `json:"name"`
	Alias               string            `json:"alias"`
	Params              map[string]string `json:"params"`
	Extras              string            `json:"extras"`
	Status              string            `json:"status"`
	Message             string            `json:"message"`
	SkipOnFailed        bool              `json:"skipOnFailed"`
	RetryCount          uint32            `json:"retryCount"`
	Start               time.Time         `json:"start"`
	End                 time.Time         `json:"end"`
	ExecutionTime       uint32            `json:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds"`
	LastUpdate          time.Time         `json:"lastUpdate"`
}
