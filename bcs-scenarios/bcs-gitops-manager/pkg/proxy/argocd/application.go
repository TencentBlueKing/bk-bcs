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
	bcsapi "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/analysis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/resources"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils/jsonq"
	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
)

// AppPlugin for internal project authorization
type AppPlugin struct {
	*mux.Router
	db             dao.Interface
	storage        store.Store
	middleware     mw.MiddlewareInterface
	analysisClient analysis.AnalysisInterface
	bcsStorage     bcsapi.Storage
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

	// 自定义接口
	plugin.Path("/dry-run").Methods("POST").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationDryRun))
	plugin.Path("/diff").Methods("POST").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationDiff))

	// Put,Patch,Delete with prefix /api/v1/applications/{name}
	appRouter := plugin.PathPrefix("/{name}").Subrouter()

	// 自定义接口
	appRouter.Path("/collect").Methods("PUT").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationCollect))
	appRouter.Path("/collect").Methods("DELETE").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationCancelCollect))
	appRouter.Path("/pod_resources").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationPodResources))
	appRouter.Path("/clean").Methods("DELETE").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationCleanHandler))
	appRouter.Path("/delete_resources").Methods("DELETE").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationDeleteResourcesHandler))
	appRouter.Path("/history_state").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationHistoryState))
	appRouter.Path("/custom_revisions").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.customRevisionsMetadata))

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
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("query param 'projects' cannot be empty"))
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
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "parse by query 'fields' failed"))
	}
	return r, mw.ReturnDirectResponse(string(bs))
}

// Put,Patch,Delete with prefix /api/v1/applications/{name}
func (plugin *AppPlugin) applicationEditHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	app, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName,
		iamnamespace.NameSpaceScopedUpdate)
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
	app, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName,
		iamnamespace.NameSpaceScopedDelete)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	r = middleware.SetAuditMessage(r, app, middleware.ApplicationClean)
	results := plugin.storage.DeleteApplicationResource(r.Context(), app, nil)
	errs := make([]string, 0)
	for i := range results {
		res := results[i]
		if !res.Succeeded {
			errs = append(errs, res.ErrMessage)
		}
	}
	if len(errs) != 0 {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError,
			errors.Errorf("clean application resources failed: %v", errs))
	}
	return r, mw.ReturnJSONResponse("clean application subresource success")
}

// ApplicationDeleteResourceRequest defines the request that delete resources of application
type ApplicationDeleteResourceRequest struct {
	Resources []*store.ApplicationResource `json:"resources"`
}

// ApplicationDeleteResourceResponse defines the response that delete resources of application
type ApplicationDeleteResourceResponse struct {
	Code      int32                                   `json:"code"`
	Message   string                                  `json:"message"`
	RequestID string                                  `json:"requestID"`
	Data      []store.ApplicationDeleteResourceResult `json:"data"`
}

func (plugin *AppPlugin) applicationDeleteResourcesHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	app, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName,
		iamnamespace.NameSpaceScopedDelete)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	req := new(ApplicationDeleteResourceRequest)
	if err = json.Unmarshal(body, req); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Wrapf(err, "unmarshal request body failed"))
	}
	if len(req.Resources) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Errorf("request body 'resources' cannot be empty"))
	}
	for i, resource := range req.Resources {
		if resource.ResourceName == "" || resource.Kind == "" || resource.Namespace == "" || resource.Version == "" {
			return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("request 'resources[%d] "+
				"have param empty, 'resourceName/kind/namespace/version' cannot be empty'", i))
		}
	}
	r = middleware.SetAuditMessage(r, app, middleware.ApplicationDeleteResources)
	results := plugin.storage.DeleteApplicationResource(r.Context(), app, req.Resources)
	return r, mw.ReturnJSONResponse(&ApplicationDeleteResourceResponse{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      results,
	})
}

// GET with prefix /api/v1/applications/{name}
func (plugin *AppPlugin) applicationViewsHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	_, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName,
		iamnamespace.NameSpaceScopedView)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	return r, mw.ReturnArgoReverse()
}

func (plugin *AppPlugin) applicationCollect(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	argoApp, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName,
		iamnamespace.NameSpaceScopedView)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	if err = plugin.analysisClient.ApplicationCollect(argoApp.Spec.Project, appName); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse("success")
}

func (plugin *AppPlugin) applicationCancelCollect(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	argoApp, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName,
		iamnamespace.NameSpaceScopedUpdate)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	if err = plugin.analysisClient.ApplicationCancelCollect(argoApp.Spec.Project, appName); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse("success")
}

func (plugin *AppPlugin) applicationPodResources(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	argoApp, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName,
		iamnamespace.NameSpaceScopedView)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	podQuery := resources.PodQuery{
		Storage:    plugin.storage,
		BCSStorage: plugin.bcsStorage,
	}
	pods, err := podQuery.Query(r.Context(), argoApp)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}

	fields := r.URL.Query().Get("fields")
	if fields == "" {
		return r, mw.ReturnJSONResponse(pods)
	}
	bs, err := jsonq.ReserveField(pods, strings.Split(fields, ","))
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "parse by query 'fields' failed"))
	}
	return r, mw.ReturnDirectResponse(string(bs))
}

// ApplicationRevisionsMetadata defines the response that get revision metadata of application's repo
// It adapt to multiple sources.
type ApplicationRevisionsMetadata struct {
	Code      int32                        `json:"code"`
	Message   string                       `json:"message"`
	RequestID string                       `json:"requestID"`
	Data      []*v1alpha1.RevisionMetadata `json:"data"`
}

func (plugin *AppPlugin) customRevisionsMetadata(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	argoApp, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName,
		iamnamespace.NameSpaceScopedView)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	revisions := r.URL.Query()["revisions"]
	if len(revisions) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("query parameter 'revisions' cannot be empty"))
	}
	for _, revision := range revisions {
		if revision == "" {
			return r, mw.ReturnErrorResponse(http.StatusBadRequest,
				fmt.Errorf("query parameter 'revisions' has empty value"))
		}
	}
	repos := make([]string, 0)
	if argoApp.Spec.HasMultipleSources() {
		if len(argoApp.Spec.Sources) != len(revisions) {
			return r, mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf("application has multiple(%d) "+
				"sources, not same as query param 'revisions'", len(argoApp.Spec.Sources)))
		}
		for _, source := range argoApp.Spec.Sources {
			repos = append(repos, source.RepoURL)
		}
	} else {
		if len(revisions) != 1 {
			return r, mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf("application has single source"))
		}
		repos = append(repos, argoApp.Spec.Source.RepoURL)
	}

	revisionsMetadata, err := plugin.storage.GetApplicationRevisionsMetadata(r.Context(), repos, revisions)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError,
			fmt.Errorf("get application revisions metadata failed: %s", err.Error()))
	}
	return r, mw.ReturnJSONResponse(&ApplicationRevisionsMetadata{
		Code:      0,
		RequestID: mw.RequestID(r.Context()),
		Data:      revisionsMetadata,
	})
}
