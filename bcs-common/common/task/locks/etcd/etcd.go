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

// Package etcd implement the lock interface for etcd.
package etcd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/iface"
	"github.com/RichardKnop/machinery/v2/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	lockKey = "/machinery/v2/lock/%s"
)

var (
	// ErrLockFailed ..
	ErrLockFailed = errors.New("etcd lock: failed to acquire lock")
)

type etcdLock struct {
	ctx     context.Context
	client  *clientv3.Client
	retries int
}

// New ..
func New(ctx context.Context, conf *config.Config, retries int) (iface.Lock, error) {
	etcdConf := clientv3.Config{
		Endpoints:   []string{conf.Lock},
		Context:     ctx,
		DialTimeout: time.Second * 5,
		TLS:         conf.TLSConfig,
	}

	client, err := clientv3.New(etcdConf)
	if err != nil {
		return nil, err
	}

	lock := etcdLock{
		ctx:     ctx,
		client:  client,
		retries: retries,
	}

	return &lock, nil
}

// LockWithRetries lock with retries, if TTL is < 1s, the default 1s TTL will be used.
func (l *etcdLock) LockWithRetries(key string, unixTsToExpireNs int64) error {
	i := 0
	for ; i < l.retries; i++ {
		err := l.Lock(key, unixTsToExpireNs)
		if err == nil {
			// 成功拿到锁，返回
			return nil
		}

		log.DEBUG.Printf("acquired lock=%s failed, retries=%d, err=%s", key, i, err)
		time.Sleep(time.Millisecond * 100)
	}

	log.INFO.Printf("acquired lock=%s failed, retries=%d", key, i)
	return ErrLockFailed
}

// Lock If TTL is < 1s, the default 1s TTL will be used.
func (l *etcdLock) Lock(key string, unixTsToExpireNs int64) error {
	now := time.Now().UnixNano()
	expireTTL := time.Duration(unixTsToExpireNs - now)

	// etcd ttl单位是s,往上取整
	ttl := time.Duration(int(expireTTL.Seconds())) * time.Second
	if ttl < expireTTL {
		ttl += time.Second
	}

	// etcd 不能设置小于1s的ttl
	if ttl < time.Second {
		ttl = time.Second
	}

	s, err := concurrency.NewSession(l.client, concurrency.WithTTL(int(ttl.Seconds())))
	if err != nil {
		return err
	}
	defer s.Orphan()

	k := fmt.Sprintf(lockKey, strings.TrimRight(key, "/"))
	m := concurrency.NewMutex(s, k)

	ctx, cancel := context.WithTimeout(l.ctx, time.Second*5)
	defer cancel()

	// 阻塞等待锁
	if err := m.Lock(ctx); err != nil {
		_ = s.Close()
		if errors.Is(err, context.DeadlineExceeded) {
			return ErrLockFailed
		}
		return err
	}

	log.INFO.Printf("acquired lock=%s, duration=%s", key, ttl)
	return nil
}

// GetLockExpireNs 获取锁的过期时间
func GetLockExpireNs(duration time.Duration) int64 {
	return time.Now().Add(duration).UnixNano()
}
