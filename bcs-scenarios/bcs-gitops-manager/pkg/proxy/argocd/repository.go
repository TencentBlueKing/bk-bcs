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
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
)

// RepositoryPlugin for internal project authorization
type RepositoryPlugin struct {
	*mux.Router
	permitChecker permitcheck.PermissionInterface
	middleware    mw.MiddlewareInterface
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
	plugin.Path("/{repo}").Methods("DELETE").Handler(plugin.middleware.HttpWrapper(plugin.repoDeleteHandler))
	plugin.Path("/{repo}").Methods("PUT").Handler(plugin.middleware.HttpWrapper(plugin.repoUpdateHandler))
	// GET /api/v1/repositories/{repo}/{details}:
	// apps, helmcharts, refs, validate, appdetails
	plugin.Path("/{repo}/{details}").Methods("GET", "POST").
		Handler(plugin.middleware.HttpWrapper(plugin.repoHandler))

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
	statusCode, err := plugin.permitChecker.CheckRepoCreate(r.Context(), localRepo)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode,
			errors.Wrapf(err, "check project '%s' edit permission failed", localRepo.Project))
	}
	r = plugin.setRepoAudit(r, localRepo.Project, localRepo.Repo, ctxutils.RepoCreate, ctxutils.EmptyData)
	return r, mw.ReturnArgoReverse()
}

func (plugin *RepositoryPlugin) repoDeleteHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	rawRepo := mux.Vars(r)["repo"]
	project := r.URL.Query().Get("appProject")
	repo, err := url.PathUnescape(rawRepo)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "parse repo param failed"))
	}
	repo, err = url.QueryUnescape(repo)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err,
			"unescape repo '%s' failed", repo))
	}

	argoRepo, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), project, repo,
		permitcheck.RepoDeleteRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check delete repo '%s' permission failed",
			repo))
	}
	r = plugin.setRepoAudit(r, argoRepo.Project, repo, ctxutils.RepoDelete, ctxutils.EmptyData)
	return r, mw.ReturnArgoReverse()
}

func (plugin *RepositoryPlugin) repoUpdateHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	rawRepo := mux.Vars(r)["repo"]
	project := r.URL.Query().Get("appProject")
	repo, err := url.PathUnescape(rawRepo)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "parse repo param failed"))
	}
	repo, err = url.QueryUnescape(repo)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err,
			"unescape repo '%s' failed", repo))
	}

	argoRepo, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), project, repo,
		permitcheck.RepoUpdateRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check update repo '%s' permission failed",
			repo))
	}
	r = plugin.setRepoAudit(r, argoRepo.Project, repo, ctxutils.RepoUpdate, ctxutils.EmptyData)
	return r, mw.ReturnArgoReverse()
}

func (plugin *RepositoryPlugin) repoHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	rawRepo := mux.Vars(r)["repo"]
	project := r.URL.Query().Get("appProject")
	repo, err := url.PathUnescape(rawRepo)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "parse repo param failed"))
	}
	repo, err = url.QueryUnescape(repo)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err,
			"unescape repo '%s' failed", repo))
	}

	var action permitcheck.RSAction
	switch r.Method {
	case http.MethodGet:
		action = permitcheck.RepoViewRSAction
	case http.MethodPost:
		action = permitcheck.RepoUpdateRSAction
	}
	_, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), project, repo, action)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check get repo '%s' permission failed",
			repo))
	}
	return r, mw.ReturnArgoReverse()
}

func (plugin *RepositoryPlugin) setRepoAudit(r *http.Request, project, repo string,
	action ctxutils.AuditAction, data string) *http.Request {
	return ctxutils.SetAuditMessage(r, &dao.UserAudit{
		Project:      project,
		Action:       string(action),
		ResourceType: string(ctxutils.RepoResource),
		ResourceName: repo,
		ResourceData: data,
		RequestType:  string(ctxutils.HTTPRequest),
	})
}
