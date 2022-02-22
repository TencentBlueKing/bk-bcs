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
	"go-micro.dev/v4/logger"
	"gopkg.in/yaml.v2"
)

// Configurations : manage all configurations
type Configurations struct {
	Base       *BaseConf       `yaml:"base_conf"`
	BCS        *BCSConf        `yaml:"bcs_conf"`
	Redis      *RedisConf      `yaml:"redis"`
	WebConsole *WebConsoleConf `yaml:"webconsole"`
	Web        *WebConf        `yaml:"web"`
}

// ReadFrom : read from file
func (c *Configurations) Init() error {
	c.Base = &BaseConf{}
	c.Base.Init()

	// BCS Config
	c.BCS = &BCSConf{}
	c.BCS.Init()

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
		logger.Info("conf content is empty, will use default values")
		return nil
	}

	err := yaml.Unmarshal(content, &G)
	if err != nil {
		return err
	}
	return nil
}
