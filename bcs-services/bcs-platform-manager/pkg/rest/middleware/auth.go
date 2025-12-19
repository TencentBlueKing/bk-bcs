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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/rest"
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
		username := restContext.Username

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
		client, err := iam.GetProjectPermClient()
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}
		allow, url, _, err := client.CanViewProject(username, projectID)
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
		username := restContext.Username

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
		client, err := iam.GetClusterPermClient()
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}
		allow, url, _, err := client.CanViewCluster(username, projectID, clusterID)
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
	username := os.Getenv("BCS_PLATFORM_MANAGER_USERNAME")
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

	c.BindBCS = claims
	c.Username = claims.UserName
	return true
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
