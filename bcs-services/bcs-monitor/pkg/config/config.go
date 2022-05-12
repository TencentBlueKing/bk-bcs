/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-监控平台 (Blueking - Monitor) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */
package config

import (
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// UnmarshalKey 从 viper 配置中反序列化为对象
func UnmarshalKey(key string, out interface{}) error {
	conf := viper.GetStringMap(key)
	ConfBytes, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(ConfBytes, out)
}

type Configuration struct {
	Base       *BaseConf                  `yaml:"base_conf"`
	Redis      *RedisConf                 `yaml:"redis"`
	StoreGW    *StoreGWConf               `yaml:"store"`
	API        *APIConf                   `yaml:"query"`
	BKAPIGW    *BKAPIGWConf               `yaml:"bkapigw_conf"`
	BCS        *BCSConf                   `yaml:"bcs_conf"`
	BCSEnvConf []*BCSConf                 `yaml:"bcs_env_conf"`
	BCSEnvMap  map[BCSClusterEnv]*BCSConf `yaml:"-"`
	Web        *WebConf                   `yaml:"web"`
}

func (c *Configuration) init() error {
	if err := c.Web.init(); err != nil {
		return err
	}

	return nil
}

// newConfigurations 新增配置
func newConfiguration() (*Configuration, error) {
	c := &Configuration{}

	c.Base = &BaseConf{}
	c.Base.Init()

	c.StoreGW = &StoreGWConf{}
	c.StoreGW.Init()

	c.API = &APIConf{}
	c.API.Init()

	c.Redis = DefaultRedisConf()

	c.Web = defaultWebConf()
	return c, nil
}

// G : Global Configurations
var G *Configuration

// 初始化
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

// ReadFrom : read from file
func (c *Configuration) ReadFrom(content []byte) error {
	return yaml.Unmarshal(content, &G)
}
