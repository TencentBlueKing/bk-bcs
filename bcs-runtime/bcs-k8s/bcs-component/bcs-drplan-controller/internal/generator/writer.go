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
	"fmt"
	"os"
	"path/filepath"
)

// WriteOutput writes the generated DRPlan, DRWorkflow, and DRPlanExecution YAML
// files to the specified output directory. Creates the directory if it does not exist.
func WriteOutput(result *GenerateResult, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("creating output directory %s: %w", outputDir, err)
	}

	if err := writeFile(filepath.Join(outputDir, "drplan.yaml"), result.PlanYAML); err != nil {
		return err
	}

	for filename, data := range result.WorkflowYAMLs {
		if err := writeFile(filepath.Join(outputDir, filename), data); err != nil {
			return err
		}
	}

	for filename, data := range result.ExecutionYAMLs {
		if err := writeFile(filepath.Join(outputDir, filename), data); err != nil {
			return err
		}
	}

	return nil
}

func writeFile(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
