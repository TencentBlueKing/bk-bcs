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

// Package lock NOTES
package lock

import (
	"sync"
	"time"

	prm "github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"
	"golang.org/x/time/rate"
)

// Interface defines all the supported operations for the resource's lock,
// which is used to manage the lock of different kind of resources.
type Interface interface {
	// Acquire test if the caller can get the resource's lock.
	// if the caller got the lock, then it will return immediately with ture.
	// so that it can handle the cache job immediately. Otherwise, the spin
	// lock will wait until the lock is released by the one who has acquired
	// the lock.
	Acquire(resource string) *State

	// Release the resource lock so that the resource lock is released.
	// after the resource's lock is released, the other call which is still
	// waiting for the lock to be released will return immediately.
	Release(resource string, withLimit bool)
}

// Option defines options to initial a resource lock.
type Option struct {
	// QPS should >=1
	QPS uint
	// Burst should >= 1, otherwise the limiter
	// can not work correctly.
	Burst uint
}

// New initialize a resource lock instance.
func New(opt Option) Interface {
	mc := initResourceLockMetric()
	lo := &resLock{
		lo:      sync.Mutex{},
		pool:    make(map[string]*spinLock),
		mc:      mc,
		limiter: rate.NewLimiter(rate.Limit(opt.QPS), int(opt.Burst)),
		stat:    &resourceLockStatistic{mc: mc},
	}

	go lo.collectMetric()

	return lo
}

// resLock is used to manage the lock of different kind of resources.
// Note: add capacity limit, limit the max number of the resource's
// kind, if overhead of the max limit, then return with an error.
type resLock struct {
	lo      sync.Mutex
	pool    map[string]*spinLock
	limiter *rate.Limiter
	stat    *resourceLockStatistic
	mc      *resourceLockMetric
}

// Acquire test if the caller can get the resource's lock.
// if the caller got the lock, then it will return immediately with ture.
// so that it can handle the cache job immediately. Otherwise, the spin
// lock will wait until the lock is released by the one who has acquired
// the lock.
func (r *resLock) Acquire(res string) *State {

	r.stat.IncTotal()

	r.lo.Lock()
	sl, exist := r.pool[res]
	if exist {
		r.lo.Unlock()

		state := sl.Acquire()
		state.resource = res
		state.lock = r

		if state.Acquired {
			r.stat.IncAcquired()
		}

		return state
	}

	// this resource is not in the pool for now.
	// initial a new spinlock for it.
	sl = newSpinLock(r.limiter)
	r.pool[res] = sl
	r.lo.Unlock()

	state := sl.Acquire()
	state.resource = res
	state.lock = r

	if state.Acquired {
		r.stat.IncAcquired()
	}

	return state
}

// Release the resource lock so that the resource lock is released.
// after the resource's lock is released, the other call which is still
// waiting for the lock to be released will return immediately.
func (r *resLock) Release(resource string, withLimit bool) {
	r.lo.Lock()

	sl, exist := r.pool[resource]
	if !exist {
		r.lo.Unlock()
		// no spin lock is found for this resource, return directly.
		return
	}

	// remove the resource's spin lock
	delete(r.pool, resource)
	r.lo.Unlock()

	// release the spin lock to awake the awaiting caller to return.
	sl.Release(withLimit)
}

func (r *resLock) collectMetric() {
	for {
		time.Sleep(5 * time.Second)
		r.mc.acquiredRate.With(prm.Labels{}).Set(r.stat.Rate())
	}
}

// State describe the state of the caller to acquire the lock.
type State struct {
	// Acquired defines if the caller have acquired the lock,
	// which true means acquired the lock.
	Acquired bool
	// WithLimit means the caller does not acquire the lock and
	// the one who have acquired the lock request the other caller
	// to consume the resource with limit.
	WithLimit bool

	// resource defines which kind of resource is locked.
	resource string

	// lock is this resource's parent lock.
	lock *resLock
}

// Release the resource lock so that the resource lock is released.
// after the resource's lock is released, the other call which is still
// waiting for the lock to be released will return immediately.
func (s *State) Release(withLimit bool) {
	if !s.Acquired {
		return
	}

	s.lock.Release(s.resource, withLimit)
}

type resourceLockStatistic struct {
	Total    atomic.Int64
	Acquired atomic.Int64
	mc       *resourceLockMetric
}

// IncTotal increase the total by one.
func (sc *resourceLockStatistic) IncTotal() {
	sc.Total.Inc()
	sc.mc.totalCounter.With(prm.Labels{}).Inc()
}

// IncAcquired increase the acquired by one.
func (sc *resourceLockStatistic) IncAcquired() {
	sc.Total.Inc()
	sc.mc.acquiredCounter.With(prm.Labels{}).Inc()
}

// Rate is the acquired rate of total.
func (sc *resourceLockStatistic) Rate() float64 {
	if sc.Total.Load() == 0 {
		return 0
	}

	return float64(sc.Acquired.Load() / sc.Total.Load())
}
