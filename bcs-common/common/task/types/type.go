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

// Task task definition
type Task struct {
	// index for task, client should set this field
	TaskIndex string `json:"taskIndex" bson:"taskIndex"`
	TaskID    string `json:"taskId" bson:"taskId"`
	TaskType  string `json:"taskType" bson:"taskType"`
	TaskName  string `json:"taskName" bson:"taskName"`
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

// Step step definition
type Step struct {
	Name   string            `json:"name" bson:"name"`
	Alias  string            `json:"alias" bson:"alias"`
	Params map[string]string `json:"params" bson:"params"`
	// step extras for string json, need client step to parse
	Extras              string `json:"extras" bson:"extras"`
	Status              string `json:"status" bson:"status"`
	Message             string `json:"message" bson:"message"`
	SkipOnFailed        bool   `json:"skipOnFailed" bson:"skipOnFailed"`
	RetryCount          uint32 `json:"retryCount" bson:"retryCount"`
	Start               string `json:"start" bson:"start"`
	End                 string `json:"end" bson:"end"`
	ExecutionTime       uint32 `json:"executionTime" bson:"executionTime"`
	MaxExecutionSeconds uint32 `json:"maxExecutionSeconds" bson:"maxExecutionSeconds"`
	LastUpdate          string `json:"lastUpdate" bson:"lastUpdate"`
}
