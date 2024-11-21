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
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	isteps "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/internal/steps"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// NewScaleDownProjectQuotaTask build scale down project quota task
func NewScaleDownProjectQuotaTask(quotaId string, resource *bcsproject.QuotaResource,
	operator string) task.TaskBuilder {

	quotaInfo, err := store.GetModel().GetProjectQuotaById(context.Background(), quotaId)
	if err != nil {
		blog.Errorf("get project quota %s failed, err %s", quotaId, err.Error())
		return nil
	}

	return &scaleDownProjectQuota{
		quotaId:      quotaId,
		projectQuota: quotaInfo,
		resource:     resource,
		operator:     operator,
	}
}

// scaleDownProjectQuota task
type scaleDownProjectQuota struct {
	quotaId      string
	projectQuota *quota.ProjectQuota
	resource     *bcsproject.QuotaResource
	operator     string
}

// Name 任务名称
func (cpq *scaleDownProjectQuota) Name() string {
	return utils.ScaleDownProjectQuota.GetTaskName()
}

// Type 任务类型
func (cpq *scaleDownProjectQuota) Type() string {
	return utils.ScaleDownProjectQuota.GetTaskType(ProviderName)
}

// Steps 构建任务step
func (cpq *scaleDownProjectQuota) Steps(defineSteps []task.StepBuilder) []*types.Step {
	stepList := make([]*types.Step, 0)
	// 项目配额支持3种类型: CA整机资源配额管理、共享集群命名空间配额管理、联邦集群子集群配额管理
	switch cpq.projectQuota.QuotaType {
	case quota.Host:
		return cpq.buildScaleDownHostQuotaSteps()
	case quota.Shared:
		return cpq.buildScaleDownSharedQuotaSteps()
	case quota.Federation:
		return cpq.buildScaleDownFederationQuotaSteps()
	default:
	}
	return stepList
}

func (cpq *scaleDownProjectQuota) buildScaleDownHostQuotaSteps() []*types.Step {
	// CA整机资源配额管理:
	// 0. 提交用户调减资源申请单据，并等待审批(调减时需要检测当前正在使用额度)
	// 2. 预期等待执行时间 (可增加资源回收计费策略等)
	// 3. 下发CA整机配额
	stepList := make([]*types.Step, 0)

	// 1. 审批调增整机资源配额申请
	// 2. 等待审批通过，审批通过后, 更新配额大小; 审批拒绝后, 更新配额申请状态
	content := fmt.Sprintf("user %s scale down project %s CA resources quota: region(%s) zone(%s) instanceType(%s) "+
		"quota(%v)", cpq.operator, cpq.projectQuota.ProjectCode,
		cpq.projectQuota.Quota.HostResources.Region, cpq.projectQuota.Quota.HostResources.ZoneId,
		cpq.projectQuota.Quota.HostResources.InstanceType, cpq.resource.GetZoneResources().GetQuotaNum())

	stepList = append(stepList, buildItsmQuotaSteps("", itsmData{
		operator:    cpq.operator,
		projectCode: cpq.projectQuota.ProjectCode,
		clusterId:   "",
		content:     content,
	})...)

	return stepList
}

func (cpq *scaleDownProjectQuota) buildScaleDownSharedQuotaSteps() []*types.Step {
	// 共享集群命名空间配额管理:
	// 审批共享集群配额申请
	// 下发共享集群配额, 主要用于某个共享集群的命名空间配额
	stepList := make([]*types.Step, 0)

	// 1. 审批共享集群配额申请
	// 2. 等待审批通过，审批通过后, 更新配额状态; 审批拒绝后, 更新配额申请状态
	content := fmt.Sprintf("user %s scale down shared cluster %s quota: cpu(%s) mem(%s)",
		cpq.operator, cpq.projectQuota.ClusterId,
		cpq.resource.GetCpu().GetDeviceQuota(), cpq.resource.GetMem().GetDeviceQuota())

	stepList = append(stepList, buildItsmQuotaSteps("", itsmData{
		operator:    cpq.operator,
		projectCode: cpq.projectQuota.ProjectCode,
		clusterId:   cpq.projectQuota.ClusterId,
		content:     content,
	})...)

	return stepList
}

func (cpq *scaleDownProjectQuota) buildScaleDownFederationQuotaSteps() []*types.Step {
	// 联邦集群配额管理:
	// 1. 提交审批联邦配额调整申请
	// 2. 审批联邦集群配额调整申请
	// 3. 下发联邦配额, 主要用于某个子集群的配额
	stepList := make([]*types.Step, 0)

	// 1. 审批联邦集群配额申请
	// 2. 等待审批通过，审批通过后, 更新配额状态; 审批拒绝后, 更新配额申请状态
	content := fmt.Sprintf("user %s scale down federation cluster %s quota: cpu(%s) mem(%s) gpu(%s)",
		cpq.operator, cpq.projectQuota.ClusterId, cpq.resource.GetCpu().GetDeviceQuota(),
		cpq.resource.GetMem().GetDeviceQuota(), cpq.resource.GetGpu().GetDeviceQuota())
	stepList = append(stepList, buildItsmQuotaSteps("", itsmData{
		operator:    cpq.operator,
		projectCode: cpq.projectQuota.ProjectCode,
		clusterId:   cpq.projectQuota.ClusterId,
		content:     content,
	})...)

	// 3. 下发联邦集群quota资源
	federationQuota := isteps.NewFederationQuotaStep()

	quotaParams := isteps.FederationQuotaStepParams{
		Operation: isteps.OperationUpdate,
		Scale:     isteps.ScaleDown,
		QuotaId:   cpq.quotaId,
		Name:      cpq.projectQuota.QuotaName,
		ClusterId: cpq.projectQuota.ClusterId,
		NameSpace: cpq.projectQuota.Namespace,
		Cpu:       cpq.resource.GetCpu(),
		Mem:       cpq.resource.GetMem(),
		Gpu:       cpq.resource.GetGpu(),
	}
	kvs, err := quotaParams.BuildParams()
	if err != nil {
		return nil
	}
	quotaStep := federationQuota.BuildStep(kvs)
	stepList = append(stepList, quotaStep)

	return stepList
}

func (cpq *scaleDownProjectQuota) BuildTask(info types.TaskInfo, opts ...types.TaskOption) (*types.Task, error) {
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
	t.AddCommonParams(utils.TaskType.String(), utils.ScaleDownProjectQuota.GetJobType())
	t.AddCommonParams(utils.QuotaIdKey.String(), cpq.quotaId)

	// 注入调减资源详情
	resourceBytes, _ := json.Marshal(cpq.resource)
	t.AddCommonParams(utils.QuotaResource.String(), string(resourceBytes))

	return t, nil
}
