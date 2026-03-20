/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package webhook

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// validateStageBasics validates stage name and workflow count
func validateStageBasics(stage drv1alpha1.Stage, index int) []string {
	var errors []string

	if stage.Name == "" {
		errors = append(errors, fmt.Sprintf("stage[%d]: name is required", index))
	}

	if len(stage.Workflows) == 0 {
		errors = append(errors, fmt.Sprintf("stage[%d] '%s': at least one workflow is required", index, stage.Name))
	}

	return errors
}

// validateStageWorkflows validates workflow references in a stage
func (w *DRPlanWebhook) validateStageWorkflows(ctx context.Context, stage drv1alpha1.Stage, stageIndex int, planNamespace string) ([]string, []string) {
	var warnings []string
	var errors []string

	for j, wfRef := range stage.Workflows {
		if wfRef.WorkflowRef.Name == "" {
			errors = append(errors, fmt.Sprintf("stage[%d] '%s' workflow[%d]: workflowRef.name is required",
				stageIndex, stage.Name, j))
			continue
		}

		// Determine workflow namespace
		namespace := wfRef.WorkflowRef.Namespace
		if namespace == "" {
			namespace = planNamespace
		}

		// Validate workflow exists and is ready
		workflow := &drv1alpha1.DRWorkflow{}
		key := client.ObjectKey{
			Name:      wfRef.WorkflowRef.Name,
			Namespace: namespace,
		}

		if err := w.Client.Get(ctx, key, workflow); err != nil {
			errors = append(errors, fmt.Sprintf("stage[%d] '%s' workflow[%d]: workflow %s/%s not found",
				stageIndex, stage.Name, j, namespace, wfRef.WorkflowRef.Name))
		} else if workflow.Status.Phase != drv1alpha1.PlanPhaseReady {
			warnings = append(warnings, fmt.Sprintf("stage[%d] '%s' workflow[%d]: workflow %s/%s is not ready (phase=%s)",
				stageIndex, stage.Name, j, namespace, wfRef.WorkflowRef.Name, workflow.Status.Phase))
		}
	}

	return warnings, errors
}

// validateParallelStage validates parallel stage constraints
func validateParallelStage(stage drv1alpha1.Stage, index int) []string {
	var warnings []string

	if stage.Parallel && len(stage.Workflows) == 1 {
		warnings = append(warnings, fmt.Sprintf("stage[%d] '%s': parallel=true with only one workflow, consider setting parallel=false",
			index, stage.Name))
	}

	return warnings
}

// validateStageDependencies validates stage dependency references
func validateStageDependencies(stages []drv1alpha1.Stage, stageNames map[string]bool) []string {
	var errors []string

	for i, stage := range stages {
		for _, depName := range stage.DependsOn {
			if !stageNames[depName] {
				errors = append(errors, fmt.Sprintf("stage[%d] '%s': depends on non-existent stage '%s'",
					i, stage.Name, depName))
			}
			if depName == stage.Name {
				errors = append(errors, fmt.Sprintf("stage[%d] '%s': cannot depend on itself",
					i, stage.Name))
			}
		}
	}

	return errors
}

// buildStageNameMap builds a map of stage names for validation
func buildStageNameMap(stages []drv1alpha1.Stage) (map[string]bool, []string) {
	var errors []string
	stageNames := make(map[string]bool)

	for i, stage := range stages {
		if stageNames[stage.Name] {
			errors = append(errors, fmt.Sprintf("stage[%d]: duplicate stage name '%s'", i, stage.Name))
		}
		stageNames[stage.Name] = true
	}

	return stageNames, errors
}

// validateGlobalParameters validates global parameters
func validateGlobalParameters(params []drv1alpha1.Parameter) []string {
	var errors []string
	paramNames := make(map[string]bool)

	for i, param := range params {
		if param.Name == "" {
			errors = append(errors, fmt.Sprintf("globalParams[%d]: name is required", i))
		}
		if paramNames[param.Name] {
			errors = append(errors, fmt.Sprintf("globalParams[%d]: duplicate parameter name '%s'", i, param.Name))
		}
		paramNames[param.Name] = true
	}

	return errors
}
