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

// Package steps include all steps for federation manager
package steps

import (
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
)

var (
	allSteps     []task.StepWorkerInterface
	allCallBacks []task.CallbackInterface
	lock         sync.RWMutex
)

// InitSteps register all steps
func InitSteps(steps []task.StepWorkerInterface, callBacks ...task.CallbackInterface) {
	lock.Lock()
	defer lock.Unlock()
	allSteps = append(allSteps, steps...)
	allCallBacks = append(allCallBacks, callBacks...)
}

// GetAllSteps get all steps
func GetAllSteps() []task.StepWorkerInterface {
	lock.RLock()
	defer lock.RUnlock()

	var steps []task.StepWorkerInterface
	steps = append(steps, allSteps...)
	return steps
}

// GetAllCallbacks get all callbacks
func GetAllCallbacks() []task.CallbackInterface {
	lock.RLock()
	defer lock.RUnlock()

	var callbacks []task.CallbackInterface
	callbacks = append(callbacks, allCallBacks...)
	return callbacks
}
