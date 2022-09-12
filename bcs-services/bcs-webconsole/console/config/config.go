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
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Configurations : manage all configurations
type Configurations struct {
	mtx         sync.Mutex
	Base        *BaseConf                  `yaml:"base_conf"`
	Auth        *AuthConf                  `yaml:"auth_conf"`
	BkLogin     *BKLoginConf               `yaml:"bklogin_conf"`
	Logging     *LogConf                   `yaml:"logging"`
	BKAPIGW     *BKAPIGWConf               `yaml:"bkapigw_conf"`
	BCS         *BCSConf                   `yaml:"bcs_conf"`
	BCSCC       *BCSCCConf                 `yaml:"bcs_cc_conf"`
	BCSEnvConf  []*BCSConf                 `yaml:"bcs_env_conf"`
	Credentials map[string][]*Credential   `yaml:"-"`
	BCSEnvMap   map[BCSClusterEnv]*BCSConf `yaml:"-"`
	Redis       *RedisConf                 `yaml:"redis"`
	WebConsole  *WebConsoleConf            `yaml:"webconsole"`
	Web         *WebConf                   `yaml:"web"`
}

// newConfigurations 新增配置
func newConfigurations() (*Configurations, error) {
	c := &Configurations{}

	c.Base = &BaseConf{}
	c.Base.Init()

	// Auth Config
	c.Auth = &AuthConf{}
	c.Auth.Init()

	// BkLogin Config
	c.BkLogin = &BKLoginConf{}
	c.BkLogin.Init()

	c.BKAPIGW = &BKAPIGWConf{}
	c.BKAPIGW.Init()

	// logging
	c.Logging = &LogConf{}
	c.Logging.Init()

	// BCS Config
	c.BCS = &BCSConf{}
	c.BCS.Init()

	// BCS-CC Config
	c.BCSCC = &BCSCCConf{}
	c.BCSCC.Init()

	c.BCSEnvConf = []*BCSConf{}
	c.BCSEnvMap = map[BCSClusterEnv]*BCSConf{}

	c.Redis = &RedisConf{}
	c.Redis.Init()

	c.WebConsole = &WebConsoleConf{}
	c.WebConsole.Init()

	c.Credentials = map[string][]*Credential{}

	c.Web = defaultWebConf()

	return c, nil
}

// init 初始化
func (c *Configurations) init() error {
	if err := c.Web.init(); err != nil {
		return err
	}

	return nil
}

// IsDevMode 是否本地开发模式
func (c *Configurations) IsDevMode() bool {
	return c.Base.RunEnv == DevEnv
}

// G : Global Configurations
var G *Configurations

// init 初始化
func init() {
	g, err := newConfigurations()
	if err != nil {
		panic(err)
	}
	if err := g.init(); err != nil {
		panic(err)
	}

	G = g
}

// ReadCred xxx
func (c *Configurations) ReadCred(name string, content []byte) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	cred := []*Credential{}
	err := yaml.Unmarshal(content, &cred)
	if err != nil {
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

// ValidateCred 校验凭证是否合法
func (c *Configurations) ValidateCred(credType CredentialType, credName string, scopeType ScopeType,
	scopeValue string) bool {
	for _, creds := range c.Credentials {
		for _, cred := range creds {
			if cred.Matches(credType, credName, scopeType, scopeValue) {
				return true
			}
		}
	}
	return false
}

// IsManager 校验固定的 manager 和 集群维度动态凭证
func (c *Configurations) IsManager(username, clusterId string) bool {
	if _, ok := c.Base.ManagerMap[username]; ok {
		return true
	}

	if c.ValidateCred(CredentialManager, username, ScopeClusterId, clusterId) {
		return true
	}

	return false
}

// ReadFrom : read from file
func (c *Configurations) ReadFrom(content []byte) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if len(content) == 0 {
		return errors.New("conf content is empty, will use default values")
	}

	err := yaml.Unmarshal(content, &G)
	if err != nil {
		return err
	}

	if err := c.init(); err != nil {
		return err
	}

	c.Logging.InitBlog()
	c.Base.InitManagers()

	// 把列表类型转换为map，方便检索
	for _, conf := range c.BCSEnvConf {
		c.BCSEnvMap[conf.ClusterEnv] = conf
	}

	if err := c.WebConsole.InitTagPatterns(); err != nil {
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
