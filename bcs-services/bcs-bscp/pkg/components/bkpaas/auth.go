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

	"github.com/pkg/errors"

	"bscp.io/pkg/components"
)

type UserInfo struct {
	UserName string `json:"username"`
}

// getBKUserInfoByToken 外部统一鉴权
func GetBKUserInfoByToken(ctx context.Context, host, uid, token string) (*UserInfo, error) {
	url := fmt.Sprintf("%s/login/accounts/is_login/", host)
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetQueryParam("bk_token", token).
		Get(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	user := new(UserInfo)
	if err := components.UnmarshalBKResult(resp, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetBKUserInfoByTicket bicket统一鉴权
func GetBKUserInfoByTicket(ctx context.Context, host, uid, token string) (*UserInfo, error) {
	url := fmt.Sprintf("%s/user/is_login/", host)
	resp, err := components.GetClient().R().
		SetContext(ctx).
		SetQueryParam("bk_ticket", token).
		Get(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, errors.Errorf("http code %d != 200, body: %s", resp.StatusCode(), resp.Body())
	}

	user := new(UserInfo)
	if err := components.UnmarshalBKResult(resp, user); err != nil {
		return nil, err
	}

	return user, nil
}
