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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	argorepo "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

var (
	// grpcAccessUrl 定义 grpc 模式下准入的 API 列表，及处理方法
	grpcAccessUrlHandlers map[string]middleware.HttpHandler
)

// GrpcPlugin for internal project authorization
type GrpcPlugin struct {
	*mux.Router
	middleware    mw.MiddlewareInterface
	permitChecker permitcheck.PermissionInterface
}

// Init the grpc plugin
// 参见: github.com/argoproj/argocd/v2/cmd/argocd/commands
func (plugin *GrpcPlugin) Init() error {
	grpcAccessUrlHandlers = map[string]middleware.HttpHandler{
		"/project.ProjectService/List":               plugin.handleProjectList,
		"/project.ProjectService/GetDetailedProject": plugin.handleProjectGet,
		"/project.ProjectService/Get":                plugin.handleProjectGet,

		"/repository.RepositoryService/ListRepositories": plugin.handleRepoList,
		"/repository.RepositoryService/Get":              plugin.handleRepoGet,
		"/repository.RepositoryService/ValidateAccess":   plugin.handleRepoAccess,
		"/repository.RepositoryService/CreateRepository": plugin.handleRepoCreate,
		"/repository.RepositoryService/DeleteRepository": plugin.handleRepoDelete,
		// "/repository.RepositoryService/ListRefs":         nil,
		// "/repository.RepositoryService/ListApps":         nil,
		// "/repository.RepositoryService/GetAppDetails":    nil,
		// "/repository.RepositoryService/GetHelmCharts":    nil,

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

		"/applicationset.ApplicationSetService/List":   plugin.handleAppSetList,
		"/applicationset.ApplicationSetService/Get":    plugin.handleAppSetGet,
		"/applicationset.ApplicationSetService/Create": plugin.handleAppSetCreate,
		"/applicationset.ApplicationSetService/Delete": plugin.handleAppSetDelete,

		"/version.VersionService/Version": plugin.handleVersion,
	}
	plugin.Path("").Handler(plugin.middleware.HttpWrapper(plugin.serve))
	return nil
}

// ServeHTTP http handler implementation
func (plugin *GrpcPlugin) serve(r *http.Request) (*http.Request, *mw.HttpResponse) {
	if !proxy.IsGitOpsClient(r) {
		return r, mw.ReturnGRPCErrorResponse(http.StatusForbidden, fmt.Errorf("request not come from admin"))
	}
	handler, ok := grpcAccessUrlHandlers[strings.TrimPrefix(r.URL.Path, common.GitOpsProxyURL)]
	if !ok {
		return r, mw.ReturnGRPCErrorResponse(http.StatusForbidden,
			fmt.Errorf("request url '%s' unahtourized", r.URL.Path))
	}
	return handler(r)
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

func (plugin *GrpcPlugin) readRequestBody(req *http.Request, query interface{}) error {
	bs, err := io.ReadAll(req.Body)
	if err != nil {
		return errors.Wrapf(err, "read request body failed")
	}
	req.Body = io.NopCloser(bytes.NewBuffer(bs))
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
	req.Body = io.NopCloser(bytes.NewBuffer(result))
	return len(result), nil
}

// handleProjectList will handle the grpc request of list project
func (plugin *GrpcPlugin) handleProjectList(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projectList, statusCode, err := plugin.middleware.ListProjects(r.Context())
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "list projects failed"))
	}
	return r, mw.ReturnGRPCResponse(projectList)
}

// handleProjectGet will return project details by project name
func (plugin *GrpcPlugin) handleProjectGet(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &project.ProjectQuery{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), query.Name,
		permitcheck.ProjectViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err,
			"check project '%s' view permission failed", query.Name))
	}
	return r, mw.ReturnArgoReverse()
}

// handleRepoList will return repo list
func (plugin *GrpcPlugin) handleRepoList(r *http.Request) (*http.Request, *mw.HttpResponse) {
	repoList, statusCode, err := plugin.middleware.ListRepositories(r.Context(), nil, true)
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "list repositories failed"))
	}
	return r, mw.ReturnGRPCResponse(repoList)
}

// handleRepoGet will return repo details by repo name
func (plugin *GrpcPlugin) handleRepoGet(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &argorepo.RepoQuery{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	repo, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), query.Repo,
		permitcheck.RepoViewRSAction)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err,
			"check repo '%s' permission failed", query.Repo))
	}
	return r, mw.ReturnGRPCResponse(repo)
}

// handleRepoAccess will check repo access
func (plugin *GrpcPlugin) handleRepoAccess(r *http.Request) (*http.Request, *mw.HttpResponse) {
	repoAccess := &argorepo.RepoAccessQuery{}
	if err := plugin.readRequestBody(r, repoAccess); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	if repoAccess.Project == "" {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, fmt.Errorf("create repo request project cannot empty"))
	}
	_, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), repoAccess.Repo,
		permitcheck.RepoViewRSAction)
	if err != nil {
		// fix repo not create yet
		if strings.Contains(err.Error(), "not found") {
			return r, mw.ReturnArgoReverse()
		}
		return r, mw.ReturnGRPCErrorResponse(statusCode,
			errors.Wrapf(err, "check project '%s' edit permission failed", repoAccess.Project))
	}
	return r, mw.ReturnArgoReverse()
}

// handleRepoCreate will create repo to argocd
func (plugin *GrpcPlugin) handleRepoCreate(r *http.Request) (*http.Request, *mw.HttpResponse) {
	repoCreate := &argorepo.RepoCreateRequest{}
	if err := plugin.readRequestBody(r, repoCreate); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	if repoCreate.Repo == nil || repoCreate.Repo.Project == "" {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest,
			fmt.Errorf("create repo request project cannot empty"))
	}
	statusCode, err := plugin.permitChecker.CheckRepoCreate(r.Context(), repoCreate.Repo)
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode,
			errors.Wrapf(err, "check project '%s' edit permission failed", repoCreate.Repo.Project))
	}

	r = ctxutils.SetAuditMessage(r, &dao.UserAudit{
		Project:      repoCreate.Repo.Project,
		Action:       string(ctxutils.RepoCreate),
		ResourceType: string(ctxutils.RepoResource),
		ResourceName: repoCreate.Repo.Repo,
		ResourceData: ctxutils.EmptyData,
		RequestType:  string(ctxutils.GRPCRequest),
	})
	return r, mw.ReturnArgoReverse()
}

// handleRepoDelete will delete repo from argocd
func (plugin *GrpcPlugin) handleRepoDelete(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &argorepo.RepoQuery{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	if query.Repo == "" {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, fmt.Errorf("delete repo request repo cannot empty"))
	}
	argoRepo, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), query.Repo,
		permitcheck.RepoDeleteRSAction)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(statusCode,
			errors.Wrapf(err, "check repo '%s' permission failed", query.Repo))
	}
	r = ctxutils.SetAuditMessage(r, &dao.UserAudit{
		Project:      argoRepo.Project,
		Action:       string(ctxutils.RepoDelete),
		ResourceType: string(ctxutils.RepoResource),
		ResourceName: argoRepo.Repo,
		ResourceData: ctxutils.EmptyData,
		RequestType:  string(ctxutils.GRPCRequest),
	})
	return r, mw.ReturnArgoReverse()
}

// handleRepoListRefs will list repo refs from argocd
// nolint unused
func (plugin *GrpcPlugin) handleRepoListRefs(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &argorepo.RepoQuery{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	if query.Repo == "" {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, fmt.Errorf("delete repo request repo cannot empty"))
	}
	_, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), query.Repo,
		permitcheck.RepoViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode,
			errors.Wrapf(err, "check repo '%s' permission failed", query.Repo))
	}
	return r, mw.ReturnArgoReverse()
}

// handleRepoListApps will handle repo list apps
// nolint unused
func (plugin *GrpcPlugin) handleRepoListApps(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &argorepo.RepoAppsQuery{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	if query.Repo == "" {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, fmt.Errorf("delete repo request repo cannot empty"))
	}
	_, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), query.Repo,
		permitcheck.RepoViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode,
			errors.Wrapf(err, "check repo '%s' permission failed", query.Repo))
	}
	return r, mw.ReturnArgoReverse()
}

// handleRepoGetAppDetails will handle repo get application details
// nolint unused
func (plugin *GrpcPlugin) handleRepoGetAppDetails(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &argorepo.RepoAppDetailsQuery{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	if query.Source.RepoURL == "" {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, fmt.Errorf("delete repo request repo cannot empty"))
	}
	_, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), query.Source.RepoURL,
		permitcheck.RepoViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode,
			errors.Wrapf(err, "check repo '%s' permission failed", query.Source.RepoURL))
	}
	return r, mw.ReturnArgoReverse()
}

// handleRepoGetHelmCharts will handle repo get helm charts
// nolint unused
func (plugin *GrpcPlugin) handleRepoGetHelmCharts(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &argorepo.RepoQuery{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	if query.Repo == "" {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest,
			fmt.Errorf("delete repo request repo cannot empty"))
	}
	_, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), query.Repo,
		permitcheck.RepoViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "check repo '%s' permission failed",
			query.Repo))
	}
	return r, mw.ReturnArgoReverse()
}

// handleClusterList will handle cluster list
func (plugin *GrpcPlugin) handleClusterList(r *http.Request) (*http.Request, *mw.HttpResponse) {
	clusters, statusCode, err := plugin.middleware.ListClusters(r.Context(), nil)
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "list clusters failed"))
	}
	return r, mw.ReturnGRPCResponse(clusters)
}

// handleClusterGet will handle cluster get, return cluster details
func (plugin *GrpcPlugin) handleClusterGet(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &cluster.ClusterQuery{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	_, statusCode, err := plugin.permitChecker.CheckClusterPermission(r.Context(), query,
		permitcheck.ClusterViewRSAction)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(statusCode,
			errors.Wrapf(err, "check cluster '%s' permission failed", query.Name))
	}
	return r, mw.ReturnArgoReverse()
}

func (plugin *GrpcPlugin) handleClusterSettingGet(r *http.Request) (*http.Request, *mw.HttpResponse) {
	return r, mw.ReturnArgoReverse()
}

// handleAppSetList will handle applicationSet list, return applicationSets
func (plugin *GrpcPlugin) handleAppSetList(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projectList, statusCode, err := plugin.middleware.ListProjects(r.Context())
	if statusCode != http.StatusOK {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "list projects failed"))
	}
	query := new(applicationset.ApplicationSetListQuery)
	if err = plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	names := make([]string, 0, len(projectList.Items))
	if len(query.Projects) != 0 {
		queryProjects := make(map[string]struct{})
		for _, pro := range query.Projects {
			queryProjects[pro] = struct{}{}
		}
		for i := range projectList.Items {
			item := projectList.Items[i]
			if _, ok := queryProjects[item.Name]; ok {
				names = append(names, item.Name)
			}
		}
	} else {
		for i := range projectList.Items {
			item := projectList.Items[i]
			names = append(names, item.Name)
		}
	}
	query.Projects = names
	appsetList, statusCode, err := plugin.middleware.ListApplicationSets(r.Context(), query)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "list applicationsets by project "+
			"'%s' from storage failed", names))
	}
	result := make([]v1alpha1.ApplicationSet, 0, len(appsetList.Items))
	result = append(result, appsetList.Items...)
	appsetList.Items = result
	return r, mw.ReturnGRPCResponse(appsetList)
}

// handleAppSetGet handle application get, return application details
func (plugin *GrpcPlugin) handleAppSetGet(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &applicationset.ApplicationSetGetQuery{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	argoAppSet, statusCode, err := plugin.permitChecker.CheckAppSetPermission(r.Context(), query.Name,
		permitcheck.AppSetViewRSAction)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(statusCode,
			errors.Wrapf(err, "check applicationset '%s' failed", query.Name))
	}
	return r, mw.ReturnGRPCResponse(argoAppSet)
}

// handleAppSetCreate handle application create
func (plugin *GrpcPlugin) handleAppSetCreate(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appCreate := &applicationset.ApplicationSetCreateRequest{}
	if err := plugin.readRequestBody(r, appCreate); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	_, statusCode, err := plugin.permitChecker.CheckAppSetCreate(r.Context(), appCreate.Applicationset)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "check create applicationset failed"))
	}

	bs, _ := json.Marshal(appCreate)
	r = ctxutils.SetAuditMessage(r, &dao.UserAudit{
		Project:      appCreate.Applicationset.Spec.Template.Spec.Project,
		Action:       string(ctxutils.ApplicationSetCreateOrUpdate),
		ResourceType: string(ctxutils.AppSetResource),
		ResourceName: appCreate.Applicationset.Name,
		ResourceData: string(bs),
		RequestType:  string(ctxutils.GRPCRequest),
	})
	contentLen, err := plugin.rewriteRequestBody(r, appCreate)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "rewrite request body failed"))
	}
	r.Header.Set("Content-Length", strconv.Itoa(contentLen))
	r.ContentLength = int64(contentLen)
	return r, mw.ReturnArgoReverse()
}

// handleAppSetDelete will handle application delete
func (plugin *GrpcPlugin) handleAppSetDelete(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appDelete := &applicationset.ApplicationSetDeleteRequest{}
	if err := plugin.readRequestBody(r, appDelete); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	argoAppSet, statusCode, err := plugin.permitChecker.CheckAppSetPermission(r.Context(), appDelete.Name,
		permitcheck.AppSetDeleteRSAction)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "check delete applicationset failed"))
	}
	bs, _ := json.Marshal(appDelete)
	r = ctxutils.SetAuditMessage(r, &dao.UserAudit{
		Project:      argoAppSet.Spec.Template.Spec.Project,
		Action:       string(ctxutils.ApplicationSetDelete),
		ResourceType: string(ctxutils.AppSetResource),
		ResourceName: appDelete.Name,
		ResourceData: string(bs),
		RequestType:  string(ctxutils.GRPCRequest),
	})
	return r, mw.ReturnArgoReverse()
}

// handleAppList will handle application list, return applications
func (plugin *GrpcPlugin) handleAppList(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := new(application.ApplicationQuery)
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}

	appList, statusCode, err := plugin.middleware.ListApplications(r.Context(), query)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "list applications by project "+
			"from storage failed"))
	}
	return r, mw.ReturnGRPCResponse(appList)
}

// handleAppGet handle application get, return application details
func (plugin *GrpcPlugin) handleAppGet(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &application.ApplicationQuery{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	_, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), *query.Name,
		permitcheck.AppViewRSAction)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(statusCode,
			errors.Wrapf(err, "check application '%s' permission failed", *query.Name))
	}
	return r, mw.ReturnArgoReverse()
}

// handleAppCreate handle application create
func (plugin *GrpcPlugin) handleAppCreate(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appCreate := &application.ApplicationCreateRequest{}
	if err := plugin.readRequestBody(r, appCreate); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	statusCode, err := plugin.permitChecker.CheckApplicationCreate(r.Context(), appCreate.Application)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err, "check create application failed"))
	}

	bs, _ := json.Marshal(appCreate)
	r = ctxutils.SetAuditMessage(r, &dao.UserAudit{
		Project:      appCreate.Application.Spec.Project,
		Action:       string(ctxutils.RepoDelete),
		ResourceType: string(ctxutils.RepoResource),
		ResourceName: appCreate.Application.Name,
		ResourceData: string(bs),
		RequestType:  string(ctxutils.GRPCRequest),
	})
	contentLen, err := plugin.rewriteRequestBody(r, appCreate)
	if err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "rewrite request body failed"))
	}
	r.Header.Set("Content-Length", strconv.Itoa(contentLen))
	r.ContentLength = int64(contentLen)
	return r, mw.ReturnArgoReverse()
}

// handleAppSync will handle application sync
func (plugin *GrpcPlugin) handleAppSync(r *http.Request) (*http.Request, *mw.HttpResponse) {
	query := &application.ApplicationSyncRequest{}
	if err := plugin.readRequestBody(r, query); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *query.Name, permitcheck.AppUpdateRSAction,
		ctxutils.ApplicationSync, query)
}

// handleAppDelete will handle application delete
func (plugin *GrpcPlugin) handleAppDelete(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appDelete := &application.ApplicationDeleteRequest{}
	if err := plugin.readRequestBody(r, appDelete); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appDelete.Name, permitcheck.AppDeleteRSAction,
		ctxutils.ApplicationDelete, appDelete)
}

// handleAppWatch will handle application watch
func (plugin *GrpcPlugin) handleAppWatch(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appWatch := new(application.ApplicationQuery)
	if err := plugin.readRequestBody(r, appWatch); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appWatch.Name, permitcheck.AppViewRSAction, ctxutils.None, nil)
}

// handleAppUpdate will handle application update
func (plugin *GrpcPlugin) handleAppUpdate(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appUpdate := &application.ApplicationUpdateRequest{}
	if err := plugin.readRequestBody(r, appUpdate); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, appUpdate.Application.Name, permitcheck.AppUpdateRSAction,
		ctxutils.ApplicationUpdate, appUpdate)
}

// handleAppUpdateSpec will handle application update spec information
func (plugin *GrpcPlugin) handleAppUpdateSpec(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationUpdateSpecRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppUpdateRSAction,
		ctxutils.ApplicationUpdate, appReq)
}

// handleAppPatch handle application patch
func (plugin *GrpcPlugin) handleAppPatch(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationPatchRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppUpdateRSAction,
		ctxutils.ApplicationUpdate, appReq)
}

// handleAppListResourceEvents handle application list resource events
func (plugin *GrpcPlugin) handleAppListResourceEvents(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationResourceEventsQuery)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppViewRSAction, ctxutils.None, nil)
}

// handleAppGetApplicationSyncWindows handle application sync windows
func (plugin *GrpcPlugin) handleAppGetApplicationSyncWindows(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationSyncWindowsQuery)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppViewRSAction, ctxutils.None, nil)
}

// handleAppRevisionMetadata handle application revision metadata
func (plugin *GrpcPlugin) handleAppRevisionMetadata(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.RevisionMetadataQuery)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppViewRSAction, ctxutils.None, nil)
}

// handleAppGetManifests handle application get manifests
func (plugin *GrpcPlugin) handleAppGetManifests(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationManifestQuery)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppViewRSAction, ctxutils.None, nil)
}

// handleAppManagedResources handle application managed resources
func (plugin *GrpcPlugin) handleAppManagedResources(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ResourcesQuery)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.ApplicationName, permitcheck.AppViewRSAction, ctxutils.None, nil)
}

// handleAppResourceTree handle application resource tree
func (plugin *GrpcPlugin) handleAppResourceTree(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ResourcesQuery)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.ApplicationName, permitcheck.AppViewRSAction, ctxutils.None, nil)
}

// handleAppRollback handle application rollback
func (plugin *GrpcPlugin) handleAppRollback(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationRollbackRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppUpdateRSAction,
		ctxutils.ApplicationRollback, appReq)
}

// handleAppTerminateOperation handle application termination operator
func (plugin *GrpcPlugin) handleAppTerminateOperation(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.OperationTerminateRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppUpdateRSAction,
		ctxutils.ApplicationUpdate, appReq)
}

// handleAppGetResource handle application get resource
func (plugin *GrpcPlugin) handleAppGetResource(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationResourceRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppViewRSAction, ctxutils.None, nil)
}

// handleAppPatchResource handle application patch resource
func (plugin *GrpcPlugin) handleAppPatchResource(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationResourcePatchRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppUpdateRSAction,
		ctxutils.ApplicationUpdate, appReq)
}

// handleAppListResourceActions handle application list resource actions
func (plugin *GrpcPlugin) handleAppListResourceActions(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationResourceRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppViewRSAction,
		ctxutils.None, nil)
}

// handleAppRunResourceAction handle application run resource action
func (plugin *GrpcPlugin) handleAppRunResourceAction(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ResourceActionRunRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppUpdateRSAction,
		ctxutils.ApplicationUpdate, appReq)
}

// handleAppDeleteResource handle application delete resource
func (plugin *GrpcPlugin) handleAppDeleteResource(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationResourceDeleteRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppDeleteRSAction,
		ctxutils.ApplicationDelete, appReq)
}

// handleAppPodLogs handle application pod logs
func (plugin *GrpcPlugin) handleAppPodLogs(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationPodLogsQuery)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppViewRSAction,
		ctxutils.None, nil)
}

// handleAppListLinks handle application list links
func (plugin *GrpcPlugin) handleAppListLinks(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ListAppLinksRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppViewRSAction,
		ctxutils.None, nil)
}

// handleAppListResourceLinks handle application list resource links
func (plugin *GrpcPlugin) handleAppListResourceLinks(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appReq := new(application.ApplicationResourceRequest)
	if err := plugin.readRequestBody(r, appReq); err != nil {
		return r, mw.ReturnGRPCErrorResponse(http.StatusBadRequest, err)
	}
	return plugin.handleAppCommon(r, *appReq.Name, permitcheck.AppViewRSAction,
		ctxutils.None, nil)
}

// handleAppCommon handle application common handler
func (plugin *GrpcPlugin) handleAppCommon(r *http.Request, appName string, actionID permitcheck.RSAction,
	auditAction ctxutils.AuditAction, auditData interface{}) (*http.Request, *mw.HttpResponse) {
	argoApp, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appName, actionID)
	if err != nil {
		if !utils.IsClusterNotFound(err) {
			return r, mw.ReturnGRPCErrorResponse(statusCode, errors.Wrapf(err,
				"check application '%s' permission failed", appName))
		}
		return r, mw.ReturnArgoReverse()
	}
	if auditAction == ctxutils.None {
		return r, mw.ReturnArgoReverse()
	}

	var resourceData string
	if auditData != nil {
		bs, _ := json.Marshal(auditData)
		resourceData = string(bs)
	}
	auditMessage := &dao.UserAudit{
		Project:      argoApp.Spec.Project,
		ResourceType: string(ctxutils.ApplicationResource),
		ResourceName: argoApp.Name,
		RequestType:  string(ctxutils.GRPCRequest),
		Action:       string(auditAction),
		ResourceData: resourceData,
	}
	r = ctxutils.SetAuditMessage(r, auditMessage)
	return r, mw.ReturnArgoReverse()
}

// handleAppListResourceLinks handle application list resource links
func (plugin *GrpcPlugin) handleVersion(r *http.Request) (*http.Request, *mw.HttpResponse) {
	return r, mw.ReturnArgoReverse()
}
