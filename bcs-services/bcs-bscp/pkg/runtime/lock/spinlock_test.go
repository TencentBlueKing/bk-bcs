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

	"golang.org/x/time/rate"
)

func TestSpinLock(t *testing.T) {
	limiter := rate.NewLimiter(1, 2)
	sl := newSpinLock(limiter)

	wg := sync.WaitGroup{}
	total := 5
	say := ""
	for i := 1; i <= total; i++ {
		wg.Add(1)
		go func(num int) {
			start := time.Now()
			defer wg.Done()

			state := sl.Acquire()
			if state.Acquired {
				say = "hello world!!!"
				t.Logf("%d decide what to say!", num)

				// sleep a while and let them wait for releasing
				// the lock.
				time.Sleep(time.Second)
				sl.Release(false)
				return
			}

			if state.WithLimit {
				t.Errorf("should not be limited.")
				return
			}

			t.Logf("%d say '%s' after %s", num, say, time.Since(start).String())
		}(i)
	}

	wg.Wait()
	time.Sleep(time.Second)
}

func TestSpinLockWithLimit(t *testing.T) {
	limiter := rate.NewLimiter(1, 1)
	sl := newSpinLock(limiter)

	wg := sync.WaitGroup{}
	total := 5
	say := ""
	for i := 1; i <= total; i++ {
		wg.Add(1)
		go func(num int) {
			start := time.Now()
			defer wg.Done()

			state := sl.Acquire()
			if state.Acquired {
				say = "let's run with limit!!!"
				t.Logf("%d decide what to say!\n", num)

				// sleep a while and let them wait for releasing
				// the lock.
				time.Sleep(time.Second)
				sl.Release(true)
				return
			}

			if !state.WithLimit {
				t.Errorf("should be with limited.")
				return
			}

			t.Logf("%d say '%s' after %s\n", num, say, time.Since(start).String())
		}(i)
	}

	wg.Wait()

	time.Sleep(time.Second)
}
