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
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Configurations : manage all configurations
type Configurations struct {
	Base       *BaseConf                  `yaml:"base_conf"`
	Auth       *AuthConf                  `yaml:"auth_conf"`
	Logging    *LogConf                   `yaml:"logging"`
	BCS        *BCSConf                   `yaml:"bcs_conf"`
	BCSEnvConf []*BCSConf                 `yaml:"bcs_env_conf"`
	BCSEnvMap  map[BCSClusterEnv]*BCSConf `yaml:"-"`
	Redis      *RedisConf                 `yaml:"redis"`
	WebConsole *WebConsoleConf            `yaml:"webconsole"`
	Web        *WebConf                   `yaml:"web"`
}

// ReadFrom : read from file
func (c *Configurations) Init() error {
	c.Base = &BaseConf{}
	c.Base.Init()

	// Auth Config
	c.Auth = &AuthConf{}
	c.Auth.Init()

	// logging
	c.Logging = &LogConf{}
	c.Logging.Init()

	// BCS Config
	c.BCS = &BCSConf{}
	c.BCS.Init()

	c.BCSEnvConf = []*BCSConf{}
	c.BCSEnvMap = map[BCSClusterEnv]*BCSConf{}

	c.Redis = &RedisConf{}
	c.Redis.Init()

	c.WebConsole = &WebConsoleConf{}
	c.WebConsole.Init()

	c.Web = &WebConf{}
	c.Web.Init()

	return nil
}

// G : Global Configurations
var G = &Configurations{}

// 初始化
func init() {
	G.Init()
}

// ReadFrom : read from file
func (c *Configurations) ReadFrom(content []byte) error {
	if len(content) == 0 {
		return errors.New("conf content is empty, will use default values")
	}

	err := yaml.Unmarshal(content, &G)
	if err != nil {
		return err
	}
	c.Logging.InitBlog()
	c.Base.InitManagers()

	// 把列表类型转换为map，方便检索
	for _, conf := range c.BCSEnvConf {
		c.BCSEnvMap[conf.ClusterEnv] = conf
	}

	if err := c.WebConsole.InitMatchPattern(); err != nil {
		return err
	}

	if err := c.BCS.InitJWTPubKey(); err != nil {
		return err
	}

	return nil
}
