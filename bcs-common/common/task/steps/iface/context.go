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
func (c *Context) Context() context.Context {
	return c.ctx
}

// GetTaskID get task id
func (c *Context) GetTaskID() string {
	return c.task.GetTaskID()
}

// GetTaskName get task name
func (c *Context) GetTaskName() string {
	return c.task.GetTaskName()
}

// GetTaskType get task type
func (c *Context) GetTaskType() string {
	return c.task.GetTaskType()
}

// GetTaskIndex get task index
func (c *Context) GetTaskIndex() string {
	return c.task.GetTaskIndex()
}

// GetTaskStatus get task status
func (c *Context) GetTaskStatus() string {
	return c.task.GetStatus()
}

// GetCommonParam get current task param
func (c *Context) GetCommonParam(key string) (string, bool) {
	return c.task.GetCommonParam(key)
}

// AddCommonParam add task common param and save to store
func (c *Context) AddCommonParam(k, v string) error {
	_ = c.task.AddCommonParam(k, v)
	return c.store.UpdateTask(c.ctx, c.task)
}

// GetCommonPayload unmarshal task common payload to struct obj
func (c *Context) GetCommonPayload(obj any) error {
	return c.task.GetCommonPayload(obj)
}

// SetCommonPayload marshal struct obj to task common payload and save to store
func (c *Context) SetCommonPayload(obj any) error {
	if err := c.task.SetCommonPayload(obj); err != nil {
		return err
	}

	return c.store.UpdateTask(c.ctx, c.task)
}

// GetName get current step name
func (c *Context) GetName() string {
	return c.currentStep.GetName()
}

// GetStatus get current step status
func (c *Context) GetStatus() string {
	return c.currentStep.GetStatus()
}

// GetRetryCount get current step retry count
func (c *Context) GetRetryCount() uint32 {
	return c.currentStep.GetRetryCount()
}

// GetParam get current step param by key
func (c *Context) GetParam(key string) (string, bool) {
	return c.currentStep.GetParam(key)
}

// AddParam set step param by key,value and save to store
func (c *Context) AddParam(key string, value string) error {
	_ = c.currentStep.AddParam(key, value)
	return c.store.UpdateTask(c.ctx, c.task)
}

// GetParams return all step params
func (c *Context) GetParams() map[string]string {
	return c.currentStep.GetParams()
}

// SetParams set all step params and save to store
func (c *Context) SetParams(params map[string]string) error {
	c.currentStep.SetParams(params)
	return c.store.UpdateTask(c.ctx, c.task)
}

// GetPayload return unmarshal step payload
func (c *Context) GetPayload(obj any) error {
	return c.currentStep.GetPayload(obj)
}

// GetStartTime return step start time
func (c *Context) GetStartTime() time.Time {
	return c.currentStep.Start
}

// SetPayload marshal struct obj to step payload and save to store
func (c *Context) SetPayload(obj any) error {
	if err := c.currentStep.SetPayload(obj); err != nil {
		return err
	}

	return c.store.UpdateTask(c.ctx, c.task)
}
