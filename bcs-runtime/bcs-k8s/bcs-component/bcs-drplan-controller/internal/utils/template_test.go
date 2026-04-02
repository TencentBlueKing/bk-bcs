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

package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RenderTemplate", func() {
	Context("When substituting $(params.xxx) variables", func() {
		It("should replace a single parameter", func() {
			data := &TemplateData{
				Params: map[string]interface{}{
					"endpoint": "http://api.example.com",
				},
			}
			result, err := RenderTemplate("$(params.endpoint)/health", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("http://api.example.com/health"))
		})

		It("should replace multiple parameters in the same string", func() {
			data := &TemplateData{
				Params: map[string]interface{}{
					"host": "db.example.com",
					"port": "5432",
				},
			}
			result, err := RenderTemplate("postgres://$(params.host):$(params.port)/mydb", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("postgres://db.example.com:5432/mydb"))
		})

		It("should handle numeric parameter values", func() {
			data := &TemplateData{
				Params: map[string]interface{}{
					"replicas": 3,
				},
			}
			result, err := RenderTemplate("replicas: $(params.replicas)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("replicas: 3"))
		})

		It("should preserve unresolved parameters when not found", func() {
			data := &TemplateData{
				Params: map[string]interface{}{},
			}
			result, err := RenderTemplate("$(params.missing)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("$(params.missing)"))
		})
	})

	Context("When substituting $(planName)", func() {
		It("should replace planName", func() {
			data := &TemplateData{
				PlanName: "dr-plan-production",
				Params:   map[string]interface{}{},
			}
			result, err := RenderTemplate("dr-app-$(planName)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("dr-app-dr-plan-production"))
		})

		It("should handle empty planName", func() {
			data := &TemplateData{
				PlanName: "",
				Params:   map[string]interface{}{},
			}
			result, err := RenderTemplate("prefix-$(planName)-suffix", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("prefix--suffix"))
		})
	})

	Context("When substituting $(outputs.xxx) variables", func() {
		It("should replace output references", func() {
			data := &TemplateData{
				Params: map[string]interface{}{},
				Outputs: map[string]interface{}{
					"checkResult": "healthy",
				},
			}
			result, err := RenderTemplate("status: $(outputs.checkResult)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("status: healthy"))
		})

		It("should preserve unresolved output references", func() {
			data := &TemplateData{
				Params:  map[string]interface{}{},
				Outputs: map[string]interface{}{},
			}
			result, err := RenderTemplate("$(outputs.missing)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("$(outputs.missing)"))
		})
	})

	Context("When mixing multiple variable types", func() {
		It("should resolve params, planName, and outputs together", func() {
			data := &TemplateData{
				Params: map[string]interface{}{
					"cluster": "zone-b",
				},
				PlanName: "failover-01",
				Outputs: map[string]interface{}{
					"endpoint": "10.0.0.1",
				},
			}
			result, err := RenderTemplate(
				"plan=$(planName) cluster=$(params.cluster) endpoint=$(outputs.endpoint)",
				data,
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("plan=failover-01 cluster=zone-b endpoint=10.0.0.1"))
		})
	})

	Context("When handling nil data", func() {
		It("should handle nil TemplateData gracefully", func() {
			result, err := RenderTemplate("$(params.xxx)", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("$(params.xxx)"))
		})
	})

	Context("When string contains no variables", func() {
		It("should return the original string unchanged", func() {
			result, err := RenderTemplate("no-variables-here", &TemplateData{
				Params: map[string]interface{}{"key": "value"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("no-variables-here"))
		})

		It("should not touch Helm template syntax", func() {
			result, err := RenderTemplate("{{ .Values.replicaCount }}", &TemplateData{
				Params: map[string]interface{}{},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("{{ .Values.replicaCount }}"))
		})

		It("should not touch Shell variable syntax", func() {
			result, err := RenderTemplate("${HOME} $(date +%Y)", &TemplateData{
				Params: map[string]interface{}{},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("${HOME} $(date +%Y)"))
		})
	})

	Context("When rendering multi-line YAML templates", func() {
		It("should substitute variables in a YAML manifest", func() {
			data := &TemplateData{
				Params: map[string]interface{}{
					"targetCluster": "zone-b-ns",
					"drMode":        "active",
				},
			}
			tmpl := `apiVersion: v1
kind: ConfigMap
metadata:
  name: "dr-config"
  namespace: "$(params.targetCluster)"
data:
  dr-mode: "$(params.drMode)"`

			expected := `apiVersion: v1
kind: ConfigMap
metadata:
  name: "dr-config"
  namespace: "zone-b-ns"
data:
  dr-mode: "active"`

			result, err := RenderTemplate(tmpl, data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expected))
		})
	})
})

var _ = Describe("RenderTemplateMap", func() {
	It("should render all values in the map", func() {
		m := map[string]string{
			"url":    "$(params.endpoint)/api",
			"header": "Bearer $(params.token)",
		}
		data := &TemplateData{
			Params: map[string]interface{}{
				"endpoint": "https://api.example.com",
				"token":    "secret123",
			},
		}
		result, err := RenderTemplateMap(m, data)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveKeyWithValue("url", "https://api.example.com/api"))
		Expect(result).To(HaveKeyWithValue("header", "Bearer secret123"))
	})
})

var _ = Describe("BuildParamsMap", func() {
	It("should merge with correct priority: stageParams > globalParams > workflowDefaults", func() {
		workflowDefaults := map[string]interface{}{
			"a": "default-a",
			"b": "default-b",
			"c": "default-c",
		}
		globalParams := map[string]interface{}{
			"b": "global-b",
			"c": "global-c",
		}
		stageParams := map[string]interface{}{
			"c": "stage-c",
		}
		result := BuildParamsMap(workflowDefaults, globalParams, stageParams)
		Expect(result["a"]).To(Equal("default-a"))
		Expect(result["b"]).To(Equal("global-b"))
		Expect(result["c"]).To(Equal("stage-c"))
	})
})
