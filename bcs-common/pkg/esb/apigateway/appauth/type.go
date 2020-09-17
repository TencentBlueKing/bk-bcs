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

package appauth

const (
	// PaasGrantTypeClient paas grant type
	PaasGrantTypeClient = "client_credentials"
	// PaasIDProviderClient paaas id provider
	PaasIDProviderClient = "client"
	// PaasOAuthEnvProd paas environment prod
	PaasOAuthEnvProd = "prod"
	// PaasOAuthEnvTest paas environment test
	PaasOAuthEnvTest = "test"
)

// OAuthRequest auth request
type OAuthRequest struct {
	//EnvName now "prod", "test" are available
	EnvName string `json:"env_name"`
	//code & secret are for PaaS authorization
	AppCode   string `json:"app_code"`
	AppSecret string `json:"app_secret"`
	//default client_credentials
	GrantType string `json:"grant_type"`
}

// OAuthData auth data
type OAuthData struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	UserType     string `json:"user_type"`
	Scope        string `json:"scope"`
}

// OAuthResponse auth response
type OAuthResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    *OAuthData `json:"data"`
}
