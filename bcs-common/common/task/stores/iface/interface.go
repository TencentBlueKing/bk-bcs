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

// UpdateOption ...
type UpdateOption struct {
	CurrentStep   string                       `json:"currentStep"`
	CommonParams  map[string]string            `json:"commonParams"`
	ExtraJson     string                       `json:"extraJson"`
	Status        string                       `json:"status"`
	Message       string                       `json:"message"`
	Start         time.Time                    `json:"start"`
	End           time.Time                    `json:"end"`
	ExecutionTime uint32                       `json:"executionTime"`
	LastUpdate    time.Time                    `json:"lastUpdate"`
	Updater       string                       `json:"updater"`
	StepOptions   map[string]*UpdateStepOption `json:"stepOptions"`
}

// UpdateStepOption ...
type UpdateStepOption struct {
	Params        map[string]string `json:"params"`
	Extras        string            `json:"extras"`
	Status        string            `json:"status"`
	Message       string            `json:"message"`
	RetryCount    uint32            `json:"retryCount"`
	Start         time.Time         `json:"start"`
	End           time.Time         `json:"end"`
	ExecutionTime uint32            `json:"executionTime"`
	LastUpdate    time.Time         `json:"lastUpdate"`
}

// Store model for TaskManager
type Store interface {
	EnsureTable(ctx context.Context, dst ...any) error
	CreateTask(ctx context.Context, task *types.Task) error
	ListTask(ctx context.Context, opt *ListOption) ([]types.Task, error)
	GetTask(ctx context.Context, taskID string) (*types.Task, error)
	// UpdateTask(ctx context.Context, taskID string, opt *UpdateOption) error
	DeleteTask(ctx context.Context, taskID string) error
	// GetStep(ctx context.Context, taskID string, stepName string) (*types.Step, error)
	// UpdateStep(ctx context.Context, taskID string, stepName string, opt *UpdateStepOption) error
	UpdateTask(ctx context.Context, task *types.Task) error
	PatchTask(ctx context.Context, taskID string, patchs map[string]interface{}) error
	// WriteStepOutput(ctx context.Context, taskId string, name string, output map[string]string) error
}
