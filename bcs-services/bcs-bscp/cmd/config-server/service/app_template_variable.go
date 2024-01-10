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

package service

import (
	"context"
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbatv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app-template-variable"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// ExtractAppTmplVariables extract app template variables
func (s *Service) ExtractAppTmplVariables(ctx context.Context, req *pbcs.ExtractAppTmplVariablesReq) (
	*pbcs.ExtractAppTmplVariablesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ExtractAppTmplVariablesReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}

	rp, err := s.client.DS.ExtractAppTmplVariables(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("extract app template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ExtractAppTmplVariablesResp{
		Details: rp.Details,
	}
	return resp, nil
}

// GetAppTmplVariableRefs get app template variable references
func (s *Service) GetAppTmplVariableRefs(ctx context.Context, req *pbcs.GetAppTmplVariableRefsReq) (
	*pbcs.GetAppTmplVariableRefsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.GetAppTmplVariableRefsReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}

	rp, err := s.client.DS.GetAppTmplVariableRefs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get app template variable references failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.GetAppTmplVariableRefsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// GetReleasedAppTmplVariableRefs get released app template variable references
func (s *Service) GetReleasedAppTmplVariableRefs(ctx context.Context, req *pbcs.GetReleasedAppTmplVariableRefsReq) (
	*pbcs.GetReleasedAppTmplVariableRefsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	if req.ReleaseId <= 0 {
		return nil, fmt.Errorf("invalid release id %d, it must bigger than 0", req.ReleaseId)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.GetReleasedAppTmplVariableRefsReq{
		BizId:     req.BizId,
		AppId:     req.AppId,
		ReleaseId: req.ReleaseId,
	}

	rp, err := s.client.DS.GetReleasedAppTmplVariableRefs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get app template variable references failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.GetReleasedAppTmplVariableRefsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListAppTmplVariables list app template variables
func (s *Service) ListAppTmplVariables(ctx context.Context, req *pbcs.ListAppTmplVariablesReq) (
	*pbcs.ListAppTmplVariablesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListAppTmplVariablesReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}

	rp, err := s.client.DS.ListAppTmplVariables(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list app template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListAppTmplVariablesResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListReleasedAppTmplVariables list released app template variables
func (s *Service) ListReleasedAppTmplVariables(ctx context.Context, req *pbcs.ListReleasedAppTmplVariablesReq) (
	*pbcs.ListReleasedAppTmplVariablesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	if req.ReleaseId <= 0 {
		return nil, fmt.Errorf("invalid release id %d, it must bigger than 0", req.ReleaseId)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListReleasedAppTmplVariablesReq{
		BizId:     req.BizId,
		AppId:     req.AppId,
		ReleaseId: req.ReleaseId,
	}

	rp, err := s.client.DS.ListReleasedAppTmplVariables(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list released app template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListReleasedAppTmplVariablesResp{
		Details: rp.Details,
	}
	return resp, nil
}

// UpdateAppTmplVariables update app template variables
func (s *Service) UpdateAppTmplVariables(ctx context.Context, req *pbcs.UpdateAppTmplVariablesReq) (
	*pbcs.UpdateAppTmplVariablesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.UpdateAppTmplVariablesReq{
		Attachment: &pbatv.AppTemplateVariableAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbatv.AppTemplateVariableSpec{
			Variables: req.Variables,
		},
	}

	_, err := s.client.DS.UpdateAppTmplVariables(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("update app template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UpdateAppTmplVariablesResp{}, nil
}
