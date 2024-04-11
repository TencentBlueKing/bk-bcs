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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/analyze"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// AnalysisPlugin for gitops analysis
type AnalysisPlugin struct {
	*mux.Router

	analysisClient analyze.AnalysisOverview
	middleware     mw.MiddlewareInterface
	store          store.Store
}

// Init the http route for analysis
func (plugin *AnalysisPlugin) Init() error {
	plugin.Path("").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.analysis))

	plugin.Path("/overview/all").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.overview))
	plugin.Path("/overview/internal").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.overviewInternal))
	plugin.Path("/overview/external").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.overviewExternal))

	plugin.Path("/top_projects").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.topProjects))
	plugin.Path("/managed_resources").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.managedResources))

	plugin.Path("/bkmonitor/common").Methods(http.MethodPost).
		Handler(plugin.middleware.HttpWrapper(plugin.bkmCommon))
	plugin.Path("/bkmonitor/activity_projects/internal").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.bkmActivityProjectsInternal))
	plugin.Path("/bkmonitor/activity_projects/external").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.bkmActivityProjectsExternal))

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

func (plugin *AnalysisPlugin) analysis(r *http.Request) (*http.Request, *mw.HttpResponse) {
	isAll, projects, err := plugin.checkQuery(r)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, err)
	}
	if isAll {
		return r, mw.ReturnJSONResponse(&AnalysisResponse{
			Code:      0,
			RequestID: mw.RequestID(r.Context()),
			Data:      plugin.analysisClient.AnalysisProjectAll(),
		})
	}
	result, err := plugin.analysisClient.AnalysisProject(r.Context(), projects)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse(&AnalysisResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      result,
	})
}

// checkQuery it will check the permission with projects
func (plugin *AnalysisPlugin) checkQuery(r *http.Request) (bool, []v1alpha1.AppProject, error) {
	projects := r.URL.Query()["projects"]
	if len(projects) != 0 {
		result := make([]v1alpha1.AppProject, 0, len(projects))
		for i := range projects {
			projectName := projects[i]
			argoProj, statusCode, err := plugin.middleware.CheckProjectPermission(r.Context(),
				projectName, iam.ProjectView)
			if statusCode != http.StatusOK {
				return false, nil, errors.Wrapf(err, "check project '%s' permission failed", projectName)
			}
			result = append(result, *argoProj)
		}
		return false, result, nil
	}

	user := mw.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return false, nil, fmt.Errorf("query param 'projects' cannot be empty")
	}
	projList, err := plugin.store.ListProjectsWithoutAuth(r.Context())
	if err != nil {
		return false, nil, err
	}
	return true, projList.Items, nil
}

func (plugin *AnalysisPlugin) topProjects(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := mw.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, errors.Errorf("admin api"))
	}
	return r, mw.ReturnJSONResponse(&AnalysisResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      plugin.analysisClient.TopProjects(),
	})
}

func (plugin *AnalysisPlugin) managedResources(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := mw.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, errors.Errorf("admin api"))
	}
	return r, mw.ReturnJSONResponse(&AnalysisResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      plugin.analysisClient.ResourceInfos(),
	})
}

func (plugin *AnalysisPlugin) overview(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := mw.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, errors.Errorf("admin api"))
	}
	internal := plugin.analysisClient.OverviewAllInternal()
	internal.Type = "国内"
	external := plugin.analysisClient.OverviewAllExternal()
	external.Type = "海外"
	return r, mw.ReturnJSONResponse(&AnalysisResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      []*analyze.AnalysisOverviewAll{internal, external},
	})
}

func (plugin *AnalysisPlugin) overviewInternal(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := mw.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, errors.Errorf("admin api"))
	}
	return r, mw.ReturnJSONResponse(&AnalysisResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      []*analyze.AnalysisOverviewAll{plugin.analysisClient.OverviewAllInternal()},
	})
}

func (plugin *AnalysisPlugin) overviewExternal(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := mw.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, errors.Errorf("admin api"))
	}
	return r, mw.ReturnJSONResponse(&AnalysisResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      []*analyze.AnalysisOverviewAll{plugin.analysisClient.OverviewAllExternal()},
	})
}

func (plugin *AnalysisPlugin) bkmCommon(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := mw.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, errors.Errorf("admin api"))
	}
	bkmMessage, err := plugin.buildBKMRequest(r)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, err)
	}
	series, err := plugin.analysisClient.BKMonitorCommonGet(bkmMessage)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse(&AnalysisResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      series,
	})
}

func (plugin *AnalysisPlugin) bkmActivityProjectsInternal(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := mw.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, errors.Errorf("admin api"))
	}
	projects, err := plugin.analysisClient.BKMonitorTopActivityProjectsInternal()
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse(&AnalysisResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      projects,
	})
}

func (plugin *AnalysisPlugin) bkmActivityProjectsExternal(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := mw.User(r.Context())
	if !common.IsAdminUser(user.ClientID) {
		return r, mw.ReturnErrorResponse(http.StatusForbidden, errors.Errorf("admin api"))
	}
	projects, err := plugin.analysisClient.BKMonitorTopActivityProjectsExternal()
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse(&AnalysisResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      projects,
	})
}

func (plugin *AnalysisPlugin) buildBKMRequest(r *http.Request) (*analyze.BKMonitorGetMessage, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read body failed")
	}
	req := new(analyze.BKMonitorGetMessage)
	if err = json.Unmarshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "unmarshal request body failed")
	}
	return req, nil
}
