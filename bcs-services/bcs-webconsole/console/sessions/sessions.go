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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

const (
	// "BCS-WebConsole:{project_id}:{cluster_id}:{session_id}"
	cacheKeyTmpl = "BCS-WebConsole:%s:%s:%s"
)

type RedisStore struct {
	client    *redis.Client
	projectId string
	clusterId string
	Id        string
}

func NewRedisStore(projectId, clusterId string) *RedisStore {
	redisClient := storage.GetDefaultRedisSession().Client
	return &RedisStore{client: redisClient, projectId: projectId, clusterId: clusterId}
}

func (rs *RedisStore) cacheKey(id string) string {
	key := fmt.Sprintf(cacheKeyTmpl, rs.projectId, rs.clusterId, id)
	return key
}

func (rs *RedisStore) Get(ctx context.Context, id string) (*types.PodContext, error) {
	values, err := rs.client.Get(ctx, rs.cacheKey(id)).Result()
	if err != nil {
		return nil, err
	}
	var podCtx types.PodContext
	if err := json.Unmarshal([]byte(values), &podCtx); err != nil {
		return nil, err
	}

	return &podCtx, nil
}

// Save 保存数据到 Redis, 使用 HSET 数据结构
func (rs *RedisStore) Set(ctx context.Context, values *types.PodContext) (string, error) {
	id := uuid.NewString()
	payload, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	if _, err := rs.client.Set(ctx, rs.cacheKey(id), payload, time.Minute*30).Result(); err != nil {
		return "", err
	}
	return id, nil
}
