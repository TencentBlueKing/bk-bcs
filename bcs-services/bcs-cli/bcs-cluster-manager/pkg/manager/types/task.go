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
 *
 */

package types

// CreateTaskReq 创建任务request
type CreateTaskReq struct {
	TaskType     string          `json:"taskType"`
	Status       string          `json:"status"`
	StepSequence []string        `json:"stepSequence"`
	Steps        map[string]Step `json:"steps"`
	ClusterID    string          `json:"clusterID"`
	ProjectID    string          `json:"projectID"`
}

// RetryTaskReq 重试创建任务request
type RetryTaskReq struct {
	TaskID string `json:"taskID"`
}

// UpdateTaskReq 更新任务request
type UpdateTaskReq struct {
	TaskID        string          `json:"taskID"`
	Status        string          `json:"status"`
	Message       string          `json:"message"`
	End           string          `json:"end"`
	ExecutionTime uint32          `json:"executionTime"`
	CurrentStep   string          `json:"currentStep"`
	Steps         map[string]Step `json:"steps"`
}

// DeleteTaskReq 删除任务request
type DeleteTaskReq struct {
	TaskID string `json:"taskID"`
}

// GetTaskReq 查询任务信息request
type GetTaskReq struct {
	TaskID string `json:"taskID"`
}

// ListTaskReq 查询任务列表request
type ListTaskReq struct {
	ClusterID string `json:"clusterID"`
	ProjectID string `json:"projectID"`
}

// CreateTaskResp 创建任务response
type CreateTaskResp struct {
	TaskID string `json:"taskID"`
}

// RetryTaskResp 重试创建任务response
type RetryTaskResp struct {
	TaskID string `json:"taskID"`
}

// GetTaskResp 查询任务信息response
type GetTaskResp struct {
	Data Task `json:"data"`
}

// ListTaskResp 查询任务列表response
type ListTaskResp struct {
	Data []*Task `json:"data"`
}

// Step 任务步骤
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
}

// Task 任务信息
type Task struct {
	TaskID         string            `json:"taskID"`
	TaskType       string            `json:"taskType"`
	Status         string            `json:"status"`
	Message        string            `json:"message"`
	Start          string            `json:"start"`
	End            string            `json:"end"`
	ExecutionTime  uint32            `json:"executionTime"`
	CurrentStep    string            `json:"currentStep"`
	StepSequence   []string          `json:"stepSequence"`
	Steps          map[string]Step   `json:"steps"`
	ClusterID      string            `json:"clusterID"`
	ProjectID      string            `json:"projectID"`
	Creator        string            `json:"creator"`
	LastUpdate     string            `json:"lastUpdate"`
	Updater        string            `json:"updater"`
	ForceTerminate bool              `json:"forceTerminate"`
	CommonParams   map[string]string `json:"commonParams"`
	TaskName       string            `json:"taskName"`
	NodeIPList     []string          `json:"nodeIPList"`
	NodeGroupID    string            `json:"nodeGroupID"`
}

// TaskMgr 任务管理接口
type TaskMgr interface {
	// Create 创建任务
	Create(CreateTaskReq) (CreateTaskResp, error)
	// Retry 重试创建任务
	Retry(RetryTaskReq) (RetryTaskResp, error)
	// Update 更新任务
	Update(UpdateTaskReq) error
	// Delete 删除任务
	Delete(DeleteTaskReq) error
	// Get 查询任务
	Get(GetTaskReq) (GetTaskResp, error)
	// List 查询任务列表
	List(ListTaskReq) (ListTaskResp, error)
}
