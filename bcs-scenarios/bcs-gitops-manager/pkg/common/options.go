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

package common

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
)

// ServerConfig option for server side
type ServerConfig struct {
	Debug       bool        `json:"debug"`
	Address     string      `json:"address,omitempty"`
	IPv6Address string      `json:"ipv6address,omitempty"`
	Port        uint        `json:"port,omitempty"`
	HTTPPort    uint        `json:"httpport,omitempty"`
	MetricPort  uint        `json:"metricport,omitempty"`
	ServerCert  string      `json:"servercert,omitempty"`
	ServerKey   string      `json:"serverkey,omitempty"`
	ServerCa    string      `json:"serverca,omitempty"`
	ServerTLS   *tls.Config `json:"-"`
}

// Complete unset item
func (s *ServerConfig) Complete() error {
	// loading server tls configuration
	svrConfig, err := ssl.ServerTslConfVerityClient(s.ServerCa, s.ServerCert,
		s.ServerKey, static.ServerCertPwd)
	if err != nil {
		return fmt.Errorf("loading server side tls config failed, %s", err.Error())
	}
	s.ServerTLS = svrConfig
	s.ServerTLS.ClientAuth = tls.NoClientCert
	s.ServerTLS.InsecureSkipVerify = true
	return nil
}

// ClientConfig option for as client side
type ClientConfig struct {
	ClientCert string      `json:"clientcert,omitempty"`
	ClientKey  string      `json:"clientkey,omitempty"`
	ClientCa   string      `json:"clientca,omitempty"`
	ClientTLS  *tls.Config `json:"-,omitempty"`
}

// Complete unset item
func (c *ClientConfig) Complete() error {
	// loading client tls configuration
	cliConfig, err := ssl.ClientTslConfVerity(c.ClientCa, c.ClientCert,
		c.ClientKey, static.ClientCertPwd)
	if err != nil {
		return fmt.Errorf("loading client side tls configuration failed, %s", err.Error())
	}
	// NOTE: clean setting when releasing office version
	c.ClientTLS = cliConfig
	c.ClientTLS.ClientAuth = tls.NoClientCert
	c.ClientTLS.InsecureSkipVerify = true
	return nil
}

// Registry definition for gitops
type Registry struct {
	// Endpooints ip address for registry, split by comma
	Endpoints string      `json:"endpoints,omitempty"`
	CA        string      `json:"ca,omitempty"`
	Key       string      `json:"key,omitempty"`
	Cert      string      `json:"cert,omitempty"`
	TLSConfig *tls.Config `json:"-,omitempty"`
}

// Complete unset item
func (r *Registry) Complete() error {
	// loading registry tls configuration
	etcdConfig, err := ssl.ClientTslConfVerity(r.CA, r.Cert, r.Key, "")
	if err != nil {
		return fmt.Errorf("loading etcd registry tls configuration failed, %s", err.Error())
	}
	r.TLSConfig = etcdConfig
	return nil
}

// LoadConfigFile loading json config file
func LoadConfigFile(fileName string, opt interface{}) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, opt)
}

// TraceConfig defines the config of trace from config file
type TraceConfig struct {
	Endpoint string `json:"endpoint,omitempty"`
	Token    string `json:"token,omitempty"`
}

// AuthConfig for AuthCenter
type AuthConfig struct {
	VerifyKeyFile string `json:"verifykeyfile,omitempty"`
	SignKeyFile   string `json:"signkeyfile,omitempty"`
	External      bool   `json:"external,omitempty"`
	SystemID      string `json:"systemid,omitempty"`
	AppCode       string `json:"appcode,omitempty"`
	AppSecret     string `json:"appsecret,omitempty"`
	Gateway       string `json:"gateway,omitempty"`
	IAMHost       string `json:"iamhost,omitempty"`
	BKIAM         string `json:"bkiam,omitempty"`
}

// AuditConfig defines the config of audit
type AuditConfig struct {
	BCSGateway string `json:"bcsGateway"`
	Token      string `json:"token"`
}

// Complete unset item
func (config *AuthConfig) Complete() error {
	return nil
}

// Validate check
func (config *AuthConfig) Validate() error {
	if len(config.SignKeyFile) == 0 || len(config.VerifyKeyFile) == 0 {
		return fmt.Errorf("lost token validation")
	}
	if len(config.SystemID) == 0 || len(config.AppCode) == 0 || len(config.AppSecret) == 0 {
		return fmt.Errorf("lost auth system ID")
	}
	return nil
}
