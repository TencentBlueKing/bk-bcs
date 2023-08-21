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
	"errors"
	"fmt"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbtset "bscp.io/pkg/protocol/core/template-set"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/search"
	"bscp.io/pkg/types"
)

// CreateTemplateSet create template set.
func (s *Service) CreateTemplateSet(ctx context.Context, req *pbds.CreateTemplateSetReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.TemplateSet().GetByUniqueKey(
		kt, req.Attachment.BizId, req.Attachment.TemplateSpaceId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("template set's same name %s already exists", req.Spec.Name)
	}

	if req.Spec.Public == true {
		req.Spec.BoundApps = []uint32{}
	}

	templateSet := &table.TemplateSet{
		Spec:       req.Spec.TemplateSetSpec(),
		Attachment: req.Attachment.TemplateSetAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	id, err := s.dao.TemplateSet().Create(kt, templateSet)
	if err != nil {
		logs.Errorf("create template set failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListTemplateSets list template set.
func (s *Service) ListTemplateSets(ctx context.Context, req *pbds.ListTemplateSetsReq) (*pbds.ListTemplateSetsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.TemplateSet)
	if err != nil {
		return nil, err
	}

	details, count, err := s.dao.TemplateSet().List(kt, req.BizId, req.TemplateSpaceId, searcher, opt)

	if err != nil {
		logs.Errorf("list template sets failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplateSetsResp{
		Count:   uint32(count),
		Details: pbtset.PbTemplateSets(details),
	}
	return resp, nil
}

// UpdateTemplateSet update template set.
func (s *Service) UpdateTemplateSet(ctx context.Context, req *pbds.UpdateTemplateSetReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	var (
		hasInvisibleApp bool
		invisibleApps   []uint32
		err             error
	)

	if req.Spec.Public == false {
		invisibleApps, err = s.dao.TemplateBindingRelation().ListTemplateSetInvisibleApps(kt, req.Attachment.BizId,
			req.Id, req.Spec.BoundApps)
		if err != nil {
			logs.Errorf("update template set failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(invisibleApps) > 0 {
			hasInvisibleApp = true
			if !req.Force {
				return nil, errors.New("template set is bound to unnamed app, please unbind first")
			}
		}
	}

	if len(req.Spec.TemplateIds) > 0 {
		if err := s.dao.Validator().ValidateTemplatesExist(kt, req.Spec.TemplateIds); err != nil {
			return nil, err
		}
	}

	if _, err := s.dao.TemplateSet().GetByUniqueKey(
		kt, req.Attachment.BizId, req.Attachment.TemplateSpaceId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("template set's same name %s already exists", req.Spec.Name)
	}

	templateSet := &table.TemplateSet{
		ID:         req.Id,
		Spec:       req.Spec.TemplateSetSpec(),
		Attachment: req.Attachment.TemplateSetAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if req.Spec.Public == true {
		templateSet.Spec.BoundApps = []uint32{}
	}

	tx := s.dao.GenQuery().Begin()

	// 1. update template set
	if err = s.dao.TemplateSet().UpdateWithTx(kt, tx, templateSet); err != nil {
		logs.Errorf("update template set failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 2. delete template set for invisible apps if exists
	if hasInvisibleApp {
		if err = s.dao.TemplateBindingRelation().DeleteTmplSetForInvisibleAppsWithTx(kt, tx, req.Attachment.BizId,
			req.Id, invisibleApps); err != nil {
			logs.Errorf("delete template set for invisible apps failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	return new(pbbase.EmptyResp), nil
}

// DeleteTemplateSet delete template set.
func (s *Service) DeleteTemplateSet(ctx context.Context, req *pbds.DeleteTemplateSetReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	r := &pbds.ListTemplateSetBoundCountsReq{
		BizId:           req.Attachment.BizId,
		TemplateSpaceId: req.Attachment.TemplateSpaceId,
		TemplateSetIds:  []uint32{req.Id},
	}
	boundCnt, err := s.ListTemplateSetBoundCounts(ctx, r)
	if err != nil {
		logs.Errorf("delete template set failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var hasUnnamedApp bool
	if len(boundCnt.Details) > 0 {
		if boundCnt.Details[0].BoundUnnamedAppCount > 0 {
			hasUnnamedApp = true
			if !req.Force {
				return nil, errors.New("template set is bound to unnamed app, please unbind first")
			}
		}
	}

	tx := s.dao.GenQuery().Begin()

	// 1. delete template set
	templateSet := &table.TemplateSet{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateSetAttachment(),
	}
	if err = s.dao.TemplateSet().DeleteWithTx(kt, tx, templateSet); err != nil {
		logs.Errorf("delete template set failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 2. delete bound unnamed app if exists
	if hasUnnamedApp {
		if err = s.dao.TemplateBindingRelation().DeleteTmplSetWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
			logs.Errorf("delete template set failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	return new(pbbase.EmptyResp), nil
}

// ListAppTemplateSets list app template set.
func (s *Service) ListAppTemplateSets(ctx context.Context, req *pbds.ListAppTemplateSetsReq) (*pbds.
	ListAppTemplateSetsResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	details, err := s.dao.TemplateSet().ListAppTmplSets(kt, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("list template sets failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListAppTemplateSetsResp{
		Details: pbtset.PbTemplateSets(details),
	}
	return resp, nil
}

// ListTemplateSetsByIDs list template set by ids.
func (s *Service) ListTemplateSetsByIDs(ctx context.Context, req *pbds.ListTemplateSetsByIDsReq) (*pbds.
	ListTemplateSetsByIDsResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	details, err := s.dao.TemplateSet().ListByIDs(kt, req.Ids)
	if err != nil {
		logs.Errorf("list template sets failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplateSetsByIDsResp{
		Details: pbtset.PbTemplateSets(details),
	}
	return resp, nil
}
