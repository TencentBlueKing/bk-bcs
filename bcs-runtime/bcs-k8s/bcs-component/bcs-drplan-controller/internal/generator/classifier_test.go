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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func makeResource(kind, name string, annotations map[string]string) unstructured.Unstructured {
	obj := unstructured.Unstructured{}
	obj.SetKind(kind)
	obj.SetName(name)
	obj.SetAPIVersion("v1")
	if annotations != nil {
		obj.SetAnnotations(annotations)
	}
	return obj
}

var _ = Describe("Classify", func() {
	It("should classify resources without hook annotations as main resources", func() {
		resources := []unstructured.Unstructured{
			makeResource("ConfigMap", "cm-1", nil),
			makeResource("Deployment", "deploy-1", nil),
			makeResource("Service", "svc-1", nil),
		}

		analysis := Classify(resources)
		Expect(analysis.MainResources).To(HaveLen(3))
		Expect(analysis.Hooks).To(BeEmpty())
		Expect(analysis.SkippedResources).To(BeEmpty())
	})

	It("should classify pre-install hook resources", func() {
		resources := []unstructured.Unstructured{
			makeResource("Job", "db-migrate", map[string]string{
				HookAnnotation: HookPreInstall,
			}),
			makeResource("ConfigMap", "cm-1", nil),
		}

		analysis := Classify(resources)
		Expect(analysis.MainResources).To(HaveLen(1))
		Expect(analysis.Hooks).To(HaveKey(HookPreInstall))
		Expect(analysis.Hooks[HookPreInstall]).To(HaveLen(1))
		Expect(analysis.Hooks[HookPreInstall][0].Resource.GetName()).To(Equal("db-migrate"))
	})

	It("should sort hooks by weight ascending", func() {
		resources := []unstructured.Unstructured{
			makeResource("Job", "job-weight-0", map[string]string{
				HookAnnotation:       HookPreInstall,
				HookWeightAnnotation: "0",
			}),
			makeResource("Job", "job-weight-neg5", map[string]string{
				HookAnnotation:       HookPreInstall,
				HookWeightAnnotation: "-5",
			}),
			makeResource("Job", "job-weight-10", map[string]string{
				HookAnnotation:       HookPreInstall,
				HookWeightAnnotation: "10",
			}),
		}

		analysis := Classify(resources)
		hooks := analysis.Hooks[HookPreInstall]
		Expect(hooks).To(HaveLen(3))
		Expect(hooks[0].Resource.GetName()).To(Equal("job-weight-neg5"))
		Expect(hooks[1].Resource.GetName()).To(Equal("job-weight-0"))
		Expect(hooks[2].Resource.GetName()).To(Equal("job-weight-10"))
	})

	It("should parse hook-delete-policy", func() {
		resources := []unstructured.Unstructured{
			makeResource("Job", "cleanup-job", map[string]string{
				HookAnnotation:   HookPostInstall,
				HookDeletePolicy: DeletePolicyHookSucceeded,
			}),
		}

		analysis := Classify(resources)
		hooks := analysis.Hooks[HookPostInstall]
		Expect(hooks).To(HaveLen(1))
		Expect(hooks[0].DeletePolicy).To(Equal(DeletePolicyHookSucceeded))
	})

	It("should skip test hooks", func() {
		resources := []unstructured.Unstructured{
			makeResource("Pod", "test-pod", map[string]string{
				HookAnnotation: HookTest,
			}),
			makeResource("Pod", "test-success-pod", map[string]string{
				HookAnnotation: HookTestSuccess,
			}),
			makeResource("ConfigMap", "cm-1", nil),
		}

		analysis := Classify(resources)
		Expect(analysis.MainResources).To(HaveLen(1))
		Expect(analysis.Hooks).To(BeEmpty())
		Expect(analysis.SkippedResources).To(HaveLen(2))
	})

	It("should handle mixed hook types", func() {
		resources := []unstructured.Unstructured{
			makeResource("Job", "pre-job", map[string]string{
				HookAnnotation: HookPreInstall,
			}),
			makeResource("ConfigMap", "cm-1", nil),
			makeResource("Deployment", "deploy-1", nil),
			makeResource("Job", "post-job", map[string]string{
				HookAnnotation: HookPostInstall,
			}),
		}

		analysis := Classify(resources)
		Expect(analysis.MainResources).To(HaveLen(2))
		Expect(analysis.Hooks).To(HaveKey(HookPreInstall))
		Expect(analysis.Hooks).To(HaveKey(HookPostInstall))
		Expect(analysis.Hooks[HookPreInstall]).To(HaveLen(1))
		Expect(analysis.Hooks[HookPostInstall]).To(HaveLen(1))
	})

	It("should split multi-hook annotation into multiple hook entries", func() {
		resources := []unstructured.Unstructured{
			makeResource("Job", "multi-hook-job", map[string]string{
				HookAnnotation: "pre-install, pre-upgrade",
			}),
		}

		analysis := Classify(resources)
		Expect(analysis.Hooks).To(HaveKey(HookPreInstall))
		Expect(analysis.Hooks).To(HaveKey(HookPreUpgrade))
		Expect(analysis.Hooks[HookPreInstall]).To(HaveLen(1))
		Expect(analysis.Hooks[HookPreUpgrade]).To(HaveLen(1))
		Expect(analysis.Hooks[HookPreInstall][0].Resource.GetName()).To(Equal("multi-hook-job"))
		Expect(analysis.Hooks[HookPreUpgrade][0].Resource.GetName()).To(Equal("multi-hook-job"))
	})

	It("should default weight to 0 when not specified", func() {
		resources := []unstructured.Unstructured{
			makeResource("Job", "no-weight", map[string]string{
				HookAnnotation: HookPreInstall,
			}),
		}

		analysis := Classify(resources)
		Expect(analysis.Hooks[HookPreInstall][0].Weight).To(Equal(0))
	})
})
