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

	istore "github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// Context 当前执行的任务
type Context struct {
	ctx         context.Context
	store       istore.Store
	task        *types.Task
	currentStep *types.Step
}

// NewContext ...
func NewContext(ctx context.Context, store istore.Store, task *types.Task, currentStep *types.Step) *Context {
	return &Context{
		ctx:         ctx,
		store:       store,
		task:        task,
		currentStep: currentStep,
	}
}

// Context returns the step's context
func (t *Context) Context() context.Context {
	return t.ctx
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

// GetTaskStatus get task status
func (t *Context) GetTaskStatus() string {
	return t.task.GetStatus()
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

// GetCommonPayload get task extra json
func (t *Context) GetCommonPayload(obj interface{}) error {
	return t.task.GetCommonPayload(obj)
}

// SetCommonPayload set task extra json
func (t *Context) SetCommonPayload(obj interface{}) error {
	if err := t.task.SetCommonPayload(obj); err != nil {
		return err
	}

	return t.store.UpdateTask(t.ctx, t.task)
}

// GetName get current step name
func (t *Context) GetName() string {
	return t.currentStep.GetName()
}

// GetStatus get current step status
func (t *Context) GetStatus() string {
	return t.currentStep.GetStatus()
}

// GetParam get current step param
func (t *Context) GetParam(key string) (string, bool) {
	return t.currentStep.GetParam(key)
}

// AddParam set step param by key,value
func (t *Context) AddParam(key string, value string) {
	_ = t.currentStep.AddParam(key, value)
}

// GetParamsAll return all step params
func (t *Context) GetParamsAll() map[string]string {
	return t.currentStep.GetParamsAll()
}

// SetParamMulti return all step params
func (t *Context) SetParamMulti(params map[string]string) error {
	t.currentStep.SetParamMulti(params)
	return t.store.UpdateTask(t.ctx, t.task)
}

// GetPayload return unmarshal step extras
func (t *Context) GetPayload(obj interface{}) error {
	return t.currentStep.GetPayload(obj)
}

// SetPayload set step extras by json string
func (t *Context) SetPayload(obj interface{}) error {
	if err := t.currentStep.SetPayload(obj); err != nil {
		return err
	}

	return t.store.UpdateTask(t.ctx, t.task)
}
