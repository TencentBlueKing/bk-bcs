/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbatb "bscp.io/pkg/protocol/core/app-template-binding"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateAppTemplateBinding create app template binding.
func (s *Service) CreateAppTemplateBinding(ctx context.Context, req *pbds.CreateAppTemplateBindingReq) (*pbds.
	CreateResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	AppTemplateBinding := &table.AppTemplateBinding{
		Spec:       req.Spec.AppTemplateBindingSpec(),
		Attachment: req.Attachment.AppTemplateBindingAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	id, err := s.dao.AppTemplateBinding().Create(kt, AppTemplateBinding)
	if err != nil {
		logs.Errorf("create app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListAppTemplateBindings list app template binding.
func (s *Service) ListAppTemplateBindings(ctx context.Context, req *pbds.ListAppTemplateBindingsReq) (*pbds.
	ListAppTemplateBindingsResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	details, count, err := s.dao.AppTemplateBinding().List(kt, req.BizId, req.AppId, opt)

	if err != nil {
		logs.Errorf("list app template bindings failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListAppTemplateBindingsResp{
		Count:   uint32(count),
		Details: pbatb.PbAppTemplateBindings(details),
	}
	return resp, nil
}

// UpdateAppTemplateBinding update app template binding.
func (s *Service) UpdateAppTemplateBinding(ctx context.Context, req *pbds.UpdateAppTemplateBindingReq) (*pbbase.
	EmptyResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	AppTemplateBinding := &table.AppTemplateBinding{
		ID:         req.Id,
		Spec:       req.Spec.AppTemplateBindingSpec(),
		Attachment: req.Attachment.AppTemplateBindingAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if err := s.dao.AppTemplateBinding().Update(kt, AppTemplateBinding); err != nil {
		logs.Errorf("update app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteAppTemplateBinding delete app template binding.
func (s *Service) DeleteAppTemplateBinding(ctx context.Context, req *pbds.DeleteAppTemplateBindingReq) (*pbbase.
	EmptyResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	AppTemplateBinding := &table.AppTemplateBinding{
		ID:         req.Id,
		Attachment: req.Attachment.AppTemplateBindingAttachment(),
	}
	if err := s.dao.AppTemplateBinding().Delete(kt, AppTemplateBinding); err != nil {
		logs.Errorf("delete app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
