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
	"path"
	"sort"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbtemplate "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template"
	pbtr "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-revision"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateTemplate create template.
//
//nolint:funlen
func (s *Service) CreateTemplate(ctx context.Context, req *pbds.CreateTemplateReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// Get all configuration files under a certain package of the service
	items, _, err := s.dao.Template().List(kt, req.Attachment.BizId, req.Attachment.TemplateSpaceId,
		nil, &types.BasePage{All: true}, nil, "")
	if err != nil {
		return nil, err
	}
	existingPaths := []string{}
	for _, v := range items {
		existingPaths = append(existingPaths, path.Join(v.Spec.Path, v.Spec.Name))
	}

	if tools.CheckPathConflict(path.Join(req.Spec.Path, req.Spec.Name), existingPaths) {
		return nil, errf.Errorf(errf.InvalidRequest, i18n.T(kt, "config item's same name %s and path %s already exists",
			req.Spec.Name, req.Spec.Path))
	}

	if len(req.TemplateSetIds) > 0 {
		// ValidateTmplSetsExist validate if template sets exists
		if err = s.dao.Validator().ValidateTmplSetsExist(kt, req.TemplateSetIds); err != nil {
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
	// CreateWithTx create one template instance with transaction.
	id, err := s.dao.Template().CreateWithTx(kt, tx, template)
	if err != nil {
		logs.Errorf("create template failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// validate template set's templates count.
	for _, tmplSetID := range req.TemplateSetIds {
		if err = s.dao.TemplateSet().ValidateTmplNumber(kt, tx, req.Attachment.BizId, tmplSetID); err != nil {
			logs.Errorf("validate template set's templates count failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}

	// 2. create template revision
	spec := req.TrSpec.TemplateRevisionSpec()
	// if no revision name is specified, generate it by system
	if spec.RevisionName == "" {
		spec.RevisionName = tools.GenerateRevisionName()
	}
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
	// CreateWithTx create one template revision instance with transaction.
	if _, err = s.dao.TemplateRevision().CreateWithTx(kt, tx, templateRevision); err != nil {
		logs.Errorf("create template revision failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 3. add current template to template sets if necessary
	if len(req.TemplateSetIds) > 0 {
		if err = s.dao.TemplateSet().AddTmplToTmplSetsWithTx(kt, tx, id, req.TemplateSetIds); err != nil {
			logs.Errorf("add current template to template sets failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}

		// 3-1. update app template bindings if necessary
		atbs, err := s.dao.TemplateBindingRelation().
			ListTemplateSetsBoundATBs(kt, template.Attachment.BizID, req.TemplateSetIds)
		if err != nil {
			logs.Errorf("list template sets bound app template bindings failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
		if len(atbs) > 0 {
			for _, atb := range atbs {
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

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return &pbds.CreateResp{Id: id}, nil
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
	topIds, _ := tools.StrToUint32Slice(req.Ids)
	// List templates with options.
	details, count, err := s.dao.Template().List(kt, req.BizId, req.TemplateSpaceId, searcher,
		opt, topIds, req.SearchValue)

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

	template := &table.Template{
		ID:         req.Id,
		Spec:       req.Spec.TemplateSpec(),
		Attachment: req.Attachment.TemplateAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	// Update one template's info.
	if err := s.dao.Template().Update(kt, template); err != nil {
		logs.Errorf("update template failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteTemplate delete template.
//
//nolint:funlen
func (s *Service) DeleteTemplate(ctx context.Context, req *pbds.DeleteTemplateReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	r := &pbds.ListTmplBoundCountsReq{
		BizId:           req.Attachment.BizId,
		TemplateSpaceId: req.Attachment.TemplateSpaceId,
		TemplateIds:     []uint32{req.Id},
	}
	boundCnt, err := s.ListTmplBoundCounts(ctx, r)
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
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 2. delete template revisions of current template
	if err = s.dao.TemplateRevision().DeleteForTmplWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
		logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 3. delete bound template set if exists
	if hasTmplSet {
		if err = s.dao.TemplateSet().DeleteTmplFromAllTmplSetsWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
			logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}

	// 4. update app template bindings if necessary
	if hasUnnamedApp {
		atbs, err := s.dao.TemplateBindingRelation().ListTemplatesBoundATBs(kt, req.Attachment.BizId, []uint32{req.Id})
		if err != nil {
			logs.Errorf("list templates bound app template bindings failed, err: %v, rid: %s", err, kt.Rid)
			if e := tx.Rollback(); e != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", e, kt.Rid)
			}
			return nil, err
		}
		if len(atbs) > 0 {
			for _, atb := range atbs {
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
	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// BatchDeleteTemplate delete template in batch.
// nolint: funlen
func (s *Service) BatchDeleteTemplate(ctx context.Context, req *pbds.BatchDeleteTemplateReq) (*pbbase.EmptyResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	r := &pbds.ListTmplBoundCountsReq{
		BizId:           req.Attachment.BizId,
		TemplateSpaceId: req.Attachment.TemplateSpaceId,
		TemplateIds:     req.Ids,
	}
	boundCnt, err := s.ListTmplBoundCounts(ctx, r)
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
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}

		// 2. delete template revisions of current template
		if err = s.dao.TemplateRevision().DeleteForTmplWithTx(kt, tx, req.Attachment.BizId, templateID); err != nil {
			logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}

		// 3. delete bound template set if exists
		if hasTmplSets[templateID] {
			if err = s.dao.TemplateSet().DeleteTmplFromAllTmplSetsWithTx(kt, tx, req.Attachment.BizId, templateID); err != nil {
				logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
				}
				return nil, err
			}
		}

		// 4. update app template bindings if necessary
		if hasUnnamedApps[templateID] {
			atbs, e := s.dao.TemplateBindingRelation().ListTemplatesBoundATBs(kt, req.Attachment.BizId,
				[]uint32{templateID})
			if e != nil {
				logs.Errorf("list templates bound app template bindings failed, err: %v, rid: %s", e, kt.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
				}
				return nil, e
			}
			if len(atbs) > 0 {
				for _, atb := range atbs {
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
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// AddTmplsToTmplSets add templates to template sets.
func (s *Service) AddTmplsToTmplSets(ctx context.Context, req *pbds.AddTmplsToTmplSetsReq) (
	*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplatesExist(kt, req.TemplateIds); err != nil {
		return nil, err
	}
	if err := s.dao.Validator().ValidateTmplSetsExist(kt, req.TemplateSetIds); err != nil {
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()

	// 1. add templates to template sets
	if err := s.dao.TemplateSet().AddTmplsToTmplSetsWithTx(kt, tx, req.TemplateIds,
		req.TemplateSetIds); err != nil {
		logs.Errorf(" add templates to template sets failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// validate template set's templates count.
	for _, tmplSetID := range req.TemplateSetIds {
		if err := s.dao.TemplateSet().ValidateTmplNumber(kt, tx, req.BizId, tmplSetID); err != nil {
			logs.Errorf("validate template set's templates count failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}

	// 2. update app template bindings if necessary
	atbs, err := s.dao.TemplateBindingRelation().
		ListTemplateSetsBoundATBs(kt, req.BizId, req.TemplateSetIds)
	if err != nil {
		logs.Errorf("list template sets bound app template bindings failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}
	if len(atbs) > 0 {
		for _, atb := range atbs {
			if err := s.CascadeUpdateATB(kt, tx, atb); err != nil {
				logs.Errorf("cascade update app template binding failed, err: %v, rid: %s", err, kt.Rid)
				if e := tx.Rollback(); e != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", e, kt.Rid)
				}
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteTmplsFromTmplSets delete templates from template sets.
func (s *Service) DeleteTmplsFromTmplSets(ctx context.Context, req *pbds.DeleteTmplsFromTmplSetsReq) (
	*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplatesExist(kt, req.TemplateIds); err != nil {
		return nil, err
	}
	if err := s.dao.Validator().ValidateTmplSetsExist(kt, req.TemplateSetIds); err != nil {
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()

	// 1. delete templates from template sets
	if err := s.dao.TemplateSet().DeleteTmplsFromTmplSetsWithTx(kt, tx, req.TemplateIds,
		req.TemplateSetIds); err != nil {
		logs.Errorf(" delete template from template sets failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 2. update app template bindings if necessary
	atbs, err := s.dao.TemplateBindingRelation().
		ListTemplateSetsBoundATBs(kt, req.BizId, req.TemplateSetIds)
	if err != nil {
		logs.Errorf("list template sets bound app template bindings failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}
	if len(atbs) > 0 {
		for _, atb := range atbs {
			if e := s.CascadeUpdateATB(kt, tx, atb); e != nil {
				logs.Errorf("cascade update app template binding failed, err: %v, rid: %s", e, kt.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
				}
				return nil, e
			}
		}
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}

// ListTemplatesByIDs list templates by ids.
func (s *Service) ListTemplatesByIDs(ctx context.Context, req *pbds.ListTemplatesByIDsReq) (
	*pbds.ListTemplatesByIDsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplatesExist(kt, req.Ids); err != nil {
		return nil, err
	}

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

	// validate the page params
	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

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
	if err != nil {
		logs.Errorf("list app bound template revisions failed, err: %v, rid: %s", err, kt.Rid)
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
		fieldsMap["combinedPathName"] = true
		newDetails := make([]*pbtemplate.Template, 0)
		for _, detail := range details {
			combinedPathName := path.Join(detail.Spec.Path, detail.Spec.Name)
			if (fieldsMap["combinedPathName"] && strings.Contains(combinedPathName, req.SearchValue)) ||
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

// ListTmplsOfTmplSet list templates of template set.
// 获取到该套餐的template_ids字段，根据这批ID获取对应的详情，做逻辑分页和搜索
func (s *Service) ListTmplsOfTmplSet(ctx context.Context, req *pbds.ListTmplsOfTmplSetReq) (
	*pbds.ListTmplsOfTmplSetResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// validate the page params
	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	if err := s.dao.Validator().ValidateTmplSetExist(kt, req.TemplateSetId); err != nil {
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
		fieldsMap["combinedPathName"] = true
		newDetails := make([]*pbtemplate.Template, 0)
		for _, detail := range details {
			// 拼接path和name
			combinedPathName := path.Join(detail.Spec.Path, detail.Spec.Name)
			if (fieldsMap["combinedPathName"] && strings.Contains(combinedPathName, req.SearchValue)) ||
				(fieldsMap["memo"] && strings.Contains(detail.Spec.Memo, req.SearchValue)) ||
				(fieldsMap["creator"] && strings.Contains(detail.Revision.Creator, req.SearchValue)) ||
				(fieldsMap["reviser"] && strings.Contains(detail.Revision.Reviser, req.SearchValue)) {
				newDetails = append(newDetails, detail)
			}
		}
		details = newDetails
	}
	topId, _ := tools.StrToUint32Slice(req.Ids)
	sort.SliceStable(details, func(i, j int) bool {
		iInTopID := tools.Contains(topId, details[i].Id)
		jInTopID := tools.Contains(topId, details[j].Id)
		if iInTopID && jInTopID {
			return i < j
		}
		if iInTopID {
			return true
		}
		if jInTopID {
			return false
		}
		return i < j
	})

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListTmplsOfTmplSetResp{
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

	return &pbds.ListTmplsOfTmplSetResp{
		Count:   totalCnt,
		Details: details,
	}, nil
}

// ListTemplateByTuple 按照多个字段in查询
func (s *Service) ListTemplateByTuple(ctx context.Context, req *pbds.ListTemplateByTupleReq) (
	*pbds.ListTemplateByTupleReqResp, error) {
	kt := kit.FromGrpcContext(ctx)
	data := [][]interface{}{}
	for _, item := range req.Items {
		data = append(data, []interface{}{item.BizId, item.TemplateSpaceId, item.Name, item.Path})
	}
	templates, err := s.dao.Template().ListTemplateByTuple(kt, data)
	if err != nil {
		logs.Errorf("list templates by tuple failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	templateRevision := map[uint32]*table.TemplateRevision{}
	templateIds := []uint32{}
	if len(templates) > 0 {
		for _, item := range templates {
			templateIds = append(templateIds, item.ID)
		}
		revision, err := s.dao.TemplateRevision().ListLatestRevisionsGroupByTemplateIds(kt, templateIds)
		if err != nil {
			return nil, err
		}
		for _, item := range revision {
			templateRevision[item.Attachment.TemplateID] = item
		}
	}

	templatesData := []*pbds.ListTemplateByTupleReqResp_Item{}
	for _, item := range templates {
		templatesData = append(templatesData,
			&pbds.ListTemplateByTupleReqResp_Item{
				Template:         pbtemplate.PbTemplate(item),
				TemplateRevision: pbtr.PbTemplateRevision(templateRevision[item.ID]),
			})
	}

	resp := &pbds.ListTemplateByTupleReqResp{Items: templatesData}

	return resp, nil

}

// BatchUpsertTemplates batch upsert templates.
// nolint:funlen
func (s *Service) BatchUpsertTemplates(ctx context.Context, req *pbds.BatchUpsertTemplatesReq) (
	*pbds.BatchUpsertTemplatesReqResp, error) {

	kt := kit.FromGrpcContext(ctx)
	// 1. 验证套餐是否存在
	if err := s.dao.Validator().ValidateTmplSetsExist(kt, req.TemplateSetIds); err != nil {
		return nil, err
	}

	// 2. 创建模板配置
	createData, updateData := make([]*pbds.BatchUpsertTemplatesReq_Item, 0), make([]*pbds.BatchUpsertTemplatesReq_Item, 0)
	updateIds := []uint32{}
	// 筛选出新增和修改的数据
	for _, item := range req.Items {
		if item.GetTemplate().GetId() != 0 {
			updateIds = append(updateIds, item.GetTemplate().GetId())
			updateData = append(updateData, item)
		} else {
			createData = append(createData, item)
		}
	}

	now := time.Now().UTC()
	tx := s.dao.GenQuery().Begin()
	createIds, e := s.doBatchCreateTemplates(kt, tx, createData, now)
	if e != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}

	if len(updateIds) > 0 {
		oldTemplateData, err := s.validateBatchUpsertTemplates(kt, updateIds, updateData)
		if err != nil {
			return nil, err
		}
		if e := s.doBatchUpdateTemplates(kt, tx, updateData, oldTemplateData, now); e != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, e
		}
	}

	// 合并创建和编辑后的模板ID
	templateIds := tools.MergeAndDeduplicate(createIds, updateIds)

	// 3. 添加至套餐中
	if err := s.dao.TemplateSet().AddTmplsToTmplSetsWithTx(kt, tx, templateIds,
		req.TemplateSetIds); err != nil {
		logs.Errorf("add templates to template sets failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, fmt.Sprintf("add templates to template sets failed, err: %v, rid: %s", err, kt.Rid)))
	}

	// 验证套餐中模板数量
	if len(req.GetTemplateSetIds()) > 0 {
		for _, tmplSetID := range req.TemplateSetIds {
			if err := s.dao.TemplateSet().ValidateTmplNumber(kt, tx, req.GetBizId(), tmplSetID); err != nil {
				logs.Errorf("validate template set's templates count failed, err: %v, rid: %s", err, kt.Rid)
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
				}
				return nil, err
			}
		}
	}

	// 4. 查询绑定服务的套餐以及更新绑定的数据
	atbs, err := s.dao.TemplateBindingRelation().
		ListTemplateSetsBoundATBs(kt, req.BizId, req.TemplateSetIds)
	if err != nil {
		logs.Errorf("list template sets bound app template bindings failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(0,
			i18n.T(kt, fmt.Sprintf("list template sets bound app template bindings failed, err: %v, rid: %s", err, kt.Rid)))
	}

	// 验证被引用的套餐是否超出限额
	if len(atbs) > 0 {
		for _, atb := range atbs {
			if err := s.CascadeUpdateATB(kt, tx, atb); err != nil {
				logs.Errorf("cascade update app template binding failed, err: %v, rid: %s", err, kt.Rid)
				if e := tx.Rollback(); e != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", e, kt.Rid)
				}
				return nil, err
			}
		}
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return &pbds.BatchUpsertTemplatesReqResp{
		Ids: templateIds,
	}, nil
}

// 创建模板配置以及模板配置版本
func (s *Service) doBatchCreateTemplates(kt *kit.Kit, tx *gen.QueryTx, createData []*pbds.BatchUpsertTemplatesReq_Item,
	now time.Time) ([]uint32, error) {

	createIds := []uint32{}
	toCreate := []*table.Template{}

	for _, item := range createData {
		toCreate = append(toCreate, &table.Template{
			Spec:       item.GetTemplate().GetSpec().TemplateSpec(),
			Attachment: item.GetTemplate().GetAttachment().TemplateAttachment(),
			Revision: &table.Revision{
				Creator:   kt.User,
				Reviser:   kt.User,
				CreatedAt: now,
				UpdatedAt: now,
			},
		})
	}

	if err := s.dao.Template().BatchCreateWithTx(kt, tx, toCreate); err != nil {
		logs.Errorf("batch create templates failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, fmt.Sprintf("batch create templates failed, err: %v, rid: %s", err, kt.Rid)))
	}

	toCreateTr := []*table.TemplateRevision{}
	for i, item := range createData {
		createIds = append(createIds, toCreate[i].ID)
		toCreateTr = append(toCreateTr, &table.TemplateRevision{
			Spec: item.GetTemplateRevision().GetSpec().TemplateRevisionSpec(),
			Attachment: &table.TemplateRevisionAttachment{
				BizID:           item.GetTemplateRevision().GetAttachment().TemplateRevisionAttachment().BizID,
				TemplateSpaceID: item.GetTemplateRevision().GetAttachment().TemplateRevisionAttachment().TemplateSpaceID,
				TemplateID:      toCreate[i].ID,
			},
			Revision: &table.CreatedRevision{
				Creator:   kt.User,
				CreatedAt: now,
			},
		})
	}

	if err := s.doCreateTemplateRevisions(kt, tx, toCreateTr); err != nil {
		return nil, err
	}

	return createIds, nil
}

// 更新模板配置以及模板配置版本
func (s *Service) doBatchUpdateTemplates(kt *kit.Kit, tx *gen.QueryTx, updateData []*pbds.BatchUpsertTemplatesReq_Item,
	oldTemplateData map[uint32]*table.Template, now time.Time) error {
	toUpdate := []*table.Template{}
	for _, item := range updateData {
		toUpdate = append(toUpdate, &table.Template{
			ID:         item.GetTemplate().GetId(),
			Spec:       item.GetTemplate().GetSpec().TemplateSpec(),
			Attachment: item.GetTemplate().GetAttachment().TemplateAttachment(),
			Revision: &table.Revision{
				Creator:   oldTemplateData[item.GetTemplate().GetId()].Revision.Creator,
				Reviser:   kt.User,
				CreatedAt: oldTemplateData[item.GetTemplate().GetId()].Revision.CreatedAt,
				UpdatedAt: now,
			},
		})
	}
	if err := s.dao.Template().BatchUpdateWithTx(kt, tx, toUpdate); err != nil {
		logs.Errorf("batch update templates failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	toUpdateTr := []*table.TemplateRevision{}
	for i, item := range updateData {
		toUpdateTr = append(toUpdateTr, &table.TemplateRevision{
			Spec: item.GetTemplateRevision().GetSpec().TemplateRevisionSpec(),
			Attachment: &table.TemplateRevisionAttachment{
				BizID:           item.GetTemplateRevision().GetAttachment().TemplateRevisionAttachment().BizID,
				TemplateSpaceID: item.GetTemplateRevision().GetAttachment().TemplateRevisionAttachment().TemplateSpaceID,
				TemplateID:      toUpdate[i].ID,
			},
			Revision: &table.CreatedRevision{
				Creator:   kt.User,
				CreatedAt: now,
			},
		})
	}

	if err := s.doCreateTemplateRevisions(kt, tx, toUpdateTr); err != nil {
		return err
	}

	return nil
}

func (s *Service) validateBatchUpsertTemplates(grpcKit *kit.Kit, updateIds []uint32,
	updateData []*pbds.BatchUpsertTemplatesReq_Item) (map[uint32]*table.Template, error) {
	// 针对更新的数据做查询 获取创建时间和创建人
	template, err := s.dao.Template().ListByIDs(grpcKit, updateIds)
	if err != nil {
		logs.Errorf("list templates by template ids failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	oldTemplateData := make(map[uint32]*table.Template)
	for _, item := range template {
		oldTemplateData[item.ID] = item
	}

	// 通过模板id获取最新的template revision数据
	templateRevisions, err := s.dao.TemplateRevision().ListLatestRevisionsGroupByTemplateIds(grpcKit, updateIds)
	if err != nil {
		return nil, err
	}
	oldTemplateRevisionData := make(map[uint32]*table.TemplateRevision)
	for _, item := range templateRevisions {
		oldTemplateRevisionData[item.Attachment.TemplateID] = item
	}

	// 验证类型是否变更
	for _, item := range updateData {
		if item.GetTemplateRevision().GetSpec().GetFileType() !=
			string(oldTemplateRevisionData[item.GetTemplate().GetId()].Spec.FileType) {
			logs.Errorf("batch create templates failed, err: %v, rid: %s, templateId: %s",
				err, grpcKit.Rid, item.GetTemplate().GetId())
			return nil, fmt.Errorf("模板配置文件名 %s: 不支持更改文件类型", item.Template.GetSpec().Name)
		}
	}
	return oldTemplateData, nil
}

// BatchUpdateTemplatePermissions 批量更新模板权限
// nolint:funlen
func (s *Service) BatchUpdateTemplatePermissions(ctx context.Context, req *pbds.BatchUpdateTemplatePermissionsReq) (
	*pbds.BatchUpdateTemplatePermissionsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// 获取最新的模板配置
	tmps, err := s.dao.TemplateRevision().ListLatestRevisionsGroupByTemplateIds(kt, req.GetTemplateIds())
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, fmt.Sprintf("lists the latest version by template ids failed, err: %s", err.Error())))
	}

	now := time.Now().UTC()
	toCreate := make([]*table.TemplateRevision, 0)
	for _, v := range tmps {
		v.Spec.RevisionName = tools.GenerateRevisionName()
		if req.User != "" {
			v.Spec.Permission.User = req.User
		}
		if req.UserGroup != "" {
			v.Spec.Permission.UserGroup = req.UserGroup
		}
		if req.Privilege != "" {
			v.Spec.Permission.Privilege = req.Privilege
		}
		v.Revision = &table.CreatedRevision{Creator: kt.User, CreatedAt: now}
		toCreate = append(toCreate, &table.TemplateRevision{
			Spec:       v.Spec,
			Attachment: v.Attachment,
			Revision:   v.Revision,
		})
	}

	tx := s.dao.GenQuery().Begin()
	if err := s.dao.TemplateRevision().BatchCreateWithTx(kt, tx, toCreate); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, fmt.Sprintf("batch update of template permissions failed, err: %s", err.Error())))
	}

	ids := []uint32{}
	templateIds := []uint32{}
	templateRevisionID := map[uint32]uint32{}
	for _, v := range toCreate {
		ids = append(ids, v.ID)
		templateIds = append(templateIds, v.Attachment.TemplateID)
		templateRevisionID[v.Attachment.TemplateID] = v.ID
	}

	if len(req.GetAppIds()) > 0 {
		// 更新引用的服务
		items, err := s.dao.AppTemplateBinding().ListAppTemplateBindingByAppIds(kt, req.GetBizId(), req.GetAppIds())
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			templateRevisionIDs := make([]uint32, 0, len(item.Spec.Bindings))
			for _, binding := range item.Spec.Bindings {
				for _, revision := range binding.TemplateRevisions {
					// 如果存在更新TemplateRevisionID
					if id, exists := templateRevisionID[revision.TemplateID]; exists && id > 0 {
						revision.TemplateRevisionID = id
						revision.IsLatest = true
						templateRevisionIDs = append(templateRevisionIDs, id)
					} else {
						templateRevisionIDs = append(templateRevisionIDs, revision.TemplateRevisionID)
					}
				}
			}
			item.Revision.Reviser = kt.User
			item.Revision.UpdatedAt = now
			item.Spec.TemplateRevisionIDs = tools.RemoveDuplicates(templateRevisionIDs)
			item.Spec.LatestTemplateIDs = tools.RemoveDuplicates(tools.MergeAndDeduplicate(templateIds,
				item.Spec.LatestTemplateIDs))
		}

		// 更新未命名版本绑定关系
		if err := s.dao.AppTemplateBinding().BatchUpdateWithTx(kt, tx, items); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, errf.Errorf(errf.DBOpFailed,
				i18n.T(kt, fmt.Sprintf("batch update of template permissions failed, err: %s", err.Error())))
		}
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, fmt.Sprintf("batch update of template permissions failed, err: %s", e.Error())))
	}

	return &pbds.BatchUpdateTemplatePermissionsResp{Ids: ids}, nil
}
