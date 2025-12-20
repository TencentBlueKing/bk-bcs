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
	"context"
	"fmt"
	"time"

	common_task "github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	v2 "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/itsm/v2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
)

const (
	itsmApproveStepName   = "额度管理单据审批"
	itsmApproveStepMethod = "itsm-approve"
)

// NewItsmApproveStep itsm approve step
func NewItsmApproveStep() common_task.StepBuilder {
	return &itsmApproveStep{}
}

// itsmApproveStep itsm approve step
type itsmApproveStep struct{}

// GetName stepName
func (s itsmApproveStep) GetName() string {
	return itsmApproveStepMethod
}

// Alias method name
func (s itsmApproveStep) Alias() string {
	return itsmApproveStepMethod
}

// DoWork for worker exec task
func (s itsmApproveStep) DoWork(task *types.Task) error {
	sn, ok := task.GetCommonParams(utils.ItsmSnKey.String())
	if !ok {
		return fmt.Errorf("itsmApproveStep[%s] get itsmSn failed", task.TaskID)
	}

	step, ok := task.GetStep(s.Alias())
	if !ok {
		return fmt.Errorf("task[%s] step[%s] not exist", task.GetTaskID(), s.GetName())
	}
	projectCode, ok := step.GetParam(utils.ProjectCodeKey.String())
	if !ok {
		return fmt.Errorf("task[%s] step[%s] project empty", task.GetTaskID(), s.GetName())
	}

	ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(),
		tenant.ResourceMetaData{ProjectCode: projectCode})
	if err != nil {
		return err
	}

	// 查询单据状态，当前不会超时。后续可根据默认超时时间取消该单据(30天等)
	err = common_task.LoopDoFunc(context.Background(), func() error {
		ticket, lerr := v2.ListTicketsApprovalResult(ctx, []string{sn})
		if lerr != nil {
			logging.Error("itsmApproveStep[%s] ListTicketsApprovalResult failed: %v", task.GetTaskID(), err)
			return nil
		}

		logging.Info("itsmApproveStep[%s] quotaManagerItsmApproveStep sm %s currentStatus %s, approval %v",
			task.GetTaskID(), sn, ticket[0].CurrentStatus, ticket[0].ApprovalResult)
		// RUNNING（处理中）/FINISHED（已结束）/TERMINATED（被终止）/ SUSPENDED（被挂起）
		switch ticket[0].CurrentStatus {
		case v2.RUNNING:
			return nil
		case v2.FINISHED:
			if ticket[0].ApprovalResult {
				return common_task.ErrEndLoop
			}

			return fmt.Errorf("ticket sn[%s] approval result is false", sn)
		case v2.SUSPENDED, v2.TERMINATED, v2.REVOKED:
			return fmt.Errorf("ticket sn[%s] status[%s] is not expected", sn, ticket[0].CurrentStatus)
		default:
		}

		return nil
	}, common_task.LoopInterval(20*time.Second))

	if err != nil {
		logging.Error("itsmApproveStep[%s] apprive sn[%s] failed, err: %s",
			task.GetTaskID(), sn, err.Error())
		return err
	}

	logging.Info("itsmApproveStep[%s] approve itsm %s:%s", task.GetTaskID(), sn)

	return nil
}

// BuildStep build step
func (s itsmApproveStep) BuildStep(kvs []common_task.KeyValue, opts ...types.StepOption) *types.Step {
	step := types.NewStep(s.GetName(), s.Alias(), opts...)

	// build step paras
	for _, v := range kvs {
		step.AddParam(v.Key.String(), v.Value)
	}

	return step
}
