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
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"go-micro.dev/v4/logger"
	"gopkg.in/yaml.v2"
)

// CertConfig is configuration of Cert
type CertConfig struct {
	CAFile     string
	CertFile   string
	KeyFile    string
	CertPasswd string
	IsSSL      bool
}

// ConsoleConfig Config is a configuration
type ConsoleConfig struct {
	Address         string
	Port            int
	ServCert        *CertConfig
	WebConsoleImage string
}

// Configurations : manage all configurations
type Configurations struct {
	BCS   *BCSConf   `yaml:"bcs_conf"`
	Redis *RedisConf `yaml:"redis"`
}

// ReadFrom : read from file
func (c *Configurations) Init() error {
	// BCS Config
	c.BCS = &BCSConf{}
	c.BCS.Init()

	c.Redis = &RedisConf{}
	c.Redis.Init()

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

// NewConsoleConfig create a config object
func NewConsoleConfig() ConsoleConfig {
	return ConsoleConfig{
		ServCert: &CertConfig{
			CertPasswd: static.ServerCertPwd,
			IsSSL:      false,
		},
	}
}
