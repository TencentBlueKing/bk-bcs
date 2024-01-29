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
	"os"

	"gopkg.in/yaml.v3"
)

// Configuration 配置
type Configuration struct {
	Base         *BaseConf                    `yaml:"base_conf"`
	BCS          *BCSConf                     `yaml:"bcs_conf"`
	IAM          *IAMConf                     `yaml:"iam_conf"`
	BKNotice     *BKNoticeConf                `yaml:"bk_notice"`
	Web          *WebConf                     `yaml:"web"`
	Tracing      *TracingConf                 `yaml:"tracing"`
	FrontendConf *FrontendConf                `yaml:"frontend_conf"`
	FeatureFlags map[string]FeatureFlagOption `yaml:"feature_flags"`
	Etcd         *EtcdConf                    `yaml:"etcd"`
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
	_ = c.BCS.Init()

	c.IAM = &IAMConf{}
	c.FrontendConf = defaultFrontendConf()

	// 链路追踪初始化
	c.Tracing = &TracingConf{}
	c.Tracing.Init()

	// etcdc初始化
	c.Etcd = &EtcdConf{}
	c.Etcd.Init()
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

// IsLocalDevMode 是否本地开发模式
func (c *Configuration) IsLocalDevMode() bool {
	return c.Base.RunEnv == LocalEnv
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
	if c.Base.BKUsername == "" {
		c.Base.BKUsername = BK_USERNAME
	}
	if c.Base.Domain == "" {
		c.Base.Domain = BK_DOMAIN
	}

	// bcs env
	if c.BCS.Token == "" {
		c.BCS.Token = BCS_APIGW_TOKEN
	}
	if c.BCS.JWTPubKey == "" {
		c.BCS.JWTPubKey = BCS_APIGW_PUBLIC_KEY
	}
	if c.BCS.NamespacePrefix == "" {
		c.BCS.NamespacePrefix = BCS_NAMESPACE_PREFIX
	}

	// iam env
	if c.IAM.GatewayServer == "" {
		c.IAM.GatewayServer = BKIAM_GATEWAY_SERVER
	}

	// etcd env
	if c.Etcd.Endpoints == "" {
		c.Etcd.Endpoints = BCS_ETCD_HOST
	}

	if err := c.init(); err != nil {
		return err
	}

	if err := c.Base.InitBaseConf(); err != nil {
		return err
	}
	return nil
}

// BCSDebugAPIHost 事件未分离, 在前端分流
func (c *Configuration) BCSDebugAPIHost() string {
	return c.BCS.Host
}

// ReadFromFile : read from config file
func (c *Configuration) ReadFromFile(cfgFile string) error {
	// 不支持inline, 需要使用 yaml 库
	content, err := os.ReadFile(cfgFile)
	if err != nil {
		return err
	}
	return c.ReadFrom(content)
}
