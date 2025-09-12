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

const (
	// ProdCluster xxx
	ProdCluster BCSClusterEnv = "prod" // 正式环境
	// DebugCLuster xxx
	DebugCLuster BCSClusterEnv = "debug" // debug 环境
	// UatCluster xxx
	UatCluster BCSClusterEnv = "uat" // uat 环境
)

// BCSConf :
type BCSConf struct {
	Target       string         `yaml:"target"`
	Host         string         `yaml:"host"`
	Token        string         `yaml:"token"`
	JWTPubKey    string         `yaml:"jwt_public_key"`
	JWTPubKeyObj *rsa.PublicKey `yaml:"-"`
}

// Init :
func (c *BCSConf) Init() error {
	// only for development
	c.Host = ""
	c.Token = ""
	c.JWTPubKey = ""
	c.JWTPubKeyObj = nil
	return nil
}

// InitJWTPubKey :
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
