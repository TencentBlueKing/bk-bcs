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

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbts "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-space"
	pbds "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// CreateTemplateSpace create a template space
func (s *Service) CreateTemplateSpace(ctx context.Context, req *pbcs.CreateTemplateSpaceReq) (
	*pbcs.CreateTemplateSpaceResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	if req.Name == constant.DefaultTmplSpaceName || req.Name == constant.DefaultTmplSpaceCNName {
		return nil, fmt.Errorf("can't create template space %s which is created by system", req.Name)
	}

	r := &pbds.CreateTemplateSpaceReq{
		Attachment: &pbts.TemplateSpaceAttachment{
			BizId: grpcKit.BizID,
		},
		Spec: &pbts.TemplateSpaceSpec{
			Name: req.Name,
			Memo: req.Memo,
		},
	}
	rp, err := s.client.DS.CreateTemplateSpace(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create template space failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateTemplateSpaceResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteTemplateSpace delete a template space
func (s *Service) DeleteTemplateSpace(ctx context.Context, req *pbcs.DeleteTemplateSpaceReq) (
	*pbcs.DeleteTemplateSpaceResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.DeleteTemplateSpaceReq{
		Id: req.TemplateSpaceId,
		Attachment: &pbts.TemplateSpaceAttachment{
			BizId: grpcKit.BizID,
		},
	}
	if _, err := s.client.DS.DeleteTemplateSpace(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete template space failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.DeleteTemplateSpaceResp{}, nil
}

// UpdateTemplateSpace update a template space
func (s *Service) UpdateTemplateSpace(ctx context.Context, req *pbcs.UpdateTemplateSpaceReq) (
	*pbcs.UpdateTemplateSpaceResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.UpdateTemplateSpaceReq{
		Id: req.TemplateSpaceId,
		Attachment: &pbts.TemplateSpaceAttachment{
			BizId: grpcKit.BizID,
		},
		Spec: &pbts.TemplateSpaceSpec{
			Memo: req.Memo,
		},
	}
	if _, err := s.client.DS.UpdateTemplateSpace(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update template space failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UpdateTemplateSpaceResp{}, nil
}

// ListTemplateSpaces list template spaces
func (s *Service) ListTemplateSpaces(ctx context.Context, req *pbcs.ListTemplateSpacesReq) (
	*pbcs.ListTemplateSpacesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateSpacesReq{
		BizId:        grpcKit.BizID,
		SearchFields: req.SearchFields,
		SearchValue:  req.SearchValue,
		Start:        req.Start,
		Limit:        req.Limit,
		All:          req.All,
	}

	rp, err := s.client.DS.ListTemplateSpaces(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template spaces failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTemplateSpacesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// GetAllBizsOfTmplSpaces get all biz ids of template spaces
func (s *Service) GetAllBizsOfTmplSpaces(ctx context.Context, req *pbbase.EmptyReq) (
	*pbcs.GetAllBizsOfTmplSpacesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	rp, err := s.client.DS.GetAllBizsOfTmplSpaces(grpcKit.RpcCtx(), req)
	if err != nil {
		logs.Errorf("get all bizs of template space failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.GetAllBizsOfTmplSpacesResp{
		BizIds: rp.BizIds,
	}
	return resp, nil
}

// CreateDefaultTmplSpace create default template space
func (s *Service) CreateDefaultTmplSpace(ctx context.Context, req *pbcs.CreateDefaultTmplSpaceReq) (
	*pbcs.CreateDefaultTmplSpaceResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	r := &pbds.CreateDefaultTmplSpaceReq{
		BizId: req.BizId,
	}

	rp, err := s.client.DS.CreateDefaultTmplSpace(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create default template space failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateDefaultTmplSpaceResp{
		Id: rp.Id,
	}
	return resp, nil
}

// ListTmplSpacesByIDs list template spaces by ids
func (s *Service) ListTmplSpacesByIDs(ctx context.Context, req *pbcs.ListTmplSpacesByIDsReq) (*pbcs.
	ListTmplSpacesByIDsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	ids := tools.SliceRepeatedElements(req.Ids)
	if len(ids) > 0 {
		return nil, fmt.Errorf("repeated ids: %v, id must be unique", ids)
	}
	idsLen := len(req.Ids)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplSpacesByIDsReq{
		Ids: req.Ids,
	}

	rp, err := s.client.DS.ListTmplSpacesByIDs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template spaces failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplSpacesByIDsResp{
		Details: rp.Details,
	}
	return resp, nil
}
