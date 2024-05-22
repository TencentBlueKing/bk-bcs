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

// Package store implements task storage
package store

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// ListOption options for list task
type ListOption struct {
	// Sort map for sort list results
	Sort map[string]int
	// Offset offset for list results
	Offset int64
	// Limit limit for list results
	Limit int64
	// All for all results
	All bool
	// Count for index
	Count bool
	// SkipDecrypt skip data decrypt
	SkipDecrypt bool
}

// TaskManagerModel model for TaskManager
type TaskManagerModel interface {
	// task information storage management
	CreateTask(ctx context.Context, task *types.Task) error
	UpdateTask(ctx context.Context, task *types.Task) error
	PatchTask(ctx context.Context, taskID string, patchs map[string]interface{}) error
	DeleteTask(ctx context.Context, taskID string) error
	GetTask(ctx context.Context, taskID string) (*types.Task, error)
	ListTask(ctx context.Context, cond *operator.Condition, opt *ListOption) ([]types.Task, error)
}

// ModelSet model for task
type ModelSet struct {
	*ModelTask
}

// NewModelSet create a new model set
func NewModelSet(db drivers.DB, taskPrefix string) TaskManagerModel {
	storeClient := &ModelSet{
		ModelTask: New(db, taskPrefix),
	}
	return storeClient
}
