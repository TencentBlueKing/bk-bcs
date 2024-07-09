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
	"fmt"
	"time"

	prm "github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/bedis"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// RedisLock is a distributed resource lock
type RedisLock struct {
	ctx context.Context
	bds bedis.Client
	// The lock would automatically releas after ttl seconds.
	ttl  uint
	stat *redisLockStatistic
	mc   *redisLockMetric
}

// NewRedisLock initialize a redis lock instance.
func NewRedisLock(bds bedis.Client, ttl uint) *RedisLock {
	mc := initRedisLockMetric()
	lo := &RedisLock{
		ctx:  context.Background(),
		bds:  bds,
		ttl:  ttl,
		mc:   mc,
		stat: &redisLockStatistic{mc: mc},
	}

	go lo.collectMetric()

	return lo
}

// Acquire caller try to get the resource's lock.
// if the caller got the lock, then it will return immediately with ture.
// Otherwise, the caller will wait until the lock is released by the one who has acquired
// the lock.
func (r *RedisLock) Acquire(res string) {

	r.stat.IncTotal()

	key := fmt.Sprintf("Lock:%s", res)

	for {
		timeout := time.Now().Add(time.Duration(r.ttl) * time.Second)
		ok, err := r.bds.SetNX(r.ctx, key, timeout, int(r.ttl+5))

		// if ok is true, then the lock is acquired by the caller.
		if err == nil && ok {
			r.stat.IncAcquired()
			return
		}

		if err != nil {
			logs.Errorf("acquire redis lock %s failed, setnx resource lock failed, err: %s", key, err.Error())
			continue
		}

		// check if the lock is expired
		str, err := r.bds.Get(r.ctx, key)
		if err != nil {
			logs.Errorf("acquire redis lock %s failed, get resource lock failed, err: %s", key, err.Error())
			continue
		}

		if str == "" {
			logs.Warnf("acquire redis lock %s failed, resource lock not exist in redis", key)
			continue
		}

		oldTimeout, err := time.Parse(time.RFC3339, str)
		if err != nil {
			logs.Errorf("acquire redis lock %s failed, parse timeout failed, err: %s", key, err.Error())
			continue
		}

		if time.Now().After(oldTimeout) {
			oldStr, err := r.bds.GetSet(r.ctx, key, timeout.Format(time.RFC3339))
			if err != nil {
				logs.Errorf("acquire redis lock %s failed, getset resource lock failed, err: %s", key, err.Error())
				continue
			}
			oldTimeout, err := time.Parse(time.RFC3339, oldStr)
			if err != nil {
				logs.Errorf("acquire redis lock %s failed, parse timeout failed, err: %s", key, err.Error())
				continue
			}
			if time.Now().After(oldTimeout) {
				logs.Warnf("acquire redis lock %s success, resource lock is expired", key)
				r.stat.IntTimeout()
				r.stat.IncAcquired()
				return
			}
		}

		time.Sleep(time.Millisecond * 50)
	}
}

// TryAcquire caller try to get the resource's lock.
// Whether or not the caller got the lock, it will return immediately.
func (r *RedisLock) TryAcquire(res string) bool {

	r.stat.IncTotal()

	key := fmt.Sprintf("Lock:%s", res)

	timeout := time.Now().Add(time.Duration(r.ttl) * time.Second)
	ok, err := r.bds.SetNX(r.ctx, key, timeout, int(r.ttl+5))

	// if ok is true, then the lock is acquired by the caller.
	if err == nil && ok {
		r.stat.IncAcquired()
		return true
	}

	if err != nil {
		logs.Errorf("acquire redis lock %s failed, setnx resource lock failed, err: %s", key, err.Error())
		return false
	}

	// check if the lock is expired
	str, err := r.bds.Get(r.ctx, key)
	if err != nil {
		logs.Errorf("try acquire redis lock %s failed, get resource lock failed, err: %s", key, err.Error())
		return false
	}

	if str == "" {
		logs.Warnf("acquire redis lock %s failed, resource lock not exist in redis", key)
		return false
	}

	oldTimeout, err := time.Parse(time.RFC3339, str)
	if err != nil {
		logs.Errorf("try acquire redis lock %s failed, parse timeout failed, err: %s", key, err.Error())
		return false
	}

	if time.Now().After(oldTimeout) {
		oldStr, err := r.bds.GetSet(r.ctx, key, timeout.Format(time.RFC3339))
		if err != nil {
			logs.Errorf("try acquire redis lock %s failed, getset resource lock failed, err: %s", key, err.Error())
			return false
		}
		oldTimeout, err := time.Parse(time.RFC3339, oldStr)
		if err != nil {
			logs.Errorf("try acquire redis lock %s failed, parse timeout failed, err: %s", key, err.Error())
			return false
		}
		if time.Now().After(oldTimeout) {
			logs.Warnf("try acquire redis lock %s success, resource lock is expired", key)
			r.stat.IntTimeout()
			r.stat.IncAcquired()
			return true
		}
	}

	return false
}

// Release the resource lock so that the resource lock is released.
// after the resource's lock is released, the other call which is still
// waiting for the lock to be released will return immediately.
func (r *RedisLock) Release(res string) {

	key := fmt.Sprintf("Lock:%s", res)

	str, err := r.bds.Get(r.ctx, key)
	if err != nil {
		logs.Errorf("release redis lock %s failed, get resource lock failed, err: %s", key, err.Error())
		return
	}

	if str == "" {
		logs.Warnf("release redis lock %s failed, resource lock not exist in redis", key)
		return
	}

	timeout, err := time.Parse(time.RFC3339, str)
	if err != nil {
		logs.Errorf("release redis lock %s failed, parse timeout failed, err: %s", key, err.Error())
		return
	}
	if time.Now().Before(timeout) {
		if err = r.bds.Delete(r.ctx, key); err != nil {
			logs.Errorf("release redis lock %s failed, delete resource lock failed, err: %s", key, err.Error())
		}
		return
	}
	logs.Errorf("trying to release a timeout lock: %s", key)
}

func (r *RedisLock) collectMetric() {
	for {
		time.Sleep(5 * time.Second)
		r.mc.acquiredRate.With(prm.Labels{}).Set(r.stat.Rate())
	}
}

type redisLockStatistic struct {
	TimeOut  atomic.Int64
	Total    atomic.Int64
	Acquired atomic.Int64
	mc       *redisLockMetric
}

// IncTimeout increase the total by one.
func (sc *redisLockStatistic) IntTimeout() {
	sc.TimeOut.Inc()
	sc.mc.timeoutCounter.With(prm.Labels{}).Inc()
}

// IncTotal increase the total by one.
func (sc *redisLockStatistic) IncTotal() {
	sc.Total.Inc()
	sc.mc.totalCounter.With(prm.Labels{}).Inc()
}

// IncAcquired increase the acquired by one.
func (sc *redisLockStatistic) IncAcquired() {
	sc.Total.Inc()
	sc.mc.acquiredCounter.With(prm.Labels{}).Inc()
}

// Rate is the acquired rate of total.
func (sc *redisLockStatistic) Rate() float64 {
	if sc.Total.Load() == 0 {
		return 0
	}

	return float64(sc.Acquired.Load() / sc.Total.Load())
}
