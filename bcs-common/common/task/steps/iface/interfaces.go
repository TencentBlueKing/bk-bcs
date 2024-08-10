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

// Package iface is a package for task step interface
package iface

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

var (
	// ErrParamNotFound 参数未找到
	ErrParamNotFound = errors.New("param not found")
)

// StepExecutor that client must implement
type StepExecutor interface {
	Execute(*Context) error
}

// The StepExecutorFunc type is an adapter to allow the use of
// ordinary functions as a Executor. If f is a function
// with the appropriate signature, StepExecutorFunc(f) is a
// Executor that calls f.
type StepExecutorFunc func(*Context) error

// Execute calls f(ctx, w)
func (f StepExecutorFunc) Execute(c *Context) error {
	return f(c)
}

// CallbackExecutor that callback client must implement
type CallbackExecutor interface {
	Callback(*Context, error)
}

// The CallbackExecutorFunc type is an adapter to allow the use of
// ordinary functions as a Executor. If f is a function
// with the appropriate signature, CallbackExecutorFunc(f) is a
// Executor that calls f.
type CallbackExecutorFunc func(*Context, error)

// Callback calls f(ctx, w)
func (f CallbackExecutorFunc) Callback(c *Context, err error) {
	f(c, err)
}

// TaskBuilder build task
type TaskBuilder interface { // nolint
	Name() string
	Type() string
	BuildTask(info types.TaskInfo, opts ...types.TaskOption) (*types.Task, error)
	Steps(defineSteps []StepBuilder) []*types.Step
}

// StepBuilder build step
type StepBuilder interface {
	Alias() string
	GetName() string
	BuildStep(kvs []KeyValue, opts ...types.StepOption) *types.Step
	DoWork(task *types.Task) error
}

// KeyValue key-value paras
type KeyValue struct {
	Key   ParamKey
	Value string
}

// ParamKey xxx
type ParamKey string

// String xxx
func (pk ParamKey) String() string {
	return string(pk)
}

// TaskType taskType
type TaskType string // nolint

// String toString
func (tt TaskType) String() string {
	return string(tt)
}

// StepName 步骤名称, 通过这个查找Executor, 必须全局唯一
type StepName string

// CallbackName 步骤名称, 通过这个查找callback Executor, 必须全局唯一
type CallbackName string

// TaskName xxx
type TaskName string // nolint

// String xxx
func (tn TaskName) String() string {
	return string(tn)
}
