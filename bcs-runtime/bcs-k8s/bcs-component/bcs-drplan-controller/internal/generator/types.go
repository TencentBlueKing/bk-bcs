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

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// Helm hook annotation keys and well-known hook types.
const (
	HookAnnotation       = "helm.sh/hook"
	HookWeightAnnotation = "helm.sh/hook-weight"
	HookDeletePolicy     = "helm.sh/hook-delete-policy"

	HookPreInstall   = "pre-install"
	HookPostInstall  = "post-install"
	HookPreUpgrade   = "pre-upgrade"
	HookPostUpgrade  = "post-upgrade"
	HookPreDelete    = "pre-delete"
	HookPostDelete   = "post-delete"
	HookPreRollback  = "pre-rollback"
	HookPostRollback = "post-rollback"
	HookTest         = "test"
	HookTestSuccess  = "test-success"

	DeletePolicyHookSucceeded      = "hook-succeeded"
	DeletePolicyHookFailed         = "hook-failed"
	DeletePolicyBeforeHookCreation = "before-hook-creation"

	DefaultTTLSecondsAfterFinished int32 = 60
	DefaultPollInterval                  = "5s"
)

// HookResource represents a Helm hook resource extracted from rendered YAML.
type HookResource struct {
	Resource     unstructured.Unstructured
	HookType     string
	Weight       int
	DeletePolicy string
}

// MainResource represents a non-hook Kubernetes resource.
type MainResource struct {
	Resource unstructured.Unstructured
}

// ChartAnalysis holds the result of classifying rendered YAML resources.
type ChartAnalysis struct {
	// Hooks maps hook type (e.g. "pre-install") to a weight-sorted list of hook resources.
	Hooks map[string][]HookResource
	// MainResources are non-hook resources to be deployed via Subscription.
	MainResources []MainResource
	// SkippedResources are resources that were skipped (e.g. test hooks).
	SkippedResources []unstructured.Unstructured
}

// GenerateConfig holds CLI parameters for plan generation.
type GenerateConfig struct {
	ReleaseName string
	Namespace   string
	OutputDir   string
}

// GenerateResult holds the output of plan generation.
type GenerateResult struct {
	PlanYAML       []byte
	WorkflowYAMLs  map[string][]byte // filename -> YAML content
	ExecutionYAMLs map[string][]byte // filename -> YAML content
}
