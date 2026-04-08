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
	"sort"
	"strings"

	sigyaml "sigs.k8s.io/yaml"
)

const (
	// feedNamespaceRef is resolved at execution time from DRPlan global params.
	feedNamespaceRef = "$(params.feedNamespace)"
	// mainActionName is the canonical action that applies chart main resources.
	mainActionName = "create-subscription"
	// deleteActionName is the canonical action that removes chart main resources.
	deleteActionName = "delete-subscription"
)

// GeneratePlan creates DRPlan, DRWorkflow, and DRPlanExecution YAML from a ChartAnalysis.
func GeneratePlan(analysis ChartAnalysis, config GenerateConfig) (*GenerateResult, error) {
	result := &GenerateResult{
		WorkflowYAMLs:  make(map[string][]byte),
		ExecutionYAMLs: make(map[string][]byte),
	}

	stages := buildStages(analysis, config)
	if err := generateWorkflows(analysis, config, result); err != nil {
		return nil, err
	}

	plan := buildPlanYAML(config, stages)
	planBytes, err := sigyaml.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("marshaling DRPlan: %w", err)
	}
	result.PlanYAML = planBytes

	if err := generateExecutionSamples(config, result); err != nil {
		return nil, err
	}

	return result, nil
}

// buildStages emits a single logical stage when there is anything to execute.
// Keeping one stage preserves ordering guarantees while still allowing per-action DAG.
func buildStages(analysis ChartAnalysis, config GenerateConfig) []map[string]interface{} {
	if !hasAnyActions(analysis) {
		return nil
	}
	return []map[string]interface{}{
		{
			"name":        "install",
			"description": "Unified workflow for install/upgrade hooks and main resources",
			"workflows": []map[string]interface{}{
				{"workflowRef": map[string]interface{}{"name": fmt.Sprintf("%s-install", config.ReleaseName)}},
			},
		},
	}
}

// buildPlanYAML builds the top-level DRPlan document.
// It also injects feedNamespace as a global param for all generated actions.
func buildPlanYAML(config GenerateConfig, stages []map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
		"kind":       "DRPlan",
		"metadata": map[string]interface{}{
			"name":      config.ReleaseName,
			"namespace": config.Namespace,
		},
		"spec": map[string]interface{}{
			"description":   fmt.Sprintf("Auto-generated from Helm template: %s", config.ReleaseName),
			"failurePolicy": "Stop",
			"globalParams": []map[string]interface{}{
				{"name": "feedNamespace", "value": config.Namespace},
			},
			"stages": stages,
		},
	}
}

// generateWorkflows generates the unified install workflow YAML.
// The planner currently emits one workflow to cover install/upgrade/delete/rollback hooks.
func generateWorkflows(analysis ChartAnalysis, config GenerateConfig, result *GenerateResult) error {
	if !hasAnyActions(analysis) {
		return nil
	}

	wf := buildUnifiedWorkflow(analysis, config)
	wfBytes, err := sigyaml.Marshal(wf)
	if err != nil {
		return fmt.Errorf("marshaling install workflow: %w", err)
	}
	result.WorkflowYAMLs["workflow-install.yaml"] = wfBytes

	return nil
}

// buildUnifiedWorkflow assembles all hook and main-resource actions into one DAG-ready workflow.
// Hook groups are chained serially (stable Helm semantics), while cross-group dependencies are explicit.
func buildUnifiedWorkflow(analysis ChartAnalysis, config GenerateConfig) map[string]interface{} {
	actions := make([]map[string]interface{}, 0)
	var createDeps []string

	actions, installPreDeps := appendHookActionsSerial(actions, buildHookEntries(analysis.Hooks[HookPreInstall]), nil)
	actions, upgradePreDeps := appendHookActionsSerial(actions, buildHookEntries(analysis.Hooks[HookPreUpgrade]), nil)
	createDeps = appendUnique(createDeps, installPreDeps...)
	createDeps = appendUnique(createDeps, upgradePreDeps...)

	if len(analysis.MainResources) > 0 {
		mainAction := buildInstallAction(analysis.MainResources, config)
		if len(createDeps) > 0 {
			mainAction["dependsOn"] = createDeps
		}
		actions = append(actions, mainAction)
	}

	postInstallDeps := []string{}
	postUpgradeDeps := []string{}
	if len(analysis.MainResources) > 0 {
		postInstallDeps = append(postInstallDeps, mainActionName)
		postUpgradeDeps = append(postUpgradeDeps, mainActionName)
	}
	actions, _ = appendHookActionsSerial(actions, buildHookEntries(analysis.Hooks[HookPostInstall]), postInstallDeps)
	actions, _ = appendHookActionsSerial(actions, buildHookEntries(analysis.Hooks[HookPostUpgrade]), postUpgradeDeps)

	deletePreDeps := []string{}
	actions, deletePreDeps = appendHookActionsSerial(actions, buildHookEntries(analysis.Hooks[HookPreDelete]), deletePreDeps)
	if len(analysis.MainResources) > 0 {
		deleteAction := buildDeleteAction(config)
		if len(deletePreDeps) > 0 {
			deleteAction["dependsOn"] = deletePreDeps
		}
		actions = append(actions, deleteAction)
		deletePreDeps = []string{deleteActionName}
	}
	actions, _ = appendHookActionsSerial(actions, buildHookEntries(analysis.Hooks[HookPostDelete]), deletePreDeps)

	actions, _ = appendHookActionsSerial(actions, buildHookEntries(analysis.Hooks[HookPreRollback]), nil)
	actions, _ = appendHookActionsSerial(actions, buildHookEntries(analysis.Hooks[HookPostRollback]), nil)

	return map[string]interface{}{
		"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
		"kind":       "DRWorkflow",
		"metadata": map[string]interface{}{
			"name":      fmt.Sprintf("%s-install", config.ReleaseName),
			"namespace": config.Namespace,
		},
		"spec": map[string]interface{}{
			"failurePolicy": "FailFast",
			"parameters": []map[string]interface{}{
				{"name": "feedNamespace", "type": "string", "default": config.Namespace},
			},
			"actions": actions,
		},
	}
}

// plannedHook holds a single hook resource with its resolved when condition.
type plannedHook struct {
	hook HookResource
	when string
}

// buildHookEntries normalizes hooks into plannedHook and sorts by (weight, name),
// matching Helm's deterministic hook ordering behavior.
func buildHookEntries(hooks []HookResource) []plannedHook {
	if len(hooks) == 0 {
		return nil
	}
	result := make([]plannedHook, 0, len(hooks))
	for _, h := range hooks {
		result = append(result, plannedHook{
			hook: h,
			when: whenForHookType(h.HookType),
		})
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].hook.Weight != result[j].hook.Weight {
			return result[i].hook.Weight < result[j].hook.Weight
		}
		return result[i].hook.Resource.GetName() < result[j].hook.Resource.GetName()
	})
	return result
}

// whenForHookType maps each Helm hook type to the runtime mode expression.
// Unknown hook types are left without a when guard.
func whenForHookType(hookType string) string {
	switch hookType {
	case HookPreInstall, HookPostInstall:
		return `mode == "install"`
	case HookPreUpgrade, HookPostUpgrade:
		return `mode == "upgrade"`
	case HookPreDelete, HookPostDelete:
		return `mode == "delete"`
	case HookPreRollback, HookPostRollback:
		return `mode == "rollback"`
	default:
		return ""
	}
}

// buildInstallAction creates the primary Subscription action for install/upgrade modes.
// Main resources are converted to feeds and applied through Subscription operation=Apply.
func buildInstallAction(resources []MainResource, config GenerateConfig) map[string]interface{} {
	var feeds []map[string]interface{}
	for _, res := range resources {
		feed := map[string]interface{}{
			"apiVersion": res.Resource.GetAPIVersion(),
			"kind":       res.Resource.GetKind(),
			"name":       res.Resource.GetName(),
		}
		if res.Resource.GetNamespace() != "" {
			feed["namespace"] = feedNamespaceRef
		}
		feeds = append(feeds, feed)
	}

	return map[string]interface{}{
		"name":    mainActionName,
		"type":    "Subscription",
		"timeout": "5m",
		"when":    `mode == "install" || mode == "upgrade"`,
		"subscription": map[string]interface{}{
			"operation": "Apply",
			"name":      fmt.Sprintf("%s-subscription", config.ReleaseName),
			"namespace": feedNamespaceRef,
			"spec": map[string]interface{}{
				"schedulingStrategy": "Replication",
				"feeds":              feeds,
				"subscribers": []map[string]interface{}{
					{"clusterAffinity": map[string]interface{}{}},
				},
			},
		},
	}
}

// buildDeleteAction creates the main-resource cleanup action for delete mode.
func buildDeleteAction(config GenerateConfig) map[string]interface{} {
	return map[string]interface{}{
		"name": deleteActionName,
		"type": "Subscription",
		"when": `mode == "delete"`,
		"subscription": map[string]interface{}{
			"operation": "Delete",
			"name":      fmt.Sprintf("%s-subscription", config.ReleaseName),
			"namespace": feedNamespaceRef,
		},
	}
}

// appendHookActionsSerial appends hooks in a single explicit dependsOn chain.
// The first hook depends on prevDeps; each subsequent hook depends on the
// immediately preceding hook, matching Helm's stable serial execution model.
func appendHookActionsSerial(
	actions []map[string]interface{},
	hooks []plannedHook,
	prevDeps []string,
) ([]map[string]interface{}, []string) {
	if len(hooks) == 0 {
		return actions, prevDeps
	}

	currentDeps := append([]string(nil), prevDeps...)
	lastActionName := ""
	for _, ph := range hooks {
		action := buildHookAction(ph)
		if len(currentDeps) > 0 {
			action["dependsOn"] = currentDeps
		}
		actions = append(actions, action)
		lastActionName = hookActionName(ph)
		currentDeps = []string{lastActionName}
	}
	return actions, []string{lastActionName}
}

// buildHookAction converts one hook resource to a per-cluster subscription action
// with waitReady and cleanup policy settings.
func buildHookAction(ph plannedHook) map[string]interface{} {
	res := ph.hook.Resource
	feed := map[string]interface{}{
		"apiVersion": res.GetAPIVersion(),
		"kind":       res.GetKind(),
		"name":       res.GetName(),
	}
	if res.GetNamespace() != "" {
		feed["namespace"] = feedNamespaceRef
	}
	action := map[string]interface{}{
		"name":    hookActionName(ph),
		"type":    "Subscription",
		"timeout": "5m",
		"subscription": map[string]interface{}{
			"operation": "Create",
			"name":      fmt.Sprintf("%s-sub", res.GetName()),
			"namespace": feedNamespaceRef,
			"spec": map[string]interface{}{
				"schedulingStrategy": "Replication",
				"feeds":              []map[string]interface{}{feed},
				"subscribers": []map[string]interface{}{
					{"clusterAffinity": map[string]interface{}{}},
				},
			},
		},
		"waitReady":            true,
		"clusterExecutionMode": "PerCluster",
		"hookType":             ph.hook.HookType,
		"hookCleanup":          hookCleanupFor(ph.hook),
	}
	if ph.when != "" {
		action["when"] = ph.when
	}
	return action
}

// hookActionName returns a stable action name: "<resource>-<normalized-hook-type>".
func hookActionName(ph plannedHook) string {
	return fmt.Sprintf("%s-%s", ph.hook.Resource.GetName(), normalizeHookTypeForName(ph.hook.HookType))
}

// hookCleanupFor translates Helm hook-delete-policy into hookCleanup toggles.
// beforeCreate defaults to true to avoid stale hook subscriptions.
func hookCleanupFor(h HookResource) map[string]interface{} {
	policyTokens := parseHookDeletePolicy(h.DeletePolicy)
	cleanup := map[string]interface{}{
		"beforeCreate": true,
		"onSuccess":    false,
		"onFailure":    false,
	}

	if _, ok := policyTokens[DeletePolicyBeforeHookCreation]; ok {
		cleanup["beforeCreate"] = true
	}
	if _, ok := policyTokens[DeletePolicyHookSucceeded]; ok {
		cleanup["onSuccess"] = true
	}
	if _, ok := policyTokens[DeletePolicyHookFailed]; ok {
		cleanup["onFailure"] = true
	}

	return cleanup
}

// parseHookDeletePolicy parses a comma-separated Helm hook-delete-policy string.
func parseHookDeletePolicy(policy string) map[string]struct{} {
	if policy == "" {
		return nil
	}
	parts := strings.Split(policy, ",")
	result := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		result[token] = struct{}{}
	}
	return result
}

// appendUnique appends non-empty strings that are not already present in dst.
func appendUnique(dst []string, items ...string) []string {
	seen := make(map[string]struct{}, len(dst))
	for _, item := range dst {
		seen[item] = struct{}{}
	}
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		dst = append(dst, item)
	}
	return dst
}

// normalizeHookTypeForName converts hook constants to readable name suffixes.
func normalizeHookTypeForName(hookType string) string {
	switch hookType {
	case HookPreInstall:
		return "pre-install"
	case HookPreUpgrade:
		return "pre-upgrade"
	case HookPostInstall:
		return "post-install"
	case HookPostUpgrade:
		return "post-upgrade"
	case HookPreDelete:
		return "pre-delete"
	case HookPostDelete:
		return "post-delete"
	case HookPreRollback:
		return "pre-rollback"
	case HookPostRollback:
		return "post-rollback"
	default:
		return hookType
	}
}

// hasAnyActions reports whether analysis contains either main resources or hooks.
func hasAnyActions(analysis ChartAnalysis) bool {
	if len(analysis.MainResources) > 0 {
		return true
	}
	for _, hooks := range analysis.Hooks {
		if len(hooks) > 0 {
			return true
		}
	}
	return false
}

// generateExecutionSamples emits example DRPlanExecution manifests for each mode.
// These files are intended as operator-facing quick-start templates.
func generateExecutionSamples(config GenerateConfig, result *GenerateResult) error {
	samples := []struct {
		filename string
		data     map[string]interface{}
	}{
		{
			filename: "drplanexecution-install.yaml",
			data: map[string]interface{}{
				"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
				"kind":       "DRPlanExecution",
				"metadata":   map[string]interface{}{"name": fmt.Sprintf("%s-install-001", config.ReleaseName), "namespace": config.Namespace},
				"spec":       map[string]interface{}{"planRef": config.ReleaseName, "operationType": "Execute", "mode": "Install"},
			},
		},
		{
			filename: "drplanexecution-upgrade.yaml",
			data: map[string]interface{}{
				"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
				"kind":       "DRPlanExecution",
				"metadata":   map[string]interface{}{"name": fmt.Sprintf("%s-upgrade-001", config.ReleaseName), "namespace": config.Namespace},
				"spec":       map[string]interface{}{"planRef": config.ReleaseName, "operationType": "Execute", "mode": "Upgrade"},
			},
		},
		{
			filename: "drplanexecution-delete.yaml",
			data: map[string]interface{}{
				"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
				"kind":       "DRPlanExecution",
				"metadata":   map[string]interface{}{"name": fmt.Sprintf("%s-delete-001", config.ReleaseName), "namespace": config.Namespace},
				"spec":       map[string]interface{}{"planRef": config.ReleaseName, "operationType": "Execute", "mode": "Delete"},
			},
		},
		{
			filename: "drplanexecution-revert.yaml",
			data: map[string]interface{}{
				"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
				"kind":       "DRPlanExecution",
				"metadata":   map[string]interface{}{"name": fmt.Sprintf("%s-revert-001", config.ReleaseName), "namespace": config.Namespace},
				"spec": map[string]interface{}{
					"planRef":            config.ReleaseName,
					"operationType":      "Revert",
					"mode":               "Rollback",
					"revertExecutionRef": fmt.Sprintf("%s-install-001", config.ReleaseName),
				},
			},
		},
	}

	for _, s := range samples {
		b, err := sigyaml.Marshal(s.data)
		if err != nil {
			return fmt.Errorf("marshaling %s: %w", s.filename, err)
		}
		result.ExecutionYAMLs[s.filename] = b
	}
	return nil
}
