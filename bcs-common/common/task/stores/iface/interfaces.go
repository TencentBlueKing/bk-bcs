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

// Package iface defines the interface for store.
package iface

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// ListOption ...
type ListOption struct {
	TaskID        string
	TaskType      string
	TaskName      string
	TaskIndex     string
	TaskIndexType string
	CurrentStep   string
	Status        string
	Creator       string
	CreatedGte    *time.Time     // CreatedGte create time greater or equal to
	CreatedLte    *time.Time     // CreatedLte create time less or equal to
	Sort          map[string]int // Sort map for sort list results
	Offset        int64          // Offset offset for list results
	Limit         int64          // Limit limit for list results
}

// Pagination generic pagination for list results
type Pagination[T any] struct {
	Count int64 `json:"count"`
	Items []*T  `json:"items"`
}

// PatchOption 主要实时更新params, payload信息
type PatchOption struct {
	Task        *types.Task
	CurrentStep *types.Step
}

// Store model for TaskManager
type Store interface {
	EnsureTable(ctx context.Context, dst ...any) error
	CreateTask(ctx context.Context, task *types.Task) error
	ListTask(ctx context.Context, opt *ListOption) (*Pagination[types.Task], error)
	GetTask(ctx context.Context, taskID string) (*types.Task, error)
	DeleteTask(ctx context.Context, taskID string) error
	UpdateTask(ctx context.Context, task *types.Task) error
}
