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

	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/headerkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
)

// NoCheckTenantMethod 不需要校验租户的方法
var NoCheckTenantMethod = []string{
	"BCSProject.CreateProject",
	"BCSProject.ListProjects",
	"BCSProject.ListAuthorizedProjects",
	"BCSProject.ListProjectsForIAM",
	"Healthz.Healthz",
	"Healthz.Ping",
	"Business.ListBusiness",
}

// SkipMethod skip method tenant validation
func SkipMethod(ctx context.Context, req server.Request) bool {
	for _, v := range NoCheckTenantMethod {
		if v == req.Method() {
			return true
		}
	}
	return false
}

// SkipTenantValidation implementation for skip client tenant validation
func SkipTenantValidation(ctx context.Context, req server.Request, client string) bool {
	if len(client) == 0 {
		return false
	}

	for _, p := range config.GlobalConf.ClientActionExemptTenant.ClientActions {
		if client != p.ClientID {
			continue
		}
		if p.All {
			return true
		}
		for _, method := range p.Actions {
			if method == req.Method() {
				return true
			}
		}
	}
	return false
}

// CheckUserResourceTenantAttrFunc xxx
func CheckUserResourceTenantAttrFunc(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		if !tenant.IsMultiTenantEnabled() {
			return fn(ctx, req, rsp)
		}

		var (
			tenantId = ""
			// 用户请求的资源租户ID
			headerTenantId = GetHeaderTenantIdFromCtx(ctx)
			// 认证后的用户信息以及所属租户信息
			user, _ = middleauth.GetUserFromContext(ctx)
		)

		logging.Info("CheckUserResourceTenantAttrFunc clientName[%s] tenant[%s] username[%s], innerClient[%s]",
			user.ClientName, user.TenantId, user.Username, user.InnerClient)

		// exempt inner user
		if user.IsInner() {
			logging.Info("CheckUserResourceTenantAttrFunc user[%s] inner client",
				user.GetUsername())
			return fn(ctx, req, rsp)
		}
		// skip method tenant validation
		if SkipMethod(ctx, req) {
			logging.Info("CheckUserResourceTenantAttrFunc skip method[%s]", req.Method())
			return fn(ctx, req, rsp)
		}
		// exempt client
		if SkipTenantValidation(ctx, req, user.GetUsername()) {
			logging.Info("CheckUserResourceTenantAttrFunc skip tenant[%s] validate", user.GetUsername())
			return fn(ctx, req, rsp)
		}

		// get tenant id
		if headerTenantId == "" {
			tenantId = user.GetTenantId()
		} else {
			if user.GetTenantId() != headerTenantId {
				tenantId = user.GetTenantId()
			} else {
				tenantId = headerTenantId
			}
		}

		// get resource tenant id
		resourceTenantId, err := getResourceTenantId(ctx, req)
		if err != nil {
			logging.Error("CheckUserResourceTenantAttrFunc getResourceTenantId failed[%s], err: %s",
				req.Method(), err.Error())
			return err
		}
		logging.Info("CheckUserResourceTenantAttrFunc headerTenantId[%s] userTenantId[%s] "+
			"tenantId[%s] resourceTenantId[%s]", headerTenantId, user.GetTenantId(), tenantId, resourceTenantId)

		if tenantId != resourceTenantId {
			return fmt.Errorf("user[%s] tenant[%s] not match resource tenant[%s]",
				user, tenantId, resourceTenantId)
		}

		// 注入租户信息
		ctx = context.WithValue(ctx, headerkey.TenantIdKey, tenantId)

		return fn(ctx, req, rsp)
	}
}

// getResourceTenantId get tenant id
func getResourceTenantId(ctx context.Context, req server.Request) (string, error) {
	b, err := json.Marshal(req.Body())
	if err != nil {
		return "", err
	}
	r := &resourceID{}
	if err := json.Unmarshal(b, r); err != nil {
		return "", err
	}

	// 选择第一个非空值作为查询参数
	queryParam := r.ProjectIDOrCode
	if queryParam == "" {
		queryParam = r.ProjectCode
	}
	if queryParam == "" {
		queryParam = r.ProjectID
	}

	if queryParam == "" {
		return "", errorx.NewReadableErr(errorx.TenantResourceCheckErr, "resource is empty")
	}

	p, err := store.GetModel().GetProject(ctx, queryParam)
	if err != nil {
		return "", errorx.NewReadableErr(errorx.TenantResourceCheckErr, "project is not exists")
	}

	r.ProjectID, r.ProjectCode = p.ProjectID, p.ProjectCode
	return p.TenantID, nil
}

// GetHeaderTenantIdFromCtx get tenantID from header
func GetHeaderTenantIdFromCtx(ctx context.Context) string {
	md, _ := metadata.FromContext(ctx)
	tenantId, _ := md.Get(headerkey.TenantIdKey)
	return tenantId
}
