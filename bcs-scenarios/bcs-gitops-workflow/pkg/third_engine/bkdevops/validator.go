/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bkdevops

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	gitopsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/apis/gitopsworkflow/v1"
)

type workflowValidator struct {
	workflow *gitopsv1.Workflow
}

// Validate the workflow whether legal
func (c *workflowValidator) Validate() ([]string, []string) {
	return nil, nil
}

// validateStepTemplates validate the step templates
func (c *workflowValidator) validateWorkflow() (map[string]*gitopsv1.StepTemplate, error) {
	errs := make([]string, 0)
	stpTemplates := make(map[string]*gitopsv1.StepTemplate)
	for i := range c.workflow.Spec.StepTemplates {
		stp := &c.workflow.Spec.StepTemplates[i]
		stpTemplates[stp.Name] = stp

		pluginSlice := strings.Split(stp.Plugin, "@")
		if len(pluginSlice) != 3 {
			errs = append(errs, fmt.Sprintf("stepTemplates[%d]'s plugin split by '@' length not 3", i))
		}
	}
	for i := range c.workflow.Spec.Stages {
		stage := &c.workflow.Spec.Stages[i]
		for j := range stage.Jobs {
			job := &stage.Jobs[j]
			runType := job.RunsOn.RunsType
			if runType != gitopsv1.NormalRunType && runType != gitopsv1.LinuxRunType {
				errs = append(errs, fmt.Sprintf("stage[%d].job[%d].runsOn.runsType illegal", i, j))
			}
			for k := range job.Steps {
				tpl := job.Steps[k].Template
				tplObj, ok := stpTemplates[tpl]
				if !ok {
					errs = append(errs, fmt.Sprintf("stage[%d].job[%d].steps[%d]'s template '%s' not exist",
						i, j, k, tpl))
					continue
				}
				switch runType {
				case gitopsv1.NormalRunType:
					if strings.Contains(tplObj.Plugin, string(marketBuild)+"@") {
						errs = append(errs, fmt.Sprintf("stage[%d].job[%d] runsType is '%s', "+
							"cannot use '%s' plugin", i, j, runType, string(marketBuild)))
					}
				case gitopsv1.LinuxRunType:
					if strings.Contains(tplObj.Plugin, string(marketBuildLess)+"@") {
						errs = append(errs, fmt.Sprintf("stage[%d].job[%d] runsType is '%s', "+
							"cannot use '%s' plugin", i, j, runType, string(marketBuildLess)))
					}
				}
			}
		}
	}
	if len(errs) == 0 {
		return stpTemplates, nil
	}
	return nil, errors.Errorf("check steps template failed: %s", strings.Join(errs, ", "))
}
