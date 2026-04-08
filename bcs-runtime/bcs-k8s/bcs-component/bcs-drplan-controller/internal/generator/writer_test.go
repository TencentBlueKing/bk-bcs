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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("WriteOutput", func() {
	var tmpDir string

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "drplan-gen-test-*")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	It("should write all files to the output directory", func() {
		result := &GenerateResult{
			PlanYAML: []byte("apiVersion: dr.bkbcs.tencent.com/v1alpha1\nkind: DRPlan\n"),
			WorkflowYAMLs: map[string][]byte{
				"workflow-install.yaml": []byte("apiVersion: dr.bkbcs.tencent.com/v1alpha1\nkind: DRWorkflow\n"),
			},
			ExecutionYAMLs: map[string][]byte{
				"drplanexecution-execute.yaml": []byte("kind: DRPlanExecution\n"),
				"drplanexecution-revert.yaml":  []byte("kind: DRPlanExecution\n"),
			},
		}

		err := WriteOutput(result, tmpDir)
		Expect(err).NotTo(HaveOccurred())

		Expect(filepath.Join(tmpDir, "drplan.yaml")).To(BeAnExistingFile())
		Expect(filepath.Join(tmpDir, "workflow-install.yaml")).To(BeAnExistingFile())
		Expect(filepath.Join(tmpDir, "drplanexecution-execute.yaml")).To(BeAnExistingFile())
		Expect(filepath.Join(tmpDir, "drplanexecution-revert.yaml")).To(BeAnExistingFile())
	})

	It("should create output directory if it does not exist", func() {
		nestedDir := filepath.Join(tmpDir, "sub", "dir")
		result := &GenerateResult{
			PlanYAML:       []byte("kind: DRPlan\n"),
			WorkflowYAMLs:  map[string][]byte{},
			ExecutionYAMLs: map[string][]byte{},
		}

		err := WriteOutput(result, nestedDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(filepath.Join(nestedDir, "drplan.yaml")).To(BeAnExistingFile())
	})

	It("should write correct file contents", func() {
		content := []byte("test-plan-content")
		result := &GenerateResult{
			PlanYAML:       content,
			WorkflowYAMLs:  map[string][]byte{},
			ExecutionYAMLs: map[string][]byte{},
		}

		err := WriteOutput(result, tmpDir)
		Expect(err).NotTo(HaveOccurred())

		data, err := os.ReadFile(filepath.Clean(filepath.Join(tmpDir, "drplan.yaml")))
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(Equal(content))
	})
})
