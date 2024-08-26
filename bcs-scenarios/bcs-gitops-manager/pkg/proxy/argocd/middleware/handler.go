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

package middleware

import (
	"context"
	"net/http"
	"reflect"
	"runtime"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/project"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/exp/slices"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/session"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// MiddlewareInterface defines the middleware interface
// nolint
type MiddlewareInterface interface {
	Init() error
	HttpWrapper(handler HttpHandler) http.Handler

	CheckBusinessPermission(ctx context.Context, bizID string, action permitcheck.RSAction) (int, error)

	ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, int, error)
	ListClusters(ctx context.Context, projectNames []string) (*v1alpha1.ClusterList, int, error)
	ListRepositories(ctx context.Context, projectNames []string,
		needCheckPermission bool) (*v1alpha1.RepositoryList, int, error)
	ListApplications(ctx context.Context, query *appclient.ApplicationQuery) (*v1alpha1.ApplicationList, int, error)
	ListApplicationSets(ctx context.Context, query *appsetpkg.ApplicationSetListQuery) (
		*v1alpha1.ApplicationSetList, int, error)
}

// handler 定义 http 中间件处理对象
type handler struct {
	permitChecker permitcheck.PermissionInterface

	projectPermission   *project.BCSProjectPerm
	clusterPermission   *cluster.BCSClusterPerm
	namespacePermission *namespace.BCSNamespacePerm

	option *options.Options
	store  store.Store
	db     dao.Interface

	secretSession     *session.SecretSession
	monitorSession    *session.MonitorSession
	argoSession       *session.ArgoSession
	argoStreamSession *session.ArgoStreamSession
	terraformSession  *session.TerraformSession
	analysisSession   *session.AnalysisSession

	tracer func(context.Context) error
}

// NewMiddlewareHandler create handler instance
func NewMiddlewareHandler(permitChecker permitcheck.PermissionInterface) MiddlewareInterface {
	op := options.GlobalOptions()
	return &handler{
		option:              op,
		permitChecker:       permitChecker,
		db:                  dao.GlobalDB(),
		store:               store.GlobalStore(),
		argoSession:         session.NewArgoSession(),
		argoStreamSession:   session.NewArgoStreamSession(),
		secretSession:       session.NewSecretSession(),
		terraformSession:    session.NewTerraformSession(),
		analysisSession:     session.NewAnalysisSession(),
		monitorSession:      session.NewMonitorSession(),
		projectPermission:   project.NewBCSProjectPermClient(op.IAMClient),
		clusterPermission:   cluster.NewBCSClusterPermClient(op.IAMClient),
		namespacePermission: namespace.NewBCSNamespacePermClient(op.IAMClient),
	}
}

// Init will init the tracer
func (h *handler) Init() error {
	opts := []trace.Option{
		trace.OTLPEndpoint(h.option.TraceConfig.Endpoint),
	}
	attrs := make([]attribute.KeyValue, 0)
	attrs = append(attrs, attribute.String("bk.data.token", h.option.TraceConfig.Token))
	opts = append(opts, trace.ResourceAttrs(attrs))
	// InitTracingProvider Initializes an OTLP exporter, and configures the corresponding trace and
	// metric providers.
	tracer, err := trace.InitTracingProvider("bcs-gitops-manager", opts...)
	if err != nil {
		return errors.Wrapf(err, "init tracer failed")
	}
	h.tracer = tracer
	return nil
}

// HttpWrapper 创建 http wrapper 中间件
func (h *handler) HttpWrapper(handler HttpHandler) http.Handler {
	handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	hw := &httpWrapper{
		handler:           handler,
		handlerName:       handlerName,
		option:            h.option,
		argoSession:       h.argoSession,
		argoStreamSession: h.argoStreamSession,
		secretSession:     h.secretSession,
		terraformSession:  h.terraformSession,
		analysisSession:   h.analysisSession,
		monitorSession:    h.monitorSession,
	}
	blog.Infof("[Trace] request handler '%s' add to otel", handlerName)
	return otelhttp.NewHandler(hw, handlerName)
}

// CheckBusinessPermission 检查用户是否具备业务权限
func (h *handler) CheckBusinessPermission(ctx context.Context, bizID string, action permitcheck.RSAction) (int, error) {
	if bizID == "" {
		return http.StatusBadRequest, errors.Errorf("bizID cannot be empty")
	}
	projectList, statusCode, err := h.ListProjects(ctx)
	if err != nil {
		return statusCode, err
	}

	for _, proj := range projectList.Items {
		// GetBCSProjectBusinessKey return the business id of project
		projectBizID := common.GetBCSProjectBusinessKey(proj.Annotations)
		if projectBizID == bizID {
			// checkBCSProjectPermissionByID 检查登录态用户对于项目的权限
			_, statusCode, _ = h.permitChecker.CheckProjectPermission(ctx, proj.Name, action)
			// 只要拥有一个project的权限，则允许操作
			if statusCode == http.StatusOK {
				return http.StatusOK, nil
			}
		}
	}
	return http.StatusForbidden, errors.Errorf("businessID '%s' for action '%s' forbidden", bizID, action)
}

// ListProjects 根据用户权限列出具备权限的 Projects
func (h *handler) ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, int, error) {
	return h.listAuthorizedProjects(ctx, nil)
}

func (h *handler) listAuthorizedProjects(ctx context.Context, names []string) (*v1alpha1.AppProjectList, int, error) {
	projectList, err := h.store.ListProjects(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "list projects failed")
	}
	projectIDs := make([]string, 0, len(projectList.Items))
	controlledProjects := make(map[string]v1alpha1.AppProject)
	for i, proj := range projectList.Items {
		projectID := common.GetBCSProjectID(proj.Annotations)
		if projectID == "" {
			continue
		}
		if len(names) != 0 && !slices.Contains(names, proj.Name) {
			continue
		}
		controlledProjects[projectID] = projectList.Items[i]
		projectIDs = append(projectIDs, projectID)
	}
	if len(projectIDs) == 0 {
		return &v1alpha1.AppProjectList{}, http.StatusOK, nil
	}

	projectAuth, err := h.listAuthorizedProjectsByID(ctx, projectIDs)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "project permission auth center failed")
	}

	authedProjects := make([]v1alpha1.AppProject, 0)
	for projectID, auth := range projectAuth {
		if !auth {
			continue
		}
		authedProjects = append(authedProjects, controlledProjects[projectID])
	}
	projectList.Items = authedProjects
	return projectList, http.StatusOK, nil
}

func (h *handler) listAuthorizedProjectsByID(ctx context.Context, projectIDs []string) (map[string]bool, error) {
	result, err := h.permitChecker.GetProjectMultiPermission(ctx, projectIDs, []permitcheck.RSAction{
		permitcheck.ProjectViewRSAction})
	if err != nil {
		return nil, err
	}
	projectAuth := make(map[string]bool)
	for pid, permits := range result {
		projectAuth[pid] = permits[permitcheck.ProjectViewRSAction]
	}
	return projectAuth, nil
}

// ListClusters 根据项目名获取用户态下可以 view 的集群列表
func (h *handler) ListClusters(ctx context.Context, projectNames []string) (*v1alpha1.ClusterList, int, error) {
	projList, statusCode, err := h.listAuthorizedProjects(ctx, projectNames)
	if err != nil {
		return nil, statusCode, err
	}
	authedProjMap := make(map[string]struct{})
	for i := range projList.Items {
		authedProjMap[projList.Items[i].Name] = struct{}{}
	}

	result := &v1alpha1.ClusterList{}
	clusterList, err := h.store.ListCluster(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "list clusters from storage failure")
	}
	for _, cls := range clusterList.Items {
		if _, ok := authedProjMap[cls.Project]; !ok {
			continue
		}
		result.Items = append(result.Items, cls)
	}
	return result, http.StatusOK, nil
}

// ListRepositories 根据项目名称获取用户态下可以 view 的仓库列表
func (h *handler) ListRepositories(ctx context.Context, projectNames []string,
	needCheckPermission bool) (*v1alpha1.RepositoryList, int, error) {
	repositories, err := h.store.ListRepository(ctx, projectNames)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "list repository from storage failed")
	}
	if !needCheckPermission {
		return repositories, http.StatusOK, nil
	}
	projList, statusCode, err := h.listAuthorizedProjects(ctx, projectNames)
	if err != nil {
		return nil, statusCode, err
	}
	authedProjMap := make(map[string]struct{})
	for i := range projList.Items {
		authedProjMap[projList.Items[i].Name] = struct{}{}
	}
	repoList := &v1alpha1.RepositoryList{}
	for _, repo := range repositories.Items {
		if _, ok := authedProjMap[repo.Project]; ok {
			repoList.Items = append(repoList.Items, repo)
		}
	}
	return repoList, http.StatusOK, nil
}

// ListApplications 根据项目名称获取所有应用
func (h *handler) ListApplications(ctx context.Context, query *appclient.ApplicationQuery) (*v1alpha1.ApplicationList,
	int, error) {
	argoAppList, err := h.store.ListApplications(ctx, query)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "list application swith project '%v' failed",
			query.Projects)
	}
	queryProjMap := make(map[string]struct{})
	for i := range argoAppList.Items {
		queryProjMap[argoAppList.Items[i].Spec.Project] = struct{}{}
	}
	queryProjs := make([]string, 0)
	for proj := range queryProjMap {
		queryProjs = append(queryProjs, proj)
	}
	blog.Infof("RequestID[%s] queried applications output projects: %v", ctxutils.RequestID(ctx), queryProjs)
	projList, statusCode, err := h.listAuthorizedProjects(ctx, queryProjs)
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "list authorized projects failed")
	}
	authedProjMap := make(map[string]struct{})
	authedProjs := make([]string, 0)
	for i := range projList.Items {
		projName := projList.Items[i].Name
		authedProjMap[projName] = struct{}{}
		authedProjs = append(authedProjs, projName)
	}
	blog.Infof("RequestID[%s] authed projects: %v", ctxutils.RequestID(ctx), authedProjs)
	result := make([]v1alpha1.Application, 0, len(argoAppList.Items))
	for i := range argoAppList.Items {
		if _, ok := authedProjMap[argoAppList.Items[i].Spec.Project]; ok {
			result = append(result, argoAppList.Items[i])
		}
	}
	argoAppList.Items = result
	return argoAppList, http.StatusOK, nil
}

// ListApplicationSets list applicationsets
func (h *handler) ListApplicationSets(ctx context.Context, query *appsetpkg.ApplicationSetListQuery) (
	*v1alpha1.ApplicationSetList, int, error) {
	projList, statusCode, err := h.listAuthorizedProjects(ctx, query.Projects)
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "list authorized projects failed")
	}
	authedProjMap := make(map[string]struct{})
	queryProjects := make([]string, 0)
	for i := range projList.Items {
		projName := projList.Items[i].Name
		authedProjMap[projName] = struct{}{}
		queryProjects = append(queryProjects, projName)
	}

	query.Projects = queryProjects
	appSetList, err := h.store.ListApplicationSets(ctx, query)
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "list applicationsets failed")
	}
	return appSetList, http.StatusOK, nil
}
