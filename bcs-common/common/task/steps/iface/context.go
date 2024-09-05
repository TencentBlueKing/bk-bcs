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
	"errors"
	"time"

	istore "github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

var (
	// ErrRevoked step has been revoked
	ErrRevoked = errors.New("revoked")
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
	return t.task.GetTaskName()
}

// GetTaskType get task type
func (t *Context) GetTaskType() string {
	return t.task.GetTaskType()
}

// GetTaskIndex get task index
func (t *Context) GetTaskIndex() string {
	return t.task.GetTaskIndex()
}

// GetTaskStatus get task status
func (t *Context) GetTaskStatus() string {
	return t.task.GetStatus()
}

// GetCommonParam get current task param
func (t *Context) GetCommonParam(key string) (string, bool) {
	return t.task.GetCommonParam(key)
}

// AddCommonParam add task common params
func (t *Context) AddCommonParam(k, v string) error {
	_ = t.task.AddCommonParam(k, v)
	return t.store.UpdateTask(t.ctx, t.task)
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

// GetRetryCount get current step retrycount
func (t *Context) GetRetryCount() uint32 {
	return t.currentStep.GetRetryCount()
}

// GetParam get current step param
func (t *Context) GetParam(key string) (string, bool) {
	return t.currentStep.GetParam(key)
}

// AddParam set step param by key,value
func (t *Context) AddParam(key string, value string) error {
	_ = t.currentStep.AddParam(key, value)
	return t.store.UpdateTask(t.ctx, t.task)
}

// GetParams return all step params
func (t *Context) GetParams() map[string]string {
	return t.currentStep.GetParams()
}

// SetParams return all step params
func (t *Context) SetParams(params map[string]string) error {
	t.currentStep.SetParams(params)
	return t.store.UpdateTask(t.ctx, t.task)
}

// GetPayload return unmarshal step extras
func (t *Context) GetPayload(obj interface{}) error {
	return t.currentStep.GetPayload(obj)
}

// GetStartTime return step start time
func (t *Context) GetStartTime() time.Time {
	return t.currentStep.Start
}

// SetPayload set step extras by json string
func (t *Context) SetPayload(obj interface{}) error {
	if err := t.currentStep.SetPayload(obj); err != nil {
		return err
	}

	return t.store.UpdateTask(t.ctx, t.task)
}
