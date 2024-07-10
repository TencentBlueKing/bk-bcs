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

package argocd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// PermissionPlugin defines the plugin form permission
type PermissionPlugin struct {
	*mux.Router
	middleware    mw.MiddlewareInterface
	permitChecker permitcheck.PermissionInterface
	storage       store.Store
	db            dao.Interface
}

// Init the permission router
func (plugin *PermissionPlugin) Init() error {
	// 接口供给蓝盾流水线插件"BCS集群执行"使用
	plugin.Path("/check_cluster_scoped").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.checkCluster))

	// 更新权限
	plugin.Path("/{project}").Methods("PUT").
		Handler(plugin.middleware.HttpWrapper(plugin.updatePermissions))
	// 获取资源具备权限的用户
	plugin.Path("/{project}/resource_users").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.queryResourceUsers))
	// 获取登录用户具备权限的资源
	plugin.Path("/{project}/user_permissions").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.userPermissions))

	// 校验用户是否具备 BCS 的资源权限
	plugin.Path("/check_bcs_permissions").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.checkBcsPermissions))
	return nil
}

func (plugin *PermissionPlugin) checkCluster(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := ctxutils.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, fmt.Errorf("user not authorized"))
	}
	userID := r.URL.Query().Get("user")
	if userID == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("query param 'user' cannot be empty"))
	}
	clusterID := r.URL.Query().Get("cluster")
	if clusterID == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("query param 'cluster' cannot be empty"))
	}
	iamAction := r.URL.Query().Get("action")
	if iamAction == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf("query param 'user' cannot be empty"))
	}
	status, err := plugin.permitChecker.CheckBCSClusterPermission(r.Context(), userID, clusterID,
		iam.ActionID(iamAction))
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	return r, mw.ReturnDirectResponse("ok")
}

func (plugin *PermissionPlugin) updatePermissions(r *http.Request) (*http.Request, *mw.HttpResponse) {
	resourceType := r.URL.Query().Get("resourceType")
	if resourceType == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("query param 'resourceType' cannot be empty"))
	}
	bs, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read request body failed"))
	}
	req := new(permitcheck.UpdatePermissionRequest)
	if err = json.Unmarshal(bs, req); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal request failed"))
	}
	if len(req.Users) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf("request 'users' cannot be empty"))
	}
	if len(req.ResourceNames) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request 'resourceNames' cannot be empty"))
	}
	if len(req.ResourceActions) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request 'resourceActions' cannot be empty"))
	}
	project := mux.Vars(r)["project"]
	statusCode, err := plugin.permitChecker.UpdatePermissions(r.Context(), project,
		permitcheck.RSType(resourceType), req)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	return r, mw.ReturnJSONResponse("ok")
}

func (plugin *PermissionPlugin) queryResourceUsers(r *http.Request) (*http.Request, *mw.HttpResponse) {
	project := mux.Vars(r)["project"]
	resourceType := r.URL.Query().Get("resourceType")
	if resourceType == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf(
			"query param 'resourceType' cannot be empty"))
	}
	resourceNames := r.URL.Query()["resourceNames"]
	result, statusCode, err := plugin.permitChecker.QueryResourceUsers(r.Context(),
		project, permitcheck.RSType(resourceType), resourceNames)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	return r, mw.ReturnJSONResponse(result)
}

func (plugin *PermissionPlugin) userPermissions(r *http.Request) (*http.Request, *mw.HttpResponse) {
	project := mux.Vars(r)["project"]
	resourceType := r.URL.Query().Get("resourceType")
	if resourceType == "" {
		permits, statusCode, err := plugin.permitChecker.UserAllPermissions(r.Context(), project)
		if err != nil {
			return r, mw.ReturnErrorResponse(statusCode, err)
		}
		return r, mw.ReturnJSONResponse(permits)
	}

	resourceNames := r.URL.Query()["resourceNames"]
	_, result, statusCode, err := plugin.permitChecker.QueryUserPermissions(r.Context(),
		project, permitcheck.RSType(resourceType), resourceNames)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	return r, mw.ReturnJSONResponse(result)
}

func (plugin *PermissionPlugin) checkBcsPermissions(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := ctxutils.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, fmt.Errorf("user not admin"))
	}
	permit, statusCode, err := plugin.permitChecker.CheckBCSPermissions(r)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	if !permit {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, fmt.Errorf("forbidden"))
	}
	return r, mw.ReturnJSONResponse("ok")
}
