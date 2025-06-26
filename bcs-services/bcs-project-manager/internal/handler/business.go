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

// Package handler xxx
package handler

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// BusinessHandler xxx
type BusinessHandler struct {
	model store.ProjectModel
}

// NewBusiness return a business service hander
func NewBusiness(model store.ProjectModel) *BusinessHandler {
	return &BusinessHandler{
		model: model,
	}
}

// GetBusiness implement for GetBusiness interface
func (p *BusinessHandler) GetBusiness(ctx context.Context,
	req *proto.GetBusinessRequest, resp *proto.GetBusinessResponse) error {

	authUser := tenant.GetAuthAndTenantInfoFromCtx(ctx)
	ctx = tenant.WithTenantIdFromContext(ctx, authUser.ResourceTenantId)

	// 查询业务信息
	ga := business.NewGetAction(p.model)
	businessInfo, e := ga.Do(ctx, req)
	if e != nil {
		return e
	}
	// 处理返回数据及权限
	resp.Data = businessInfo
	return nil
}

// ListBusiness query authorized business info list
func (p *BusinessHandler) ListBusiness(ctx context.Context,
	req *proto.ListBusinessRequest, resp *proto.ListBusinessResponse) error {

	authUser := tenant.GetAuthAndTenantInfoFromCtx(ctx)
	ctx = tenant.WithTenantIdFromContext(ctx, authUser.ResourceTenantId)

	la := business.NewListAction(p.model)
	biz, e := la.Do(ctx, req)
	if e != nil {
		return e
	}
	resp.Data = biz
	return nil
}

// GetBusinessTopology query business topology info
func (p *BusinessHandler) GetBusinessTopology(ctx context.Context,
	req *proto.GetBusinessTopologyRequest, resp *proto.GetBusinessTopologyResponse) error {

	authUser := tenant.GetAuthAndTenantInfoFromCtx(ctx)
	ctx = tenant.WithTenantIdFromContext(ctx, authUser.ResourceTenantId)

	ga := business.NewGetTopologyAction(p.model)
	if err := ga.Do(ctx, req, resp); err != nil {
		return err
	}
	return nil
}
