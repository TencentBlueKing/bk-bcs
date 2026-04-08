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
	var data *TemplateData

	BeforeEach(func() {
		data = &TemplateData{
			Params: map[string]interface{}{
				"namespace":   "production",
				"clusterName": "cluster-a",
				"replicas":    3,
			},
			PlanName: "my-dr-plan",
			Outputs: map[string]interface{}{
				"step1": map[string]interface{}{
					"phase":   "Succeeded",
					"message": "done",
				},
			},
		}
	})

	Context("basic parameter substitution", func() {
		It("should replace $(params.xxx) with the value", func() {
			result, err := RenderTemplate("ns: $(params.namespace)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("ns: production"))
		})

		It("should replace legacy {{ .params.xxx }} syntax", func() {
			result, err := RenderTemplate("ns: {{ .params.namespace }}", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("ns: production"))
		})

		It("should replace multiple params in one string", func() {
			result, err := RenderTemplate("$(params.clusterName)/$(params.namespace)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("cluster-a/production"))
		})

		It("should support mixing recommended and legacy syntax", func() {
			result, err := RenderTemplate("$(params.clusterName)/{{ .params.namespace }}", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("cluster-a/production"))
		})

		It("should handle integer parameter values", func() {
			result, err := RenderTemplate("replicas: $(params.replicas)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("replicas: 3"))
		})
	})

	Context("planName substitution", func() {
		It("should replace $(planName)", func() {
			result, err := RenderTemplate("plan: $(planName)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("plan: my-dr-plan"))
		})

		It("should replace legacy {{ .planName }}", func() {
			result, err := RenderTemplate("plan: {{ .planName }}", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("plan: my-dr-plan"))
		})
	})

	Context("outputs substitution", func() {
		It("should replace $(outputs.step1.phase) with nested value", func() {
			result, err := RenderTemplate("status: $(outputs.step1.phase)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("status: Succeeded"))
		})

		It("should replace $(outputs.step1.message)", func() {
			result, err := RenderTemplate("msg: $(outputs.step1.message)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("msg: done"))
		})

		It("should replace legacy {{ .outputs.step1.phase }}", func() {
			result, err := RenderTemplate("status: {{ .outputs.step1.phase }}", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("status: Succeeded"))
		})
	})

	Context("pass-through (no variables)", func() {
		It("should return the string unchanged when no variables exist", func() {
			result, err := RenderTemplate("plain text without variables", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("plain text without variables"))
		})

		It("should not modify Helm template syntax {{ }}", func() {
			result, err := RenderTemplate("{{ .Release.Name }}-config", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("{{ .Release.Name }}-config"))
		})
	})

	Context("error handling", func() {
		It("should return error for undefined parameter", func() {
			_, err := RenderTemplate("$(params.unknown)", data)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown"))
		})

		It("should return error for undefined legacy parameter", func() {
			_, err := RenderTemplate("{{ .params.unknown }}", data)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown"))
		})

		It("should return error for undefined top-level key", func() {
			_, err := RenderTemplate("$(nonexistent)", data)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("nonexistent"))
		})

		It("should return error for nested path on non-map value", func() {
			_, err := RenderTemplate("$(params.namespace.sub)", data)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not a map"))
		})
	})

	Context("nil data", func() {
		It("should handle nil data gracefully and return error for any variable", func() {
			_, err := RenderTemplate("$(params.x)", nil)
			Expect(err).To(HaveOccurred())
		})

		It("should return plain string unchanged with nil data", func() {
			result, err := RenderTemplate("no-vars-here", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("no-vars-here"))
		})
	})

	Context("edge cases", func() {
		It("should handle empty string", func() {
			result, err := RenderTemplate("", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(""))
		})

		It("should handle adjacent variables without separator", func() {
			result, err := RenderTemplate("$(params.namespace)$(params.clusterName)", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("productioncluster-a"))
		})

		It("should preserve surrounding text", func() {
			result, err := RenderTemplate("http://$(params.clusterName).example.com/api/$(params.namespace)/health", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("http://cluster-a.example.com/api/production/health"))
		})

		It("should not match incomplete syntax like $(params.xxx without closing paren", func() {
			result, err := RenderTemplate("$(params.namespace is broken", data)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal("$(params.namespace is broken"))
		})
	})
})

var _ = Describe("RenderTemplateMap", func() {
	It("should render all values in the map", func() {
		data := &TemplateData{
			Params: map[string]interface{}{
				"host": "db.example.com",
				"port": "5432",
			},
		}
		m := map[string]string{
			"Content-Type":  "application/json",
			"X-Target-Host": "$(params.host):$(params.port)",
		}
		result, err := RenderTemplateMap(m, data)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveKeyWithValue("Content-Type", "application/json"))
		Expect(result).To(HaveKeyWithValue("X-Target-Host", "db.example.com:5432"))
	})

	It("should return error if any value fails to render", func() {
		data := &TemplateData{
			Params: map[string]interface{}{},
		}
		m := map[string]string{
			"good": "static",
			"bad":  "$(params.missing)",
		}
		_, err := RenderTemplateMap(m, data)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("BuildParamsMap", func() {
	It("should merge with correct priority: stageParams > globalParams > workflowDefaults", func() {
		defaults := map[string]interface{}{
			"namespace": "default",
			"replicas":  1,
			"onlyInDef": "from-default",
		}
		global := map[string]interface{}{
			"namespace":  "production",
			"onlyInGlob": "from-global",
		}
		stage := map[string]interface{}{
			"namespace":   "staging",
			"onlyInStage": "from-stage",
		}

		result := BuildParamsMap(defaults, global, stage)

		By("stage param should win for namespace")
		Expect(result["namespace"]).To(Equal("staging"))

		By("global param should provide onlyInGlob")
		Expect(result["onlyInGlob"]).To(Equal("from-global"))

		By("default param should provide onlyInDef")
		Expect(result["onlyInDef"]).To(Equal("from-default"))

		By("default replicas should survive since not overridden")
		Expect(result["replicas"]).To(Equal(1))

		By("stage param should provide onlyInStage")
		Expect(result["onlyInStage"]).To(Equal("from-stage"))
	})

	It("should handle nil maps gracefully", func() {
		result := BuildParamsMap(nil, nil, nil)
		Expect(result).NotTo(BeNil())
		Expect(result).To(BeEmpty())
	})

	It("should handle single source only", func() {
		global := map[string]interface{}{
			"cluster": "cluster-a",
		}
		result := BuildParamsMap(nil, global, nil)
		Expect(result["cluster"]).To(Equal("cluster-a"))
	})
})
