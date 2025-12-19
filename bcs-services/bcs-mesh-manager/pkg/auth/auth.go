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

// Package auth xxx
package auth

import (
	"context"
	"encoding/json"
	"slices"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	jwtGo "github.com/golang-jwt/jwt/v4"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

// JWTClientConfig jwt client config
type JWTClientConfig struct {
	Enable         bool
	PublicKey      string
	PublicKeyFile  string
	PrivateKey     string
	PrivateKeyFile string
}

var (
	jwtClient *jwt.JWTClient
)

// NewJWTClient new a jwt client
func NewJWTClient(c JWTClientConfig) (*jwt.JWTClient, error) {
	jwtOpt, err := getJWTOpt(c)
	if err != nil {
		return nil, common.UnauthError.GetErr()
	}
	jwtClient, err = jwt.NewJWTClient(*jwtOpt)
	if err != nil {
		return nil, err
	}

	return jwtClient, nil
}

// GetJWTClient get jwt client
func GetJWTClient() *jwt.JWTClient {
	return jwtClient
}

func getJWTOpt(c JWTClientConfig) (*jwt.JWTOptions, error) {
	jwtOpt := &jwt.JWTOptions{
		VerifyKeyFile: c.PublicKeyFile,
		SignKeyFile:   c.PrivateKeyFile,
	}
	publicKey := c.PublicKey
	privateKey := c.PrivateKey

	if publicKey != "" {
		key, err := jwtGo.ParseRSAPublicKeyFromPEM([]byte(publicKey))
		if err != nil {
			return nil, err
		}
		jwtOpt.VerifyKey = key
	}
	if privateKey != "" {
		key, err := jwtGo.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
		if err != nil {
			return nil, err
		}
		jwtOpt.SignKey = key
	}
	return jwtOpt, nil
}

// SkipHandler skip handler
func SkipHandler(ctx context.Context, req server.Request) bool {
	// disable auth
	if !options.GlobalOptions.Auth.Enable {
		return true
	}

	// 检查配置中的方法
	if len(options.GlobalOptions.Auth.NoAuthMethod) > 0 {
		methods := strings.Split(options.GlobalOptions.Auth.NoAuthMethod, ",")
		for _, method := range methods {
			method = strings.TrimSpace(method)
			if method == req.Method() {
				return true
			}
		}
	}

	return false
}

// SkipClient skip client
func SkipClient(ctx context.Context, req server.Request, client string) bool {
	// 解析客户端权限配置
	clientPermissions, err := parseClientPermissions(options.GlobalOptions.Auth.ClientPermissions)
	if err != nil {
		blog.Errorf("failed to parse client permissions: %v", err)
		return false
	}

	// 检查客户端是否存在
	permissions, exists := clientPermissions[client]
	if !exists {
		blog.Errorf("client %s not found in permissions config", client)
		return false
	}

	// 检查是否有所有权限
	if slices.Contains(permissions, "*") {
		return true
	}

	// 检查是否有特定方法的权限
	method := req.Method()
	if slices.Contains(permissions, method) {
		return true
	}

	blog.Errorf("client %s does not have permission for method %s", client, method)
	return false
}

// parseClientPermissions 解析客户端权限配置
func parseClientPermissions(config string) (map[string][]string, error) {
	if config == "" {
		return make(map[string][]string), nil
	}

	var clientPermissions map[string][]string
	err := json.Unmarshal([]byte(config), &clientPermissions)
	if err != nil {
		return nil, err
	}

	return clientPermissions, nil
}

// GetUserFromCtx 通过 ctx 获取当前用户
func GetUserFromCtx(ctx context.Context) string {
	authUser, err := middleauth.GetUserFromContext(ctx)
	if err != nil {
		blog.Errorf("get user from context failed, err: %s", err)
		return ""
	}
	return authUser.GetUsername()
}
