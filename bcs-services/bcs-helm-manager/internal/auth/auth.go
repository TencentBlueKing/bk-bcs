/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	jwtGo "github.com/dgrijalva/jwt-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/util/stringx"
)

type JWTClientConfig struct {
	Enable         bool
	PublicKey      string
	PublicKeyFile  string
	PrivateKey     string
	PrivateKeyFile string
	ExemptClients  string
}

var (
	jwtClient *jwt.JWTClient
	JWTConfig *JWTClientConfig
)

func NewJWTClient(c JWTClientConfig) (*jwt.JWTClient, error) {
	jwtOpt, err := getJWTOpt(c)
	if err != nil {
		return nil, common.ErrHelmManagerAuthFailed.GenError()
	}
	JWTConfig = &c
	jwtClient, err = jwt.NewJWTClient(*jwtOpt)
	if err != nil {
		return nil, err
	}

	return jwtClient, nil
}

// GetUserFromCtx 通过 ctx 获取当前用户
func GetUserFromCtx(ctx context.Context) string {
	username, ok := ctx.Value(ctxkey.UsernameKey).(string)
	if !ok {
		blog.Warnf("获取用户信息异常, 非字符串类型!")
		return ""
	}
	return username
}

type AuthUser struct {
	Username string
	UserType string
	ClientID string
}

// ParseUserFromJWT 通过 jwt token 解析当前用户
func ParseUserFromJWT(jwtToken string) (*AuthUser, error) {
	claims, err := parseClaims(jwtToken)
	if err != nil {
		return nil, err
	}

	return &AuthUser{
		Username: claims.UserName,
		UserType: claims.SubType,
		ClientID: claims.ClientID,
	}, nil
}

func parseClaims(jwtToken string) (*jwt.UserClaimsInfo, error) {
	// 解析token
	claims, err := jwtClient.JWTDecode(jwtToken)
	if err != nil {
		return nil, err
	}
	return claims, nil
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

// NoAuthEndpoints 不需要用户身份认证的方法
var NoAuthEndpoints = []string{
	"Common.Available",
}

// 检查当前请求是否允许免除用户认证
func CanExemptAuth(ep string) bool {
	// 禁用身份认证
	if !JWTConfig.Enable {
		return true
	}
	// 特殊指定的Handler，不需要认证的方法
	return stringx.StringInSlice(ep, NoAuthEndpoints)
}
