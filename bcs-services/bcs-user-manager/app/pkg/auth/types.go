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

// GrantType xxx
type GrantType string

// String xxx
func (gt GrantType) String() string {
	return string(gt)
}

var (
	clientCredentialGrant GrantType = "client_credentials"
)

const (
	clientProvider = "client"
	prodEnv        = "prod"
)

// CommonResp common resp
type CommonResp struct {
	Code      uint   `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// AccessRequest request
type AccessRequest struct {
	AppCode    string `json:"app_code"`
	AppSecret  string `json:"app_secret"`
	IDProvider string `json:"id_provider"`
	GrantType  string `json:"grant_type"`
	Env        string `json:"env"`
}

// AccessTokenResp response
type AccessTokenResp struct {
	CommonResp
	Data *AccessTokenInfo `json:"data"`
}

// AccessTokenInfo data
type AccessTokenInfo struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    uint32 `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}
