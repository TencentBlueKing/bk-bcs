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

package wrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcsJwt "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

// NewTenantWrapper 租户校验中间件
func NewTenantWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			if !config.G.BCSAPIGW.EnableMultiTenantMode {
				return fn(ctx, req, rsp)
			}

			var (
				user     *bcsJwt.UserClaimsInfo
				username = ctx.Value(ctxkey.UsernameKey).(string)
				tenantID = ctx.Value(ctxkey.TenantIDKey).(string)
			)
			if v, ok := ctx.Value(ctxkey.UserinfoKey).(*bcsJwt.UserClaimsInfo); ok {
				user = v
			}
			// 不是jwt鉴权的跳过租户校验
			if user == nil {
				return fn(ctx, req, rsp)
			}

			// skip method tenant validation
			if SkipMethod(req) {
				return fn(ctx, req, rsp)
			}

			// exempt client
			if SkipTenantValidation(req, username) {
				return fn(ctx, req, rsp)
			}

			// get tenant id
			resourceTenantId, err := GetResourceTenantId(ctx, req)
			if err != nil {
				blog.Errorf("NewTenantWrapper GetResourceTenantId failed, err: %s", err.Error())
				return err
			}

			if tenantID != resourceTenantId {
				return fmt.Errorf("user[%s] tenant[%s] not match resource tenant[%s]",
					user.UserName, tenantID, resourceTenantId)
			}
			return fn(ctx, req, rsp)
		}
	}
}

// TenantClientWhiteList tenant client white list
var TenantClientWhiteList = map[string][]string{}

// SkipMethod skip method tenant validation
func SkipMethod(req server.Request) bool {
	for _, v := range NoInjectProjClusterEndpoints {
		if v == req.Method() {
			return true
		}
	}
	return false
}

// SkipTenantValidation skip tenant validation
func SkipTenantValidation(req server.Request, client string) bool {
	if len(client) == 0 {
		return false
	}
	for _, v := range TenantClientWhiteList[client] {
		if strings.HasPrefix(v, "*") || v == req.Method() {
			return true
		}
	}
	return false
}

// resource id
type resourceID struct {
	ProjectID   string
	ProjectCode string
}

// GetResourceTenantId get resource tenant id
func GetResourceTenantId(ctx context.Context, req server.Request) (string, error) {
	b, err := json.Marshal(req.Body())
	if err != nil {
		return "", err
	}
	resource := &resourceID{}
	if err := json.Unmarshal(b, resource); err != nil {
		return "", err
	}
	return getTenantldByResource(ctx, resource)
}

// getTenantldByResource get tenant id by resource
func getTenantldByResource(ctx context.Context, resource *resourceID) (string, error) {
	projectID := resource.ProjectID
	if len(projectID) == 0 {
		projectID = resource.ProjectCode
	}

	project, err := project.GetProjectInfo(ctx, projectID)
	if err != nil {
		return "", err
	}

	return project.TenantID, nil
}

// GetHeaderTenantIdFromCtx get header tenant id from ctx
func GetHeaderTenantIdFromCtx(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	return md[ctxkey.TenantIdHeaderKey]
}
