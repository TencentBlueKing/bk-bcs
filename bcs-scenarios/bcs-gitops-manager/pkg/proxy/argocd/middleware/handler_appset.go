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
	"github.com/argoproj/argo-cd/v2/applicationset/services"
	"github.com/argoproj/argo-cd/v2/applicationset/utils"
	appsetpkg "github.com/argoproj/argo-cd/v2/pkg/apiclient/applicationset"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

var (
	render = utils.Render{}
)

// CheckCreateApplicationSet check the applicationset creation
func (h *handler) CheckCreateApplicationSet(ctx context.Context,
	appset *v1alpha1.ApplicationSet) ([]*v1alpha1.Application, int, error) {
	projName := appset.Spec.Template.Spec.Project
	if !strings.HasPrefix(appset.Name, projName) {
		appset.Name = projName + "-" + appset.Name
	}
	if !strings.HasPrefix(appset.Spec.Template.Name, appset.Spec.Template.Spec.Project) {
		appset.Spec.Template.Name = appset.Spec.Template.Spec.Project + "-" + appset.Spec.Template.Name
	}
	argoProject, statusCode, err := h.CheckProjectPermission(ctx, projName, iam.ProjectEdit)
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "check project '%s' permission failed", projName)
	}
	if appset.Spec.Template.Annotations == nil {
		appset.Spec.Template.Annotations = make(map[string]string)
	}
	appset.Spec.Template.Annotations[common.ProjectIDKey] = common.GetBCSProjectID(argoProject.Annotations)
	appset.Spec.Template.Annotations[common.ProjectBusinessIDKey] =
		common.GetBCSProjectBusinessKey(argoProject.Annotations)
	if err = h.checkApplicationSetGenerator(ctx, appset); err != nil {
		return nil, http.StatusBadRequest, errors.Wrapf(err, "check applicationset generator failed")
	}
	blog.Infof("RequestID[%s] check applicationset generator success", RequestID(ctx))

	repoClientSet := apiclient.NewRepoServerClientset(h.option.RepoServerUrl, 60,
		apiclient.TLSConfiguration{
			DisableTLS:       false,
			StrictValidation: false,
		})
	argoCDService, err := services.NewArgoCDService(h.option.Storage.GetArgoDB(),
		true, repoClientSet, false)
	// this will render the Applications by ApplicationSet's generators
	// refer to: https://github.com/argoproj/argo-cd/blob/v2.8.2/applicationset/controllers/applicationset_controller.go#L499
	results := make([]*v1alpha1.Application, 0)
	for i := range appset.Spec.Generators {
		generator := appset.Spec.Generators[i]
		if generator.List == nil && generator.Git == nil && generator.Matrix == nil {
			continue
		}
		var tsResult []generators.TransformResult
		listGenerator := generators.NewListGenerator()
		gitGenerator := generators.NewGitGenerator(argoCDService)
		tsResult, err = generators.Transform(generator, map[string]generators.Generator{
			"List": listGenerator,
			"Git":  gitGenerator,
			"Matrix": generators.NewMatrixGenerator(map[string]generators.Generator{
				"List": listGenerator,
				"Git":  gitGenerator,
			}),
		}, appset.Spec.Template, appset, map[string]interface{}{})
		if err != nil {
			return nil, http.StatusBadRequest, errors.Wrapf(err, "transform generator[%d] failed", i)
		}
		for j := range tsResult {
			ts := tsResult[j]
			tmplApplication := getTempApplication(ts.Template)
			if tmplApplication.Labels == nil {
				tmplApplication.Labels = make(map[string]string)
			}
			for _, p := range ts.Params {
				var app *v1alpha1.Application
				app, err = render.RenderTemplateParams(tmplApplication, appset.Spec.SyncPolicy,
					p, appset.Spec.GoTemplate, nil)
				if err != nil {
					return nil, http.StatusBadRequest, errors.Wrap(err, "error generating application from params")
				}
				statusCode, err = h.CheckCreateApplication(ctx, app)
				if err != nil {
					return nil, statusCode, errors.Wrapf(err, "check create application failed")
				}
				results = append(results, app)
			}
		}
	}
	return results, 0, nil
}

func (h *handler) checkApplicationSetGenerator(ctx context.Context, appset *v1alpha1.ApplicationSet) error {
	projName := appset.Spec.Template.Spec.Project
	var errs []string
	var repoUrls []string
	// NOTE: 仅允许 List/Git/Matrix Generator
	for i := range appset.Spec.Generators {
		generator := appset.Spec.Generators[i]
		if generator.Clusters != nil || generator.SCMProvider != nil ||
			generator.ClusterDecisionResource != nil || generator.PullRequest != nil ||
			generator.Merge != nil {
			errs = append(errs, fmt.Sprintf("generator[%d] has not allowed generator type", i))
			continue
		}
		if generator.Git != nil {
			generator.Git.Template = v1alpha1.ApplicationSetTemplate{}
			repoUrls = append(repoUrls, generator.Git.RepoURL)
		}
		if generator.List != nil {
			// rewrite generator.List.Template to empty value
			generator.List.Template = v1alpha1.ApplicationSetTemplate{}
		}
		if generator.Matrix != nil {
			if len(generator.Matrix.Generators) == 0 {
				errs = append(errs, fmt.Sprintf("generator[%d] with matrix generator have 0 generators", i))
				continue
			}
			for j := range generator.Matrix.Generators {
				matrixGenerator := generator.Matrix.Generators[j]
				if matrixGenerator.Clusters != nil || matrixGenerator.SCMProvider != nil ||
					matrixGenerator.ClusterDecisionResource != nil || matrixGenerator.PullRequest != nil ||
					matrixGenerator.Merge != nil {
					errs = append(errs, fmt.Sprintf("generator[%d] with matrix generator[%d] "+
						"has not allowed generator type", i, j))
				}
				if matrixGenerator.Git != nil {
					repoUrls = append(repoUrls, matrixGenerator.Git.RepoURL)
				}
			}
			generator.Matrix.Template = v1alpha1.ApplicationSetTemplate{}
		}
	}
	if len(errs) != 0 {
		return errors.Errorf("gitops only allowed [list,git,matrix] generator: %s", strings.Join(errs, "; "))
	}
	// check repository permission
	for _, repoUrl := range repoUrls {
		repoBelong, err := h.checkRepositoryBelongProject(ctx, repoUrl, projName)
		if err != nil {
			return errors.Wrapf(err, "check repository permission failed")
		}
		if !repoBelong {
			return errors.Errorf("repo '%s' not belong to project '%s'", repoUrl, projName)
		}
	}
	return nil
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
func (h *handler) CheckDeleteApplicationSet(ctx context.Context,
	appsetName string) (*v1alpha1.ApplicationSet, int, error) {
	appset, err := h.option.Storage.GetApplicationSet(ctx, appsetName)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "get applicationset failed")
	}
	if appset == nil {
		return nil, http.StatusNotFound, errors.Errorf("not found")
	}
	projName := appset.Spec.Template.Spec.Project
	_, statusCode, err := h.CheckProjectPermission(ctx, projName, iam.ProjectEdit)
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "check project '%s' permission failed", projName)
	}
	return appset, 0, nil
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
