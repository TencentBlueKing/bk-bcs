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
	"testing"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
	clusternetapps "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"
)

func TestDRWorkflowWebhook_DefaultSetsHelmChartDefaults(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{
					Name: "chart",
					Type: drv1alpha1.ActionTypeHelmChart,
					HelmChart: &drv1alpha1.HelmChartAction{
						Name:      "demo-chart",
						Namespace: "default",
						Spec: &clusternetapps.HelmChartSpec{
							HelmOptions: clusternetapps.HelmOptions{
								Repository: "oci://registry.example.com/charts",
								Chart:      "demo-app",
							},
							TargetNamespace: "default",
						},
					},
				},
			},
		},
	}

	if err := webhook.Default(context.Background(), workflow); err != nil {
		t.Fatalf("default failed: %v", err)
	}

	action := workflow.Spec.Actions[0]
	if action.HelmChart.Operation != drv1alpha1.OperationCreate {
		t.Fatalf("expected default operation Create, got %q", action.HelmChart.Operation)
	}
}

func TestDRWorkflowWebhook_DefaultSetsGlobalizationDefaults(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{
					Name: "glob",
					Type: drv1alpha1.ActionTypeGlobalization,
					Globalization: &drv1alpha1.GlobalizationAction{
						Name: "demo-global-values",
						Spec: &clusternetapps.GlobalizationSpec{
							Feed: clusternetapps.Feed{
								APIVersion: "apps.clusternet.io/v1alpha1",
								Kind:       "HelmChart",
								Name:       "demo-app",
								Namespace:  "default",
							},
						},
					},
				},
			},
		},
	}

	if err := webhook.Default(context.Background(), workflow); err != nil {
		t.Fatalf("default failed: %v", err)
	}

	action := workflow.Spec.Actions[0]
	if action.Globalization.Operation != drv1alpha1.OperationCreate {
		t.Fatalf("expected default operation Create, got %q", action.Globalization.Operation)
	}
	if action.Globalization.Spec.Priority != 500 {
		t.Fatalf("expected default priority 500, got %d", action.Globalization.Spec.Priority)
	}
}

func TestDRWorkflowWebhook_ValidateWorkflow_HelmChartApplyAllowsMissingRollback(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{
					Name: "chart",
					Type: drv1alpha1.ActionTypeHelmChart,
					HelmChart: &drv1alpha1.HelmChartAction{
						Operation: drv1alpha1.OperationApply,
						Name:      "demo-chart",
						Namespace: "default",
						Spec: &clusternetapps.HelmChartSpec{
							HelmOptions: clusternetapps.HelmOptions{
								Repository: "oci://registry.example.com/charts",
								Chart:      "demo-app",
							},
							TargetNamespace: "default",
						},
					},
				},
			},
		},
	}

	warnings, errs := webhook.validateWorkflow(workflow)
	if len(errs) != 0 {
		t.Fatalf("expected Apply without rollback to be valid, got errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for HelmChart Apply without rollback, got %v", warnings)
	}
}

func TestDRWorkflowWebhook_ValidateWorkflow_HelmChartDeleteDoesNotRequireSpec(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{
					Name: "chart",
					Type: drv1alpha1.ActionTypeHelmChart,
					HelmChart: &drv1alpha1.HelmChartAction{
						Operation: drv1alpha1.OperationDelete,
						Name:      "demo-chart",
						Namespace: "default",
					},
				},
			},
		},
	}

	warnings, errs := webhook.validateWorkflow(workflow)
	if len(errs) != 0 {
		t.Fatalf("expected delete action to be valid, got errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for delete action, got %v", warnings)
	}
}

func TestDRWorkflowWebhook_ValidateWorkflow_KubernetesResourceApplyAllowsMissingRollback(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{
					Name: "manifest",
					Type: drv1alpha1.ActionTypeKubernetesResource,
					Resource: &drv1alpha1.KubernetesResourceAction{
						Operation: drv1alpha1.OperationApply,
						Manifest:  "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: demo\n",
					},
				},
			},
		},
	}

	_, errs := webhook.validateWorkflow(workflow)
	if len(errs) != 0 {
		t.Fatalf("expected KubernetesResource Apply without rollback to be valid, got errors: %v", errs)
	}
}

func TestDRWorkflowWebhook_ValidateWorkflow_KubernetesResourcePatchRequiresRollback(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{
					Name: "manifest",
					Type: drv1alpha1.ActionTypeKubernetesResource,
					Resource: &drv1alpha1.KubernetesResourceAction{
						Operation: drv1alpha1.OperationPatch,
						Manifest:  "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: demo\n",
					},
				},
			},
		},
	}

	_, errs := webhook.validateWorkflow(workflow)
	if len(errs) == 0 {
		t.Fatal("expected KubernetesResource Patch without rollback to be invalid")
	}
}

func TestDRWorkflowWebhook_ValidateWorkflow_GlobalizationApplyAllowsMissingRollback(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{
					Name: "glob",
					Type: drv1alpha1.ActionTypeGlobalization,
					Globalization: &drv1alpha1.GlobalizationAction{
						Operation: drv1alpha1.OperationApply,
						Name:      "demo-global-values",
						Spec: &clusternetapps.GlobalizationSpec{
							Feed: clusternetapps.Feed{
								APIVersion: "apps.clusternet.io/v1alpha1",
								Kind:       "HelmChart",
								Name:       "demo-app",
								Namespace:  "default",
							},
						},
					},
				},
			},
		},
	}

	warnings, errs := webhook.validateWorkflow(workflow)
	if len(errs) != 0 {
		t.Fatalf("expected Apply without rollback to be valid, got errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for Globalization Apply without rollback, got %v", warnings)
	}
}

func TestDRWorkflowWebhook_ValidateWorkflow_GlobalizationDeleteDoesNotRequireSpec(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{
					Name: "glob",
					Type: drv1alpha1.ActionTypeGlobalization,
					Globalization: &drv1alpha1.GlobalizationAction{
						Operation: drv1alpha1.OperationDelete,
						Name:      "demo-global-values",
					},
				},
			},
		},
	}

	warnings, errs := webhook.validateWorkflow(workflow)
	if len(errs) != 0 {
		t.Fatalf("expected delete action to be valid, got errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for delete action, got %v", warnings)
	}
}

func TestDRWorkflowWebhook_ValidateWorkflow_LocalizationApplyAllowsMissingRollback(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{
					Name: "loc",
					Type: drv1alpha1.ActionTypeLocalization,
					Localization: &drv1alpha1.LocalizationAction{
						Operation: drv1alpha1.OperationApply,
						Name:      "demo-localization",
						Namespace: "cluster-a",
						Spec: &clusternetapps.LocalizationSpec{
							Feed: clusternetapps.Feed{
								APIVersion: "apps.clusternet.io/v1alpha1",
								Kind:       "HelmChart",
								Name:       "demo-app",
								Namespace:  "default",
							},
						},
					},
				},
			},
		},
	}

	warnings, errs := webhook.validateWorkflow(workflow)
	if len(errs) != 0 {
		t.Fatalf("expected Apply without rollback to be valid, got errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for Localization Apply without rollback, got %v", warnings)
	}
}

func TestDRWorkflowWebhook_ValidateWorkflow_LocalizationDeleteDoesNotRequireSpec(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{
					Name: "loc",
					Type: drv1alpha1.ActionTypeLocalization,
					Localization: &drv1alpha1.LocalizationAction{
						Operation: drv1alpha1.OperationDelete,
						Name:      "demo-localization",
						Namespace: "cluster-a",
					},
				},
			},
		},
	}

	warnings, errs := webhook.validateWorkflow(workflow)
	if len(errs) != 0 {
		t.Fatalf("expected delete action to be valid, got errors: %v", errs)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for Localization delete action, got %v", warnings)
	}
}
