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
	"github.com/argoproj/argo-cd/v2/applicationset/generators"
	"github.com/argoproj/argo-cd/v2/applicationset/services"
	"github.com/argoproj/argo-cd/v2/applicationset/utils"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
)

var (
	render = utils.Render{}
)

// CheckAppSetPermission check appset permission
func (c *checker) CheckAppSetPermission(ctx context.Context, appSet string, action RSAction) (
	*v1alpha1.ApplicationSet, int, error) {
	argoAppSet, err := c.store.GetApplicationSet(ctx, appSet)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrapf(err, "get appset from storage failed")
	}
	if argoAppSet == nil {
		return nil, http.StatusBadRequest, errors.Errorf("appset '%s' not found", appSet)
	}
	var statusCode int
	_, statusCode, err = c.checkSingleResourcePermission(ctx, argoAppSet.Spec.Template.Spec.Project,
		AppSetRSType, appSet, action)
	if err != nil {
		return nil, statusCode, err
	}
	return argoAppSet, http.StatusOK, nil
}

// CheckAppSetCreate check appset create permission
func (c *checker) CheckAppSetCreate(ctx context.Context, appSet *v1alpha1.ApplicationSet) (
	[]*v1alpha1.Application, int, error) {
	projName := appSet.Spec.Template.Spec.Project
	if !strings.HasPrefix(appSet.Name, projName) {
		appSet.Name = projName + "-" + appSet.Name
	}
	if !strings.HasPrefix(appSet.Spec.Template.Name, projName) {
		appSet.Spec.Template.Name = projName + "-" + appSet.Spec.Template.Name
	}
	argoProject, statusCode, err := c.CheckProjectPermission(ctx, projName, ProjectEditRSAction)
	if err != nil {
		return nil, statusCode, errors.Wrapf(err, "check project permission failed")
	}
	if appSet.Spec.Template.Annotations == nil {
		appSet.Spec.Template.Annotations = make(map[string]string)
	}
	appSet.Spec.Template.Annotations[common.ProjectIDKey] = common.GetBCSProjectID(argoProject.Annotations)
	appSet.Spec.Template.Annotations[common.ProjectBusinessIDKey] =
		common.GetBCSProjectBusinessKey(argoProject.Annotations)
	if err = c.checkApplicationSetGenerator(ctx, appSet); err != nil {
		return nil, http.StatusBadRequest, errors.Wrapf(err, "check applicationset generator failed")
	}
	blog.Infof("RequestID[%s] check application set generator success", ctxutils.RequestID(ctx))

	repoClientSet := apiclient.NewRepoServerClientset(c.option.GitOps.RepoServer, 300,
		apiclient.TLSConfiguration{
			DisableTLS:       false,
			StrictValidation: false,
		})
	argoCDService, _ := services.NewArgoCDService(c.store.GetArgoDB(),
		true, repoClientSet, false)
	// this will render the Applications by ApplicationSet's generators
	// refer to:
	// https://github.com/argoproj/argo-cd/blob/v2.8.2/applicationset/controllers/applicationset_controller.go#L499
	results := make([]*v1alpha1.Application, 0)
	for i := range appSet.Spec.Generators {
		generator := appSet.Spec.Generators[i]
		if generator.List == nil && generator.Git == nil && generator.Matrix == nil && generator.Merge == nil {
			continue
		}
		var tsResult []generators.TransformResult
		listGenerator := generators.NewListGenerator()
		gitGenerator := generators.NewGitGenerator(argoCDService)
		terminalGenerators := map[string]generators.Generator{
			"List": listGenerator,
			"Git":  gitGenerator,
		}
		tsResult, err = generators.Transform(generator, map[string]generators.Generator{
			"List":   listGenerator,
			"Git":    gitGenerator,
			"Matrix": generators.NewMatrixGenerator(terminalGenerators),
			"Merge":  generators.NewMergeGenerator(terminalGenerators),
		}, appSet.Spec.Template, appSet, map[string]interface{}{})
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
				app, err = render.RenderTemplateParams(tmplApplication, appSet.Spec.SyncPolicy,
					p, appSet.Spec.GoTemplate, nil)
				if err != nil {
					return nil, http.StatusBadRequest, errors.Wrap(err, "error generating application from params")
				}
				statusCode, err = c.CheckApplicationCreate(ctx, app)
				if err != nil {
					return nil, statusCode, errors.Wrapf(err, "check create application failed")
				}
				results = append(results, app)
			}
		}
	}
	return results, 0, nil
}

// checkApplicationSetGenerator check applicationset generator
func (c *checker) checkApplicationSetGenerator(ctx context.Context, appSet *v1alpha1.ApplicationSet) error {
	projName := appSet.Spec.Template.Spec.Project
	var errs []string
	var repoUrls []string
	// NOTE: 仅允许 List/Git/Matrix/Merge Generator
	for i := range appSet.Spec.Generators {
		generator := appSet.Spec.Generators[i]
		if generator.Clusters != nil || generator.SCMProvider != nil ||
			generator.ClusterDecisionResource != nil || generator.PullRequest != nil {
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
		if generator.Merge != nil {
			if len(generator.Merge.Generators) == 0 {
				errs = append(errs, fmt.Sprintf("generator[%d] with merge generator have 0 generators", i))
				continue
			}
			for j := range generator.Merge.Generators {
				mergeGenerator := generator.Merge.Generators[j]
				if mergeGenerator.Clusters != nil || mergeGenerator.SCMProvider != nil ||
					mergeGenerator.ClusterDecisionResource != nil || mergeGenerator.PullRequest != nil ||
					mergeGenerator.Merge != nil {
					errs = append(errs, fmt.Sprintf("generator[%d] with merge generator[%d] "+
						"has not allowed generator type", i, j))
				}
				if mergeGenerator.Git != nil {
					repoUrls = append(repoUrls, mergeGenerator.Git.RepoURL)
				}
			}
			generator.Merge.Template = v1alpha1.ApplicationSetTemplate{}
		}
	}
	if len(errs) != 0 {
		return errors.Errorf("gitops only allowed [list,git,matrix] generator: %s", strings.Join(errs, "; "))
	}
	// check repository permission
	for _, repoUrl := range repoUrls {
		repoBelong, err := c.checkRepositoryBelongProject(ctx, repoUrl, projName)
		if err != nil {
			return errors.Wrapf(err, "check repository permission failed")
		}
		if !repoBelong {
			return errors.Errorf("repo '%s' not belong to project '%s'", repoUrl, projName)
		}
	}
	return nil
}

// refer to:
// https://github.com/argoproj/argo-cd/blob/v2.8.2/applicationset/controllers/applicationset_controller.go#L487
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
