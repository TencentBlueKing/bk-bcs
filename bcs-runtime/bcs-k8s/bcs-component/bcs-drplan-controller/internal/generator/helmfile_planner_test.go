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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sigyaml "sigs.k8s.io/yaml"
)

var _ = Describe("GenerateHelmfilePlan", func() {
	It("should generate HelmChart, Globalization and Subscription actions when values exist", func() {
		release := HelmfileResolvedRelease{
			ReleaseName:     "nginx-demo",
			Namespace:       "default",
			TargetNamespace: "default",
			Chart:           "nginx",
			ChartVersion:    "0.1.1",
			ChartRepo:       "oci://registry.example.com/charts",
			ValuesYAML:      "replicaCount: 2\n",
			Atomic:          boolPtr(true),
			PlainHTTP:       boolPtr(true),
		}

		result, err := GenerateHelmfilePlan(release)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.WorkflowYAMLs).To(HaveKey("workflow-execute.yaml"))

		wf := string(result.WorkflowYAMLs["workflow-execute.yaml"])
		plan := string(result.PlanYAML)
		Expect(wf).To(ContainSubstring("type: HelmChart"))
		Expect(wf).To(ContainSubstring("type: Globalization"))
		Expect(wf).To(ContainSubstring("type: Subscription"))
		Expect(wf).To(ContainSubstring("repo: oci://registry.example.com/charts"))
		Expect(wf).To(ContainSubstring("atomic: true"))
		Expect(wf).To(ContainSubstring("upgradeAtomic: true"))
		Expect(wf).To(ContainSubstring("plainHTTP: true"))
		Expect(wf).To(ContainSubstring("createNamespace: true"))
		Expect(wf).To(ContainSubstring("feedNamespace"))
		Expect(wf).To(ContainSubstring("targetNamespace"))
		Expect(wf).To(ContainSubstring("namespace: $(params.feedNamespace)"))
		Expect(wf).To(ContainSubstring("targetNamespace: $(params.targetNamespace)"))
		Expect(wf).NotTo(ContainSubstring("dependsOn:"))
		Expect(wf).NotTo(ContainSubstring("rollback:"))
		Expect(wf).NotTo(ContainSubstring("operation: Delete"))
		Expect(wf).NotTo(ContainSubstring("mode == \"install\""))
		Expect(wf).NotTo(ContainSubstring("mode == \"upgrade\""))
		Expect(wf).NotTo(ContainSubstring("mode == \"delete\""))
		Expect(plan).To(ContainSubstring("name: feedNamespace"))
		Expect(plan).To(ContainSubstring("value: default"))
		Expect(plan).To(ContainSubstring("name: targetNamespace"))
	})

	It("should skip Globalization action when values are empty", func() {
		release := HelmfileResolvedRelease{
			ReleaseName:     "nginx-demo",
			Namespace:       "default",
			TargetNamespace: "default",
			Chart:           "nginx",
			ChartVersion:    "0.1.1",
			ChartRepo:       "oci://registry.example.com/charts",
		}

		result, err := GenerateHelmfilePlan(release)
		Expect(err).NotTo(HaveOccurred())

		wf := string(result.WorkflowYAMLs["workflow-execute.yaml"])
		Expect(wf).To(ContainSubstring("type: HelmChart"))
		Expect(wf).NotTo(ContainSubstring("type: Globalization"))
		Expect(wf).To(ContainSubstring("type: Subscription"))
		Expect(wf).To(ContainSubstring("namespace: $(params.feedNamespace)"))
		Expect(wf).To(ContainSubstring("targetNamespace: $(params.targetNamespace)"))
		Expect(wf).NotTo(ContainSubstring("dependsOn:"))
		Expect(wf).NotTo(ContainSubstring("rollback:"))
		Expect(wf).NotTo(ContainSubstring("operation: Delete"))
		Expect(wf).NotTo(ContainSubstring("mode == \"install\""))
		Expect(wf).NotTo(ContainSubstring("mode == \"upgrade\""))
		Expect(wf).NotTo(ContainSubstring("mode == \"delete\""))
	})

	It("should preserve explicit createNamespace false", func() {
		release := HelmfileResolvedRelease{
			ReleaseName:     "nginx-demo",
			Namespace:       "default",
			TargetNamespace: "default",
			Chart:           "nginx",
			ChartVersion:    "0.1.1",
			ChartRepo:       "oci://registry.example.com/charts",
			CreateNamespace: boolPtr(false),
		}

		result, err := GenerateHelmfilePlan(release)
		Expect(err).NotTo(HaveOccurred())

		wf := string(result.WorkflowYAMLs["workflow-execute.yaml"])
		Expect(wf).To(ContainSubstring("createNamespace: false"))
		Expect(wf).NotTo(ContainSubstring("createNamespace: true"))
	})

	It("should preserve explicit createNamespace true", func() {
		release := HelmfileResolvedRelease{
			ReleaseName:     "nginx-demo",
			Namespace:       "default",
			TargetNamespace: "default",
			Chart:           "nginx",
			ChartVersion:    "0.1.1",
			ChartRepo:       "oci://registry.example.com/charts",
			CreateNamespace: boolPtr(true),
		}

		result, err := GenerateHelmfilePlan(release)
		Expect(err).NotTo(HaveOccurred())

		wf := string(result.WorkflowYAMLs["workflow-execute.yaml"])
		Expect(wf).To(ContainSubstring("createNamespace: true"))
		Expect(wf).NotTo(ContainSubstring("createNamespace: false"))
	})
})

func boolPtr(v bool) *bool {
	return &v
}

func TestGenerateHelmfilePlan_WithHooks_UsesMultiStageWorkflows(t *testing.T) {
	t.Helper()

	release := HelmfileResolvedRelease{
		ReleaseName:     "demo-app",
		Namespace:       "default",
		TargetNamespace: "demo-target",
		Chart:           "demo",
		ChartRepo:       "oci://registry.example.com/charts",
		HookImage:       "registry.example.com/hook-runner:v1",
		Hooks: []HelmfileResolvedHook{
			{Event: "preapply", Command: "/hooks/preapply.sh", Order: 0},
			{Event: "presync", Command: "/hooks/presync.sh", Order: 1},
			{Event: "postsync", Command: "/hooks/postsync.sh", Order: 2},
		},
	}

	result, err := GenerateHelmfilePlan(release)
	if err != nil {
		t.Fatalf("GenerateHelmfilePlan() error = %v", err)
	}

	if len(result.WorkflowYAMLs) != 4 {
		t.Fatalf("expected 4 workflow files, got %d", len(result.WorkflowYAMLs))
	}
	mustHaveKey(t, result.WorkflowYAMLs, "workflow-preapply.yaml")
	mustHaveKey(t, result.WorkflowYAMLs, "workflow-presync.yaml")
	mustHaveKey(t, result.WorkflowYAMLs, "workflow-execute.yaml")
	mustHaveKey(t, result.WorkflowYAMLs, "workflow-postsync.yaml")

	plan := decodeYAMLMap(t, result.PlanYAML)
	spec := mustMap(t, plan["spec"], "plan.spec")
	assertEqual(t, spec["failurePolicy"], "Continue", "plan.spec.failurePolicy")
	globalParams := mustSlice(t, spec["globalParams"], "plan.spec.globalParams")
	assertGlobalParamValue(t, globalParams, "hookImage", "registry.example.com/hook-runner:v1")
	stages := mustSlice(t, spec["stages"], "plan.spec.stages")
	assertStageOrder(t, stages, []string{"preapply", "presync", "execute", "postsync"})

	execute := decodeYAMLMap(t, result.WorkflowYAMLs["workflow-execute.yaml"])
	executeSpec := mustMap(t, execute["spec"], "workflow-execute.spec")
	executeActions := mustSlice(t, executeSpec["actions"], "workflow-execute.spec.actions")
	assertActionTypes(t, executeActions, []string{"HelmChart", "Subscription"})

	postsync := decodeYAMLMap(t, result.WorkflowYAMLs["workflow-postsync.yaml"])
	postsyncSpec := mustMap(t, postsync["spec"], "workflow-postsync.spec")
	assertEqual(t, postsyncSpec["failurePolicy"], "Continue", "workflow-postsync.spec.failurePolicy")
}

func TestGenerateHelmfilePlan_HookWorkflowUsesManifestAndPerClusterSubscription(t *testing.T) {
	t.Helper()

	release := HelmfileResolvedRelease{
		ReleaseName:     "demo-app",
		Namespace:       "default",
		TargetNamespace: "demo-target",
		Chart:           "demo",
		ChartRepo:       "oci://registry.example.com/charts",
		HookImage:       "registry.example.com/hook-runner:v1",
		Hooks: []HelmfileResolvedHook{
			{
				Event:   "preapply",
				Command: "/bin/sh",
				Args:    []string{"-c", "echo preparing"},
				Order:   0,
			},
		},
	}

	result, err := GenerateHelmfilePlan(release)
	if err != nil {
		t.Fatalf("GenerateHelmfilePlan() error = %v", err)
	}

	wf := decodeYAMLMap(t, result.WorkflowYAMLs["workflow-preapply.yaml"])
	spec := mustMap(t, wf["spec"], "workflow-preapply.spec")
	actions := mustSlice(t, spec["actions"], "workflow-preapply.spec.actions")
	assertActionTypes(t, actions, []string{"KubernetesResource", "Subscription"})

	manifestAction := mustMap(t, actions[0], "workflow-preapply.spec.actions[0]")
	resource := mustMap(t, manifestAction["resource"], "workflow-preapply.spec.actions[0].resource")
	assertEqual(t, resource["operation"], "Apply", "manifest action operation")
	manifestYAML, ok := resource["manifest"].(string)
	if !ok {
		t.Fatalf("expected manifest to be string, got %T", resource["manifest"])
	}
	manifest := decodeYAMLStringMap(t, manifestYAML)
	assertEqual(t, manifest["kind"], "Manifest", "manifest kind")
	if _, ok := manifest["spec"]; ok {
		t.Fatal("expected Clusternet Manifest to use top-level template, got spec")
	}
	manifestMetadata := mustMap(t, manifest["metadata"], "manifest.metadata")
	assertEqual(t, manifestMetadata["name"], "jobs.$(params.targetNamespace).demo-app-preapply-hook-0", "manifest metadata name")
	assertEqual(t, manifestMetadata["namespace"], "clusternet-reserved", "manifest metadata namespace")
	manifestLabels := mustMap(t, manifestMetadata["labels"], "manifest.metadata.labels")
	assertEqual(t, manifestLabels["apps.clusternet.io/config.group"], "batch", "manifest label config.group")
	assertEqual(t, manifestLabels["apps.clusternet.io/config.version"], "v1", "manifest label config.version")
	assertEqual(t, manifestLabels["apps.clusternet.io/config.kind"], "Job", "manifest label config.kind")
	assertEqual(t, manifestLabels["apps.clusternet.io/config.name"], "demo-app-preapply-hook-0", "manifest label config.name")
	assertEqual(t, manifestLabels["apps.clusternet.io/config.namespace"], "$(params.targetNamespace)", "manifest label config.namespace")
	job := mustMap(t, manifest["template"], "manifest.template")
	assertEqual(t, job["kind"], "Job", "job kind")
	jobMetadata := mustMap(t, job["metadata"], "job.metadata")
	assertEqual(t, jobMetadata["name"], "demo-app-preapply-hook-0", "job metadata name")
	assertEqual(t, jobMetadata["namespace"], "$(params.targetNamespace)", "job metadata namespace")
	jobSpec := mustMap(t, job["spec"], "job.spec")
	assertInt(t, jobSpec["ttlSecondsAfterFinished"], helmfileHookJobTTLSeconds, "job.spec.ttlSecondsAfterFinished")
	assertInt(t, jobSpec["backoffLimit"], 0, "job.spec.backoffLimit")
	template := mustMap(t, jobSpec["template"], "job.spec.template")
	podSpec := mustMap(t, template["spec"], "job.spec.template.spec")
	assertEqual(t, podSpec["restartPolicy"], "Never", "job.spec.template.spec.restartPolicy")
	containers := mustSlice(t, podSpec["containers"], "job.spec.template.spec.containers")
	container := mustMap(t, containers[0], "job.spec.template.spec.containers[0]")
	assertEqual(t, container["image"], "$(params.hookImage)", "job container image")

	subscriptionAction := mustMap(t, actions[1], "workflow-preapply.spec.actions[1]")
	assertEqual(t, subscriptionAction["clusterExecutionMode"], "PerCluster", "subscription clusterExecutionMode")
	assertEqual(t, subscriptionAction["waitReady"], true, "subscription waitReady")
	hookCleanup := mustMap(t, subscriptionAction["hookCleanup"], "subscription hookCleanup")
	assertEqual(t, hookCleanup["beforeCreate"], true, "subscription hookCleanup.beforeCreate")
	subscription := mustMap(t, subscriptionAction["subscription"], "subscription action payload")
	assertEqual(t, subscription["operation"], "Apply", "subscription operation")
	subscriptionSpec := mustMap(t, subscription["spec"], "subscription spec")
	feeds := mustSlice(t, subscriptionSpec["feeds"], "subscription spec feeds")
	feed := mustMap(t, feeds[0], "subscription spec feeds[0]")
	assertEqual(t, feed["apiVersion"], "batch/v1", "hook feed apiVersion")
	assertEqual(t, feed["kind"], "Job", "hook feed kind")
	assertEqual(t, feed["name"], "demo-app-preapply-hook-0", "hook feed name")
	assertEqual(t, feed["namespace"], "$(params.targetNamespace)", "hook feed namespace")
}

func TestBuildHelmfileHookManifest_ReturnsStructuredManifest(t *testing.T) {
	t.Helper()

	manifestYAML, err := buildHelmfileHookManifest(HelmfileResolvedRelease{
		ReleaseName:     "demo-app",
		Namespace:       "default",
		TargetNamespace: "demo-target",
		HookImage:       "registry.example.com/hook-runner:v1",
	}, HelmfileResolvedHook{
		Event:   "preapply",
		Command: "/bin/sh",
		Args:    []string{"-c", "echo preparing"},
		Order:   0,
	})
	if err != nil {
		t.Fatalf("buildHelmfileHookManifest() error = %v", err)
	}

	manifest := decodeYAMLStringMap(t, manifestYAML)
	assertEqual(t, manifest["kind"], "Manifest", "manifest kind")
	if _, ok := manifest["spec"]; ok {
		t.Fatal("expected Clusternet Manifest to use top-level template, got spec")
	}
	metadata := mustMap(t, manifest["metadata"], "manifest.metadata")
	assertEqual(t, metadata["name"], "jobs.$(params.targetNamespace).demo-app-preapply-hook-0", "manifest metadata name")
	assertEqual(t, metadata["namespace"], "clusternet-reserved", "manifest metadata namespace")
	job := mustMap(t, manifest["template"], "manifest.template")
	jobSpec := mustMap(t, job["spec"], "job.spec")
	assertInt(t, jobSpec["ttlSecondsAfterFinished"], helmfileHookJobTTLSeconds, "job ttlSecondsAfterFinished")
	template := mustMap(t, jobSpec["template"], "job.spec.template")
	podSpec := mustMap(t, template["spec"], "job.spec.template.spec")
	assertEqual(t, podSpec["restartPolicy"], "Never", "job restartPolicy")
	containers := mustSlice(t, podSpec["containers"], "job containers")
	container := mustMap(t, containers[0], "job container")
	assertEqual(t, container["image"], "$(params.hookImage)", "job container image")
}

func TestGenerateHelmfilePlan_WithOnlyPresyncHook_OmitsEmptyHookWorkflows(t *testing.T) {
	t.Helper()

	release := HelmfileResolvedRelease{
		ReleaseName:     "demo-app",
		Namespace:       "default",
		TargetNamespace: "demo-target",
		Chart:           "demo",
		ChartRepo:       "oci://registry.example.com/charts",
		HookImage:       "registry.example.com/hook-runner:v1",
		Hooks: []HelmfileResolvedHook{
			{Event: "presync", Command: "/hooks/presync.sh", Order: 0},
		},
	}

	result, err := GenerateHelmfilePlan(release)
	if err != nil {
		t.Fatalf("GenerateHelmfilePlan() error = %v", err)
	}

	if len(result.WorkflowYAMLs) != 2 {
		t.Fatalf("expected 2 workflow files, got %d", len(result.WorkflowYAMLs))
	}
	mustHaveKey(t, result.WorkflowYAMLs, "workflow-presync.yaml")
	mustHaveKey(t, result.WorkflowYAMLs, "workflow-execute.yaml")
	mustNotHaveKey(t, result.WorkflowYAMLs, "workflow-preapply.yaml")
	mustNotHaveKey(t, result.WorkflowYAMLs, "workflow-postsync.yaml")

	plan := decodeYAMLMap(t, result.PlanYAML)
	spec := mustMap(t, plan["spec"], "plan.spec")
	stages := mustSlice(t, spec["stages"], "plan.spec.stages")
	assertStageOrder(t, stages, []string{"presync", "execute"})
}

func mustHaveKey(t *testing.T, values map[string][]byte, key string) {
	t.Helper()
	if _, ok := values[key]; !ok {
		t.Fatalf("expected key %q to exist", key)
	}
}

func mustNotHaveKey(t *testing.T, values map[string][]byte, key string) {
	t.Helper()
	if _, ok := values[key]; ok {
		t.Fatalf("expected key %q not to exist", key)
	}
}

func decodeYAMLMap(t *testing.T, data []byte) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := sigyaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}
	return result
}

func decodeYAMLStringMap(t *testing.T, data string) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := sigyaml.Unmarshal([]byte(data), &result); err != nil {
		t.Fatalf("failed to unmarshal yaml string: %v", err)
	}
	return result
}

func mustMap(t *testing.T, value interface{}, field string) map[string]interface{} {
	t.Helper()
	result, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("expected %s to be map, got %T", field, value)
	}
	return result
}

func mustSlice(t *testing.T, value interface{}, field string) []interface{} {
	t.Helper()
	result, ok := value.([]interface{})
	if !ok {
		t.Fatalf("expected %s to be slice, got %T", field, value)
	}
	return result
}

func assertEqual(t *testing.T, got interface{}, want interface{}, field string) {
	t.Helper()
	if got != want {
		t.Fatalf("expected %s = %#v, got %#v", field, want, got)
	}
}

func assertInt(t *testing.T, got interface{}, want int, field string) {
	t.Helper()
	switch v := got.(type) {
	case int:
		if v != want {
			t.Fatalf("expected %s = %d, got %d", field, want, v)
		}
	case int32:
		if int(v) != want {
			t.Fatalf("expected %s = %d, got %d", field, want, v)
		}
	case int64:
		if int(v) != want {
			t.Fatalf("expected %s = %d, got %d", field, want, v)
		}
	case float64:
		if int(v) != want {
			t.Fatalf("expected %s = %d, got %v", field, want, v)
		}
	default:
		t.Fatalf("expected %s to be numeric, got %T", field, got)
	}
}

func assertGlobalParamValue(t *testing.T, params []interface{}, name, wantValue string) {
	t.Helper()
	for _, item := range params {
		param := mustMap(t, item, "global param")
		if param["name"] == name {
			assertEqual(t, param["value"], wantValue, "global param value")
			return
		}
	}
	t.Fatalf("global param %q not found", name)
}

func assertStageOrder(t *testing.T, stages []interface{}, want []string) {
	t.Helper()
	if len(stages) != len(want) {
		t.Fatalf("expected %d stages, got %d", len(want), len(stages))
	}
	for i, item := range stages {
		stage := mustMap(t, item, "stage")
		assertEqual(t, stage["name"], want[i], "stage name")
	}
}

func assertActionTypes(t *testing.T, actions []interface{}, want []string) {
	t.Helper()
	if len(actions) != len(want) {
		t.Fatalf("expected %d actions, got %d", len(want), len(actions))
	}
	for i, item := range actions {
		action := mustMap(t, item, "action")
		assertEqual(t, action["type"], want[i], "action type")
	}
}
