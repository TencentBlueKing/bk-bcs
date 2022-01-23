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
	"fmt"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

const (
	// "BCS-WebConsole:{project_id}:{cluster_id}:{session_id}"
	cacheKeyTmpl = "BCS-WebConsole:{%s}:{%s}:{%s}"
)

// RedisStore github.com/gorilla/sessions/store.Store interface 实现
type RedisStore struct {
	client    *redis.Client
	projectId string
	clusterId string
}

func NewRedisStore(client *redis.Client, projectId, clusterId string) *RedisStore {
	return &RedisStore{client: client, projectId: projectId, clusterId: clusterId}
}

func (rs *RedisStore) cacheKey(id string) string {
	key := fmt.Sprintf(cacheKeyTmpl, rs.projectId, rs.clusterId, id)
	return key
}

func (rs *RedisStore) New(r *http.Request, id string) (*sessions.Session, error) {
	session := sessions.NewSession(rs, id)

	if id == "" {
		id = uuid.NewString()
	}

	session.IsNew = true
	session.ID = id

	return session, nil
}

func (rs *RedisStore) GetValue(r *http.Request, id, valueKey string) (string, error) {
	session, err := rs.Get(r, id)
	if err != nil {
		return "", err
	}
	value, ok := session.Values[valueKey]
	if !ok {
		return "", nil
	}
	v, ok := value.(string)
	if !ok {
		return "", nil
	}
	return v, nil
}

func (rs *RedisStore) GetValues(r *http.Request, id string) (map[string]string, error) {
	values := map[string]string{}
	session, err := rs.Get(r, id)
	if err != nil {
		return values, err
	}

	for k, v := range session.Values {
		values[k.(string)] = v.(string)
	}

	return values, nil
}

func (rs *RedisStore) Get(r *http.Request, id string) (*sessions.Session, error) {
	session, err := rs.New(r, id)
	if err != nil {
		return nil, err
	}
	session.IsNew = false

	values, err := rs.client.HGetAll(r.Context(), rs.cacheKey(id)).Result()
	if err != nil {
		return nil, err
	}

	for k, v := range values {
		session.Values[k] = v
	}

	return session, nil
}

// Save 保存数据到 Redis, 使用 HSET 数据结构
func (rs *RedisStore) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	_, err := rs.client.HSet(r.Context(), rs.cacheKey(s.ID), s.Values).Result()
	if err != nil {
		return err
	}
	return nil
}
