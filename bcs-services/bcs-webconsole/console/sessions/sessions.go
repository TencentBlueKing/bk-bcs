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

package sessions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	// bcs::webconsole::sessions::{run_env}
	keyPrefix      = "bcs::webconsole::sessions::%s"
	fieldKeyPrefix = "%s:%s:%s" // "{project_id}:{cluster_id}:{session_id}"
	expireDuration = time.Minute * 30
)

type RedisStore struct {
	client    *redis.Client
	projectId string
	clusterId string
	Id        string
	key       string
}

func NewRedisStore(projectId, clusterId string) *RedisStore {
	redisClient := storage.GetDefaultRedisSession().Client
	key := fmt.Sprintf(keyPrefix, config.G.Base.RunEnv)
	return &RedisStore{client: redisClient, projectId: projectId, clusterId: clusterId, key: key}
}

func (rs *RedisStore) cacheKey(id string) string {
	key := fmt.Sprintf(fieldKeyPrefix, rs.projectId, rs.clusterId, id)
	return key
}

func (rs *RedisStore) Get(ctx context.Context, id string) (*types.PodContext, error) {
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

// Save 保存数据到 Redis, 使用 HSET 数据结构
func (rs *RedisStore) Set(ctx context.Context, values *types.PodContext) (string, error) {
	podCtx := types.TimestampPodContext{
		PodContext: *values,
		Timestamp:  time.Now().Unix(),
	}
	id := uuid.NewString()
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
func (rs *RedisStore) Cleanup(ctx context.Context) error {
	values, err := rs.client.HGetAll(ctx, rs.key).Result()
	if err != nil {
		return nil
	}
	var expireSessions []string

	expireTimestamp := time.Now().Add(-expireDuration).Unix()

	for key, value := range values {
		var podCtx types.TimestampPodContext
		err := json.Unmarshal([]byte(value), &podCtx)
		if err == nil && podCtx.Timestamp >= expireTimestamp {
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
