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

package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/lock"

	"github.com/google/uuid"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// JitterFactor jitter factor for retry acquire lock
	JitterFactor = 0.5
	// DefaultTimeoutDuration default timeout duration for acquiring or renewing lock
	DefaultTimeoutDuration = 3 * time.Second
	// DefaultRenewDuration default duration for renewing lock
	DefaultRenewDuration = 3 * time.Second
	// DefaultRetryDuration default duration for retrying acquiring lock
	DefaultRetryDuration = 1 * time.Second
)

type kubeConfigKey struct{}

type kubeLock struct {
	SessionID string
	stopCh    chan struct{}
}

// KubeLockerConfig set kubernetes locker config
func KubeLockerConfig(config *LockerConfig) lock.Option {
	return func(o *lock.Options) {
		if o.Ctx == nil {
			o.Ctx = context.Background()
		}
		o.Ctx = context.WithValue(o.Ctx, kubeConfigKey{}, config)
	}
}

// LockerConfig configs for initializing locker
type LockerConfig struct {
	// LockerName locker name
	LockerName string
	// Prefix prefix for store
	Prefix string
	// Namespace namespace for store
	Namespace string
	// Kubeconfig kubeconfig for store
	Kubeconfig string
	// TimeoutDuration timeout duration for acquire or renew lock
	TimeoutDuration time.Duration
	// RenewDuration renew interval for renew the lock lease
	RenewDuration time.Duration
	// RetryDuration retry interval for acquire the lock
	RetryDuration time.Duration
}

// Locker is locker for kubernetes
type Locker struct {
	lockerName      string
	store           Store
	timeoutDuration time.Duration
	renewDuration   time.Duration
	retryDuration   time.Duration

	mtx   sync.Mutex
	locks map[string]*kubeLock
}

// New create locker
func New(opts ...lock.Option) (*Locker, error) {
	var options lock.Options
	for _, o := range opts {
		o(&options)
	}
	if options.Ctx == nil {
		return nil, fmt.Errorf("kube locker config cannot be empty")
	}
	cfg, ok := options.Ctx.Value(kubeConfigKey{}).(*LockerConfig)
	if !ok {
		return nil, fmt.Errorf("get kube locker config from context failed")
	}
	if cfg == nil {
		return nil, fmt.Errorf("locker config cannot be empty")
	}
	locker := &Locker{
		locks: make(map[string]*kubeLock),
	}
	// do configure
	if err := locker.configure(cfg); err != nil {
		return nil, err
	}
	return locker, nil
}

// configure do configure for kube lock
func (l *Locker) configure(cfg *LockerConfig) error {
	cmStore, err := NewConfigmapStore(cfg.Prefix, cfg.Namespace, cfg.Kubeconfig)
	if err != nil {
		return fmt.Errorf("create configmap store failed, err %s", err.Error())
	}
	l.store = cmStore
	if cfg.TimeoutDuration <= 0 {
		l.timeoutDuration = DefaultTimeoutDuration
	}
	if cfg.RenewDuration <= 0 {
		l.renewDuration = DefaultRenewDuration
	}
	if cfg.RetryDuration <= 0 {
		l.retryDuration = DefaultRetryDuration
	}
	return nil
}

// Lock acquires a lock
func (l *Locker) Lock(key string, opts ...lock.LockOption) error {
	var options lock.LockOptions
	for _, o := range opts {
		o(&options)
	}
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), l.timeoutDuration)
	defer timeoutCancel()
	if !l.acquire(timeoutCtx, key, options) {
		return fmt.Errorf("acquire lock for key %s failed", key)
	}
	go l.renewLoop(context.Background(), key, options)
	return nil
}

// Unlock releases a lock
func (l *Locker) Unlock(key string) error {
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), l.timeoutDuration)
	defer timeoutCancel()
	return l.release(timeoutCtx, key)
}

func (l *Locker) acquire(ctx context.Context, key string, opt lock.LockOptions) bool {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	succeeded := false
	var lr *LockRecord
	sessionID := l.lockerName + "-" + uuid.New().String()
	blog.V(5).Infof("locker %s attempting to acquire lease for %s with session %s", l.lockerName, key, sessionID)
	wait.JitterUntil(func() {
		lr, succeeded = l.tryAcquireOrRenew(ctx, key, sessionID, opt)
		if !succeeded {
			return
		}
		blog.V(5).Infof("locker %s successfully acquired lease for %s, lock record %v", l.lockerName, key, lr)
		l.mtx.Lock()
		l.locks[key] = &kubeLock{
			SessionID: sessionID,
			stopCh:    make(chan struct{}, 5),
		}
		l.mtx.Unlock()
		cancel()
	}, l.retryDuration, JitterFactor, true, ctx.Done())
	return succeeded
}

func (l *Locker) renewLoop(ctx context.Context, key string, opt lock.LockOptions) {
	ticker := time.NewTicker(l.renewDuration)
	l.mtx.Lock()
	kLock, ok := l.locks[key]
	l.mtx.Unlock()
	if !ok {
		blog.Warnf("locker %s lock with key %s not existed", l.lockerName, key)
		return
	}
	sessionID := kLock.SessionID
	for {
		select {
		case <-ticker.C:
			var lr *LockRecord
			var succeeded bool
			timeoutCtx, timeoutCancel := context.WithTimeout(ctx, l.timeoutDuration)
			defer timeoutCancel()
			err := wait.PollImmediateUntil(l.retryDuration, func() (bool, error) {
				lr, succeeded = l.tryAcquireOrRenew(timeoutCtx, key, sessionID, opt)
				return succeeded, nil
			}, timeoutCtx.Done())
			if err == nil {
				blog.V(5).Infof("locker %s successfully renewed lease for key %s, lock record %v",
					l.lockerName, key, lr)
				continue
			}
			blog.V(5).Infof("locker %s failed to renew lease for key %s, err %s", l.lockerName, key, err.Error())
			l.mtx.Lock()
			defer l.mtx.Unlock()
			_, ok = l.locks[key]
			if !ok {
				blog.Warnf("locker %s lock with key %s not existed when stop renew loop", l.lockerName, key)
				return
			}
			delete(l.locks, key)
			return
		case <-kLock.stopCh:
			l.mtx.Lock()
			defer l.mtx.Unlock()
			_, ok = l.locks[key]
			if !ok {
				blog.Warnf("locker %s lock with key %s not existed when stop renew loop", l.lockerName, key)
				return
			}
			delete(l.locks, key)
			return
		}
	}
}

func (l *Locker) release(ctx context.Context, key string) error {
	// ask renew loop to exit
	l.mtx.Lock()
	kLock, ok := l.locks[key]
	l.mtx.Unlock()
	if !ok {
		blog.Warnf("locker %s release lock with key %s not existed", l.lockerName, key)
		return nil
	}

	kLock.stopCh <- struct{}{}
	// get lock record
	lr, _, err := l.store.Get(ctx, key)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("locker %s get lock record by key %s failed when release, err %s",
			l.lockerName, key, err.Error())
	}
	lrToUpdate := LockRecord{
		ResourceVersion: lr.ResourceVersion,
	}

	// update the lock
	_, err = l.store.Update(ctx, key, lrToUpdate)
	if err != nil {
		return fmt.Errorf("locker %s update lock %s to null failed, err %s", l.lockerName, key, err.Error())
	}
	blog.V(5).Infof("locker %s successfully release lease for key %s", l.lockerName, key)
	return nil
}

func (l *Locker) tryAcquireOrRenew(ctx context.Context, key, sessionID string, opt lock.LockOptions) (
	*LockRecord, bool) {
	lr, _, err := l.store.Get(ctx, key)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, false
		}
		createTime := metav1.Now()
		lrToCreate := LockRecord{
			OwnerID:        sessionID,
			ExpireDuration: opt.TTL,
			AcquireTime:    createTime,
			RenewTime:      createTime,
		}
		newLr, createErr := l.store.Create(ctx, key, lrToCreate)
		if createErr != nil {
			blog.Warnf("locker %s failed to create lock record %v, err %s", l.lockerName, lr, err.Error())
			return nil, false
		}
		return newLr, true
	}
	now := metav1.Now()
	lrToUpdate := LockRecord{
		OwnerID:         sessionID,
		ExpireDuration:  opt.TTL,
		RenewTime:       now,
		ResourceVersion: lr.ResourceVersion,
	}
	if lr.OwnerID != sessionID {
		if lr.RenewTime.Add(lr.ExpireDuration).After(now.Time) {
			blog.V(5).Infof("locker %s lock %s held by %s not expired", l.lockerName, key, lr.OwnerID)
			return nil, false
		}
	} else {
		lrToUpdate.AcquireTime = lr.AcquireTime
	}
	// update the lock
	updatedLr, err := l.store.Update(ctx, key, lrToUpdate)
	if err != nil {
		blog.Warnf("locker %s update lock %s to %v failed, err %s", l.lockerName, key, lrToUpdate, err.Error())
		return nil, false
	}
	return updatedLr, true
}
