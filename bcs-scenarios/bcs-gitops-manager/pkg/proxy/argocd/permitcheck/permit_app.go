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
	"strings"

	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	appclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	argocluster "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	clusterclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
)

// CheckApplicationPermission check application permission
func (c *checker) CheckApplicationPermission(ctx context.Context, app string, action RSAction) (
	*v1alpha1.Application, int, error) {
	objs, permits, statusCode, err := c.getMultiApplicationMultiActionPermission(ctx, "", []string{app})
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "get application '%s' permission failed", app)
	}
	pm, ok := permits.ResourcePerms[app]
	if !ok {
		return nil, http.StatusBadRequest, errors.Errorf("application '%s' not exist", app)
	}
	if !pm[action] {
		return nil, http.StatusForbidden, errors.Errorf("user '%s' not have app permission '%s/%s'",
			ctxutils.User(ctx).GetUser(), action, app)
	}
	if len(objs) != 1 {
		return nil, http.StatusBadRequest, errors.Errorf("query app '%s' got '%d' apps", app, len(objs))
	}
	return objs[0].(*v1alpha1.Application), http.StatusOK, nil
}

func (c *checker) filterApplications(ctx context.Context, project string, apps []string) ([]interface{},
	map[string][]*v1alpha1.Application, int, error) {
	resultApps := make([]interface{}, 0, len(apps))
	projApps := make(map[string][]*v1alpha1.Application)
	if project != "" && len(apps) == 0 {
		appList, err := c.store.ListApplications(ctx, &appclient.ApplicationQuery{
			Projects: []string{project},
		})
		if err != nil {
			return nil, nil, http.StatusInternalServerError, err
		}
		for i := range appList.Items {
			argoApp := appList.Items[i]
			resultApps = append(resultApps, &argoApp)
			projApps[project] = append(projApps[project], &argoApp)
		}
	}
	for i := range apps {
		app := apps[i]
		argoApp, err := c.store.GetApplication(ctx, app)
		if err != nil {
			return nil, nil, http.StatusInternalServerError, errors.Wrapf(err, "get application failed")
		}
		if argoApp == nil {
			return nil, nil, http.StatusBadRequest, errors.Errorf("application '%s' not found", app)
		}
		proj := argoApp.Spec.Project
		_, ok := projApps[proj]
		if ok {
			projApps[proj] = append(projApps[proj], argoApp)
		} else {
			projApps[proj] = []*v1alpha1.Application{argoApp}
		}
		resultApps = append(resultApps, argoApp)
	}
	return resultApps, projApps, http.StatusOK, nil
}

func (c *checker) getMultiApplicationMultiActionPermission(ctx context.Context, project string, apps []string) (
	[]interface{}, *UserResourcePermission, int, error) {
	resultApps, projApps, statusCode, err := c.filterApplications(ctx, project, apps)
	if err != nil {
		return nil, nil, statusCode, err
	}
	result := &UserResourcePermission{
		ResourceType:  AppRSType,
		ResourcePerms: make(map[string]map[RSAction]bool),
		ActionPerms: map[RSAction]bool{AppViewRSAction: true, AppCreateRSAction: true, AppUpdateRSAction: true,
			AppDeleteRSAction: true},
	}
	if len(resultApps) == 0 {
		return resultApps, result, http.StatusOK, nil
	}

	canView := false
	canCreate := false
	canUpdate := false
	canDelete := false
	for proj, argoApps := range projApps {
		ctx, statusCode, err = c.createPermitContext(ctx, proj)
		if err != nil {
			return nil, nil, statusCode, err
		}
		clusterNamespaceMap, clusterServerNameMap, err := c.buildClusterNamespaceMap(ctx, argoApps)
		if err != nil {
			return nil, nil, http.StatusInternalServerError, errors.Wrapf(err,
				"build cluster namespace map failed")
		}
		permits, err := c.getBCSNamespaceScopedPermission(ctx, proj, contextGetProjID(ctx), clusterNamespaceMap)
		if err != nil {
			return nil, nil, http.StatusInternalServerError, errors.Wrapf(err,
				"auth center failed for project '%s'", contextGetProjName(ctx))
		}

		for _, argoApp := range argoApps {
			clsServer := argoApp.Spec.Destination.Server
			clsName := clusterServerNameMap[clsServer]
			ns := argoApp.Spec.Destination.Namespace
			nsPermits, ok := permits[authutils.CalcIAMNsID(clsName, ns)]
			if !ok {
				result.ResourcePerms[argoApp.Name] = map[RSAction]bool{
					AppCreateRSAction: false, AppUpdateRSAction: false, AppDeleteRSAction: false,
				}
				continue
			}
			appPermits := map[RSAction]bool{
				AppViewRSAction:   contextGetProjPermits(ctx)[ProjectViewRSAction],
				AppCreateRSAction: nsPermits[string(iamnamespace.NameSpaceScopedCreate)],
				AppUpdateRSAction: nsPermits[string(iamnamespace.NameSpaceScopedUpdate)],
				AppDeleteRSAction: nsPermits[string(iamnamespace.NameSpaceScopedDelete)],
			}
			result.ResourcePerms[argoApp.Name] = appPermits
			if appPermits[AppViewRSAction] {
				canView = true
			}
			if appPermits[AppCreateRSAction] {
				canCreate = true
			}
			if appPermits[AppUpdateRSAction] {
				canUpdate = true
			}
			if appPermits[AppDeleteRSAction] {
				canDelete = true
			}
		}
	}
	result.ActionPerms = map[RSAction]bool{
		AppViewRSAction:   canView,
		AppCreateRSAction: canCreate,
		AppUpdateRSAction: canUpdate,
		AppDeleteRSAction: canDelete,
	}
	return resultApps, result, http.StatusOK, nil
}

func (c *checker) buildClusterNamespaceMap(ctx context.Context, argoApps []*v1alpha1.Application) (
	map[string]map[string]struct{}, map[string]string, error) {
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
		argoCluster, err := c.store.GetCluster(ctx, &argocluster.ClusterQuery{
			Server: clsServer,
		})
		if err != nil {
			return nil, nil, errors.Wrapf(err, "get cluster '%s' failed", clsServer)
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
	return clusterNSMap, clusterServerNameMap, nil
}

// CheckApplicationCreate check application create permission
func (c *checker) CheckApplicationCreate(ctx context.Context, app *v1alpha1.Application) (int, error) {
	projectName := app.Spec.Project
	if projectName == "" || projectName == "default" { // nolint
		return http.StatusBadRequest, errors.Errorf("project information lost")
	}
	// 校验仓库是否归属于项目下
	var repoUrls []string
	if app.Spec.HasMultipleSources() {
		for i := range app.Spec.Sources {
			repoUrls = append(repoUrls, app.Spec.Sources[i].RepoURL)
		}
	} else {
		repoUrls = append(repoUrls, app.Spec.Source.RepoURL)
	}
	for i := range repoUrls {
		repoUrl := repoUrls[i]
		repoBelong, err := c.checkRepositoryBelongProject(ctx, repoUrl, projectName)
		if err != nil {
			return http.StatusBadRequest, err
		}
		if !repoBelong {
			return http.StatusBadRequest, errors.Errorf("repo '%s' not belong to project '%s'",
				repoUrl, projectName)
		}
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
		return http.StatusBadRequest, errors.Errorf("cluster '%v' not found", clusterQuery)
	}
	if argoCluster.Project != app.Spec.Project {
		return http.StatusBadRequest, errors.Errorf("cluster '%v' not belong to project '%s'",
			clusterQuery, app.Spec.Project)
	}

	// 校验用户是否具备创建权限
	clusterName := argoCluster.Name
	clusterNamespace := app.Spec.Destination.Namespace
	var statusCode int
	ctx, statusCode, err = c.createPermitContext(ctx, app.Spec.Project)
	if err != nil {
		return statusCode, err
	}
	permits, err := c.getBCSNamespaceScopedPermission(ctx, projectName, contextGetProjID(ctx),
		map[string]map[string]struct{}{
			clusterName: {clusterNamespace: struct{}{}},
		})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	nsPM, ok := permits[authutils.CalcIAMNsID(clusterName, clusterNamespace)]
	if !ok {
		return http.StatusBadRequest, errors.Errorf("cluster-namespace '%s/%s' permission not found",
			clusterName, clusterNamespace)
	}
	if !nsPM[string(iamnamespace.NameSpaceScopedCreate)] {
		return http.StatusForbidden, errors.Errorf("user '%s' not have 'namespace_scoped_create' "+
			"permission for '%s/%s'", ctxutils.User(ctx).GetUser(), clusterName, clusterNamespace)
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
