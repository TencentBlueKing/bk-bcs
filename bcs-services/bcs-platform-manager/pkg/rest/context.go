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

// Package rest rest context
package rest

import (
	"fmt"
	"net/http"
	"time"

	httpin_integration "github.com/ggicci/httpin/integration"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
)

func init() {
	httpin_integration.UseGochiURLParam("path", chi.URLParam)
}

// Context xxx
type Context struct {
	*http.Request
	RequestId     string          `json:"request_id"`
	StartTime     time.Time       `json:"start_time"`
	Operator      string          `json:"operator"`
	Username      string          `json:"username"`
	ProjectId     string          `json:"project_id"`
	ProjectCode   string          `json:"project_code"`
	ClusterId     string          `json:"cluster_id"`
	SharedCluster bool            `json:"shared_cluster"`
	HandleName    string          `json:"handle_name"`
	BindEnv       *EnvToken       `json:"bind_env"`
	BindBCS       *UserClaimsInfo `json:"bind_bcs"`
	BindAPIGW     *APIGWToken     `json:"bind_apigw"`
}

// WriteAttachment 提供附件下载能力
func WriteAttachment(w http.ResponseWriter, data []byte, filename string) {
	w.Header().Set("Content-Type", "application/octet-stream")
	attachment := fmt.Sprintf("attachment; filename=%s", filename)
	w.Header().Set("Content-Disposition", attachment)
	w.Write(data)
}

// EnvToken xxx
type EnvToken struct {
	Username string
}

// APIGWApp xxx
type APIGWApp struct {
	AppCode  string `json:"app_code"`
	Verified bool   `json:"verified"`
}

// APIGWUser xxx
type APIGWUser struct {
	Username string `json:"username"`
	Verified bool   `json:"verified"`
}

// APIGWToken 返回信息
type APIGWToken struct {
	App  *APIGWApp  `json:"app"`
	User *APIGWUser `json:"user"`
	*jwt.StandardClaims
}

// String :
func (a *APIGWToken) String() string {
	return fmt.Sprintf("<%s, %v>", a.App.AppCode, a.App.Verified)
}

// UserClaimsInfo custom jwt claims
type UserClaimsInfo struct {
	SubType      string `json:"sub_type"`
	UserName     string `json:"username"`
	BKAppCode    string `json:"bk_app_code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	// https://tools.ietf.org/html/rfc7519#section-4.1
	// aud: 接收jwt一方; exp: jwt过期时间; jti: jwt唯一身份认证; IssuedAt: 签发时间; Issuer: jwt签发者
	*jwt.StandardClaims
}
