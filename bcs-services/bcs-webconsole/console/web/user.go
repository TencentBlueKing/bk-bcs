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

package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/rest"
)

// UserPermRequestRedirect 用户权限申请URL
func (s *service) UserPermRequestRedirect(c *gin.Context) {
	projectId := c.Query("project_id")
	clusterId := c.Query("cluster_id")
	if projectId == "" {
		rest.APIError(c, i18n.T(c, "project_id is required"))
		return
	}
	project, err := bcs.GetProject(c.Request.Context(), config.G.BCS, projectId)
	if err != nil {
		rest.APIError(c, i18n.T(c, "项目不正确"))
		return
	}

	redirectUrl, err := iam.MakeResourceApplyUrl(c.Request.Context(),
		project.ProjectId, clusterId, "", "", project.TenantID)
	if err != nil {
		rest.APIError(c, i18n.T(c, "%s", err))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}
