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
	"strconv"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// AppPlugin for internal project authorization
type AppPlugin struct {
	*mux.Router

	Session    *Session
	Permission *project.BCSProjectPerm
	option     *proxy.GitOpsOptions
}

// all argocd application URL:
// * required Project Edit permission
// POST：  /api/v1/applications，创建
// DELETE：/api/v1/applications/{name}，指定删除
// PUT：   /api/v1/applications/{name}，指定更新
// PATCH： /api/v1/applications/{name}，指定字段更新
// DELETE：/api/v1/applications/{name}/operation，终止当前操作
// POST：  /api/v1/applications/{name}/resource，patch资源
// DELETE：/api/v1/applications/{name}/resource，删除资源
// POST：  /api/v1/applications/{name}/resource/actions，run resource action
// POST：  /api/v1/applications/{name}/rollback，回退到上个版本
// PUT：   /api/v1/applications/{name}/spec，更新spec
// POST：  /api/v1/applications/{name}/sync，发起app同步
//
// * required Project View permission
// GET：/api/v1/applications?projects={projects}，获取列表，强制启用projects参数
//
// path prefix format: /api/v1/applications/{name}
// GET：/api/v1/applications/{name}，获取具体信息
// GET：/api/v1/applications/{name}/managed-resources，返回管理资源
// GET：/api/v1/applications/{name}/resource-tree，返回资源树
// GET：/api/v1/applications/{name}/events
// GET：/api/v1/applications/{name}/logs，日志，建议直接访问集群接口
// GET：/api/v1/applications/{name}/manifests
// GET：/api/v1/applications/{name}/pods/{podName}/logs，获取Pod日志，建议直接访问集群接口
// GET：/api/v1/applications/{name}/resource，获取资源
// GET：/api/v1/applications/{name}/resource/actions，获取actions
// GET：/api/v1/applications/{name}/revisions/{revision}/metadata，获取指定版本的meta数据
// GET：/api/v1/applications/{name}/syncwindws，获取syncwindows
//

// Init all project sub path handler
// project plugin is a subRouter, all path registed is relative
func (plugin *AppPlugin) Init() error {
	// init BCSProjectPermitssion
	plugin.Permission = project.NewBCSProjectPermClient(plugin.option.IAMClient)

	// POST /api/v1/applications, create new application
	plugin.Path("").Methods("POST").HandlerFunc(plugin.createApplicationHandler)
	// force check query, GET /api/v1/applications?projects={projects}
	plugin.Path("").Methods("GET").Queries("projects", "{projects}").HandlerFunc(plugin.listApplicationsHandler)

	// Put,Patch,Delete with preifx /api/v1/applications/{name}
	plugin.PathPrefix("/{name}").Methods("PUT", "POST", "DELETE", "PATCH").HandlerFunc(plugin.applicationEditHandler)

	// GET with prefix /api/v1/applications/{name}
	plugin.PathPrefix("/{name}").Methods("GET").HandlerFunc(plugin.applicationViewsHandler)

	// todo(DeveloperJim): GET /api/v1/stream/applications?project={project}
	// todo(DeveloperJim): GET /api/v1/stream/applications/{name}/resource-tree
	blog.Infof("argocd application plugin init successfully")
	return nil
}

// POST /api/v1/applications, create new application
// validate project detail from request
func (plugin *AppPlugin) createApplicationHandler(w http.ResponseWriter, r *http.Request) {
	// check header user info
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("request %s get jwt token failure, %s", r.URL.Path, err.Error())
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}
	// decode requst for project information
	body, _ := ioutil.ReadAll(r.Body)
	app := &v1alpha1.Application{}
	if err := json.Unmarshal(body, app); err != nil {
		blog.Infof("AppPlugin decode request %s create application body failure, %s", r.URL.Path, err.Error())
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}
	// all validation workflow
	if app.Spec.Project == "" || app.Spec.Project == "default" {
		blog.Errorf("AppPlugin Serve %s for user %s, create application failure, project lost",
			r.URL.Path, user.GetUser())
		http.Error(w, "Bad Request: project lost", http.StatusBadRequest)
		return
	}
	project, err := plugin.option.Storage.GetProject(r.Context(), app.Spec.Project)
	if err != nil {
		blog.Errorf("AppPlugin Serve %s for user %s failed at get project %s details, %s",
			r.URL.Path, user.GetUser(), app.Spec.Project, err.Error())
		http.Error(w, "gitops project storage failure", http.StatusInternalServerError)
		return
	}
	if project == nil {
		blog.Errorf("AppPlugin Serve %s get no project %s", r.URL.Path, app.Spec.Project)
		http.Error(w, "Not Found: Not Found project", http.StatusNotFound)
		return
	}
	projectID := common.GetBCSProjectID(project.Annotations)
	if projectID == "" {
		blog.Errorf("AppPlugin Serve %s for user %s Unauthorized, project %s lost control info",
			r.URL.Path, user.GetUser(), project.Name)
		http.Error(w, "Unauthorized: project Unauthorized", http.StatusUnauthorized)
		return
	}
	permit, _, err := plugin.Permission.CanEditProject(user.GetUser(), projectID)
	if err != nil {
		blog.Errorf("AppPlugin Serve %s for user %s at project %s failed, AuthCenter err %s",
			r.URL.Path, user.GetUser(), project.Name, err.Error())
		http.Error(w, "AuthCenter request failure", http.StatusInternalServerError)
		return
	}
	blog.Infof("Applugin Serve [%s] %s user %s project %s, permission: %t",
		r.Method, r.URL.Path, user.GetUser(), project.Name, permit)
	if !permit {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// setting control annotations
	if app.Annotations == nil {
		app.Annotations = make(map[string]string)
	}
	app.Annotations[common.ProjectIDKey] = projectID
	app.Annotations[common.ProjectBusinessIDKey] = project.Annotations[common.ProjectBusinessIDKey]
	// recover request body
	updatedBody, _ := json.Marshal(app)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(updatedBody))
	length := len(updatedBody)
	r.Header.Set("Content-Length", strconv.Itoa(length))
	r.ContentLength = int64(length)
	plugin.Session.ServeHTTP(w, r)
}

// GET /api/v1/applications?projects={projects}
func (plugin *AppPlugin) listApplicationsHandler(w http.ResponseWriter, r *http.Request) {
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
		blog.Errorf("AppPlugin Serve %s get no project %s", r.URL.Path, project)
		http.Error(w, "Not Found: Not Found project", http.StatusNotFound)
		return
	}
	projectID := common.GetBCSProjectID(appProject.Annotations)
	if projectID == "" {
		blog.Errorf("AppPlugin Serve %s, get project %s ID failure, not under control", r.URL.Path, project)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	permit, _, err := plugin.Permission.CanViewProject(user.GetUser(), projectID)
	if err != nil {
		blog.Infof("AppPlugin Serve %s user %s project %s failure, %s",
			r.URL.Path, user.GetUser(), project, err.Error())
		http.Error(w, "Unauthorized: Auth center failure", http.StatusInternalServerError)
		return
	}
	blog.Infof("AppPlugin Serve [%s] %s user %s project %s, permission: %t",
		r.Method, r.URL.Path, user.GetUser(), project, permit)
	if !permit {
		http.Error(w, "Unauthorized", http.StatusInternalServerError)
		return
	}
	appList, err := plugin.option.Storage.ListApplications(r.Context(), &store.ListAppOptions{Project: project})
	if err != nil {
		blog.Errorf("request %s user %s get all applications from storage failure, %s",
			r.URL.Path, user.GetUser(), err.Error())
		http.Error(w, "gitops storage failure", http.StatusInternalServerError)
		return
	}
	blog.Infof("user %s request %s, %d applications retrive", user.GetUser(), r.URL.Path, len(appList.Items))
	proxy.JSONResponse(w, appList)
}

// Put,Patch,Delete with preifx /api/v1/applications/{name}
func (plugin *AppPlugin) applicationEditHandler(w http.ResponseWriter, r *http.Request) {
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("AppPlugin get jwt info from request %s failed, %s", r.URL.Path, err.Error())
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}
	name := mux.Vars(r)["name"]
	app, err := plugin.option.Storage.GetApplication(r.Context(), name)
	if err != nil {
		blog.Errorf("AppPlugin Serve %s for user %s get Application %s failure, %s",
			r.URL.Path, user.GetUser(), name, err.Error())
		http.Error(w, "gitops application storage failure", http.StatusInternalServerError)
		return
	}
	if app == nil {
		blog.Errorf("AppPlugin Serve %s get no application %s", r.URL.Path, name)
		http.Error(w, "gitops resource Not Found", http.StatusNotFound)
		return
	}
	projectID := common.GetBCSProjectID(app.Annotations)
	if projectID == "" {
		blog.Errorf("AppPlugin Serve %s for user %s, get no bcs project control information",
			r.URL.Path, user.GetUser())
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	permit, _, err := plugin.Permission.CanEditProject(user.GetUser(), projectID)
	if err != nil {
		blog.Errorf("AppPlugin Serve %s for user %s at project %s failure, %s",
			r.URL.Path, user.GetUser(), app.Spec.Project, err.Error())
		http.Error(w, "Unauthorized: Auth center failure", http.StatusInternalServerError)
		return
	}
	blog.Infof("AppPlugin Serve [%s] %s user %s project %s, permission %t",
		r.Method, r.URL.Path, user.GetUser(), app.Spec.Project, permit)
	if !permit {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// backend serving
	plugin.Session.ServeHTTP(w, r)
}

// GET with prefix /api/v1/applications/{name}
func (plugin *AppPlugin) applicationViewsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("AppPlugin get jwt info from request %s failed, %s", r.URL.Path, err.Error())
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}
	name := mux.Vars(r)["name"]
	app, err := plugin.option.Storage.GetApplication(r.Context(), name)
	if err != nil {
		blog.Errorf("AppPlugin Serve %s for user %s get Application %s failure, %s",
			r.URL.Path, user.GetUser(), name, err.Error())
		http.Error(w, "gitops application storage failure", http.StatusInternalServerError)
		return
	}
	if app == nil {
		blog.Errorf("AppPlugin Serve %s get no application %s", r.URL.Path, name)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	projectID := common.GetBCSProjectID(app.Annotations)
	if projectID == "" {
		blog.Errorf("AppPlugin Serve %s for user %s, get no bcs project control information",
			r.URL.Path, user.GetUser())
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	permit, _, err := plugin.Permission.CanViewProject(user.GetUser(), projectID)
	if err != nil {
		blog.Errorf("AppPlugin Serve %s for user %s at project %s failure, %s",
			r.URL.Path, user.GetUser(), app.Spec.Project, err.Error())
		http.Error(w, "Unauthorized: Auth center failure", http.StatusInternalServerError)
		return
	}
	blog.Infof("AppPlugin Serve [%s] %s user %s project %s, permission %t",
		r.Method, r.URL.Path, user.GetUser(), app.Spec.Project, permit)
	if !permit {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	// authorized then go through reverse proxy
	plugin.Session.ServeHTTP(w, r)
}
