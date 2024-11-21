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

// Package steps xxx
package steps

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/itsm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
)

const (
	itsmSubmitStepName   = "提交额度管理通用审批单据"
	itsmSubmitStepMethod = "itsm-submit"
)

// NewItsmSubmitStep itsm common step
func NewItsmSubmitStep() task.StepBuilder {
	return &itsmSubmitStep{}
}

// itsmSubmitStep itsm submit step
type itsmSubmitStep struct{}

// GetName stepName
func (s itsmSubmitStep) GetName() string {
	return itsmSubmitStepMethod
}

// Alias method name
func (s itsmSubmitStep) Alias() string {
	return itsmSubmitStepMethod
}

func (s itsmSubmitStep) getStepParams(task *types.Task) (*ItsmStepParams, error) {
	step, ok := task.GetStep(s.Alias())
	if !ok {
		return nil, fmt.Errorf("task[%s] step[%s] not exist", task.GetTaskID(), s.GetName())
	}
	user, ok := step.GetParam(utils.UserNameKey.String())
	if !ok {
		return nil, fmt.Errorf("task[%s] step[%s] user empty", task.GetTaskID(), s.GetName())
	}
	projectCode, ok := step.GetParam(utils.ProjectCodeKey.String())
	if !ok {
		return nil, fmt.Errorf("task[%s] step[%s] project empty", task.GetTaskID(), s.GetName())
	}
	clusterId, ok := step.GetParam(utils.ClusterIDKey.String())
	if !ok {
		return nil, fmt.Errorf("task[%s] step[%s] cluster empty", task.GetTaskID(), s.GetName())
	}
	content, ok := step.GetParam(utils.ContentKey.String())
	if !ok {
		return nil, fmt.Errorf("task[%s] step[%s] content empty", task.GetTaskID(), s.GetName())
	}

	return &ItsmStepParams{
		User:        user,
		ProjectCode: projectCode,
		ClusterId:   clusterId,
		Content:     content,
	}, nil
}

// DoWork for worker exec task
func (s itsmSubmitStep) DoWork(task *types.Task) error {
	// step params && handle logic
	params, err := s.getStepParams(task)
	if err != nil {
		return err
	}

	itsmData, err := itsm.SubmitQuotaManagerCommonTicket(params.User, params.ProjectCode, params.ClusterId, params.Content)
	if err != nil {
		logging.Error("quotaManagerItsmSubmitStep[%s] SubmitQuotaManagerCommonTicket failed, err: %s",
			task.GetTaskID(), err.Error())
		return err
	}

	logging.Info("quotaManagerItsmSubmitStep[%s] success itsm %s:%s", task.GetTaskID(), itsmData.SN, itsmData.TicketURL)

	if task.CommonParams == nil {
		task.CommonParams = make(map[string]string)
	}

	task.CommonParams[utils.ItsmSnKey.String()] = itsmData.SN

	return nil
}

// BuildStep build step
func (s itsmSubmitStep) BuildStep(kvs []task.KeyValue, opts ...types.StepOption) *types.Step {
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}

// ItsmStepParams xxx
type ItsmStepParams struct {
	User        string
	ProjectCode string
	ClusterId   string
	Content     string
}

// TransItsmStepParamsToKeyValue 转换itsm参数为key value
func TransItsmStepParamsToKeyValue(params ItsmStepParams) []task.KeyValue {
	kvs := make([]task.KeyValue, 0)

	kvs = append(kvs, task.KeyValue{
		Key:   utils.UserNameKey,
		Value: params.User,
	})
	kvs = append(kvs, task.KeyValue{
		Key:   utils.ProjectCodeKey,
		Value: params.ProjectCode,
	})
	kvs = append(kvs, task.KeyValue{
		Key:   utils.ClusterIDKey,
		Value: params.ClusterId,
	})
	kvs = append(kvs, task.KeyValue{
		Key:   utils.ContentKey,
		Value: params.Content,
	})

	return kvs
}
