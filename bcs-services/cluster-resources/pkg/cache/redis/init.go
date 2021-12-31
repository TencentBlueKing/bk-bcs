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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/options"
)

var rds *redis.Client

var redisClientInitOnce sync.Once

const (
	// dialTimeout unit: s
	dialTimeout = 2
	// readTimeout unit: s
	readTimeout = 1
	// writeTimeout unit: s
	writeTimeout = 1
	// pollSizeMultiple * NumCPU
	pollSizeMultiple = 20
	// minIdleConnsMultiple * NumCPU
	minIdleConnsMultiple = 10
	// idleTimeout unit: min
	idleTimeout = 3
)

func newStandaloneClient(redisConf options.RedisConf) *redis.Client {
	opt := &redis.Options{
		Addr:     redisConf.Address,
		Password: redisConf.Password,
		DB:       redisConf.DB,
	}

	// 默认配置
	opt.DialTimeout = time.Duration(dialTimeout) * time.Second
	opt.ReadTimeout = time.Duration(readTimeout) * time.Second
	opt.WriteTimeout = time.Duration(writeTimeout) * time.Second
	opt.PoolSize = pollSizeMultiple * runtime.NumCPU()
	opt.MinIdleConns = minIdleConnsMultiple * runtime.NumCPU()
	opt.IdleTimeout = time.Duration(idleTimeout) * time.Minute

	// 若指定配置中指定，则使用
	if redisConf.DialTimeout > 0 {
		opt.DialTimeout = time.Duration(redisConf.DialTimeout) * time.Second
	}
	if redisConf.ReadTimeout > 0 {
		opt.ReadTimeout = time.Duration(redisConf.ReadTimeout) * time.Second
	}
	if redisConf.WriteTimeout > 0 {
		opt.WriteTimeout = time.Duration(redisConf.WriteTimeout) * time.Second
	}

	if redisConf.PoolSize > 0 {
		opt.PoolSize = redisConf.PoolSize
	}
	if redisConf.MinIdleConns > 0 {
		opt.MinIdleConns = redisConf.MinIdleConns
	}

	// TODO Add Redis Conn Info fmt.Println("connect to redis: "+
	//	"%s [db=%d, dialTimeout=%s, readTimeout=%s, writeTimeout=%s, poolSize=%d, minIdleConns=%d, idleTimeout=%s]",
	//	opt.Addr, opt.DB, opt.DialTimeout, opt.ReadTimeout, opt.WriteTimeout,
	//	opt.PoolSize, opt.MinIdleConns, opt.IdleTimeout)

	return redis.NewClient(opt)
}

// InitRedisClient ...
func InitRedisClient(debugMode bool, opts *options.ClusterResourcesOptions) {
	if rds != nil {
		return
	}
	redisClientInitOnce.Do(func() {
		//
		rds = newStandaloneClient(opts.Redis)
		_, err := rds.Ping(context.TODO()).Result()
		if err != nil {
			// redis is important
			if !debugMode {
				panic(err)
			}
		}
	})
}

// GetDefaultRedisClient 获取默认 Redis 实例
func GetDefaultRedisClient() *redis.Client {
	return rds
}
