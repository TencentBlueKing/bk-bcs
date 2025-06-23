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

// Package route xxx
package route

import (
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

const (
	// HeaderTenantId header tenant id
	HeaderTenantId = "X-Bk-Tenant-Id"
	// DefaultTenantId default tenant id
	DefaultTenantId = "default"
)

// TenantCheck 租户校验中间件
func TenantCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.G.Base.EnableTenant {
			c.Next()
			return
		}

		var (
			headerTenantId = c.GetHeader(HeaderTenantId)
			authCtx        = MustGetAuthContext(c)
			user           = MustGetAuthContext(c).BindBCS
		)
		// 不是jwt鉴权的跳过租户校验
		if user == nil {
			c.Next()
			return
		}

		// skip method tenant validation
		if SkipMethod(c) {
			c.Next()
			return
		}

		// get tenant id
		tenantId := func() string {
			if headerTenantId != "" {
				return headerTenantId
			}
			if user.TenantId != "" {
				return user.TenantId
			}
			return DefaultTenantId
		}()

		// get tenant id
		resourceTenantId, err := GetResourceTenantId(c)
		if err != nil {
			blog.Errorf("TenantHandler getTenantld failed, err: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   i18n.T(c, "%s", err),
				RequestID: authCtx.RequestId,
			})
			return
		}
		blog.Infof("TenantHandler headerTenantId[%s] userTenantId[%s] tenantId[%s] resourceTenantId[%s]",
			headerTenantId, user.TenantId, tenantId, resourceTenantId)

		if tenantId != resourceTenantId {
			err := fmt.Errorf("user[%s] tenant[%s] not match resource tenant[%s]",
				user.UserName, tenantId, resourceTenantId)
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   i18n.T(c, "%s", err),
				RequestID: authCtx.RequestId,
			})
			return
		}
		c.Next()
	}
}

// NoCheckTenantMethod no check tenant method
var NoCheckTenantMethod = []string{
	"/api/command/delay",
	"/api/command/delay/:username",
	"/api/command/delay/:username/meter",
}

// SkipMethod skip method tenant validation
func SkipMethod(c *gin.Context) bool {
	for _, v := range NoCheckTenantMethod {
		if v == c.FullPath() {
			return true
		}
	}
	return false
}

// GetResourceTenantId get resource tenant id
func GetResourceTenantId(c *gin.Context) (string, error) {

	projectId := GetProjectIdOrCode(c)
	if projectId == "" {
		// param url中是否有project_id
		projectId = c.Query("project_id")
		if projectId == "" {
			return "", errors.New("project_id or code is required")
		}
	}

	project, err := bcs.GetProject(c.Request.Context(), config.G.BCS, projectId)
	if err != nil {
		return "", err
	}

	return project.TenantID, nil
}
