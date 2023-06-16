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

// Package sessions xxx
package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

const (
	// bcs::webconsole::sessions::{run_env} 格式
	keyPrefix = "bcs::webconsole::sessions::%s"
	// "{scope}:{session_id}" 格式
	fieldKeyPrefix = "%s:%s"
)

// redisStore: redis 全局 session 存储
type redisStore struct {
	client *redis.Client
	scope  string
	key    string
}

// NewStore 新建 session 存储
// NOCC:golint/ret(设计如此:)
func NewStore() *redisStore {
	redisClient := storage.GetDefaultRedisSession().Client
	key := fmt.Sprintf(keyPrefix, config.G.Base.RunEnv)
	return &redisStore{client: redisClient, key: key, scope: "internal"}
}

// WebSocketScope 类型
func (rs *redisStore) WebSocketScope() *redisStore {
	rs.scope = "websocket"
	return rs
}

// OpenAPIScope xxx
// OpenAPI 类型
func (rs *redisStore) OpenAPIScope() *redisStore {
	rs.scope = "openapi"
	return rs
}

func (rs *redisStore) cacheKey(id string) string {
	key := fmt.Sprintf(fieldKeyPrefix, rs.scope, id)
	return key
}

// Get 读取 session
func (rs *redisStore) Get(ctx context.Context, id string) (*types.PodContext, error) {
	value, err := rs.client.HGet(ctx, rs.key, rs.cacheKey(id)).Result()
	if err != nil {
		return nil, err
	}

	var podCtx types.TimestampPodContext
	if err := json.Unmarshal([]byte(value), &podCtx); err != nil {
		return nil, err
	}

	return &podCtx.PodContext, nil
}

// Set 保存数据到 Redis, 使用 HSET 数据结构
func (rs *redisStore) Set(ctx context.Context, values *types.PodContext) (string, error) {
	podCtx := types.TimestampPodContext{
		PodContext: *values,
		Timestamp:  time.Now().Unix(),
	}
	id := sessionIdGenerator()
	payload, err := json.Marshal(podCtx)
	if err != nil {
		return "", err
	}
	if _, err := rs.client.HSet(ctx, rs.key, rs.cacheKey(id), payload).Result(); err != nil {
		return "", err
	}
	return id, nil
}

// Cleanup 清理过期数据
func (rs *redisStore) Cleanup(ctx context.Context) error {
	values, err := rs.client.HGetAll(ctx, rs.key).Result()
	if err != nil {
		return nil
	}
	var expireSessions []string

	for key, value := range values {
		var podCtx types.TimestampPodContext
		err := json.Unmarshal([]byte(value), &podCtx)
		if err == nil && !podCtx.IsExpired() {
			continue
		}

		expireSessions = append(expireSessions, key)
	}

	if len(expireSessions) > 0 {
		if _, err := rs.client.HDel(ctx, rs.key, expireSessions...).Result(); err != nil {
			return err
		}
	}

	return nil
}

// sessionIdGenerator xxx
func sessionIdGenerator() string {
	uid := uuid.New().String()
	requestId := strings.Replace(uid, "-", "", -1)
	return requestId
}
