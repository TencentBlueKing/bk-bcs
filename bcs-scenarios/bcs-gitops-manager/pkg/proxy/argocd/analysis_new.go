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
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// AnalysisNewPlugin for gitops analysis
type AnalysisNewPlugin struct {
	*mux.Router

	middleware    mw.MiddlewareInterface
	permitChecker permitcheck.PermissionInterface
	store         store.Store
}

// Init the http route for analysis
func (plugin *AnalysisNewPlugin) Init() error {
	plugin.PathPrefix("").Handler(plugin.middleware.HttpWrapper(plugin.analysis))
	return nil
}

// analysis return the project analysis
func (plugin *AnalysisNewPlugin) analysis(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := ctxutils.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, errors.Errorf("admin api"))
	}
	return r, mw.ReturnAnalysisReverse()
}
