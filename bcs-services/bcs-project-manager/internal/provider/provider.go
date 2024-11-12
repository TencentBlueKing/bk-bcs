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

// Package provider xxx
package provider

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// 业务针对不同provider的抽象逻辑层，可根据provider对接不同的系统

// ValidateManager validate interface for check provider resource
type ValidateManager interface {
}

// QuotaManager interface for quota resource management, handler logic && task builder
type QuotaManager interface {
	// CreateProjectQuota create project quota
	CreateProjectQuota(quota *proto.ProjectQuota, opt *CreateProjectQuotaOptions) (*types.Task, error)
	// DeleteProjectQuota build delete project quota task
	DeleteProjectQuota(quotaId string, opt *DeleteProjectQuotaOptions) (*types.Task, error)
	// UpdateProjectQuota build update project quota task
	UpdateProjectQuota(quotaId string, opt *UpdateProjectQuotaOptions) (*types.Task, error)
	// ScaleUpProjectQuota build scale up project quota task
	ScaleUpProjectQuota(quotaId string, quota *proto.QuotaResource,
		opt *ScaleUpProjectQuotaOptions) (*types.Task, error)
	// ScaleDownProjectQuota build scale down project quota task
	ScaleDownProjectQuota(quotaId string, quota *proto.QuotaResource,
		opt *ScaleDownProjectQuotaOptions) (*types.Task, error)
	// GetProjectQuotaById get projectQuota by quotaId
	GetProjectQuotaById(quotaId string) (*proto.ProjectQuota, error)
	// ListProjectQuotasByProjectId list project quotas by projectIdOrCode
	ListProjectQuotasByProjectId(projectIdOrCode string) ([]*proto.ProjectQuota, error)
	// ListProjectQuotasByBizId list project quotas by projectIdOrCode
	ListProjectQuotasByBizId(bizId string) ([]*proto.ProjectQuota, error)
	// GetProjectQuotaUsage get projectQuota usage by quotaId
	GetProjectQuotaUsage(quotaId string) (*proto.ProjectQuota, error)
}

// CapacityManager interface for capacity management (shared cluster && CA resource capacity)
type CapacityManager interface {
}

// TaskManager interface for async platform task management
type TaskManager interface {
	Name() string
	// CreateProjectQuotaTask build create project quota task
	CreateProjectQuotaTask(quota *proto.ProjectQuota, opt *CreateProjectQuotaOptions) (*types.Task, error)
	// DeleteProjectQuotaTask build delete project quota task
	DeleteProjectQuotaTask(quotaId string, opt *DeleteProjectQuotaOptions) (*types.Task, error)
	// UpdateProjectQuotaTask build update project quota task
	UpdateProjectQuotaTask(quotaId string, opt *UpdateProjectQuotaOptions) (*types.Task, error)
	// ScaleUpProjectQuotaTask build scale up project quota task
	ScaleUpProjectQuotaTask(quotaId string, quota *proto.QuotaResource,
		opt *ScaleUpProjectQuotaOptions) (*types.Task, error)
	// ScaleDownProjectQuotaTask build scale down project quota task
	ScaleDownProjectQuotaTask(quotaId string, quota *proto.QuotaResource,
		opt *ScaleDownProjectQuotaOptions) (*types.Task, error)
}

// CreateProjectQuotaOptions options params
type CreateProjectQuotaOptions struct {
}

// DeleteProjectQuotaOptions options params
type DeleteProjectQuotaOptions struct {
	Operator string
}

// UpdateProjectQuotaOptions options params
type UpdateProjectQuotaOptions struct {
}

// ScaleUpProjectQuotaOptions options params
type ScaleUpProjectQuotaOptions struct {
	Operator string
}

// ScaleDownProjectQuotaOptions options params
type ScaleDownProjectQuotaOptions struct {
	Operator string
}

// CommonOptions options params
type CommonOptions struct {
}
