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

// Package redis xxx
package redis

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/redisclient"
	redisotel2 "github.com/go-redis/redis/extra/redisotel/v8"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	crRuntime "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

var rds redisclient.Client

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
	// Sentinel mode
	Sentinel = "sentinel"
	// Cluster mode
	Cluster = "cluster"
	// Single mode
	Single = "single"
)

// NewRedisClient 根据配置创建不同模式的 RedisClient
func NewRedisClient(conf *config.RedisConf) (redisclient.Client, error) {
	clientConf := redisclient.Config{
		Addrs:    stringx.Split(conf.Address),
		DB:       conf.DB,
		Password: conf.Password,
		// 默认配置
		DialTimeout:  time.Duration(dialTimeout) * time.Second,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		PoolSize:     pollSizeMultiple * runtime.NumCPU(),
		MinIdleConns: minIdleConnsMultiple * runtime.NumCPU(),
		IdleTimeout:  time.Duration(idleTimeout) * time.Minute,
	}

	// 若配置中指定，则使用
	if conf.DialTimeout > 0 {
		clientConf.DialTimeout = time.Duration(conf.DialTimeout) * time.Second
	}
	if conf.ReadTimeout > 0 {
		clientConf.ReadTimeout = time.Duration(conf.ReadTimeout) * time.Second
	}
	if conf.WriteTimeout > 0 {
		clientConf.WriteTimeout = time.Duration(conf.WriteTimeout) * time.Second
	}
	if conf.PoolSize > 0 {
		clientConf.PoolSize = conf.PoolSize
	}
	if conf.MinIdleConns > 0 {
		clientConf.MinIdleConns = conf.MinIdleConns
	}

	switch conf.RedisMode {
	case Sentinel:
		clientConf.Mode = redisclient.SentinelMode
		clientConf.MasterName = conf.MasterName
	case Cluster:
		clientConf.Mode = redisclient.ClusterMode
	default:
		clientConf.Mode = redisclient.SingleMode
	}

	log.Info(context.TODO(),
		"start connect redis: %v [mode=%s db=%d, dialTimeout=%s, readTimeout=%s, writeTimeout=%s, poolSize=%d, minIdleConns=%d, idleTimeout=%s]", //nolint:lll
		clientConf.Addrs, clientConf.Mode, clientConf.DB, clientConf.DialTimeout, clientConf.ReadTimeout, clientConf.WriteTimeout, clientConf.PoolSize, clientConf.MinIdleConns, clientConf.IdleTimeout) //nolint:lll
	return redisclient.NewClient(clientConf)
}

// InitRedisClient 初始化 Redis 客户端
func InitRedisClient(conf *config.RedisConf) {
	if rds != nil {
		return
	}
	initOnce.Do(func() {
		var err error
		// 初始化失败，panic
		if rds, err = NewRedisClient(conf); err != nil {
			panic(err)
		}
		rds.GetCli().AddHook(redisotel2.NewTracingHook())
		// 若 Redis 服务异常，应重置 rds 并 panic
		if _, err = rds.Ping(context.TODO()); err != nil {
			rds = nil
			panic(err)
		}
	})
}

// GetDefaultClient 获取默认 Redis 客户端
func GetDefaultClient() redisclient.Client {
	if rds == nil {
		// 单元测试模式下，自动启用测试用 Redis，否则需要提前初始化
		if crRuntime.RunMode == runmode.UnitTest {
			var err error
			if rds, err = redisclient.NewTestClient(); err != nil {
				panic(err)
			}
			rds.GetCli().AddHook(redisotel2.NewTracingHook())
			return rds
		}
		panic("prod and stag run mode need init redis!")
	}
	return rds
}
