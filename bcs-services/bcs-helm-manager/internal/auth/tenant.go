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

package auth

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
)

// CheckUserResourceTenantAttrFunc is the authorization function for go-micro
func CheckUserResourceTenantAttrFunc(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		if !options.GlobalOptions.EnableMultiTenant {
			return fn(ctx, req, rsp)
		}

		var (
			headerTenantId = contextx.GetHeaderTenantIDFromCtx(ctx)
			user           = GetAuthUserFromCtx(ctx)
		)

		// get tenant id
		tenantId := func() string {
			if headerTenantId != "" {
				return headerTenantId
			}
			return user.GetTenantId()
		}()
		ctx = context.WithValue(ctx, contextx.TenantIDContextKey, tenantId)

		// exempt inner user
		if user.IsInner() {
			return fn(ctx, req, rsp)
		}

		// skip method tenant validation
		if SkipHandler(ctx, req) {
			return fn(ctx, req, rsp)
		}

		// exempt client
		if SkipClient(ctx, req, user.GetUsername()) {
			return fn(ctx, req, rsp)
		}

		// 暂不校验
		return fn(ctx, req, rsp)
		// get resource tenant id
		// nolint:govet
		resourceTenantId, err := GetResourceTenantId(ctx, req)
		if err != nil {
			blog.Errorf("CheckUserResourceTenantAttrFunc GetResourceTenantId failed, err: %s", err.Error())
			return err
		}
		blog.Infof("CheckUserResourceTenantAttrFunc headerTenantId[%s] userTenantId[%s] tenantId[%s] resourceTenantId[%s]",
			headerTenantId, user.GetTenantId(), tenantId, resourceTenantId)

		if tenantId != resourceTenantId {
			return fmt.Errorf("user[%s] tenant[%s] not match resource tenant[%s]",
				user.GetUsername(), tenantId, resourceTenantId)
		}

		return fn(ctx, req, rsp)
	}
}

// GetResourceTenantId get resource tenant id
func GetResourceTenantId(ctx context.Context, req server.Request) (string, error) {
	projectCode := contextx.GetProjectCodeFromCtx(ctx)

	pro, err := project.GetProjectByCode(ctx, projectCode)
	if err != nil {
		return "", err
	}

	// 待租户ID支持后修改
	return pro.ProjectCode, nil
}
