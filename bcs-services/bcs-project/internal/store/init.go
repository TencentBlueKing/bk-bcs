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

package store

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
)

var mongoInitOnce sync.Once

var mongoDB *mongo.DB

// InitMongo 初始化 mongo
func InitMongo(conf *config.MongoConfig) {
	mongoInitOnce.Do(func() {
		mongoDB = NewMongo(conf)
	})
}

// NewMongo ...
func NewMongo(conf *config.MongoConfig) *mongo.DB {
	if len(conf.Address) == 0 {
		panic("mongo address cannot be empty")
	}
	if len(conf.Database) == 0 {
		panic("mongo database cannot be empty")
	}
	// 判断 password 是否加密，如果加密需要解密获取到原始数据
	// 使用 bcs service 统一的 pwd
	password := conf.Password
	if password != "" && conf.Encrypted {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
		password = string(realPwd)
	}

	mongoOptions := &mongo.Options{
		Hosts:                 strings.Split(conf.Address, ","),
		ConnectTimeoutSeconds: int(conf.ConnectTimeout),
		Database:              conf.Database,
		Username:              conf.Username,
		Password:              password,
		MaxPoolSize:           uint64(conf.MaxPoolSize),
		MinPoolSize:           uint64(conf.MinPoolSize),
	}

	db, err := mongo.NewDB(mongoOptions)
	if err != nil {
		panic(fmt.Sprintf("create mongo error, err: %s", err.Error()))
	}
	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("connect mongo error, err: %s", err.Error()))
	}
	return db
}

// GetMongo get mongo client
func GetMongo() *mongo.DB {
	return mongoDB
}
