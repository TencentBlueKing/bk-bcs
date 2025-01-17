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

// Package httpx xxx
package httpx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	bcsJwt "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	jwtGo "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam"
	projectAuth "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/contextx"
)

// LoggingMiddleware log http request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), contextx.LangContectKey, i18n.GetLangFromReqCookies(r))
		r = r.WithContext(ctx)
		blog.Infof("request_id %s, method %s, url %s", getRequestID(r), r.Method, r.URL.String())
		next.ServeHTTP(w, r)
	})
}

// AuthenticationMiddleware authentication
func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authUser := middleauth.AuthUser{}
		// parse client token from header
		clientName := r.Header.Get(middleauth.InnerClientHeaderKey)
		if clientName != "" {
			authUser.ClientName = clientName
		}

		// parse user token from header
		jwtToken := r.Header.Get(middleauth.AuthorizationHeaderKey)
		if jwtToken != "" {
			u, err := parseJwtToken(jwtToken)
			if err != nil {
				ResponseAuthError(w, r, err)
				return
			}
			// !NOTO: bk-apigw would set SubType to "user" even if use client's app code and secret
			if u.SubType == jwt.User.String() {
				authUser.Username = u.UserName
			}
			if u.SubType == jwt.Client.String() {
				authUser.ClientName = u.ClientID
			}
			if len(u.BKAppCode) != 0 {
				authUser.ClientName = u.BKAppCode
			}
		}

		// If and only if client name from jwt token is not empty, we will check username in header
		if authUser.ClientName != "" {
			username := r.Header.Get(middleauth.CustomUsernameHeaderKey)
			if username != "" {
				authUser.Username = username
			}
		}

		// set auth user to context
		ctx := context.WithValue(r.Context(), middleauth.AuthUserKey, authUser)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// ParseProjectIDMiddleware parse projectID
func ParseProjectIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		projectCode := vars["projectCode"]
		if len(projectCode) == 0 {
			blog.Warn("ParseProjectID error: projectCode is empty")
			next.ServeHTTP(w, r)
			return
		}

		pj, err := project.GetProjectInfo(r.Context(), projectCode)
		if err != nil {
			msg := fmt.Errorf("ParseProjectID get projectID error, projectCode: %s, err: %s", projectCode, err.Error())
			ResponseSystemError(w, r, msg)
			return
		}

		ctx := context.WithValue(r.Context(), contextx.ProjectCodeContextKey, pj.Code)
		ctx = context.WithValue(ctx, contextx.ProjectIDContextKey, pj.ID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// ParseClusterIDMiddleware parse clusterID
func ParseClusterIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clusterID := vars["clusterID"]
		if clusterID == "" {
			blog.Warn("clusterID error: clusterID is empty")
			next.ServeHTTP(w, r)
			return
		}

		cluster, err := cluster.GetClusterInfo(r.Context(), clusterID)
		if err != nil {
			msg := fmt.Errorf("ParseClusterID get clusterID error, clusterID: %s, err: %s", clusterID, err.Error())
			ResponseSystemError(w, r, msg)
			return
		}

		projectID := contextx.GetProjectIDFromCtx(r.Context())

		if !cluster.IsShared && cluster.ProjID != projectID {
			msg := fmt.Errorf("cluster is invalid")
			ResponseSystemError(w, r, msg)
			return
		}

		ctx := context.WithValue(r.Context(), contextx.ClusterIDContextKey, cluster.ID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// AuthorizationMiddleware authorization
func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// skip handler
		if config.G.Auth.Disabled {
			next.ServeHTTP(w, r)
			return
		}

		authUser, err := middleauth.GetUserFromContext(r.Context())
		if err != nil {
			ResponseAuthError(w, r, err)
			return
		}

		if authUser.IsInner() {
			next.ServeHTTP(w, r)
			return
		}

		projectID := contextx.GetProjectIDFromCtx(r.Context())

		// 权限控制为项目查看
		permCtx := &projectAuth.PermCtx{
			Username:  authUser.Username,
			ProjectID: projectID,
		}
		if allow, err := iam.NewProjectPerm().CanView(permCtx); err != nil {
			ResponseAuthError(w, r, err)
			return
		} else if !allow {
			ResponseAuthError(w, r, errors.New(i18n.GetMsg(r.Context(), "无项目查看权限")))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseJwtToken(jwtToken string) (*jwt.UserClaimsInfo, error) {
	if len(jwtToken) == 0 || !strings.HasPrefix(jwtToken, "Bearer ") {
		return nil, errors.New("authorization token error")
	}
	claims, err := jwtDecode(jwtToken[7:])
	if err != nil {
		return nil, err
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("authorization token expired")
	}
	return claims, nil
}

// jwtDecode 解析 jwt
func jwtDecode(jwtToken string) (*bcsJwt.UserClaimsInfo, error) {
	if config.G.Auth.JWTPubKeyObj == nil {
		return nil, errors.New("jwt public key uninitialized")
	}

	token, err := jwtGo.ParseWithClaims(
		jwtToken,
		&bcsJwt.UserClaimsInfo{},
		func(token *jwtGo.Token) (interface{}, error) {
			return config.G.Auth.JWTPubKeyObj, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("jwt token invalid")
	}

	claims, ok := token.Claims.(*bcsJwt.UserClaimsInfo)
	if !ok {
		return nil, errors.New("jwt token's issuer isn't bcs")
	}
	return claims, nil
}

// getRequestID 获取 request id
func getRequestID(req *http.Request) string {
	// 当request id不存在或者为空时，生成id
	requestID := req.Header.Get(contextx.RequestIDHeaderKey)
	if requestID == "" {
		return uuid.New().String()
	}

	return requestID
}
