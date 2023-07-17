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

package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/auth"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/bcs"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/iam"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/rest"
)

// NeedProjectAuthorization middleware for project authorization
func NeedProjectAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		claims, err := decodeBCSJwtFromContext(ctx, r)
		if err != nil {
			rest.AbortWithUnauthorized(w, r, http.StatusUnauthorized, err.Error())
			return
		}
		if claims.UserName == "" {
			rest.AbortWithUnauthorized(w, r, http.StatusUnauthorized, "auth failed, username is empty")
			return
		}

		projectCode := r.URL.Query().Get("projectCode")
		if projectCode == "" {
			rest.AbortWithBadRequest(w, r, http.StatusBadRequest, "projectCode is empty")
			return
		}
		project, err := bcs.GetProject(ctx, projectCode)
		if err != nil {
			rest.AbortWithBadRequest(w, r, http.StatusBadRequest, err.Error())
			return
		}
		// iam 鉴权，先校验 cluster_view 权限，因为 cluster_view 权限包含 project_view 权限，
		// 避免用户申请 project_view 之后还要再单独申请 cluster_view 权限
		// 如果有 clusterID 参数，先校验 cluster_view 权限
		clusterID := r.URL.Query().Get("clusterID")
		if clusterID != "" {
			_, err := bcs.GetCluster(ctx, clusterID)
			if err != nil {
				rest.AbortWithInternalServerError(w, r, http.StatusInternalServerError, err.Error())
				return
			}
			client, err := iam.GetClusterPermClient()
			allow, url, actionList, err := client.CanViewCluster(claims.UserName, project.ProjectID, clusterID)
			if err != nil {
				rest.AbortWithInternalServerError(w, r, http.StatusInternalServerError, err.Error())
				return
			}
			if !allow {
				rest.AbortWithForbidden(w, r, &rest.Perms{
					ActionList: actionList,
					ApplyURL:   url,
				})
				return
			}
		}
		// 校验 project_view 权限
		client, err := iam.GetProjectPermClient()
		if err != nil {
			rest.AbortWithBadRequest(w, r, http.StatusBadRequest, err.Error())
			return
		}
		allow, url, actionList, err := client.CanViewProject(claims.UserName, project.ProjectID)
		if err != nil {
			rest.AbortWithInternalServerError(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		if !allow {
			rest.AbortWithForbidden(w, r, &rest.Perms{
				ActionList: actionList,
				ApplyURL:   url,
			})
			return
		}

		// pass the span through the request context
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}

func decodeBCSJwtFromContext(ctx context.Context, r *http.Request) (*auth.UserClaimsInfo, error) {
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) == 0 || !strings.HasPrefix(tokenString, "Bearer ") {
		return nil, errors.New("auth token not found")
	}
	tokenString = tokenString[7:]

	claims, err := BCSJWTDecode(tokenString)
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// BCSJWTDecode BCS 网关 JWT 解码
func BCSJWTDecode(jwtToken string) (*auth.UserClaimsInfo, error) {
	if config.G.BCS.JWTPubKeyObj == nil {
		return nil, errors.New("jwt public key not set")
	}

	token, err := jwt.ParseWithClaims(jwtToken, &auth.UserClaimsInfo{}, func(token *jwt.Token) (interface{}, error) {
		return config.G.BCS.JWTPubKeyObj, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("jwt token not valid")
	}

	claims, ok := token.Claims.(*auth.UserClaimsInfo)
	if !ok {
		return nil, errors.New("jwt token not bcs issuer")

	}
	return claims, nil
}
