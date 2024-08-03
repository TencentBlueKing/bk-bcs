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

// Package task is a package for task management
package iface

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// StepWorkerInterface that client must implement
type StepWorkerInterface interface {
	DoWork(context.Context, *Work) error
}

// The StepWorkerFunc type is an adapter to allow the use of
// ordinary functions as a Handler. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type StepWorkerFunc func(context.Context, *Work) error

// CallbackInterface that client must implement
type CallbackInterface interface {
	GetName() string
	Callback(isSuccess bool, task *types.Task)
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

// TaskName xxx
type TaskName string // nolint

// String xxx
func (tn TaskName) String() string {
	return string(tn)
}

// Task 当前执行的任务
type Work struct {
	task *types.Task
	step *types.Step
}

// NewWork ...
func NewWork(t *types.Task, currentStep *types.Step) *Work {
	return &Work{
		task: t,
		step: currentStep,
	}
}

// GetTaskID ...
func (t *Work) GetTaskID() string {
	return t.task.GetTaskID()
}

func (t *Work) GetName() string {
	return t.step.GetName()
}

func (t *Work) GetTaskType() string {
	return t.task.GetTaskID()
}

func (t *Work) AddCommonParams(k, v string) error {
	_ = t.task.AddCommonParams(k, v)
	return nil
}

func (t *Work) GetParam(key string) (string, bool) {
	return t.step.GetParam(key)
}

func (t *Work) GetTaskName() string {
	return t.task.GetTaskID()
}

func (t *Work) GetStatus() string {
	return t.step.GetStatus()
}
