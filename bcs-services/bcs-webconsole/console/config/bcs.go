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

package config

import (
	"crypto/rsa"

	"github.com/golang-jwt/jwt/v4"
)

// BCSClusterEnv xxx
type BCSClusterEnv string

// BCSConf BCS配置
type BCSConf struct {
	InnerHost             string         `yaml:"inner_host"`
	Host                  string         `yaml:"host"`
	Token                 string         `yaml:"token"`
	InsecureSkipVerify    bool           `yaml:"insecure_skip_verify"`
	JWTPubKey             string         `yaml:"jwt_public_key"`
	EnableMultiTenantMode bool           `yaml:"enable_multi_tenant_mode"`
	JWTPubKeyObj          *rsa.PublicKey `yaml:"-"`
}

// Init xxx
func (c *BCSConf) Init() {
	// only for development
	c.InnerHost = ""
	c.Host = ""
	c.Token = ""
	c.JWTPubKey = ""
	c.JWTPubKeyObj = nil
	c.InsecureSkipVerify = false
	c.EnableMultiTenantMode = false
}

// InitJWTPubKey xxx
func (c *BCSConf) InitJWTPubKey() error {
	if c.JWTPubKey == "" {
		return nil
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(c.JWTPubKey))
	if err != nil {
		return err
	}

	c.JWTPubKeyObj = pubKey
	return nil
}

func (c *BCSConf) initInnerHost() {
	if c.InnerHost == "" {
		c.InnerHost = c.Host
	}
}
