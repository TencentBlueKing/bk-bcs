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
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Helmfile loader helpers", func() {
	It("should normalize local tgz chart name", func() {
		got := normalizeChartName("./charts/bcs-services-stack-1.2.3.tgz", "1.2.3")
		Expect(got).To(Equal("bcs-services-stack"))
	})

	It("should select exactly one release", func() {
		_, err := selectSingleRelease([]HelmfileResolvedRelease{
			{ReleaseName: "a"},
			{ReleaseName: "b"},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expected exactly 1 release"))
	})

	It("should strip repo alias when chart is repo/chart", func() {
		got := normalizeChartName("bitnami/nginx", "1.2.3")
		Expect(got).To(Equal("nginx"))
	})

	It("should preserve basename when version suffix does not match", func() {
		got := normalizeChartName("./charts/custom-bcs-stack.tgz", "1.2.3")
		Expect(got).To(Equal("custom-bcs-stack"))
	})

	It("should reject empty release selection", func() {
		_, err := selectSingleRelease(nil)
		Expect(err).To(HaveOccurred())
		Expect(strings.ToLower(err.Error())).To(ContainSubstring("expected exactly 1 release"))
	})

	It("should load one release from helmfile fixture and merge final values", func() {
		fixture := filepath.Join(projectRoot(), "testdata", "helmfile", "basic", "helmfile.yaml.gotmpl")

		release, err := LoadHelmfileRelease(HelmfileLoadInput{
			File:      fixture,
			Selectors: []string{"name=nginx-demo"},
			ChartRepo: "oci://registry.example.com/charts",
			PlainHTTP: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(release.ReleaseName).To(Equal("nginx-demo"))
		Expect(release.Namespace).To(Equal("default"))
		Expect(release.Chart).To(Equal("nginx"))
		Expect(release.ChartVersion).To(Equal("0.1.1"))
		Expect(release.TimeoutSeconds).To(Equal(int32(600)))
		Expect(release.Atomic).NotTo(BeNil())
		Expect(*release.Atomic).To(BeTrue())
		Expect(release.Wait).NotTo(BeNil())
		Expect(*release.Wait).To(BeFalse())
		Expect(release.PlainHTTP).NotTo(BeNil())
		Expect(*release.PlainHTTP).To(BeTrue())
		Expect(release.ValuesYAML).To(ContainSubstring("replicaCount: 2"))
		Expect(release.ValuesYAML).To(ContainSubstring("tag: 0.1.1"))
		Expect(release.ValuesYAML).To(ContainSubstring("commonLabels:"))
		Expect(release.ValuesYAML).NotTo(ContainSubstring("repository: nginx"))
		Expect(release.ValuesYAML).NotTo(ContainSubstring("type: ClusterIP"))
		Expect(release.ValuesYAML).NotTo(ContainSubstring("port: 80"))
	})

	It("should keep full values when keep-full-values behavior is requested", func() {
		fixture := filepath.Join(projectRoot(), "testdata", "helmfile", "basic", "helmfile.yaml.gotmpl")

		release, err := LoadHelmfileRelease(HelmfileLoadInput{
			File:           fixture,
			Selectors:      []string{"name=nginx-demo"},
			ChartRepo:      "oci://registry.example.com/charts",
			KeepFullValues: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(release.ValuesYAML).To(ContainSubstring("replicaCount: 2"))
		Expect(release.ValuesYAML).To(ContainSubstring("repository: nginx"))
		Expect(release.ValuesYAML).To(ContainSubstring("type: ClusterIP"))
		Expect(release.ValuesYAML).To(ContainSubstring("port: 80"))
	})

	It("should tolerate duplicate keys in rendered values files by keeping the last value", func() {
		fixture := filepath.Join(projectRoot(), "testdata", "helmfile", "duplicate", "helmfile.yaml.gotmpl")

		release, err := LoadHelmfileRelease(HelmfileLoadInput{
			File:      fixture,
			Selectors: []string{"name=dup-demo"},
			ChartRepo: "oci://registry.example.com/charts",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(release.ValuesYAML).To(ContainSubstring("foo: second"))
		Expect(release.ValuesYAML).NotTo(ContainSubstring("foo: first"))
	})

	It("should reject remote chart refs when default-value diff is requested", func() {
		stateFile := filepath.Join(projectRoot(), "testdata", "helmfile", "basic", "helmfile.yaml.gotmpl")

		_, err := diffValuesAgainstChartDefaults(stateFile, "oci://registry.example.com/charts/nginx", map[string]interface{}{
			"replicaCount": int64(2),
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("--keep-full-values"))
	})
})
