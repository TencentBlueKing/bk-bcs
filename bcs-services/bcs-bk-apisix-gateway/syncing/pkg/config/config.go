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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// SyncConfig is a configuration of gateway syncing
type SyncConfig struct {
	EtcdConf    EtcdConf        `json:"etcd_conf"`
	GatewayConf GatewayConf     `json:"gateway_conf"`
	Logging     *conf.LogConfig `yaml:"logging"`
}

// EtcdConf is configuration of Etcd
type EtcdConf struct {
	// 证书内容，直接填入证书的内容（优先级低）
	EtcdCaCert   string `json:"etcd_ca_cert"`
	EtcdCertCert string `json:"etcd_cert_cert"`
	EtcdCertKey  string `json:"etcd_cert_key"`
	// 证书文件路径，用于挂载 Secret（优先级高，如果设置了路径则忽略直接内容）
	EtcdCaCertPath   string `json:"etcd_ca_cert_path"`
	EtcdCertCertPath string `json:"etcd_cert_cert_path"`
	EtcdCertKeyPath  string `json:"etcd_cert_key_path"`
	// 其他 Etcd 配置
	EtcdEndpoints  []string `json:"etcd_endpoints"`
	EtcdPassword   string   `json:"etcd_password"`
	EtcdPrefix     string   `json:"etcd_prefix"`
	EtcdSchemaType string   `json:"etcd_schema_type"`
	EtcdUsername   string   `json:"etcd_username"`
}

// GetEtcdCaCert 获取 CA 证书内容，优先从文件读取，否则返回配置的内容
func (e *EtcdConf) GetEtcdCaCert() (string, error) {
	if e.EtcdCaCertPath != "" {
		content, err := os.ReadFile(e.EtcdCaCertPath)
		if err != nil {
			blog.Errorf("failed to read etcd ca cert from path %s: %v", e.EtcdCaCertPath, err)
			return "", err
		}
		return string(content), nil
	}
	return e.EtcdCaCert, nil
}

// GetEtcdCertCert 获取客户端证书内容，优先从文件读取，否则返回配置的内容
func (e *EtcdConf) GetEtcdCertCert() (string, error) {
	if e.EtcdCertCertPath != "" {
		content, err := os.ReadFile(e.EtcdCertCertPath)
		if err != nil {
			blog.Errorf("failed to read etcd cert from path %s: %v", e.EtcdCertCertPath, err)
			return "", err
		}
		return string(content), nil
	}
	return e.EtcdCertCert, nil
}

// GetEtcdCertKey 获取客户端私钥内容，优先从文件读取，否则返回配置的内容
func (e *EtcdConf) GetEtcdCertKey() (string, error) {
	if e.EtcdCertKeyPath != "" {
		content, err := os.ReadFile(e.EtcdCertKeyPath)
		if err != nil {
			blog.Errorf("failed to read etcd key from path %s: %v", e.EtcdCertKeyPath, err)
			return "", err
		}
		return string(content), nil
	}
	return e.EtcdCertKey, nil
}

// GatewayConf is configuration of Apisix Gateway
type GatewayConf struct {
	GatewayHost   string   `json:"gateway_host"`
	XBkApiToken   string   `json:"x_bk_api_token"`
	ApisixType    string   `json:"apisix_type"`
	ApisixVersion string   `json:"apisix_version"`
	Description   string   `json:"description"`
	Maintainers   []string `json:"maintainers"`
	Mode          int      `json:"mode"`
	Name          string   `json:"name"`
	ReadOnly      bool     `json:"read_only"`
	ResourcesPath string   `json:"resources_path"`
}

func (s *SyncConfig) defaultSyncConfig() {
	if s.Logging == nil {
		s.Logging = &conf.LogConfig{
			ToStdErr: true,
		}
	}
}
