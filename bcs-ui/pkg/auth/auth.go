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

// Package auth xxx
package auth

import "github.com/golang-jwt/jwt"

// UserClaimsInfo custom jwt claims
type UserClaimsInfo struct {
	SubType      string `json:"sub_type"`
	UserName     string `json:"username"`
	BKAppCode    string `json:"bk_app_code"`
	ClientID     string `json:"client_id"`
	TenantID     string `json:"tenant_id"`
	ClientSecret string `json:"client_secret"`
	// https://tools.ietf.org/html/rfc7519#section-4.1
	// aud: 接收jwt一方; exp: jwt过期时间; jti: jwt唯一身份认证; IssuedAt: 签发时间; Issuer: jwt签发者
	*jwt.StandardClaims
}
