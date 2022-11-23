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
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// RepositoryPlugin for internal project authorization
type RepositoryPlugin struct {
	*mux.Router

	Session    *Session
	Permission *project.BCSProjectPerm
	option     *proxy.GitOpsOptions
}

// Init all project sub path handler
// project plugin is a subRouter, all path registed is relative
func (plugin *RepositoryPlugin) Init() error {
	// init BCSProjectPermitssion
	plugin.Permission = project.NewBCSProjectPermClient(plugin.option.IAMClient)
	plugin.UseEncodedPath()
	// GET /api/v1/repositories?projects={projects}
	plugin.Path("").Methods("GET").Queries("projects", "{projects}").HandlerFunc(plugin.listRepositoryHandler)
	// POST /api/v1/repositories, create new repository
	plugin.Path("").Methods("POST").HandlerFunc(plugin.repositoryCreateHandler)
	// DELETE and Update /api/v1/repositories/{name}
	plugin.Path("/{repo}").Methods("PUT", "DELETE").HandlerFunc(plugin.repositoryEditHandler)
	// GET /api/v1/repositories/{name}
	plugin.Path("/{repo}").Methods("GET").HandlerFunc(plugin.repositoryViewsHandler)

	// GET /api/v1/repositories/{repo}/{details}:
	// apps, helmcharts, refs, validate, appdetails
	plugin.Path("/{repo}/{details}").Methods("GET", "POST").HandlerFunc(plugin.repositoryViewsHandler)

	blog.Infof("argocd repository plugin init successfully")
	return nil
}

// GET /api/v1/repositories?projects={projects}
func (plugin *RepositoryPlugin) listRepositoryHandler(w http.ResponseWriter, r *http.Request) {
	// check header user info
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("request %s get jwt token failure, %s", r.URL.Path, err.Error())
		http.Error(w,
			fmt.Sprintf("Bad Request: %s", err.Error()),
			http.StatusBadRequest,
		)
		return
	}
	project := r.URL.Query().Get("projects")
	// first check relatice project existence
	appProject, err := plugin.option.Storage.GetProject(r.Context(), project)
	if err != nil {
		blog.Errorf("request %s get specified project %s from storage failure, %s", r.URL.Path, project, err.Error())
		http.Error(w, "gitops storage failure", http.StatusInternalServerError)
		return
	}
	if appProject == nil {
		blog.Errorf("RepositoryPlugin Serve %s get no project %s", r.URL.Path, project)
		http.Error(w, "Not Found: Not Found project", http.StatusNotFound)
		return
	}
	projectID := common.GetBCSProjectID(appProject.Annotations)
	if projectID == "" {
		blog.Errorf("request %s failure, relative project %s lost control information", r.URL.Path, project)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// verify project View permistion
	permit, _, err := plugin.Permission.CanViewProject(user.GetUser(), projectID)
	if err != nil {
		blog.Errorf("Repository validate project %s view permission for %s failed, %s", project, r.URL.Path, err.Error())
		http.Error(w, "Authentication system failure", http.StatusInternalServerError)
		return
	}
	blog.Infof("user %s request %s, permission %t", user.GetUser(), r.URL.Path, permit)
	if !permit {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// permission pass, list all repositories in gitops storage
	repositories, err := plugin.option.Storage.ListRepository(r.Context())
	if err != nil {
		blog.Errorf("request %s get all repositories from storage failure, %s", r.URL.Path, err.Error())
		http.Error(w, "gitops storage failure", http.StatusInternalServerError)
		return
	}
	// filter specified project
	items := v1alpha1.Repositories{}
	for _, repo := range repositories.Items {
		if repo.Project == project {
			items = append(items, repo)
		}
	}
	if len(items) == 0 {
		blog.Infof("user %s request %s get no repositories information", user.GetUser(), r.URL.Path)
		proxy.JSONResponse(w, &v1alpha1.RepositoryList{})
		return
	}
	blog.Infof("user %s request %s, %d repositories retrive", user.GetUser(), r.URL.Path, len(items))
	repositories.Items = items
	proxy.JSONResponse(w, repositories)
}

// repository only for local json parse
type repository struct {
	// Repo contains the URL to the remote repository
	Repo string `json:"repo"`
	// Project that Repo belongs to
	Project string `json:"project"`
}

// POST /api/v1/repositories
func (plugin *RepositoryPlugin) repositoryCreateHandler(w http.ResponseWriter, r *http.Request) {
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("RepositoryPlugin get jwt info from request %s failed, %s", r.URL.Path, err.Error())
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}
	localRepo := &repository{}
	body, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(body, localRepo); err != nil {
		blog.Errorf("RepositoryPlugin decode create Repository json failed, %s", err.Error())
		http.Error(w, "Bad Request: error request format", http.StatusBadRequest)
		return
	}
	repo := localRepo.Repo
	proj := localRepo.Project
	blog.Infof("RespositoryPlugin Serve %s %s, read body for repo parse", r.Method, r.URL.Path)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	// check repo info
	if repo == "" || proj == "" {
		blog.Errorf("Respository serve %s failure, no repo or project in request body", r.URL.Path)
		http.Error(w, "Bad Request: no repo or project information", http.StatusBadRequest)
		return
	}
	argoProject, err := plugin.option.Storage.GetProject(r.Context(), proj)
	if err != nil {
		blog.Errorf("RepositoryPlugin Serve %s get project %s info failed, %s", r.URL.Path, proj, err.Error())
		http.Error(w, "gitops project storage failure", http.StatusInternalServerError)
		return
	}
	if argoProject == nil {
		blog.Errorf("RepositoryPlugin Serve %s get no project %s", r.URL.Path, proj)
		http.Error(w, "Not Found: Not Found project", http.StatusNotFound)
		return
	}
	projectID := common.GetBCSProjectID(argoProject.Annotations)
	if projectID == "" {
		blog.Errorf("RepositoryPlugin Serve %s get no bcs project control information", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	permit, _, err := plugin.Permission.CanEditProject(user.GetUser(), projectID)
	if err != nil {
		blog.Errorf("Project %s permission edit %s validate for %s failed, %s",
			proj, r.URL.Path, user.GetUser(), err.Error())
		http.Error(w, "Unauthorized: Auth center failure", http.StatusUnauthorized)
		return
	}
	blog.Infof("user %s request POST %s, permission %s", user.GetUser(), r.URL.Path, permit)
	if !permit {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// backend serving
	plugin.Session.ServeHTTP(w, r)
}

// DELETE and Update /api/v1/repositories/{name}
func (plugin *RepositoryPlugin) repositoryEditHandler(w http.ResponseWriter, r *http.Request) {
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("RepositoryPlugin get jwt info from request %s failed, %s", r.URL.Path, err.Error())
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}
	rawRepo := mux.Vars(r)["repo"]
	repo, err := url.PathUnescape(rawRepo)
	if err != nil {
		blog.Errorf("RepositoryPlugin Serve %s decode repo %s failure, %s", r.URL.Path, rawRepo, err.Error())
		http.Error(w, "Bad Request: malform repository", http.StatusBadRequest)
		return
	}
	repository, err := plugin.option.Storage.GetRepository(r.Context(), repo)
	if err != nil {
		blog.Errorf("RepositoryPlugin Serve %s get Repository %s failure, %s", r.URL.Path, repo, err.Error())
		http.Error(w, "gitops repository storage failure", http.StatusInternalServerError)
		return
	}
	if repository == nil {
		blog.Errorf("RepositoryPlugin Serve %s get no Repository", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	argoProject, err := plugin.option.Storage.GetProject(r.Context(), repository.Project)
	if err != nil {
		blog.Errorf("RepositoryPlugin Serve %s get project %s info failed, %s", r.URL.Path, repository.Project, err.Error())
		http.Error(w, "gitops project storage failure", http.StatusInternalServerError)
		return
	}
	if argoProject == nil {
		blog.Errorf("RepositoryPlugin Serve %s get no project %s", r.URL.Path, repository.Project)
		http.Error(w, "Not Found: Not Found project", http.StatusNotFound)
		return
	}
	projectID := common.GetBCSProjectID(argoProject.Annotations)
	if projectID == "" {
		blog.Errorf("RepositoryPlugin Serve %s get no bcs project control information", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	permit, _, err := plugin.Permission.CanEditProject(user.GetUser(), projectID)
	if err != nil {
		blog.Errorf("Project %s permission validate for %s failed, %s", repository.Project, r.URL.Path, err.Error())
		http.Error(w, "Unauthorized: Auth center failure", http.StatusUnauthorized)
		return
	}
	blog.Infof("user %s request %s, permission %s", user.GetUser(), r.URL.Path, permit)
	if !permit {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// backend serving
	plugin.Session.ServeHTTP(w, r)
}

// handle path permission belows:
// GET /api/v1/repositories/{repo}
// GET /api/v1/repositories/{repo}/apps
// GET /api/v1/repositories/{repo}/helmcharts
// GET /api/v1/repositories/{repo}/refs
// GET /api/v1/repositories/{repo}/appdetails
func (plugin *RepositoryPlugin) repositoryViewsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("RepositoryPlugin get jwt info from request %s failed, %s", r.URL.Path, err.Error())
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}
	rawRepo := mux.Vars(r)["repo"]
	repo, err := url.PathUnescape(rawRepo)
	if err != nil {
		blog.Errorf("RepositoryPlugin Serve %s decode repo %s failure, %s", r.URL.Path, rawRepo, err.Error())
		http.Error(w, "Bad Request: malform repository", http.StatusBadRequest)
		return
	}
	repository, err := plugin.option.Storage.GetRepository(r.Context(), repo)
	if err != nil {
		blog.Errorf("RepositoryPlugin Serve %s get Repository %s failure, %s", r.URL.Path, repo, err.Error())
		http.Error(w, "gitops repository storage failure", http.StatusInternalServerError)
		return
	}
	if repository == nil {
		blog.Errorf("RepositoryPlugin Serve %s get no Repository", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	argoProject, err := plugin.option.Storage.GetProject(r.Context(), repository.Project)
	if err != nil {
		blog.Errorf("RepositoryPlugin Serve %s get project %s info failed, %s", r.URL.Path, repository.Project, err.Error())
		http.Error(w, "gitops project storage failure", http.StatusInternalServerError)
		return
	}
	projectID := common.GetBCSProjectID(argoProject.Annotations)
	if projectID == "" {
		blog.Errorf("RepositoryPlugin Serve %s get no bcs project control information", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	permit, _, err := plugin.Permission.CanViewProject(user.GetUser(), projectID)
	if err != nil {
		blog.Errorf("Project %s permission validate for %s failed, %s", repository.Project, r.URL.Path, err.Error())
		http.Error(w, "Unauthorized: Auth center failure", http.StatusUnauthorized)
		return
	}
	blog.Infof("user %s request %s, permission %t", user.GetUser(), r.URL.Path, permit)
	if !permit {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// authorized then go through reverse proxy
	plugin.Session.ServeHTTP(w, r)
}
