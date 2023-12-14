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
	pbtset "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-set"
	pbds "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// CreateTemplateSet create a template set
func (s *Service) CreateTemplateSet(ctx context.Context, req *pbcs.CreateTemplateSetReq) (*pbcs.CreateTemplateSetResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	idsLen := len(req.TemplateIds)
	if idsLen > 500 {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [0,500]",
			idsLen)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.CreateTemplateSetReq{
		Attachment: &pbtset.TemplateSetAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
		},
		Spec: &pbtset.TemplateSetSpec{
			Name:        req.Name,
			Memo:        req.Memo,
			TemplateIds: req.TemplateIds,
			Public:      req.Public,
			BoundApps:   req.BoundApps,
		},
	}
	rp, err := s.client.DS.CreateTemplateSet(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create template set failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateTemplateSetResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteTemplateSet delete a template set
func (s *Service) DeleteTemplateSet(ctx context.Context, req *pbcs.DeleteTemplateSetReq) (*pbcs.DeleteTemplateSetResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.DeleteTemplateSetReq{
		Id: req.TemplateSetId,
		Attachment: &pbtset.TemplateSetAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
		},
		Force: req.Force,
	}
	if _, err := s.client.DS.DeleteTemplateSet(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete template set failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.DeleteTemplateSetResp{}, nil
}

// UpdateTemplateSet update a template set
func (s *Service) UpdateTemplateSet(ctx context.Context, req *pbcs.UpdateTemplateSetReq) (*pbcs.UpdateTemplateSetResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.UpdateTemplateSetReq{
		Id: req.TemplateSetId,
		Attachment: &pbtset.TemplateSetAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
		},
		Spec: &pbtset.TemplateSetSpec{
			Name:        req.Name,
			Memo:        req.Memo,
			TemplateIds: req.TemplateIds,
			Public:      req.Public,
			BoundApps:   req.BoundApps,
		},
		Force: req.Force,
	}
	if _, err := s.client.DS.UpdateTemplateSet(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update template set failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UpdateTemplateSetResp{}, nil
}

// ListTemplateSets list template sets
func (s *Service) ListTemplateSets(ctx context.Context, req *pbcs.ListTemplateSetsReq) (*pbcs.ListTemplateSetsResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateSetsReq{
		BizId:           grpcKit.BizID,
		TemplateSpaceId: req.TemplateSpaceId,
		SearchFields:    req.SearchFields,
		SearchValue:     req.SearchValue,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTemplateSets(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template sets failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTemplateSetsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListAppTemplateSets list app template sets
func (s *Service) ListAppTemplateSets(ctx context.Context, req *pbcs.ListAppTemplateSetsReq) (
	*pbcs.ListAppTemplateSetsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListAppTemplateSetsReq{
		BizId: grpcKit.BizID,
		AppId: req.AppId,
	}

	rp, err := s.client.DS.ListAppTemplateSets(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list app template sets failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListAppTemplateSetsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateSetsByIDs list template sets by ids
func (s *Service) ListTemplateSetsByIDs(ctx context.Context, req *pbcs.ListTemplateSetsByIDsReq) (
	*pbcs.ListTemplateSetsByIDsResp, error) {
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

	r := &pbds.ListTemplateSetsByIDsReq{
		Ids: req.Ids,
	}

	rp, err := s.client.DS.ListTemplateSetsByIDs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template sets by ids failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTemplateSetsByIDsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplSetsOfBiz list template sets of one biz
func (s *Service) ListTmplSetsOfBiz(ctx context.Context, req *pbcs.ListTmplSetsOfBizReq) (
	*pbcs.ListTmplSetsOfBizResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplSetsOfBizReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}

	rp, err := s.client.DS.ListTmplSetsOfBiz(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template sets of biz failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplSetsOfBizResp{
		Details: rp.Details,
	}
	return resp, nil
}
