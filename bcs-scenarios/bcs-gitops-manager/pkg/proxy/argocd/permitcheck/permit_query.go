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

package permitcheck

import (
	"context"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
)

// UserAllPermissions return all user permission
func (c *checker) UserAllPermissions(ctx context.Context, project string) ([]*UserResourcePermission, int, error) {
	permitType := []RSType{ProjectRSType, ClusterRSType, RepoRSType, AppRSType, AppSetRSType}
	permits := make([]*UserResourcePermission, 0)
	for _, pt := range permitType {
		_, permit, statusCode, err := c.queryUserSingleTypePermission(ctx, project, pt, nil)
		if err != nil {
			return nil, statusCode, errors.Wrapf(err, "query permission for type '%s' failed", pt)
		}
		permits = append(permits, permit)
	}
	return permits, http.StatusOK, nil
}

// queryUserSingleTypePermission query user permission with single resource type
func (c *checker) queryUserSingleTypePermission(ctx context.Context, project string, rsType RSType,
	rsNames []string) ([]interface{}, *UserResourcePermission, int, error) {
	start := time.Now()
	defer blog.Infof("RequestID[%s] query permission for project '%s' with resource '%s/%v' cost time: %v",
		ctxutils.RequestID(ctx), project, rsType, rsNames, time.Since(start)) // nolint
	var result *UserResourcePermission
	var resources []interface{}
	var statusCode int
	var err error
	switch rsType {
	case ProjectRSType:
		resources, result, statusCode, err = c.getMultiProjectsMultiActionsPermit(ctx, []string{project})
	case ClusterRSType:
		resources, result, statusCode, err = c.getMultiClustersMultiActionsPermission(ctx, project, rsNames)
	case RepoRSType:
		resources, result, statusCode, err = c.getMultiRepoMultiActionPermission(ctx, project, rsNames)
	case AppRSType:
		resources, result, statusCode, err = c.getMultiAppMultiActionPermission(ctx, project, rsNames)
	case AppSetRSType:
		resources, result, statusCode, err = c.getMultiAppSetMultiActionPermission(ctx, project, rsNames)
	default:
		return nil, nil, http.StatusBadRequest, errors.Errorf("unknown resource type '%s'", rsType)
	}
	if err != nil {
		return nil, nil, statusCode, errors.Wrapf(err, "query user reosurces failed")
	}
	return resources, result, http.StatusOK, nil
}

// QueryUserPermissions 获取用户对应的资源的权限信息
func (c *checker) QueryUserPermissions(ctx context.Context, project string, rsType RSType, rsNames []string) (
	[]interface{}, *UserResourcePermission, int, error) {
	return c.queryUserSingleTypePermission(ctx, project, rsType, rsNames)
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
