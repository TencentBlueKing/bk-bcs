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

// Package config xxx
package config

import (
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Configuration 配置
type Configuration struct {
	Viper        *viper.Viper  `yaml:"-"`
	Base         *BaseConf     `yaml:"base_conf"`
	BCS          *BCSConf      `yaml:"bcs_conf"`
	Web          *WebConf      `yaml:"web"`
	FrontendConf *FrontendConf `yaml:"frontend_conf"`
	Tracing      *TracingConf  `yaml:"tracing"`
}

// init 初始化
func (c *Configuration) init() error {
	if err := c.Web.init(); err != nil {
		return err
	}

	if err := c.BCS.InitJWTPubKey(); err != nil {
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

	c.Web = defaultWebConf()

	// BCS Config
	c.BCS = &BCSConf{}
	c.BCS.Init()

	c.FrontendConf = defaultFrontendConf()

	c.Tracing = &TracingConf{}
	c.Tracing.Init()
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
func (c *Configuration) IsLocalDevMode() bool {
	return c.Base.RunEnv == LocalEnv
}

// ReadFrom : read from file
func (c *Configuration) ReadFrom(content []byte) error {
	if err := yaml.Unmarshal(content, c); err != nil {
		return err
	}
	if err := c.init(); err != nil {
		return err
	}

	if err := c.Base.InitBaseConf(); err != nil {
		return err
	}
	return nil
}

// DebugAPIHost 事件未分离, 在前端分流
func (c *Configuration) BCSDebugAPIHost() string {
	return c.BCS.Host
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
