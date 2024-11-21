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
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// NewQuotaManager init internal quota manager
func NewQuotaManager() provider.QuotaManager {
	return &quotaManager{}
}

// quotaManager is the manager for quota
type quotaManager struct{}

// CreateProjectQuota create project quota
func (q *quotaManager) CreateProjectQuota(quota *proto.ProjectQuota,
	opt *provider.CreateProjectQuotaOptions) (*types.Task, error) {
	task, err := newTaskManager().CreateProjectQuotaTask(quota, opt)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteProjectQuota build delete project quota task
func (q *quotaManager) DeleteProjectQuota(quotaId string,
	opt *provider.DeleteProjectQuotaOptions) (*types.Task, error) {
	task, err := newTaskManager().DeleteProjectQuotaTask(quotaId, opt)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// UpdateProjectQuota build update project quota task
func (q *quotaManager) UpdateProjectQuota(quotaId string,
	opt *provider.UpdateProjectQuotaOptions) (*types.Task, error) {
	task, err := newTaskManager().UpdateProjectQuotaTask(quotaId, opt)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// ScaleUpProjectQuota build scale up project quota task
func (q *quotaManager) ScaleUpProjectQuota(quotaId string, quota *proto.QuotaResource,
	opt *provider.ScaleUpProjectQuotaOptions) (*types.Task, error) {
	task, err := newTaskManager().ScaleUpProjectQuotaTask(quotaId, quota, opt)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// ScaleDownProjectQuota build scale doen project quota task
func (q *quotaManager) ScaleDownProjectQuota(quotaId string, quota *proto.QuotaResource,
	opt *provider.ScaleDownProjectQuotaOptions) (*types.Task, error) {
	task, err := newTaskManager().ScaleDownProjectQuotaTask(quotaId, quota, opt)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetProjectQuotaById get projectQuota by quotaId
func (q *quotaManager) GetProjectQuotaById(quotaId string) (*proto.ProjectQuota, error) {
	sQuota, err := store.GetModel().GetProjectQuotaById(context.Background(), quotaId)
	if err != nil {
		return nil, err
	}

	return quota.TransStore2ProtoQuota(sQuota), nil
}

// ListProjectQuotasByProjectId list project quotas by projectIdOrCode
func (q *quotaManager) ListProjectQuotasByProjectId(projectIdOrCode string) ([]*proto.ProjectQuota, error) {
	return nil, utils.NewNotImplemented(ProviderName)
}

// ListProjectQuotasByBizId list project quotas by projectIdOrCode
func (q *quotaManager) ListProjectQuotasByBizId(bizId string) ([]*proto.ProjectQuota, error) {
	return nil, utils.NewNotImplemented(ProviderName)
}

// GetProjectQuotaUsage get projectQuota usage by quotaId
func (q *quotaManager) GetProjectQuotaUsage(quotaId string) (*proto.ProjectQuota, error) {
	return nil, utils.NewNotImplemented(ProviderName)
}
