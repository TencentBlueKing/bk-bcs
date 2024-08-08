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

// Package permitcheck xx
package permitcheck

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	iamcluster "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/cluster"
	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
	iamproject "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/project"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	argocluster "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

// RSAction defines the action of resource
type RSAction string

// RSType defines the type of resource
type RSType string

// nolint
var (
	ProjectViewRSAction   RSAction = "project_view"
	ProjectEditRSAction   RSAction = "project_edit"
	ProjectDeleteRSAction RSAction = "project_delete"
	ClusterViewRSAction   RSAction = "cluster_view"
	RepoCreateRSAction    RSAction = "repo_create"
	RepoViewRSAction      RSAction = "repo_view"
	RepoDeleteRSAction    RSAction = "repo_delete"
	RepoUpdateRSAction    RSAction = "repo_update"
	AppViewRSAction       RSAction = "app_view"
	AppUpdateRSAction     RSAction = "app_update"
	AppDeleteRSAction     RSAction = "app_delete"
	AppCreateRSAction     RSAction = "app_create"
	AppSetViewRSAction    RSAction = "appset_view"
	AppSetCreateRSAction  RSAction = "appset_create"
	AppSetUpdateRSAction  RSAction = "appset_update"
	AppSetDeleteRSAction  RSAction = "appset_delete"
	SecretViewRSAction    RSAction = "secret_view"
	SecretOperateSAction  RSAction = "secret_operate"

	ProjectRSType RSType = "project"
	ClusterRSType RSType = "cluster"
	RepoRSType    RSType = "repo"
	AppRSType     RSType = "app"
	AppSetRSType  RSType = "appset"
)

// ResourceUserPermission 某资源对应的具备操作权限的用户
type ResourceUserPermission struct {
	ResourceType RSType                     `json:"resourceType"`
	ResourceName string                     `json:"resourceName"`
	UserPerms    map[string]map[string]bool `json:"userPerms"`
}

// UserResourcePermission 某用户具备的某个类型资源下的资源操作权限
type UserResourcePermission struct {
	ResourceType  RSType                       `json:"resourceType"`
	ActionPerms   map[RSAction]bool            `json:"actionPerms"`
	ResourcePerms map[string]map[RSAction]bool `json:"resourcePerms"`
}

// UpdatePermissionRequest defines the request that update permissions
type UpdatePermissionRequest struct {
	Users           []string `json:"users"`
	ResourceNames   []string `json:"resourceNames"`
	ResourceActions []string `json:"resourceActions"`
}

// PermissionInterface defines the interface of permission
type PermissionInterface interface {
	CheckProjectPermission(ctx context.Context, project string, action RSAction) (*v1alpha1.AppProject, int, error)
	GetProjectMultiPermission(ctx context.Context, projectIDs []string, actions []RSAction) (
		map[string]map[RSAction]bool, error)

	CheckClusterPermission(ctx context.Context, query *cluster.ClusterQuery, action RSAction) (*v1alpha1.Cluster,
		int, error)
	CheckRepoPermission(ctx context.Context, repo string, action RSAction) (*v1alpha1.Repository, int, error)
	CheckRepoCreate(ctx context.Context, repo *v1alpha1.Repository) (int, error)
	CheckApplicationPermission(ctx context.Context, app string, action RSAction) (*v1alpha1.Application, int, error)
	CheckApplicationCreate(ctx context.Context, app *v1alpha1.Application) (int, error)
	CheckAppSetPermission(ctx context.Context, appSet string, action RSAction) (*v1alpha1.ApplicationSet, int, error)
	CheckAppSetCreate(ctx context.Context, appSet *v1alpha1.ApplicationSet) ([]*v1alpha1.Application, int, error)

	CheckBCSPermissions(req *http.Request) (bool, int, error)
	CheckBCSClusterPermission(ctx context.Context, user, clusterID string, action iam.ActionID) (int, error)

	UpdatePermissions(ctx context.Context, project string, resourceType RSType,
		req *UpdatePermissionRequest) (int, error)
	UserAllPermissions(ctx context.Context, project string) ([]*UserResourcePermission, int, error)
	QueryUserPermissions(ctx context.Context, project string, rsType RSType,
		rsNames []string) ([]interface{}, *UserResourcePermission, int, error)
	QueryResourceUsers(ctx context.Context, project string, rsType RSType,
		resources []string) (map[string]*ResourceUserPermission, int, error)
}

type checker struct {
	projectPermission   *iamproject.BCSProjectPerm
	clusterPermission   *iamcluster.BCSClusterPerm
	namespacePermission *iamnamespace.BCSNamespacePerm

	option *options.Options
	store  store.Store
	db     dao.Interface
}

// NewPermitChecker create permit checker instance
func NewPermitChecker() PermissionInterface {
	op := options.GlobalOptions()
	return &checker{
		option:              op,
		db:                  dao.GlobalDB(),
		store:               store.GlobalStore(),
		projectPermission:   iamproject.NewBCSProjectPermClient(op.IAMClient),
		clusterPermission:   iamcluster.NewBCSClusterPermClient(op.IAMClient),
		namespacePermission: iamnamespace.NewBCSNamespacePermClient(op.IAMClient),
	}
}

func (c *checker) isAdminUser(user string) bool {
	return slices.Contains(c.option.AdminUsers, user)
}

// CheckProjectPermission check permission for project
func (c *checker) CheckProjectPermission(ctx context.Context, project string, action RSAction) (
	*v1alpha1.AppProject, int, error) {
	result, statusCode, err := c.checkSingleResourcePermission(ctx, project, ProjectRSType, project, action)
	if err != nil {
		return nil, statusCode, err
	}
	return result.(*v1alpha1.AppProject), http.StatusOK, nil
}

// GetProjectMultiPermission get multi projects permission
func (c *checker) GetProjectMultiPermission(ctx context.Context, projectIDs []string,
	actions []RSAction) (map[string]map[RSAction]bool, error) {
	return c.getBCSMultiProjectPermission(ctx, projectIDs, actions)
}

// CheckClusterPermission check cluster permission
func (c *checker) CheckClusterPermission(ctx context.Context, query *cluster.ClusterQuery, action RSAction) (
	*v1alpha1.Cluster, int, error) {
	argoCluster, err := c.store.GetCluster(ctx, query)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "get cluster from storage failure")
	}
	if argoCluster == nil {
		return nil, http.StatusBadRequest, errors.Errorf("cluster '%v' not found", query)
	}
	var statusCode int
	_, statusCode, err = c.checkSingleResourcePermission(ctx, argoCluster.Project, ClusterRSType,
		argoCluster.Name, action)
	if err != nil {
		return nil, statusCode, err
	}
	return argoCluster, http.StatusOK, nil
}

// CheckRepoPermission check repo permission
func (c *checker) CheckRepoPermission(ctx context.Context, repo string, action RSAction) (*v1alpha1.Repository,
	int, error) {
	argoRepo, err := c.store.GetRepository(ctx, repo)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "get repository from storage failed")
	}
	if argoRepo == nil {
		return nil, http.StatusBadRequest, errors.Errorf("repository '%s' not found", repo)
	}
	var statusCode int
	_, statusCode, err = c.checkSingleResourcePermission(ctx, argoRepo.Project, RepoRSType, repo, action)
	if err != nil {
		return nil, statusCode, err
	}
	return argoRepo, http.StatusOK, nil
}

// CheckRepoCreate check repo create
func (c *checker) CheckRepoCreate(ctx context.Context, repo *v1alpha1.Repository) (int, error) {
	_, projectID, statusCode, err := c.getProjectWithID(ctx, repo.Project)
	if err != nil {
		return statusCode, errors.Wrapf(err, "get proejct failed")
	}
	permits, err := c.getBCSMultiProjectPermission(ctx, []string{projectID}, []RSAction{ProjectViewRSAction})
	if err != nil {
		return statusCode, errors.Wrapf(err, "get project permission failed")
	}
	if v := permits[projectID]; v == nil || !v[ProjectViewRSAction] {
		return http.StatusForbidden, errors.Errorf("user '%s' not have '%s/%s' permission",
			ctxutils.User(ctx).GetUser(), repo.Project, ProjectViewRSAction)
	}
	return http.StatusOK, nil
}

// QueryResourceUsers 获取某些资源对应的具备权限的用户信息
// NOTE: 此接口仅针对存储在数据库中的权限
func (c *checker) QueryResourceUsers(ctx context.Context, project string, rsType RSType, resources []string) (
	map[string]*ResourceUserPermission, int, error) {
	_, statusCode, err := c.CheckProjectPermission(ctx, project, ProjectViewRSAction)
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "check permission for project '%s' failed", project)
	}
	permissions, err := c.db.ListResourceUsers(project, string(rsType), resources)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "list reosurce's users failed")
	}
	resourceMap := make(map[string][]*dao.UserPermission)
	for _, permit := range permissions {
		resourceMap[permit.ResourceName] = append(resourceMap[permit.ResourceName], permit)
	}
	result := make(map[string]*ResourceUserPermission)
	for rsName, permits := range resourceMap {
		rpr := &ResourceUserPermission{
			ResourceType: rsType,
			ResourceName: rsName,
			UserPerms:    make(map[string]map[string]bool),
		}
		for _, permit := range permits {
			if _, ok := rpr.UserPerms[permit.User]; ok {
				rpr.UserPerms[permit.User][permit.ResourceAction] = true
			} else {
				rpr.UserPerms[permit.User] = map[string]bool{
					permit.ResourceAction: true,
				}
			}
		}
		result[rsName] = rpr
	}
	return result, http.StatusOK, nil
}

// UserAllPermissions return all user permission
func (c *checker) UserAllPermissions(ctx context.Context, project string) ([]*UserResourcePermission, int, error) {
	argoProject, projectID, statusCode, err := c.getProjectWithID(ctx, project)
	if err != nil {
		return nil, statusCode, err
	}
	allPermits, err := c.GetProjectMultiPermission(ctx, []string{projectID}, []RSAction{
		ProjectViewRSAction, ProjectEditRSAction, ProjectDeleteRSAction,
	})
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "get project permission failed")
	}
	if v := allPermits[projectID]; v == nil || !v[ProjectViewRSAction] {
		return nil, http.StatusBadRequest, errors.Errorf("user '%s' not have project_view "+
			"permission for project '%s'", ctxutils.User(ctx).GetUser(), project)
	}

	permitType := []RSType{ProjectRSType, ClusterRSType, RepoRSType, AppRSType, AppSetRSType}
	permits := make([]*UserResourcePermission, 0)
	for _, pt := range permitType {
		var permit *UserResourcePermission
		_, permit, statusCode, err = c.queryUserPermissionSingleType(ctx, argoProject, pt, nil, allPermits[projectID])
		if err != nil {
			return nil, statusCode, errors.Wrapf(err, "query permission for type '%s' failed", pt)
		}
		permits = append(permits, permit)
	}
	return permits, http.StatusOK, nil
}

func (c *checker) checkSingleResourcePermission(ctx context.Context, project string, resourceType RSType,
	resource string, action RSAction) (interface{}, int, error) {
	resources, permit, statusCode, err := c.QueryUserPermissions(ctx, project,
		resourceType, []string{resource})
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "query '%s/%s/%s' permission failed", resourceType, resource,
			action)
	}
	if _, ok := permit.ResourcePerms[resource]; !ok {
		return nil, http.StatusBadRequest, errors.Errorf("user '%s' not have permission for resource '%s'",
			ctxutils.User(ctx).GetUser(), resource)
	}
	if !permit.ResourcePerms[resource][action] {
		return nil, http.StatusForbidden, errors.Errorf("user '%s' not have '%s' permission for resource '%s/%s'",
			ctxutils.User(ctx).GetUser(), action, string(resourceType), resource)
	}
	if len(resources) != 1 {
		return nil, http.StatusInternalServerError, errors.Errorf("not get project when query permission")
	}
	return resources[0], http.StatusOK, nil
}

// QueryUserPermissions 获取用户对应的资源的权限信息
func (c *checker) QueryUserPermissions(ctx context.Context, project string, rsType RSType, rsNames []string) (
	[]interface{}, *UserResourcePermission, int, error) {
	// insert operate user if check project permission
	user := ctxutils.User(ctx)
	go c.db.UpdateActivityUserWithName(&dao.ActivityUserItem{
		Project: project, User: user.GetUser(),
	})

	argoProject, projectID, statusCode, err := c.getProjectWithID(ctx, project)
	if err != nil {
		return nil, nil, statusCode, err
	}
	projPermits, err := c.GetProjectMultiPermission(ctx, []string{projectID}, []RSAction{
		ProjectViewRSAction, ProjectEditRSAction, ProjectDeleteRSAction,
	})
	if err != nil {
		return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "get project permission failed")
	}
	if v := projPermits[projectID]; v == nil || !v[ProjectViewRSAction] {
		return nil, nil, statusCode, errors.Errorf("user '%s' not have project_view permission for project",
			user.GetUser())
	}
	return c.queryUserPermissionSingleType(ctx, argoProject, rsType, rsNames, projPermits[projectID])
}

// queryUserPermissionSingleType query the user permission with single resource type
func (c *checker) queryUserPermissionSingleType(ctx context.Context, argoProject *v1alpha1.AppProject, rsType RSType,
	rsNames []string, projPermits map[RSAction]bool) ([]interface{}, *UserResourcePermission, int, error) {
	start := time.Now()
	defer blog.Infof("RequestID[%s] query permission for project '%s' with resource '%s/%v' cost time: %v",
		ctxutils.RequestID(ctx), argoProject.Name, rsType, rsNames, time.Since(start))
	var result *UserResourcePermission
	var resources []interface{}
	var statusCode int
	var err error
	switch rsType {
	case ProjectRSType:
		resources, result, statusCode = c.queryUserResourceForProject(argoProject, projPermits)
	case ClusterRSType:
		resources, result, statusCode, err = c.queryUserResourceForCluster(ctx, argoProject, projPermits, rsNames)
	case RepoRSType:
		resources, result, statusCode, err = c.queryUserResourceForRepo(ctx, argoProject, projPermits, rsNames)
	case AppRSType:
		resources, result, statusCode, err = c.queryUserResourceForApp(ctx, argoProject, rsNames)
	case AppSetRSType:
		resources, result, statusCode, err = c.queryUserResourceForAppSets(ctx, projPermits, argoProject, rsNames)
	default:
		return nil, nil, http.StatusBadRequest, errors.Errorf("unknown resource type '%s'", rsType)
	}
	if err != nil {
		return nil, nil, statusCode, errors.Wrapf(err, "query user reosurces failed")
	}
	return resources, result, http.StatusOK, nil
}

// queryUserResourceForProject query project permission
func (c *checker) queryUserResourceForProject(argoProj *v1alpha1.AppProject,
	projPermits map[RSAction]bool) ([]interface{}, *UserResourcePermission, int) {
	result := &UserResourcePermission{
		ResourceType: ProjectRSType,
		ActionPerms:  projPermits,
		ResourcePerms: map[string]map[RSAction]bool{
			argoProj.Name: projPermits,
		},
	}
	resources := []interface{}{argoProj}
	return resources, result, http.StatusOK
}

// queryUserResourceForCluster query cluster permission
func (c *checker) queryUserResourceForCluster(ctx context.Context, argoProj *v1alpha1.AppProject,
	projPermits map[RSAction]bool, rsNames []string) ([]interface{}, *UserResourcePermission, int, error) {
	var clusterList *v1alpha1.ClusterList
	var err error
	if len(rsNames) == 1 {
		var argoCluster *v1alpha1.Cluster
		argoCluster, err = c.store.GetCluster(ctx, &cluster.ClusterQuery{
			Name: rsNames[0],
		})
		if err != nil {
			return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "list clusters failed")
		}
		if argoCluster == nil {
			return nil, nil, http.StatusBadRequest, errors.Errorf("cluster '%s' not found", rsNames[0])
		}
		if argoCluster.Project != argoProj.Name {
			return nil, nil, http.StatusBadRequest, errors.Errorf("cluster '%s' not belongs to '%s'",
				rsNames[0], argoProj.Name)
		}
		clusterList = &v1alpha1.ClusterList{
			Items: []v1alpha1.Cluster{*argoCluster},
		}
	} else {
		clusterList, err = c.store.ListClustersByProject(ctx, common.GetBCSProjectID(argoProj.Annotations))
		if err != nil {
			return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "list clusters failed")
		}
	}

	result := &UserResourcePermission{
		ResourceType: ClusterRSType,
		ActionPerms: map[RSAction]bool{
			ClusterViewRSAction: projPermits[ProjectViewRSAction],
		},
		ResourcePerms: make(map[string]map[RSAction]bool),
	}
	resources := make([]interface{}, 0)
	for _, cls := range clusterList.Items {
		if len(rsNames) != 0 && !slices.Contains(rsNames, cls.Name) {
			continue
		}
		result.ResourcePerms[cls.Name] = map[RSAction]bool{
			ClusterViewRSAction: projPermits[ProjectViewRSAction],
		}
		resources = append(resources, &cls)
	}
	return resources, result, http.StatusOK, nil
}

// queryUserResourceForRepo query repo permission
func (c *checker) queryUserResourceForRepo(ctx context.Context, argoProj *v1alpha1.AppProject,
	projPermits map[RSAction]bool, rsNames []string) ([]interface{}, *UserResourcePermission, int, error) {
	var repoList *v1alpha1.RepositoryList
	var err error
	if len(rsNames) == 1 {
		var argoRepo *v1alpha1.Repository
		argoRepo, err = c.store.GetRepository(ctx, rsNames[0])
		if err != nil {
			return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "list repositories failed")
		}
		if argoRepo == nil {
			return nil, nil, http.StatusBadRequest, errors.Errorf("repository '%s' not found", rsNames[0])
		}
		if argoRepo.Project != argoProj.Name {
			return nil, nil, http.StatusBadRequest, errors.Errorf("repository '%s' not belongs to '%s'",
				rsNames[0], argoProj.Name)
		}
		repoList = &v1alpha1.RepositoryList{Items: []*v1alpha1.Repository{argoRepo}}
	} else {
		repoList, err = c.store.ListRepository(ctx, []string{argoProj.Name})
		if err != nil {
			return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "list repositories failed")
		}
	}

	result := &UserResourcePermission{
		ResourceType: RepoRSType,
		ActionPerms: map[RSAction]bool{
			RepoViewRSAction:   projPermits[ProjectViewRSAction],
			RepoCreateRSAction: projPermits[ProjectViewRSAction],
			RepoUpdateRSAction: projPermits[ProjectViewRSAction],
			RepoDeleteRSAction: projPermits[ProjectEditRSAction],
		},
		ResourcePerms: make(map[string]map[RSAction]bool),
	}
	resources := make([]interface{}, 0)
	for _, repo := range repoList.Items {
		if len(rsNames) != 0 && !slices.Contains(rsNames, repo.Repo) {
			continue
		}
		result.ResourcePerms[repo.Repo] = map[RSAction]bool{
			RepoViewRSAction:   projPermits[ProjectViewRSAction],
			RepoCreateRSAction: projPermits[ProjectViewRSAction],
			RepoUpdateRSAction: projPermits[ProjectViewRSAction],
			RepoDeleteRSAction: projPermits[ProjectEditRSAction],
		}
		resources = append(resources, repo)
	}
	return resources, result, http.StatusOK, nil
}

// buildClusterNSForQueryByApp build cluster namespace for query by application
func (c *checker) buildClusterNSForQueryByApp(ctx context.Context, project string, resources []string) (
	map[string]map[string]struct{}, map[string]string, []*v1alpha1.Application, error) {
	appList, err := c.store.ListApplications(ctx, &appclient.ApplicationQuery{
		Projects: []string{project},
	})
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "query applications failed")
	}
	argoApps := make([]*v1alpha1.Application, 0)
	for i := range appList.Items {
		argoApp := appList.Items[i]
		if len(resources) != 0 && !slices.Contains(resources, argoApp.Name) {
			continue
		}
		argoApps = append(argoApps, &argoApp)
	}
	clusterServerNSMap := make(map[string]map[string]struct{})
	for _, argoApp := range argoApps {
		clsServer := argoApp.Spec.Destination.Server
		ns := argoApp.Spec.Destination.Namespace
		_, ok := clusterServerNSMap[clsServer]
		if ok {
			clusterServerNSMap[clsServer][ns] = struct{}{}
		} else {
			clusterServerNSMap[clsServer] = map[string]struct{}{ns: {}}
		}
	}
	clusterServerNameMap := make(map[string]string)
	for clsServer := range clusterServerNSMap {
		var argoCluster *v1alpha1.Cluster
		argoCluster, err = c.store.GetCluster(ctx, &argocluster.ClusterQuery{
			Server: clsServer,
		})
		if err != nil {
			return nil, nil, nil, errors.Wrapf(err, "get cluster '%s' failed", clsServer)
		}
		if argoCluster == nil {
			continue
		}
		clusterServerNameMap[clsServer] = argoCluster.Name
	}
	clusterNSMap := make(map[string]map[string]struct{})
	for clsServer, nsMap := range clusterServerNSMap {
		clsName := clusterServerNameMap[clsServer]
		clusterNSMap[clsName] = nsMap
	}
	return clusterNSMap, clusterServerNameMap, argoApps, nil
}

// buildClusterNSForQueryByNamespace build query cluster namespace for query by namespace
func (c *checker) buildClusterNSForQueryByNamespace(resources []string) map[string]map[string]struct{} {
	clusterNsMap := make(map[string]map[string]struct{})
	for _, res := range resources {
		t := strings.Split(res, ":")
		if len(t) != 2 {
			continue
		}
		cls := t[0]
		ns := t[1]
		if _, ok := clusterNsMap[cls]; ok {
			clusterNsMap[cls][ns] = struct{}{}
		} else {
			clusterNsMap[cls] = map[string]struct{}{ns: {}}
		}
	}
	return clusterNsMap
}

// queryUserResourceForApp query application permission
func (c *checker) queryUserResourceForApp(ctx context.Context, argoProj *v1alpha1.AppProject, rsNames []string) (
	[]interface{}, *UserResourcePermission, int, error) {
	var clusterNamespaceMap map[string]map[string]struct{}
	var clusterServerNameMap map[string]string
	var argoApps []*v1alpha1.Application
	// 如果请求的第一条数据是"集群ID:命名空间"格式，则认为是获取 AppCreate 权限
	queriedByNamespace := false
	if len(rsNames) != 0 && strings.HasPrefix(rsNames[0], "BCS-K8S-") &&
		len(strings.Split(rsNames[0], ":")) == 2 {
		queriedByNamespace = true
		clusterNamespaceMap = c.buildClusterNSForQueryByNamespace(rsNames)
	} else {
		var err error
		clusterNamespaceMap, clusterServerNameMap, argoApps, err = c.buildClusterNSForQueryByApp(ctx,
			argoProj.Name, rsNames)
		if err != nil {
			return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "build cluster namespace failed")
		}
	}
	permits, err := c.getBCSNamespaceScopedPermission(ctx, common.GetBCSProjectID(argoProj.Annotations),
		clusterNamespaceMap)
	if err != nil {
		return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "auth center failed")
	}

	result := &UserResourcePermission{
		ResourceType: AppRSType,
		ActionPerms: map[RSAction]bool{AppViewRSAction: true, AppCreateRSAction: true,
			AppUpdateRSAction: true, AppDeleteRSAction: true},
		ResourcePerms: make(map[string]map[RSAction]bool),
	}
	resources := make([]interface{}, 0)
	if queriedByNamespace {
		for cls, nsMap := range clusterNamespaceMap {
			for ns := range nsMap {
				rsName := cls + ":" + ns
				nsPermit, ok := permits[authutils.CalcIAMNsID(cls, ns)]
				if !ok {
					result.ResourcePerms[rsName] = map[RSAction]bool{
						AppCreateRSAction: false, AppUpdateRSAction: false, AppDeleteRSAction: false,
					}
					continue
				}
				result.ResourcePerms[rsName] = map[RSAction]bool{
					AppViewRSAction:   true,
					AppCreateRSAction: nsPermit[string(iamnamespace.NameSpaceScopedCreate)],
					AppUpdateRSAction: nsPermit[string(iamnamespace.NameSpaceScopedUpdate)],
					AppDeleteRSAction: nsPermit[string(iamnamespace.NameSpaceScopedDelete)],
				}
				resources = append(resources, rsNames)
			}
		}
		return resources, result, http.StatusOK, nil
	}

	for _, argoApp := range argoApps {
		clsServer := argoApp.Spec.Destination.Server
		ns := argoApp.Spec.Destination.Namespace
		cls := clusterServerNameMap[clsServer]
		nsPermit, ok := permits[authutils.CalcIAMNsID(cls, ns)]
		if !ok {
			result.ResourcePerms[argoApp.Name] = map[RSAction]bool{
				AppCreateRSAction: false, AppUpdateRSAction: false, AppDeleteRSAction: false,
			}
			continue
		}
		result.ResourcePerms[argoApp.Name] = map[RSAction]bool{
			AppViewRSAction:   true,
			AppCreateRSAction: nsPermit[string(iamnamespace.NameSpaceScopedCreate)],
			AppUpdateRSAction: nsPermit[string(iamnamespace.NameSpaceScopedUpdate)],
			AppDeleteRSAction: nsPermit[string(iamnamespace.NameSpaceScopedDelete)],
		}
		resources = append(resources, argoApp)
	}
	return resources, result, http.StatusOK, nil
}

// queryUserResourceForAppSets query appset permission
func (c *checker) queryUserResourceForAppSets(ctx context.Context, projPermits map[RSAction]bool,
	argoProj *v1alpha1.AppProject, rsNames []string) ([]interface{}, *UserResourcePermission, int, error) {
	result := &UserResourcePermission{
		ResourceType: AppSetRSType,
		ActionPerms: map[RSAction]bool{
			AppSetViewRSAction:   projPermits[ProjectViewRSAction],
			AppSetCreateRSAction: projPermits[ProjectEditRSAction],
		},
		ResourcePerms: make(map[string]map[RSAction]bool),
	}
	result.ActionPerms[AppSetUpdateRSAction] = projPermits[ProjectEditRSAction]
	result.ActionPerms[AppSetDeleteRSAction] = projPermits[ProjectEditRSAction]

	// 获取所有的 appset
	appSetList, err := c.store.ListApplicationSets(ctx, &appsetpkg.ApplicationSetListQuery{
		Projects: []string{argoProj.Name},
	})
	if err != nil {
		return nil, nil, http.StatusInternalServerError, errors.Wrapf(err,
			"list applicationsets for project failed")
	}
	resources := make([]interface{}, 0)
	for i := range appSetList.Items {
		item := &appSetList.Items[i]
		if len(rsNames) != 0 && !slices.Contains(rsNames, item.Name) {
			continue
		}
		result.ResourcePerms[item.Name] = map[RSAction]bool{
			AppSetViewRSAction:   true,
			AppSetDeleteRSAction: projPermits[ProjectEditRSAction],
			// NOTE: 暂时维持 ProjectEdit 权限
			AppSetUpdateRSAction: projPermits[ProjectEditRSAction],
		}
		resources = append(resources, item)
	}

	// 如果数据库中具备权限，则将 Update 权限设置为 true
	user := ctxutils.User(ctx)
	permissions, err := c.db.ListUserPermissions(user.GetUser(), argoProj.Name, string(AppSetRSType))
	if err != nil {
		return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "list user's resources failed")
	}
	if len(permissions) != 0 {
		result.ActionPerms[AppSetUpdateRSAction] = true
	}
	for _, permit := range permissions {
		if _, ok := result.ResourcePerms[permit.ResourceName]; ok {
			result.ResourcePerms[permit.ResourceName][AppSetUpdateRSAction] = true
		}
	}
	return resources, result, http.StatusOK, nil
}

// UpdatePermissions 更新用户权限
func (c *checker) UpdatePermissions(ctx context.Context, project string, resourceType RSType,
	req *UpdatePermissionRequest) (int, error) {
	switch resourceType {
	case AppSetRSType:
		return c.updateAppSetPermissions(ctx, project, req)
	default:
		return http.StatusBadRequest, fmt.Errorf("not handler for resourceType=%s", resourceType)
	}
}

// updateAppSetPermissions update the appset permissions
func (c *checker) updateAppSetPermissions(ctx context.Context, project string,
	req *UpdatePermissionRequest) (int, error) {
	for i := range req.ResourceActions {
		action := req.ResourceActions[i]
		if action != string(AppSetUpdateRSAction) {
			return http.StatusBadRequest, fmt.Errorf("not allowed action '%s'", action)
		}
	}
	argoProject, statusCode, err := c.CheckProjectPermission(ctx, project, ProjectViewRSAction)
	if err != nil {
		return statusCode, errors.Wrapf(err, "check permission for project '%s' failed", project)
	}
	clusterCreate, err := c.getBCSClusterCreatePermission(ctx, common.GetBCSProjectID(argoProject.Annotations))
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "check cluster_create permission failed")
	}
	if !clusterCreate {
		return http.StatusForbidden, errors.Errorf("user '%s' not have cluster_create permission",
			ctxutils.User(ctx).GetUser())
	}

	appSets := c.store.AllApplicationSets()
	appSetMap := make(map[string]*v1alpha1.ApplicationSet)
	for _, appSet := range appSets {
		appSetMap[appSet.Name] = appSet
	}
	notFoundAppSet := make([]string, 0)
	resultAppSets := make([]*v1alpha1.ApplicationSet, 0)
	// 校验请求的 resource_name
	for i := range req.ResourceNames {
		rsName := req.ResourceNames[i]
		tmp, ok := appSetMap[rsName]
		if !ok {
			notFoundAppSet = append(notFoundAppSet, rsName)
			continue
		}
		resultAppSets = append(resultAppSets, tmp)
		if tmpProj := tmp.Spec.Template.Spec.Project; tmpProj != project {
			return http.StatusBadRequest, fmt.Errorf("appset '%s' project '%s' not same as '%s'",
				rsName, tmpProj, project)
		}
	}
	if len(notFoundAppSet) != 0 {
		return http.StatusBadRequest, fmt.Errorf("appset '%v' not found", notFoundAppSet)
	}

	// 添加权限
	errs := make([]string, 0)
	for _, appSet := range resultAppSets {
		for _, action := range req.ResourceActions {
			err = c.db.UpdateResourcePermissions(project, string(AppSetRSType), appSet.Name, action, req.Users)
			if err == nil {
				blog.Infof("RequestID[%s] update resource '%s/%s' permissions success", ctxutils.RequestID(ctx),
					string(AppSetRSType), appSet.Name)
				continue
			}

			errMsg := fmt.Sprintf("update resource '%s/%s' permissions failed", string(AppSetRSType), appSet.Name)
			errs = append(errs, errMsg)
			blog.Errorf("RequestID[%s] update permission failed: %s", ctxutils.RequestID(ctx), errMsg)
		}
	}
	if len(errs) != 0 {
		return http.StatusInternalServerError, fmt.Errorf("create permission with multiple error: %v", errs)
	}
	return http.StatusOK, nil
}

// getBCSMultiProjectPermission get mutli-projects permission
func (c *checker) getBCSMultiProjectPermission(ctx context.Context, projectIDs []string,
	actions []RSAction) (map[string]map[RSAction]bool, error) {
	user := ctxutils.User(ctx)
	if c.isAdminUser(user.GetUser()) {
		result := make(map[string]map[RSAction]bool)
		for _, projectID := range projectIDs {
			result[projectID] = map[RSAction]bool{
				ProjectViewRSAction: true, ProjectEditRSAction: true, ProjectDeleteRSAction: true,
			}
		}
		return result, nil
	}

	bcsActions := make([]string, 0)
	for _, action := range actions {
		switch action {
		case ProjectViewRSAction:
			bcsActions = append(bcsActions, string(iam.ProjectView))
		case ProjectEditRSAction:
			bcsActions = append(bcsActions, string(iam.ProjectEdit))
		case ProjectDeleteRSAction:
			bcsActions = append(bcsActions, string(iam.ProjectDelete))
		}
	}
	var permits map[string]map[string]bool
	var err error
	for i := 0; i < 5; i++ {
		permits, err = c.projectPermission.GetMultiProjectMultiActionPerm(user.GetUser(), projectIDs, bcsActions)
		if err == nil {
			break
		}
		if !utils.NeedRetry(err) {
			break
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "get project permission failed")
	}

	newResult := make(map[string]map[RSAction]bool)
	for projID, projPermits := range permits {
		newResult[projID] = make(map[RSAction]bool)
		for act, perm := range projPermits {
			switch act {
			case string(iam.ProjectView):
				newResult[projID][ProjectViewRSAction] = perm
			case string(iam.ProjectEdit):
				newResult[projID][ProjectEditRSAction] = perm
			case string(iam.ProjectDelete):
				newResult[projID][ProjectDeleteRSAction] = perm
			}
		}
	}
	return newResult, nil
}

// getBCSClusterCreatePermission get bcs cluster creat permission
func (c *checker) getBCSClusterCreatePermission(ctx context.Context, projectID string) (bool, error) {
	user := ctxutils.User(ctx)
	if c.isAdminUser(user.GetUser()) {
		return true, nil
	}
	var err error
	for i := 0; i < 5; i++ {
		var permit bool
		permit, _, _, err = c.clusterPermission.CanCreateCluster(user.GetUser(), projectID)
		if err == nil {
			return permit, nil
		}
		if !utils.NeedRetry(err) {
			break
		}
	}
	return false, errors.Wrapf(err, "get cluster create permission failed")
}

// getBCSNamespaceScopedPermission get bcs namespace scoped permission
func (c *checker) getBCSNamespaceScopedPermission(ctx context.Context, projectID string,
	clusterNS map[string]map[string]struct{}) (map[string]map[string]bool, error) {
	user := ctxutils.User(ctx)
	if c.isAdminUser(user.GetUser()) {
		result := make(map[string]map[string]bool)
		for cls, nsMap := range clusterNS {
			for ns := range nsMap {
				result[authutils.CalcIAMNsID(cls, ns)] = map[string]bool{
					string(iamnamespace.NameSpaceScopedCreate): true, string(iamnamespace.NameSpaceScopedDelete): true,
					string(iamnamespace.NameSpaceScopedUpdate): true,
				}
			}
		}
		return result, nil
	}
	projNsData := make([]iamnamespace.ProjectNamespaceData, 0)
	for cls, nsMap := range clusterNS {
		for ns := range nsMap {
			projNsData = append(projNsData, iamnamespace.ProjectNamespaceData{
				Project:   projectID,
				Cluster:   cls,
				Namespace: ns,
			})
		}
	}

	var err error
	for i := 0; i < 5; i++ {
		var permits map[string]map[string]bool
		permits, err = c.namespacePermission.GetMultiNamespaceMultiActionPerm(user.GetUser(), projNsData, []string{
			string(iamnamespace.NameSpaceScopedCreate), string(iamnamespace.NameSpaceScopedDelete),
			string(iamnamespace.NameSpaceScopedUpdate),
		})
		if err == nil {
			return permits, nil
		}
		if !utils.NeedRetry(err) {
			break
		}
	}
	return nil, errors.Wrapf(err, "get nameespace scoped permission failed")
}

// getProjectWithID get project with id
func (c *checker) getProjectWithID(ctx context.Context, projectName string) (*v1alpha1.AppProject, string, int, error) {
	if projectName == "" {
		return nil, "", http.StatusBadRequest, errors.Errorf("project name cannot be empty")
	}
	// get project info and validate projectPermission
	argoProject, err := c.store.GetProject(ctx, projectName)
	if err != nil {
		return nil, "", http.StatusInternalServerError, errors.Wrapf(err, "get project from storage failure")
	}
	if argoProject == nil {
		return nil, "", http.StatusBadRequest, errors.Errorf("project '%s' not found", projectName)
	}
	projectID := common.GetBCSProjectID(argoProject.Annotations)
	if projectID == "" {
		return nil, "", http.StatusForbidden,
			errors.Errorf("project '%s' got id failed, not under control", projectName)
	}
	return argoProject, projectID, http.StatusOK, nil
}
