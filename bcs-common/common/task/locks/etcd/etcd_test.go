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

package etcd

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/RichardKnop/machinery/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLock(t *testing.T) {
	endpoints := os.Getenv("ETCDCTL_ENDPOINTS")
	if endpoints == "" {
		t.Skip("ETCDCTL_ENDPOINTS is not set")
	}
	t.Parallel()

	locker, err := New(context.Background(), &config.Config{Lock: endpoints}, 3)
	require.NoError(t, err)

	lockDuration := time.Second * 10
	err = locker.Lock("test_lock", int64(lockDuration))
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		st := time.Now()
		err = locker.Lock("test_lock", int64(lockDuration))
		assert.ErrorIs(t, err, ErrLockFailed)

		time.Sleep(time.Second * 5)
		err = locker.Lock("test_lock", int64(lockDuration))
		assert.NoError(t, err)
		duration := time.Since(st)
		assert.True(t, duration > lockDuration, "lock duration %s should be greater than %s", duration, lockDuration)
	}()
	wg.Wait()
}

func TestLockWithRetries(t *testing.T) {
	endpoints := os.Getenv("ETCDCTL_ENDPOINTS")
	if endpoints == "" {
		t.Skip("ETCDCTL_ENDPOINTS is not set")
	}
	t.Parallel()

	locker, err := New(context.Background(), &config.Config{Lock: endpoints}, 3)
	require.NoError(t, err)

	lockDuration := time.Second * 10
	err = locker.Lock("test_retry_lock", int64(lockDuration))
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		st := time.Now()
		err = locker.LockWithRetries("test_retry_lock", int64(lockDuration))
		assert.NoError(t, err)
		duration := time.Since(st)
		assert.True(t, duration > lockDuration, "lock duration %s should be greater than %s", duration, lockDuration)
	}()
	wg.Wait()
}

func TestLockWithMs(t *testing.T) {
	endpoints := os.Getenv("ETCDCTL_ENDPOINTS")
	if endpoints == "" {
		t.Skip("ETCDCTL_ENDPOINTS is not set")
	}
	t.Parallel()

	locker, err := New(context.Background(), &config.Config{Lock: endpoints}, 3)
	require.NoError(t, err)

	lockDuration := time.Millisecond * 10
	err = locker.Lock("test_retry_lock_ms", int64(lockDuration))
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		st := time.Now()
		err = locker.LockWithRetries("test_retry_lock_ms", int64(lockDuration))
		assert.NoError(t, err)
		duration := time.Since(st)
		assert.True(t, duration > lockDuration, "lock duration %s should be greater than %s", duration, lockDuration)
	}()
	wg.Wait()
}
