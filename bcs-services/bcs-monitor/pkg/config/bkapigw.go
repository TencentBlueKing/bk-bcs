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
	"crypto/rsa"

	"github.com/dgrijalva/jwt-go"
)

// BKAPIGWConf :
type BKAPIGWConf struct {
	Host         string         `yaml:"host"`
	JWTPubKey    string         `yaml:"jwt_public_key"`
	JWTPubKeyObj *rsa.PublicKey `yaml:"-"`
}

// Init
func (c *BKAPIGWConf) Init() error {
	// only for development
	c.Host = ""
	c.JWTPubKey = ""
	return nil
}

// InitJWTPubKey
func (c *BKAPIGWConf) InitJWTPubKey() error {
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
