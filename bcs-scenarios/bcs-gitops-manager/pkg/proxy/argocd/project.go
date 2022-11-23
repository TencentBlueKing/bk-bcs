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

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// ProjectPlugin for internal project authorization
type ProjectPlugin struct {
	*mux.Router

	Session    *Session
	Permission *project.BCSProjectPerm
	option     *proxy.GitOpsOptions
}

// Init all project sub path handler
// project plugin is a subRouter, all path registed is relative
func (plugin *ProjectPlugin) Init() error {
	// init BCSProjectPermitssion
	plugin.Permission = project.NewBCSProjectPermClient(plugin.option.IAMClient)

	// not allow requests
	// POST /api/v1/projects
	plugin.Path("").Methods("POST").HandlerFunc(plugin.forbidden)
	// DELETE and Update /api/v1/projects/{name}
	plugin.HandleFunc("/{name}", plugin.forbidden).Methods("DELETE")
	plugin.HandleFunc("/{name}", plugin.forbidden).Methods("PUT")

	// requests by authorization
	// GET /api/v1/projects
	plugin.Path("").Methods("GET").HandlerFunc(plugin.listProjectsHandler)
	// GET /api/v1/projects/{name}
	plugin.HandleFunc("/{name}", plugin.projectViewsHandler).Methods("GET")
	// GET /api/v1/projects/{name}/{details}:
	// detailed, events, syncwindows, globalprojects all in one handler
	plugin.HandleFunc("/{name}/{details}", plugin.projectViewsHandler).Methods("GET")

	// deny token access
	plugin.PathPrefix("/{name}/roles/").HandlerFunc(plugin.forbidden)

	blog.Infof("argocd project plugin init successfully")
	return nil
}

// forbidden specified path
func (plugin *ProjectPlugin) forbidden(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
}

// GET /api/v1/projects
func (plugin *ProjectPlugin) listProjectsHandler(w http.ResponseWriter, r *http.Request) {
	// check header user info
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("request %s get jwt token failure, %s", r.URL.Path, err.Error())
		http.Error(w,
			fmt.Sprintf("%s: %s", http.StatusText(http.StatusBadRequest), err.Error()),
			http.StatusBadRequest,
		)
		return
	}
	// list all projects in gitops storage
	appProjectList, err := plugin.option.Storage.ListProjects(r.Context())
	if err != nil {
		blog.Errorf("request %s get all projects from storage failure, %s", r.URL.Path, err.Error())
		http.Error(w, "gitops storage failure", http.StatusInternalServerError)
		return
	}
	// validate access permission
	IDList := make([]string, 0)
	controlledProjects := make(map[string]v1alpha1.AppProject)
	for _, app := range appProjectList.Items {
		projectID := common.GetBCSProjectID(app.Annotations)
		if projectID == "" {
			continue
		}
		controlledProjects[projectID] = app
		IDList = append(IDList, projectID)
	}
	if len(IDList) == 0 {
		blog.Infof("user %s request %s get no projects information", user.GetUser(), r.URL.Path)
		proxy.JSONResponse(w, &v1alpha1.AppProjectList{})
		return
	}
	action := string(project.ProjectView)
	result, err := plugin.Permission.GetMultiProjectMultiActionPermission(
		user.GetUser(), IDList, []string{action})
	if err != nil {
		blog.Errorf("request %s authenticate projects %+v failure, %s", r.URL.Path, IDList, err.Error())
		http.Error(w, "authentication failure", http.StatusInternalServerError)
		return
	}
	// permission check
	finnals := make([]v1alpha1.AppProject, 0)
	for projectID, permits := range result {
		if permits[action] {
			appProject := controlledProjects[projectID]
			finnals = append(finnals, appProject)
			blog.Infof("project %s view permission pass", appProject.Name)
		}
	}
	blog.Infof("user %s request %s, %d projects retrive", user.GetUser(), r.URL.Path, len(finnals))
	appProjectList.Items = finnals
	proxy.JSONResponse(w, appProjectList)
}

// handle path permission belows:
// GET /api/v1/projects/{name}
// GET /api/v1/projects/{name}/detailed
// GET /api/v1/projects/{name}/events
// GET /api/v1/projects/{name}/globalprojects
// GET /api/v1/projects/{name}/syncwindows
func (plugin *ProjectPlugin) projectViewsHandler(w http.ResponseWriter, r *http.Request) {
	authorized, err := plugin.checkProjectViewPermission(r)
	if err != nil || !authorized {
		w.WriteHeader(http.StatusUnauthorized)
		message := http.StatusText(http.StatusUnauthorized)
		if err != nil {
			// maybe inner error such as IAM service down
			message = err.Error()
		}
		w.Write([]byte(message)) // nolint
		return
	}
	// authorized then go through reverse proxy
	plugin.Session.ServeHTTP(w, r)
}

// checkProjectViewPermission check single project permission
func (plugin *ProjectPlugin) checkProjectViewPermission(r *http.Request) (bool, error) {
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("ProjectPlugin get jwt info from request %s failed, %s", r.URL.Path, err.Error())
		return false, err
	}
	projectName := mux.Vars(r)["name"]
	argoProject, err := plugin.option.Storage.GetProject(r.Context(), projectName)
	if err != nil {
		blog.Errorf("ProjectPlugin Serve %s get project info failed, %s", r.URL.Path, err.Error())
		return false, err
	}
	if argoProject == nil {
		return false, fmt.Errorf("project Unauthorized")
	}
	projectID := common.GetBCSProjectID(argoProject.Annotations)
	if projectID == "" {
		blog.Errorf("ProjectPlugin Serve %s get no bcs project control information", r.URL.Path)
		return false, nil
	}
	permit, _, err := plugin.Permission.CanViewProject(user.GetUser(), projectID)
	if err != nil {
		blog.Errorf("Project permission validate for %s failed, %s", r.URL.Path, err.Error())
		return false, err
	}
	blog.Infof("user %s request %s, permission %s", user.GetUser(), r.URL.Path, permit)
	return permit, nil
}
