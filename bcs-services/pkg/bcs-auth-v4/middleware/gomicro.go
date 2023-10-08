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

package middleware

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"
)

// GoMicroAuth is the authentication middleware for go-micro
type GoMicroAuth struct {
	skipHandler   func(ctx context.Context, req server.Request) bool
	exemptClient  func(ctx context.Context, req server.Request, client string) bool
	checkUserPerm func(ctx context.Context, req server.Request, username string) (bool, error)
	jwtClient     *jwt.JWTClient
}

// NewGoMicroAuth creates a new go-micro authentication middleware
func NewGoMicroAuth(jwtClient *jwt.JWTClient) *GoMicroAuth {
	return &GoMicroAuth{
		jwtClient: jwtClient,
	}
}

// EnableSkipHandler enable skip method, if skipHandler return true, skip authorization, like Health handler
func (g *GoMicroAuth) EnableSkipHandler(skipHandler func(ctx context.Context, req server.Request) bool) *GoMicroAuth {
	g.skipHandler = skipHandler
	return g
}

// EnableSkipClient enable skip client, if skip client return true, skip authorization
// nolint
func (g *GoMicroAuth) EnableSkipClient(exemptClient func(ctx context.Context, req server.Request, client string) bool) *GoMicroAuth {
	g.exemptClient = exemptClient
	return g
}

// SetCheckUserPerm set check user permission function
func (g *GoMicroAuth) SetCheckUserPerm(checkUserPerm func(ctx context.Context,
	req server.Request, username string) (bool, error)) *GoMicroAuth {
	g.checkUserPerm = checkUserPerm
	return g
}

// AuthenticationFunc is the authentication function for go-micro
func (g *GoMicroAuth) AuthenticationFunc() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			md, ok := metadata.FromContext(ctx)
			if !ok {
				return errors.New("failed to get micro's metadata")
			}
			authUser := AuthUser{}

			// parse client token from header
			clientName, ok := md.Get(InnerClientHeaderKey)
			if ok {
				authUser.InnerClient = clientName
			}

			// parse jwt token from header
			jwtToken, ok := md.Get(AuthorizationHeaderKey)
			if ok {
				u, err := g.parseJwtToken(jwtToken)
				if err != nil {
					return err
				}
				// !NOTO: bk-apigw would set SubType to "user" even if use client's app code and secret
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

			// If and only if client name from jwt token is not empty, we will check username in header
			if authUser.ClientName != "" {
				username, ok := md.Get(CustomUsernameHeaderKey)
				if ok && username != "" {
					authUser.Username = username
				}
			}

			// set auth user to context
			ctx = context.WithValue(ctx, AuthUserKey, authUser)
			return fn(ctx, req, rsp)
		}
	}
}

// AuthorizationFunc is the authorization function for go-micro
func (g *GoMicroAuth) AuthorizationFunc() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			if g.skipHandler != nil && g.skipHandler(ctx, req) {
				return fn(ctx, req, rsp)
			}

			authUser, err := GetUserFromContext(ctx)
			if err != nil {
				return err
			}

			if authUser.IsInner() {
				return fn(ctx, req, rsp)
			}

			if g.exemptClient != nil && g.exemptClient(ctx, req, authUser.ClientName) {
				return fn(ctx, req, rsp)
			}

			if len(authUser.Username) == 0 {
				return errors.New("username is empty")
			}

			if g.checkUserPerm == nil {
				return errors.New("check user permission function is not set")
			}

			if allow, err := g.checkUserPerm(ctx, req, authUser.Username); err != nil {
				return err
			} else if !allow {
				return errors.New("user not authorized")
			}

			return fn(ctx, req, rsp)
		}
	}
}

func (g *GoMicroAuth) parseJwtToken(jwtToken string) (*jwt.UserClaimsInfo, error) {
	if len(jwtToken) == 0 || !strings.HasPrefix(jwtToken, "Bearer ") {
		return nil, errors.New("authorization token error")
	}
	claims, err := g.jwtClient.JWTDecode(jwtToken[7:])
	if err != nil {
		return nil, err
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("authorization token expired")
	}
	return claims, nil
}
