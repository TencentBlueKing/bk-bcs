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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/resources"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store/secretstore"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils/jsonq"
)

// AppPlugin for internal project authorization
type AppPlugin struct {
	*mux.Router
	db            dao.Interface
	storage       store.Store
	middleware    mw.MiddlewareInterface
	permitChecker permitcheck.PermissionInterface
	bcsStorage    bcsapi.Storage
	secretStore   secretstore.SecretInterface
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
// PUT:    /api/v1/applications/{name}/parameters/set, 设置app临时参数
// PUT:    /api/v1/applications/{name}/parameters/unset, 取消app临时参数
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
	appRouter.Path("/parameters/set").Methods("PUT").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationParameters))
	appRouter.Path("/parameters/unset").Methods("PUT").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationParameters))
	appRouter.Path("/workload_replicas_zero").Methods("PUT").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationWorkloadReplicasZero))
	appRouter.Path("/sync_refresh").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.syncRefresh))
	appRouter.Path("/forbid_sync").Methods("POST").
		Handler(plugin.middleware.HttpWrapper(plugin.forbidAppSync))

	appRouter.PathPrefix("").Methods("PUT", "POST", "DELETE", "PATCH").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationEditHandler))
	// GET with prefix /api/v1/applications/{name}
	appRouter.Path("").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationGetHandler))
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
	statusCode, err := plugin.permitChecker.CheckApplicationCreate(r.Context(), app)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check create application failed"))
	}
	r = plugin.setApplicationAudit(r, app.Spec.Project, app.Name, ctxutils.ApplicationCreate, string(body))
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
	query := &appclient.ApplicationQuery{Projects: projects}
	repo := r.URL.Query().Get("repo")
	if repo != "" {
		query.Repo = &repo
	}
	targetRevision := r.URL.Query().Get("targetRevision")
	if targetRevision != "" && repo == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("query param 'repo' cannot be empty when target revision is not empty"))
	}
	selectors := r.URL.Query()["selector"]
	appList := &v1alpha1.ApplicationList{Items: make([]v1alpha1.Application, 0)}
	if len(selectors) != 0 {
		for _, selector := range selectors {
			query.Selector = &selector
			apps, statusCode, err := plugin.middleware.ListApplications(r.Context(), query)
			if err != nil {
				return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "list applications by project '%v' from "+
					"storage failed", projects))
			}
			if len(apps.Items) != 0 {
				appList.ListMeta = apps.ListMeta
				appList.TypeMeta = apps.TypeMeta
				appList.Items = append(appList.Items, apps.Items...)
			}
		}
	} else {
		apps, statusCode, err := plugin.middleware.ListApplications(r.Context(), query)
		if err != nil {
			return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "list applications by project '%v' from "+
				"storage failed", projects))
		}
		appList = apps
	}
	if targetRevision != "" {
		filterAppsByTargetRevision(appList, targetRevision, repo)
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
	argoApp, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName,
		permitcheck.AppUpdateRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check application permission failed"))
	}
	var action ctxutils.AuditAction
	var data string
	switch r.Method {
	case http.MethodPut, http.MethodPost, http.MethodPatch:
		var body []byte
		body, err = io.ReadAll(r.Body)
		if err != nil {
			blog.Errorf("RequestID[%s] read request body failed: %v", ctxutils.RequestID(r.Context()), err)
		} else {
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			data = string(body)
		}
		if strings.Contains(r.URL.String(), "sync") {
			action = ctxutils.ApplicationSync
		} else {
			action = ctxutils.ApplicationUpdate
		}
	case http.MethodDelete:
		action = ctxutils.ApplicationDelete
	}
	r = plugin.setApplicationAudit(r, argoApp.Spec.Project, argoApp.Name, action, data)
	return r, mw.ReturnArgoReverse()
}

func (plugin *AppPlugin) applicationCleanHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf("request application name cannot be empty"))
	}
	argoApp, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName,
		permitcheck.AppDeleteRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	r = plugin.setApplicationAudit(r, argoApp.Spec.Project, argoApp.Name, ctxutils.ApplicationDelete,
		ctxutils.EmptyData)
	results := plugin.storage.DeleteApplicationResource(r.Context(), argoApp, nil)
	errs := make([]string, 0)
	for i := range results {
		res := results[i]
		if !res.Succeeded {
			errs = append(errs, res.ErrMessage)
		}
	}
	if len(errs) != 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
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
	argoApp, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName,
		permitcheck.AppDeleteRSAction)
	if err != nil {
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
		if resource.ResourceName == "" || resource.Kind == "" || resource.Version == "" {
			return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("request 'resources[%d] "+
				"have param empty, 'resourceName/kind/version' cannot be empty'", i))
		}
	}
	r = plugin.setApplicationAudit(r, argoApp.Spec.Project, argoApp.Name, ctxutils.ApplicationDelete, string(body))
	results := plugin.storage.DeleteApplicationResource(r.Context(), argoApp, req.Resources)
	return r, mw.ReturnJSONResponse(&ApplicationDeleteResourceResponse{
		Code:      0,
		RequestID: ctxutils.RequestID(r.Context()),
		Data:      results,
	})
}

func (plugin *AppPlugin) applicationGetHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	argoApp, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName,
		permitcheck.AppViewRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	// 优化项：如果是单独获取 application 详情，则返回缓存内容
	queries := r.URL.Query()
	for k := range queries {
		// 如果出现 query 参数非只 projects，则表明非是获取 application 详情
		if k != "projects" {
			return r, mw.ReturnArgoReverse()
		}
	}
	blog.Infof("RequestID[%s] get application from cache", ctxutils.RequestID(r.Context()))
	return r, mw.ReturnJSONResponse(argoApp)
}

// GET with prefix /api/v1/applications/{name}
func (plugin *AppPlugin) applicationViewsHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	_, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName,
		permitcheck.AppViewRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	return r, mw.ReturnArgoReverse()
}

func (plugin *AppPlugin) applicationPodResources(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	argoApp, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName,
		permitcheck.AppViewRSAction)
	if err != nil {
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

	argoApp, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName,
		permitcheck.AppViewRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	if sourceNum := len(argoApp.Spec.GetSources()); sourceNum != len(revisions) {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, fmt.Errorf(
			"application have '%d' source, but query param 'revisions' have '%d (%v)'",
			sourceNum, len(revisions), revisions))
	}
	result := make([]*v1alpha1.RevisionMetadata, 0)
	for i := range argoApp.Spec.GetSources() {
		src := argoApp.Spec.GetSources()[i]
		// no-need get revision metadata for helm chart
		if src.IsHelm() {
			result = append(result, &v1alpha1.RevisionMetadata{
				Message: "Helm Chart",
			})
			continue
		}

		var metadata []*v1alpha1.RevisionMetadata
		if metadata, err = plugin.storage.GetApplicationRevisionsMetadata(r.Context(), argoApp.Spec.Project,
			[]string{src.RepoURL}, []string{revisions[i]}); err != nil {
			return r, mw.ReturnErrorResponse(http.StatusInternalServerError,
				fmt.Errorf("get application revisions metadata failed: %s", err.Error()))
		}
		result = append(result, metadata...)
	}
	return r, mw.ReturnJSONResponse(&ApplicationRevisionsMetadata{
		Code:      0,
		RequestID: ctxutils.RequestID(r.Context()),
		Data:      result,
	})
}

type parameterAction string

const (
	parameterSet   parameterAction = "set"
	parameterUnset parameterAction = "unset"
)

// set body: [param1=value1, param2=value2]
// unset body: [param1, param2]
func (plugin *AppPlugin) applicationParameters(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	var parameters []string
	if err = json.Unmarshal(body, &parameters); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal body failed"))
	}

	argoApp, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName,
		permitcheck.AppUpdateRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}

	r = plugin.setApplicationAudit(r, argoApp.Spec.Project, argoApp.Name, ctxutils.ApplicationUpdate, string(body))
	var action parameterAction
	if strings.HasSuffix(r.URL.Path, "/set") {
		action = parameterSet
	} else if strings.HasSuffix(r.URL.Path, "/unset") {
		action = parameterUnset
	}
	switch action {
	case parameterSet:
		if err = setParameterOverrides(argoApp, parameters, parameterSet); err != nil {
			return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
		}
	case parameterUnset:
		if err = setParameterOverrides(argoApp, parameters, parameterUnset); err != nil {
			return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
		}
	default:
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError,
			errors.Errorf("url path error: %s", r.URL.Path))
	}

	if _, err = plugin.storage.UpdateApplicationSpec(r.Context(), &appclient.ApplicationUpdateSpecRequest{
		Name: &argoApp.Name,
		Spec: &argoApp.Spec,
	}); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse("success")
}

// setParameterOverrides updates an existing or appends a new parameters override in the application
func setParameterOverrides(app *v1alpha1.Application, parameters []string, action parameterAction) error {
	if len(parameters) == 0 {
		return errors.New("parameters cannot be null")
	}
	source := app.Spec.GetSource()
	sourceType, err := source.ExplicitType()
	if err != nil {
		return errors.Wrapf(err, "check sourceType error")
	}

	if *sourceType != v1alpha1.ApplicationSourceTypeHelm {
		return errors.New("parameters can only be set against Helm applications")
	}
	if source.Helm == nil {
		return errors.New("source helm is nil cannot unset")
	}
	switch action {
	case parameterSet:
		for _, p := range parameters {
			newParam, err := v1alpha1.NewHelmParameter(p, false)
			if err != nil {
				return errors.Wrapf(err, "set helm parameter error")
			}
			source.Helm.AddParameter(*newParam)
		}
	case parameterUnset:
		for _, paramStr := range parameters {
			helmParams := source.Helm.Parameters
			for i, p := range helmParams {
				if p.Name == paramStr {
					source.Helm.Parameters = append(helmParams[0:i], helmParams[i+1:]...) // nolint gocritic
					break
				}
			}
		}
	default:
		return errors.Errorf("action error: '%s'", action)
	}

	return nil
}

func (plugin *AppPlugin) setApplicationAudit(r *http.Request, project, appName string,
	action ctxutils.AuditAction, data string) *http.Request {
	return ctxutils.SetAuditMessage(r, &dao.UserAudit{
		Project:      project,
		Action:       string(action),
		ResourceType: string(ctxutils.ApplicationResource),
		ResourceName: appName,
		ResourceData: data,
		RequestType:  string(ctxutils.HTTPRequest),
	})
}

var (
	workloadKind = map[string]struct{}{
		"Deployment":      {},
		"StatefulSet":     {},
		"GameStatefulSet": {},
		"GameDeployment":  {},
	}
)

const (
	defaultPathType  = "application/merge-patch+json"
	defaultPatchJson = `{"spec": {"replicas": 0}}`
)

// applicationWorkloadReplicasZero set app's workload replicas to zero
func (plugin *AppPlugin) applicationWorkloadReplicasZero(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	argoApp, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName,
		permitcheck.AppUpdateRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	r = plugin.setApplicationAudit(r, argoApp.Spec.Project, appName, ctxutils.ApplicationUpdate, ctxutils.EmptyData)
	errs := make([]string, 0)
	patched := make([]string, 0)
	for i := range argoApp.Status.Resources {
		res := argoApp.Status.Resources[i]
		if _, ok := workloadKind[res.Kind]; !ok {
			continue
		}
		patched = append(patched, fmt.Sprintf("%s/%s/%s-%s", res.Group, res.Version, res.Kind, res.Name))
		if err = plugin.storage.PatchApplicationResource(r.Context(), argoApp.Name, &res,
			defaultPatchJson, defaultPathType); err != nil {
			if utils.IsArgoNotFoundAsPartOf(err) {
				blog.Warnf("RequestID[%s] patch argo resource failed: %s", ctxutils.RequestID(r.Context()), err.Error())
				continue
			}
			errs = append(errs, err.Error())
		}
	}
	blog.Infof("RequestID[%s] patched resources: %s", ctxutils.RequestID(r.Context()), strings.Join(patched, ", "))
	if len(errs) != 0 {
		blog.Warnf("RequestID[%s] patch app '%s' resources failed: %s", ctxutils.RequestID(r.Context()),
			argoApp.Name, strings.Join(errs, "; "))
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			errors.Errorf("patch application workload to 0 failed: %s", strings.Join(errs, ", ")))
	}
	return r, mw.ReturnJSONResponse("ok")
}

func (plugin *AppPlugin) syncRefresh(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	argoApp, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName,
		permitcheck.AppViewRSAction)
	if err != nil {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	r = plugin.setApplicationAudit(r, argoApp.Spec.Project, appName, ctxutils.ApplicationRefresh, ctxutils.EmptyData)

	// refresh application
	blog.Infof("RequestID[%s] refresh application", ctxutils.RequestID(r.Context()))
	argoAppClient := plugin.storage.GetAppClient()
	refreshType := string(v1alpha1.RefreshTypeNormal)
	if _, err = argoAppClient.Get(r.Context(), &appclient.ApplicationQuery{
		Name:    &appName,
		Refresh: &refreshType,
	}); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, errors.Wrapf(err,
			"refresh application '%s' failed", appName))
	}
	// just sleep to wait application refresh if application sync by-local
	if len(argoApp.Status.Sync.Revisions) == 0 && argoApp.Status.Sync.Revision == "" {
		time.Sleep(5 * time.Second)
		return r, mw.ReturnJSONResponse(argoApp)
	}

	// get remote repo last-commit-id
	blog.Infof("RequestID[%s] start get last-commit ids", ctxutils.RequestID(r.Context()))
	lastCommitIDs, err := plugin.getApplicationLastCommitIDs(r.Context(), argoApp)
	if err != nil {
		// we cannot check without clone repo if targetRevision is just a commit-hash, just sleep 5 seconds
		blog.Warnf("RequestID[%s] got last-commit for '%s' failed: %s", ctxutils.
			RequestID(r.Context()), appName, err)
		time.Sleep(5 * time.Second)
		return r, mw.ReturnJSONResponse(argoApp)
	}
	blog.Infof("RequestID[%s] got the last-commit-ids: %v", ctxutils.RequestID(r.Context()), lastCommitIDs)
	// ticker for check application got the latest commit-id
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			argoApp, err = plugin.storage.GetApplication(r.Context(), appName)
			if err != nil {
				return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
			}
			if !argoApp.Spec.HasMultipleSources() {
				if len(lastCommitIDs) != 1 {
					return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf(
						"application with single resource got not '1' last-commit-ids '%v'", lastCommitIDs))
				}
				if !strings.HasPrefix(argoApp.Status.Sync.Revision, lastCommitIDs[0]) {
					continue
				}
				return r, mw.ReturnJSONResponse(argoApp)
			}
			revisions := argoApp.Status.Sync.Revisions
			if len(lastCommitIDs) != len(revisions) {
				return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf(
					"application with multi-resources got last-commit-ids '%v', not same with reivions '%v'",
					lastCommitIDs, revisions))
			}
			allMatch := true
			blog.Infof("RequestID[%s] check match between app-revisions(%v) and last-commits(%v)",
				revisions, lastCommitIDs)
			for i := range lastCommitIDs {
				if !strings.HasPrefix(revisions[i], lastCommitIDs[i]) {
					allMatch = false
				}
			}
			if allMatch {
				return r, mw.ReturnJSONResponse(argoApp)
			}
		case <-r.Context().Done():
			return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("sync-refresh timeout"))
		}
	}
}

func (plugin *AppPlugin) getApplicationLastCommitIDs(ctx context.Context, argoApp *v1alpha1.Application) ([]string,
	error) {
	lastCommitIDs := make([]string, 0)
	if !argoApp.Spec.HasMultipleSources() {
		repoURL := argoApp.Spec.Source.RepoURL
		revision := argoApp.Spec.Source.TargetRevision
		commitID, err := plugin.storage.GetRepoLastCommitID(ctx, argoApp.Spec.Project, repoURL, revision)
		if err != nil {
			return nil, err
		}
		lastCommitIDs = append(lastCommitIDs, commitID)
		return lastCommitIDs, nil
	}
	for i := range argoApp.Spec.Sources {
		source := argoApp.Spec.Sources[i]
		repoURL := source.RepoURL
		revision := source.TargetRevision
		commitID, err := plugin.storage.GetRepoLastCommitID(ctx, argoApp.Spec.Project, repoURL, revision)
		if err != nil {
			return nil, err
		}
		lastCommitIDs = append(lastCommitIDs, commitID)
	}
	return lastCommitIDs, nil
}

func filterAppsByTargetRevision(appList *v1alpha1.ApplicationList, target string, repo string) {
	filterApps := make([]v1alpha1.Application, 0)
	for _, item := range appList.Items {
		if item.Spec.HasMultipleSources() {
			for _, source := range item.Spec.Sources {
				if source.RepoURL == repo && source.TargetRevision == target {
					filterApps = append(filterApps, item)
				}
			}
			continue
		}
		if item.Spec.Source.TargetRevision == target {
			filterApps = append(filterApps, item)
		}
	}
	appList.Items = filterApps
}

func (plugin *AppPlugin) forbidAppSync(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	forbid := r.URL.Query()["forbid"]
	if len(forbid) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("param forbid must not empty"))
	}
	app, err := plugin.storage.GetApplication(r.Context(), appName)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	annotations := make(map[string]interface{})
	if forbid[0] == "false" {
		annotations[common.ApplicationSyncForbidden] = nil
	} else {
		annotations[common.ApplicationSyncForbidden] = "true"
	}
	if err := plugin.storage.PatchApplicationAnnotation(
		r.Context(), app.Name, app.Namespace, annotations); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnJSONResponse(map[string]string{"forbid": forbid[0]})
}
