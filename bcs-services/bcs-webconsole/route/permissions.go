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
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/iam"
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

		authCtx, err := GetAuthContext(c)
		if err != nil {
			panic(err)
		}

		// 管理员不校验权限
		if config.G.Base.IsManager(authCtx.Username) {
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

// initContextWithDevEnv Dev环境, 可以设置环境变量
func initContextWithIAMProject(c *gin.Context, authCtx *AuthContext) error {
	allow, err := iam.IsAllowedWithResource(c.Request.Context(), authCtx.ProjectId, authCtx.Username)
	if err != nil {
		return err
	}
	if !allow {
		return errors.New("not allowed")
	}

	return nil
}
