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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-chi/render"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// ProjectParse 解析 project
func ProjectParse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restContext, err := rest.GetRestContext(r.Context())
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithBadRequestError(rest.InitRestContext(w, r), err))
			return
		}

		// get project code
		projectIDOrCode := restContext.ProjectId
		if len(restContext.ProjectCode) != 0 {
			projectIDOrCode = restContext.ProjectCode
		}
		project, err := bcs.GetProject(r.Context(), config.G.BCS, projectIDOrCode)
		if err != nil {
			blog.Errorf("get project error for project %s, error: %s", projectIDOrCode, err.Error())
			_ = render.Render(w, r, rest.AbortWithBadRequestError(restContext, err))
			return
		}
		restContext.ProjectId = project.ProjectId
		restContext.ProjectCode = project.Code

		// get cluster info
		cls, err := bcs.GetCluster(restContext.ClusterId)
		if err != nil {
			_ = render.Render(w, r, rest.AbortWithWithForbiddenError(restContext, err))
			return
		}
		restContext.SharedCluster = cls.IsShared

		next.ServeHTTP(w, r)
	})
}
