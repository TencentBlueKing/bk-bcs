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

package iface

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// Context 当前执行的任务
type Context struct {
	ctx         context.Context
	task        *types.Task
	currentStep *types.Step
}

// NewContext ...
func NewContext(ctx context.Context, task *types.Task, currentStep *types.Step) *Context {
	return &Context{
		ctx:         ctx,
		task:        task,
		currentStep: currentStep,
	}
}

// GetTaskID get task id
func (t *Context) GetTaskID() string {
	return t.task.GetTaskID()
}

// GetTaskName get task name
func (t *Context) GetTaskName() string {
	return t.task.GetTaskID()
}

// GetTaskType get task type
func (t *Context) GetTaskType() string {
	return t.task.GetTaskType()
}

// GetCommonParams get current task param
func (t *Context) GetCommonParams(key string) (string, bool) {
	return t.task.GetCommonParams(key)
}

// AddCommonParams add task common params
func (t *Context) AddCommonParams(k, v string) error {
	_ = t.task.AddCommonParams(k, v)
	return nil
}

// GetCommonPayload add task common params
func (t *Context) GetCommonPayload(obj interface{}) error {
	return t.task.GetCommonPayload(obj)
}

// GetName get current step name
func (t *Context) GetName() string {
	return t.currentStep.GetName()
}

// GetParam get current step param
func (t *Context) GetParam(key string) (string, bool) {
	return t.currentStep.GetParam(key)
}

// GetParamsAll return all step params
func (t *Context) GetParamsAll() map[string]string {
	return t.currentStep.GetParamsAll()
}

// GetStatus get current step status
func (t *Context) GetStatus() string {
	return t.currentStep.GetStatus()
}

// GetPayload ...
func (t *Context) GetPayload(obj interface{}) error {
	return t.currentStep.GetPayload(obj)
}
