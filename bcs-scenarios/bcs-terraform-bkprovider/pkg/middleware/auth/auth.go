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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/middleware"
)

// AuthTokenKey xx
const AuthTokenKey string = "X-JWT-Token"

// UserInfo for token validate
type UserInfo struct {
	Token string `json:"token"`
	*jwt.UserClaimsInfo
}

// GetUser string
func (user *UserInfo) GetUser() string {
	if len(user.UserName) != 0 {
		return user.UserName
	}
	if len(user.ClientID) != 0 {
		return user.ClientID
	}
	return ""
}

// JWTAuth jwt auth
type JWTAuth struct {
	client *jwt.JWTClient
}

// NewJWTAuth init jwt client
func NewJWTAuth(publicKeyFile, privateKeyFile string) (*JWTAuth, error) {
	cli, err := jwt.NewJWTClient(jwt.JWTOptions{
		VerifyKeyFile: publicKeyFile,
		SignKeyFile:   privateKeyFile,
	})
	if err != nil {
		return nil, err
	}
	return &JWTAuth{
		client: cli,
	}, nil
}

// GetJWTClient get jwt client
func (j *JWTAuth) GetJWTClient() *jwt.JWTClient {
	return j.client
}

// SkipHandler skip handler
func (j *JWTAuth) SkipHandler(ctx context.Context, req server.Request) bool {
	for _, v := range NoAuthMethod {
		if v == req.Method() {
			return true
		}
	}
	return false
}

// GetJWTInfoWithAuthorization 根据 token 获取用户信息
func (j *JWTAuth) GetJWTInfoWithAuthorization(authorization string) (*UserInfo, error) {
	if len(authorization) == 0 {
		return nil, fmt.Errorf("lost 'Authorization' header")
	}
	if !strings.HasPrefix(authorization, "Bearer ") {
		return nil, fmt.Errorf("hader 'Authorization' malform")
	}
	token := strings.TrimPrefix(authorization, "Bearer ")
	claim, err := j.client.JWTDecode(token)
	if err != nil {
		return nil, err
	}
	u := &UserInfo{
		Token:          token,
		UserClaimsInfo: claim,
	}
	if u.GetUser() == "" {
		return nil, fmt.Errorf("lost user information")
	}
	return u, nil
}

// AuthorizationFunc will trans the user auth
func (j *JWTAuth) AuthorizationFunc(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		md, ok := metadata.FromContext(ctx)
		if !ok {
			return errors.New("failed to get micro's metadata")
		}
		authUser := middleware.AuthUser{}

		// parse client token from header
		clientName, ok := md.Get(middleware.InnerClientHeaderKey)
		if ok {
			authUser.InnerClient = clientName
		}

		// parse username from header
		username, ok := md.Get(middleware.CustomUsernameHeaderKey)
		if ok {
			authUser.Username = username
		}

		// parse jwt token from header
		jwtToken, ok := md.Get(middleware.AuthorizationHeaderKey)
		if ok {
			u, err := j.GetJWTInfoWithAuthorization(jwtToken)
			if err != nil {
				return err
			}
			if u.SubType == jwt.User.String() {
				authUser.Username = u.UserName
			}
			if u.SubType == jwt.Client.String() {
				authUser.ClientName = u.ClientID
			}
			if len(u.BKAppCode) != 0 {
				authUser.ClientName = u.BKAppCode
			}
		}

		// set auth user to context
		data, _ := json.Marshal(authUser)
		md.Set(string(middleware.AuthUserKey), string(data))
		md.Set(AuthTokenKey, jwtToken)
		ctx = metadata.MergeContext(ctx, md, false)

		return fn(ctx, req, rsp)
	}
}
