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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

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

// getBCSMultiProjectPermission get mutli-projects permission
func (c *checker) getBCSMultiProjectPermission(ctx context.Context, projects map[string]string,
	actions []RSAction) (map[string]map[RSAction]bool, error) {
	// set admin-user have all permissions
	user := ctxutils.User(ctx)
	if c.isAdminUser(user.GetUser()) {
		result := make(map[string]map[RSAction]bool)
		for _, projectID := range projects {
			result[projectID] = map[RSAction]bool{
				ProjectViewRSAction: true, ProjectEditRSAction: true, ProjectDeleteRSAction: true,
			}
		}
		return result, nil
	}

	publicProjIDs := make(map[string]struct{})
	for projName, projID := range projects {
		if slices.Contains(c.option.PublicProjects, projName) {
			publicProjIDs[projID] = struct{}{}
		}
	}
	result := make(map[string]map[RSAction]bool)
	defer func() {
		// rewrite project_view for public projects
		for k := range result {
			if _, ok := publicProjIDs[k]; !ok {
				continue
			}
			result[k][ProjectViewRSAction] = true
		}
	}()
	// rewrite not tencent user with only-view permission
	if !user.IsTencent {
		for projName, projID := range projects {
			result[projID] = map[RSAction]bool{ProjectViewRSAction: false,
				ProjectEditRSAction: false, ProjectDeleteRSAction: false}
			authed, err := dao.GlobalDB().CheckExternalUserPermission(user.GetUser(), projName)
			if err != nil {
				blog.Errorf("check external user permission error: %v", err)
				continue
			}
			result[projID][ProjectViewRSAction] = authed
		}
		return result, nil
	}

	projectIDs := make([]string, 0, len(projects))
	for _, projectID := range projects {
		projectIDs = append(projectIDs, projectID)
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
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "get project permission failed")
	}

	for projID, projPermits := range permits {
		result[projID] = make(map[RSAction]bool)
		for act, perm := range projPermits {
			switch act {
			case string(iam.ProjectView):
				result[projID][ProjectViewRSAction] = perm
			case string(iam.ProjectEdit):
				result[projID][ProjectEditRSAction] = perm
			case string(iam.ProjectDelete):
				result[projID][ProjectDeleteRSAction] = perm
			}
		}
	}
	return result, nil
}

// getBCSClusterCreatePermission get bcs cluster creat permission
func (c *checker) getBCSClusterCreatePermission(ctx context.Context, projectID string) (bool, error) {
	user := ctxutils.User(ctx)
	if c.isAdminUser(user.GetUser()) {
		return true, nil
	}
	// rewrite not tencent user with only-view permission
	if !user.IsTencent {
		return false, nil
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
		time.Sleep(2 * time.Second)
	}
	return false, errors.Wrapf(err, "get cluster create permission failed")
}

// getBCSNamespaceScopedPermission get bcs namespace scoped permission
func (c *checker) getBCSNamespaceScopedPermission(ctx context.Context, proj, projectID string,
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

	result := make(map[string]map[string]bool)
	defer func() {
		// rewrite namespace_scoped_view for public-project
		if slices.Contains(c.option.PublicProjects, proj) {
			for k := range result {
				result[k][string(iamnamespace.NameSpaceScopedView)] = true
			}
		}
	}()
	// rewrite not tencent user with only-view permission
	if !user.IsTencent {
		authed, err := dao.GlobalDB().CheckExternalUserPermission(user.GetUser(), proj)
		if err != nil {
			blog.Errorf("check external user permission error: %v", err)
			authed = false
		}
		for cls, nsMap := range clusterNS {
			for ns := range nsMap {
				result[authutils.CalcIAMNsID(cls, ns)] = map[string]bool{
					string(iamnamespace.NameSpaceScopedView):   authed,
					string(iamnamespace.NameSpaceScopedCreate): false,
					string(iamnamespace.NameSpaceScopedDelete): false,
					string(iamnamespace.NameSpaceScopedUpdate): false,
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
		result, err = c.namespacePermission.GetMultiNamespaceMultiActionPerm(user.GetUser(), projNsData, []string{
			string(iamnamespace.NameSpaceScopedCreate), string(iamnamespace.NameSpaceScopedDelete),
			string(iamnamespace.NameSpaceScopedUpdate), string(iamnamespace.NameSpaceScopedView),
		})
		if err == nil {
			return result, nil
		}
		if !utils.NeedRetry(err) {
			break
		}
		time.Sleep(2 * time.Second)
	}
	return nil, errors.Wrapf(err, "get nameespace scoped permission failed")
}
