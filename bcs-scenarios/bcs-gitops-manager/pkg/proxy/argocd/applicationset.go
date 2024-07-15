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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// ApplicationSetPlugin defines the application http serve plugin
type ApplicationSetPlugin struct {
	*mux.Router
	storage       store.Store
	db            dao.Interface
	permitChecker permitcheck.PermissionInterface
	middleware    mw.MiddlewareInterface
}

// Init will init the path that http need served
func (plugin *ApplicationSetPlugin) Init() error {
	// 自定义接口
	plugin.Path("/generate").Methods(http.MethodPost).
		Handler(plugin.middleware.HttpWrapper(plugin.Generate))
	plugin.Path("/{name}/orphan-delete").Methods(http.MethodDelete).
		Handler(plugin.middleware.HttpWrapper(plugin.OrphanDelete))
	plugin.Path("/{name}/cluster-scope").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.getClusterScope))
	plugin.Path("/{name}/cluster-scope").Methods(http.MethodPut).
		Handler(plugin.middleware.HttpWrapper(plugin.setClusterScope))

	plugin.Path("").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.List))
	plugin.Path("").Methods(http.MethodPost).Handler(plugin.middleware.HttpWrapper(plugin.CreateOrUpdate))
	plugin.Path("/{name}").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.Get))
	plugin.Path("/{name}").Methods(http.MethodDelete).Handler(plugin.middleware.HttpWrapper(plugin.Delete))
	return nil
}

// ApplicationSetGenerateResponse defines the return object of generate applicationset
type ApplicationSetGenerateResponse struct {
	Code      int32                   `json:"code"`
	Message   string                  `json:"message"`
	RequestID string                  `json:"requestID"`
	Data      []*v1alpha1.Application `json:"data"`
}

// Generate the applicationset and return applications it rendered
func (plugin *ApplicationSetPlugin) Generate(r *http.Request) (*http.Request, *mw.HttpResponse) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	appset := &v1alpha1.ApplicationSet{}
	if err = json.Unmarshal(body, appset); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal body failed"))
	}
	apps, statusCode, err := plugin.permitChecker.CheckAppSetCreate(r.Context(), appset)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check create applicationset failed"))
	}
	return r, mw.ReturnJSONResponse(&ApplicationSetGenerateResponse{
		Code:      0,
		RequestID: ctxutils.RequestID(r.Context()),
		Data:      apps,
	})
}

// OrphanDelete delete appset with 'orphan' cascade
func (plugin *ApplicationSetPlugin) OrphanDelete(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appsetName := mux.Vars(r)["name"]
	if appsetName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request applicationset name cannot be empty"))
	}
	argoAppSet, statusCode, err := plugin.permitChecker.CheckAppSetPermission(r.Context(), appsetName,
		permitcheck.AppSetDeleteRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check delete applicationset failed"))
	}
	r = plugin.setApplicationSetAudit(r, argoAppSet.Spec.Template.Spec.Project, appsetName,
		ctxutils.ApplicationSetDelete, ctxutils.EmptyData)
	if err = plugin.storage.DeleteApplicationSetOrphan(r.Context(), appsetName); err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "delete applicationset orphan failed"))
	}
	return r, mw.ReturnJSONResponse("delete applicationset success")
}

// AppSetClusterScopeResponse defines the response of appset's cluster scope
type AppSetClusterScopeResponse struct {
	Code      int32                   `json:"code"`
	Message   string                  `json:"message"`
	RequestID string                  `json:"requestID"`
	Data      *dao.AppSetClusterScope `json:"data"`
}

func (plugin *ApplicationSetPlugin) getClusterScope(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appsetName := mux.Vars(r)["name"]
	if appsetName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request applicationset name cannot be empty"))
	}
	_, statusCode, err := plugin.permitChecker.CheckAppSetPermission(r.Context(), appsetName,
		permitcheck.AppSetViewRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check delete applicationset failed"))
	}
	scope, err := plugin.db.GetAppSetClusterScope(appsetName)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError,
			fmt.Errorf("get cluster scope failed: %s", err.Error()))
	}
	return r, mw.ReturnJSONResponse(&AppSetClusterScopeResponse{
		Code:      0,
		RequestID: ctxutils.RequestID(r.Context()),
		Data:      scope,
	})
}

// AppSetClusterScopeSetRequest defines the request that appset's set cluster scope
type AppSetClusterScopeSetRequest struct {
	Clusters []string `json:"clusters"`
}

func (plugin *ApplicationSetPlugin) setClusterScope(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appsetName := mux.Vars(r)["name"]
	if appsetName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request applicationset name cannot be empty"))
	}
	argoAppSet, statusCode, err := plugin.permitChecker.CheckAppSetPermission(r.Context(), appsetName,
		permitcheck.AppSetUpdateRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check update applicationset failed"))
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	setReq := new(AppSetClusterScopeSetRequest)
	if err = json.Unmarshal(body, setReq); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal request body failed"))
	}
	project := argoAppSet.Spec.Template.Spec.Project
	r = plugin.setApplicationSetAudit(r, project, appsetName, ctxutils.ApplicationSetCreateOrUpdate, string(body))

	clsList, err := plugin.storage.ListCluster(r.Context())
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, errors.Wrapf(err, "list cluster failed"))
	}
	projClusters := make(map[string]string)
	for i := range clsList.Items {
		cls := clsList.Items[i]
		if cls.Project == project {
			projClusters[cls.Name] = cls.Server
		}
	}
	notFoundClusters := make([]string, 0)
	for i := range setReq.Clusters {
		cls := setReq.Clusters[i]
		if _, ok := projClusters[cls]; !ok {
			notFoundClusters = append(notFoundClusters, cls)
		}
	}
	if len(notFoundClusters) != 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("request clusters %v  not found",
			notFoundClusters))
	}
	if err = plugin.db.UpdateAppSetClusterScope(appsetName, strings.Join(setReq.Clusters, ",")); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "update appset's cluster scope failed'"))
	}
	return r, mw.ReturnJSONResponse(&AppSetClusterScopeResponse{
		Code:      0,
		RequestID: ctxutils.RequestID(r.Context()),
	})
}

// CreateOrUpdate create or update application set
func (plugin *ApplicationSetPlugin) CreateOrUpdate(r *http.Request) (*http.Request, *mw.HttpResponse) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	argoAppSet := &v1alpha1.ApplicationSet{}
	if err = json.Unmarshal(body, argoAppSet); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal body failed"))
	}
	r = plugin.setApplicationSetAudit(r, argoAppSet.Spec.Template.Spec.Project, argoAppSet.Name,
		ctxutils.ApplicationSetCreateOrUpdate, string(body))
	_, statusCode, err := plugin.permitChecker.CheckAppSetCreate(r.Context(), argoAppSet)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check create applicationset failed"))
	}
	updatedBody, err := json.Marshal(argoAppSet)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "json marshal applicationset failed"))
	}
	r.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
	length := len(updatedBody)
	r.Header.Set("Content-Length", strconv.Itoa(length))
	r.ContentLength = int64(length)
	return r, mw.ReturnArgoReverse()
}

// Delete delete application set
func (plugin *ApplicationSetPlugin) Delete(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appSetName := mux.Vars(r)["name"]
	if appSetName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request applicationset name cannot be empty"))
	}
	argoAppSet, statusCode, err := plugin.permitChecker.CheckAppSetPermission(r.Context(), appSetName,
		permitcheck.AppSetDeleteRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check delete applicationset failed"))
	}
	r = plugin.setApplicationSetAudit(r, argoAppSet.Spec.Template.Spec.Project, argoAppSet.Name,
		ctxutils.ApplicationSetDelete, ctxutils.EmptyData)
	return r, mw.ReturnArgoReverse()
}

// List all applicationsets
func (plugin *ApplicationSetPlugin) List(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projects := r.URL.Query()["projects"]
	if len(projects) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf("query param 'projects' cannot be empty"))
	}
	for i := range projects {
		projectName := projects[i]
		_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), projectName,
			permitcheck.ProjectViewRSAction)
		if err != nil {
			return r, mw.ReturnErrorResponse(statusCode,
				errors.Wrapf(err, "check project '%s' permission failed", projectName))
		}
	}
	appsetList, statusCode, err := plugin.middleware.ListApplicationSets(r.Context(),
		&applicationset.ApplicationSetListQuery{
			Projects: projects,
			Selector: r.URL.Query().Get("selector")})
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "list applications by project '%v' from "+
			"storage failed", projects))
	}
	return r, mw.ReturnJSONResponse(appsetList)
}

// Get one applicationset
func (plugin *ApplicationSetPlugin) Get(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appsetName := mux.Vars(r)["name"]
	if appsetName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request applicationset name cannot be empty"))
	}
	appSet, statusCode, err := plugin.permitChecker.CheckAppSetPermission(r.Context(), appsetName,
		permitcheck.AppSetViewRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check get applicationset failed"))
	}
	return r, mw.ReturnJSONResponse(appSet)
}

func (plugin *ApplicationSetPlugin) setApplicationSetAudit(r *http.Request, project, appSetName string,
	action ctxutils.AuditAction, data string) *http.Request {
	httpRequest := ctxutils.SetAuditMessage(r, &dao.UserAudit{
		Project:      project,
		Action:       string(action),
		ResourceType: string(ctxutils.AppSetResource),
		ResourceName: appSetName,
		ResourceData: data,
		RequestType:  string(ctxutils.HTTPRequest),
	})
	return httpRequest
}
