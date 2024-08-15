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

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
)

// CheckRepoPermission check repo permission
func (c *checker) CheckRepoPermission(ctx context.Context, repo string, action RSAction) (*v1alpha1.Repository,
	int, error) {
	objs, permits, statusCode, err := c.getMultiRepoMultiActionPermission(ctx, "", []string{repo})
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "get repository permission failed")
	}
	pm, ok := permits.ResourcePerms[repo]
	if !ok {
		return nil, http.StatusBadRequest, errors.Errorf("repository '%s' not exist", repo)
	}
	if !pm[action] {
		return nil, http.StatusForbidden, errors.Errorf("user '%s' not have repo permission '%s/%s'",
			ctxutils.User(ctx).GetUser(), action, repo)
	}
	if len(objs) != 1 {
		return nil, http.StatusBadRequest, errors.Errorf("query repository '%s' got '%d' repos", repo, len(objs))
	}
	return objs[0].(*v1alpha1.Repository), http.StatusOK, nil
}

// CheckRepoCreate check repo create
func (c *checker) CheckRepoCreate(ctx context.Context, repo *v1alpha1.Repository) (int, error) {
	_, statusCode, err := c.CheckProjectPermission(ctx, repo.Project, ProjectViewRSAction)
	if err != nil {
		return statusCode, errors.Wrapf(err, "check repo create permission failed")
	}
	return http.StatusOK, nil
}

// getMultiRepoMultiActionPermission query repo permission
func (c *checker) getMultiRepoMultiActionPermission(ctx context.Context, project string, repos []string) ([]interface{},
	*UserResourcePermission, int, error) {
	resultRepos := make([]interface{}, 0, len(repos))
	projRepos := make(map[string][]*v1alpha1.Repository)
	if project != "" && len(repos) == 0 {
		repoList, err := c.store.ListRepository(ctx, []string{project})
		if err != nil {
			return nil, nil, http.StatusInternalServerError, err
		}
		for _, argoRepo := range repoList.Items {
			resultRepos = append(resultRepos, argoRepo)
			projRepos[project] = append(projRepos[project], argoRepo)
		}
	} else {
		for i := range repos {
			repo := repos[i]
			argoRepo, err := c.store.GetRepository(ctx, repo)
			if err != nil {
				return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "get repos failed")
			}
			if argoRepo == nil {
				return nil, nil, http.StatusBadRequest, errors.Errorf("repo '%s' not found", repo)
			}
			proj := argoRepo.Project
			_, ok := projRepos[proj]
			if ok {
				projRepos[proj] = append(projRepos[proj], argoRepo)
			} else {
				projRepos[proj] = []*v1alpha1.Repository{argoRepo}
			}
			resultRepos = append(resultRepos, argoRepo)
		}
	}

	canDelete := false
	canViewOrCreateOrUpdate := false
	result := &UserResourcePermission{
		ResourceType:  RepoRSType,
		ResourcePerms: make(map[string]map[RSAction]bool),
	}
	for proj, argoRepos := range projRepos {
		_, projPermits, statusCode, err := c.getProjectMultiActionsPermission(ctx, proj)
		if err != nil {
			return nil, nil, statusCode, err
		}
		for _, repo := range argoRepos {
			result.ResourcePerms[repo.Repo] = map[RSAction]bool{
				RepoViewRSAction:   projPermits[ProjectViewRSAction],
				RepoCreateRSAction: projPermits[ProjectViewRSAction],
				RepoUpdateRSAction: projPermits[ProjectViewRSAction],
				RepoDeleteRSAction: projPermits[ProjectEditRSAction],
			}
		}
		if projPermits[ProjectEditRSAction] {
			canDelete = true
		}
		if projPermits[ProjectViewRSAction] {
			canViewOrCreateOrUpdate = true
		}
	}
	result.ActionPerms = map[RSAction]bool{
		RepoViewRSAction:   canViewOrCreateOrUpdate,
		RepoCreateRSAction: canViewOrCreateOrUpdate,
		RepoUpdateRSAction: canViewOrCreateOrUpdate,
		RepoDeleteRSAction: canDelete,
	}
	return resultRepos, result, http.StatusOK, nil
}
