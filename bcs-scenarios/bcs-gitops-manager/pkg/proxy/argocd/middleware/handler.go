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
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/project"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	argocluster "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	clusterclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/exp/slices"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/analysis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/session"
)

// MiddlewareInterface defines the middleware interface
// nolint
type MiddlewareInterface interface {
	Init() error

	HttpWrapper(handler HttpHandler) http.Handler

	CheckMultiProjectsPermission(ctx context.Context, projectIDs []string,
		action []string) (map[string]map[string]bool, error)
	CheckProjectPermission(ctx context.Context, projectName string,
		action iam.ActionID) (*v1alpha1.AppProject, int, error)
	CheckProjectPermissionByID(ctx context.Context, projectName, projectID string, action iam.ActionID) (int, error)
	CheckBusinessPermission(ctx context.Context, bizID string, action iam.ActionID) (int, error)

	CheckMultiClustersPermission(ctx context.Context, projectID string, clusterIDs []string,
		actions []string) (map[string]map[string]bool, error)
	CheckClusterPermission(ctx context.Context, query *clusterclient.ClusterQuery, action iam.ActionID) (int, error)

	CheckCreateApplication(ctx context.Context, app *v1alpha1.Application) (int, error)
	CheckRepositoryPermission(ctx context.Context, repoName string,
		action iam.ActionID) (*v1alpha1.Repository, int, error)
	CheckApplicationPermission(ctx context.Context, appName string,
		action iam.ActionID) (*v1alpha1.Application, int, error)

	ListProjectsWithoutAuth(ctx context.Context) (*v1alpha1.AppProjectList, int, error)
	ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, int, error)
	ListClusters(ctx context.Context, projectNames []string) (*v1alpha1.ClusterList, int, error)
	ListRepositories(ctx context.Context, projectNames []string,
		needCheckPermission bool) (*v1alpha1.RepositoryList, int, error)
	ListApplications(ctx context.Context, query *appclient.ApplicationQuery) (*v1alpha1.ApplicationList, error)

	CheckCreateApplicationSet(ctx context.Context,
		appset *v1alpha1.ApplicationSet) ([]*v1alpha1.Application, int, error)
	CheckDeleteApplicationSet(ctx context.Context, appsetName string) (*v1alpha1.ApplicationSet, int, error)
	CheckGetApplicationSet(ctx context.Context, appsetName string) (int, error)
	ListApplicationSets(ctx context.Context, query *appsetpkg.ApplicationSetListQuery) (
		*v1alpha1.ApplicationSetList, error)
}

// handler 定义 http 中间件处理对象
type handler struct {
	projectPermission   *project.BCSProjectPerm
	clusterPermission   *cluster.BCSClusterPerm
	namespacePermission *namespace.BCSNamespacePerm
	option              *proxy.GitOpsOptions
	argoSession         *session.ArgoSession
	secretSession       *session.SecretSession
	analysisClient      analysis.AnalysisInterface
	monitorSession      *session.MonitorSession

	tracer func(context.Context) error
}

// NewMiddlewareHandler create handler instance
func NewMiddlewareHandler(option *proxy.GitOpsOptions, session *session.ArgoSession,
	secretSession *session.SecretSession,
	monitorSession *session.MonitorSession) MiddlewareInterface {
	return &handler{
		option:              option,
		argoSession:         session,
		secretSession:       secretSession,
		monitorSession:      monitorSession,
		projectPermission:   project.NewBCSProjectPermClient(option.IAMClient),
		clusterPermission:   cluster.NewBCSClusterPermClient(option.IAMClient),
		namespacePermission: namespace.NewBCSNamespacePermClient(option.IAMClient),
		analysisClient:      analysis.GetAnalysisClient(),
	}
}

// Init will init the tracer
func (h *handler) Init() error {
	opts := []trace.Option{
		trace.OTLPEndpoint(h.option.TraceOption.Endpoint),
	}
	attrs := make([]attribute.KeyValue, 0)
	attrs = append(attrs, attribute.String("bk.data.token", h.option.TraceOption.Token))
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
		handler:        handler,
		handlerName:    handlerName,
		option:         h.option,
		argoSession:    h.argoSession,
		secretSession:  h.secretSession,
		monitorSession: h.monitorSession,
	}
	blog.Infof("[Trace] request handler '%s' add to otel", handlerName)
	return otelhttp.NewHandler(hw, handlerName)
}

// CheckMultiProjectsPermission check multi projects with action
func (h *handler) CheckMultiProjectsPermission(ctx context.Context, projectIDs []string,
	actions []string) (map[string]map[string]bool, error) {
	user := ctx.Value(ctxKeyUser).(*proxy.UserInfo)
	// GetMultiProjectMultiActionPerm only support same instanceSelection
	result, err := h.projectPermission.GetMultiProjectMultiActionPerm(
		user.GetUser(), projectIDs, actions)
	if err != nil {
		return nil, errors.Wrapf(err, "check multi project action failed")
	}
	return result, nil
}

// CheckMultiClustersPermission check multi clusters with action
func (h *handler) CheckMultiClustersPermission(ctx context.Context, projectID string,
	clusterIDs []string, actions []string) (map[string]map[string]bool, error) {
	user := ctx.Value(ctxKeyUser).(*proxy.UserInfo)
	// GetMultiClusterMultiActionPerm only support same instanceSelection
	result, err := h.clusterPermission.GetMultiClusterMultiActionPerm(user.GetUser(), projectID,
		clusterIDs, actions)
	if err != nil {
		return nil, errors.Wrapf(err, "check multi clusters action failed")
	}
	return result, nil
}

// CheckProjectPermission 检查登录态用户对于项目的权限
func (h *handler) CheckProjectPermission(ctx context.Context, projectName string,
	action iam.ActionID) (*v1alpha1.AppProject, int, error) {
	if projectName == "" {
		return nil, http.StatusBadRequest, errors.Errorf("project name cannot be empty")
	}
	// get project info and validate projectPermission
	argoProject, err := h.option.Storage.GetProject(ctx, projectName)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "get project from storage failure")
	}
	if argoProject == nil {
		return nil, http.StatusNotFound, errors.Errorf("project '%s' not found", projectName)
	}
	// GetBCSProjectID get projectID from annotations
	projectID := common.GetBCSProjectID(argoProject.Annotations)
	if projectID == "" {
		return nil, http.StatusForbidden,
			errors.Errorf("project '%s' got ID failed, not under control", projectName)
	}
	var statusCode int
	// CheckProjectPermissionByID 检查登录态用户对于项目的权限
	statusCode, err = h.CheckProjectPermissionByID(ctx, projectName, projectID, action)
	return argoProject, statusCode, err
}

// CheckBusinessPermission 检查用户是否具备业务权限
func (h *handler) CheckBusinessPermission(ctx context.Context, bizID string, action iam.ActionID) (int, error) {
	if bizID == "" {
		return http.StatusBadRequest, errors.Errorf("bizID cannot be empty")
	}
	projectList, statusCode, err := h.ListProjects(ctx)
	if statusCode != http.StatusOK {
		return statusCode, err
	}

	for _, proj := range projectList.Items {
		// GetBCSProjectBusinessKey return the business id of project
		projectBizID := common.GetBCSProjectBusinessKey(proj.Annotations)
		if projectBizID == bizID {
			// CheckProjectPermissionByID 检查登录态用户对于项目的权限
			statusCode, _ = h.CheckProjectPermissionByID(ctx, proj.Name,
				common.GetBCSProjectID(proj.Annotations), action)
			// 只要拥有一个project的权限，则允许操作
			if statusCode == http.StatusOK {
				return http.StatusOK, nil
			}
		}
	}
	return http.StatusForbidden, errors.Errorf("businessID '%s' for action '%s' forbidden", bizID, action)
}

// CheckNamespaceScopedResourcePermission 检查 NamespaceScopedResource 的权限
// nolint
func (h *handler) CheckNamespaceScopedResourcePermission(ctx context.Context, projectName, projectID, clusterID,
	namespace string, action iam.ActionID) (int, error) {
	user := ctx.Value(ctxKeyUser).(*proxy.UserInfo)
	if h.isAdminUser(user.GetUser()) {
		return http.StatusOK, nil
	}
	var permit bool
	var err error
	switch action {
	case iamnamespace.NameSpaceScopedView:
		permit, _, _, err = h.namespacePermission.
			CanViewNamespaceScopedResource(user.GetUser(), projectID, clusterID, namespace)
	case iamnamespace.NameSpaceScopedCreate:
		permit, _, _, err = h.namespacePermission.
			CanCreateNamespaceScopedResource(user.GetUser(), projectID, clusterID, namespace)
	case iamnamespace.NameSpaceScopedUpdate:
		permit, _, _, err = h.namespacePermission.
			CanUpdateNamespaceScopedResource(user.GetUser(), projectID, clusterID, namespace)
	case iamnamespace.NameSpaceScopedDelete:
		permit, _, _, err = h.namespacePermission.
			CanDeleteNamespaceScopedResource(user.GetUser(), projectID, clusterID, namespace)
	default:
		return http.StatusInternalServerError, errors.Errorf("unknown iam action '%s'", action)
	}
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "auth center failed")
	}
	if !permit {
		return http.StatusForbidden, errors.Errorf("project '%s' namespace '%s/%s' for action '%s' forbidden",
			projectName, clusterID, namespace, action)
	}
	return http.StatusOK, nil
}

// CheckProjectPermissionByID 检查登录态用户对于项目的权限
func (h *handler) CheckProjectPermissionByID(ctx context.Context, projectName, projectID string,
	action iam.ActionID) (int, error) {
	user := ctx.Value(ctxKeyUser).(*proxy.UserInfo)
	h.analysisClient.UpdateActivityUser(projectName, user.GetUser())
	if h.isAdminUser(user.GetUser()) {
		return http.StatusOK, nil
	}
	var permit bool
	var err error
	switch action {
	case iam.ProjectView:
		permit, _, _, err = h.projectPermission.CanViewProject(user.GetUser(), projectID)
	case iam.ProjectEdit:
		permit, _, _, err = h.projectPermission.CanEditProject(user.GetUser(), projectID)
	case iam.ProjectDelete:
		permit, _, _, err = h.projectPermission.CanDeleteProject(user.GetUser(), projectID)
	case iam.ProjectCreate:
		permit, _, _, err = h.projectPermission.CanCreateProject(user.GetUser())
	default:
		return http.StatusBadRequest, errors.Errorf("unknown iam action '%s'", action)
	}
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "auth center failed")
	}
	if !permit {
		return http.StatusForbidden, errors.Errorf("project '%s' for action '%s' forbidden", projectID, action)
	}
	return http.StatusOK, nil
}

// CheckClusterPermission 检查登录态用户对于集群的权限
func (h *handler) CheckClusterPermission(ctx context.Context, query *clusterclient.ClusterQuery,
	action iam.ActionID) (statusCode int, err error) {
	user := ctx.Value(ctxKeyUser).(*proxy.UserInfo)
	if h.isAdminUser(user.GetUser()) {
		return http.StatusOK, nil
	}
	argoCluster, err := h.option.Storage.GetCluster(ctx, query)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "get cluster from storage failure")
	}
	if argoCluster == nil {
		return http.StatusNotFound, errors.Errorf("cluster '%v' not found", *query)
	}
	projectID := common.GetBCSProjectID(argoCluster.Annotations)
	if projectID == "" {
		return http.StatusForbidden, errors.Errorf("cluster no project control information")
	}

	var permit bool
	switch action {
	case iam.ClusterView:
		permit, _, _, err = h.clusterPermission.CanViewCluster(user.GetUser(), projectID, argoCluster.Name)
	case iam.ClusterManage:
		permit, _, _, err = h.clusterPermission.CanManageCluster(user.GetUser(), projectID, argoCluster.Name)
	case iam.ClusterDelete:
		permit, _, _, err = h.clusterPermission.CanDeleteCluster(user.GetUser(), projectID, argoCluster.Name)
	default:
		return http.StatusBadRequest, errors.Errorf("unknown iam action '%s'", action)
	}
	if err != nil {
		return http.StatusInternalServerError, errors.Errorf("auth center failed")
	}
	if !permit {
		return http.StatusForbidden, errors.Errorf("cluster '%v' forbidden", *query)
	}
	return http.StatusOK, nil
}

// CheckRepositoryPermission 检查登录态用户对于 Repo 仓库权限，Repo 权限与 Project 权限挂钩
func (h *handler) CheckRepositoryPermission(ctx context.Context, repoName string,
	action iam.ActionID) (*v1alpha1.Repository, int, error) {
	repo, err := h.option.Storage.GetRepository(ctx, repoName)
	if err != nil {
		return nil, http.StatusInternalServerError,
			errors.Wrapf(err, "get repository '%s' from storage failed", repoName)
	}
	if repo == nil {
		return nil, http.StatusNotFound, errors.Errorf("repository '%s' not found", repoName)
	}
	if slices.Contains[string](h.option.PublicProjects, repo.Project) {
		return repo, http.StatusOK, nil
	}
	projectName := repo.Project
	_, statusCode, err := h.CheckProjectPermission(ctx, projectName, action)
	return repo, statusCode, err
}

func (h *handler) checkRepositoryBelongProject(ctx context.Context, repoUrl, project string) (bool, error) {
	repo, err := h.option.Storage.GetRepository(ctx, repoUrl)
	if err != nil {
		return false, errors.Wrapf(err, "get repo '%s' failed", repoUrl)
	}
	if repo == nil {
		return false, fmt.Errorf("repo '%s' not found", repoUrl)
	}
	// passthrough if repository's project equal to public projects
	if slices.Contains[string](h.option.PublicProjects, repo.Project) {
		return true, nil
	}

	if repo.Project != project {
		return false, nil
	}
	return true, nil
}

// CheckApplicationPermission 检查应用的权限
func (h *handler) CheckApplicationPermission(ctx context.Context, appName string,
	action iam.ActionID) (*v1alpha1.Application, int, error) {
	app, err := h.option.Storage.GetApplication(ctx, appName)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err,
			"get application '%s' from storage failed", appName)
	}
	if app == nil {
		return nil, http.StatusNotFound, errors.Errorf("application '%s' not found", appName)
	}
	// 检查是否具备 ProjectView 权限
	projectID := common.GetBCSProjectID(app.Annotations)
	if projectID != "" {
		// CheckProjectPermissionByID 检查登录态用户对于项目的权限
		statusCode, err := h.CheckProjectPermissionByID(ctx, app.Spec.Project, projectID, iam.ProjectView) // nolint
		if err != nil {
			return nil, statusCode, errors.Wrapf(err, "check project '%s' permission failed", projectID)
		}
		return app, http.StatusOK, nil
	}
	argoProject, statusCode, err := h.CheckProjectPermission(ctx, app.Spec.Project, iam.ProjectView)
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "check project '%s' permission failed", app.Spec.Project)
	}
	// GetBCSProjectID get projectID from annotations
	projectID = common.GetBCSProjectID(argoProject.Annotations)

	// 获取集群信息
	argoCluster, err := h.option.Storage.GetCluster(ctx, &argocluster.ClusterQuery{
		Server: app.Spec.Destination.Server,
	})
	if err != nil {
		return nil, http.StatusInternalServerError,
			errors.Wrapf(err, "get cluster '%s' failed", app.Spec.Destination.Server)
	}
	if argoCluster == nil {
		return nil, http.StatusNotFound, fmt.Errorf("cluster '%s' not found", app.Spec.Destination.Server)
	}
	// 检查是否具备 NamespaceScopedResource 权限
	if app.Spec.Destination.Namespace == "" {
		app.Spec.Destination.Namespace = "default" // nolint
	}
	statusCode, err = h.CheckNamespaceScopedResourcePermission(ctx, app.Spec.Project, projectID, argoCluster.Name,
		app.Spec.Destination.Namespace, action)
	if err != nil {
		return nil, statusCode,
			errors.Wrapf(err, "check namespaceScoped resource permission for action '%s' failed", action)
	}
	return app, http.StatusOK, nil
}

// CheckCreateApplication 检查创建某个应用是否具备权限
func (h *handler) CheckCreateApplication(ctx context.Context, app *v1alpha1.Application) (int, error) {
	projectName := app.Spec.Project
	if projectName == "" || projectName == "default" { // nolint
		return http.StatusBadRequest, errors.Errorf("project information lost")
	}
	// CheckProjectPermission 检查登录态用户对于项目的权限
	argoProject, statusCode, err := h.CheckProjectPermission(ctx, projectName, iam.ProjectView)
	if statusCode != http.StatusOK {
		return statusCode, errors.Wrapf(err, "check application '%s' permission failed", projectName)
	}

	if app.Spec.HasMultipleSources() {
		for i := range app.Spec.Sources {
			appSource := app.Spec.Sources[i]
			repoUrl := appSource.RepoURL
			var repoBelong bool
			repoBelong, err = h.checkRepositoryBelongProject(ctx, repoUrl, projectName)
			if err != nil {
				return http.StatusBadRequest,
					errors.Wrapf(err, "check multi-source repository '%s' permission failed", repoUrl)
			}
			if !repoBelong {
				return http.StatusForbidden,
					errors.Errorf("check multi-source repo '%s' not belong to project '%s'", repoUrl, projectName)
			}
			blog.Infof("RequestID[%s] check multi-source repo '%s' success", RequestID(ctx), repoUrl)
		}
	} else if app.Spec.Source != nil {
		repoUrl := app.Spec.Source.RepoURL
		var repoBelong bool
		repoBelong, err = h.checkRepositoryBelongProject(ctx, repoUrl, projectName)
		if err != nil {
			return http.StatusBadRequest, errors.Wrapf(err, "check repository permission failed")
		}
		if !repoBelong {
			return http.StatusForbidden, errors.Errorf("repo '%s' not belong to project '%s'",
				repoUrl, projectName)
		}
		blog.Infof("RequestID[%s] check source repo '%s' success", RequestID(ctx), repoUrl)
	}

	clusterQuery := clusterclient.ClusterQuery{
		Server: app.Spec.Destination.Server,
		Name:   app.Spec.Destination.Name,
	}
	argoCluster, err := h.option.Storage.GetCluster(ctx, &clusterQuery)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "get cluster '%v' failed", clusterQuery)
	}
	if argoCluster == nil {
		return http.StatusNotFound, fmt.Errorf("cluster '%v' not found", clusterQuery)
	}
	// 检查是否具备 NamespaceScopedResource 权限
	if app.Spec.Destination.Namespace == "" {
		app.Spec.Destination.Namespace = "default" // nolint
	}
	// GetBCSProjectID get projectID from annotations
	projectID := common.GetBCSProjectID(argoProject.Annotations)
	statusCode, err = h.CheckNamespaceScopedResourcePermission(ctx, app.Spec.Project, projectID, argoCluster.Name,
		app.Spec.Destination.Namespace, iamnamespace.NameSpaceScopedCreate)
	if err != nil {
		return statusCode,
			errors.Wrapf(err, "check namespaceScoped resource permission for action '%s' failed",
				iamnamespace.NameSpaceScopedCreate)
	}

	// setting application name with project prefix
	if !strings.HasPrefix(app.Name, projectName+"-") {
		app.Name = projectName + "-" + app.Name
	}
	// setting control annotations
	if app.Annotations == nil {
		app.Annotations = make(map[string]string)
	}
	common.AddCustomAnnotationForApplication(argoProject, app)
	return 0, nil
}

// ListProjectsWithoutAuth list all projects that argo controlled
func (h *handler) ListProjectsWithoutAuth(ctx context.Context) (*v1alpha1.AppProjectList, int, error) {
	projectList, err := h.option.Storage.ListProjects(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "list projects failed")
	}
	result := make([]v1alpha1.AppProject, 0, len(projectList.Items))
	for i := range projectList.Items {
		appProj := projectList.Items[i]
		projectID := common.GetBCSProjectID(appProj.Annotations)
		if projectID == "" {
			continue
		}
		result = append(result, appProj)
	}
	projectList.Items = result
	return projectList, http.StatusOK, nil
}

// ListProjects 根据用户权限列出具备权限的 Projects
func (h *handler) ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, int, error) {
	projectList, err := h.option.Storage.ListProjects(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "list projects failed")
	}
	projectIDs := make([]string, 0, len(projectList.Items))
	controlledProjects := make(map[string]v1alpha1.AppProject)
	for i, app := range projectList.Items {
		projectID := common.GetBCSProjectID(app.Annotations)
		if projectID == "" {
			continue
		}
		controlledProjects[projectID] = projectList.Items[i]
		projectIDs = append(projectIDs, projectID)
	}
	if len(projectIDs) == 0 {
		return &v1alpha1.AppProjectList{}, http.StatusOK, nil
	}
	action := string(project.ProjectView)
	result, err := h.CheckMultiProjectsPermission(ctx, projectIDs, []string{action})
	if err != nil {
		return nil, http.StatusInternalServerError,
			errors.Wrapf(err, "project permission auth center failed")
	}
	finals := make([]v1alpha1.AppProject, 0, len(result))
	for projectID, permits := range result {
		if permits[action] {
			appProject := controlledProjects[projectID]
			finals = append(finals, appProject)
		}
	}
	projectList.Items = finals
	return projectList, http.StatusOK, nil
}

// ListClusters 根据项目名获取用户态下可以 view 的集群列表
func (h *handler) ListClusters(ctx context.Context, projectNames []string) (
	*v1alpha1.ClusterList, int, error) {
	clusterList, err := h.option.Storage.ListCluster(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "list clusters from storage failure")
	}

	projectClusters := make(map[string][]string)
	controlledClusters := make(map[string]v1alpha1.Cluster)
	for _, cls := range clusterList.Items {
		if !slices.Contains[string](projectNames, cls.Project) {
			continue
		}
		controlProjectID := common.GetBCSProjectID(cls.Annotations)
		if controlProjectID == "" {
			continue
		}
		projectClusters[controlProjectID] = append(projectClusters[controlProjectID], cls.Name)
		controlledClusters[cls.Name] = cls
	}
	action := string(cluster.ClusterView)
	resultClusterList := &v1alpha1.ClusterList{}
	for projectID, clusterIDs := range projectClusters {
		result, err := h.CheckMultiClustersPermission(ctx, projectID, clusterIDs, []string{action})
		if err != nil {
			return nil, http.StatusInternalServerError,
				errors.Wrapf(err, "check cluster permission occurred internal error")
		}
		for clusterName, permit := range result {
			if permit[action] {
				resultClusterList.Items = append(resultClusterList.Items, controlledClusters[clusterName])
			}
		}
	}
	return resultClusterList, http.StatusOK, nil
}

// ListRepositories 根据项目名称获取用户态下可以 view 的仓库列表
func (h *handler) ListRepositories(ctx context.Context, projectNames []string,
	needCheckPermission bool) (*v1alpha1.RepositoryList, int, error) {
	if needCheckPermission {
		for _, name := range projectNames {
			_, statusCode, err := h.CheckProjectPermission(ctx, name, iam.ProjectView)
			if statusCode != http.StatusOK {
				return nil, statusCode, err
			}
		}
	}

	// projectPermission pass, list all repositories in gitops storage
	repositories, err := h.option.Storage.ListRepository(ctx, projectNames)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "list repository from storage failed")
	}
	return repositories, http.StatusOK, nil
}

// ListApplications 根据项目名称获取所有应用
func (h *handler) ListApplications(ctx context.Context, query *appclient.ApplicationQuery) (
	*v1alpha1.ApplicationList, error) {
	apps, err := h.option.Storage.ListApplications(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "list application swith project '%v' failed", query.Projects)
	}
	projectAppsMap := make(map[string][]v1alpha1.Application)
	for i := range apps.Items {
		item := apps.Items[i]
		_, ok := projectAppsMap[item.Spec.Project]
		if ok {
			projectAppsMap[item.Spec.Project] = append(projectAppsMap[item.Spec.Project], item)
		} else {
			projectAppsMap[item.Spec.Project] = []v1alpha1.Application{item}
		}
	}
	result := make([]v1alpha1.Application, 0, len(apps.Items))
	for proj, projApps := range projectAppsMap {
		var prefers []*dao.ResourcePreference
		prefers, err = h.analysisClient.ListApplicationCollects(proj)
		if err != nil {
			return nil, errors.Wrapf(err, "list application collects for project '%s' failed", proj)
		}
		preferMap := make(map[string]struct{})
		for _, prefer := range prefers {
			preferMap[prefer.Name] = struct{}{}
		}

		for i := range projApps {
			if _, ok := preferMap[projApps[i].Name]; !ok {
				continue
			}
			if projApps[i].Annotations == nil {
				projApps[i].Annotations = map[string]string{
					common.ApplicationCollectAnnotation: "true",
				}
			} else {
				projApps[i].Annotations[common.ApplicationCollectAnnotation] = "true"
			}
		}
		sort.Sort(applicationSort(projApps))
		result = append(result, projApps...)
	}
	apps.Items = result
	return apps, nil
}

func (h *handler) isAdminUser(user string) bool {
	return slices.Contains[string](h.option.AdminUsers, user)
}
