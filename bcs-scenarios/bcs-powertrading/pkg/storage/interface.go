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

// Package storage xxx
package storage

import (
	"context"
)

// ListOptions for list operation
type ListOptions struct {
	Limit                  int
	Page                   int
	ReturnSoftDeletedItems bool
	DoPagination           bool
}

// CreateOptions for create strategy
type CreateOptions struct {
	OverWriteIfExist bool
}

// UpdateOptions for update strategy
type UpdateOptions struct {
	CreateIfNotExist        bool
	OverwriteZeroOrEmptyStr bool
}

// DeleteOptions for delete strategy
type DeleteOptions struct {
	ErrIfNotExist bool
}

// GetOptions for get single data
type GetOptions struct {
	ErrIfNotExist  bool
	GetSoftDeleted bool
}

// Storage interface define data object store behavior
// that is independent of any kind of implementation,
// such as MySQL, MongoDB
type Storage interface {
	CreateMachineTestTask(ctx context.Context, task *MachineTask, opt *CreateOptions) error
	UpdateTask(ctx context.Context, task *MachineTask, opt *UpdateOptions) (*MachineTask, error)
	ListTasks(ctx context.Context, taskType string, opt *ListOptions) ([]*MachineTask, error)
	DeleteTask(ctx context.Context, taskID string, opt *DeleteOptions) (*MachineTask, error)
	GetTask(ctx context.Context, taskID string, opt *GetOptions) (*MachineTask, error)
	CreateDeviceData(ctx context.Context, data *DeviceOperationData, opt *CreateOptions) error
}
