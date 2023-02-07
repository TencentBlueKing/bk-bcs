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
 *
 */

package storage

import (
	"net"
	"strconv"
	"time"

	redis "github.com/go-redis/redis/v8"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
)

// RedisSession :
type RedisSession struct {
	Client *redis.Client
}

// Init : init the redis client
func (r *RedisSession) Init() error {
	redisConf := config.G.Redis

	r.Client = redis.NewClient(&redis.Options{
		Addr:        net.JoinHostPort(redisConf.Host, strconv.Itoa(redisConf.Port)),
		Password:    redisConf.Password,
		DB:          redisConf.DB,
		DialTimeout: time.Duration(redisConf.MaxConnTimeout) * time.Second,
		ReadTimeout: time.Duration(redisConf.ReadTimeout) * time.Second,
		PoolSize:    redisConf.MaxPoolSize,
		IdleTimeout: time.Duration(redisConf.IdleTimeout) * time.Second,
	})
	return nil
}

// Close : close redis session
func (r *RedisSession) Close() {
	r.Client.Close()
}
