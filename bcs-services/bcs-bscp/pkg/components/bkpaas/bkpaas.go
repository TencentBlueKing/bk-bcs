/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package bkpaas

import (
	"context"
	"fmt"
	"net/http"

	"bscp.io/pkg/cc"
)

type userInfo struct {
	Username string `json:"username"`
}

// UserCredential
type LoginCredential struct {
	UID   string
	Token string
}

// AuthLoginClient 登入鉴权
type AuthLoginClient interface {
	GetLoginCredentialFromCookies(r *http.Request) (*LoginCredential, error)
	GetUserInfoByToken(ctx context.Context, host, uid, token string) (string, error)
	BuildLoginRedirectURL(r *http.Request, webHost string) string
}

// NewAuthLoginClient init client
func NewAuthLoginClient(conf *cc.LoginAuthSettings) AuthLoginClient {
	// 部分场景 conf 可为空
	if conf == nil {
		return nil
	}

	if conf.Provider == "BK_LOGIN" {
		return &bkLoginAuthClient{conf: conf}
	}
	return &bkPaaSAuthClient{conf: conf}
}

// BuildAbsoluteUri
func buildAbsoluteUri(webHost string, r *http.Request) string {
	// fallback use request host
	if webHost == "" {
		webHost = "http://" + r.Host
	}

	return fmt.Sprintf("%s%s", webHost, r.RequestURI)
}

// BuildLoginURL 返回前端的URL
func BuildLoginURL(r *http.Request, Loginhost string) string {
	u := fmt.Sprintf("%s/login/?c_url=", Loginhost)
	return u
}

// BuildLoginPlainURL 返回前端的URL
func BuildLoginPlainURL(r *http.Request, Loginhost string) string {
	u := fmt.Sprintf("%s/login/plain/?c_url=", Loginhost)
	return u
}
