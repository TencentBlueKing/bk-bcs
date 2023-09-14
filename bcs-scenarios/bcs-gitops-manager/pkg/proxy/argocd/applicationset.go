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
 *
 */

package argocd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
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
	plugin.Path("").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.List))
	plugin.Path("").Methods(http.MethodPost).Handler(plugin.middleware.HttpWrapper(plugin.CreateOrUpdate))
	plugin.Path("/{name}").Methods(http.MethodGet).Handler(plugin.middleware.HttpWrapper(plugin.Get))
	plugin.Path("/{name}").Methods(http.MethodDelete).Handler(plugin.middleware.HttpWrapper(plugin.Delete))
	return nil
}

// CreateOrUpdate create or update application set
func (plugin *ApplicationSetPlugin) CreateOrUpdate(ctx context.Context, r *http.Request) *mw.HttpResponse {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	appset := &applicationset.ApplicationSetCreateRequest{}
	if err = json.Unmarshal(body, appset); err != nil {
		return mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal body failed"))
	}
	statusCode, err := plugin.middleware.CheckCreateApplicationSet(ctx, appset.Applicationset)
	if err != nil {
		return mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check create applicationset failed"))
	}
	updatedBody, err := json.Marshal(appset)
	if err != nil {
		return mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "json marshal applicationset failed"))
	}
	r.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
	length := len(updatedBody)
	r.Header.Set("Content-Length", strconv.Itoa(length))
	r.ContentLength = int64(length)
	return mw.ReturnArgoReverse()
}

// Delete delete application set
func (plugin *ApplicationSetPlugin) Delete(ctx context.Context, r *http.Request) *mw.HttpResponse {
	appsetName := mux.Vars(r)["name"]
	if appsetName == "" {
		return mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request applicationset name cannot be empty"))
	}
	statusCode, err := plugin.middleware.CheckDeleteApplicationSet(ctx, appsetName)
	if err != nil {
		return mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check delete applicationset failed"))
	}
	return mw.ReturnArgoReverse()
}

// List all applicationsets
func (plugin *ApplicationSetPlugin) List(ctx context.Context, r *http.Request) *mw.HttpResponse {
	projects := r.URL.Query()["projects"]
	if len(projects) == 0 {
		return mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf("query param 'projects' cannot be empty"))
	}
	for i := range projects {
		projectName := projects[i]
		_, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, projectName, iam.ProjectView)
		if statusCode != http.StatusOK {
			return mw.ReturnErrorResponse(statusCode,
				errors.Wrapf(err, "check project '%s' permission failed", projectName))
		}
	}
	appsetList, err := plugin.middleware.ListApplicationSets(ctx, projects)
	if err != nil {
		return mw.ReturnErrorResponse(http.StatusInternalServerError,
			errors.Wrapf(err, "list applications by project '%v' from storage failed", projects))
	}
	return mw.ReturnJSONResponse(appsetList)
}

// Get one applicationset
func (plugin *ApplicationSetPlugin) Get(ctx context.Context, r *http.Request) *mw.HttpResponse {
	appsetName := mux.Vars(r)["name"]
	if appsetName == "" {
		return mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request applicationset name cannot be empty"))
	}
	statusCode, err := plugin.middleware.CheckGetApplicationSet(ctx, appsetName)
	if err != nil {
		return mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check get applicationset failed"))
	}
	return mw.ReturnArgoReverse()
}
