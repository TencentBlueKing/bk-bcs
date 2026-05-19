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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generator Suite")
}

var _ = Describe("ParseYAML", func() {
	It("should parse a single YAML document", func() {
		input := `apiVersion: v1
kind: ConfigMap
metadata:
  name: test-cm
  namespace: default
data:
  key: value`

		resources, err := ParseYAML(strings.NewReader(input))
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).To(HaveLen(1))
		Expect(resources[0].GetKind()).To(Equal("ConfigMap"))
		Expect(resources[0].GetName()).To(Equal("test-cm"))
	})

	It("should parse multiple YAML documents", func() {
		input := `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy-1
---
apiVersion: v1
kind: Service
metadata:
  name: svc-1`

		resources, err := ParseYAML(strings.NewReader(input))
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).To(HaveLen(3))
		Expect(resources[0].GetKind()).To(Equal("ConfigMap"))
		Expect(resources[1].GetKind()).To(Equal("Deployment"))
		Expect(resources[2].GetKind()).To(Equal("Service"))
	})

	It("should skip empty documents between separators", func() {
		input := `---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-1
---
---
apiVersion: v1
kind: Service
metadata:
  name: svc-1
---`

		resources, err := ParseYAML(strings.NewReader(input))
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).To(HaveLen(2))
	})

	It("should skip comment-only documents", func() {
		input := `# Source: my-chart/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-1
---
# This is just a comment block
# with no actual resource
---
apiVersion: v1
kind: Service
metadata:
  name: svc-1`

		resources, err := ParseYAML(strings.NewReader(input))
		Expect(err).NotTo(HaveOccurred())
		Expect(resources).To(HaveLen(2))
	})

	It("should return error for completely empty input", func() {
		resources, err := ParseYAML(strings.NewReader(""))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no valid"))
		Expect(resources).To(BeNil())
	})

	It("should return error for invalid YAML", func() {
		input := `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-1
---
this is: [not: valid: yaml: {{`

		_, err := ParseYAML(strings.NewReader(input))
		Expect(err).To(HaveOccurred())
	})
})
