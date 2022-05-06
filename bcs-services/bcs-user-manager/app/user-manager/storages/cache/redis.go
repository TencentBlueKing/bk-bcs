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

package cache

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
	"github.com/go-redis/redis/v8"
)

// RDB is the global redis cache
var RDB Cache

// InitRedis init redis cache with url
func InitRedis(conf *config.UserMgrConfig) error {
	options, err := redis.ParseURL(conf.RedisDSN)
	if err != nil {
		return err
	}
	client := redis.NewClient(options)
	RDB = &redisCache{client: client}
	return nil
}

// Cache is the interface of redis cache
type Cache interface {
	Set(key string, value interface{}, expiration time.Duration) (string, error)
	SetNX(key string, value interface{}, expiration time.Duration) (bool, error)
	SetEX(key string, value interface{}, expiration time.Duration) (string, error)
	Get(key string) (string, error)
	Del(key string) (uint64, error)
	Expire(key string, expiration time.Duration) (bool, error)
}

var _ Cache = &redisCache{}

type redisCache struct {
	client *redis.Client
}

// Set implements Cache.Set
func (r *redisCache) Set(key string, value interface{}, expiration time.Duration) (string, error) {
	return r.client.Set(context.TODO(), key, value, expiration).Result()
}

// SetNX implements Cache.SetNX
func (r *redisCache) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(context.TODO(), key, value, expiration).Result()
}

// SetEX implements Cache.SetEX
func (r *redisCache) SetEX(key string, value interface{}, expiration time.Duration) (string, error) {
	return r.client.SetEX(context.TODO(), key, value, expiration).Result()
}

// Get implements Cache.Get
func (r *redisCache) Get(key string) (string, error) {
	return r.client.Get(context.TODO(), key).Result()
}

// Del implements Cache.Del
func (r *redisCache) Del(key string) (uint64, error) {
	return r.client.Del(context.TODO(), key).Uint64()
}

// Expire implements Cache.Expire
func (r *redisCache) Expire(key string, expiration time.Duration) (bool, error) {
	return r.client.Expire(context.TODO(), key, expiration).Result()
}
