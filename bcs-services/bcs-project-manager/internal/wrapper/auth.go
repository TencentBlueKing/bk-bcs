/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package wrapper

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/headerkey"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"
)

// NoAuthEndpoints 不需要用户身份认证的方法
var NoAuthEndpoints = []string{
	"Healthz.Ping",
	"Healthz.Healthz",
	"BCSProject.ListAuthorizedProjects",
	"BCSProject.ListProjects",
	"Namespace.ListNamespaces",
}

// NewAuthHeaderAdapter 转换旧的请求头，适配新的鉴权中间件
func NewAuthHeaderAdapter(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
		md, ok := metadata.FromContext(ctx)
		if !ok {
			return errors.New("failed to get micro's metadata")
		}
		if username, ok := md.Get(headerkey.UsernameKey); ok {
			ctx = metadata.Set(ctx, middleauth.CustomUsernameHeaderKey, username)
		}
		return fn(ctx, req, rsp)
	}
}

// NewAuthWrapper return auth middleware
func NewAuthWrapper() *middleauth.GoMicroAuth {
	return middleauth.NewGoMicroAuth(auth.GetJwtClient()).
		EnableSkipHandler(SkipHandler).
		EnableSkipClient(SkipClient).
		SetCheckUserPerm(CheckUserPerm)
}

// SkipHandler implementation for SkipHandler interface
func SkipHandler(ctx context.Context, req server.Request) bool {
	// 禁用身份认证
	if !config.GlobalConf.JWT.Enable {
		return true
	}
	// 特殊指定的Handler，不需要认证的方法
	return stringx.StringInSlice(req.Method(), NoAuthEndpoints)
}

// SkipClient implementation for SkipClient interface
func SkipClient(ctx context.Context, req server.Request, client string) bool {
	if len(client) == 0 {
		return false
	}
	for _, p := range config.GlobalConf.ClientActionExemptPerm.ClientActions {
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

type resourceID struct {
	ProjectID       string `json:"projectID,omitempty"`
	ProjectCode     string `json:"projectCode,omitempty"`
	ProjectIDOrCode string `json:"projectIDOrCode,omitempty"`
	ClusterID       string `json:"clusterID,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
}

func (r *resourceID) check() error {
	if r.ProjectIDOrCode != "" && r.ProjectID == "" {
		if p, err := store.GetModel().GetProject(context.Background(), r.ProjectIDOrCode); err == nil {
			r.ProjectID = p.ProjectID
		}
	}
	if r.ProjectCode != "" && r.ProjectID == "" {
		if p, err := store.GetModel().GetProject(context.Background(), r.ProjectCode); err == nil {
			r.ProjectID = p.ProjectID
		}
	}
	return nil
}

// CheckUserPerm implementation for CheckUserPerm interface
func CheckUserPerm(ctx context.Context, req server.Request, username string) (bool, error) {
	logging.Info("CheckUserPerm: method/%s, username: %s", req.Method(), username)

	if len(username) == 0 {
		return false, errorx.NewReadableErr(errorx.PermDeniedErr, "用户名为空")
	}
	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	resourceID := &resourceID{}
	if uErr := json.Unmarshal(b, resourceID); uErr != nil {
		return false, uErr
	}

	if cErr := resourceID.check(); cErr != nil {
		return false, errorx.NewReadableErr(errorx.ParamErr, "权限校验失败")
	}

	action, ok := auth.ActionPermissions[req.Method()]
	if !ok {
		return false, errorx.NewReadableErr(errorx.PermDeniedErr, "校验用户权限失败")
	}

	allow, _, err := callIAM(username, action, *resourceID)
	if err != nil {
		return false, errorx.NewReadableErr(errorx.PermDeniedErr, "校验用户权限失败")
	}
	return allow, nil
}

func callIAM(username, action string, resourceID resourceID) (bool, string, error) {
	// related actions
	switch action {
	case project.CanViewProjectOperation:
		return auth.ProjectIamClient.CanViewProject(username, resourceID.ProjectID)
	case project.CanCreateProjectOperation:
		return auth.ProjectIamClient.CanCreateProject(username)
	case project.CanEditProjectOperation:
		return auth.ProjectIamClient.CanEditProject(username, resourceID.ProjectID)
	case project.CanDeleteProjectOperation:
		return auth.ProjectIamClient.CanDeleteProject(username, resourceID.ProjectID)
	case namespace.CanViewNamespaceOperation:
		return auth.NamespaceIamClient.CanViewNamespace(username,
			resourceID.ProjectID, resourceID.ClusterID, resourceID.Namespace)
	case namespace.CanListNamespaceOperation:
		return auth.NamespaceIamClient.CanListNamespace(username,
			resourceID.ProjectID, resourceID.ClusterID)
	case namespace.CanCreateNamespaceOperation:
		return auth.NamespaceIamClient.CanCreateNamespace(username, resourceID.ProjectID, resourceID.ClusterID)
	case auth.CanUpdateNamespaceOperation:
		return auth.NamespaceIamClient.CanUpdateNamespace(username,
			resourceID.ProjectID, resourceID.ClusterID, resourceID.Namespace)
	case namespace.CanDeleteNamespaceOperation:
		return auth.NamespaceIamClient.CanDeleteNamespace(username,
			resourceID.ProjectID, resourceID.ClusterID, resourceID.Namespace)
	default:
		return false, "", errorx.NewReadableErr(errorx.PermDeniedErr, "校验用户权限失败")
	}
}
