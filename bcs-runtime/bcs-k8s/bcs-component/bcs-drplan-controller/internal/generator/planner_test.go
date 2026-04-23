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
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sigyaml "sigs.k8s.io/yaml"
)

var _ = Describe("GeneratePlan", func() {
	var config GenerateConfig

	BeforeEach(func() {
		config = GenerateConfig{
			ReleaseName: "my-app",
			Namespace:   "default",
			OutputDir:   "/tmp/test-output",
		}
	})

	Context("with only main resources (no hooks)", func() {
		It("should generate a single execute stage", func() {
			analysis := ChartAnalysis{
				Hooks: make(map[string][]HookResource),
				MainResources: []MainResource{
					{Resource: makeResource("ConfigMap", "cm-1", nil)},
					{Resource: makeResource("Deployment", "deploy-1", nil)},
				},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.PlanYAML).NotTo(BeEmpty())
			Expect(result.WorkflowYAMLs).To(HaveKey("workflow-execute.yaml"))
			Expect(result.WorkflowYAMLs).To(HaveLen(1))

			planStr := string(result.PlanYAML)
			Expect(planStr).To(ContainSubstring("name: execute"))
			Expect(planStr).NotTo(ContainSubstring("dependsOn"))
		})
	})

	Context("with pre-install and post-install hooks", func() {
		It("should generate one unified stage and one workflow", func() {
			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreInstall: {
						{
							Resource: makeResource("Job", "db-migrate", map[string]string{
								HookAnnotation: HookPreInstall,
							}),
							HookType: HookPreInstall,
							Weight:   -5,
						},
					},
					HookPostInstall: {
						{
							Resource: makeResource("Job", "health-check", map[string]string{
								HookAnnotation: HookPostInstall,
							}),
							HookType:     HookPostInstall,
							DeletePolicy: DeletePolicyHookSucceeded,
						},
					},
				},
				MainResources: []MainResource{
					{Resource: makeResource("Deployment", "app", nil)},
				},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			planStr := string(result.PlanYAML)
			Expect(planStr).To(ContainSubstring("name: install"))
			Expect(planStr).NotTo(ContainSubstring("name: pre-install"))
			Expect(planStr).NotTo(ContainSubstring("name: post-install"))
			Expect(planStr).NotTo(ContainSubstring("dependsOn"))
			Expect(result.WorkflowYAMLs).To(HaveKey("workflow-install.yaml"))
			Expect(result.WorkflowYAMLs).To(HaveLen(1))
		})
	})

	Context("with hook actions", func() {
		It("should set waitReady: true and clusterExecutionMode: PerCluster on hook actions", func() {
			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreInstall: {
						{
							Resource: makeResource("Job", "pre-job", map[string]string{
								HookAnnotation: HookPreInstall,
							}),
							HookType: HookPreInstall,
						},
					},
				},
				MainResources: []MainResource{
					{Resource: makeResource("Deployment", "app", nil)},
				},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			workflowYAML := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(workflowYAML).To(ContainSubstring("waitReady: true"))
			Expect(workflowYAML).To(ContainSubstring("clusterExecutionMode: PerCluster"))
			Expect(workflowYAML).To(ContainSubstring(`when: mode == "install"`))
			Expect(workflowYAML).To(ContainSubstring("name: create-subscription"))
		})
	})

	Context("hook workflow actions", func() {
		It("should use Subscription type with hook Job as feed", func() {
			job := makeResource("Job", "db-migrate", map[string]string{
				HookAnnotation: HookPreInstall,
			})
			job.SetAPIVersion("batch/v1")
			job.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreInstall: {
						{Resource: job, HookType: HookPreInstall},
					},
				},
				MainResources: []MainResource{
					{Resource: makeResource("Deployment", "app", nil)},
				},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(wf).To(ContainSubstring("type: Subscription"))
			Expect(wf).To(ContainSubstring("kind: Job"))
			Expect(wf).To(ContainSubstring("name: db-migrate"))
			Expect(wf).To(ContainSubstring("db-migrate-sub"))
			Expect(wf).NotTo(ContainSubstring("type: Job"))
			Expect(wf).To(ContainSubstring("$(params.feedNamespace)"))
			Expect(wf).To(ContainSubstring("feedNamespace"))
			Expect(wf).To(ContainSubstring(`when: mode == "install"`))
		})
	})

	Context("upgrade hooks", func() {
		It("should set when expression to upgrade for pre-upgrade hooks", func() {
			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreUpgrade: {
						{
							Resource: makeResource("Job", "upgrade-job", map[string]string{
								HookAnnotation: HookPreUpgrade,
							}),
							HookType: HookPreUpgrade,
						},
					},
				},
				MainResources: []MainResource{
					{Resource: makeResource("Deployment", "app", nil)},
				},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(wf).To(ContainSubstring(`when: mode == "upgrade"`))
		})
	})

	Context("subscription feeds", func() {
		It("should include all main resources as feeds", func() {
			cm := makeResource("ConfigMap", "cm-1", nil)
			cm.SetAPIVersion("v1")
			cm.SetNamespace("default")

			deploy := makeResource("Deployment", "app", nil)
			deploy.SetAPIVersion("apps/v1")
			deploy.SetNamespace("default")

			svc := makeResource("Service", "svc-1", nil)
			svc.SetAPIVersion("v1")
			svc.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: make(map[string][]HookResource),
				MainResources: []MainResource{
					{Resource: cm},
					{Resource: deploy},
					{Resource: svc},
				},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			workflowYAML := string(result.WorkflowYAMLs["workflow-execute.yaml"])
			Expect(workflowYAML).To(ContainSubstring("cm-1"))
			Expect(workflowYAML).To(ContainSubstring("app"))
			Expect(workflowYAML).To(ContainSubstring("svc-1"))
			Expect(workflowYAML).NotTo(ContainSubstring(`when: mode == "install" || mode == "upgrade"`))
			Expect(workflowYAML).NotTo(ContainSubstring("operation: Delete"))
		})
	})

	It("should generate execution YAML samples", func() {
		analysis := ChartAnalysis{
			Hooks:         make(map[string][]HookResource),
			MainResources: []MainResource{{Resource: makeResource("ConfigMap", "cm-1", nil)}},
		}

		result, err := GeneratePlan(analysis, config)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.ExecutionYAMLs).To(HaveKey("drplanexecution-install.yaml"))
		Expect(result.ExecutionYAMLs).To(HaveKey("drplanexecution-delete.yaml"))
		Expect(result.ExecutionYAMLs).To(HaveKey("drplanexecution-revert.yaml"))
	})

	Context("with only hooks (no main resources)", func() {
		It("should generate unified install stage and workflow", func() {
			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreInstall: {
						{
							Resource: makeResource("Job", "db-init", map[string]string{
								HookAnnotation: HookPreInstall,
							}),
							HookType: HookPreInstall,
						},
					},
				},
				MainResources: nil,
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			planStr := string(result.PlanYAML)
			Expect(planStr).To(ContainSubstring("name: install"))
			Expect(result.WorkflowYAMLs).To(HaveKey("workflow-install.yaml"))
			Expect(result.WorkflowYAMLs).To(HaveLen(1))
		})
	})

	Context("with multiple pre-install hooks (different weights)", func() {
		It("should chain weight layers via dependsOn", func() {
			job1 := makeResource("Job", "migrate-1", map[string]string{
				HookAnnotation: HookPreInstall,
			})
			job1.SetNamespace("default")
			job2 := makeResource("Job", "migrate-2", map[string]string{
				HookAnnotation: HookPreInstall,
			})
			job2.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreInstall: {
						{Resource: job1, HookType: HookPreInstall, Weight: -5},
						{Resource: job2, HookType: HookPreInstall, Weight: 0},
					},
				},
				MainResources: []MainResource{
					{Resource: makeResource("ConfigMap", "cm-1", nil)},
				},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(wf).To(ContainSubstring("migrate-1"))
			Expect(wf).To(ContainSubstring("migrate-2"))
			Expect(wf).To(ContainSubstring("migrate-1-sub"))
			Expect(wf).To(ContainSubstring("migrate-2-sub"))

			By("migrate-2 (weight 0) depends on migrate-1 (weight -5)")
			Expect(wf).To(ContainSubstring("dependsOn"))
		})
	})

	Context("DRPlan metadata", func() {
		It("should use releaseName and namespace from config", func() {
			config.ReleaseName = "custom-release"
			config.Namespace = "staging"

			analysis := ChartAnalysis{
				Hooks:         make(map[string][]HookResource),
				MainResources: []MainResource{{Resource: makeResource("ConfigMap", "cm-1", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			planStr := string(result.PlanYAML)
			Expect(planStr).To(ContainSubstring("name: custom-release"))
			Expect(planStr).To(ContainSubstring("namespace: staging"))
			Expect(planStr).To(ContainSubstring("feedNamespace"))
		})
	})

	Context("globalParams in plan", func() {
		It("should include feedNamespace param with config namespace", func() {
			config.Namespace = "production"

			analysis := ChartAnalysis{
				Hooks:         make(map[string][]HookResource),
				MainResources: []MainResource{{Resource: makeResource("ConfigMap", "cm-1", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			planStr := string(result.PlanYAML)
			Expect(planStr).To(ContainSubstring("name: feedNamespace"))
			Expect(planStr).To(ContainSubstring("value: production"))
		})
	})

	Context("execution sample content", func() {
		It("should contain correct operationType values", func() {
			analysis := ChartAnalysis{
				Hooks:         make(map[string][]HookResource),
				MainResources: []MainResource{{Resource: makeResource("ConfigMap", "cm-1", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			execStr := string(result.ExecutionYAMLs["drplanexecution-install.yaml"])
			Expect(execStr).To(ContainSubstring("operationType: Execute"))
			Expect(execStr).To(ContainSubstring("mode: Install"))
			Expect(execStr).To(ContainSubstring("planRef"))

			revertStr := string(result.ExecutionYAMLs["drplanexecution-revert.yaml"])
			Expect(revertStr).To(ContainSubstring("operationType: Revert"))
			Expect(revertStr).To(ContainSubstring("revertExecutionRef"))
		})
	})

	Context("resources without namespace", func() {
		It("should include default namespace template in feed when resource has no namespace", func() {
			cm := makeResource("ConfigMap", "cm-no-ns", nil)

			analysis := ChartAnalysis{
				Hooks:         make(map[string][]HookResource),
				MainResources: []MainResource{{Resource: cm}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			workflowYAML := string(result.WorkflowYAMLs["workflow-execute.yaml"])
			Expect(workflowYAML).To(ContainSubstring("name: cm-no-ns"))
			Expect(workflowYAML).To(ContainSubstring("namespace: $(params.feedNamespace)"))
		})
	})

	Context("resources with explicit namespace", func() {
		It("should still use default namespace template in feed", func() {
			deploy := makeResource("Deployment", "web", nil)
			deploy.SetAPIVersion("apps/v1")
			deploy.SetNamespace("prod")

			analysis := ChartAnalysis{
				Hooks:         make(map[string][]HookResource),
				MainResources: []MainResource{{Resource: deploy}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			workflowYAML := string(result.WorkflowYAMLs["workflow-execute.yaml"])
			Expect(workflowYAML).To(ContainSubstring("name: web"))
			Expect(workflowYAML).To(ContainSubstring("namespace: $(params.feedNamespace)"))
		})
	})

	Context("cluster-scoped resources", func() {
		It("should not include namespace in main resource feed", func() {
			clusterRole := makeResource("ClusterRole", "read-all", nil)
			clusterRole.SetAPIVersion("rbac.authorization.k8s.io/v1")

			analysis := ChartAnalysis{
				Hooks:         make(map[string][]HookResource),
				MainResources: []MainResource{{Resource: clusterRole}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			feed := extractFirstFeedForAction(result.WorkflowYAMLs["workflow-execute.yaml"], mainActionName)
			Expect(feed).To(HaveKeyWithValue("kind", "ClusterRole"))
			Expect(feed).NotTo(HaveKey("namespace"))
		})

		It("should not include namespace in hook resource feed", func() {
			clusterRole := makeResource("ClusterRole", "bootstrap-role", map[string]string{
				HookAnnotation: HookPreInstall,
			})
			clusterRole.SetAPIVersion("rbac.authorization.k8s.io/v1")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreInstall: {{Resource: clusterRole, HookType: HookPreInstall}},
				},
				MainResources: []MainResource{{Resource: makeResource("Deployment", "app", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			feed := extractFirstFeedForAction(result.WorkflowYAMLs["workflow-install.yaml"], "bootstrap-role-pre-install")
			Expect(feed).To(HaveKeyWithValue("kind", "ClusterRole"))
			Expect(feed).NotTo(HaveKey("namespace"))
		})
	})

	Context("weight-based DAG generation", func() {
		It("should generate dependsOn chain: pre-hooks → main → post-hooks", func() {
			preJob := makeResource("Job", "pre-job", nil)
			preJob.SetNamespace("default")
			postJob := makeResource("Job", "post-job", nil)
			postJob.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreInstall:  {{Resource: preJob, HookType: HookPreInstall, Weight: 0}},
					HookPostInstall: {{Resource: postJob, HookType: HookPostInstall, Weight: 0}},
				},
				MainResources: []MainResource{{Resource: makeResource("Deployment", "app", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])

			By("main resource depends on pre-hook")
			Expect(wf).To(MatchRegexp(`dependsOn:\n\s*- pre-job-pre-install\n\s*name: create-subscription`))

			By("post-hook depends on main resource")
			Expect(wf).To(ContainSubstring("name: post-job-post-install"))
			Expect(wf).To(ContainSubstring("dependsOn:\n    - create-subscription"))
		})

		It("should serialize same-weight hooks into a stable chain", func() {
			hookA := makeResource("Job", "hook-a", nil)
			hookA.SetNamespace("default")
			hookB := makeResource("Job", "hook-b", nil)
			hookB.SetNamespace("default")
			hookC := makeResource("Job", "hook-c", nil)
			hookC.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPostInstall: {
						{Resource: hookA, HookType: HookPostInstall, Weight: 0},
						{Resource: hookB, HookType: HookPostInstall, Weight: 0},
						{Resource: hookC, HookType: HookPostInstall, Weight: 10},
					},
				},
				MainResources: []MainResource{{Resource: makeResource("Deployment", "app", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])

			By("hook-a depends on create-subscription")
			Expect(wf).To(ContainSubstring("name: hook-a-post-install"))
			Expect(wf).To(ContainSubstring("dependsOn:\n    - create-subscription"))

			By("hook-b depends on hook-a, even with the same weight")
			Expect(wf).To(ContainSubstring("name: hook-b-post-install"))
			Expect(wf).To(ContainSubstring("dependsOn:\n    - hook-a-post-install"))

			By("hook-c depends on hook-b")
			Expect(wf).To(ContainSubstring("name: hook-c-post-install"))
			Expect(wf).To(ContainSubstring("dependsOn:\n    - hook-b-post-install"))
		})

		It("should not generate dependsOn for a single action workflow", func() {
			analysis := ChartAnalysis{
				Hooks:         make(map[string][]HookResource),
				MainResources: []MainResource{{Resource: makeResource("ConfigMap", "cm-1", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-execute.yaml"])
			Expect(wf).NotTo(ContainSubstring("dependsOn"))
		})
	})

	Context("main resource operation", func() {
		It("should use Apply operation without explicit when/delete in simplified mode", func() {
			analysis := ChartAnalysis{
				Hooks:         make(map[string][]HookResource),
				MainResources: []MainResource{{Resource: makeResource("Deployment", "app", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-execute.yaml"])
			Expect(wf).To(ContainSubstring("operation: Apply"))
			Expect(wf).NotTo(ContainSubstring(`when: mode == "install" || mode == "upgrade"`))
			Expect(wf).NotTo(ContainSubstring("operation: Delete"))
		})
	})

	Context("install and upgrade hooks on the same resource", func() {
		It("should generate separate actions and preserve independent lifecycle when/weight", func() {
			job := makeResource("Job", "auth-register", map[string]string{
				HookAnnotation: "post-install, post-upgrade",
			})
			job.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPostInstall: {{Resource: job, HookType: HookPostInstall, Weight: -4}},
					HookPostUpgrade: {{Resource: job, HookType: HookPostUpgrade, Weight: -4}},
				},
				MainResources: []MainResource{{Resource: makeResource("Deployment", "app", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(wf).To(ContainSubstring("name: auth-register-post-install"))
			Expect(wf).To(ContainSubstring("name: auth-register-post-upgrade"))
			Expect(wf).To(ContainSubstring(`when: mode == "install"`))
			Expect(wf).To(ContainSubstring(`when: mode == "upgrade"`))
			Expect(wf).To(ContainSubstring("name: auth-register-post-install"))
			Expect(wf).To(ContainSubstring("name: auth-register-post-upgrade"))
			Expect(strings.Count(wf, "dependsOn:\n    - create-subscription")).To(BeNumerically(">=", 2))
		})

		It("should keep when for hooks only in one mode", func() {
			installOnly := makeResource("Job", "install-only-job", nil)
			installOnly.SetNamespace("default")
			upgradeOnly := makeResource("Job", "upgrade-only-job", nil)
			upgradeOnly.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPostInstall: {{Resource: installOnly, HookType: HookPostInstall, Weight: 0}},
					HookPostUpgrade: {{Resource: upgradeOnly, HookType: HookPostUpgrade, Weight: 0}},
				},
				MainResources: []MainResource{{Resource: makeResource("Deployment", "app", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(wf).To(ContainSubstring("install-only-job"))
			Expect(wf).To(ContainSubstring("upgrade-only-job"))
			Expect(wf).To(ContainSubstring(`mode == "install"`))
			Expect(wf).To(ContainSubstring(`mode == "upgrade"`))
		})
	})

	Context("same resource across hook positions", func() {
		It("should generate unique action names for pre and post hooks on the same resource", func() {
			job := makeResource("Job", "shared-hook", map[string]string{
				HookAnnotation: "pre-install, post-install",
			})
			job.SetAPIVersion("batch/v1")
			job.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreInstall: {
						{Resource: job, HookType: HookPreInstall},
					},
					HookPostInstall: {
						{Resource: job, HookType: HookPostInstall},
					},
				},
				MainResources: []MainResource{
					{Resource: makeResource("Deployment", "app", nil)},
				},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(wf).To(ContainSubstring("name: shared-hook-pre-install"))
			Expect(wf).To(ContainSubstring("name: shared-hook-post-install"))
			Expect(wf).To(ContainSubstring("- shared-hook-pre-install"))
		})
	})

	Context("hook operation type", func() {
		It("should generate Create operation plus structured hook cleanup policy", func() {
			job1 := makeResource("Job", "db-migrate", nil)
			job1.SetNamespace("default")
			job2 := makeResource("Job", "health-check", nil)
			job2.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreInstall: {{
						Resource:     job1,
						HookType:     HookPreInstall,
						Weight:       -5,
						DeletePolicy: DeletePolicyBeforeHookCreation,
					}},
					HookPostInstall: {{
						Resource:     job2,
						HookType:     HookPostInstall,
						DeletePolicy: DeletePolicyHookSucceeded,
					}},
				},
				MainResources: []MainResource{{Resource: makeResource("Deployment", "app", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])

			By("all hooks use Create and declare cleanup timing explicitly")
			Expect(strings.Count(wf, "operation: Create")).To(Equal(2))
			Expect(wf).To(ContainSubstring("hookCleanup:"))
			Expect(wf).To(ContainSubstring("beforeCreate: true"))
			Expect(wf).To(ContainSubstring("onSuccess: true"))

			By("main resource uses Apply (SSA idempotent)")
			Expect(wf).To(ContainSubstring("operation: Apply"))
		})
	})

	Context("upgrade execution sample", func() {
		It("should generate upgrade execution YAML", func() {
			analysis := ChartAnalysis{
				Hooks:         make(map[string][]HookResource),
				MainResources: []MainResource{{Resource: makeResource("ConfigMap", "cm-1", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.ExecutionYAMLs).To(HaveKey("drplanexecution-upgrade.yaml"))

			upgradeStr := string(result.ExecutionYAMLs["drplanexecution-upgrade.yaml"])
			Expect(upgradeStr).To(ContainSubstring("operationType: Execute"))
			Expect(upgradeStr).To(ContainSubstring("mode: Upgrade"))
		})
	})

	Context("delete and rollback hooks", func() {
		It("should generate delete hook chain and delete execution sample", func() {
			preDelete := makeResource("Job", "cleanup-pre", nil)
			preDelete.SetNamespace("default")
			postDelete := makeResource("Job", "cleanup-post", nil)
			postDelete.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreDelete:  {{Resource: preDelete, HookType: HookPreDelete, Weight: 0}},
					HookPostDelete: {{Resource: postDelete, HookType: HookPostDelete, Weight: 0}},
				},
				MainResources: []MainResource{{Resource: makeResource("Deployment", "app", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(wf).To(ContainSubstring("name: cleanup-pre-pre-delete"))
			Expect(wf).To(ContainSubstring(`when: mode == "delete"`))
			Expect(wf).To(ContainSubstring("name: delete-subscription"))
			Expect(wf).To(ContainSubstring("operation: Delete"))
			Expect(wf).To(ContainSubstring("name: cleanup-post-post-delete"))
			Expect(wf).To(ContainSubstring(`name: create-subscription`))
			Expect(wf).To(ContainSubstring(`when: mode == "install" || mode == "upgrade"`))

			deleteExec := string(result.ExecutionYAMLs["drplanexecution-delete.yaml"])
			Expect(deleteExec).To(ContainSubstring("operationType: Execute"))
			Expect(deleteExec).To(ContainSubstring("mode: Delete"))
		})

		It("should generate rollback hooks and rollback mode in revert sample", func() {
			preRollback := makeResource("Job", "backup", nil)
			preRollback.SetNamespace("default")
			postRollback := makeResource("Job", "verify", nil)
			postRollback.SetNamespace("default")

			analysis := ChartAnalysis{
				Hooks: map[string][]HookResource{
					HookPreRollback:  {{Resource: preRollback, HookType: HookPreRollback, Weight: 0}},
					HookPostRollback: {{Resource: postRollback, HookType: HookPostRollback, Weight: 0}},
				},
				MainResources: []MainResource{{Resource: makeResource("Deployment", "app", nil)}},
			}

			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(wf).To(ContainSubstring("name: backup-pre-rollback"))
			Expect(wf).To(ContainSubstring("hookType: pre-rollback"))
			Expect(wf).To(ContainSubstring("name: verify-post-rollback"))
			Expect(wf).To(ContainSubstring("hookType: post-rollback"))
			Expect(wf).To(ContainSubstring(`when: mode == "rollback"`))

			revertStr := string(result.ExecutionYAMLs["drplanexecution-revert.yaml"])
			Expect(revertStr).To(ContainSubstring("operationType: Revert"))
			Expect(revertStr).To(ContainSubstring("mode: Rollback"))
		})
	})
})

func extractFirstFeedForAction(workflowYAML []byte, actionName string) map[string]interface{} {
	var wf map[string]interface{}
	Expect(sigyaml.Unmarshal(workflowYAML, &wf)).To(Succeed())

	spec, ok := wf["spec"].(map[string]interface{})
	Expect(ok).To(BeTrue())
	actions, ok := spec["actions"].([]interface{})
	Expect(ok).To(BeTrue())

	for _, item := range actions {
		action, ok := item.(map[string]interface{})
		Expect(ok).To(BeTrue())
		name, _ := action["name"].(string)
		if name != actionName {
			continue
		}
		subscription, ok := action["subscription"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		subSpec, ok := subscription["spec"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		feeds, ok := subSpec["feeds"].([]interface{})
		Expect(ok).To(BeTrue())
		Expect(feeds).NotTo(BeEmpty())
		feed, ok := feeds[0].(map[string]interface{})
		Expect(ok).To(BeTrue())
		return feed
	}

	Fail("action not found: " + actionName)
	return nil
}
