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

// Package storage init
package storage

// once synchronization
import (
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// Globals
var (
	GlobalRedisSession *RedisSession
	GlobalStorage      Storage
)

var (
	redisOnce   sync.Once
	storageOnce sync.Once
)

// GetDefaultRedisSession : get default redis session for default database
func GetDefaultRedisSession() *RedisSession {
	if GlobalRedisSession == nil {
		redisOnce.Do(func() {
			session := &RedisSession{}
			err := session.Init()
			if err != nil {
				panic(err)
			}
			GlobalRedisSession = session
		})
	}
	return GlobalRedisSession
}

// InitStorage init storage client
func InitStorage() {
	storageOnce.Do(func() {
		mongoConf := config.G.Mongo
		mongoOptions := &mongo.Options{
			Hosts:                 strings.Split(mongoConf.Address, ","),
			Replicaset:            mongoConf.Replicaset,
			ConnectTimeoutSeconds: int(mongoConf.ConnectTimeout),
			AuthDatabase:          mongoConf.AuthDatabase,
			Database:              mongoConf.Database,
			Username:              mongoConf.Username,
			Password:              mongoConf.Password,
			MaxPoolSize:           uint64(mongoConf.MaxPoolSize),
			MinPoolSize:           uint64(mongoConf.MinPoolSize),
			Monitor:               otelmongo.NewMonitor(),
		}
		mongoDB, err := mongo.NewDB(mongoOptions)
		if err != nil {
			blog.Errorf("init mongo db failed, err %s", err.Error())
			return
		}
		if err = mongoDB.Ping(); err != nil {
			blog.Errorf("ping mongo db failed, err %s", err.Error())
			return
		}
		blog.Info("init mongo db successfully")
		GlobalStorage = New(mongoDB)
	})
}
