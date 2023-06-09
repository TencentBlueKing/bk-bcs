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

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
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

// SessionRequired session 权限校验
func SessionRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		authCtx := MustGetAuthContext(c)

		podCtx, err := validateSession(c, authCtx)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   err.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		authCtx.BindSession = podCtx
		c.Set("auth_context", authCtx)

		c.Next()
	}
}

// 校验用户session权限
func validateSession(c *gin.Context, authCtx *AuthContext) (*types.PodContext, error) {
	sessionId := GetSessionId(c)
	if sessionId == "" {
		return nil, errors.New("session_id is required")
	}

	podCtx, err := sessions.NewStore().OpenAPIScope().Get(c.Request.Context(), sessionId)
	if err != nil {
		return nil, errors.New("session已经过期或不合法")
	}

	if config.G.IsManager(authCtx.Username, podCtx.ClusterId) {
		return podCtx, nil
	}

	if podCtx.HasPerm(authCtx.Username) {
		return podCtx, nil
	}

	return nil, errors.New("用户无权限登入此session")
}

// ValidateProjectCluster xxx
func ValidateProjectCluster(c *gin.Context, authCtx *AuthContext) error {
	projectId := GetProjectIdOrCode(c)
	if projectId == "" {
		return errors.New("project_id or code is required")
	}

	clusterId := GetClusterId(c)
	if clusterId == "" {
		return errors.New("clusterId required")
	}

	project, err := bcs.GetProject(c.Request.Context(), config.G.BCS, projectId)
	if err != nil {
		return errors.Wrap(err, i18n.GetMessage(c, "项目不正确"))
	}

	cluster, err := bcs.GetCluster(c.Request.Context(), project.ProjectId, clusterId)
	if err != nil {
		return errors.Wrap(err, i18n.GetMessage(c, "项目或者集群Id不正确"))
	}

	authCtx.BindProject = project
	authCtx.ProjectId = project.ProjectId
	authCtx.ProjectCode = project.Code

	authCtx.BindCluster = cluster
	authCtx.ClusterId = cluster.ClusterId

	return nil
}

// initContextWithIAMProject Dev环境, 可以设置环境变量
func initContextWithIAMProject(c *gin.Context, authCtx *AuthContext) error {
	namespace := GetNamespace(c)
	allow, err := iam.IsAllowedWithResource(c.Request.Context(), authCtx.ProjectId, authCtx.ClusterId, namespace,
		authCtx.Username)
	if err != nil {
		return err
	}
	if !allow {
		return errors.New(i18n.GetMessage(c, "没有权限"))
	}

	return nil
}

// CredentialRequired xxx
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

		bkAppCode := authCtx.BKAppCode()

		if bkAppCode == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   "not valid bk apigw request",
				RequestID: authCtx.RequestId,
			})
			return
		}

		// 校验项目 or 集群权限
		switch {
		case config.G.ValidateCred(config.CredentialAppCode, bkAppCode, config.ScopeProjectCode, authCtx.ProjectCode):
		case config.G.ValidateCred(config.CredentialAppCode, bkAppCode, config.ScopeClusterId, authCtx.ClusterId): // 校验集群权限
		default:
			c.AbortWithStatusJSON(http.StatusForbidden, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   fmt.Sprintf("app %s have no permission, %s, %s", bkAppCode, authCtx.BindProject, authCtx.BindCluster),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Next()
	}
}
