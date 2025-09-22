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

// Package config xxx
package config

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	// APIServicePrefix API 服务前缀
	APIServicePrefix = "/api"
)

// Configuration 配置
type Configuration struct {
	Viper       *viper.Viper
	Base        *BaseConf       `yaml:"base_conf"`
	Redis       *RedisConf      `yaml:"redis"`
	Mongo       *MongoConf      `yaml:"mongo"`
	Logging     *conf.LogConfig `yaml:"logging"`
	BKAPIGW     *BKAPIGWConf    `yaml:"bkapigw_conf"`
	BCS         *BCSConf        `yaml:"bcs_conf"`
	IAM         *IAMConfig      `yaml:"iam_conf"`
	Web         *WebConf        `yaml:"web"`
	TracingConf *TracingConf    `yaml:"tracing_conf"`
}

// init 初始化
func (c *Configuration) init() error {
	if err := c.Web.init(); err != nil {
		return err
	}

	if err := c.BCS.InitJWTPubKey(); err != nil {
		return err
	}

	if err := c.BKAPIGW.InitJWTPubKey(); err != nil {
		return err
	}

	return nil
}

// newConfiguration s 新增配置
func newConfiguration() (*Configuration, error) {
	c := &Configuration{}

	c.Base = &BaseConf{}
	if err := c.Base.Init(); err != nil {
		return nil, err
	}

	c.Redis = DefaultRedisConf()
	c.Mongo = DefaultMongoConf()

	c.Logging = defaultLogConf()
	c.Web = defaultWebConf()

	c.IAM = &IAMConfig{}

	c.BKAPIGW = &BKAPIGWConf{}
	_ = c.BKAPIGW.Init()

	// BCS Config
	c.BCS = &BCSConf{}
	_ = c.BCS.Init()

	c.TracingConf = &TracingConf{}

	return c, nil
}

// G : Global Configurations
var G *Configuration

// init 初始化
func init() {
	g, err := newConfiguration()
	if err != nil {
		panic(err)
	}
	if err := g.init(); err != nil {
		panic(err)
	}

	G = g
}

// ReadFrom : read from file
func (c *Configuration) ReadFrom(content []byte) error {
	if err := yaml.Unmarshal(content, c); err != nil {
		return err
	}

	// 添加环境变量
	if c.Base.AppCode == "" {
		c.Base.AppCode = BK_APP_CODE
	}
	if c.Base.AppSecret == "" {
		c.Base.AppSecret = BK_APP_SECRET
	}
	if c.Base.SystemID == "" {
		c.Base.SystemID = BK_SYSTEM_ID
	}
	if c.Base.BindAddress == "" {
		c.Base.BindAddress = POD_IP
	}
	if c.Redis.Password == "" {
		c.Redis.Password = REDIS_PASSWORD
	}
	if c.BCS.Token == "" {
		c.BCS.Token = BCS_APIGW_TOKEN
	}

	if c.BCS.JWTPubKey == "" {
		c.BCS.JWTPubKey = BCS_APIGW_PUBLIC_KEY
	}

	// iam env
	if c.IAM.GatewayServer == "" {
		c.IAM.GatewayServer = BKIAM_GATEWAY_SERVER
	}

	// mongo
	if c.Mongo.Address == "" {
		c.Mongo.Address = MONGO_ADDRESS
	}
	if c.Mongo.Replicaset == "" {
		c.Mongo.Replicaset = MONGO_REPLICASET
	}
	if c.Mongo.Username == "" {
		c.Mongo.Username = MONGO_USERNAME
	}
	if c.Mongo.Password == "" {
		c.Mongo.Password = MONGO_PASSWORD
	}

	if err := c.init(); err != nil {
		return err
	}
	return nil
}

// ReadFromViper : read from viper
func (c *Configuration) ReadFromViper(v *viper.Viper) error {
	// 不支持inline, 需要使用 yaml 库
	content, err := yaml.Marshal(v.AllSettings())
	if err != nil {
		return err
	}
	c.Viper = v
	return c.ReadFrom(content)
}
