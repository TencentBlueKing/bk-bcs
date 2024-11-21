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

// Package internal xxx
package internal

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/common/callback"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/internal/tasks"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// newTaskManager create task manager
func newTaskManager() provider.TaskManager {
	return &taskManager{}
}

type taskManager struct{}

// Name task provider name
func (t *taskManager) Name() string {
	return ProviderName
}

// CreateProjectQuotaTask build create project quota task
func (t *taskManager) CreateProjectQuotaTask(quota *proto.ProjectQuota,
	opt *provider.CreateProjectQuotaOptions) (*types.Task, error) {
	builder := tasks.NewCreateProjectQuotaTask(*quota)

	return builder.BuildTask(types.TaskInfo{
		TaskIndex: quota.QuotaId,
		TaskType:  utils.CreateProjectQuota.GetTaskType(ProviderName),
		TaskName:  utils.CreateProjectQuota.GetTaskName(),
		Creator:   quota.Creator,
	}, types.WithTaskCallBackFunc(callback.QuotaCallBackName))
}

// DeleteProjectQuotaTask build delete project quota task
func (t *taskManager) DeleteProjectQuotaTask(quotaId string,
	opt *provider.DeleteProjectQuotaOptions) (*types.Task, error) {
	builder := tasks.NewDeleteProjectQuotaTask(quotaId, opt.Operator)

	return builder.BuildTask(types.TaskInfo{
		TaskIndex: quotaId,
		TaskType:  utils.DeleteProjectQuota.GetTaskType(ProviderName),
		TaskName:  utils.DeleteProjectQuota.GetTaskName(),
		Creator:   opt.Operator,
	}, types.WithTaskCallBackFunc(callback.QuotaCallBackName))
}

// UpdateProjectQuotaTask build update project quota task
func (t *taskManager) UpdateProjectQuotaTask(quotaId string,
	opt *provider.UpdateProjectQuotaOptions) (*types.Task, error) {
	return nil, nil
}

// ScaleUpProjectQuotaTask build scale up project quota task
func (t *taskManager) ScaleUpProjectQuotaTask(quotaId string, quota *proto.QuotaResource,
	opt *provider.ScaleUpProjectQuotaOptions) (*types.Task, error) {
	builder := tasks.NewScaleUpProjectQuotaTask(quotaId, quota, opt.Operator)

	return builder.BuildTask(types.TaskInfo{
		TaskIndex: quotaId,
		TaskType:  utils.ScaleUpProjectQuota.GetTaskType(ProviderName),
		TaskName:  utils.ScaleUpProjectQuota.GetTaskName(),
		Creator:   opt.Operator,
	}, types.WithTaskCallBackFunc(callback.QuotaCallBackName))
}

// ScaleDownProjectQuotaTask build scale doen project quota task
func (t *taskManager) ScaleDownProjectQuotaTask(quotaId string, quota *proto.QuotaResource,
	opt *provider.ScaleDownProjectQuotaOptions) (*types.Task, error) {
	builder := tasks.NewScaleDownProjectQuotaTask(quotaId, quota, opt.Operator)

	return builder.BuildTask(types.TaskInfo{
		TaskIndex: quotaId,
		TaskType:  utils.ScaleDownProjectQuota.GetTaskType(ProviderName),
		TaskName:  utils.ScaleDownProjectQuota.GetTaskName(),
		Creator:   opt.Operator,
	}, types.WithTaskCallBackFunc(callback.QuotaCallBackName))
}
