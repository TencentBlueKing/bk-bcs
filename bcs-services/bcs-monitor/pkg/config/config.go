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

package config

import (
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Configuration 配置
type Configuration struct {
	mtx         sync.Mutex
	Base        *BaseConf                  `yaml:"base_conf"`
	Redis       *RedisConf                 `yaml:"redis"`
	StoreGW     *StoreGWConf               `yaml:"store"`
	API         *APIConf                   `yaml:"query"`
	Logging     *LogConf                   `yaml:"logging"`
	BKAPIGW     *BKAPIGWConf               `yaml:"bkapigw_conf"`
	BCS         *BCSConf                   `yaml:"bcs_conf"`
	Credentials map[string][]*Credential   `yaml:"-"`
	BCSEnvConf  []*BCSConf                 `yaml:"bcs_env_conf"`
	BCSEnvMap   map[BCSClusterEnv]*BCSConf `yaml:"-"`
	Web         *WebConf                   `yaml:"web"`
}

// init 初始化
func (c *Configuration) init() error {
	if err := c.Web.init(); err != nil {
		return err
	}

	if err := c.Logging.init(); err != nil {
		return err
	}

	return nil
}

// newConfigurations 新增配置
func newConfiguration() (*Configuration, error) {
	c := &Configuration{}

	c.Base = &BaseConf{}
	if err := c.Base.Init(); err != nil {
		return nil, err
	}

	c.StoreGW = &StoreGWConf{}
	if err := c.StoreGW.Init(); err != nil {
		return nil, err
	}

	c.API = &APIConf{}
	if err := c.API.Init(); err != nil {
		return nil, err
	}

	c.Redis = DefaultRedisConf()

	c.Logging = defaultLogConf()
	c.Web = defaultWebConf()

	c.Credentials = map[string][]*Credential{}
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
	return c.ReadFrom(content)
}
