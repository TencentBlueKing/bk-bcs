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

// Package types pod types
package types

// ListTaskReq list task request
type ListTaskReq struct {
	ClusterID   string `json:"clusterID" in:"query=clusterID"`
	ProjectID   string `json:"projectID" in:"query=projectID"`
	Creator     string `json:"creator" in:"query=creator"`
	Updater     string `json:"updater" in:"query=updater"`
	TaskType    string `json:"taskType" in:"query=taskType"`
	Status      string `json:"status" in:"query=status"`
	NodeIP      string `json:"nodeIP" in:"query=nodeIP"`
	NodeGroupID string `json:"nodeGroupID" in:"query=nodeGroupID"`
	StartTime   uint64 `json:"startTime" in:"query=startTime"`
	EndTime     uint64 `json:"endTime" in:"query=endTime"`
	Limit       uint32 `json:"limit" in:"query=limit"`
	Page        uint32 `json:"page" in:"query=page"`
}

// GetTaskReq get task request
type GetTaskReq struct {
	TaskID string `json:"taskID" in:"path=taskID"`
}

// RetryTaskkReq retry task request
type RetryTaskkReq struct {
	TaskID  string `json:"taskID" in:"path=taskID"`
	Updater string `json:"updater"`
}

// SkipTaskReq skip task request
type SkipTaskReq struct {
	TaskID  string `json:"taskID" in:"path=taskID"`
	Updater string `json:"updater"`
}

// UpdateTaskReq update task request
type UpdateTaskReq struct {
	TaskID        string           `json:"taskID" in:"path=taskID"`
	Status        string           `json:"status"`
	Message       string           `json:"message"`
	End           string           `json:"end"`
	ExecutionTime uint32           `json:"executionTime"`
	CurrentStep   string           `json:"currentStep"`
	Steps         map[string]*Step `json:"steps"`
	Updater       string           `json:"updater"`
}

// Step step info
type Step struct {
	Name          string            `json:"name"`
	System        string            `json:"system"`
	Link          string            `json:"link"`
	Params        map[string]string `json:"params"`
	Retry         uint32            `json:"retry"`
	Start         string            `json:"start"`
	End           string            `json:"end"`
	ExecutionTime uint32            `json:"executionTime"`
	Status        string            `json:"status"`
	Message       string            `json:"message"`
	LastUpdate    string            `json:"lastUpdate"`
	TaskMethod    string            `json:"taskMethod"`
	TaskName      string            `json:"taskName"`
	SkipOnFailed  bool              `json:"skipOnFailed"`
	Translate     string            `json:"translate"`
	AllowSkip     bool              `json:"allowSkip"`
	MaxRetry      uint32            `json:"maxRetry"`
}

// ListTaskResp list task response
type ListTaskResp struct {
	Result []*Task `json:"result"`
	Total  uint32  `json:"total"`
}

// Task 任务详情
type Task struct {
	TaskID         string           `json:"taskID"`
	TaskType       string           `json:"taskType"`
	Status         string           `json:"status"`
	Message        string           `json:"message"`
	Start          string           `json:"start"`
	End            string           `json:"end"`
	ExecutionTime  uint32           `json:"executionTime"`
	CurrentStep    string           `json:"currentStep"`
	StepSequence   []string         `json:"stepSequence"`
	Steps          map[string]*Step `json:"steps"`
	ClusterID      string           `json:"clusterID"`
	ProjectID      string           `json:"projectID"`
	Creator        string           `json:"creator"`
	LastUpdate     string           `json:"lastUpdate"`
	Updater        string           `json:"updater"`
	ForceTerminate bool             `json:"forceTerminate"`
	TaskName       string           `json:"taskName"`
	NodeGroupID    string           `json:"nodeGroupID"`
	NodeIPList     []string         `json:"nodeIPList"`
}
