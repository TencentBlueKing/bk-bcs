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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	isteps "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/internal/steps"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// NewCreateProjectQuotaTask build create project quota task
func NewCreateProjectQuotaTask(quota *bcsproject.ProjectQuota) task.TaskBuilder {
	return &createProjectQuota{
		projectQuota: quota,
	}
}

// createProjectQuota task
type createProjectQuota struct {
	projectQuota *bcsproject.ProjectQuota
}

// Name 任务名称
func (cpq *createProjectQuota) Name() string {
	return utils.CreateProjectQuota.GetTaskName()
}

// Type 任务类型
func (cpq *createProjectQuota) Type() string {
	return utils.CreateProjectQuota.GetTaskType(ProviderName)
}

// Steps 构建任务step
func (cpq *createProjectQuota) Steps(defineSteps []task.StepBuilder) []*types.Step {
	// 项目配额支持3种类型: CA整机资源配额管理、共享集群命名空间配额管理、联邦集群子集群配额管理
	switch cpq.projectQuota.QuotaType {
	case quota.Host.String():
		return cpq.buildCaInstanceTypesQuotaSteps()
	case quota.Shared.String():
		return cpq.buildSharedClusterQuotaSteps()
	case quota.Federation.String():
		steps, err := cpq.buildFederationClusterQuotaSteps()
		if err != nil {
			logging.Error("build federation cluster quota steps failed: %s", cpq.projectQuota.QuotaId, err.Error())
			return nil
		}
		return steps
	default:
	}
	return nil
}

func (cpq *createProjectQuota) buildCaInstanceTypesQuotaSteps() []*types.Step {
	// CA整机资源配额管理:
	// 0. 提交CA整机资源预测单据
	// 1. 提交CA整机资源配额申请
	// 2. 审批CA整机资源配额申请
	// 3. 预期等待执行时间，并检测海垒预测单据是否通过
	// 4. 下发CA整机配额
	stepList := make([]*types.Step, 0)

	// 1. 审批联邦集群配额申请
	// 2. 等待审批通过，审批通过后, 更新配额状态; 审批拒绝后, 更新配额申请状态
	content := fmt.Sprintf("user %s apply for project %s CA resources quota: region(%s) "+ // nolint
		"zone(%s) instanceType(%s) quota(%v)", cpq.projectQuota.GetCreator(), cpq.projectQuota.GetProjectCode(),
		cpq.projectQuota.GetQuota().GetZoneResources().Region, cpq.projectQuota.GetQuota().GetZoneResources().ZoneId,
		cpq.projectQuota.GetQuota().GetZoneResources().InstanceType,
		cpq.projectQuota.GetQuota().GetZoneResources().GetQuotaNum())

	stepList = append(stepList, buildItsmQuotaSteps("", itsmData{
		operator:    cpq.projectQuota.GetCreator(),
		projectCode: cpq.projectQuota.GetProjectCode(),
		clusterId:   "",
		content:     content,
	})...)

	return stepList
}

func (cpq *createProjectQuota) buildSharedClusterQuotaSteps() []*types.Step {
	// 共享集群命名空间配额管理:
	// 审批共享集群配额申请
	// 下发共享集群配额, 主要用于某个共享集群的命名空间配额
	stepList := make([]*types.Step, 0)

	// 1. 审批共享集群配额申请
	// 2. 等待审批通过，审批通过后, 更新配额状态; 审批拒绝后, 更新配额申请状态
	content := fmt.Sprintf("user %s apply for shared cluster %s quota: cpu(%s) mem(%s)",
		cpq.projectQuota.GetCreator(), cpq.projectQuota.GetClusterId(),
		cpq.projectQuota.GetQuota().GetCpu().GetDeviceQuota(), cpq.projectQuota.GetQuota().GetMem().GetDeviceQuota())

	stepList = append(stepList, buildItsmQuotaSteps("", itsmData{
		operator:    cpq.projectQuota.GetCreator(),
		projectCode: cpq.projectQuota.GetProjectCode(),
		clusterId:   "",
		content:     content,
	})...)

	return stepList
}

func (cpq *createProjectQuota) buildFederationClusterQuotaSteps() ([]*types.Step, error) {
	// 联邦集群配额管理:
	// 1. 提交审批联邦配额申请
	// 2. 审批联邦集群配额申请
	// 3. 下发联邦配额, 主要用于某个子集群的配额
	stepList := make([]*types.Step, 0)

	// 1. 审批联邦集群配额申请
	// 2. 等待审批通过，审批通过后, 更新配额状态; 审批拒绝后, 更新配额申请状态
	content := fmt.Sprintf("user %s apply for federation cluster %s quota: cpu(%s) mem(%s) gpu(%s)",
		cpq.projectQuota.GetCreator(), cpq.projectQuota.GetClusterId(),
		cpq.projectQuota.GetQuota().GetCpu().GetDeviceQuota(), cpq.projectQuota.GetQuota().GetMem().GetDeviceQuota(),
		cpq.projectQuota.GetQuota().GetGpu().GetDeviceQuota())

	stepList = append(stepList, buildItsmQuotaSteps("", itsmData{
		operator:    cpq.projectQuota.GetCreator(),
		projectCode: cpq.projectQuota.GetProjectCode(),
		clusterId:   "",
		content:     content,
	})...)

	// 3. 下发联邦集群quota资源
	federationQuota := isteps.NewFederationQuotaStep()

	quotaparams := isteps.FederationQuotaStepParams{
		Operation: isteps.OperationCreate,
		QuotaId:   cpq.projectQuota.GetQuotaId(),
		NameSpace: cpq.projectQuota.GetNameSpace(),
		Name:      cpq.projectQuota.GetQuotaName(),
		Cpu:       cpq.projectQuota.GetQuota().GetCpu(),
		Mem:       cpq.projectQuota.GetQuota().GetMem(),
		Gpu:       cpq.projectQuota.GetQuota().GetGpu(),
	}
	kvs, err := quotaparams.BuildParams()
	if err != nil {
		logging.Info("createProjectQuota buildFederationClusterQuotaSteps[%s] failed: %v",
			cpq.projectQuota.GetQuotaId(), err)
		return nil, err
	}
	quotaStep := federationQuota.BuildStep(kvs)
	stepList = append(stepList, quotaStep)

	return stepList, nil
}

func (cpq *createProjectQuota) BuildTask(info types.TaskInfo, opts ...types.TaskOption) (*types.Task, error) {
	t := types.NewTask(&info, opts...)
	if len(cpq.Steps(nil)) == 0 {
		return nil, fmt.Errorf("task steps empty")
	}

	for _, step := range cpq.Steps(nil) {
		t.Steps[step.GetAlias()] = step
		t.StepSequence = append(t.StepSequence, step.GetAlias())
	}
	t.CurrentStep = t.StepSequence[0]

	// 任务类型，可通过回调函数进行配额的调整 & 状态更新
	if t.CommonParams == nil {
		t.CommonParams = make(map[string]string, 0)
	}
	t.AddCommonParams(utils.TaskType.String(), utils.CreateProjectQuota.GetJobType())
	t.AddCommonParams(utils.QuotaIdKey.String(), cpq.projectQuota.GetQuotaId())

	return t, nil
}
