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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// AnalysisPlugin for gitops analysis
type AnalysisPlugin struct {
	*mux.Router

	middleware    mw.MiddlewareInterface
	permitChecker permitcheck.PermissionInterface
	store         store.Store
}

// Init the http route for analysis
func (plugin *AnalysisPlugin) Init() error {
	plugin.Path("").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.analysis))
	plugin.Path("/overview").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.common))
	plugin.Path("/overview/compare").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.common))
	plugin.Path("/top_projects").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.common))
	plugin.Path("/managed_resources").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.common))

	plugin.Path("/bkmonitor/activity_projects").Methods(http.MethodGet).Handler(
		plugin.middleware.HttpWrapper(plugin.common))
	plugin.Path("/bkmonitor/common").Methods(http.MethodPost).Handler(plugin.middleware.HttpWrapper(plugin.common))
	plugin.Path("/bkmonitor/slo").Methods(http.MethodPost).Handler(plugin.middleware.HttpWrapper(plugin.common))
	plugin.Path("/bkmonitor/slo_unavailable").Methods(http.MethodPost).Handler(plugin.middleware.
		HttpWrapper(plugin.common))

	plugin.Path("/query/projects").Methods(http.MethodGet).Handler(plugin.middleware.
		HttpWrapper(plugin.common))
	plugin.Path("/query/applications").Methods(http.MethodGet).Handler(plugin.middleware.
		HttpWrapper(plugin.common))

	blog.Infof("argocd analysis init successfully")
	return nil
}

// AnalysisResponse defines the response of analysis data
type AnalysisResponse struct {
	Code      int32       `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"requestID"`
	Data      interface{} `json:"data"`
}

// analysis return the project analysis
func (plugin *AnalysisPlugin) analysis(r *http.Request) (*http.Request, *mw.HttpResponse) {
	_, _, err := plugin.checkQuery(r)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, err)
	}
	return r, mw.ReturnAnalysisReverse()
}

// checkQuery it will check the permission with projects
func (plugin *AnalysisPlugin) checkQuery(r *http.Request) (bool, []v1alpha1.AppProject, error) {
	projects := r.URL.Query()["projects"]
	if len(projects) != 0 {
		result := make([]v1alpha1.AppProject, 0, len(projects))
		for i := range projects {
			projectName := projects[i]
			argoProj, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), projectName,
				permitcheck.ProjectViewRSAction)
			if statusCode != http.StatusOK {
				return false, nil, errors.Wrapf(err, "check project '%s' permission failed", projectName)
			}
			result = append(result, *argoProj)
		}
		return false, result, nil
	}

	user := ctxutils.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return false, nil, fmt.Errorf("query param 'projects' cannot be empty")
	}
	projList, err := plugin.store.ListProjectsWithoutAuth(r.Context())
	if err != nil {
		return false, nil, err
	}
	return true, projList.Items, nil
}

// common check admin user
func (plugin *AnalysisPlugin) common(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := ctxutils.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, errors.Errorf("admin api"))
	}
	return r, mw.ReturnAnalysisReverse()
}
