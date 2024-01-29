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
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Configurations : manage all configurations
type Configurations struct {
	mtx         sync.Mutex
	Base        *BaseConf                `yaml:"base_conf"`
	Auth        *AuthConf                `yaml:"auth_conf"`
	Logging     *LogConf                 `yaml:"logging"`
	BCS         *BCSConf                 `yaml:"bcs_conf"`
	Credentials map[string][]*Credential `yaml:"-"`
	Redis       *RedisConf               `yaml:"redis"`
	WebConsole  *WebConsoleConf          `yaml:"webconsole"`
	Web         *WebConf                 `yaml:"web"`
	Etcd        *EtcdConf                `yaml:"etcd"`
	Tracing     *TracingConf             `yaml:"tracing"`
	Audit       *AuditConf               `yaml:"audit"`
	Repository  *RepositoryConf          `yaml:"repository"`
}

// newConfigurations 新增配置
func newConfigurations() (*Configurations, error) {
	c := &Configurations{}

	c.Base = &BaseConf{}
	err := c.Base.Init()
	if err != nil {
		return c, err
	}

	// Auth Config
	c.Auth = &AuthConf{}
	c.Auth.Init()

	// logging
	c.Logging = &LogConf{}
	c.Logging.Init()

	// BCS Config
	c.BCS = &BCSConf{}
	c.BCS.Init()

	c.Redis = &RedisConf{}
	c.Redis.Init()

	c.WebConsole = &WebConsoleConf{}
	err = c.WebConsole.Init()
	if err != nil {
		return c, err
	}

	c.Etcd = &EtcdConf{}
	c.Etcd.Init()

	// Tracing Config
	c.Tracing = &TracingConf{}
	c.Tracing.Init()

	c.Audit = &AuditConf{}
	// repository conf
	c.Repository = &RepositoryConf{}

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
func (c *Configurations) ValidateCred(credType CredentialType, credName string, scopeValues map[ScopeType]string) bool {
	for _, creds := range c.Credentials {
		for _, cred := range creds {
			if cred.Matches(credType, credName, scopeValues) {
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

	scopeValues := map[ScopeType]string{
		ScopeClusterId: clusterId,
	}

	return c.ValidateCred(CredentialManager, username, scopeValues)
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

	// 添加环境变量
	if c.Base.AppCode == "" {
		c.Base.AppCode = BK_APP_CODE
	}
	if c.Base.AppSecret == "" {
		c.Base.AppSecret = BK_APP_SECRET
	}
	if c.Base.BKPaaSHost == "" {
		c.Base.BKPaaSHost = BK_PAAS_HOST
	}
	if c.Auth.Host == "" {
		c.Auth.Host = BK_IAM_HOST
	}
	if c.Auth.GatewayHost == "" {
		c.Auth.GatewayHost = BK_IAM_GATEWAY_HOST
	}
	// 为空以配置文件为准, 如果设置，以环境变量为准, false代表网关模式
	if BK_IAM_EXTERNAL != "" {
		if external, e := strconv.ParseBool(BK_IAM_EXTERNAL); e == nil {
			c.Auth.UseGateway = !external
		}
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

	if e := c.init(); e != nil {
		return err
	}

	if c.Audit.DataDir == "" {
		c.Audit.DataDir = c.Audit.defaultPath()
	}

	c.Audit.DataDir, err = filepath.Abs(c.Audit.DataDir)
	if err != nil {
		return err
	}
	// 右边的目录去除
	c.Audit.DataDir = strings.TrimRight(c.Audit.DataDir, "/")

	if err := c.Logging.InitBlog(); err != nil {
		return err
	}
	c.Base.InitManagers()
	c.BCS.initInnerHost()

	if err := c.WebConsole.InitTagPatterns(); err != nil {
		return err
	}

	if err := c.WebConsole.parseRes(); err != nil {
		return err
	}

	if err := c.BCS.InitJWTPubKey(); err != nil {
		return err
	}

	return nil
}
