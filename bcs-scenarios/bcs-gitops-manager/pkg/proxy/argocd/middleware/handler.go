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
 *
 */

package middleware

import (
	"context"
	"net/http"
	"reflect"
	"runtime"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"k8s.io/kubernetes/pkg/kubelet/util/sliceutils"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/session"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// MiddlewareInterface defines the middleware interface
type MiddlewareInterface interface {
	Init() error

	HttpWrapper(handler httpHandler) http.Handler

	CheckMultiProjectsPermission(ctx context.Context, projectIDs []string,
		action []string) (map[string]map[string]bool, error)
	CheckProjectPermission(ctx context.Context, projectName string,
		action iam.ActionID) (*v1alpha1.AppProject, int, error)
	CheckProjectPermissionByID(ctx context.Context, projectID string, action iam.ActionID) (int, error)

	CheckMultiClustersPermission(ctx context.Context, projectID string, clusterIDs []string,
		actions []string) (map[string]map[string]bool, error)
	CheckClusterPermission(ctx context.Context, clusterName string, action iam.ActionID) (int, error)

	CheckRepositoryPermission(ctx context.Context, repoName string,
		action iam.ActionID) (*v1alpha1.Repository, int, error)
	CheckApplicationPermission(ctx context.Context, appName string,
		action iam.ActionID) (*v1alpha1.Application, int, error)

	ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, int, error)
	ListClusters(ctx context.Context, projectNames []string) (*v1alpha1.ClusterList, int, error)
	ListRepositories(ctx context.Context, projectNames []string,
		needCheckPermission bool) (*v1alpha1.RepositoryList, int, error)
	ListApplications(ctx context.Context, projectNames []string) (*v1alpha1.ApplicationList, error)
}

// handler 定义 http 中间件处理对象
type handler struct {
	projectPermission *project.BCSProjectPerm
	clusterPermission *cluster.BCSClusterPerm
	option            *proxy.GitOpsOptions
	argoSession       *session.ArgoSession
	secretSession     *session.SecretSession

	tracer func(context.Context) error
}

// NewMiddlewareHandler create handler instance
func NewMiddlewareHandler(option *proxy.GitOpsOptions, session *session.ArgoSession,
	secretSession *session.SecretSession) MiddlewareInterface {
	return &handler{
		option:            option,
		argoSession:       session,
		secretSession:     secretSession,
		projectPermission: project.NewBCSProjectPermClient(option.IAMClient),
		clusterPermission: cluster.NewBCSClusterPermClient(option.IAMClient),
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
	tracer, err := trace.InitTracingProvider("bcs-gitops-manager", opts...)
	if err != nil {
		return errors.Wrapf(err, "init tracer failed")
	}
	h.tracer = tracer
	return nil
}

// HttpWrapper 创建 http wrapper 中间件
func (h *handler) HttpWrapper(handler httpHandler) http.Handler {
	handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	hw := &httpWrapper{
		handler:       handler,
		handlerName:   handlerName,
		option:        h.option,
		argoSession:   h.argoSession,
		secretSession: h.secretSession,
	}
	blog.Infof("[Trace] request handler '%s' add to otel", handlerName)
	return otelhttp.NewHandler(hw, handlerName)
}

// CheckMultiProjectsPermission check multi projects with action
func (h *handler) CheckMultiProjectsPermission(ctx context.Context, projectIDs []string,
	actions []string) (map[string]map[string]bool, error) {
	user := ctx.Value(ctxKeyUser).(*proxy.UserInfo)
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
	// get project info and validate projectPermission
	argoProject, err := h.option.Storage.GetProject(ctx, projectName)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "get project from storage failure")
	}
	if argoProject == nil {
		return nil, http.StatusNotFound, errors.Errorf("project '%s' not found", projectName)
	}
	projectID := common.GetBCSProjectID(argoProject.Annotations)
	if projectID == "" {
		return nil, http.StatusForbidden,
			errors.Errorf("project '%s' got ID failed, not under control", projectName)
	}
	statusCode, err := h.CheckProjectPermissionByID(ctx, projectID, action)
	return argoProject, statusCode, err
}

// CheckProjectPermissionByID 检查登录态用户对于项目的权限
func (h *handler) CheckProjectPermissionByID(ctx context.Context, projectID string,
	action iam.ActionID) (int, error) {
	user := ctx.Value(ctxKeyUser).(*proxy.UserInfo)
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
func (h *handler) CheckClusterPermission(ctx context.Context, clusterName string,
	action iam.ActionID) (statusCode int, err error) {
	user := ctx.Value(ctxKeyUser).(*proxy.UserInfo)
	argoCluster, err := h.option.Storage.GetCluster(ctx, clusterName)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "get cluster from storage failure")
	}
	if argoCluster == nil {
		return http.StatusNotFound, errors.Errorf("cluster '%s' not found", clusterName)
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
		return http.StatusForbidden, errors.Errorf("cluster '%s' forbidden", clusterName)
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
	projectName := repo.Project
	_, statusCode, err := h.CheckProjectPermission(ctx, projectName, action)
	return repo, statusCode, err
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
	projectID := common.GetBCSProjectID(app.Annotations)
	if projectID != "" {
		statusCode, err := h.CheckProjectPermissionByID(ctx, projectID, action)
		if err != nil {
			return nil, statusCode, errors.Wrapf(err, "check project '%s' permission failed", projectID)
		}
		return app, http.StatusOK, nil
	}

	_, statusCode, err := h.CheckProjectPermission(ctx, app.Spec.Project, action)
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "check project '%s' permission failed", app.Spec.Project)
	}
	return app, http.StatusOK, nil
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
		if !sliceutils.StringInSlice(cls.Project, projectNames) {
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
	repositories, err := h.option.Storage.ListRepository(ctx)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "list repository from storage failed")
	}
	// filter specified project
	items := v1alpha1.Repositories{}
	for _, repo := range repositories.Items {
		if sliceutils.StringInSlice(repo.Project, projectNames) {
			items = append(items, repo)
		}
	}
	if len(items) == 0 {
		return &v1alpha1.RepositoryList{}, http.StatusOK, nil
	}
	repositories.Items = items
	return repositories, http.StatusOK, nil
}

// ListApplications 根据项目名称获取所有应用
func (h *handler) ListApplications(ctx context.Context, projectNames []string) (*v1alpha1.ApplicationList, error) {
	apps, err := h.option.Storage.ListApplications(ctx, &store.ListAppOptions{Projects: projectNames})
	if err != nil {
		return nil, errors.Wrapf(err, "list application swith project '%v' failed", projectNames)
	}
	return apps, nil
}
