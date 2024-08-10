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
	"sync"
)

// StepName 步骤名称, 通过这个查找Executor, 必须全局唯一
type StepName string

// String ...
func (s StepName) String() string {
	return string(s)
}

// CallbackName 步骤名称, 通过这个查找callback Executor, 必须全局唯一
type CallbackName string

// String ...
func (cb CallbackName) String() string {
	return string(cb)
}

var (
	stepMu    sync.RWMutex
	steps     = make(map[StepName]StepExecutor)
	callBacks = make(map[CallbackName]CallbackExecutor)
)

// Register makes a StepExecutor available by the provided name.
// If Register is called twice with the same name or if StepExecutor is nil,
// it panics.
func Register(name StepName, step StepExecutor) {
	stepMu.Lock()
	defer stepMu.Unlock()

	if step == nil {
		panic("task: Register step is nil")
	}

	if _, dup := steps[name]; dup {
		panic("task: Register step twice for executor " + name)
	}

	steps[name] = step
}

// GetRegisters get all steps instance
func GetRegisters() map[StepName]StepExecutor {
	stepMu.Lock()
	defer stepMu.Unlock()

	return steps
}

// RegisterCallback ...
func RegisterCallback(name CallbackName, cb CallbackExecutor) {
	stepMu.Lock()
	defer stepMu.Unlock()

	if cb == nil {
		panic("task: Register callback is nil")
	}

	if _, dup := callBacks[name]; dup {
		panic("task: Register callback twice for executor " + name)
	}

	callBacks[name] = cb
}

// GetCallbackRegisters get all steps instance
func GetCallbackRegisters() map[CallbackName]CallbackExecutor {
	stepMu.Lock()
	defer stepMu.Unlock()

	return callBacks
}
