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

package argocd

import (
	"context"
	"net/http"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/kubelet/util/sliceutils"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

type httpHandler func(ctx context.Context, r *http.Request) *httpResponse

type httpWrapper struct {
	handler httpHandler
	option  *proxy.GitOpsOptions
	session *Session
}

type httpResponse struct {
	isGrpc     bool
	obj        interface{}
	statusCode int
	err        error
}

const (
	ctxKeyRequestID = "requestID"
	ctxKeyUser      = "user"
)

// ServeHTTP 接收请求的入口，获取请求登录态信息并设置到 context 中
func (p *httpWrapper) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()
	// 统一获取 User 信息，并存入 context 中
	user, err := proxy.GetJWTInfo(r, p.option.JWTDecoder)
	if err != nil || user == nil {
		http.Error(rw, errors.Wrapf(err, "get user info failed").Error(), http.StatusUnauthorized)
		return
	}
	requestID := uuid.New().String()
	blog.Infof("[requestID=%s] manager received user '%s' serve [%s/%s]",
		requestID, user.GetUser(), r.Method, r.URL.Path)
	defer func() {
		blog.Infof("[requestID=%s] handle request cost time: %v", requestID, time.Since(start))
	}()

	ctx := context.WithValue(r.Context(), ctxKeyRequestID, requestID)
	ctx = context.WithValue(ctx, ctxKeyUser, user)
	resp := p.handler(ctx, r)
	blog.V(5).Infof("[requestID=%s] handler '%s' cost time: %v", requestID, p.handler, time.Since(start))
	if resp == nil {
		// 如果返回值为空，直接将请求 proxy 给 argo-cd
		p.session.ServeHTTP(rw, r)
		return
	}
	// 如果返回对象为空，直接返回错误给客户端
	if resp.obj == nil {
		blog.Errorf("[requestID=%s] handler return code '%d': %s",
			requestID, resp.statusCode, resp.err.Error())
		http.Error(rw, resp.err.Error(), resp.statusCode)
		return
	}
	if resp.isGrpc {
		proxy.GRPCResponse(rw, resp.obj)
	} else {
		proxy.JSONResponse(rw, resp.obj)
	}
}

// MiddlewareInterface defines the middleware interface
type MiddlewareInterface interface {
	HttpWrapper(handler httpHandler) *httpWrapper

	ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, int, error)
	CheckProjectPermission(ctx context.Context, projectName string,
		action iam.ActionID) (*v1alpha1.AppProject, int, error)
	CheckProjectPermissionByID(ctx context.Context, projectID string, action iam.ActionID) (int, error)

	ListClusters(ctx context.Context, projectNames []string) (*v1alpha1.ClusterList, int, error)
	CheckClusterCreatePermission(ctx context.Context, projectID string) (int, error)
	CheckClusterPermission(ctx context.Context, clusterName string, action iam.ActionID) (int, error)

	ListRepositories(ctx context.Context, projectNames []string,
		needCheckPermission bool) (*v1alpha1.RepositoryList, int, error)
	CheckRepositoryPermission(ctx context.Context, repoName string,
		action iam.ActionID) (*v1alpha1.Repository, int, error)

	ListApplications(ctx context.Context, projectNames []string) (*v1alpha1.ApplicationList, error)
	CheckApplicationPermission(ctx context.Context, appName string,
		action iam.ActionID) (*v1alpha1.Application, int, error)
}

// MiddlewareHandler 定义 http 中间件处理对象
type MiddlewareHandler struct {
	session           *Session
	projectPermission *project.BCSProjectPerm
	clusterPermission *cluster.BCSClusterPerm
	option            *proxy.GitOpsOptions
}

// NewMiddlewareHandler create MiddlewareHandler instance
func NewMiddlewareHandler(option *proxy.GitOpsOptions, session *Session) MiddlewareInterface {
	return &MiddlewareHandler{
		session:           session,
		option:            option,
		projectPermission: project.NewBCSProjectPermClient(option.IAMClient),
		clusterPermission: cluster.NewBCSClusterPermClient(option.IAMClient),
	}
}

// HttpWrapper 创建 http wrapper 中间件
func (h *MiddlewareHandler) HttpWrapper(handler httpHandler) *httpWrapper {
	return &httpWrapper{
		handler: handler,
		option:  h.option,
		session: h.session,
	}
}

// ListProjects 根据用户权限列出具备权限的 Projects
func (h *MiddlewareHandler) ListProjects(ctx context.Context) (*v1alpha1.AppProjectList, int, error) {
	user := ctx.Value("user").(*proxy.UserInfo)
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
	result, err := h.projectPermission.GetMultiProjectMultiActionPermission(
		user.GetUser(), projectIDs, []string{action})
	if err != nil {
		return nil, http.StatusInternalServerError,
			errors.Wrapf(err, "project permission auth center failed")
	}
	// projectPermission check
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

// CheckProjectPermission 检查登录态用户对于项目的权限
func (h *MiddlewareHandler) CheckProjectPermission(ctx context.Context, projectName string,
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
func (h *MiddlewareHandler) CheckProjectPermissionByID(ctx context.Context, projectID string,
	action iam.ActionID) (int, error) {
	user := ctx.Value("user").(*proxy.UserInfo)
	var permit bool
	var err error
	switch action {
	case iam.ProjectView:
		permit, _, err = h.projectPermission.CanViewProject(user.GetUser(), projectID)
	case iam.ProjectEdit:
		permit, _, err = h.projectPermission.CanEditProject(user.GetUser(), projectID)
	case iam.ProjectDelete:
		permit, _, err = h.projectPermission.CanDeleteProject(user.GetUser(), projectID)
	case iam.ProjectCreate:
		permit, _, err = h.projectPermission.CanCreateProject(user.GetUser())
	default:
		return http.StatusBadRequest, errors.Errorf("unknown iam action '%s'", action)
	}
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "auth center failed")
	}
	if !permit {
		return http.StatusForbidden, errors.Errorf("project '%s' forbidden", projectID)
	}
	return http.StatusOK, nil
}

// CheckClusterCreatePermission 检查登录态用户是否有创建集群权限
func (h *MiddlewareHandler) CheckClusterCreatePermission(ctx context.Context, projectID string) (int, error) {
	user := ctx.Value("user").(*proxy.UserInfo)
	permit, _, err := h.clusterPermission.CanCreateCluster(user.GetUser(), projectID)
	if err != nil {
		return http.StatusInternalServerError, errors.Errorf("auth center failed")
	}
	if !permit {
		return http.StatusForbidden, errors.Errorf("create cluster for project '%s' forbidden", projectID)
	}
	return http.StatusOK, nil
}

// CheckClusterPermission 检查登录态用户对于集群的权限
func (h *MiddlewareHandler) CheckClusterPermission(ctx context.Context, clusterName string,
	action iam.ActionID) (statusCode int, err error) {
	user := ctx.Value("user").(*proxy.UserInfo)
	argoCluster, err := h.option.Storage.GetCluster(ctx, clusterName)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "get cluster from storage failure")
	}
	if argoCluster == nil {
		return http.StatusNotFound, errors.Errorf("cluster '%s' not found", clusterName)
	}
	projectID := common.GetBCSProjectID(argoCluster.Annotations)
	if projectID == "" {
		return http.StatusForbidden, errors.Errorf("cluster no project control infomation")
	}

	var permit bool
	switch action {
	case iam.ClusterView:
		permit, _, err = h.clusterPermission.CanViewCluster(user.GetUser(), projectID, argoCluster.Name)
	case iam.ClusterManage:
		permit, _, err = h.clusterPermission.CanManageCluster(user.GetUser(), projectID, argoCluster.Name)
	case iam.ClusterDelete:
		permit, _, err = h.clusterPermission.CanDeleteCluster(user.GetUser(), projectID, argoCluster.Name)
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

// ListClusters 根据项目名获取用户态下可以 view 的集群列表
func (h *MiddlewareHandler) ListClusters(ctx context.Context, projectNames []string) (*v1alpha1.ClusterList,
	int, error) {
	user := ctx.Value("user").(*proxy.UserInfo)
	// get all clusters from argocd
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
		result, err := h.clusterPermission.GetMultiClusterMultiActionPermission(user.GetUser(), projectID,
			clusterIDs, []string{action})
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
func (h *MiddlewareHandler) ListRepositories(ctx context.Context, projectNames []string,
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

// CheckRepositoryPermission 检查登录态用户对于 Repo 仓库权限，Repo 权限与 Project 权限挂钩
func (h *MiddlewareHandler) CheckRepositoryPermission(ctx context.Context, repoName string,
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

// ListApplications 根据项目名称获取所有应用
func (h *MiddlewareHandler) ListApplications(ctx context.Context,
	projectNames []string) (*v1alpha1.ApplicationList, error) {
	appList := make([]v1alpha1.Application, 0)
	for _, name := range projectNames {
		apps, err := h.option.Storage.ListApplications(ctx, &store.ListAppOptions{Project: name})
		if err != nil {
			return nil, errors.Wrapf(err, "list applications with project '%s' failed", name)
		}
		appList = append(appList, apps.Items...)
	}
	return &v1alpha1.ApplicationList{
		Items: appList,
	}, nil
}

// CheckApplicationPermission 检查应用的权限
func (h *MiddlewareHandler) CheckApplicationPermission(ctx context.Context, appName string,
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
