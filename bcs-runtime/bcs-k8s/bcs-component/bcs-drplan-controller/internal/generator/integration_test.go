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
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

func stripHookAnnotations(resources []unstructured.Unstructured) []unstructured.Unstructured {
	result := make([]unstructured.Unstructured, 0, len(resources))
	for i := range resources {
		copy := resources[i].DeepCopy()
		annotations := copy.GetAnnotations()
		if len(annotations) > 0 {
			delete(annotations, HookAnnotation)
			delete(annotations, HookWeightAnnotation)
			delete(annotations, HookDeletePolicy)
			if len(annotations) == 0 {
				copy.SetAnnotations(nil)
			} else {
				copy.SetAnnotations(annotations)
			}
		}
		result = append(result, *copy)
	}
	return result
}

var _ = Describe("Integration: demo-app chart", func() {
	var (
		resources []unstructured.Unstructured
		config    GenerateConfig
	)

	BeforeEach(func() {
		fixturePath := filepath.Join(projectRoot(), "testdata", "rendered", "demo-app.yaml")
		f, err := os.Open(filepath.Clean(fixturePath))
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = f.Close() }()

		resources, err = ParseYAML(f)
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).NotTo(BeEmpty())
		config = GenerateConfig{
			ReleaseName: "demo-app",
			Namespace:   "default",
			OutputDir:   "",
		}
	})

	Context("with hook annotations", func() {
		var analysis ChartAnalysis

		BeforeEach(func() {
			analysis = Classify(resources)
		})

		It("should classify 3 main resources and all non-rollback hook types", func() {
			Expect(analysis.MainResources).To(HaveLen(3))
			Expect(analysis.Hooks).To(HaveKey(HookPreInstall))
			Expect(analysis.Hooks).To(HaveKey(HookPostInstall))
			Expect(analysis.Hooks).To(HaveKey(HookPreUpgrade))
			Expect(analysis.Hooks).To(HaveKey(HookPostUpgrade))
			Expect(analysis.Hooks).To(HaveKey(HookPreDelete))
			Expect(analysis.Hooks).To(HaveKey(HookPostDelete))
		})

		It("should sort pre-install hooks by weight (db-migrate=-5 before validate=0)", func() {
			preHooks := analysis.Hooks[HookPreInstall]
			Expect(preHooks).To(HaveLen(2))
			Expect(preHooks[0].Resource.GetName()).To(Equal("release-name-db-migrate"))
			Expect(preHooks[0].Weight).To(Equal(-5))
			Expect(preHooks[1].Resource.GetName()).To(Equal("release-name-validate"))
			Expect(preHooks[1].Weight).To(Equal(0))
		})

		It("should generate DRPlan with unified install stage", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			planStr := string(result.PlanYAML)
			Expect(planStr).To(ContainSubstring("name: install"))
			Expect(planStr).NotTo(ContainSubstring("name: pre-install"))
			Expect(planStr).NotTo(ContainSubstring("name: post-install"))
		})

		It("should not include stage dependsOn in unified mode", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			planStr := string(result.PlanYAML)
			Expect(planStr).NotTo(ContainSubstring("dependsOn"))
		})

		It("should generate 1 workflow file + 4 execution samples", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.WorkflowYAMLs).To(HaveLen(1))
			Expect(result.WorkflowYAMLs).To(HaveKey("workflow-install.yaml"))
			Expect(result.ExecutionYAMLs).To(HaveLen(4))
			Expect(result.ExecutionYAMLs).To(HaveKey("drplanexecution-delete.yaml"))
			Expect(result.ExecutionYAMLs).To(HaveKey("drplanexecution-upgrade.yaml"))
		})

		It("should keep hooks in unified workflow with waitReady, clusterExecutionMode, and when", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			installWf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(installWf).To(ContainSubstring("waitReady: true"))
			Expect(installWf).To(ContainSubstring("clusterExecutionMode: PerCluster"))
			Expect(installWf).To(ContainSubstring(`when: mode == "install"`))
			Expect(installWf).To(ContainSubstring("hookCleanup:"))
		})

		It("should use Subscription type for hook resources (not Job type)", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			wf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(wf).To(ContainSubstring("type: Subscription"))
			Expect(wf).To(ContainSubstring("kind: Job"))
			Expect(wf).To(ContainSubstring("release-name-db-migrate"))
			Expect(wf).To(ContainSubstring("release-name-validate"))
			Expect(wf).To(ContainSubstring("release-name-install-smoke"))
			Expect(wf).To(ContainSubstring("release-name-upgrade-verify"))
			Expect(wf).NotTo(ContainSubstring("type: Job"))
		})

		It("should use template variable for namespace in all workflows", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			for name, wfBytes := range result.WorkflowYAMLs {
				wf := string(wfBytes)
				Expect(wf).To(ContainSubstring("$(params.feedNamespace)"),
					"workflow %s should use feedNamespace template variable", name)
				Expect(wf).To(ContainSubstring("feedNamespace"),
					"workflow %s should declare feedNamespace parameter", name)
			}
		})

		It("should include all 3 main resources as Subscription feeds", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			installWf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(installWf).To(ContainSubstring("release-name-config"))
			Expect(installWf).To(ContainSubstring("release-name-server"))
			Expect(installWf).To(ContainSubstring("release-name-svc"))
		})

		It("should write all files to output directory", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			tmpDir, err := os.MkdirTemp("", "integration-test-*")
			Expect(err).NotTo(HaveOccurred())
			defer func() { _ = os.RemoveAll(tmpDir) }()

			err = WriteOutput(result, tmpDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Join(tmpDir, "drplan.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "workflow-install.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "drplanexecution-install.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "drplanexecution-delete.yaml")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "drplanexecution-revert.yaml")).To(BeAnExistingFile())
		})

		Context("golden file comparison", func() {
			goldenFiles := []string{
				"drplan.yaml",
				"workflow-install.yaml",
				"drplanexecution-install.yaml",
				"drplanexecution-revert.yaml",
			}

			It("should match all golden output files exactly", func() {
				result, err := GeneratePlan(analysis, config)
				Expect(err).NotTo(HaveOccurred())

				goldenDir := filepath.Join(projectRoot(), "testdata", "output")

				allGenerated := map[string][]byte{
					"drplan.yaml": result.PlanYAML,
				}
				for k, v := range result.WorkflowYAMLs {
					allGenerated[k] = v
				}
				for k, v := range result.ExecutionYAMLs {
					allGenerated[k] = v
				}

				for _, fname := range goldenFiles {
					By("comparing " + fname)
					expectedBytes, readErr := os.ReadFile(filepath.Clean(filepath.Join(goldenDir, fname)))
					Expect(readErr).NotTo(HaveOccurred(), "golden file %s should exist", fname)

					actualBytes, exists := allGenerated[fname]
					Expect(exists).To(BeTrue(), "generated result should contain %s", fname)

					Expect(string(actualBytes)).To(Equal(string(expectedBytes)),
						"generated %s should match golden file", fname)
				}
			})
		})
	})

	Context("without hook annotations", func() {
		var analysis ChartAnalysis

		BeforeEach(func() {
			noHookResources := stripHookAnnotations(resources)
			analysis = Classify(noHookResources)
		})

		It("should classify all resources as main resources", func() {
			Expect(analysis.MainResources).To(HaveLen(10))
			Expect(analysis.Hooks).To(BeEmpty())
		})

		It("should generate DRPlan with only install stage", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			planStr := string(result.PlanYAML)
			Expect(planStr).To(ContainSubstring("name: install"))
			Expect(planStr).NotTo(ContainSubstring("dependsOn"))
		})

		It("should generate only install workflow", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.WorkflowYAMLs).To(HaveLen(1))
			Expect(result.WorkflowYAMLs).To(HaveKey("workflow-install.yaml"))
		})

		It("should include all 10 resources in install subscription feeds", func() {
			result, err := GeneratePlan(analysis, config)
			Expect(err).NotTo(HaveOccurred())

			installWf := string(result.WorkflowYAMLs["workflow-install.yaml"])
			Expect(installWf).To(ContainSubstring("release-name-config"))
			Expect(installWf).To(ContainSubstring("release-name-server"))
			Expect(installWf).To(ContainSubstring("release-name-svc"))
			Expect(installWf).To(ContainSubstring("release-name-db-migrate"))
			Expect(installWf).To(ContainSubstring("release-name-install-smoke"))
			Expect(installWf).To(ContainSubstring("release-name-upgrade-verify"))
			Expect(installWf).To(ContainSubstring("release-name-upgrade-prepare"))
			Expect(installWf).To(ContainSubstring("release-name-pre-delete-backup"))
			Expect(installWf).To(ContainSubstring("release-name-post-delete-notify"))
			Expect(installWf).To(ContainSubstring("release-name-validate"))
		})
	})

})
