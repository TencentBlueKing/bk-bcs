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
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	iamcluster "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/cluster"
	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
	iamproject "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/project"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"k8s.io/utils/strings/slices"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
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
	GetProjectMultiPermission(ctx context.Context, projects map[string]string, actions []RSAction) (
		map[string]map[RSAction]bool, error)

	CheckProjectPermission(ctx context.Context, project string, action RSAction) (*v1alpha1.AppProject, int, error)
	CheckClusterPermission(ctx context.Context, query *cluster.ClusterQuery, action RSAction) (*v1alpha1.Cluster,
		int, error)
	CheckRepoPermission(ctx context.Context, repo string, action RSAction) (*v1alpha1.Repository, int, error)
	CheckRepoCreate(ctx context.Context, repo *v1alpha1.Repository) (int, error)
	CheckApplicationPermission(ctx context.Context, app string, action RSAction) (*v1alpha1.Application, int, error)
	CheckApplicationCreate(ctx context.Context, app *v1alpha1.Application) (int, error)
	CheckAppSetPermission(ctx context.Context, appSet string, action RSAction) (*v1alpha1.ApplicationSet, int, error)
	CheckAppSetCreate(ctx context.Context, appSet *v1alpha1.ApplicationSet) ([]*v1alpha1.Application, int, error)

	UpdatePermissions(ctx context.Context, project string, resourceType RSType,
		req *UpdatePermissionRequest) (int, error)
	UserAllPermissions(ctx context.Context, project string) ([]*UserResourcePermission, int, error)
	QueryUserPermissions(ctx context.Context, project string, rsType RSType,
		rsNames []string) ([]interface{}, *UserResourcePermission, int, error)
	QueryResourceUsers(ctx context.Context, project string, rsType RSType,
		resources []string) (map[string]*ResourceUserPermission, int, error)

	// out-of-tree functions

	CheckBCSPermissions(req *http.Request) (bool, int, error)
	CheckBCSClusterPermission(ctx context.Context, user, clusterID string, action iam.ActionID) (int, error)
}

type checker struct {
	option *options.Options
	store  store.Store
	db     dao.Interface

	projectPermission   *iamproject.BCSProjectPerm
	clusterPermission   *iamcluster.BCSClusterPerm
	namespacePermission *iamnamespace.BCSNamespacePerm
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
