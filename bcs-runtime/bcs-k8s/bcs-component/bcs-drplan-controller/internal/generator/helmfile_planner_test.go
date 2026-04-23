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
})

func boolPtr(v bool) *bool {
	return &v
}
