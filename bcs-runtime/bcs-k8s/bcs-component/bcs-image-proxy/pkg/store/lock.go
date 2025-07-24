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

package store

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisLock defines the instance of redis lock
type RedisLock struct {
	client     *redis.Client
	lockKey    string
	lockValue  string
	expiration time.Duration
	stopChan   chan struct{}
}

// NewRedisLock 创建新的 Redis 锁
func NewRedisLock(client *redis.Client, lockKey, lockValue string, expiration time.Duration) *RedisLock {
	return &RedisLock{
		client:     client,
		lockKey:    lockKey,
		lockValue:  lockValue,
		expiration: expiration,
		stopChan:   make(chan struct{}),
	}
}

// Acquire 获取锁
func (lock *RedisLock) Acquire(ctx context.Context) (bool, error) {
	ok, err := lock.client.SetNX(ctx, lock.lockKey, lock.lockValue, lock.expiration).Result()
	if err != nil {
		return false, err
	}
	if ok {
		go lock.heartbeat(ctx)
	}
	return ok, nil
}

// Release 释放锁
func (lock *RedisLock) Release(ctx context.Context) error {
	close(lock.stopChan)
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)
	_, err := script.Run(ctx, lock.client, []string{lock.lockKey}, lock.lockValue).Result()
	return err
}

// heartbeat 心跳机制，定期更新锁的过期时间
func (lock *RedisLock) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(lock.expiration / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lock.client.Expire(ctx, lock.lockKey, lock.expiration)
		case <-lock.stopChan:
			return
		}
	}
}
