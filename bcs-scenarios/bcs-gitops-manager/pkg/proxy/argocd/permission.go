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
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
)

// PermissionPlugin defines the plugin form permission
type PermissionPlugin struct {
	*mux.Router
	middleware mw.MiddlewareInterface
}

// Init init the permission router
func (plugin *PermissionPlugin) Init() error {
	plugin.Path("/check_cluster_scoped").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.checkCluster))
	return nil
}

func (plugin *PermissionPlugin) checkCluster(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := mw.User(r.Context())
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
	status, err := plugin.middleware.CheckClusterScopedPermission(r.Context(), userID,
		clusterID, iam.ActionID(iamAction))
	if err != nil {
		return r, mw.ReturnErrorResponse(status, err)
	}
	return r, mw.ReturnDirectResponse("ok")
}
