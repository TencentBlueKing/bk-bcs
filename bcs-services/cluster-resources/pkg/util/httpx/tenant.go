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

package httpx

import (
	"fmt"
	"net/http"
	"strings"

	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"github.com/gorilla/mux"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

// TenantMiddleware 租户校验中间件
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !config.G.BCSAPIGW.EnableMultiTenantMode {
			next.ServeHTTP(w, r)
			return
		}

		user, err := middleauth.GetUserFromContext(r.Context())
		if err != nil {
			ResponseAuthError(w, r, err)
			return
		}

		// skip method tenant validation
		if SkipMethod(r) {
			next.ServeHTTP(w, r)
			return
		}

		// exempt client
		if SkipTenantValidation(r, user.ClientName) {
			next.ServeHTTP(w, r)
			return
		}

		// get tenant id
		resourceTenantId, err := getTenantldByResource(r)
		if err != nil {
			ResponseAuthError(w, r, err)
			return
		}

		if user.GetTenantId() != resourceTenantId {
			err = fmt.Errorf("user[%s] tenant[%s] not match resource tenant[%s]",
				user.Username, user.GetTenantId(), resourceTenantId)
			ResponseAuthError(w, r, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// TenantClientWhiteList tenant client white list
var TenantClientWhiteList = map[string][]string{}

// NoCheckTenantMethod no check tenant method
var NoCheckTenantMethod = []string{}

// SkipMethod skip method tenant validation
func SkipMethod(req *http.Request) bool {
	for _, v := range NoCheckTenantMethod {
		// 获取原始uri
		path, err := mux.CurrentRoute(req).GetPathTemplate()
		if err != nil {
			klog.Errorf("get path template err: %s", err)
			return false
		}
		if v == path {
			return true
		}
	}
	return false
}

// SkipTenantValidation skip tenant validation
func SkipTenantValidation(req *http.Request, client string) bool {
	if len(client) == 0 {
		return false
	}
	for _, v := range TenantClientWhiteList[client] {
		// 获取原始uri
		path, err := mux.CurrentRoute(req).GetPathTemplate()
		if err != nil {
			klog.Errorf("get path template err: %s", err)
			return false
		}
		if strings.HasPrefix(v, "*") || v == path {
			return true
		}
	}
	return false
}

// getTenantldByResource get tenant id by resource
func getTenantldByResource(req *http.Request) (string, error) {
	vars := mux.Vars(req)
	projectCode := vars["projectCode"]
	project, err := project.GetProjectInfo(req.Context(), projectCode)
	if err != nil {
		return "", err
	}

	return project.TenantID, nil
}
