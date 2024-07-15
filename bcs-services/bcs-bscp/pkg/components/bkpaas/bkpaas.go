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

// Package bkpaas provides bkpaas auth client.
package bkpaas

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
)

const (
	// BKLoginProvider 蓝鲸内部统一登入
	BKLoginProvider = "BK_LOGIN"
	// BKPaaSProvider 外部统一登入, 可使用主域名或者ESB查询
	BKPaaSProvider = "BK_PAAS"
)

// LoginCredential uid/token for grpc auth
type LoginCredential struct {
	UID   string
	Token string
}

// AuthLoginClient 登入鉴权
type AuthLoginClient interface {
	GetLoginCredentialFromCookies(r *http.Request) (*LoginCredential, error)
	GetUserInfoByToken(ctx context.Context, host, uid, token string) (string, error)
	BuildLoginRedirectURL(r *http.Request, webHost string) string
	BuildLoginURL(r *http.Request) (string, string)
}

// NewAuthLoginClient init client
func NewAuthLoginClient(conf *cc.LoginAuthSettings) AuthLoginClient {
	if conf.Provider == BKLoginProvider {
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
