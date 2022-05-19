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

package route

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// PermissionRequired 权限控制，必须都为真才可以
func PermissionRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		authCtx := MustGetAuthContext(c)

		// 校验项目，集群信息的正确性
		if err := ValidateProjectCluster(c, authCtx); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Set("auth_context", authCtx)

		// 管理员不校验权限, 包含管理员凭证
		if config.G.IsManager(authCtx.Username, authCtx.ClusterId) {
			c.Next()
			return
		}

		if err := initContextWithIAMProject(c, authCtx); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Next()
	}
}

func ValidateProjectCluster(c *gin.Context, authCtx *AuthContext) error {
	projectId := GetProjectIdOrCode(c)
	if projectId == "" {
		return errors.New("project_id or code is required")
	}

	clusterId := GetClusterId(c)
	if clusterId == "" {
		return errors.New("clusterId required")
	}

	project, err := bcs.GetProject(c.Request.Context(), projectId)
	if err != nil {
		return errors.Wrap(err, "项目不正确")
	}

	bcsConf := k8sclient.GetBCSConfByClusterId(clusterId)

	cluster, err := bcs.GetCluster(c.Request.Context(), bcsConf, project.ProjectId, clusterId)
	if err != nil {
		return errors.Wrap(err, "项目或者集群Id不正确")
	}

	authCtx.BindProject = project
	authCtx.ProjectId = project.ProjectId
	authCtx.ProjectCode = project.Code

	authCtx.BindCluster = cluster
	authCtx.ClusterId = cluster.ClusterId

	return nil
}

// initContextWithDevEnv Dev环境, 可以设置环境变量
func initContextWithIAMProject(c *gin.Context, authCtx *AuthContext) error {
	allow, err := iam.IsAllowedWithResource(c.Request.Context(), authCtx.ProjectId, authCtx.ClusterId, authCtx.Username)
	if err != nil {
		return err
	}
	if !allow {
		return errors.New("没有权限")
	}

	return nil
}

// CredentialRequired
func CredentialRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		authCtx := MustGetAuthContext(c)

		if err := ValidateProjectCluster(c, authCtx); err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Set("auth_context", authCtx)

		if authCtx.BindAPIGW == nil || !authCtx.BindAPIGW.App.Verified {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   "not valid bk apigw request",
				RequestID: authCtx.RequestId,
			})
			return
		}

		if !config.G.ValidateCred(config.CredentialAppCode, authCtx.BindAPIGW.App.AppCode, config.ScopeProjectCode, authCtx.ProjectCode) {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   fmt.Sprintf("app %s have no permission, %s, %s", authCtx.BindAPIGW.App.AppCode, authCtx.BindProject, authCtx.BindCluster),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Next()

	}
}
