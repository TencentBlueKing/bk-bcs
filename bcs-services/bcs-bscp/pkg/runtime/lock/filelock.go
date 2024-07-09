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
	"sync"
)

// FileLock is used to protect the file operation, lock the file by path.
// !It is essentially a resource lock that cannot be shared between processes or between FileLock instances.
type FileLock struct {
	lock     sync.Mutex             // used to protect the mutex of the lock map
	lockPool map[string]*sync.Mutex // mutex lock map, key is the file path
}

// NewFileLock create a file lock
func NewFileLock() *FileLock {
	return &FileLock{
		lock:     sync.Mutex{},
		lockPool: make(map[string]*sync.Mutex),
	}
}

// Acquire try get the file lock by file path
func (flm *FileLock) Acquire(filePath string) {
	flm.lock.Lock()
	if _, ok := flm.lockPool[filePath]; !ok {
		flm.lockPool[filePath] = &sync.Mutex{}
	}
	flm.lock.Unlock()

	flm.lockPool[filePath].Lock()
}

// TryAcquire try get the file lock by file path
func (flm *FileLock) TryAcquire(filePath string) bool {
	flm.lock.Lock()
	if _, ok := flm.lockPool[filePath]; !ok {
		flm.lockPool[filePath] = &sync.Mutex{}
	}
	flm.lock.Unlock()

	return flm.lockPool[filePath].TryLock()
}

// Release release the file lock by file path
func (flm *FileLock) Release(filePath string) {
	flm.lock.Lock()
	defer flm.lock.Unlock()

	if _, ok := flm.lockPool[filePath]; ok {
		flm.lockPool[filePath].Unlock()
	}
}
