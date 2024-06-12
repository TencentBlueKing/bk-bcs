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
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/internal/logctx"
	gitopsv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-workflow/pkg/apis/gitopsworkflow/v1"
)

type workflowTransfer struct {
	ctx           context.Context
	workflow      *gitopsv1.Workflow
	stepTemplates map[string]*gitopsv1.StepTemplate
}

// transToPipeline will transfer workflow object to bkdevops pipeline
func (t *workflowTransfer) transToPipeline() (*pipeline, error) {
	pp := &pipeline{
		Name: t.workflow.Spec.Name,
		Desc: t.workflow.Spec.Desc,
	}

	// check step templates
	checker := &workflowValidator{workflow: t.workflow}
	var err error
	t.stepTemplates, err = checker.validateStepTemplates()
	if err != nil {
		return nil, err
	}

	// do transfer
	pp.Stages = append(pp.Stages, t.buildStartStage())
	for i := range t.workflow.Spec.Stages {
		pp.Stages = append(pp.Stages, t.transferStage(&t.workflow.Spec.Stages[i]))
	}
	return pp, nil
}

// transferStage transfer stage object
func (t *workflowTransfer) transferStage(defineStage *gitopsv1.Stage) *pipelineStage {
	ppStage := &pipelineStage{
		Name:        defineStage.Name,
		StageEnable: !defineStage.Disabled,
	}
	ppStage.StageControlOption = t.buildControlOption(stageDfaultConditionType,
		defineStage.Condition, defineStage.Timeout)
	if len(defineStage.ReviewUsers) != 0 {
		ppStage.CheckIn = &checkIn{
			ManualTrigger: true,
			NotifyType:    []string{"RTX", "WEWORK_GROUP"},
			NotifyGroup:   defineStage.ReviewNotifyGroup,
		}
		message := defineStage.ReviewMessage
		if message == "" {
			message = "是否同意执行？"
		}
		ppStage.CheckIn.ReviewGroups = append(ppStage.CheckIn.ReviewGroups, reviewGroup{
			Name:      message,
			Reviewers: defineStage.ReviewUsers,
		})
	}
	for i := range defineStage.Jobs {
		ppStage.Containers = append(ppStage.Containers, t.transferJob(&defineStage.Jobs[i]))
	}
	return ppStage
}

// transferJob transfer job object
func (t *workflowTransfer) transferJob(defineJob *gitopsv1.Job) *container {
	c := &container{
		Type:            vmBuildJobType,
		Name:            defineJob.Name,
		ContainerEnable: !defineJob.Disable,
	}
	c.JobControlOption = t.buildControlOption(jobDefaultConditionType, defineJob.Condition, defineJob.Timeout)

	if len(defineJob.Strategy.Matrix) != 0 {
		c.MatrixGroupFlag = true
		matrix := make([]string, 0)
		for k, v := range defineJob.Strategy.Matrix {
			matrix = append(matrix, fmt.Sprintf(`%s: [ %s ]`, k, strings.Join(v, ", ")))
		}
		c.MatrixControlOption = &matrixControlOption{
			StrategyStr:    strings.Join(matrix, "\\n"),
			MaxConcurrency: 5,
		}
	}
	c.BaseOS = "LINUX"
	c.DispatchType = &dispatchType{
		BuildType:    "PUBLIC_DEVCLOUD",
		Value:        "tlinux3_ci",
		ImageType:    "BKSTORE",
		ImageCode:    "tlinux3_ci",
		ImageVersion: "2.*",
	}
	for i := range defineJob.Steps {
		stp := t.transferStep(&defineJob.Steps[i])
		if stp == nil {
			continue
		}
		c.Elements = append(c.Elements, stp)
	}
	return c
}

// transferStep transfer step object
func (t *workflowTransfer) transferStep(step *gitopsv1.Step) interface{} {
	tpl, ok := t.stepTemplates[step.Template]
	if !ok {
		logctx.Warnf(t.ctx, "template '%s' not found", step.Template)
		return nil
	}
	newStep := step.DeepCopy()
	if len(newStep.Condition) == 0 {
		newStep.Condition = tpl.Condition
	}
	if newStep.Timeout == 0 {
		newStep.Timeout = tpl.Timeout
	}
	if newStep.With == nil {
		newStep.With = make(map[string]string)
	}
	for k, v := range tpl.With {
		_, ok = newStep.With[k]
		if !ok {
			newStep.With[k] = v
		}
	}
	pluginSlice := strings.Split(tpl.Plugin, "@")
	if len(pluginSlice) != 3 {
		logctx.Warnf(t.ctx, "template '%s' plugin '%s' split by '@' length not 3", step.Template, tpl.Plugin)
		return nil
	}
	pluginType := pluginSlice[0]
	pluginName := pluginSlice[1]
	pluginVersion := pluginSlice[2]
	switch elementType(pluginType) {
	case linuxScript:
		return t.buildElementLinuxScript(pluginVersion, newStep)
	case marketBuild, marketBuildLess:
		return t.buildElementMarketBuild(pluginType, pluginName, pluginVersion, newStep)
	default:
		logctx.Warnf(t.ctx, "unknown plugin type '%s'", pluginType)
		return t.buildElementMarketBuild(pluginType, pluginName, pluginVersion, newStep)
	}
}

func (t *workflowTransfer) buildElementLinuxScript(pluginVersion string, step *gitopsv1.Step) *elementLinuxScript {
	result := &elementLinuxScript{
		Type:          linuxScript,
		Name:          step.Name,
		ElementEnable: !step.Disable,
		Version:       pluginVersion,
		AtomCode:      string(linuxScript),
		ClassType:     string(linuxScript),
		ScriptType:    "SHELL",
		Script:        step.With["script"],
	}
	result.AdditionalOptions = t.buildControlOption(preTaskSuccess, step.Condition, step.Timeout)
	return result
}

func (t *workflowTransfer) buildElementMarketBuild(pluginType, pluginName, pluginVersion string,
	step *gitopsv1.Step) *elementMarketBuild {
	result := &elementMarketBuild{
		Type:          elementType(pluginType),
		Name:          step.Name,
		ElementEnable: !step.Disable,
		Version:       pluginVersion,
		AtomCode:      pluginName,
		ClassType:     pluginType,
		ExecuteCount:  1,
		Data: struct {
			Input map[string]string `json:"input,omitempty"`
		}{
			Input: step.With,
		},
	}
	result.AdditionalOptions = t.buildControlOption(preTaskSuccess, step.Condition, step.Timeout)
	return result
}

func (t *workflowTransfer) buildControlOption(defaultCondition conditionType, conditions map[string]string,
	timeout int64) *controlOption {
	if timeout == 0 {
		timeout = 900
	}
	result := &controlOption{
		Enable:  true,
		Timeout: timeout,
	}
	if len(conditions) == 0 {
		result.RunCondition = defaultCondition
	} else {
		result.RunCondition = customVariableMatch
		for k, v := range conditions {
			result.CustomVariables = append(result.CustomVariables, variable{Key: k, Value: v})
		}
	}
	return result
}

func (t *workflowTransfer) buildStartStage() *pipelineStage {
	params := make([]customParam, 0, len(t.workflow.Spec.Params))
	for i := range t.workflow.Spec.Params {
		p := t.workflow.Spec.Params[i]
		defaultValue := p.Value
		if defaultValue == "" {
			defaultValue = "defaultValue"
		}
		params = append(params, customParam{
			ID:           p.Name,
			Required:     true,
			Type:         "STRING",
			DefaultValue: defaultValue,
		})
	}
	return &pipelineStage{
		Name:        "stage-1",
		IsTrigger:   true,
		StageEnable: true,
		Containers: []*container{
			{
				Type: "trigger",
				Name: "构建触发",
				Elements: []interface{}{
					t.buildStartStageManualTrigger(), t.buildStartStageRemoteTrigger(),
				},
				Params: params,
			},
		},
	}
}

func (t *workflowTransfer) buildStartStageManualTrigger() *elementManualTrigger {
	return &elementManualTrigger{
		Type:           manualTrigger,
		Name:           "手动触发",
		CanElementSkip: true,
		Version:        "1.*",
		ClassType:      string(manualTrigger),
		ElementEnable:  true,
		AtomCode:       string(manualTrigger),
	}
}

func (t *workflowTransfer) buildStartStageRemoteTrigger() *elementRemoteTrigger {
	return &elementRemoteTrigger{
		Type:          remoteTrigger,
		Name:          "远程触发",
		Version:       "1.*",
		ClassType:     string(remoteTrigger),
		ElementEnable: true,
		AtomCode:      string(remoteTrigger),
	}
}
