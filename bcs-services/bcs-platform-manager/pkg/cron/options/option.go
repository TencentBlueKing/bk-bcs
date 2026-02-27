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

// Package options xxx
package options

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// NewRedisConnOpt : create redis connection option
func NewRedisConnOpt() asynq.RedisConnOpt {
	redisConf := config.G.Redis

	if redisConf.IsSentinelType() {
		// 哨兵模式
		return asynq.RedisFailoverClientOpt{
			MasterName:       redisConf.MasterName,
			SentinelAddrs:    redisConf.SentinelAddrs,
			SentinelPassword: redisConf.SentinelPassword,
			Username:         "",
			Password:         redisConf.Password,
			DB:               redisConf.DB,
			DialTimeout:      time.Duration(redisConf.MaxConnTimeout) * time.Second,
			ReadTimeout:      time.Duration(redisConf.ReadTimeout) * time.Second,
			PoolSize:         redisConf.MaxPoolSize,
		}
	}
	// 单例模式
	return asynq.RedisClientOpt{
		Addr:        fmt.Sprintf("%v:%v", redisConf.Host, redisConf.Port),
		Password:    redisConf.Password,
		DB:          redisConf.DB,
		DialTimeout: time.Duration(redisConf.MaxConnTimeout) * time.Second,
		ReadTimeout: time.Duration(redisConf.ReadTimeout) * time.Second,
		PoolSize:    redisConf.MaxPoolSize,
	}
}
