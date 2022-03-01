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

	crCache "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cache"
)

const (
	// CacheKeyPrefix 标识模块名的 Cache Key 前缀
	CacheKeyPrefix = "bcs-services-cr"
)

// Cache ...
type Cache struct {
	name      string        // 缓存键名
	keyPrefix string        // 缓存键前缀
	codec     *cache.Cache  // go-redis cache
	cli       *redis.Client // redis client
	exp       time.Duration // 默认过期时间
}

// NewCache 新建 cache 实例
func NewCache(name string, expiration time.Duration) *Cache {
	cli := GetDefaultClient()

	// key: {cache_key_prefix}:{version}:{cache_name}:{raw_key}
	keyPrefix := fmt.Sprintf("%s:%s", CacheKeyPrefix, name)

	codec := cache.New(&cache.Options{
		Redis: cli,
	})

	return &Cache{
		name:      name,
		keyPrefix: keyPrefix,
		codec:     codec,
		cli:       cli,
		exp:       expiration,
	}
}

func (c *Cache) genKey(key string) string {
	return c.keyPrefix + ":" + key
}

// Set ...
func (c *Cache) Set(key crCache.Key, value interface{}, duration time.Duration) error {
	if duration == time.Duration(0) {
		duration = c.exp
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

// DeleteByPrefix 根据键前缀删除缓存，慎用！
func (c *Cache) DeleteByPrefix(prefix string) error {
	ctx := context.TODO()
	iter := c.cli.Scan(ctx, 0, c.genKey(prefix)+"*", 0).Iterator()
	for iter.Next(ctx) {
		if err := c.cli.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}
