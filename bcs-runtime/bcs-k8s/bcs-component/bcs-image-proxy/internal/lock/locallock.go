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

package lock

import (
	"context"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// keyedMutex the local lock with string key
type keyedMutex struct {
	mutexes *sync.Map
}

// NewLocalLock create new localLock instance
func NewLocalLock() Interface {
	return &keyedMutex{
		mutexes: &sync.Map{},
	}
}

// Lock the key
func (m *keyedMutex) Lock(ctx context.Context, key string) {
	value, _ := m.mutexes.LoadOrStore(key, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()
}

// UnLock the key
func (m *keyedMutex) UnLock(ctx context.Context, key string) {
	value, _ := m.mutexes.Load(key)
	if value == nil {
		blog.Warnf("local unlock '%s' is empty", key)
		return
	}
	mtx := value.(*sync.Mutex)
	mtx.Unlock()
	return
}
