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
 *
 */

package model

import "sync"

// AtomicBool thread safe bool type
type AtomicBool struct {
	bo   bool
	lock sync.Mutex
}

// NewAtomicBool return AtomicBool obj
func NewAtomicBool() *AtomicBool {
	return &AtomicBool{
		bo: false,
	}
}

// Value get AtomicBool value
func (ab *AtomicBool) Value() bool {
	ab.lock.Lock()
	defer ab.lock.Unlock()
	return ab.bo
}

// Set set AtomicBool value
func (ab *AtomicBool) Set(value bool) {
	ab.lock.Lock()
	defer ab.lock.Unlock()
	ab.bo = value
}
