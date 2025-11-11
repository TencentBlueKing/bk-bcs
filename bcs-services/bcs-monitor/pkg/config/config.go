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
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// 子模块前缀
const (
	APIServicePrefix   = "/api"
	QueryServicePrefix = "/query"
)

// Configuration 配置
type Configuration struct {
	Viper       *viper.Viper
	mtx         sync.Mutex
	Base        *BaseConf                `yaml:"base_conf"`
	Redis       *RedisConf               `yaml:"redis"`
	Mongo       *MongoConf               `yaml:"mongo"`
	StoreGWList []*StoreConf             `yaml:"storegw"`
	Logging     *conf.LogConfig          `yaml:"logging"`
	BKAPIGW     *BKAPIGWConf             `yaml:"bkapigw_conf"`
	BKMonitor   *BKMonitorConf           `yaml:"bk_monitor_conf"`
	BKLog       *BKLogConf               `yaml:"bk_log_conf"`
	BKBase      *BKBaseConf              `yaml:"bk_base_conf"`
	BKUser      *BKUserConf              `yaml:"bk_user_conf"`
	BCS         *BCSConf                 `yaml:"bcs_conf"`
	IAM         *IAMConfig               `yaml:"iam_conf"`
	Credentials map[string][]*Credential `yaml:"-"`
	Web         *WebConf                 `yaml:"web"`
	QueryStore  *QueryStoreConf          `yaml:"query_store_conf"`
	TracingConf *TracingConf             `yaml:"tracing_conf"`
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

	c.Credentials = map[string][]*Credential{}

	c.IAM = &IAMConfig{}

	c.BKAPIGW = &BKAPIGWConf{}
	_ = c.BKAPIGW.Init()

	// BCS Config
	c.BCS = &BCSConf{}
	_ = c.BCS.Init()

	c.QueryStore = &QueryStoreConf{}

	c.BKMonitor = defaultBKMonitorConf()

	c.BKLog = &BKLogConf{}
	c.BKBase = &BKBaseConf{}
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

// IsDevMode 是否本地开发模式
func (c *Configuration) IsDevMode() bool {
	return c.Base.RunEnv == DevEnv
}

// ReadCredViper 使用 viper 读取配置
func (c *Configuration) ReadCredViper(name string, v *viper.Viper) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	cred := []*Credential{}

	// 使用 yaml tag 反序列化
	opt := viper.DecoderConfigOption(func(decoderConfig *mapstructure.DecoderConfig) {
		decoderConfig.TagName = "yaml"
	})

	if err := v.UnmarshalKey("credentials", &cred, opt); err != nil {
		return err
	}

	c.Credentials[name] = cred
	for _, v := range c.Credentials[name] {
		if err := v.InitCred(); err != nil {
			return err
		}
	}
	return nil
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

	// bklog
	if c.BKLog.APIServer == "" {
		c.BKLog.APIServer = BKLOG_API_SERVER
	}
	if c.BKUser.APIServer == "" {
		c.BKUser.APIServer = BKUSER_API_SERVER
	}
	if c.BKLog.Entrypoint == "" {
		c.BKLog.Entrypoint = BK_LOG_HOST
	}
	if c.BKLog.BKBaseEntrypoint == "" {
		c.BKLog.BKBaseEntrypoint = BK_BASE_HOST
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
