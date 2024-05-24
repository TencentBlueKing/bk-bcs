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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// ApplicationSetPlugin defines the application http serve plugin
type ApplicationSetPlugin struct {
	*mux.Router
	storage    store.Store
	middleware mw.MiddlewareInterface
}

// Init will init the path that http need served
func (plugin *ApplicationSetPlugin) Init() error {
	// 自定义接口
	plugin.Path("/generate").Methods(http.MethodPost).
		Handler(plugin.middleware.HttpWrapper(plugin.Generate))
	plugin.Path("/{name}/orphan-delete").Methods(http.MethodDelete).
		Handler(plugin.middleware.HttpWrapper(plugin.OrphanDelete))

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
	r = middleware.SetAuditMessage(r, appset, middleware.ApplicationSetGenerate)
	apps, statusCode, err := plugin.middleware.CheckCreateApplicationSet(r.Context(), appset)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check create applicationset failed"))
	}
	return r, mw.ReturnJSONResponse(&ApplicationSetGenerateResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
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
	appset, statusCode, err := plugin.middleware.CheckDeleteApplicationSet(r.Context(), appsetName)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check delete applicationset failed"))
	}
	r = middleware.SetAuditMessage(r, appset, middleware.ApplicationSetDelete)
	if err = plugin.storage.DeleteApplicationSetOrphan(r.Context(), appsetName); err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "delete applicationset orphan failed"))
	}
	return r, mw.ReturnJSONResponse("delete applicationset success")
}

// CreateOrUpdate create or update application set
func (plugin *ApplicationSetPlugin) CreateOrUpdate(r *http.Request) (*http.Request, *mw.HttpResponse) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	appset := &v1alpha1.ApplicationSet{}
	if err = json.Unmarshal(body, appset); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal body failed"))
	}
	r = middleware.SetAuditMessage(r, appset, middleware.ApplicationSetCreateOrUpdate)
	_, statusCode, err := plugin.middleware.CheckCreateApplicationSet(r.Context(), appset)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check create applicationset failed"))
	}
	updatedBody, err := json.Marshal(appset)
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
	appsetName := mux.Vars(r)["name"]
	if appsetName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request applicationset name cannot be empty"))
	}
	appset, statusCode, err := plugin.middleware.CheckDeleteApplicationSet(r.Context(), appsetName)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check delete applicationset failed"))
	}
	r = middleware.SetAuditMessage(r, appset, middleware.ApplicationSetDelete)
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
		_, statusCode, err := plugin.middleware.CheckProjectPermission(r.Context(), projectName, iam.ProjectView)
		if statusCode != http.StatusOK {
			return r, mw.ReturnErrorResponse(statusCode,
				errors.Wrapf(err, "check project '%s' permission failed", projectName))
		}
	}
	appsetList, err := plugin.middleware.ListApplicationSets(r.Context(), &applicationset.ApplicationSetListQuery{
		Projects: projects,
		Selector: r.URL.Query().Get("selector"),
	})
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError,
			errors.Wrapf(err, "list applications by project '%v' from storage failed", projects))
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
	statusCode, err := plugin.middleware.CheckGetApplicationSet(r.Context(), appsetName)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check get applicationset failed"))
	}
	return r, mw.ReturnArgoReverse()
}
