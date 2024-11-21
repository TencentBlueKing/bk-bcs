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

// Package tasks xxx
package tasks

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/common/steps"
)

const (
	// ProviderName provider name
	ProviderName = "internal"
)

type itsmData struct {
	operator    string
	projectCode string
	clusterId   string
	content     string
}

func buildItsmQuotaSteps(itsmId string, info itsmData) []*types.Step { // nolint
	stepList := make([]*types.Step, 0)

	// 1. 审批配额申请
	itsmSubmit := steps.NewItsmSubmitStep()

	itsmSubmitStep := itsmSubmit.BuildStep(steps.TransItsmStepParamsToKeyValue(steps.ItsmStepParams{
		User:        info.operator,
		ProjectCode: info.projectCode,
		ClusterId:   info.clusterId,
		Content:     info.content,
	}))
	stepList = append(stepList, itsmSubmitStep)

	// 2. 等待审批通过，审批通过后, 更新配额状态; 审批拒绝后, 更新配额申请状态
	itsmApprove := steps.NewItsmApproveStep()
	itsmApproveStep := itsmApprove.BuildStep(nil)
	stepList = append(stepList, itsmApproveStep)

	return stepList
}
