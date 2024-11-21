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

package handler

import (
	"context"

	aquota "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ProjectQuotaHandler xxx
type ProjectQuotaHandler struct {
	model store.ProjectModel
}

// NewProjectQuota return a project quota service handler
func NewProjectQuota(model store.ProjectModel) *ProjectQuotaHandler {
	return &ProjectQuotaHandler{
		model: model,
	}
}

// CreateProjectQuota implement for CreateProjectQuota interface 申请项目维度的资源额度
func (p *ProjectQuotaHandler) CreateProjectQuota(ctx context.Context,
	req *proto.CreateProjectQuotaRequest, resp *proto.ProjectQuotaResponse) error {
	ca := aquota.NewCreateQuotaAction(p.model)

	e := ca.Do(ctx, req, resp)
	if e != nil {
		return e
	}
	return nil
}

// GetProjectQuota get project quota info
func (p *ProjectQuotaHandler) GetProjectQuota(ctx context.Context,
	req *proto.GetProjectQuotaRequest, resp *proto.ProjectQuotaResponse) error {
	ga := aquota.NewGetAction(p.model)
	err := ga.Do(ctx, req, resp)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProjectQuota update project quota
func (p *ProjectQuotaHandler) UpdateProjectQuota(ctx context.Context,
	req *proto.UpdateProjectQuotaRequest, resp *proto.ProjectQuotaResponse) error {
	ua := aquota.NewUpdateQuotaAction(p.model)
	e := ua.Do(ctx, req, resp)
	if e != nil {
		return e
	}

	return nil
}

// ScaleUpProjectQuota scaleUp project quota
func (p *ProjectQuotaHandler) ScaleUpProjectQuota(ctx context.Context,
	req *proto.ScaleUpProjectQuotaRequest, resp *proto.ScaleUpProjectQuotaResponse) error {
	ua := aquota.NewScaleUpQuotaAction(p.model)
	e := ua.Do(ctx, req, resp)
	if e != nil {
		return e
	}
	return nil
}

// ScaleDownProjectQuota scaleDown project quota
func (p *ProjectQuotaHandler) ScaleDownProjectQuota(ctx context.Context,
	req *proto.ScaleDownProjectQuotaRequest, resp *proto.ScaleDownProjectQuotaResponse) error {
	ua := aquota.NewScaleDownQuotaAction(p.model)
	e := ua.Do(ctx, req, resp)
	if e != nil {
		return e
	}
	return nil
}

// DeleteProjectQuota delete project quota record
func (p *ProjectQuotaHandler) DeleteProjectQuota(ctx context.Context,
	req *proto.DeleteProjectQuotaRequest, resp *proto.ProjectQuotaResponse) error {
	pa := aquota.NewDeleteQuotaAction(p.model)
	e := pa.Do(ctx, req, resp)
	if e != nil {
		return e
	}
	return nil
}

// ListProjectQuotas list project quotas records
func (p *ProjectQuotaHandler) ListProjectQuotas(ctx context.Context,
	req *proto.ListProjectQuotasRequest, resp *proto.ListProjectQuotasResponse) error {

	return nil
}

// GetProjectQuotasUsage get project quotas usage
func (p *ProjectQuotaHandler) GetProjectQuotasUsage(ctx context.Context,
	req *proto.GetProjectQuotasUsageReq, resp *proto.GetProjectQuotasUsageResp) error {

	return nil
}
