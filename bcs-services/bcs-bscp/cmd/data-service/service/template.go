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
	"strings"
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbtemplate "bscp.io/pkg/protocol/core/template"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/search"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// CreateTemplate create template.
func (s *Service) CreateTemplate(ctx context.Context, req *pbds.CreateTemplateReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.Template().GetByUniqueKey(kt, req.Attachment.BizId, req.Attachment.TemplateSpaceId,
		req.Spec.Name, req.Spec.Path); err == nil {
		return nil, fmt.Errorf("config item's same name %s and path %s already exists", req.Spec.Name, req.Spec.Path)
	}
	if len(req.TemplateSetIds) > 0 {
		if err := s.dao.Validator().ValidateTemplateSetsExist(kt, req.TemplateSetIds); err != nil {
			return nil, err
		}
	}

	tx := s.dao.GenQuery().Begin()

	// 1. create template
	template := &table.Template{
		Spec:       req.Spec.TemplateSpec(),
		Attachment: req.Attachment.TemplateAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	id, err := s.dao.Template().CreateWithTx(kt, tx, template)
	if err != nil {
		logs.Errorf("create template failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 2. create template revision
	spec := req.TrSpec.TemplateRevisionSpec()
	spec.RevisionName = generateRevisionName()
	templateRevision := &table.TemplateRevision{
		Spec: spec,
		Attachment: &table.TemplateRevisionAttachment{
			BizID:           template.Attachment.BizID,
			TemplateSpaceID: template.Attachment.TemplateSpaceID,
			TemplateID:      id,
		},
		Revision: &table.CreatedRevision{
			Creator: kt.User,
		},
	}
	if _, err = s.dao.TemplateRevision().CreateWithTx(kt, tx, templateRevision); err != nil {
		logs.Errorf("create template revision failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 3. add current template to template sets if necessary
	if len(req.TemplateSetIds) > 0 {
		if err = s.dao.TemplateSet().AddTemplateToTemplateSetsWithTx(kt, tx, id, req.TemplateSetIds); err != nil {
			logs.Errorf("add current template to template sets failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListTemplates list templates.
func (s *Service) ListTemplates(ctx context.Context, req *pbds.ListTemplatesReq) (*pbds.ListTemplatesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.Template)
	if err != nil {
		return nil, err
	}

	details, count, err := s.dao.Template().List(kt, req.BizId, req.TemplateSpaceId, searcher, opt)

	if err != nil {
		logs.Errorf("list templates failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplatesResp{
		Count:   uint32(count),
		Details: pbtemplate.PbTemplates(details),
	}
	return resp, nil
}

// UpdateTemplate update template.
func (s *Service) UpdateTemplate(ctx context.Context, req *pbds.UpdateTemplateReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	now := time.Now()
	template := &table.Template{
		ID:         req.Id,
		Spec:       req.Spec.TemplateSpec(),
		Attachment: req.Attachment.TemplateAttachment(),
		Revision: &table.Revision{
			Reviser:   kt.User,
			UpdatedAt: now,
		},
	}
	if err := s.dao.Template().Update(kt, template); err != nil {
		logs.Errorf("update template failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteTemplate delete template.
func (s *Service) DeleteTemplate(ctx context.Context, req *pbds.DeleteTemplateReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	r := &pbds.ListTemplateBoundCountsReq{
		BizId:           req.Attachment.BizId,
		TemplateSpaceId: req.Attachment.TemplateSpaceId,
		TemplateIds:     []uint32{req.Id},
	}
	boundCnt, err := s.ListTemplateBoundCounts(ctx, r)
	if err != nil {
		logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var hasTmplSet, hasUnnamedApp bool
	if len(boundCnt.Details) > 0 {
		if boundCnt.Details[0].BoundTemplateSetCount > 0 && boundCnt.Details[0].BoundUnnamedAppCount > 0 {
			hasTmplSet, hasUnnamedApp = true, true
			if !req.Force {
				return nil, errors.New("template is bound to template set and unnamed app, please unbind first")
			}
		} else if boundCnt.Details[0].BoundTemplateSetCount > 0 {
			hasTmplSet = true
			if !req.Force {
				return nil, errors.New("template is bound to template set, please unbind first")
			}
		} else if boundCnt.Details[0].BoundUnnamedAppCount > 0 {
			hasUnnamedApp = true
			if !req.Force {
				return nil, errors.New("template is bound to unnamed app, please unbind first")
			}
		}
	}

	tx := s.dao.GenQuery().Begin()

	// 1. delete template
	template := &table.Template{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateAttachment(),
	}
	if err = s.dao.Template().DeleteWithTx(kt, tx, template); err != nil {
		logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 2. delete template revisions of current template
	if err = s.dao.TemplateRevision().DeleteForTmplWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
		logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 3. delete bound template set if exists
	if hasTmplSet {
		if err = s.dao.TemplateSet().DeleteTmplFromAllTmplSetsWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
			logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, err
		}
	}

	// 4. delete bound unnamed app if exists
	if hasUnnamedApp {
		if err = s.dao.TemplateBindingRelation().DeleteTmplWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
			logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	return new(pbbase.EmptyResp), nil
}

// BatchDeleteTemplate delete template in batch.
func (s *Service) BatchDeleteTemplate(ctx context.Context, req *pbds.BatchDeleteTemplateReq) (*pbbase.EmptyResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	r := &pbds.ListTemplateBoundCountsReq{
		BizId:           req.Attachment.BizId,
		TemplateSpaceId: req.Attachment.TemplateSpaceId,
		TemplateIds:     req.Ids,
	}
	boundCnt, err := s.ListTemplateBoundCounts(ctx, r)
	if err != nil {
		logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	hasTmplSets, hasUnnamedApps := make(map[uint32]bool), make(map[uint32]bool)
	for _, detail := range boundCnt.Details {
		if detail.BoundTemplateSetCount > 0 && detail.BoundUnnamedAppCount > 0 {
			hasTmplSets[detail.TemplateId] = true
			hasUnnamedApps[detail.TemplateId] = true
			if !req.Force {
				return nil, fmt.Errorf("template id %d is bound to template set and unnamed app, please unbind first",
					detail.TemplateId)
			}
		} else if detail.BoundTemplateSetCount > 0 {
			hasTmplSets[detail.TemplateId] = true
			if !req.Force {
				return nil, fmt.Errorf("template id %d is bound to template set, please unbind first",
					detail.TemplateId)
			}
		} else if detail.BoundUnnamedAppCount > 0 {
			hasUnnamedApps[detail.TemplateId] = true
			if !req.Force {
				return nil, fmt.Errorf("template id %d is bound to unnamed app, please unbind first", detail.TemplateId)
			}
		}
	}

	tx := s.dao.GenQuery().Begin()

	// NOTE: if consider to optimize it with batch interface, consider how to add audit record as the same time
	for _, templateID := range req.Ids {
		// 1. delete template
		template := &table.Template{
			ID:         templateID,
			Attachment: req.Attachment.TemplateAttachment(),
		}
		if err = s.dao.Template().DeleteWithTx(kt, tx, template); err != nil {
			logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, err
		}

		// 2. delete template revisions of current template
		if err = s.dao.TemplateRevision().DeleteForTmplWithTx(kt, tx, req.Attachment.BizId, templateID); err != nil {
			logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, err
		}

		// 3. delete bound template set if exists
		if hasTmplSets[templateID] {
			if err = s.dao.TemplateSet().DeleteTmplFromAllTmplSetsWithTx(kt, tx, req.Attachment.BizId, templateID); err != nil {
				logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
				tx.Rollback()
				return nil, err
			}
		}

		// 4. delete bound unnamed app if exists
		if hasUnnamedApps[templateID] {
			if err = s.dao.TemplateBindingRelation().DeleteTmplWithTx(kt, tx, req.Attachment.BizId, templateID); err != nil {
				logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
				tx.Rollback()
				return nil, err
			}
		}
	}

	tx.Commit()

	return new(pbbase.EmptyResp), nil
}

// AddTemplatesToTemplateSets add templates to template sets.
func (s *Service) AddTemplatesToTemplateSets(ctx context.Context, req *pbds.AddTemplatesToTemplateSetsReq) (
	*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplatesExist(kt, req.TemplateIds); err != nil {
		return nil, err
	}
	if err := s.dao.Validator().ValidateTemplateSetsExist(kt, req.TemplateSetIds); err != nil {
		return nil, err
	}

	if err := s.dao.TemplateSet().AddTemplatesToTemplateSets(kt, req.TemplateIds, req.TemplateSetIds); err != nil {
		logs.Errorf(" add template to template sets failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteTemplatesFromTemplateSets delete templates from template sets.
func (s *Service) DeleteTemplatesFromTemplateSets(ctx context.Context, req *pbds.DeleteTemplatesFromTemplateSetsReq) (
	*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplatesExist(kt, req.TemplateIds); err != nil {
		return nil, err
	}
	if err := s.dao.Validator().ValidateTemplateSetsExist(kt, req.TemplateSetIds); err != nil {
		return nil, err
	}

	if err := s.dao.TemplateSet().DeleteTemplatesFromTemplateSets(kt, req.TemplateIds, req.TemplateSetIds); err != nil {
		logs.Errorf(" delete template from template sets failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// ListTemplatesByIDs list templates by ids.
func (s *Service) ListTemplatesByIDs(ctx context.Context, req *pbds.ListTemplatesByIDsReq) (
	*pbds.ListTemplatesByIDsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	details, err := s.dao.Template().ListByIDs(kt, req.Ids)
	if err != nil {
		logs.Errorf("list template failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplatesByIDsResp{
		Details: pbtemplate.PbTemplates(details),
	}
	return resp, nil
}

// ListTemplatesNotBound list templates not bound.
// 先获取所有模版ID列表，再获取该空间下所有套餐的template_ids字段进行合并，做差集得到目标ID列表，根据这批ID获取对应的详情，做逻辑分页和搜索
func (s *Service) ListTemplatesNotBound(ctx context.Context, req *pbds.ListTemplatesNotBoundReq) (
	*pbds.ListTemplatesNotBoundResp, error) {
	kt := kit.FromGrpcContext(ctx)

	idsAll, err := s.dao.Template().ListAllIDs(kt, req.BizId, req.TemplateSpaceId)
	if err != nil {
		logs.Errorf("list templates not bound failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	idsBound, err := s.dao.TemplateSet().ListAllTemplateIDs(kt, req.BizId, req.TemplateSpaceId)
	if err != nil {
		logs.Errorf("list templates not bound failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	ids := tools.SliceDiff(idsAll, idsBound)
	templates, err := s.dao.Template().ListByIDs(kt, ids)
	details := pbtemplate.PbTemplates(templates)

	// search by logic
	if req.SearchValue != "" {
		searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.Template)
		if err != nil {
			return nil, err
		}
		fields := searcher.SearchFields()
		fieldsMap := make(map[string]bool)
		for _, f := range fields {
			fieldsMap[f] = true
		}
		newDetails := make([]*pbtemplate.Template, 0)
		for _, detail := range details {
			if (fieldsMap["name"] && strings.Contains(detail.Spec.Name, req.SearchValue)) ||
				(fieldsMap["path"] && strings.Contains(detail.Spec.Path, req.SearchValue)) ||
				(fieldsMap["memo"] && strings.Contains(detail.Spec.Memo, req.SearchValue)) ||
				(fieldsMap["creator"] && strings.Contains(detail.Revision.Creator, req.SearchValue)) ||
				(fieldsMap["reviser"] && strings.Contains(detail.Revision.Reviser, req.SearchValue)) {
				newDetails = append(newDetails, detail)
			}
		}
		details = newDetails
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTemplatesNotBoundResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	return &pbds.ListTemplatesNotBoundResp{
		Count:   totalCnt,
		Details: details,
	}, nil
}

// ListTemplatesOfTemplateSet list templates of template set.
// 获取到该套餐的template_ids字段，根据这批ID获取对应的详情，做逻辑分页和搜索
func (s *Service) ListTemplatesOfTemplateSet(ctx context.Context, req *pbds.ListTemplatesOfTemplateSetReq) (
	*pbds.ListTemplatesOfTemplateSetResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateSetExist(kt, req.TemplateSetId); err != nil {
		return nil, err
	}

	templateSets, err := s.dao.TemplateSet().ListByIDs(kt, []uint32{req.TemplateSetId})
	if err != nil {
		logs.Errorf("list templates of template set failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	templates, err := s.dao.Template().ListByIDs(kt, templateSets[0].Spec.TemplateIDs)
	if err != nil {
		logs.Errorf("list templates of template set failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	details := pbtemplate.PbTemplates(templates)

	// search by logic
	if req.SearchValue != "" {
		searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.Template)
		if err != nil {
			return nil, err
		}
		fields := searcher.SearchFields()
		fieldsMap := make(map[string]bool)
		for _, f := range fields {
			fieldsMap[f] = true
		}
		newDetails := make([]*pbtemplate.Template, 0)
		for _, detail := range details {
			if (fieldsMap["name"] && strings.Contains(detail.Spec.Name, req.SearchValue)) ||
				(fieldsMap["path"] && strings.Contains(detail.Spec.Path, req.SearchValue)) ||
				(fieldsMap["memo"] && strings.Contains(detail.Spec.Memo, req.SearchValue)) ||
				(fieldsMap["creator"] && strings.Contains(detail.Revision.Creator, req.SearchValue)) ||
				(fieldsMap["reviser"] && strings.Contains(detail.Revision.Reviser, req.SearchValue)) {
				newDetails = append(newDetails, detail)
			}
		}
		details = newDetails
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTemplatesOfTemplateSetResp{
			Count:   totalCnt,
			Details: details,
		}, nil
	}

	// page by logic
	if req.Start >= uint32(len(details)) {
		details = details[:0]
	} else if req.Start+req.Limit > uint32(len(details)) {
		details = details[req.Start:]
	} else {
		details = details[req.Start : req.Start+req.Limit]
	}

	return &pbds.ListTemplatesOfTemplateSetResp{
		Count:   totalCnt,
		Details: details,
	}, nil
}
