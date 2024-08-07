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

var (
	stepMu sync.RWMutex
	steps  = make(map[string]StepWorkerInterface)
)

// Register makes a StepWorkerInterface available by the provided name.
// If Register is called twice with the same name or if StepWorkerInterface is nil,
// it panics.
func Register(name string, step StepWorkerInterface) {
	stepMu.Lock()
	defer stepMu.Unlock()

	if step == nil {
		panic("task: Register step is nil")
	}

	if _, dup := steps[name]; dup {
		panic("task: Register step twice for work " + name)
	}

	steps[name] = step
}

// GetRegisters get all steps instance
func GetRegisters() map[string]StepWorkerInterface {
	stepMu.Lock()
	defer stepMu.Unlock()

	return steps
}
