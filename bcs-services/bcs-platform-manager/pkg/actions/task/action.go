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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// TaskAction task action interface
type TaskAction interface { // nolint
	GetTask(ctx context.Context, req *types.GetTaskReq) (*types.Task, error)
	RetryTask(ctx context.Context, req *types.RetryTaskkReq) (bool, error)
	SkipTask(ctx context.Context, req *types.SkipTaskReq) (bool, error)
}

// Action action for task
type Action struct{}

// NewTaskAction new task action
func NewTaskAction() TaskAction {
	return &Action{}
}

// GetTask get task
func (a *Action) GetTask(ctx context.Context, req *types.GetTaskReq) (*types.Task, error) {
	task, err := clustermgr.GetTask(ctx, &clustermanager.GetTaskRequest{
		TaskID: req.TaskID,
	})
	if err != nil {
		return nil, err
	}

	return &types.Task{
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

// RetryTask retry task
func (a *Action) RetryTask(ctx context.Context, req *types.RetryTaskkReq) (bool, error) {
	result, err := clustermgr.RetryTask(ctx, &clustermanager.RetryTaskRequest{
		TaskID:  req.TaskID,
		Updater: req.Updater,
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

// SkipTask skip task
func (a *Action) SkipTask(ctx context.Context, req *types.SkipTaskReq) (bool, error) {
	result, err := clustermgr.SkipTask(ctx, &clustermanager.SkipTaskRequest{
		TaskID:  req.TaskID,
		Updater: req.Updater,
	})
	if err != nil {
		return false, err
	}

	return result, nil
}
