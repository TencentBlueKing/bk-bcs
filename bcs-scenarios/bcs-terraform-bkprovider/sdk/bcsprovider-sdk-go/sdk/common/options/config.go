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

// Package options config
package options

import (
	"os"

	"github.com/pkg/errors"
)

const (
	// bcsGatewayAddrEnv bcs网关地址-env
	bcsGatewayAddrEnv = "BCS_GATEWAY_ADDR"
)

// Config bcsprovider-sdk-go 配置
type Config struct {
	// Username 用户名
	Username string `json:"username"`

	// Token bcs-bearer-token
	Token string `json:"token"`

	// BcsGatewayAddr bcs网关地址，如：https://xxx 或者 http://xxx
	BcsGatewayAddr string `json:"bcs_gateway_addr"`

	// InsecureSkipVerify 是否跳过验证
	InsecureSkipVerify bool `json:"insecure_skip_verify"`
}

// Check 参数检查
func (c *Config) Check() error {
	if len(c.BcsGatewayAddr) == 0 { // 如果为空，则从环境变量读取
		c.BcsGatewayAddr = os.Getenv(bcsGatewayAddrEnv)
	}
	if len(c.BcsGatewayAddr) == 0 { // 再次检查
		return errors.Errorf("bcs_gateway_addr cannot be empty.")
	}
	if len(c.Token) == 0 {
		return errors.Errorf("token cannot be empty.")
	}
	if len(c.Username) == 0 {
		return errors.Errorf("username cannot be empty.")
	}

	return nil
}
