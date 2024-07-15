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
	"errors"
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbtset "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-set"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateTemplateSet create template set.
func (s *Service) CreateTemplateSet(ctx context.Context, req *pbds.CreateTemplateSetReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.TemplateSet().GetByUniqueKey(
		kt, req.Attachment.BizId, req.Attachment.TemplateSpaceId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("template set's same name %s already exists", req.Spec.Name)
	}

	if req.Spec.Public {
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
func (s *Service) ListTemplateSets(ctx context.Context, req *pbds.ListTemplateSetsReq) (*pbds.ListTemplateSetsResp,
	error) {
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
// nolint: funlen
func (s *Service) UpdateTemplateSet(ctx context.Context, req *pbds.UpdateTemplateSetReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)
	// set for empty slice to ensure the data in db is not `null` but `[]`
	if len(req.Spec.TemplateIds) == 0 {
		req.Spec.TemplateIds = []uint32{}
	}
	if len(req.Spec.BoundApps) == 0 {
		req.Spec.BoundApps = []uint32{}
	}

	var invisibleATBs []*table.AppTemplateBinding
	var err error
	if !req.Spec.Public {
		invisibleATBs, err = s.dao.TemplateBindingRelation().ListTemplateSetInvisibleATBs(kt, req.Attachment.BizId,
			req.Id, req.Spec.BoundApps)
		if err != nil {
			logs.Errorf("update template set failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(invisibleATBs) > 0 {
			if !req.Force {
				return nil, errors.New("template set is bound to unnamed app, please unbind first")
			}
		}
	}

	if len(req.Spec.TemplateIds) > 0 {
		if e := s.dao.Validator().ValidateTemplatesExist(kt, req.Spec.TemplateIds); e != nil {
			return nil, e
		}
	}

	if _, e := s.dao.TemplateSet().GetByUniqueKeyForUpdate(
		kt, req.Attachment.BizId, req.Attachment.TemplateSpaceId, req.Id, req.Spec.Name); e == nil {
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
	if req.Spec.Public {
		templateSet.Spec.BoundApps = []uint32{}
	}

	tx := s.dao.GenQuery().Begin()

	// 1. update template set
	if err = s.dao.TemplateSet().UpdateWithTx(kt, tx, templateSet); err != nil {
		logs.Errorf("update template set failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// validate template set's templates count.
	if err = s.dao.TemplateSet().ValidateTmplNumber(kt, tx, req.Attachment.BizId, req.Id); err != nil {
		logs.Errorf("validate template set's templates count failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 2. update app template bindings if necessary
	// delete invisible template set from correspond app template bindings
	if len(invisibleATBs) > 0 {
		for _, atb := range invisibleATBs {
			// delete the specific set in the atb
			delIndex := -1
			for idx, b := range atb.Spec.Bindings {
				if b.TemplateSetID == req.Id {
					delIndex = idx
					break
				}
			}
			if delIndex >= 0 {
				atb.Spec.Bindings = append(atb.Spec.Bindings[:delIndex], atb.Spec.Bindings[delIndex+1:]...)
			}
			if err = s.CascadeUpdateATB(kt, tx, atb); err != nil {
				logs.Errorf("cascade update app template binding failed, err: %v, rid: %s", err, kt.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
				}
				return nil, err
			}
		}
	}

	// 3. update app template bindings if necessary
	// get old template set detail before the update operation, not use transaction so that it is the old one
	var oldTmplSets []*table.TemplateSet
	oldTmplSets, err = s.dao.TemplateSet().ListByIDs(kt, []uint32{req.Id})
	if err != nil {
		logs.Errorf("list template sets by ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if !tools.IsSameSlice(oldTmplSets[0].Spec.TemplateIDs, templateSet.Spec.TemplateIDs) {
		var atbs []*table.AppTemplateBinding
		atbs, err = s.dao.TemplateBindingRelation().
			ListTemplateSetsBoundATBs(kt, req.Attachment.BizId, []uint32{req.Id})
		if err != nil {
			logs.Errorf("list template set bound app template bindings failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(atbs) > 0 {
			for _, atb := range atbs {
				if err = s.CascadeUpdateATB(kt, tx, atb); err != nil {
					logs.Errorf("cascade update app template binding failed, err: %v, rid: %s", err, kt.Rid)
					if rErr := tx.Rollback(); rErr != nil {
						logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
					}
					return nil, err
				}
			}
		}
	}

	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteTemplateSet delete template set.
func (s *Service) DeleteTemplateSet(ctx context.Context, req *pbds.DeleteTemplateSetReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	r := &pbds.ListTmplSetBoundCountsReq{
		BizId:           req.Attachment.BizId,
		TemplateSpaceId: req.Attachment.TemplateSpaceId,
		TemplateSetIds:  []uint32{req.Id},
	}

	var boundCnt *pbds.ListTmplSetBoundCountsResp
	var err error
	boundCnt, err = s.ListTmplSetBoundCounts(ctx, r)
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
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 2. update app template bindings if necessary
	if hasUnnamedApp {
		var atbs []*table.AppTemplateBinding
		atbs, err = s.dao.TemplateBindingRelation().
			ListTemplateSetsBoundATBs(kt, req.Attachment.BizId, []uint32{req.Id})
		if err != nil {
			logs.Errorf("list template set bound app template bindings failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		if len(atbs) > 0 {
			for _, atb := range atbs {
				// delete the specific set in the atb
				delIndex := -1
				for idx, b := range atb.Spec.Bindings {
					if b.TemplateSetID == req.Id {
						delIndex = idx
						break
					}
				}
				if delIndex >= 0 {
					atb.Spec.Bindings = append(atb.Spec.Bindings[:delIndex], atb.Spec.Bindings[delIndex+1:]...)
				}
				if err := s.CascadeUpdateATB(kt, tx, atb); err != nil {
					logs.Errorf("cascade update app template binding failed, err: %v, rid: %s", err, kt.Rid)
					if rErr := tx.Rollback(); rErr != nil {
						logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
					}
					return nil, err
				}
			}
		}
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}
	return new(pbbase.EmptyResp), nil
}

// ListAppTemplateSets list app template set.
func (s *Service) ListAppTemplateSets(ctx context.Context, req *pbds.ListAppTemplateSetsReq) (
	*pbds.ListAppTemplateSetsResp, error) {
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
func (s *Service) ListTemplateSetsByIDs(ctx context.Context, req *pbds.ListTemplateSetsByIDsReq) (
	*pbds.ListTemplateSetsByIDsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTmplSetsExist(kt, req.Ids); err != nil {
		return nil, err
	}

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

// ListTemplateSetBriefInfoByIDs list template set by ids.
func (s *Service) ListTemplateSetBriefInfoByIDs(ctx context.Context, req *pbds.ListTemplateSetBriefInfoByIDsReq) (
	*pbds.ListTemplateSetBriefInfoByIDsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTmplSetsExist(kt, req.Ids); err != nil {
		return nil, err
	}

	// template set details
	tmplSets, err := s.dao.TemplateSet().ListByIDs(kt, req.Ids)
	if err != nil {
		logs.Errorf("list template sets failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplSetMap := make(map[uint32]*table.TemplateSet)
	tmplSpaceIDs := make([]uint32, 0)
	for _, ts := range tmplSets {
		tmplSetMap[ts.ID] = ts
		tmplSpaceIDs = append(tmplSpaceIDs, ts.Attachment.TemplateSpaceID)
	}
	tmplSpaceIDs = tools.RemoveDuplicates(tmplSpaceIDs)

	// template space details
	tmplSpaces, err := s.dao.TemplateSpace().ListByIDs(kt, tmplSpaceIDs)
	if err != nil {
		logs.Errorf("list template spaces failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplSpaceMap := make(map[uint32]*table.TemplateSpace)
	for _, ts := range tmplSpaces {
		tmplSpaceMap[ts.ID] = ts
	}

	details := make([]*pbtset.TemplateSetBriefInfo, len(tmplSets))
	for idx, t := range tmplSets {
		details[idx] = &pbtset.TemplateSetBriefInfo{
			TemplateSpaceId:   t.Attachment.TemplateSpaceID,
			TemplateSpaceName: tmplSpaceMap[t.Attachment.TemplateSpaceID].Spec.Name,
			TemplateSetId:     t.ID,
			TemplateSetName:   tmplSetMap[t.ID].Spec.Name,
		}
	}
	return &pbds.ListTemplateSetBriefInfoByIDsResp{Details: details}, nil
}

// ListTmplSetsOfBiz list template sets of one biz.
func (s *Service) ListTmplSetsOfBiz(ctx context.Context, req *pbds.ListTmplSetsOfBizReq) (
	*pbds.ListTmplSetsOfBizResp, error) {
	kt := kit.FromGrpcContext(ctx)

	tmplSets, err := s.dao.TemplateSet().ListAllTmplSetsOfBiz(kt, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("list template sets of biz failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(tmplSets) == 0 {
		return &pbds.ListTmplSetsOfBizResp{}, nil
	}

	// get the map of template space id => template set detail
	tmplSetsMap := make(map[uint32]*pbtset.TemplateSetOfBizDetail)
	for _, t := range tmplSets {
		if _, ok := tmplSetsMap[t.Attachment.TemplateSpaceID]; !ok {
			tmplSetsMap[t.Attachment.TemplateSpaceID] = &pbtset.TemplateSetOfBizDetail{}
		}
		tmplSetsMap[t.Attachment.TemplateSpaceID].TemplateSets = append(
			tmplSetsMap[t.Attachment.TemplateSpaceID].TemplateSets,
			&pbtset.TemplateSetOfBizDetail_TemplateSetOfBiz{
				TemplateSetId:   t.ID,
				TemplateSetName: t.Spec.Name,
				TemplateIds:     t.Spec.TemplateIDs,
			})
	}
	tmplSpaceIDs := make([]uint32, 0, len(tmplSetsMap))
	for tmplSpaceID := range tmplSetsMap {
		tmplSpaceIDs = append(tmplSpaceIDs, tmplSpaceID)
	}

	tmplSpaces, err := s.dao.TemplateSpace().ListByIDs(kt, tmplSpaceIDs)
	if err != nil {
		logs.Errorf("list template sets of biz failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]*pbtset.TemplateSetOfBizDetail, 0)
	for _, t := range tmplSpaces {
		details = append(details, &pbtset.TemplateSetOfBizDetail{
			TemplateSpaceId:   t.ID,
			TemplateSpaceName: t.Spec.Name,
			TemplateSets:      tmplSetsMap[t.ID].TemplateSets,
		})
	}

	resp := &pbds.ListTmplSetsOfBizResp{
		Details: details,
	}
	return resp, nil
}
