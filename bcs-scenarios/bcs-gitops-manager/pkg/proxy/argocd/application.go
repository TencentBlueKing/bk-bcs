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

	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils/jsonq"
)

// AppPlugin for internal project authorization
type AppPlugin struct {
	*mux.Router
	storage    store.Store
	middleware mw.MiddlewareInterface
}

// all argocd application URL:
// * required Project Edit projectPermission
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
// * required Project View projectPermission
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
// project plugin is a subRouter, all path registered is relative
func (plugin *AppPlugin) Init() error {
	// POST /api/v1/applications, create new application
	plugin.Path("").Methods("POST").
		Handler(plugin.middleware.HttpWrapper(plugin.createApplicationHandler))
	// force check query, GET /api/v1/applications?projects={projects}
	plugin.Path("").Methods("GET").Queries("projects", "{projects}").
		Handler(plugin.middleware.HttpWrapper(plugin.listApplicationsHandler))

	plugin.Path("/dry-run").Methods("POST").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationDryRun))
	plugin.Path("/diff").Methods("POST").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationDiff))

	// Put,Patch,Delete with preifx /api/v1/applications/{name}
	appRouter := plugin.PathPrefix("/{name}").Subrouter()
	appRouter.Path("/clean").Methods("DELETE").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationCleanHandler))
	appRouter.PathPrefix("").Methods("PUT", "POST", "DELETE", "PATCH").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationEditHandler))
	// GET with prefix /api/v1/applications/{name}
	appRouter.PathPrefix("").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationViewsHandler))

	// NOTE: GET /api/v1/stream/applications?project={project}
	// NOTE: GET /api/v1/stream/applications/{name}/resource-tree
	blog.Infof("argocd application plugin init successfully")
	return nil
}

// POST /api/v1/applications, create new application
// validate project detail from request
func (plugin *AppPlugin) createApplicationHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	app := &v1alpha1.Application{}
	if err = json.Unmarshal(body, app); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal body failed"))
	}
	statusCode, err := plugin.middleware.CheckCreateApplication(r.Context(), app)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check create application failed"))
	}
	r = middleware.SetAuditMessage(r, app, middleware.ApplicationCreate)
	updatedBody, err := json.Marshal(app)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "json marshal application failed"))
	}
	r.Body = io.NopCloser(bytes.NewBuffer(updatedBody))
	length := len(updatedBody)
	r.Header.Set("Content-Length", strconv.Itoa(length))
	r.ContentLength = int64(length)
	return r, mw.ReturnArgoReverse()
}

// GET /api/v1/applications?projects={projects}
func (plugin *AppPlugin) listApplicationsHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
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
	appList, err := plugin.middleware.ListApplications(r.Context(), &appclient.ApplicationQuery{Projects: projects})
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError,
			errors.Wrapf(err, "list applications by project '%v' from storage failed", projects))
	}

	fields := r.URL.Query().Get("fields")
	if fields == "" {
		return r, mw.ReturnJSONResponse(appList)
	}
	bs, err := jsonq.ReserveField(appList, strings.Split(fields, ","))
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "parse by query 'fields' failed"))
	}
	return r, mw.ReturnDirectResponse(string(bs))
}

// Put,Patch,Delete with preifx /api/v1/applications/{name}
func (plugin *AppPlugin) applicationEditHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	app, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check application permission failed"))
	}
	urlPrefix := fmt.Sprintf("/api/v1/applications/%s", appName)
	switch r.RequestURI {
	case urlPrefix:
		if r.Method == http.MethodPut {
			r = middleware.SetAuditMessage(r, app, middleware.ApplicationUpdate)
		}
		if r.Method == http.MethodDelete {
			r = middleware.SetAuditMessage(r, app, middleware.ApplicationDelete)
		}
	case urlPrefix + "/sync":
		r = middleware.SetAuditMessage(r, app, middleware.ApplicationSync)
	case urlPrefix + "/rollback":
		r = middleware.SetAuditMessage(r, app, middleware.ApplicationRollback)
	case urlPrefix + "/resource":
		if r.Method == http.MethodDelete {
			r = middleware.SetAuditMessage(r, app, middleware.ApplicationDeleteResource)
		}
		if r.Method == http.MethodPatch {
			r = middleware.SetAuditMessage(r, app, middleware.ApplicationPatchResource)
		}
	}
	return r, mw.ReturnArgoReverse()
}

func (plugin *AppPlugin) applicationCleanHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	app, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	r = middleware.SetAuditMessage(r, app, middleware.ApplicationClean)
	if err = plugin.storage.DeleteApplicationResource(r.Context(), app); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse("clean application subresource success")
}

// GET with prefix /api/v1/applications/{name}
func (plugin *AppPlugin) applicationViewsHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	_, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName, iam.ProjectView)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	return r, mw.ReturnArgoReverse()
}
