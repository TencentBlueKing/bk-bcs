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
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
)

// RepositoryPlugin for internal project authorization
type RepositoryPlugin struct {
	*mux.Router
	middleware mw.MiddlewareInterface
}

// Init all project sub path handler
// project plugin is a subRouter, all path registered is relative
func (plugin *RepositoryPlugin) Init() error {
	plugin.UseEncodedPath()
	// GET /api/v1/repositories?projects={projects}
	plugin.Path("").Methods("GET").Queries("projects", "{projects}").
		Handler(plugin.middleware.HttpWrapper(plugin.listRepositoryHandler))
	// POST /api/v1/repositories, create new repository
	plugin.Path("").Methods("POST").
		Handler(plugin.middleware.HttpWrapper(plugin.repositoryCreateHandler))
	// DELETE and Update /api/v1/repositories/{name}
	plugin.Path("/{repo}").Methods("PUT", "DELETE").
		Handler(plugin.middleware.HttpWrapper(plugin.repositoryEditHandler))
	// GET /api/v1/repositories/{name}
	plugin.Path("/{repo}").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.repositoryViewsHandler))

	// GET /api/v1/repositories/{repo}/{details}:
	// apps, helmcharts, refs, validate, appdetails
	plugin.Path("/{repo}/{details}").Methods("GET", "POST").
		Handler(plugin.middleware.HttpWrapper(plugin.repositoryViewsHandler))

	blog.Infof("argocd repository plugin init successfully")
	return nil
}

// GET /api/v1/repositories?projects={projects}
func (plugin *RepositoryPlugin) listRepositoryHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projects := r.URL.Query()["projects"]
	if len(projects) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf("query param 'projects' cannot be empty"))
	}
	repositoryList, statusCode, err := plugin.middleware.
		ListRepositories(r.Context(), projects, true)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode,
			errors.Wrapf(err, "list repositories for project '%v' failed", projects))
	}
	return r, mw.ReturnJSONResponse(repositoryList)
}

// repository only for local json parse
// nolint unused
type repository struct {
	// Repo contains the URL to the remote repository
	Repo string `json:"repo"`
	// Project that Repo belongs to
	Project string `json:"project"`
}

// POST /api/v1/repositories
func (plugin *RepositoryPlugin) repositoryCreateHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	localRepo := &v1alpha1.Repository{}
	if err = json.Unmarshal(body, localRepo); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal body failed"))
	}
	if localRepo.Repo == "" || localRepo.Project == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("repo or project param is empty"))
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	_, statusCode, err := plugin.middleware.CheckProjectPermission(r.Context(), localRepo.Project, iam.ProjectEdit)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode,
			errors.Wrapf(err, "check project '%s' edit permission failed", localRepo.Project))
	}
	r = middleware.SetAuditMessage(r, localRepo, middleware.RepositoryCreate)
	return r, mw.ReturnArgoReverse()
}

// DELETE and Update /api/v1/repositories/{name}
func (plugin *RepositoryPlugin) repositoryEditHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	rawRepo := mux.Vars(r)["repo"]
	repo, err := url.PathUnescape(rawRepo)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "parse repo param failed"))
	}

	argoRepo, statusCode, err := plugin.middleware.CheckRepositoryPermission(r.Context(), repo, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode,
			errors.Wrapf(err, "check update repo '%s' permission failed", repo))
	}
	switch r.Method {
	case http.MethodPut:
		r = middleware.SetAuditMessage(r, argoRepo, middleware.RepositoryUpdate)
	case http.MethodDelete:
		r = middleware.SetAuditMessage(r, argoRepo, middleware.RepositoryDelete)
	}
	return r, mw.ReturnArgoReverse()
}

// handle path projectPermission belows:
// GET /api/v1/repositories/{repo}
// GET /api/v1/repositories/{repo}/apps
// GET /api/v1/repositories/{repo}/helmcharts
// GET /api/v1/repositories/{repo}/refs
// GET /api/v1/repositories/{repo}/appdetails
func (plugin *RepositoryPlugin) repositoryViewsHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	rawRepo := mux.Vars(r)["repo"]
	repo, err := url.PathUnescape(rawRepo)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "parse repo param failed"))
	}

	_, statusCode, err := plugin.middleware.CheckRepositoryPermission(r.Context(), repo, iam.ProjectView)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode,
			errors.Wrapf(err, "check view repo '%s' permission failed", repo))
	}
	return r, mw.ReturnArgoReverse()
}
