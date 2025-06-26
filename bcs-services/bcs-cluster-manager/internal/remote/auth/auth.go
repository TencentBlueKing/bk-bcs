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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/types"
)

var (
	defaultTimeOut   = time.Second * 60
	errServerNotInit = errors.New("server not inited")
)

var accessClient *ClientAuth

// SetAccessClient set access token client
func SetAccessClient(options Options) error {
	cli := NewAccessClient(options)

	accessClient = cli
	return nil
}

// GetAccessClient get access token client
func GetAccessClient() *ClientAuth {
	return accessClient
}

// NewAccessClient init access client
func NewAccessClient(opt Options) *ClientAuth {
	cli := &ClientAuth{
		server: opt.Server,
		debug:  opt.Debug,
	}
	return cli
}

// Options opts parameter
type Options struct {
	// Server auth address
	Server string
	// Debug http debug
	Debug bool
}

// ClientAuth perm client
type ClientAuth struct {
	server string
	debug  bool
}

// GetAccessToken get application access token
func (auth *ClientAuth) GetAccessToken(app types.BkAppUser) (string, error) {
	if options.GetEditionInfo().IsCommunicationEdition() || options.GetEditionInfo().IsEnterpriseEdition() {
		return auth.GetAccessTokenBySsm(app)
	}

	return auth.GetAccessTokenByGateWay(app)
}

// GetAccessTokenBySsm by ssm for Communication/EnterpriseEdition
func (auth *ClientAuth) GetAccessTokenBySsm(app types.BkAppUser) (string, error) {
	if auth == nil {
		return "", errServerNotInit
	}

	var (
		_    = "GetAccessTokenBySsm"
		path = "/api/v1/auth/access-tokens"
	)

	var (
		url = auth.server + path
		req = &AccessSsmRequest{
			AppCode:    app.BkAppCode,
			AppSecret:  app.BkAppSecret,
			IDProvider: "client",
			GrantType:  "client_credentials",
			Env:        "prod",
		}
		resp = &AccessTokenSsmResp{}
	)

	start := time.Now()
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(url).
		Set("X-BK-APP-CODE", app.BkAppCode).
		Set("X-BK-APP-SECRET", app.BkAppSecret).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(req).
		EndStruct(resp)

	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("auth", "GetAccessTokenBySsm", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api GetAccessTokenBySsm failed: %v", errs[0])
		return "", errs[0]
	}
	metrics.ReportLibRequestMetric("auth", "GetAccessTokenBySsm", "http", metrics.LibCallStatusOK, start)

	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("call GetAccessTokenBySsm API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return "", errMsg
	}

	return resp.Data.AccessToken, nil
}

// GetAccessTokenByGateWay get application accessToken for interEdition
func (auth *ClientAuth) GetAccessTokenByGateWay(app types.BkAppUser) (string, error) {
	if auth == nil {
		return "", errServerNotInit
	}

	var (
		_    = "GetAccessTokenByGateWay"
		path = "/auth_api/token/"
	)

	var (
		url = auth.server + path
		req = &AccessGateWayRequest{
			AppCode:    app.BkAppCode,
			AppSecret:  app.BkAppSecret,
			IDProvider: "client",
			GrantType:  "client_credentials",
			Env:        "prod",
		}
		resp = &AccessTokenGateWayResp{}
	)

	start := time.Now()
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(req).
		EndStruct(resp)

	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("auth", "GetAccessTokenByGateWay", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api GetAccessTokenByGateWay failed: %v", errs[0])
		return "", errs[0]
	}
	metrics.ReportLibRequestMetric("auth", "GetAccessTokenByGateWay", "http", metrics.LibCallStatusOK, start)

	if result.StatusCode != http.StatusOK || resp.Code != "0" {
		errMsg := fmt.Errorf("call GetAccessTokenByGateWay API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return "", errMsg
	}

	return resp.Data.AccessToken, nil
}
