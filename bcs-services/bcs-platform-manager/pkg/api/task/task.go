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

	actions "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/actions/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// GetTask 获取任务详情
// @Summary 任务详情
// @Tags    Logs
// @Produce json
// @Success 200 {struct} types.Task
// @Router  /task/{taskID} [get]
func GetTask(ctx context.Context, req *types.GetTaskReq) (*types.Task, error) {
	result, err := actions.NewTaskAction().GetTask(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// RetryTask 任务重试
// @Summary 任务重试
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /task/{taskID}/retry [put]
func RetryTask(ctx context.Context, req *types.RetryTaskkReq) (*bool, error) {
	result, err := actions.NewTaskAction().RetryTask(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SkipTask 跳过当前任务
// @Summary 跳过当前失败任务
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /task/{taskID}/skip [put]
func SkipTask(ctx context.Context, req *types.SkipTaskReq) (*bool, error) {
	result, err := actions.NewTaskAction().SkipTask(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
