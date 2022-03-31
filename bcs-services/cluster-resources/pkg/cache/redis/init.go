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
	"runtime"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	crRuntime "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
)

var rds *redis.Client

var initOnce sync.Once

const (
	// dialTimeout 单位：s
	dialTimeout = 2
	// readTimeout 单位：s
	readTimeout = 1
	// writeTimeout 单位：s
	writeTimeout = 1
	// pollSizeMultiple * NumCPU
	pollSizeMultiple = 20
	// minIdleConnsMultiple * NumCPU
	minIdleConnsMultiple = 10
	// idleTimeout unit: min
	idleTimeout = 3
)

// newStandaloneClient 创建单实例模式 RedisClient（非哨兵模式）
func newStandaloneClient(conf *config.RedisConf) *redis.Client {
	opt := &redis.Options{
		Addr:     conf.Address,
		Password: conf.Password,
		DB:       conf.DB,
	}

	// 默认配置
	opt.DialTimeout = time.Duration(dialTimeout) * time.Second
	opt.ReadTimeout = time.Duration(readTimeout) * time.Second
	opt.WriteTimeout = time.Duration(writeTimeout) * time.Second
	opt.PoolSize = pollSizeMultiple * runtime.NumCPU()
	opt.MinIdleConns = minIdleConnsMultiple * runtime.NumCPU()
	opt.IdleTimeout = time.Duration(idleTimeout) * time.Minute

	// 若配置中指定，则使用
	if conf.DialTimeout > 0 {
		opt.DialTimeout = time.Duration(conf.DialTimeout) * time.Second
	}
	if conf.ReadTimeout > 0 {
		opt.ReadTimeout = time.Duration(conf.ReadTimeout) * time.Second
	}
	if conf.WriteTimeout > 0 {
		opt.WriteTimeout = time.Duration(conf.WriteTimeout) * time.Second
	}
	if conf.PoolSize > 0 {
		opt.PoolSize = conf.PoolSize
	}
	if conf.MinIdleConns > 0 {
		opt.MinIdleConns = conf.MinIdleConns
	}

	log.Info(context.TODO(), "start connect redis: %s [db=%d, dialTimeout=%s, readTimeout=%s, writeTimeout=%s, poolSize=%d, minIdleConns=%d, idleTimeout=%s]", //nolint:lll
		opt.Addr, opt.DB, opt.DialTimeout, opt.ReadTimeout, opt.WriteTimeout, opt.PoolSize, opt.MinIdleConns, opt.IdleTimeout)

	return redis.NewClient(opt)
}

// InitRedisClient 初始化 Redis 客户端
func InitRedisClient(conf *config.RedisConf) {
	if rds != nil {
		return
	}
	initOnce.Do(func() {
		rds = newStandaloneClient(conf)
		// 若 Redis 服务异常，应重置 rds 并 panic
		if _, err := rds.Ping(context.TODO()).Result(); err != nil {
			rds = nil
			panic(err)
		}
	})
}

// GetDefaultClient 获取默认 Redis 客户端
func GetDefaultClient() *redis.Client {
	if rds == nil {
		// 单元测试模式下，自动启用测试用 Redis，否则需要提前初始化
		if crRuntime.RunMode == runmode.UnitTest {
			rds = NewTestRedisClient()
			return rds
		}
		panic("prod and stag run mode need init redis!")
	}
	return rds
}
