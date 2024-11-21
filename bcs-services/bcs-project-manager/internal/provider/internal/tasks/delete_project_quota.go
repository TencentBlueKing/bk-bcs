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
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	isteps "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/internal/steps"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
)

// NewDeleteProjectQuotaTask build delete quota task
func NewDeleteProjectQuotaTask(quotaId, operator string) task.TaskBuilder {
	return &deleteProjectQuota{
		QuotaId:  quotaId,
		Operator: operator,
	}
}

// deleteProjectQuota task
type deleteProjectQuota struct {
	QuotaId  string
	Operator string
}

// Name 任务名称
func (dpq *deleteProjectQuota) Name() string {
	return utils.DeleteProjectQuota.GetTaskName()
}

// Type 任务类型
func (dpq *deleteProjectQuota) Type() string {
	return utils.DeleteProjectQuota.GetTaskType(ProviderName)
}

// Steps 构建任务step
func (dpq *deleteProjectQuota) Steps(defineSteps []task.StepBuilder) []*types.Step {

	quotaData, err := store.GetModel().GetProjectQuotaById(context.Background(), dpq.QuotaId)
	if err != nil {
		return nil
	}

	// 删除不同类型的quota额度, 下发任务前需要在逻辑层处理校验是否符合删除条件
	// 项目配额支持3种类型: CA整机资源配额管理、共享集群命名空间配额管理、联邦集群子集群配额管理
	switch quotaData.QuotaType {
	case quota.Host:
		return dpq.buildCaInstanceTypesQuotaSteps(quotaData)
	case quota.Shared:
		return dpq.buildSharedClusterQuotaSteps(quotaData)
	case quota.Federation:
		steps, errLocal := dpq.buildFederationClusterQuotaSteps(quotaData)
		if errLocal != nil {
			logging.Error("build federation cluster quota steps failed: %s", dpq.QuotaId, errLocal.Error())
			return nil
		}
		return steps
	default:
	}
	return nil
}

func (dpq *deleteProjectQuota) buildCaInstanceTypesQuotaSteps(projectQuota *quota.ProjectQuota) []*types.Step {
	// CA整机资源配额删除管理:
	// 1. 提交CA整机删除申请
	// 2. 审批CA整机删除单据
	// 3. 额度退还策略/额度退换计费等
	// 4. 回调软删除额度数据
	stepList := make([]*types.Step, 0)

	// 1. 审批联邦集群配额申请
	// 2. 等待审批通过，审批通过后, 更新配额状态; 审批拒绝后, 更新配额申请状态
	content := fmt.Sprintf("user %s revoke for project %s CA resources quota(%s): region(%s) zone(%s) "+
		"instanceType(%s) quota(%v)", dpq.Operator, projectQuota.ProjectCode, projectQuota.QuotaName,
		projectQuota.Quota.HostResources.Region, projectQuota.Quota.HostResources.ZoneId,
		projectQuota.Quota.HostResources.InstanceType, projectQuota.Quota.HostResources.QuotaNum)
	stepList = append(stepList, buildItsmQuotaSteps("", itsmData{
		operator:    dpq.Operator,
		projectCode: projectQuota.ProjectCode,
		clusterId:   "",
		content:     content,
	})...)

	return stepList
}

func (dpq *deleteProjectQuota) buildSharedClusterQuotaSteps(projectQuota *quota.ProjectQuota) []*types.Step {
	// 共享集群命名空间配额管理:
	// 审批共享集群配额回收 (是否可以删除, 业务逻辑层提前check)
	// 回调软删除
	stepList := make([]*types.Step, 0)

	// 1. 审批联邦集群配额申请
	// 2. 等待审批通过，审批通过后, 更新配额状态; 审批拒绝后, 更新配额申请状态
	content := fmt.Sprintf("user %s revoke for project %s shared cluster(%s) quota(%s): cpu(%s) mem(%s)",
		dpq.Operator, projectQuota.ProjectCode, projectQuota.ClusterId, projectQuota.QuotaName,
		projectQuota.Quota.Cpu.DeviceQuota, projectQuota.Quota.Mem.DeviceQuota)
	stepList = append(stepList, buildItsmQuotaSteps("", itsmData{
		operator:    dpq.Operator,
		projectCode: projectQuota.ProjectCode,
		clusterId:   "",
		content:     content,
	})...)

	return stepList
}

func (dpq *deleteProjectQuota) buildFederationClusterQuotaSteps(
	projectQuota *quota.ProjectQuota) ([]*types.Step, error) {
	// 联邦集群配额管理:
	// 1. 提交审批联邦配额申请, 审批联邦集群配额申请
	// 2. 删除联邦集群配额(一一对应关系)
	stepList := make([]*types.Step, 0)

	// 1. 审批联邦集群配额申请
	// 2. 等待审批通过，审批通过后, 更新配额状态; 审批拒绝后, 更新配额申请状态
	content := fmt.Sprintf("user %s revoke for federation cluster %s namespace %s quota: cpu(%s) mem(%s) gpu(%s)",
		dpq.Operator, projectQuota.ClusterId, projectQuota.Namespace,
		projectQuota.Quota.Cpu.DeviceQuota, projectQuota.Quota.Mem.DeviceQuota, projectQuota.Quota.Gpu.DeviceQuota)
	stepList = append(stepList, buildItsmQuotaSteps("", itsmData{
		operator:    dpq.Operator,
		projectCode: projectQuota.ProjectCode,
		clusterId:   projectQuota.ClusterId,
		content:     content,
	})...)

	// 3. 删除联邦集群quota资源(需要前置检测资源使用情况,是否在使用中)
	federationQuota := isteps.NewFederationQuotaStep()

	quotaparams := isteps.FederationQuotaStepParams{
		Operation: isteps.OperationDelete,
		QuotaId:   dpq.QuotaId,
		ClusterId: projectQuota.ClusterId,
		NameSpace: projectQuota.Namespace,
		Name:      projectQuota.QuotaName,
	}
	kvs, err := quotaparams.BuildParams()
	if err != nil {
		logging.Info("deleteProjectQuota buildFederationClusterQuotaSteps[%s] failed: %v",
			projectQuota.QuotaId, err)
		return nil, err
	}
	quotaStep := federationQuota.BuildStep(kvs)
	stepList = append(stepList, quotaStep)

	return stepList, nil
}

func (dpq *deleteProjectQuota) BuildTask(info types.TaskInfo, opts ...types.TaskOption) (*types.Task, error) {
	t := types.NewTask(&info, opts...)
	if len(dpq.Steps(nil)) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	for _, step := range dpq.Steps(nil) {
		t.Steps[step.GetAlias()] = step
		t.StepSequence = append(t.StepSequence, step.GetAlias())
	}
	t.CurrentStep = t.StepSequence[0]

	// 任务类型，可通过回调函数进行配额的调整 & 状态更新
	if t.CommonParams == nil {
		t.CommonParams = make(map[string]string, 0)
	}
	t.AddCommonParams(utils.TaskType.String(), utils.DeleteProjectQuota.GetJobType())
	t.AddCommonParams(utils.QuotaIdKey.String(), dpq.QuotaId)

	return t, nil
}
