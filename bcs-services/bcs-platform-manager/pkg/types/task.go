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

import "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

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

// Task 任务详情
type Task struct {
	TaskID         string                          `json:"taskID"`
	TaskType       string                          `json:"taskType"`
	Status         string                          `json:"status"`
	Message        string                          `json:"message"`
	Start          string                          `json:"start"`
	End            string                          `json:"end"`
	ExecutionTime  uint32                          `json:"executionTime"`
	CurrentStep    string                          `json:"currentStep"`
	StepSequence   []string                        `json:"stepSequence"`
	Steps          map[string]*clustermanager.Step `json:"steps"`
	ClusterID      string                          `json:"clusterID"`
	ProjectID      string                          `json:"projectID"`
	Creator        string                          `json:"creator"`
	LastUpdate     string                          `json:"lastUpdate"`
	Updater        string                          `json:"updater"`
	ForceTerminate bool                            `json:"forceTerminate"`
	TaskName       string                          `json:"taskName"`
	NodeGroupID    string                          `json:"nodeGroupID"`
}
