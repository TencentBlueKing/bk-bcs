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

// Package cache xxx
package cache

import (
	"context"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/redisclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
	"github.com/go-redis/redis/extra/redisotel/v8"
)

// RDB is the global redis cache
var RDB Cache

// InitRedis init redis cache with url
func InitRedis(conf *config.UserMgrConfig) error {
	var client redisclient.Client
	var err error
	if conf.RedisDSN != "" {
		client, err = redisclient.NewSingleClientFromDSN(conf.RedisDSN)
	} else {
		redisConf := parseRedisConfig(conf)
		client, err = redisclient.NewClient(redisConf)
	}
	if err != nil {
		return err
	}

	client.GetCli().AddHook(redisotel.NewTracingHook())
	RDB = &redisCache{client: client}
	return nil
}

// parseRedisConfig parse Redis config
func parseRedisConfig(conf *config.UserMgrConfig) redisclient.Config {
	redisConf := redisclient.Config{
		Addrs:        strings.Split(conf.RedisConfig.Addr, ","),
		Password:     conf.RedisConfig.Password,
		DB:           conf.RedisConfig.DB,
		DialTimeout:  time.Duration(conf.RedisConfig.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(conf.RedisConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.RedisConfig.WriteTimeout) * time.Second,
		PoolSize:     conf.RedisConfig.PoolSize,
		MinIdleConns: conf.RedisConfig.MinIdleConns,
		IdleTimeout:  time.Duration(conf.RedisConfig.IdleTimeout) * time.Second,
	}

	// redis mode
	switch conf.RedisConfig.RedisMode {
	case "sentinel":
		redisConf.Mode = redisclient.SentinelMode
		redisConf.MasterName = conf.RedisConfig.MasterName
	case "cluster":
		redisConf.Mode = redisclient.ClusterMode
	default:
		redisConf.Mode = redisclient.SingleMode
	}

	return redisConf
}

// Cache is the interface of redis cache
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error)
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) (uint64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
}

var _ Cache = &redisCache{}

type redisCache struct {
	client redisclient.Client
}

// Set implements Cache.Set
func (r *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	return r.client.Set(ctx, key, value, expiration)
}

// SetNX implements Cache.SetNX
func (r *redisCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration)
}

// SetEX implements Cache.SetEX
func (r *redisCache) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) (
	string, error) {
	return r.client.SetEX(ctx, key, value, expiration)
}

// Get implements Cache.Get
func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key)
}

// Del implements Cache.Del
func (r *redisCache) Del(ctx context.Context, key string) (uint64, error) {
	count, err := r.client.Del(ctx, key)
	return uint64(count), err
}

// Expire implements Cache.Expire
func (r *redisCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return r.client.Expire(ctx, key, expiration)
}
