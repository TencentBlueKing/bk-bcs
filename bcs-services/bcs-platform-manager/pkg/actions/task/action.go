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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"
)

// TaskAction task action interface
type TaskAction interface { // nolint
	ListTask(ctx context.Context, req *types.ListTaskReq) (*types.ListTaskResp, error)
	GetTask(ctx context.Context, req *types.GetTaskReq) (*types.Task, error)
	RetryTask(ctx context.Context, req *types.RetryTaskkReq) (bool, error)
	SkipTask(ctx context.Context, req *types.SkipTaskReq) (bool, error)
	UpdateTask(ctx context.Context, req *types.UpdateTaskReq) (bool, error)
}

// Action action for task
type Action struct{}

// NewTaskAction new task action
func NewTaskAction() TaskAction {
	return &Action{}
}

// ListTask list task
func (a *Action) ListTask(ctx context.Context, req *types.ListTaskReq) (*types.ListTaskResp, error) {
	/*	taskData, err := clustermgr.ListTask(ctx, &clustermanager.ListTaskV2Request{
			ClusterID:   req.ClusterID,
			ProjectID:   req.ProjectID,
			Creator:     req.Creator,
			Updater:     req.Updater,
			TaskType:    req.TaskType,
			Status:      req.Status,
			NodeIP:      req.NodeIP,
			NodeGroupID: req.NodeGroupID,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Limit:       req.Limit,
			Page:        req.Page,
		})
		if err != nil {
			return nil, utils.SystemError(err)
		}

		result := make([]*types.Task, 0)
		for _, task := range taskData.Results {
			result = append(result, &types.Task{
				TaskID:         task.TaskID,
				TaskType:       task.TaskType,
				Status:         task.Status,
				Message:        task.Message,
				Start:          task.Start,
				End:            task.End,
				ExecutionTime:  task.ExecutionTime,
				ClusterID:      task.ClusterID,
				ProjectID:      task.ProjectID,
				Creator:        task.Creator,
				LastUpdate:     task.LastUpdate,
				Updater:        task.Updater,
				ForceTerminate: task.ForceTerminate,
				TaskName:       task.TaskName,
				NodeGroupID:    task.NodeGroupID,
				NodeIPList:     task.NodeIPList,
			})
		}

		return &types.ListTaskResp{
			Result: result,
			Total:  taskData.Count,
		}, nil*/
	return nil, nil
}

// GetTask get task
func (a *Action) GetTask(ctx context.Context, req *types.GetTaskReq) (*types.Task, error) {
	task, err := clustermgr.GetTask(ctx, &clustermanager.GetTaskRequest{
		TaskID: req.TaskID,
	})
	if err != nil {
		return nil, utils.SystemError(err)
	}

	return &types.Task{
		TaskID:         task.TaskID,
		TaskType:       task.TaskType,
		Status:         task.Status,
		Message:        task.Message,
		Start:          task.Start,
		End:            task.End,
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
		return false, utils.SystemError(err)
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
		return false, utils.SystemError(err)
	}

	return result, nil
}

// UpdateTask update task
func (a *Action) UpdateTask(ctx context.Context, req *types.UpdateTaskReq) (bool, error) {
	steps := make(map[string]*clustermanager.Step)
	for k, v := range req.Steps {
		steps[k] = &clustermanager.Step{
			Name:          v.Name,
			System:        v.System,
			Link:          v.Link,
			Params:        v.Params,
			Retry:         v.Retry,
			Start:         v.Start,
			End:           v.End,
			ExecutionTime: v.ExecutionTime,
			Status:        v.Status,
			Message:       v.Message,
			LastUpdate:    v.LastUpdate,
			TaskMethod:    v.TaskMethod,
			TaskName:      v.TaskName,
			SkipOnFailed:  v.SkipOnFailed,
			Translate:     v.Translate,
			AllowSkip:     v.AllowSkip,
			MaxRetry:      v.MaxRetry,
		}
	}

	result, err := clustermgr.UpdateTask(ctx, &clustermanager.UpdateTaskRequest{
		TaskID:        req.TaskID,
		Updater:       req.Updater,
		Status:        req.Status,
		Message:       req.Message,
		End:           req.End,
		ExecutionTime: req.ExecutionTime,
		CurrentStep:   req.CurrentStep,
		Steps:         steps,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}
