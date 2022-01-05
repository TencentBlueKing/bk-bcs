/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"

	crCache "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

const (
	// go-redis/cache 版本升级可能存在不兼容先前版本的情况，例如 v7 Set 的对象无法通过 v8 读取
	// https://github.com/go-redis/cache/issues/52
	// NOTE: 【重要】若升级 go-redis/cache 版本，需要同步更新 CacheVersion，确保现有缓存失效

	// CacheVersion 取值建议为循环 00->99->00
	CacheVersion = "00"

	// CacheKeyPrefix 标识模块名的 Cache Key 前缀
	CacheKeyPrefix = "bcs-services-cr"
)

// Cache ...
type Cache struct {
	name              string
	keyPrefix         string
	codec             *cache.Cache
	cli               *redis.Client
	defaultExpiration time.Duration
	G                 singleflight.Group
}

// NewCache 新建 cache 实例
func NewCache(name string, expiration time.Duration) *Cache {
	cli := GetDefaultRedisClient()

	// key: {cache_key_prefix}:{version}:{cache_name}:{raw_key}
	keyPrefix := fmt.Sprintf("%s:%s:%s", CacheKeyPrefix, CacheVersion, name)

	codec := cache.New(&cache.Options{
		Redis: cli,
	})

	return &Cache{
		name:              name,
		keyPrefix:         keyPrefix,
		codec:             codec,
		cli:               cli,
		defaultExpiration: expiration,
	}
}

// NewMockCache 新建 mock cache 实例
func NewMockCache(name string, expiration time.Duration) *Cache {
	cli := util.NewTestRedisClient()

	// key: {cache_key_prefix}:{cache_name}:{raw_key}
	keyPrefix := fmt.Sprintf("%s:%s", CacheKeyPrefix, name)

	codec := cache.New(&cache.Options{
		Redis: cli,
	})

	return &Cache{
		name:              name,
		keyPrefix:         keyPrefix,
		codec:             codec,
		cli:               cli,
		defaultExpiration: expiration,
	}
}

func (c *Cache) genKey(key string) string {
	return c.keyPrefix + ":" + key
}

// Set ...
func (c *Cache) Set(key crCache.Key, value interface{}, duration time.Duration) error {
	if duration == time.Duration(0) {
		duration = c.defaultExpiration
	}

	k := c.genKey(key.Key())
	return c.codec.Set(&cache.Item{
		Key:   k,
		Value: value,
		TTL:   duration,
	})
}

// Exists ...
func (c *Cache) Exists(key crCache.Key) bool {
	k := c.genKey(key.Key())
	count, err := c.cli.Exists(context.TODO(), k).Result()
	return err == nil && count == 1
}

// Get ...
func (c *Cache) Get(key crCache.Key, value interface{}) error {
	k := c.genKey(key.Key())
	return c.codec.Get(context.TODO(), k, value)
}

// Delete ...
func (c *Cache) Delete(key crCache.Key) error {
	k := c.genKey(key.Key())
	_, err := c.cli.Del(context.TODO(), k).Result()
	return err
}
