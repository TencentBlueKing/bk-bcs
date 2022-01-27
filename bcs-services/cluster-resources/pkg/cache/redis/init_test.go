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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

func TestInitRedisClient(t *testing.T) {
	// 不存在的 Redis 服务，应当如预期出现 panic
	redisConfig := &config.RedisConf{
		Address:      "1.0.0.1:6379",
		DialTimeout:  2,
		ReadTimeout:  1,
		WriteTimeout: 1,
		PoolSize:     2,
		MinIdleConns: 4,
	}
	defer func() {
		err := recover()
		assert.Error(t, err.(error))
	}()
	InitRedisClient(redisConfig)
}

func TestGetDefaultClient(t *testing.T) {
	// 单元测试模式下，没有初始化会有默认的
	rdsCli := GetDefaultClient()
	assert.NotNil(t, rdsCli)

	ret, _ := rds.Ping(context.TODO()).Result()
	assert.Equal(t, "PONG", ret)
}
