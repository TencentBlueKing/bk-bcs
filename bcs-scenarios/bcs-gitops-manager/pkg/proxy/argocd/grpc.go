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
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	argorepo "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
)

type argoGrpcHandler func(ctx context.Context, req *http.Request) *httpResponse

var (
	// grpcAccessUrl 定义 grpc 模式下准入的 API 列表，及处理方法
	grpcAccessUrlHandlers map[string]argoGrpcHandler
)

// GrpcPlugin for internal project authorization
type GrpcPlugin struct {
	*mux.Router
	middleware MiddlewareInterface
}

// Init the grpc plugin
// 参见: github.com/argoproj/argocd/v2/cmd/argocd/commands
func (plugin *GrpcPlugin) Init() error {
	grpcAccessUrlHandlers = map[string]argoGrpcHandler{
		"/project.ProjectService/List":               plugin.handleProjectList,
		"/project.ProjectService/GetDetailedProject": plugin.handleProjectGet,
		"/project.ProjectService/Get":                plugin.handleProjectGet,

		"/repository.RepositoryService/ListRepositories": plugin.handleRepoList,
		"/repository.RepositoryService/Get":              plugin.handleRepoGet,
		"/repository.RepositoryService/ValidateAccess":   plugin.handleRepoAccess,
		"/repository.RepositoryService/CreateRepository": plugin.handleRepoCreate,
		"/repository.RepositoryService/DeleteRepository": plugin.handleRepoDelete,
		"/repository.RepositoryService/ListRefs":         nil,
		"/repository.RepositoryService/ListApps":         nil,
		"/repository.RepositoryService/GetAppDetails":    nil,
		"/repository.RepositoryService/GetHelmCharts":    nil,

		"/cluster.ClusterService/List": plugin.handleClusterList,
		"/cluster.SettingsService/Get": plugin.handleClusterSettingGet,
		"/cluster.ClusterService/Get":  plugin.handleClusterGet,

		"/application.ApplicationService/List":                      plugin.handleAppList,
		"/application.ApplicationService/Get":                       plugin.handleAppGet,
		"/application.ApplicationService/Create":                    plugin.handleAppCreate,
		"/application.ApplicationService/Sync":                      plugin.handleAppSync,
		"/application.ApplicationService/Watch":                     plugin.handleAppWatch,
		"/application.ApplicationService/Delete":                    plugin.handleAppDelete,
		"/application.ApplicationService/Update":                    plugin.handleAppUpdate,
		"/application.ApplicationService/UpdateSpec":                plugin.handleAppUpdateSpec,
		"/application.ApplicationService/Patch":                     plugin.handleAppPatch,
		"/application.ApplicationService/ListResourceEvents":        plugin.handleAppListResourceEvents,
		"/application.ApplicationService/GetApplicationSyncWindows": plugin.handleAppGetApplicationSyncWindows,
		"/application.ApplicationService/RevisionMetadata":          plugin.handleAppRevisionMetadata,
		"/application.ApplicationService/GetManifests":              plugin.handleAppGetManifests,
		"/application.ApplicationService/ManagedResources":          plugin.handleAppManagedResources,
		"/application.ApplicationService/ResourceTree":              plugin.handleAppResourceTree,
		"/application.ApplicationService/Rollback":                  plugin.handleAppRollback,
		"/application.ApplicationService/TerminateOperation":        plugin.handleAppTerminateOperation,
		"/application.ApplicationService/GetResource":               plugin.handleAppGetResource,
		"/application.ApplicationService/PatchResource":             plugin.handleAppPatchResource,
		"/application.ApplicationService/ListResourceActions":       plugin.handleAppListResourceActions,
		"/application.ApplicationService/RunResourceAction":         plugin.handleAppRunResourceAction,
		"/application.ApplicationService/DeleteResource":            plugin.handleAppDeleteResource,
		"/application.ApplicationService/PodLogs":                   plugin.handleAppPodLogs,
		"/application.ApplicationService/ListLinks":                 plugin.handleAppListLinks,
		"/application.ApplicationService/ListResourceLinks":         plugin.handleAppListResourceLinks,
	}
	plugin.Path("").Handler(plugin.middleware.HttpWrapper(plugin.serve))
	return nil
}

// ServeHTTP http handler implementation
func (plugin *GrpcPlugin) serve(ctx context.Context, req *http.Request) *httpResponse {
	if !proxy.IsAdmin(req) {
		return &httpResponse{
			statusCode: http.StatusForbidden,
			err:        fmt.Errorf("request not come from admin"),
		}
	}
	handler, ok := grpcAccessUrlHandlers[strings.TrimPrefix(req.URL.Path, common.GitOpsProxyURL)]
	if !ok {
		return &httpResponse{
			statusCode: http.StatusForbidden,
			err:        fmt.Errorf("request url '%s' unahtourized", req.URL.Path),
		}
	}
	return handler(ctx, req)
}

// parseRequestBytes GRPC 的前 5 位为 header，第 1 位标注是否压缩, 第 2-5 位标注 body 长度。
func (plugin *GrpcPlugin) parseRequestBytes(request []byte) ([]byte, error) {
	if len(request) < 5 {
		return nil, fmt.Errorf("request body %v bytes not over 5", request)
	}
	// NOTE: 默认未压缩，此处不做处理
	_ = request[0]
	bodyBytes := request[1:5]
	bodyLen := binary.BigEndian.Uint32(bodyBytes)
	if len(request) < int(bodyLen+5) {
		return nil, fmt.Errorf("request body %v not normal", request)
	}
	return request[5 : bodyLen+5], nil
}

func (plugin *GrpcPlugin) readRequestBody(ctx context.Context, req *http.Request, query interface{}) error {
	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.Wrapf(err, "read request body failed")
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bs))
	body, err := plugin.parseRequestBytes(bs)
	if err != nil {
		return errors.Wrapf(err, "parse request body failed")
	}
	if err = encoding.GetCodec(proto.Name).Unmarshal(body, query); err != nil {
		return errors.Wrapf(err, "unmarshal request body failed")
	}
	return nil
}

// rewriteRequestBody 对于 Application Create 需要重写 request body
func (plugin *GrpcPlugin) rewriteRequestBody(req *http.Request, body interface{}) (int, error) {
	bodyBs, err := encoding.GetCodec(proto.Name).Marshal(body)
	if err != nil {
		return 0, errors.Wrapf(err, "encoding request body failed")
	}
	contentLen := make([]byte, 4)
	binary.BigEndian.PutUint32(contentLen, uint32(len(bodyBs)))
	result := make([]byte, 0, 5+len(bodyBs))
	result = append(result, 0)
	result = append(result, contentLen...)
	result = append(result, bodyBs...)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(result))
	return len(result), nil
}

// handleProjectList will handle the grpc request of list project
func (plugin *GrpcPlugin) handleProjectList(ctx context.Context, req *http.Request) *httpResponse {
	projectList, statusCode, err := plugin.middleware.ListProjects(ctx)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list projects failed"),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: statusCode,
		obj:        projectList,
	}
}

// handleProjectGet will return project details by project name
func (plugin *GrpcPlugin) handleProjectGet(ctx context.Context, req *http.Request) *httpResponse {
	query := &project.ProjectQuery{}
	if err := plugin.readRequestBody(ctx, req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	_, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, query.Name, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check project '%s' view permission failed", query.Name),
		}
	}
	return nil
}

// handleRepoList will return repo list
func (plugin *GrpcPlugin) handleRepoList(ctx context.Context, req *http.Request) *httpResponse {
	projectList, statusCode, err := plugin.middleware.ListProjects(ctx)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list projects failed"),
		}
	}
	names := make([]string, 0, len(projectList.Items))
	for _, item := range projectList.Items {
		names = append(names, item.Name)
	}
	repoList, statusCode, err := plugin.middleware.ListRepositories(ctx, names, false)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list repositories failed"),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: statusCode,
		obj:        repoList,
	}
}

// handleRepoGet will return repo details by repo name
func (plugin *GrpcPlugin) handleRepoGet(ctx context.Context, req *http.Request) *httpResponse {
	query := &argorepo.RepoQuery{}
	if err := plugin.readRequestBody(ctx, req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	repo, statusCode, err := plugin.middleware.CheckRepositoryPermission(ctx, query.Repo, iam.ProjectView)
	if err != nil {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check repo '%s' permission failed", query.Repo),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: statusCode,
		obj:        repo,
	}
}

// handleRepoAccess will check repo access
func (plugin *GrpcPlugin) handleRepoAccess(ctx context.Context, req *http.Request) *httpResponse {
	repoAccess := &argorepo.RepoAccessQuery{}
	if err := plugin.readRequestBody(ctx, req, repoAccess); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	if repoAccess.Project == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("create repo request project cannot empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, repoAccess.Project, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check project '%s' edit permission failed", repoAccess.Project),
		}
	}
	return nil
}

// handleRepoCreate will create repo to argocd
func (plugin *GrpcPlugin) handleRepoCreate(ctx context.Context, req *http.Request) *httpResponse {
	repoCreate := &argorepo.RepoCreateRequest{}
	if err := plugin.readRequestBody(ctx, req, repoCreate); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	if repoCreate.Repo == nil || repoCreate.Repo.Project == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("create repo request project cannot empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, repoCreate.Repo.Project, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check project '%s' edit permission failed", repoCreate.Repo.Project),
		}
	}
	return nil
}

// handleRepoDelete will delete repo from argocd
func (plugin *GrpcPlugin) handleRepoDelete(ctx context.Context, req *http.Request) *httpResponse {
	query := &argorepo.RepoQuery{}
	if err := plugin.readRequestBody(ctx, req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	if query.Repo == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("delete repo request repo cannot empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckRepositoryPermission(ctx, query.Repo, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check repo '%s' permission failed", query.Repo),
		}
	}
	return nil
}

// handleRepoListRefs will list repo refs from argocd
func (plugin *GrpcPlugin) handleRepoListRefs(ctx context.Context, req *http.Request) *httpResponse {
	query := &argorepo.RepoQuery{}
	if err := plugin.readRequestBody(ctx, req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	if query.Repo == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("delete repo request repo cannot empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckRepositoryPermission(ctx, query.Repo, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check repo '%s' permission failed", query.Repo),
		}
	}
	return nil
}

// handleRepoListApps will handle repo list apps
func (plugin *GrpcPlugin) handleRepoListApps(ctx context.Context, req *http.Request) *httpResponse {
	query := &argorepo.RepoAppsQuery{}
	if err := plugin.readRequestBody(ctx, req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	if query.Repo == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("delete repo request repo cannot empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckRepositoryPermission(ctx, query.Repo, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check repo '%s' permission failed", query.Repo),
		}
	}
	return nil
}

// handleRepoGetAppDetails will handle repo get application details
func (plugin *GrpcPlugin) handleRepoGetAppDetails(ctx context.Context, req *http.Request) *httpResponse {
	query := &argorepo.RepoAppDetailsQuery{}
	if err := plugin.readRequestBody(ctx, req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	if query.Source.RepoURL == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("delete repo request repo cannot empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckRepositoryPermission(ctx, query.Source.RepoURL, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check repo '%s' permission failed", query.Source.RepoURL),
		}
	}
	return nil
}

// handleRepoGetHelmCharts will handle repo get helm charts
func (plugin *GrpcPlugin) handleRepoGetHelmCharts(ctx context.Context, req *http.Request) *httpResponse {
	query := &argorepo.RepoQuery{}
	if err := plugin.readRequestBody(ctx, req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	if query.Repo == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("delete repo request repo cannot empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckRepositoryPermission(ctx, query.Repo, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check repo '%s' permission failed", query.Repo),
		}
	}
	return nil
}

// handleClusterList will handle cluster list
func (plugin *GrpcPlugin) handleClusterList(ctx context.Context, req *http.Request) *httpResponse {
	projectList, statusCode, err := plugin.middleware.ListProjects(ctx)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list projects failed"),
		}
	}
	names := make([]string, 0, len(projectList.Items))
	for _, item := range projectList.Items {
		names = append(names, item.Name)
	}
	blog.Infof("[requestID=%s] list cluster with projects: %v", ctx.Value(ctxKeyRequestID).(string), names)
	clusters, statusCode, err := plugin.middleware.ListClusters(ctx, names)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list clusters failed"),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: statusCode,
		obj:        clusters,
	}
}

// parseClusterName will parse cluster name and check it
func (plugin *GrpcPlugin) parseClusterName(server string) (string, error) {
	arr := strings.Split(server, "/")
	clusterID := arr[len(arr)-1]
	if !strings.HasPrefix(clusterID, "BCS-K8S-") {
		return "", errors.Errorf("parse cluster '%s' failed", server)
	}
	return clusterID, nil
}

// handleClusterGet will handle cluster get, return cluster details
func (plugin *GrpcPlugin) handleClusterGet(ctx context.Context, req *http.Request) *httpResponse {
	query := &cluster.ClusterQuery{}
	if err := plugin.readRequestBody(ctx, req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	clusterID, err := plugin.parseClusterName(query.Server)
	if err != nil {
		return &httpResponse{
			statusCode: http.StatusInternalServerError,
			err:        errors.Wrapf(err, "parse cluster server failed"),
		}
	}
	statusCode, err := plugin.middleware.CheckClusterPermission(ctx, clusterID, iam.ClusterView)
	if err != nil {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check application '%s' permission failed", query.Name),
		}
	}
	return nil
}

func (plugin *GrpcPlugin) handleClusterSettingGet(ctx context.Context, req *http.Request) *httpResponse {
	return nil
}

// handleAppList will handle application list, return applications
func (plugin *GrpcPlugin) handleAppList(ctx context.Context, req *http.Request) *httpResponse {
	projectList, statusCode, err := plugin.middleware.ListProjects(ctx)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "list projects failed"),
		}
	}
	names := make([]string, 0, len(projectList.Items))
	for _, item := range projectList.Items {
		names = append(names, item.Name)
	}

	appList, err := plugin.middleware.ListApplications(ctx, names)
	if err != nil {
		return &httpResponse{
			statusCode: http.StatusInternalServerError,
			err:        errors.Wrapf(err, "list applications by project '%s' from storage failed", names),
		}
	}
	return &httpResponse{
		isGrpc:     true,
		statusCode: http.StatusOK,
		obj:        appList,
	}
}

// handleAppGet handle application get, return application details
func (plugin *GrpcPlugin) handleAppGet(ctx context.Context, req *http.Request) *httpResponse {
	query := &application.ApplicationQuery{}
	if err := plugin.readRequestBody(ctx, req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	_, statusCode, err := plugin.middleware.CheckApplicationPermission(ctx, *query.Name, iam.ProjectView)
	if err != nil {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check application '%s' permission failed", *query.Name),
		}
	}
	return nil
}

// handleAppCreate handle application create
func (plugin *GrpcPlugin) handleAppCreate(ctx context.Context, req *http.Request) *httpResponse {
	appCreate := &application.ApplicationCreateRequest{}
	if err := plugin.readRequestBody(ctx, req, appCreate); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}

	projectName := appCreate.Application.Spec.Project
	if projectName == "" || projectName == "default" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        errors.Errorf("project information lost"),
		}
	}
	argoProject, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, projectName, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check application '%s' permission failed", projectName),
		}
	}
	// setting application name with project prefix
	if !strings.HasPrefix(appCreate.Application.Name, projectName+"-") {
		appCreate.Application.Name = projectName + "-" + appCreate.Application.Name
	}
	// setting control annotations
	if appCreate.Application.Annotations == nil {
		appCreate.Application.Annotations = make(map[string]string)
	}
	appCreate.Application.Annotations[common.ProjectIDKey] = common.GetBCSProjectID(argoProject.Annotations)
	appCreate.Application.Annotations[common.ProjectBusinessIDKey] =
		argoProject.Annotations[common.ProjectBusinessIDKey]
	contentLen, err := plugin.rewriteRequestBody(req, appCreate)
	if err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        errors.Wrapf(err, "rewrite request body failed"),
		}
	}
	req.Header.Set("Content-Length", strconv.Itoa(contentLen))
	req.ContentLength = int64(contentLen)
	return nil
}

// handleAppSync will handle application sync
func (plugin *GrpcPlugin) handleAppSync(ctx context.Context, req *http.Request) *httpResponse {
	query := &application.ApplicationSyncRequest{}
	if err := plugin.readRequestBody(ctx, req, query); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	return plugin.handleAppCommon(ctx, *query.Name, iam.ProjectEdit)
}

// handleAppDelete will handle application delete
func (plugin *GrpcPlugin) handleAppDelete(ctx context.Context, req *http.Request) *httpResponse {
	appDelete := &application.ApplicationDeleteRequest{}
	if err := plugin.readRequestBody(ctx, req, appDelete); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        err,
		}
	}
	return plugin.handleAppCommon(ctx, *appDelete.Name, iam.ProjectEdit)
}

// handleAppWatch will handle application watch
func (plugin *GrpcPlugin) handleAppWatch(ctx context.Context, req *http.Request) *httpResponse {
	appWatch := new(application.ApplicationQuery)
	if err := plugin.readRequestBody(ctx, req, appWatch); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appWatch.Name, iam.ProjectEdit)
}

// handleAppUpdate will handle application update
func (plugin *GrpcPlugin) handleAppUpdate(ctx context.Context, req *http.Request) *httpResponse {
	appUpdate := &application.ApplicationUpdateRequest{}
	if err := plugin.readRequestBody(ctx, req, appUpdate); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, appUpdate.Application.Name, iam.ProjectEdit)
}

// handleAppUpdateSpec will handle application update spec information
func (plugin *GrpcPlugin) handleAppUpdateSpec(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationUpdateSpecRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppPatch handle application patch
func (plugin *GrpcPlugin) handleAppPatch(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationPatchRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppListResourceEvents handle application list resource events
func (plugin *GrpcPlugin) handleAppListResourceEvents(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationResourceEventsQuery)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppGetApplicationSyncWindows handle application sync windows
func (plugin *GrpcPlugin) handleAppGetApplicationSyncWindows(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationSyncWindowsQuery)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppRevisionMetadata handle application revision metadata
func (plugin *GrpcPlugin) handleAppRevisionMetadata(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.RevisionMetadataQuery)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppGetManifests handle application get manifests
func (plugin *GrpcPlugin) handleAppGetManifests(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationManifestQuery)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppManagedResources handle application managed resources
func (plugin *GrpcPlugin) handleAppManagedResources(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ResourcesQuery)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.ApplicationName, iam.ProjectEdit)
}

// handleAppResourceTree handle application resource tree
func (plugin *GrpcPlugin) handleAppResourceTree(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ResourcesQuery)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.ApplicationName, iam.ProjectEdit)
}

// handleAppRollback handle application rollback
func (plugin *GrpcPlugin) handleAppRollback(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationRollbackRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppTerminateOperation handle application termination operator
func (plugin *GrpcPlugin) handleAppTerminateOperation(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.OperationTerminateRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppGetResource handle application get resource
func (plugin *GrpcPlugin) handleAppGetResource(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationResourceRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppPatchResource handle application patch resource
func (plugin *GrpcPlugin) handleAppPatchResource(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationResourcePatchRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppListResourceActions handle application list resource actions
func (plugin *GrpcPlugin) handleAppListResourceActions(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationResourceRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppRunResourceAction handle application run resource action
func (plugin *GrpcPlugin) handleAppRunResourceAction(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ResourceActionRunRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppDeleteResource handle application delete resource
func (plugin *GrpcPlugin) handleAppDeleteResource(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationResourceDeleteRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppPodLogs handle application pod logs
func (plugin *GrpcPlugin) handleAppPodLogs(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationPodLogsQuery)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppListLinks handle application list links
func (plugin *GrpcPlugin) handleAppListLinks(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ListAppLinksRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppListResourceLinks handle application list resource links
func (plugin *GrpcPlugin) handleAppListResourceLinks(ctx context.Context, req *http.Request) *httpResponse {
	appReq := new(application.ApplicationResourceRequest)
	if err := plugin.readRequestBody(ctx, req, appReq); err != nil {
		return &httpResponse{statusCode: http.StatusBadRequest, err: err}
	}
	return plugin.handleAppCommon(ctx, *appReq.Name, iam.ProjectEdit)
}

// handleAppCommon handle application common handler
func (plugin *GrpcPlugin) handleAppCommon(ctx context.Context, appName string, actionID iam.ActionID) *httpResponse {
	_, statusCode, err := plugin.middleware.CheckApplicationPermission(ctx, appName, actionID)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check application '%s' permission failed", appName),
		}
	}
	return nil
}
