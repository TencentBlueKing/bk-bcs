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

package generator

import (
	"fmt"
	"strings"

	sigyaml "sigs.k8s.io/yaml"
)

const (
	helmChartApplyActionName      = "apply-helmchart"
	globalizationApplyActionName  = "apply-globalization"
	subscriptionApplyActionName   = "apply-subscription"
	helmfileWorkflowSuffix        = "execute"
	helmfileGlobalizationPriority = 600
	helmfileTargetNamespaceRef    = "$(params.targetNamespace)"
)

// GenerateHelmfilePlan creates DRPlan, DRWorkflow, and DRPlanExecution YAML from one resolved helmfile release.
func GenerateHelmfilePlan(release HelmfileResolvedRelease) (*GenerateResult, error) {
	if err := validateHelmfileRelease(release); err != nil {
		return nil, err
	}

	result := &GenerateResult{
		WorkflowYAMLs:  make(map[string][]byte),
		ExecutionYAMLs: make(map[string][]byte),
	}

	cfg := GenerateConfig{
		ReleaseName: release.ReleaseName,
		Namespace:   release.Namespace,
	}

	planBytes, err := sigyaml.Marshal(buildHelmfilePlanYAML(release))
	if err != nil {
		return nil, fmt.Errorf("marshaling helmfile DRPlan: %w", err)
	}
	result.PlanYAML = planBytes

	wfBytes, err := sigyaml.Marshal(buildHelmfileWorkflow(release))
	if err != nil {
		return nil, fmt.Errorf("marshaling helmfile workflow: %w", err)
	}
	result.WorkflowYAMLs["workflow-execute.yaml"] = wfBytes

	if err := generateExecutionSamples(cfg, result); err != nil {
		return nil, err
	}

	return result, nil
}

func validateHelmfileRelease(release HelmfileResolvedRelease) error {
	if strings.TrimSpace(release.ReleaseName) == "" {
		return fmt.Errorf("helmfile release name is required")
	}
	if strings.TrimSpace(release.Namespace) == "" {
		return fmt.Errorf("helmfile release namespace is required")
	}
	if strings.TrimSpace(release.TargetNamespace) == "" {
		return fmt.Errorf("helmfile release targetNamespace is required")
	}
	if strings.TrimSpace(release.Chart) == "" {
		return fmt.Errorf("helmfile chart name is required")
	}
	if strings.TrimSpace(release.ChartRepo) == "" {
		return fmt.Errorf("helmfile chart repo is required")
	}
	return nil
}

func buildHelmfilePlanYAML(release HelmfileResolvedRelease) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
		"kind":       "DRPlan",
		"metadata": map[string]interface{}{
			"name":      release.ReleaseName,
			"namespace": release.Namespace,
		},
		"spec": map[string]interface{}{
			"description":   fmt.Sprintf("Auto-generated from Helmfile release: %s", release.ReleaseName),
			"failurePolicy": "Stop",
			"globalParams": []map[string]interface{}{
				{"name": "feedNamespace", "value": release.Namespace},
				{"name": "targetNamespace", "value": release.TargetNamespace},
			},
			"stages": []map[string]interface{}{
				{
					"name":        helmfileWorkflowSuffix,
					"description": "HelmChart-driven workflow generated from helmfile",
					"workflows": []map[string]interface{}{
						{"workflowRef": map[string]interface{}{"name": fmt.Sprintf("%s-%s", release.ReleaseName, helmfileWorkflowSuffix)}},
					},
				},
			},
		},
	}
}

func buildHelmfileWorkflow(release HelmfileResolvedRelease) map[string]interface{} {
	actions := []map[string]interface{}{
		buildHelmChartApplyAction(release),
	}

	if strings.TrimSpace(release.ValuesYAML) != "" {
		actions = append(actions, buildGlobalizationApplyAction(release))
	}
	actions = append(actions, buildSubscriptionApplyAction(release))

	return map[string]interface{}{
		"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
		"kind":       "DRWorkflow",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("%s-%s", release.ReleaseName, helmfileWorkflowSuffix),
			"namespace": release.Namespace,
		},
		"spec": map[string]interface{}{
			"failurePolicy": "FailFast",
			"parameters": []map[string]interface{}{
				{"name": "feedNamespace", "type": "string", "default": release.Namespace},
				{"name": "targetNamespace", "type": "string", "default": release.TargetNamespace},
			},
			"actions": actions,
		},
	}
}

func buildHelmChartApplyAction(release HelmfileResolvedRelease) map[string]interface{} {
	return map[string]interface{}{
		"name": helmChartApplyActionName,
		"type": "HelmChart",
		"helmChart": map[string]interface{}{
			"operation": "Apply",
			"name":      release.ReleaseName,
			"namespace": feedNamespaceRef,
			"spec":      buildHelmChartSpec(release),
		},
	}
}

func buildGlobalizationApplyAction(release HelmfileResolvedRelease) map[string]interface{} {
	return map[string]interface{}{
		"name": globalizationApplyActionName,
		"type": "Globalization",
		"globalization": map[string]interface{}{
			"operation": "Apply",
			"name":      release.ReleaseName,
			"spec": map[string]interface{}{
				"priority":       helmfileGlobalizationPriority,
				"overridePolicy": "ApplyNow",
				"feed":           buildHelmChartFeed(release),
				"overrides": []map[string]interface{}{
					{
						"name":  "override-values",
						"type":  "Helm",
						"value": release.ValuesYAML,
					},
				},
			},
		},
	}
}

func buildSubscriptionApplyAction(release HelmfileResolvedRelease) map[string]interface{} {
	return map[string]interface{}{
		"name": subscriptionApplyActionName,
		"type": "Subscription",
		"subscription": map[string]interface{}{
			"operation": "Apply",
			"name":      fmt.Sprintf("%s-subscription", release.ReleaseName),
			"namespace": feedNamespaceRef,
			"spec": map[string]interface{}{
				"schedulingStrategy": "Replication",
				"feeds":              []map[string]interface{}{buildHelmChartFeed(release)},
				"subscribers": []map[string]interface{}{
					{"clusterAffinity": map[string]interface{}{}},
				},
			},
		},
	}
}

func buildHelmChartFeed(release HelmfileResolvedRelease) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "apps.clusternet.io/v1alpha1",
		"kind":       "HelmChart",
		"name":       release.ReleaseName,
		"namespace":  feedNamespaceRef,
	}
}

func buildHelmChartSpec(release HelmfileResolvedRelease) map[string]interface{} {
	spec := map[string]interface{}{
		"repo":            release.ChartRepo,
		"chart":           release.Chart,
		"targetNamespace": helmfileTargetNamespaceRef,
	}
	if release.ChartVersion != "" {
		spec["version"] = release.ChartVersion
	}
	if release.Wait != nil {
		spec["wait"] = *release.Wait
	}
	if release.WaitForJob != nil {
		spec["waitForJob"] = *release.WaitForJob
	}
	if release.Atomic != nil {
		spec["atomic"] = *release.Atomic
		spec["upgradeAtomic"] = *release.Atomic
	}
	if release.PlainHTTP != nil {
		spec["plainHTTP"] = *release.PlainHTTP
	}
	if release.CreateNamespace != nil {
		spec["createNamespace"] = *release.CreateNamespace
	}
	if release.TimeoutSeconds > 0 {
		spec["timeoutSeconds"] = release.TimeoutSeconds
	}
	return spec
}
