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

// Package middleware Authorization
package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	jwtauth "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

const (
	// InnerClientHeaderKey is the key for client in header
	InnerClientHeaderKey = "X-Bcs-Client"
	// AuthorizationHeaderKey is the key for authorization in header
	AuthorizationHeaderKey = "Authorization"
	// CustomUsernameHeaderKey is the key for custom username in header
	CustomUsernameHeaderKey = "X-Bcs-Username"
)

// AuthenticationRequired API类型, 兼容多种认证模式
func AuthenticationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restContext := rest.InitRestContext(w, r)
		r = restContext.Request

		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		switch {
		case initContextWithAPIGW(r, restContext):
		case initContextWithBCSJwt(r, restContext):
		case initContextWithDevEnv(r, restContext):
		default:
			_ = render.Render(w, r, rest.AbortWithUnauthorizedError(restContext, rest.ErrorUnauthorized))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ProjectAuthorization project 鉴权
func ProjectAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restContext, err := rest.GetRestContext(r.Context())
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}

		projectID := restContext.ProjectId
		clusterID := restContext.ClusterId
		user := authutils.UserInfo{
			TenantId:   restContext.TenantId,
			BkUserName: restContext.Username,
		}

		// check cluster
		cls, err := bcs.GetCluster(clusterID)
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}
		if !cls.IsShared && cls.ProjectID != projectID {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, fmt.Errorf("cluster is invalid")))
			return
		}

		// call iam
		client, err := iam.GetProjectPermClient(restContext.TenantId)
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}
		allow, url, _, err := client.CanViewProject(user.BkUserName, projectID)
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}
		if !allow {
			errMsg := fmt.Errorf("permission denied, please apply permission with %s", url)
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, errMsg))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ClusterAuthorization 集群鉴权
func ClusterAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restContext, err := rest.GetRestContext(r.Context())
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}

		projectID := restContext.ProjectId
		clusterID := restContext.ClusterId
		user := authutils.UserInfo{
			TenantId:   restContext.TenantId,
			BkUserName: restContext.Username,
		}

		// check cluster
		cls, err := bcs.GetCluster(clusterID)
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}
		if !cls.IsShared && cls.ProjectID != projectID {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, fmt.Errorf("cluster is invalid")))
			return
		}

		// call iam
		client, err := iam.GetClusterPermClient(restContext.TenantId)
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}
		allow, url, _, err := client.CanViewCluster(user.BkUserName, projectID, clusterID)
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}
		if !allow {
			errMsg := fmt.Errorf("permission denied, please apply permission with %s", url)
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, errMsg))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// initContextWithDevEnv Dev环境, 可以设置环境变量
func initContextWithDevEnv(r *http.Request, c *rest.Context) bool {
	if config.G.Base.RunEnv != config.DevEnv {
		return false
	}

	// 本地用户认证
	username := os.Getenv("BCS_MONITOR_USERNAME")
	if username != "" {
		c.BindEnv = &rest.EnvToken{Username: username}
		c.Username = username
	}

	// AppCode 认证
	appCode := r.Header.Get("X-BKAPI-JWT-APPCODE")
	if appCode != "" {
		c.BindAPIGW = &rest.APIGWToken{
			App: &rest.APIGWApp{AppCode: appCode, Verified: true},
		}
	}

	if username != "" || appCode != "" {
		return true
	}

	return false
}

// BCSJWTDecode BCS 网关 JWT 解码
func BCSJWTDecode(jwtToken string) (*rest.UserClaimsInfo, error) {
	if config.G.BCS.JWTPubKeyObj == nil {
		return nil, errors.New("jwt public key not set")
	}

	token, err := jwt.ParseWithClaims(jwtToken, &rest.UserClaimsInfo{}, func(token *jwt.Token) (interface{}, error) {
		return config.G.BCS.JWTPubKeyObj, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("jwt token not valid")
	}

	claims, ok := token.Claims.(*rest.UserClaimsInfo)
	if !ok {
		return nil, errors.New("jwt token not bcs issuer")

	}
	return claims, nil
}

// BKAPIGWJWTDecode 蓝鲸APIGW JWT 解析
func BKAPIGWJWTDecode(jwtToken string) (*rest.APIGWToken, error) {
	if config.G.BKAPIGW.JWTPubKeyObj == nil {
		return nil, errors.New("jwt public key not set")
	}

	token, err := jwt.ParseWithClaims(jwtToken, &rest.APIGWToken{}, func(token *jwt.Token) (interface{}, error) {
		return config.G.BKAPIGW.JWTPubKeyObj, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("jwt token not valid")
	}

	claims, ok := token.Claims.(*rest.APIGWToken)
	if !ok {
		return nil, errors.New("jwt token not BKAPIGW issuer")

	}
	return claims, nil
}

// initContextWithBCSJwt BCS APISix JWT 鉴权
func initContextWithBCSJwt(r *http.Request, c *rest.Context) bool {
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) == 0 || !strings.HasPrefix(tokenString, "Bearer ") {
		return false
	}
	tokenString = tokenString[7:]

	claims, err := BCSJWTDecode(tokenString)
	if err != nil {
		return false
	}
	username := getUsername(r, claims)

	claims.UserName = username
	c.BindBCS = claims
	c.TenantId = claims.TenantId
	c.Username = username
	return true
}

// getUsername 获取用户名
func getUsername(r *http.Request, claims *rest.UserClaimsInfo) string {
	clientName := ""
	username := claims.UserName
	if claims.SubType == jwtauth.Client.String() {
		clientName = claims.ClientID
	}

	if len(claims.BKAppCode) != 0 {
		clientName = claims.BKAppCode
	}

	if clientName != "" {
		cusUsername := r.Header.Get(CustomUsernameHeaderKey)
		if cusUsername != "" {
			username = cusUsername
		}
	}

	// 优先级 username > clientName > innerClientHeader
	if username != "" {
		return username
	}

	if clientName != "" {
		return clientName
	}

	return r.Header.Get(InnerClientHeaderKey)
}

func initContextWithAPIGW(r *http.Request, c *rest.Context) bool {
	// get jwt info from headers
	tokenString := r.Header.Get("X-Bkapi-Jwt")
	if tokenString == "" {
		return false
	}

	token, err := BKAPIGWJWTDecode(tokenString)
	if err != nil {
		return false
	}

	c.BindAPIGW = token

	return true
}

// GetProjectIdOrCode :
func GetProjectIdOrCode(r *http.Request) string {
	if chi.URLParam(r, "projectId") != "" {
		return chi.URLParam(r, "projectId")
	}
	return ""
}

// GetProjectCode 获取 projectCode 参数
func GetProjectCode(r *http.Request) string {
	if chi.URLParam(r, "projectCode") != "" {
		return chi.URLParam(r, "projectCode")
	}
	return ""
}

// GetClusterId :
func GetClusterId(r *http.Request) string {
	if chi.URLParam(r, "clusterId") != "" {
		return chi.URLParam(r, "clusterId")
	}
	return ""
}

// GetSessionId :
func GetSessionId(r *http.Request) string {
	if chi.URLParam(r, "sessionId") != "" {
		return chi.URLParam(r, "sessionId")
	}
	return ""
}

// ServiceEnable 服务是否可用，如果不可用则直接返回空数据
func ServiceEnable(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.G.Base.ServiceEnable {
			next.ServeHTTP(w, r)
			return
		}
		w.Write([]byte(`{"code": 0, "data": {}, "message": "service is not available"}`))
		w.WriteHeader(http.StatusOK)
	})
}
