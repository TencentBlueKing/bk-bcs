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

// Package auth xxx
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
	"github.com/parnurzeal/gorequest"
)

// Auth interface for accessToken
type Auth interface {
	// GetAccessToken get access token
	GetAccessToken() (string, error)
}

var (
	defaultTimeOut   = time.Second * 60
	errServerNotInit = errors.New("server not inited")
)

var authClient Auth

// SetAuthClient set auth client
func SetAuthClient(options Options) error {
	if !options.Enable {
		authClient = nil
		return nil
	}
	cli := NewAuthClient(options)

	authClient = cli
	return nil
}

// GetAuthClient get auth client
func GetAuthClient() Auth {
	return authClient
}

// NewAuthClient init auth client
func NewAuthClient(opt Options) *ClientAuth {
	cli := &ClientAuth{
		server:    opt.Server,
		appCode:   opt.AppCode,
		appSecret: opt.AppSecret,
	}
	return cli
}

// Options opts parameter
type Options struct {
	// Server auth address
	Server string
	// AppCode app code
	AppCode string
	// AppSecret app secret
	AppSecret string
	// Enable enable feature
	Enable bool
}

// ClientAuth auth client
type ClientAuth struct {
	server    string
	appCode   string
	appSecret string
}

// GetAccessToken get access token
func (auth *ClientAuth) GetAccessToken() (string, error) {
	if auth == nil {
		return "", errServerNotInit
	}

	path := "/oauth/token"
	if config.GetGlobalConfig().CommunityEdition {
		path = "/api/v1/auth/access-tokens"
	}

	var (
		url = auth.server + path
		req = &AccessRequest{
			AppCode:    auth.appCode,
			AppSecret:  auth.appSecret,
			IDProvider: clientProvider,
			GrantType:  clientCredentialGrant.String(),
			Env:        prodEnv,
		}
		resp = &AccessTokenResp{}
	)

	result, body, errs := gorequest.New().Timeout(defaultTimeOut).
		Post(url).
		Set("X-BK-APP-CODE", auth.appCode).
		Set("X-BK-APP-SECRET", auth.appSecret).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(req).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api GetAccessToken failed: %v", errs[0])
		return "", errs[0]
	}

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call GetAccessToken API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return "", errMsg
	}

	return resp.Data.AccessToken, nil
}
