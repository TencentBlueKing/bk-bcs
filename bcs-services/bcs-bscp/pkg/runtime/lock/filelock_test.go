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
	"testing"
	"time"
)

func TestFileLock1(t *testing.T) {
	// [1] acquire success, release after 5 seconds
	// [2] try acquire failed, retry until success
	rl := NewFileLock()
	wg := &sync.WaitGroup{}

	go func(wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		rl.Acquire("resource-key")
		t.Log("[1] acquire success, release after 5 seconds")
		time.Sleep(5 * time.Second)
		rl.Release("resource-key")
		t.Log("[1] release")
	}(wg)

	go func(wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		// sleep 1s to let goroutine [1] acquire lock first
		time.Sleep(1 * time.Second)
		for {
			if rl.TryAcquire("resource-key") {
				t.Log("[2] try acquire success, release after 5 seconds")
				time.Sleep(5 * time.Second)
				rl.Release("resource-key")
				t.Log("[2] release")
				break
			} else {
				t.Log("[2] try acquire failed")
				time.Sleep(1 * time.Second)
			}
		}
	}(wg)

	// sleep in case goroutine have not had time to add waitgroup
	time.Sleep(1 * time.Second)
	wg.Wait()
}

func TestFileLock2(t *testing.T) {
	// [2] try acquire success, release after 5 seconds
	// [1] block to acquire
	// (5 secones later...)
	// [2] release
	// [1] acquire success
	rl := NewFileLock()
	wg := &sync.WaitGroup{}

	go func(wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		// sleep 1s to let goroutine [2] acquire lock first
		time.Sleep(1 * time.Second)
		rl.Acquire("resource-key")
		t.Log("[1] acquire success, release after 5 seconds")
		time.Sleep(5 * time.Second)
		rl.Release("resource-key")
		t.Log("[1] release")
	}(wg)

	go func(wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()
		for {
			if rl.TryAcquire("resource-key") {
				t.Log("[2] try acquire success, release after 5 seconds")
				time.Sleep(5 * time.Second)
				rl.Release("resource-key")
				t.Log("[2] release")
				break
			} else {
				t.Log("[2] try acquire failed")
				time.Sleep(1 * time.Second)
			}
		}
	}(wg)

	// sleep in case goroutine have not had time to add waitgroup
	time.Sleep(1 * time.Second)
	wg.Wait()
}
