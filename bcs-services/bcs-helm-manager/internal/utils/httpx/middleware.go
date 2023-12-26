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
	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	projectClient "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
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

		pj, err := projectClient.GetProjectByCode(projectCode)
		if err != nil {
			msg := fmt.Errorf("ParseProjectID get projectID error, projectCode: %s, err: %s", projectCode, err.Error())
			ResponseSystemError(w, r, msg)
			return
		}

		ctx := context.WithValue(r.Context(), contextx.ProjectCodeContextKey, pj.ProjectCode)
		ctx = context.WithValue(ctx, contextx.ProjectIDContextKey, pj.ProjectID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// AuthorizationMiddleware authorization
func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// skip handler
		if !options.GlobalOptions.JWT.Enable {
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

		resourceID := options.CredentialScope{}
		resourceID.ProjectID = contextx.GetProjectIDFromCtx(r.Context())

		allow, url, resources, err := auth.CallIAM(authUser.GetUsername(), project.CanViewProjectOperation, resourceID)
		if err != nil {
			ResponseAuthError(w, r, err)
			return
		}
		if !allow {
			ResponsePermissionError(w, r, &authutils.PermDeniedError{
				Perms: authutils.PermData{
					ApplyURL:   url,
					ActionList: resources,
				},
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func parseJwtToken(jwtToken string) (*jwt.UserClaimsInfo, error) {
	if len(jwtToken) == 0 || !strings.HasPrefix(jwtToken, "Bearer ") {
		return nil, errors.New("authorization token error")
	}
	claims, err := auth.GetJWTClient().JWTDecode(jwtToken[7:])
	if err != nil {
		return nil, err
	}
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("authorization token expired")
	}
	return claims, nil
}

// getRequestID 获取 request id
func getRequestID(req *http.Request) string {
	// 当request id不存在或者为空时，生成id
	requestID := req.Header.Get(contextx.RequestIDHeaderKey)
	if requestID == "" {
		return stringx.GenUUID()
	}

	return requestID
}
