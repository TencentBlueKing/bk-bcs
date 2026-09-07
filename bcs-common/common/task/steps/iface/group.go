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
	"sync"

	istore "github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// GroupCallbackName 任务组回调名称, 必须全局唯一
type GroupCallbackName string

// String ...
func (name GroupCallbackName) String() string {
	return string(name)
}

// GroupContext 任务组回调的执行上下文
type GroupContext struct {
	ctx   context.Context
	store istore.Store
	group *types.TaskGroup
}

// NewGroupContext ...
func NewGroupContext(ctx context.Context, store istore.Store, group *types.TaskGroup) *GroupContext {
	return &GroupContext{ctx: ctx, store: store, group: group}
}

// Context returns the group callback's context
func (c *GroupContext) Context() context.Context {
	return c.ctx
}

// GetGroup get the task group
func (c *GroupContext) GetGroup() *types.TaskGroup {
	return c.group
}

// GetGroupID get group id
func (c *GroupContext) GetGroupID() string {
	return c.group.GetGroupID()
}

// GetGroupType get group type
func (c *GroupContext) GetGroupType() string {
	return c.group.GroupType
}

// GetGroupName get group name
func (c *GroupContext) GetGroupName() string {
	return c.group.GroupName
}

// GetGroupIndex get group index
func (c *GroupContext) GetGroupIndex() string {
	return c.group.GroupIndex
}

// GetStatus get group status
func (c *GroupContext) GetStatus() string {
	return c.group.GetStatus()
}

// GetCommonParam get group common param
func (c *GroupContext) GetCommonParam(key string) (string, bool) {
	if c.group.CommonParams == nil {
		return "", false
	}
	value, ok := c.group.CommonParams[key]
	return value, ok
}

// GetCommonPayload unmarshal group common payload to struct obj
func (c *GroupContext) GetCommonPayload(obj any) error {
	return c.group.GetCommonPayload(obj)
}

// GetTask 读取任务组内的某个任务, 便于回调按任务粒度做业务收尾
func (c *GroupContext) GetTask(taskID string) (*types.Task, error) {
	return c.store.GetTask(c.ctx, taskID)
}

// GroupCallbackExecutor 任务组级回调, 由调用方按需实现
type GroupCallbackExecutor interface {
	// OnStageBlocked 前序阶段失败导致后续任务被直接终结时触发。
	// 这些任务从未执行, 不会触发各自的 Callback, 它们占用的业务资源只能在这里回收。
	OnStageBlocked(c *GroupContext, blockedTaskIDs []string) error

	// OnGroupComplete 任务组到达终态时触发
	OnGroupComplete(c *GroupContext) error
}

var (
	groupCbMu      sync.RWMutex
	groupCallbacks = make(map[GroupCallbackName]GroupCallbackExecutor)
)

// RegisterGroupCallback makes a GroupCallbackExecutor available by the provided name.
// 只应在服务初始化阶段调用, 不支持运行时动态注册。
// If RegisterGroupCallback is called twice with the same name or if the executor is nil, it panics.
func RegisterGroupCallback(name GroupCallbackName, cb GroupCallbackExecutor) {
	groupCbMu.Lock()
	defer groupCbMu.Unlock()

	if cb == nil {
		panic("task: Register group callback is nil")
	}

	if _, dup := groupCallbacks[name]; dup {
		panic("task: Register group callback twice for executor " + name)
	}

	groupCallbacks[name] = cb
}

// GetGroupCallbackRegisters get all registered group callbacks
func GetGroupCallbackRegisters() map[GroupCallbackName]GroupCallbackExecutor {
	groupCbMu.RLock()
	defer groupCbMu.RUnlock()

	return groupCallbacks
}
