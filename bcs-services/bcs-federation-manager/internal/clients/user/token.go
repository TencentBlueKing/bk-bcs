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

// Package user xxx
package user

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

const (
	// GetUserTokenPath get user token path
	GetUserTokenPath = "/bcsapi/v4/usermanager/v1/users/%s/tokens"
)

// CommonResult common result
type CommonResult struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// TokenResp token response
type TokenResp struct {
	Token     string     `json:"token"`
	ExpiredAt *time.Time `json:"expired_at"` // nil means never expired
}

// GetUserTokenResp result for get user token
type GetUserTokenResp struct {
	CommonResult `json:",inline"`
	Data         []TokenResp `json:"data"`
}

// GetUserToken get user token
func (h userClient) GetUserToken(username string) (string, error) {
	blog.Infof("get user token for user: %s", username)

	url := fmt.Sprintf("%s%s", h.opt.Endpoint, fmt.Sprintf(GetUserTokenPath, username))

	raw, err := h.opt.Sender.DoGetRequest(url, h.defaultHeader)
	if err != nil {
		return "", fmt.Errorf("GetUserToken failed when DoGetRequest error: %s", err)
	}

	var resp GetUserTokenResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", fmt.Errorf("GetUserToken decode GetUserTokenResult response failed %s,"+
			" raw response %s", err.Error(), string(raw))
	}

	if resp.Code != 0 || !resp.Result {
		return "", fmt.Errorf("GetUserToken failed, code: %d, message: %s", resp.Code, resp.Message)
	}

	if len(resp.Data) == 0 {
		blog.Warnf("GetUserToken empty, user[%s] do not have token", username)
		return "", nil
	}

	return resp.Data[0].Token, nil
}
