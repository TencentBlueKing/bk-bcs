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

// Package task task operate
package task

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	clustermgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/clustermanager"
)

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

// GetTask 获取任务详情
// @Summary 任务详情
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /task/{taskID} [get]
func GetTask(ctx context.Context, req *GetTaskReq) (*Task, error) {
	task, err := clustermgr.GetTask(ctx, &clustermanager.GetTaskRequest{
		TaskID: req.TaskID,
	})
	if err != nil {
		return nil, err
	}

	return &Task{
		TaskID:         task.TaskID,
		TaskType:       task.TaskType,
		Status:         task.Status,
		Message:        task.Message,
		Start:          task.Start,
		End:            task.End,
		ExecutionTime:  task.ExecutionTime,
		CurrentStep:    task.CurrentStep,
		StepSequence:   task.StepSequence,
		Steps:          task.Steps,
		ClusterID:      task.ClusterID,
		ProjectID:      task.ProjectID,
		Creator:        task.Creator,
		LastUpdate:     task.LastUpdate,
		Updater:        task.Updater,
		ForceTerminate: task.ForceTerminate,
		TaskName:       task.TaskName,
		NodeGroupID:    task.NodeGroupID,
	}, nil
}

// RetryTask 任务重试
// @Summary 任务重试
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /task/{taskID}/retry [put]
func RetryTask(ctx context.Context, req *RetryTaskkReq) (*bool, error) {
	result, err := clustermgr.RetryTask(ctx, &clustermanager.RetryTaskRequest{
		TaskID:  req.TaskID,
		Updater: req.Updater,
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SkipTask 跳过当前任务
// @Summary 跳过当前失败任务
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /task/{taskID}/skip [put]
func SkipTask(ctx context.Context, req *SkipTaskReq) (*bool, error) {
	result, err := clustermgr.SkipTask(ctx, &clustermanager.SkipTaskRequest{
		TaskID:  req.TaskID,
		Updater: req.Updater,
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}
