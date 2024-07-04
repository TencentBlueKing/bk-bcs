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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	clusterclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/utils/strings/slices"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
)

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

// CheckApplicationPermission check application permission
func (c *checker) CheckApplicationPermission(ctx context.Context, app string, action RSAction) (
	*v1alpha1.Application, int, error) {
	argoApp, err := c.store.GetApplication(ctx, app)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "get application from storage failed")
	}
	if argoApp == nil {
		return nil, http.StatusBadRequest, errors.Errorf("application '%s' not found", app)
	}
	var statusCode int
	_, statusCode, err = c.checkSingleResourcePermission(ctx, argoApp.Spec.Project, AppRSType, app, action)
	if err != nil {
		return nil, statusCode, err
	}
	return argoApp, http.StatusOK, nil
}

// CheckApplicationCreate check application create permission
func (c *checker) CheckApplicationCreate(ctx context.Context, app *v1alpha1.Application) (int, error) {
	projectName := app.Spec.Project
	if projectName == "" || projectName == "default" { // nolint
		return http.StatusBadRequest, errors.Errorf("project information lost")
	}
	// 校验仓库是否归属于项目下
	if app.Spec.HasMultipleSources() {
		for i := range app.Spec.Sources {
			appSource := app.Spec.Sources[i]
			repoUrl := appSource.RepoURL
			repoBelong, err := c.checkRepositoryBelongProject(ctx, repoUrl, projectName)
			if err != nil {
				return http.StatusBadRequest,
					errors.Wrapf(err, "check multi-source repository '%s' permission failed", repoUrl)
			}
			if !repoBelong {
				return http.StatusForbidden,
					errors.Errorf("check multi-source repo '%s' not belong to project '%s'", repoUrl, projectName)
			}
			blog.Infof("RequestID[%s] check multi-source repo '%s' success", ctxutils.RequestID(ctx), repoUrl)
		}
	} else if app.Spec.Source != nil {
		repoUrl := app.Spec.Source.RepoURL
		repoBelong, err := c.checkRepositoryBelongProject(ctx, repoUrl, projectName)
		if err != nil {
			return http.StatusBadRequest, errors.Wrapf(err, "check repository permission failed")
		}
		if !repoBelong {
			return http.StatusForbidden, errors.Errorf("repo '%s' not belong to project '%s'",
				repoUrl, projectName)
		}
		blog.Infof("RequestID[%s] check source repo '%s' success", ctxutils.RequestID(ctx), repoUrl)
	}
	// 校验集群是否存在
	clusterQuery := clusterclient.ClusterQuery{
		Server: app.Spec.Destination.Server,
		Name:   app.Spec.Destination.Name,
	}
	argoCluster, err := c.store.GetCluster(ctx, &clusterQuery)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "get cluster '%v' failed", clusterQuery)
	}
	if argoCluster == nil {
		return http.StatusBadRequest, fmt.Errorf("cluster '%v' not found", clusterQuery)
	}
	if argoCluster.Project != app.Spec.Project {
		return http.StatusBadRequest, fmt.Errorf("cluster '%v' not belong to project '%s'",
			clusterQuery, app.Spec.Project)
	}

	// 校验用户是否具备创建权限
	var statusCode int
	_, statusCode, err = c.checkSingleResourcePermission(ctx, app.Spec.Project, AppRSType, argoCluster.Name+":"+
		app.Spec.Destination.Namespace, AppCreateRSAction)
	if err != nil {
		return statusCode, err
	}
	// setting application name with project prefix
	if !strings.HasPrefix(app.Name, projectName+"-") {
		app.Name = projectName + "-" + app.Name
	}
	if app.Annotations == nil {
		app.Annotations = make(map[string]string)
	}
	var argoProject *v1alpha1.AppProject
	argoProject, _, statusCode, err = c.getProjectWithID(ctx, projectName)
	if err != nil {
		return statusCode, nil
	}
	common.AddCustomAnnotationForApplication(argoProject, app)
	return http.StatusOK, nil
}
