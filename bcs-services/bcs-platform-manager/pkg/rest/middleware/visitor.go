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

package middleware

import (
	"net/http"

	"github.com/go-chi/render"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"
)

// VisitorsRequired 访问人员校验
func VisitorsRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restContext, err := rest.GetRestContext(r.Context())
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithBadRequestError(rest.InitRestContext(w, r), err))
			return
		}

		if restContext.Username == "" || !utils.StringInSlice(restContext.Username, config.G.Base.VisitorsWhiteList) {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, rest.ErrorNoPermission))
			return
		}
		next.ServeHTTP(w, r)
	})
}
