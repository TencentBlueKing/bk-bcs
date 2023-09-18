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
	"fmt"
	"net/http"
	"strings"

	"github.com/argoproj/argo-cd/v2/applicationset/generators"
	"github.com/argoproj/argo-cd/v2/applicationset/utils"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	clusterclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

var (
	render = utils.Render{}
)

func (h *handler) checkDryRunApplication(ctx context.Context, app *v1alpha1.Application) (int, error) {
	repoUrl := app.Spec.Source.RepoURL
	repo, err := h.option.Storage.GetRepository(ctx, repoUrl)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "get repo '%s' failed", repoUrl)
	}
	if repo == nil {
		return http.StatusNotFound, fmt.Errorf("repo '%s' not found", repoUrl)
	}
	clusterQuery := clusterclient.ClusterQuery{
		Server: app.Spec.Destination.Server,
		Name:   app.Spec.Destination.Name,
	}
	argoCluster, err := h.option.Storage.GetCluster(ctx, &clusterQuery)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "get cluster '%v' failed", clusterQuery)
	}
	if argoCluster == nil {
		return http.StatusNotFound, fmt.Errorf("cluster '%v' not found", clusterQuery)
	}
	return 0, nil
}

func (h *handler) CheckCreateApplicationSet(ctx context.Context, appset *v1alpha1.ApplicationSet) (int, error) {
	projName := appset.Spec.Template.Spec.Project
	if !strings.HasPrefix(appset.Name, projName) {
		appset.Name = projName + "-" + appset.Name
	}
	if !strings.HasPrefix(appset.Spec.Template.Name, appset.Spec.Template.Spec.Project) {
		appset.Spec.Template.Name = appset.Spec.Template.Spec.Project + "-" + appset.Spec.Template.Name
	}
	argoProject, statusCode, err := h.CheckProjectPermission(ctx, projName, iam.ProjectEdit)
	if err != nil {
		return statusCode, errors.Wrapf(err, "check project '%s' permission failed", projName)
	}
	if appset.Spec.Template.Annotations == nil {
		appset.Spec.Template.Annotations = make(map[string]string)
	}
	appset.Spec.Template.Annotations[common.ProjectIDKey] =
		common.GetBCSProjectID(argoProject.Annotations)
	appset.Spec.Template.Annotations[common.ProjectBusinessIDKey] =
		common.GetBCSProjectBusinessKey(argoProject.Annotations)

	var errs []string
	// NOTE: 仅允许 List Generator
	for i := range appset.Spec.Generators {
		generator := appset.Spec.Generators[i]
		if generator.Clusters != nil || generator.Git != nil || generator.SCMProvider != nil ||
			generator.ClusterDecisionResource != nil || generator.PullRequest != nil ||
			generator.Matrix != nil || generator.Merge != nil {
			errs = append(errs, fmt.Sprintf("generator[%d] has not allowed generator type", i))
			continue
		}
		if generator.List != nil {
			// rewrite generator.List.Template to empty value
			generator.List.Template = v1alpha1.ApplicationSetTemplate{}
		}
	}
	if len(errs) != 0 {
		return http.StatusBadRequest,
			errors.Errorf("gitops only allowed [list] generator: %s", strings.Join(errs, "; "))
	}
	// this will render the Applications by ApplicationSet's generators
	// refer to: https://github.com/argoproj/argo-cd/blob/v2.8.2/applicationset/controllers/applicationset_controller.go#L499
	for i := range appset.Spec.Generators {
		generator := appset.Spec.Generators[i]
		tsResult, err := generators.Transform(generator, map[string]generators.Generator{
			"List": generators.NewListGenerator(),
		}, appset.Spec.Template, appset, map[string]interface{}{})
		if err != nil {
			return http.StatusBadRequest, errors.Wrapf(err, "transform generator[%d] failed", i)
		}
		for j := range tsResult {
			ts := tsResult[j]
			tmplApplication := getTempApplication(ts.Template)
			if tmplApplication.Labels == nil {
				tmplApplication.Labels = make(map[string]string)
			}
			for _, p := range ts.Params {
				app, err := render.RenderTemplateParams(tmplApplication, appset.Spec.SyncPolicy, p, appset.Spec.GoTemplate)
				if err != nil {
					return http.StatusBadRequest, errors.Wrap(err, "error generating application from params")
				}
				statusCode, err = h.checkDryRunApplication(ctx, app)
				if err != nil {
					return statusCode, errors.Wrapf(err, "check create application failed")
				}
			}
		}
	}
	return 0, nil
}

// refer to: https://github.com/argoproj/argo-cd/blob/v2.8.2/applicationset/controllers/applicationset_controller.go#L487
func getTempApplication(applicationSetTemplate v1alpha1.ApplicationSetTemplate) *v1alpha1.Application {
	var tmplApplication v1alpha1.Application
	tmplApplication.Annotations = applicationSetTemplate.Annotations
	tmplApplication.Labels = applicationSetTemplate.Labels
	tmplApplication.Namespace = applicationSetTemplate.Namespace
	tmplApplication.Name = applicationSetTemplate.Name
	tmplApplication.Spec = applicationSetTemplate.Spec
	tmplApplication.Finalizers = applicationSetTemplate.Finalizers

	return &tmplApplication
}

// CheckDeleteApplicationSet check delete applicationset
func (h *handler) CheckDeleteApplicationSet(ctx context.Context, appsetName string) (int, error) {
	appset, err := h.option.Storage.GetApplicationSet(ctx, appsetName)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "get applicationset failed")
	}
	if appset == nil {
		return http.StatusNotFound, errors.Errorf("not found")
	}
	projName := appset.Spec.Template.Spec.Project
	_, statusCode, err := h.CheckProjectPermission(ctx, projName, iam.ProjectEdit)
	if err != nil {
		return statusCode, errors.Wrapf(err, "check project '%s' permission failed", projName)
	}
	return 0, nil
}

func (h *handler) CheckGetApplicationSet(ctx context.Context, appsetName string) (int, error) {
	appset, err := h.option.Storage.GetApplicationSet(ctx, appsetName)
	if err != nil {
		return http.StatusInternalServerError, errors.Wrapf(err, "get applicationset failed")
	}
	if appset == nil {
		return http.StatusNotFound, errors.Errorf("not found")
	}
	projName := appset.Spec.Template.Spec.Project
	_, statusCode, err := h.CheckProjectPermission(ctx, projName, iam.ProjectEdit)
	if err != nil {
		return statusCode, errors.Wrapf(err, "check project '%s' permission failed", projName)
	}
	return 0, nil
}

// ListApplicationSets list applicationsets
func (h *handler) ListApplicationSets(ctx context.Context, query *appsetpkg.ApplicationSetListQuery) (
	*v1alpha1.ApplicationSetList, error) {
	appsets, err := h.option.Storage.ListApplicationSets(ctx, query)
	if err != nil {
		return nil, errors.Wrapf(err, "list application swith project '%v' failed", *query)
	}
	return appsets, nil
}
