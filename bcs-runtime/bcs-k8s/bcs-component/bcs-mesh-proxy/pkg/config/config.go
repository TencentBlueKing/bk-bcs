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

// Package config provides configuration for the mesh proxy.
package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

// Config 代理服务器配置
type Config struct {
	// 本集群配置（程序部署的集群）
	TargetCluster TargetClusterConfig `yaml:"targetCluster"`

	// 代理配置
	Proxy ProxyConfig `yaml:"proxy"`
}

// TargetClusterConfig 目标集群配置
type TargetClusterConfig struct {
	// 本集群的API服务器地址
	APIServer string `yaml:"apiServer"`

	// 本集群的CA证书路径
	CACertPath string `yaml:"caCertPath"`

	// 本集群的客户端证书路径
	ClientCertPath string `yaml:"clientCertPath"`

	// 本集群的客户端密钥路径
	ClientKeyPath string `yaml:"clientKeyPath"`

	// 本集群的Bearer Token
	BearerToken string `yaml:"bearerToken"`

	// 连接超时时间
	Timeout time.Duration `yaml:"timeout"`

	// 是否使用in-cluster配置
	UseInClusterConfig bool `yaml:"useInClusterConfig"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	// 监听的端口
	Port int `yaml:"port"`

	// 是否跳过TLS验证
	InsecureSkipTLSVerify bool `yaml:"insecureSkipTLSVerify"`

	// 请求超时时间
	RequestTimeout time.Duration `yaml:"requestTimeout"`

	// 允许的API组和版本
	AllowedAPIGroups []string `yaml:"allowedAPIGroups"`

	// TLS配置
	TLS TLSConfig `yaml:"tls"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	// 是否启用TLS
	Enabled bool `yaml:"enabled"`

	// 证书文件路径
	CertFile string `yaml:"certFile"`

	// 密钥文件路径
	KeyFile string `yaml:"keyFile"`

	// 客户端认证模式
	// "NoClientCert", "RequestClientCert", "RequireAnyClientCert",
	// "VerifyClientCertIfGiven", "RequireAndVerifyClientCert"
	ClientAuth string `yaml:"clientAuth"`

	// CA证书文件路径
	CAFile string `yaml:"caFile"`
}

// Load 从文件加载配置
func Load(configFile string) (*Config, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 设置默认值
	if config.TargetCluster.Timeout == 0 {
		config.TargetCluster.Timeout = 30 * time.Second
	}

	if config.Proxy.RequestTimeout == 0 {
		config.Proxy.RequestTimeout = 60 * time.Second
	}

	if config.Proxy.Port == 0 {
		config.Proxy.Port = 61011
	}

	return &config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 如果使用in-cluster配置，则不需要验证API服务器地址
	if !c.TargetCluster.UseInClusterConfig {
		if c.TargetCluster.APIServer == "" {
			return fmt.Errorf("本集群API服务器地址不能为空")
		}
	}

	return nil
}
