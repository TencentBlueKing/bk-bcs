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

	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// * addinational stream path, application wrapper
// GET：/api/v1/stream/applications?projects={projects}，获取event事件流，强制启用projects
// GET：/api/v1/stream/applications/{name}/resource-tree，指定资源树事件流

// StreamPlugin for internal streaming
type StreamPlugin struct {
	*mux.Router

	appHandler *AppPlugin
	Session    *Session
	Permission *project.BCSProjectPerm
	option     *proxy.GitOpsOptions
}

// Init all project sub path handler
// project plugin is a subRouter, all path registed is relative
func (plugin *StreamPlugin) Init() error {
	// init BCSProjectPermitssion
	plugin.Permission = project.NewBCSProjectPermClient(plugin.option.IAMClient)

	// done(DeveloperJim): GET /api/v1/stream/applications?projects={projects}
	plugin.Path("").Methods("GET").Queries("projects", "{projects}").HandlerFunc(plugin.projectViewHandler)

	// done(DeveloperJim): GET /api/v1/stream/applications/{name}/resource-tree
	plugin.Path("/{name}/resource-tree").Methods("GET").HandlerFunc(plugin.appHandler.applicationViewsHandler)
	blog.Infof("argocd stream applications plugin init successfully")
	return nil
}

func (plugin *StreamPlugin) projectViewHandler(w http.ResponseWriter, r *http.Request) {
	// check header user info
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("request %s get jwt token failure, %s", r.URL.Path, err.Error())
		http.Error(w,
			fmt.Sprintf("Bad Request: %s", err.Error()), http.StatusBadRequest,
		)
		return
	}
	project := r.URL.Query().Get("projects")
	// get project info and validate permission
	appProject, err := plugin.option.Storage.GetProject(r.Context(), project)
	if err != nil {
		blog.Errorf("request %s get project %s from storage failure, %s", r.URL.Path, project, err.Error())
		http.Error(w, "gitops project storage failure", http.StatusInternalServerError)
		return
	}
	if appProject == nil {
		blog.Errorf("StreamPlugin Serve %s get no project %s", r.URL.Path, project)
		http.Error(w, "Not Found: Found no project", http.StatusNotFound)
		return
	}
	projectID := common.GetBCSProjectID(appProject.Annotations)
	if projectID == "" {
		blog.Errorf("StreamPlugin Serve %s, get project %s ID failure, not under control", r.URL.Path, project)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	permit, _, err := plugin.Permission.CanViewProject(user.GetUser(), projectID)
	if err != nil {
		blog.Infof("StreamPlugin Serve %s user %s project %s failure, %s",
			r.URL.Path, user.GetUser(), project, err.Error())
		http.Error(w, "Unauthorized: Auth center failure", http.StatusInternalServerError)
		return
	}
	blog.Infof("StreamPlugin Serve [%s] %s user %s project %s, permission: %t",
		r.Method, r.URL.Path, user.GetUser(), project, permit)
	if !permit {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// backend serving
	plugin.Session.ServeHTTP(w, r)
}
