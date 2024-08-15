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
	"fmt"
	"net/http"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
)

// GetProjectMultiPermission get multi projects permission
func (c *checker) GetProjectMultiPermission(ctx context.Context, projectIDs []string,
	actions []RSAction) (map[string]map[RSAction]bool, error) {
	return c.getBCSMultiProjectPermission(ctx, projectIDs, actions)
}

// CheckProjectPermission check permission for project
func (c *checker) CheckProjectPermission(ctx context.Context, project string, action RSAction) (
	*v1alpha1.AppProject, int, error) {
	argoProj, permits, statusCode, err := c.getProjectMultiActionsPermission(ctx, project)
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "check project permission failed")
	}
	user := ctxutils.User(ctx)
	if !permits[action] {
		return nil, http.StatusForbidden, errors.Errorf("user '%s' not have project permission '%s/%s'",
			user.GetUser(), action, project)
	}
	return argoProj, http.StatusOK, nil
}

func (c *checker) getMultiProjectsMultiActionsPermission(ctx context.Context, projects []string) (
	[]interface{}, *UserResourcePermission, int, error) {
	resultObjs := make([]interface{}, 0, len(projects))
	urp := &UserResourcePermission{
		ResourceType:  ProjectRSType,
		ResourcePerms: make(map[string]map[RSAction]bool),
	}
	canView := false
	canEdit := false
	canDelete := false
	for _, project := range projects {
		argoProj, permits, statusCode, err := c.getProjectMultiActionsPermission(ctx, project)
		if err != nil {
			return nil, nil, statusCode, err
		}
		resultObjs = append(resultObjs, argoProj)
		urp.ResourcePerms[project] = permits
		if permits[ProjectViewRSAction] {
			canView = true
		}
		if permits[ProjectEditRSAction] {
			canEdit = true
		}
		if permits[ProjectDeleteRSAction] {
			canDelete = true
		}
	}
	urp.ActionPerms = map[RSAction]bool{
		ProjectViewRSAction:   canView,
		ProjectEditRSAction:   canEdit,
		ProjectDeleteRSAction: canDelete,
	}
	return resultObjs, urp, http.StatusOK, nil
}

func (c *checker) getProjectMultiActionsPermission(ctx context.Context, project string) (
	*v1alpha1.AppProject, map[RSAction]bool, int, error) {
	pctx, statusCode, err := c.createPermitContext(ctx, project)
	if err != nil {
		return nil, nil, statusCode, errors.Wrapf(err, "check project permission failed")
	}
	return pctx.argoProj, pctx.projPermits, http.StatusOK, nil
}

// checkRepositoryBelongProject check repo belong to project
func (c *checker) checkRepositoryBelongProject(ctx context.Context, repoUrl, project string) (bool, error) {
	repo, err := c.store.GetRepository(ctx, repoUrl)
	if err != nil {
		return false, errors.Wrapf(err, "get repo '%s' failed", repoUrl)
	}
	if repo == nil {
		return false, fmt.Errorf("repo '%s' not found", repoUrl)
	}
	// pass if repository's project equal to public projects
	if slices.Contains(c.option.PublicProjects, repo.Project) {
		return true, nil
	}

	if repo.Project != project {
		return false, nil
	}
	return true, nil
}
