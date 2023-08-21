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
	"fmt"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbts "bscp.io/pkg/protocol/core/template-space"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/search"
	"bscp.io/pkg/types"
)

// CreateTemplateSpace create template space.
func (s *Service) CreateTemplateSpace(ctx context.Context, req *pbds.CreateTemplateSpaceReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.TemplateSpace().GetByUniqueKey(kt, req.Attachment.BizId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("template space's same name %s already exists", req.Spec.Name)
	}

	templateSpace := &table.TemplateSpace{
		Spec:       req.Spec.TemplateSpaceSpec(),
		Attachment: req.Attachment.TemplateSpaceAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	id, err := s.dao.TemplateSpace().Create(kt, templateSpace)
	if err != nil {
		logs.Errorf("create template space failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListTemplateSpaces list template space.
func (s *Service) ListTemplateSpaces(ctx context.Context, req *pbds.ListTemplateSpacesReq) (*pbds.ListTemplateSpacesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.TemplateSpace)
	if err != nil {
		return nil, err
	}

	details, count, err := s.dao.TemplateSpace().List(kt, req.BizId, searcher, opt)

	if err != nil {
		logs.Errorf("list template spaces failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplateSpacesResp{
		Count:   uint32(count),
		Details: pbts.PbTemplateSpaces(details),
	}
	return resp, nil
}

// UpdateTemplateSpace update template space.
func (s *Service) UpdateTemplateSpace(ctx context.Context, req *pbds.UpdateTemplateSpaceReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	templateSpace := &table.TemplateSpace{
		ID:         req.Id,
		Spec:       req.Spec.TemplateSpaceSpec(),
		Attachment: req.Attachment.TemplateSpaceAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if err := s.dao.TemplateSpace().Update(kt, templateSpace); err != nil {
		logs.Errorf("update template space failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteTemplateSpace delete template space.
func (s *Service) DeleteTemplateSpace(ctx context.Context, req *pbds.DeleteTemplateSpaceReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateSpaceNoSubResource(kt, req.Id); err != nil {
		return nil, err
	}

	templateSpace := &table.TemplateSpace{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateSpaceAttachment(),
	}
	if err := s.dao.TemplateSpace().Delete(kt, templateSpace); err != nil {
		logs.Errorf("delete template space failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// GetAllBizsOfTemplateSpaces get all biz ids of template spaces
func (s *Service) GetAllBizsOfTemplateSpaces(ctx context.Context, req *pbbase.EmptyReq) (
	*pbds.GetAllBizsOfTemplateSpacesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	bizIDs, err := s.dao.TemplateSpace().GetAllBizs(kt)
	if err != nil {
		logs.Errorf("get all bizs of template space failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.GetAllBizsOfTemplateSpacesResp{BizIds: bizIDs}
	return resp, nil
}

// CreateDefaultTemplateSpace get all biz ids of template spaces
func (s *Service) CreateDefaultTemplateSpace(ctx context.Context, req *pbds.CreateDefaultTemplateSpaceReq) (
	*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	id, err := s.dao.TemplateSpace().CreateDefault(kt, req.BizId)
	if err != nil {
		logs.Errorf("create default template space failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListTemplateSpacesByIDs list template space by ids.
func (s *Service) ListTemplateSpacesByIDs(ctx context.Context, req *pbds.ListTemplateSpacesByIDsReq) (*pbds.
	ListTemplateSpacesByIDsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	details, err := s.dao.TemplateSpace().ListByIDs(kt, req.Ids)
	if err != nil {
		logs.Errorf("list template spaces failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplateSpacesByIDsResp{
		Details: pbts.PbTemplateSpaces(details),
	}
	return resp, nil
}
