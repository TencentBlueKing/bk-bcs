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
	"net/url"

	"github.com/pkg/errors"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/components"
)

type bkLoginAuthClient struct {
	conf *cc.LoginAuthSettings
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

	user := new(userInfo)
	if err := components.UnmarshalBKResult(resp, user); err != nil {
		return "", err
	}

	return user.Username, nil
}

// BuildLoginRedirectURL 登入跳转URL
func (b *bkLoginAuthClient) BuildLoginRedirectURL(r *http.Request, webHost string) string {
	redirectURL := fmt.Sprintf("%s/?c_url=%s", b.conf.Host, url.QueryEscape(buildAbsoluteUri(webHost, r)))
	return redirectURL
}

// GetLoginCredentialFromCookies 从 cookie 获取 LoginCredential
func (b *bkLoginAuthClient) GetLoginCredentialFromCookies(r *http.Request) (*LoginCredential, error) {
	uid, err := r.Cookie("bk_uid")
	if err != nil {
		return nil, err
	}

	token, err := r.Cookie("bk_ticket")
	if err != nil {
		return nil, err
	}

	return &LoginCredential{UID: uid.Value, Token: token.Value}, nil
}
