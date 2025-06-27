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

package types

import (
	"errors"
	"time"
)

var (
	// DefaultTimeOut default timeout
	DefaultTimeOut = time.Second * 60
	// ErrServerNotInit server not init
	ErrServerNotInit = errors.New("server not inited")
)

// BkAccessToken bk app token
type BkAccessToken struct {
	AccessToken string `json:"access_token"`
}

// BkAppUser appCode/appSecret
type BkAppUser struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
}

// AuthInfo auth user for gateway
type AuthInfo struct {
	BkAppUser
	BkUserName  string `json:"bk_username,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

// CommonClient client common section
type CommonClient struct {
	AppCode   string
	AppSecret string
	Server    string
	Debug     bool
}

// BaseResponse baseResp
type BaseResponse struct {
	Code      int    `json:"code"`
	Result    bool   `json:"result"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}
