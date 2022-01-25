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

package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/parnurzeal/gorequest"
)

var (
	defaultTimeOut   = time.Second * 60
	errServerNotInit = errors.New("server not inited")
)

var ssmClient *ClientSSM

// SetSSMClient set ssm client
func SetSSMClient(options Options) error {
	if !options.Enable {
		ssmClient = nil
		return nil
	}
	cli := NewSSMClient(options)

	ssmClient = cli
	return nil
}

// GetSSMClient get perm client
func GetSSMClient() *ClientSSM {
	return ssmClient
}

// NewSSMClient init SSM client
func NewSSMClient(opt Options) *ClientSSM {
	cli := &ClientSSM{
		server:    opt.Server,
		appCode:   opt.AppCode,
		appSecret: opt.AppSecret,
		debug:     opt.Debug,
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
	// Debug http debug
	Debug bool
}

// ClientSSM ssm client
type ClientSSM struct {
	server string

	appCode   string
	appSecret string
	debug     bool
}

// GetAccessToken get access token
func (ssm *ClientSSM) GetAccessToken() (string, error) {
	if ssm == nil {
		return "", errServerNotInit
	}

	const (
		apiName = "GetAccessToken"
		path    = "/api/v1/auth/access-tokens"
	)

	var (
		url = ssm.server + path
		req = &AccessRequest{
			AppCode:    ssm.appCode,
			AppSecret:  ssm.appSecret,
			IDProvider: "client",
			GrantType:  "client_credentials",
			Env:        "prod",
		}
		resp = &AccessTokenResp{}
	)

	result, body, errs := gorequest.New().Timeout(defaultTimeOut).
		Post(url).
		Set("X-BK-APP-CODE", ssm.appCode).
		Set("X-BK-APP-SECRET", ssm.appSecret).
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
