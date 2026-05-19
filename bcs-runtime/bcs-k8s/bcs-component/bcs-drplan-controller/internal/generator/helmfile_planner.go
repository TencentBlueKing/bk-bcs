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
	helmfilePreapplyWorkflow      = "preapply"
	helmfilePresyncWorkflow       = "presync"
	helmfilePostsyncWorkflow      = "postsync"
	helmfileGlobalizationPriority = 600
	helmfileTargetNamespaceRef    = "$(params.targetNamespace)"
	helmfileHookImageRef          = "$(params.hookImage)"
	helmfileHookJobTTLSeconds     = 300
	helmfileReservedNamespace     = "clusternet-reserved"
	helmfileHookJobAPIVersion     = "batch/v1"
	helmfileHookJobKind           = "Job"
)

const (
	clusternetCreatedByLabel       = "clusternet.io/created-by"
	clusternetHubName              = "clusternet-hub"
	clusternetConfigGroupLabel     = "apps.clusternet.io/config.group"
	clusternetConfigVersionLabel   = "apps.clusternet.io/config.version"
	clusternetConfigKindLabel      = "apps.clusternet.io/config.kind"
	clusternetConfigNameLabel      = "apps.clusternet.io/config.name"
	clusternetConfigNamespaceLabel = "apps.clusternet.io/config.namespace"
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

	// A release without hooks can use the compact single-workflow model. Once a
	// release has helmfile hooks, the generated output must preserve hook timing,
	// so planner switches to the hook-aware multi-stage model.
	if !hasHelmfileHooks(release) {
		wfBytes, marshalErr := sigyaml.Marshal(buildHelmfileWorkflow(release))
		if marshalErr != nil {
			return nil, fmt.Errorf("marshaling helmfile workflow: %w", marshalErr)
		}
		result.WorkflowYAMLs["workflow-execute.yaml"] = wfBytes
	} else {
		workflows, buildErr := buildHookAwareHelmfileWorkflows(release)
		if buildErr != nil {
			return nil, buildErr
		}
		for name, data := range workflows {
			result.WorkflowYAMLs[name] = data
		}
	}

	if err := generateExecutionSamples(cfg, result); err != nil {
		return nil, err
	}

	return result, nil
}

func validateHelmfileRelease(release HelmfileResolvedRelease) error {
	// These fields are required after helmfile resolution. Validation lives in
	// the planner instead of the CLI so tests and future callers using the
	// package API get the same error behavior.
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
	if hasHelmfileHooks(release) && strings.TrimSpace(release.HookImage) == "" {
		return fmt.Errorf("helmfile hook image is required when release hooks are present")
	}
	return nil
}

func buildHelmfilePlanYAML(release HelmfileResolvedRelease) map[string]interface{} {
	if hasHelmfileHooks(release) {
		return buildHookAwareHelmfilePlanYAML(release)
	}

	// In the no-hook path, install and upgrade are both represented by idempotent
	// Apply actions inside the execute workflow. Delete mode is inferred by the
	// workflow executor from the same positive actions, so the generated YAML
	// stays small and does not need explicit delete actions.
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

func buildHookAwareHelmfilePlanYAML(release HelmfileResolvedRelease) map[string]interface{} {
	stages := []map[string]interface{}{}
	// Only emit hook stages that contain actions. DRWorkflow validation rejects
	// empty action lists, and omitting empty stages also keeps the visible plan
	// aligned with the original helmfile release.
	if hasHelmfileHooksForEvent(release, helmfilePreapplyWorkflow) {
		stages = append(stages,
			buildHelmfileStage(release, helmfilePreapplyWorkflow, "Pre-apply hook workflow generated from helmfile release hooks"),
		)
	}
	if hasHelmfileHooksForEvent(release, helmfilePresyncWorkflow) {
		stages = append(stages,
			buildHelmfileStage(release, helmfilePresyncWorkflow, "Pre-sync hook workflow generated from helmfile release hooks"),
		)
	}
	stages = append(stages,
		buildHelmfileStage(release, helmfileWorkflowSuffix, "HelmChart-driven workflow generated from helmfile"),
	)
	// postsync must be a later stage rather than the final action in execute.
	// This lets the plan-level Continue policy run it after execute failures,
	// matching helmfile's expectation that postsync still gets a chance to run.
	if hasHelmfileHooksForEvent(release, helmfilePostsyncWorkflow) {
		stages = append(stages,
			buildHelmfileStage(release, helmfilePostsyncWorkflow, "Post-sync hook workflow generated from helmfile release hooks"),
		)
	}

	return map[string]interface{}{
		"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
		"kind":       "DRPlan",
		"metadata": map[string]interface{}{
			"name":      release.ReleaseName,
			"namespace": release.Namespace,
		},
		"spec": map[string]interface{}{
			"description":   fmt.Sprintf("Auto-generated from Helmfile release: %s", release.ReleaseName),
			"failurePolicy": "Continue",
			"globalParams": []map[string]interface{}{
				{"name": "feedNamespace", "value": release.Namespace},
				{"name": "targetNamespace", "value": release.TargetNamespace},
				{"name": "hookImage", "value": release.HookImage},
			},
			"stages": stages,
		},
	}
}

func buildHelmfileStage(release HelmfileResolvedRelease, stageName, description string) map[string]interface{} {
	return map[string]interface{}{
		"name":        stageName,
		"description": description,
		"workflows": []map[string]interface{}{
			{"workflowRef": map[string]interface{}{"name": fmt.Sprintf("%s-%s", release.ReleaseName, stageName)}},
		},
	}
}

func buildHelmfileWorkflow(release HelmfileResolvedRelease) map[string]interface{} {
	// Ordering is intentional: the HelmChart feed must exist before values are
	// attached via Globalization, and Subscription is applied last so Clusternet
	// distributes the final feed state.
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

func buildHookAwareHelmfileWorkflows(release HelmfileResolvedRelease) (map[string][]byte, error) {
	// workflow-execute is always present. Hook workflows are added selectively
	// below, which prevents generating invalid DRWorkflow objects with
	// `actions: []` for hook events that do not exist in the source release.
	workflows := map[string]map[string]interface{}{
		"workflow-execute.yaml": buildHelmfileWorkflow(release),
	}
	if hasHelmfileHooksForEvent(release, helmfilePreapplyWorkflow) {
		workflow, err := buildHelmfileHookWorkflow(release, helmfilePreapplyWorkflow, "FailFast")
		if err != nil {
			return nil, fmt.Errorf("building %s workflow: %w", helmfilePreapplyWorkflow, err)
		}
		workflows["workflow-preapply.yaml"] = workflow
	}
	if hasHelmfileHooksForEvent(release, helmfilePresyncWorkflow) {
		workflow, err := buildHelmfileHookWorkflow(release, helmfilePresyncWorkflow, "FailFast")
		if err != nil {
			return nil, fmt.Errorf("building %s workflow: %w", helmfilePresyncWorkflow, err)
		}
		workflows["workflow-presync.yaml"] = workflow
	}
	if hasHelmfileHooksForEvent(release, helmfilePostsyncWorkflow) {
		workflow, err := buildHelmfileHookWorkflow(release, helmfilePostsyncWorkflow, "Continue")
		if err != nil {
			return nil, fmt.Errorf("building %s workflow: %w", helmfilePostsyncWorkflow, err)
		}
		workflows["workflow-postsync.yaml"] = workflow
	}

	result := make(map[string][]byte, len(workflows))
	for name, workflow := range workflows {
		data, err := sigyaml.Marshal(workflow)
		if err != nil {
			return nil, fmt.Errorf("marshaling %s: %w", name, err)
		}
		result[name] = data
	}
	return result, nil
}

func buildHelmfileHookWorkflow(
	release HelmfileResolvedRelease,
	event string,
	failurePolicy string,
) (map[string]interface{}, error) {
	// Each hook workflow receives the same namespace parameters as execute plus
	// hookImage. This keeps generated YAML relocatable without re-rendering the
	// original helmfile templates.
	actions, err := buildHelmfileHookActions(release, event)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
		"kind":       "DRWorkflow",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("%s-%s", release.ReleaseName, event),
			"namespace": release.Namespace,
		},
		"spec": map[string]interface{}{
			"failurePolicy": failurePolicy,
			"parameters": []map[string]interface{}{
				{"name": "feedNamespace", "type": "string", "default": release.Namespace},
				{"name": "targetNamespace", "type": "string", "default": release.TargetNamespace},
				{"name": "hookImage", "type": "string", "default": release.HookImage},
			},
			"actions": actions,
		},
	}, nil
}

func buildHelmfileHookActions(release HelmfileResolvedRelease, event string) ([]map[string]interface{}, error) {
	actions := make([]map[string]interface{}, 0)
	for _, hook := range release.Hooks {
		if hook.Event != event {
			continue
		}
		manifestAction, err := buildHelmfileHookManifestAction(release, hook)
		if err != nil {
			return nil, fmt.Errorf("building manifest action for hook %s[%d]: %w", hook.Event, hook.Order, err)
		}
		actions = append(actions,
			manifestAction,
			buildHelmfileHookSubscriptionAction(release, hook),
		)
	}
	// nil is used to signal "no workflow should be generated" to callers. It is
	// never marshaled into YAML because empty action workflows are invalid.
	if len(actions) == 0 {
		return nil, nil
	}
	return actions, nil
}

func buildHelmfileHookManifestAction(release HelmfileResolvedRelease, hook HelmfileResolvedHook) (map[string]interface{}, error) {
	manifest, err := buildHelmfileHookManifest(release, hook)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name": helmfileHookManifestActionName(release, hook),
		"type": "KubernetesResource",
		"resource": map[string]interface{}{
			// Apply keeps hook feeds reusable across repeated upgrades. If a
			// revert path is needed, the executor can default this Apply to delete.
			"operation": "Apply",
			"manifest":  manifest,
		},
	}, nil
}

func buildHelmfileHookSubscriptionAction(release HelmfileResolvedRelease, hook HelmfileResolvedHook) map[string]interface{} {
	hookName := helmfileHookName(release, hook)
	return map[string]interface{}{
		"name":    helmfileHookSubscriptionActionName(release, hook),
		"type":    "Subscription",
		"timeout": "5m",
		"subscription": map[string]interface{}{
			"operation": "Apply",
			"name":      fmt.Sprintf("%s-subscription", hookName),
			"namespace": feedNamespaceRef,
			"spec": map[string]interface{}{
				"schedulingStrategy": "Replication",
				"feeds":              []map[string]interface{}{buildHelmfileHookJobFeed(release, hook)},
				"subscribers": []map[string]interface{}{
					{"clusterAffinity": map[string]interface{}{}},
				},
			},
		},
		"waitReady":            true,
		"clusterExecutionMode": "PerCluster",
		// BeforeCreate removes stale per-cluster hook subscriptions from a prior
		// run. Success/failure cleanup stays disabled so operators can inspect
		// Description and child-cluster Job state after a hook has executed.
		"hookCleanup": map[string]interface{}{
			"beforeCreate": true,
			"onSuccess":    false,
			"onFailure":    false,
		},
	}
}

func buildHelmfileHookManifest(release HelmfileResolvedRelease, hook HelmfileResolvedHook) (string, error) {
	hookName := helmfileHookName(release, hook)
	// The Job runs in the target namespace of each child cluster. It is stored
	// in a hub-side Clusternet Manifest below, so the hub cluster never executes
	// the script directly.
	job := map[string]interface{}{
		"apiVersion": helmfileHookJobAPIVersion,
		"kind":       helmfileHookJobKind,
		"metadata": map[string]interface{}{
			"name":      hookName,
			"namespace": helmfileTargetNamespaceRef,
			"labels": map[string]interface{}{
				clusternetCreatedByLabel: clusternetHubName,
			},
		},
		"spec": map[string]interface{}{
			"backoffLimit":            0,
			"ttlSecondsAfterFinished": helmfileHookJobTTLSeconds,
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"restartPolicy": "Never",
					"containers": []map[string]interface{}{
						buildHelmfileHookContainer(hook),
					},
				},
			},
		},
	}

	manifest := map[string]interface{}{
		"apiVersion": "apps.clusternet.io/v1alpha1",
		"kind":       "Manifest",
		"metadata": map[string]interface{}{
			"name":      helmfileHookManifestName(release, hook),
			"namespace": helmfileReservedNamespace,
			"labels":    buildHelmfileHookManifestLabels(hookName),
		},
		// Clusternet Manifest uses a top-level template field. Do not move this
		// under spec.template; that shape is rejected by the v0.18 CRD.
		"template": job,
	}

	data, err := sigyaml.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("marshaling hook manifest for %s: %w", hookName, err)
	}
	return string(data), nil
}

func buildHelmfileHookJobFeed(release HelmfileResolvedRelease, hook HelmfileResolvedHook) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": helmfileHookJobAPIVersion,
		"kind":       helmfileHookJobKind,
		"name":       helmfileHookName(release, hook),
		"namespace":  helmfileTargetNamespaceRef,
	}
}

func buildHelmfileHookManifestLabels(hookName string) map[string]interface{} {
	return map[string]interface{}{
		clusternetCreatedByLabel:       clusternetHubName,
		clusternetConfigGroupLabel:     "batch",
		clusternetConfigVersionLabel:   "v1",
		clusternetConfigKindLabel:      helmfileHookJobKind,
		clusternetConfigNameLabel:      hookName,
		clusternetConfigNamespaceLabel: helmfileTargetNamespaceRef,
	}
}

func buildHelmfileHookContainer(hook HelmfileResolvedHook) map[string]interface{} {
	container := map[string]interface{}{
		"name":  "hook",
		"image": helmfileHookImageRef,
	}
	if strings.TrimSpace(hook.Command) != "" {
		container["command"] = []string{hook.Command}
	}
	// Preserve helmfile's command/args split. Some existing hooks depend on the
	// command being argv[0] instead of being shell-joined into a single string.
	if len(hook.Args) > 0 {
		container["args"] = append([]string(nil), hook.Args...)
	}
	return container
}

func helmfileHookName(release HelmfileResolvedRelease, hook HelmfileResolvedHook) string {
	// Stable names make repeated Apply runs update the same Manifest and
	// Subscription instead of creating a new feed for every execution.
	return fmt.Sprintf("%s-%s-hook-%d", release.ReleaseName, hook.Event, hook.Order)
}

func helmfileHookManifestName(release HelmfileResolvedRelease, hook HelmfileResolvedHook) string {
	return fmt.Sprintf("jobs.%s.%s", helmfileTargetNamespaceRef, helmfileHookName(release, hook))
}

func helmfileHookManifestActionName(release HelmfileResolvedRelease, hook HelmfileResolvedHook) string {
	return fmt.Sprintf("create-%s-manifest", helmfileHookName(release, hook))
}

func helmfileHookSubscriptionActionName(release HelmfileResolvedRelease, hook HelmfileResolvedHook) string {
	return fmt.Sprintf("apply-%s-subscription", helmfileHookName(release, hook))
}

func hasHelmfileHooks(release HelmfileResolvedRelease) bool {
	return len(release.Hooks) > 0
}

func hasHelmfileHooksForEvent(release HelmfileResolvedRelease, event string) bool {
	for _, hook := range release.Hooks {
		if hook.Event == event {
			return true
		}
	}
	return false
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
	createNamespace := true
	if release.CreateNamespace != nil {
		createNamespace = *release.CreateNamespace
	}

	spec := map[string]interface{}{
		"repo":            release.ChartRepo,
		"chart":           release.Chart,
		"targetNamespace": helmfileTargetNamespaceRef,
		"createNamespace": createNamespace,
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
		// Clusternet exposes install and upgrade atomicity separately. Helmfile's
		// atomic flag maps to both so install and upgrade keep the same safety
		// behavior as the source release.
		spec["atomic"] = *release.Atomic
		spec["upgradeAtomic"] = *release.Atomic
	}
	if release.PlainHTTP != nil {
		spec["plainHTTP"] = *release.PlainHTTP
	}
	if release.TimeoutSeconds > 0 {
		spec["timeoutSeconds"] = release.TimeoutSeconds
	}
	return spec
}
