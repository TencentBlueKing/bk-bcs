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

package bkpaas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components"
)

type bkLoginResult struct {
	Msg  string `json:"msg"`
	Code int    `json:"ret"`
}

// bkLoginAuthClient 蓝鲸内部统一登入
type bkLoginAuthClient struct {
	conf *cc.LoginAuthSettings
}

// GetLoginCredentialFromCookies 从 cookie 获取 LoginCredential
func (b *bkLoginAuthClient) GetLoginCredentialFromCookies(r *http.Request) (*LoginCredential, error) {
	uid, err := r.Cookie("bk_uid")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, fmt.Errorf("%s cookie not present", "bk_uid")
		}
		return nil, err
	}

	token, err := r.Cookie("bk_ticket")
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, fmt.Errorf("%s cookie not present", "bk_ticket")
		}
		return nil, err
	}

	return &LoginCredential{UID: uid.Value, Token: token.Value}, nil
}

// GetUserInfoByToken BK_LIGIN 统一登入服务 bk_ticket 统一鉴权
func (b *bkLoginAuthClient) GetUserInfoByToken(ctx context.Context, host, uid, token string) (string, error) {
	url := fmt.Sprintf("%s/user/is_login/", host)
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetQueryParam("bk_ticket", token).
		Get(url)

	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	result := new(bkLoginResult)
	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return "", err
	}

	if result.Code != 0 {
		return "", errors.Errorf("ret code %d != 0, body: %s", result.Code, resp.Body())
	}

	return uid, nil
}

// BuildLoginRedirectURL 登入跳转URL
func (b *bkLoginAuthClient) BuildLoginRedirectURL(r *http.Request, webHost string) string {
	redirectURL := fmt.Sprintf("%s/?c_url=%s", b.conf.Host, url.QueryEscape(buildAbsoluteUri(webHost, r)))
	return redirectURL
}

// BuildLoginURL API未登入访问URL
func (b *bkLoginAuthClient) BuildLoginURL(r *http.Request) (string, string) {
	loginURL := fmt.Sprintf("%s/?c_url=", b.conf.Host)
	loginPlainURL := fmt.Sprintf("%s/plain/?c_url=", b.conf.Host)
	return loginURL, loginPlainURL
}
