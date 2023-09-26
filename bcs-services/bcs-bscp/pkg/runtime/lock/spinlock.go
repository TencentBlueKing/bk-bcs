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

	"go.uber.org/atomic"
	"golang.org/x/time/rate"
)

func newSpinLock(limiter *rate.Limiter) *spinLock {
	return &spinLock{
		cond:      sync.NewCond(&sync.Mutex{}),
		signal:    false,
		state:     &stateLock{status: false},
		withLimit: atomic.NewBool(false),
		limiter:   limiter,
	}
}

type spinLock struct {
	cond      *sync.Cond
	signal    bool
	state     *stateLock
	withLimit *atomic.Bool
	limiter   *rate.Limiter
}

// Acquire test if the caller can acquire the resource's lock.
// if the caller got the lock, then it will return immediately with ture.
// so that it can handle the cache job immediately. Otherwise, the spin
// lock will wait until the lock is released by the one who has acquired
// the lock.
func (sl *spinLock) Acquire() *State {
	if sl.state.tryLock() {
		// acquire the lock, return directly
		return &State{Acquired: true, WithLimit: false}
	}

	// can not get the lock.
	sl.cond.L.Lock()
	for !sl.signal {
		// wait until the lock is released.
		sl.cond.Wait()
	}
	sl.cond.L.Unlock()

	if sl.withLimit.Load() {
		// need to limit the request, then pick one bucket from
		// the limiter, so that the caller can be released slowly.
		_ = sl.limiter.Wait(context.TODO())

		return &State{Acquired: false, WithLimit: true}
	}

	// not acquired the lock.
	return &State{Acquired: false, WithLimit: false}
}

// Release the spin lock so that the resource lock is released.
func (sl *spinLock) Release(withLimit bool) {
	sl.cond.L.Lock()
	// update the signal to ture, so that the awaiting caller
	// can stop wait.
	sl.signal = true
	sl.cond.L.Unlock()

	// Firstly, set with limit control. so that the can pick
	// bucket later.
	sl.withLimit.Store(withLimit)

	// Secondly, notify all the caller that the lock has already
	// been released, stop wait and return.
	sl.cond.Broadcast()

	// Thirdly, reset the state
	sl.state.Reset()
}

type stateLock struct {
	lo     sync.Mutex
	status bool
}

// tryLock is to try to get the lock, if success then return
// true, otherwise, returned false.
func (s *stateLock) tryLock() bool {
	s.lo.Lock()
	defer s.lo.Unlock()
	if s.status {
		// already locked by others,
		// not acquire the lock.
		return false
	}

	// get lock success, set status to true which is
	// used to tell others that this lock is already
	// acquired by someone.
	s.status = true
	return true
}

// Reset the state to unlocked.
func (s *stateLock) Reset() {
	s.lo.Lock()
	defer s.lo.Unlock()
	s.status = false
}
