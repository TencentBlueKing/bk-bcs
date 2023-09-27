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

func TestResourceLock_Acquire(t *testing.T) {
	lo := New(Option{QPS: 2, Burst: 2})
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		state := lo.Acquire("tom")
		if !state.Acquired {
			t.Errorf("tom acquired lock failed")
			return
		}

		t.Logf("tom acquire lock success")

		state = lo.Acquire("tom")
		if state.Acquired {
			t.Errorf("should not run here, tom re-acquired, should not get lock")
			return
		}

		t.Errorf("test tom re-acquire, should not run here without release operation")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		state := lo.Acquire("jerry")
		if !state.Acquired {
			t.Errorf("jerry acquired lock failed")
			return
		}

		t.Logf("jerry acquire lock success")

		lo.Release("jerry", false)

		state = lo.Acquire("jerry")
		if !state.Acquired {
			t.Errorf("jerry re-acquired lock failed, should acquire lock.")
			return
		}

		t.Logf("jerry re-acquire lock success")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		state := lo.Acquire("sam")
		if !state.Acquired {
			t.Errorf("sam acquired lock failed")
			return
		}

		t.Logf("sam acquire lock success")

		state.Release(false)

		state = lo.Acquire("sam")
		if !state.Acquired {
			t.Errorf("sam re-acquired lock failed, should acquire lock.")
			return
		}

		t.Logf("sam re-acquire lock success")
	}()

	wg.Wait()
	time.Sleep(3 * time.Second)
}
