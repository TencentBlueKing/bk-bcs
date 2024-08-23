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

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
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

	// 1. 同一空间不能出现相同绝对路径的配置文件且同路径下不能出现同名的文件夹和文件 比如: /a.txt 和 /a/1.txt
	if tools.CheckPathConflict(path.Join(req.Spec.Path, req.Spec.Name), existingPaths) {
		return nil, errors.New(i18n.T(kt, "the config file %s already exists in this space and cannot be created again",
			path.Join(req.Spec.Path, req.Spec.Name)))
	}

	if len(req.TemplateSetIds) > 0 {
		// ValidateTmplSetsExist validate if template sets exists
		if err = s.dao.Validator().ValidateTmplSetsExist(kt, req.TemplateSetIds); err != nil {
			return nil, err
		}
	}

	tx := s.dao.GenQuery().Begin()

	// 2. 验证套餐是否超出
	templateSets, err := s.verifyTemplateSetAndReturnData(kt, tx, req.GetAttachment().BizId,
		req.GetTemplateSetIds(), []uint32{}, 1)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 3. 通过套餐获取绑定的服务数据
	bindings, err := s.dao.AppTemplateBinding().GetBindingAppByTemplateSetID(kt, req.GetAttachment().BizId,
		req.GetTemplateSetIds())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, "get app template bindings by template set ids, err: %s", err))
	}

	// 4. 验证服务套餐下是否超出限制
	if err = s.verifyAppReferenceTmplSetExceedsLimit(kt, req.GetAttachment().BizId, bindings, req.GetTemplateSetIds(),
		[]uint32{}, 1); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 5. create template
	template := &table.Template{
		Spec:       req.Spec.TemplateSpec(),
		Attachment: req.Attachment.TemplateAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	// CreateWithTx create one template instance with transaction.
	templateID, err := s.dao.Template().CreateWithTx(kt, tx, template)
	if err != nil {
		logs.Errorf("create template failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 6. create template revision
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
			TemplateID:      templateID,
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

	// 7. 处理和服务之前的绑定关系
	appTemplateBindings, errH := s.handleAppTemplateBindings(kt, tx, req.GetAttachment().BizId, bindings,
		req.GetTemplateSetIds(), []uint32{templateID})
	if errH != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errH
	}

	// 8. 更新引用的服务
	if len(appTemplateBindings) > 0 {
		if err = s.dao.AppTemplateBinding().BatchUpdateWithTx(kt, tx, appTemplateBindings); err != nil {
			logs.Errorf("batch update app template binding's failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "batch update app template binding's failed, err: %s", err))
		}
	}

	for _, v := range templateSets {
		v.Spec.TemplateIDs = tools.MergeAndDeduplicate(v.Spec.TemplateIDs, []uint32{templateID})
	}

	// 9. 添加至模板套餐中
	if err := s.dao.TemplateSet().BatchAddTmplsToTmplSetsWithTx(kt, tx, templateSets); err != nil {
		logs.Errorf("batch add templates to template sets failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "batch add templates to template sets failed, err: %s", err))
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return &pbds.CreateResp{Id: templateID}, nil
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
	// List templates with options.
	details, count, err := s.dao.Template().List(kt, req.BizId, req.TemplateSpaceId, searcher,
		opt, req.Ids, req.SearchValue)

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

	var err error
	templateIds := req.GetTemplateIds()
	// 1. 判断是否排除操作
	if req.ExclusionOperation {
		templateIds, err = s.getExclusionOperationID(kt, req.BizId, req.TemplateSetId,
			req.TemplateSpaceId, req.NoSetSpecified, req.TemplateIds)
		if err != nil {
			return nil, err
		}
	}

	if err = s.dao.Validator().ValidateTemplatesExist(kt, templateIds); err != nil {
		return nil, err
	}
	if err = s.dao.Validator().ValidateTmplSetsExist(kt, req.TemplateSetIds); err != nil {
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()

	// 2. 验证套餐是否超出
	templateSets, err := s.verifyTemplateSetAndReturnData(kt, tx, req.GetBizId(), req.TemplateSetIds, templateIds, 0)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 3. 通过套餐获取绑定的服务数据
	bindings, err := s.dao.AppTemplateBinding().GetBindingAppByTemplateSetID(kt, req.BizId, req.TemplateSetIds)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, "get app template bindings by template set ids, err: %s", err))
	}

	// 4. 验证需要编辑和新增的模板是否超出服务限制
	if err = s.verifyAppReferenceTmplSetExceedsLimit(kt, req.GetBizId(), bindings,
		req.GetTemplateSetIds(), templateIds, 0); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 5. 处理模板绑定的数据
	appTemplateBindings, err := s.handleAppTemplateBindings(kt, tx, req.GetBizId(), bindings,
		req.GetTemplateSetIds(), templateIds)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 8. 更新引用的服务
	if len(appTemplateBindings) > 0 {
		if err = s.dao.AppTemplateBinding().BatchUpdateWithTx(kt, tx, appTemplateBindings); err != nil {
			logs.Errorf("batch update app template binding's failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "batch update app template binding's failed, err: %s", err))
		}
	}
	for _, v := range templateSets {
		v.Spec.TemplateIDs = tools.MergeAndDeduplicate(v.Spec.TemplateIDs, templateIds)
	}

	// 10. 添加至模板套餐中
	if err := s.dao.TemplateSet().BatchAddTmplsToTmplSetsWithTx(kt, tx, templateSets); err != nil {
		logs.Errorf("batch add templates to template sets failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "batch add templates to template sets failed, err: %s", err))
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

	var err error
	templateIds := req.GetTemplateIds()
	// 判断是否排除操作
	if req.ExclusionOperation {
		templateIds, err = s.getExclusionOperationID(kt, req.BizId, req.TemplateSetId,
			req.TemplateSpaceId, req.NoSetSpecified, req.TemplateIds)
		if err != nil {
			return nil, err
		}
	}

	if err = s.dao.Validator().ValidateTemplatesExist(kt, templateIds); err != nil {
		return nil, err
	}
	if err = s.dao.Validator().ValidateTmplSetsExist(kt, req.TemplateSetIds); err != nil {
		return nil, err
	}

	// 1. 获取套餐下数据
	templateSet, err := s.dao.TemplateSet().GetByTemplateSetByID(kt, req.BizId, req.TemplateSetId)
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "get template set data failed, err: %s", err))
	}

	if len(templateSet.Spec.TemplateIDs) == 0 {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "there is no template file under this template set"))
	}

	// 排除移除的模板
	excludedTemplateIDs := tools.Difference(templateSet.Spec.TemplateIDs, templateIds)
	templateSet.Spec.TemplateIDs = excludedTemplateIDs

	tx := s.dao.GenQuery().Begin()

	// 2. 移除套餐中指定的模板
	if err = s.dao.TemplateSet().UpdateWithTx(kt, tx, &table.TemplateSet{
		ID:         templateSet.ID,
		Spec:       templateSet.Spec,
		Attachment: templateSet.Attachment,
		Revision: &table.Revision{
			Creator:   templateSet.Revision.Creator,
			Reviser:   kt.User,
			CreatedAt: templateSet.Revision.CreatedAt,
			UpdatedAt: time.Now().UTC(),
		},
	}); err != nil {
		logs.Errorf("delete template from template sets failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, "delete template from template sets failed, err: %v", err))
	}

	// 3. 通过套餐获取绑定的服务数据
	bindings, err := s.dao.AppTemplateBinding().GetBindingAppByTemplateSetID(kt, req.BizId, req.TemplateSetIds)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, "get app template bindings by template set ids, err: %v", err))
	}

	// 4. 移除服务引用套餐下的模板
	if bindings != nil {
		appTemplateBindings := deleteTemplateSetReferencedApp(req.TemplateSetId, req.TemplateIds, bindings)
		if err = s.dao.AppTemplateBinding().BatchUpdateWithTx(kt, tx, appTemplateBindings); err != nil {
			logs.Errorf("batch update app template binding's failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, errf.Errorf(errf.DBOpFailed,
				i18n.T(kt, "batch update app template binding's failed, err: %v", err))
		}
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, "delete template from template sets failed, err: %v", e))
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
	sort.SliceStable(details, func(i, j int) bool {
		iInTopID := tools.Contains(req.Ids, details[i].Id)
		jInTopID := tools.Contains(req.Ids, details[j].Id)
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
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "list templates by tuple failed, err: %v", err))
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

	// 2. 过滤出创建和更新的数据
	createData, updateData := make([]*pbds.BatchUpsertTemplatesReq_Item, 0), make([]*pbds.BatchUpsertTemplatesReq_Item, 0)
	updateIds := []uint32{}
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

	// 3. 验证套餐是否超出
	templateSets, err := s.verifyTemplateSetAndReturnData(kt, tx, req.GetBizId(),
		req.GetTemplateSetIds(), updateIds, len(createData))
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 4. 通过套餐获取绑定的服务数据
	bindings, err := s.dao.AppTemplateBinding().GetBindingAppByTemplateSetID(kt, req.GetBizId(), req.GetTemplateSetIds())
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kt, "get app template bindings by template set ids, err: %s", err))
	}

	// 5. 验证需要编辑和新增的模板是否超出服务限制
	if err = s.verifyAppReferenceTmplSetExceedsLimit(kt, req.GetBizId(), bindings, req.GetTemplateSetIds(),
		updateIds, len(createData)); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// 6. 批量创建模板以及模板版本
	createIds, e := s.doBatchCreateTemplates(kt, tx, createData, now)
	if e != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}

	// 7.批量更新模板以及模板版本
	if len(updateIds) > 0 {
		oldTemplateData, errV := s.validateBatchUpsertTemplates(kt, updateIds, updateData)
		if errV != nil {
			return nil, errV
		}
		if e := s.doBatchUpdateTemplates(kt, tx, updateData, oldTemplateData, now); e != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, e
		}
	}

	// 合并创建和编辑后的模板ID
	templateIDs := tools.MergeAndDeduplicate(createIds, updateIds)

	// 8. 处理和服务之前的绑定关系
	appTemplateBindings, errH := s.handleAppTemplateBindings(kt, tx, req.GetBizId(), bindings,
		req.GetTemplateSetIds(), templateIDs)
	if errH != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errH
	}

	// 9. 更新引用的服务
	if len(appTemplateBindings) > 0 {
		if err = s.dao.AppTemplateBinding().BatchUpdateWithTx(kt, tx, appTemplateBindings); err != nil {
			logs.Errorf("batch update app template binding's failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "batch update app template binding's failed, err: %s", err))
		}
	}

	for _, v := range templateSets {
		v.Spec.TemplateIDs = tools.MergeAndDeduplicate(v.Spec.TemplateIDs, templateIDs)
	}

	// 10. 添加至模板套餐中
	if err := s.dao.TemplateSet().BatchAddTmplsToTmplSetsWithTx(kt, tx, templateSets); err != nil {
		logs.Errorf("batch add templates to template sets failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "batch add templates to template sets failed, err: %s", err))
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &pbds.BatchUpsertTemplatesReqResp{
		Ids: templateIDs,
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

	var err error
	templateIds := req.GetTemplateIds()
	// 判断是否排除操作
	if req.ExclusionOperation {
		templateIds, err = s.getExclusionOperationID(kt, req.BizId, req.TemplateSetId,
			req.TemplateSpaceId, req.NoSetSpecified, req.TemplateIds)
		if err != nil {
			return nil, err
		}
	}

	// 获取最新的模板配置
	tmps, err := s.dao.TemplateRevision().ListLatestRevisionsGroupByTemplateIds(kt, templateIds)
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
	templateIdsMap := []uint32{}
	templateRevisionID := map[uint32]uint32{}
	for _, v := range toCreate {
		ids = append(ids, v.ID)
		templateIdsMap = append(templateIdsMap, v.Attachment.TemplateID)
		templateRevisionID[v.Attachment.TemplateID] = v.ID
	}

	if len(req.GetAppIds()) > 0 {
		// 更新引用的服务
		items, err := s.dao.AppTemplateBinding().ListAppTemplateBindingByAppIds(kt, req.GetBizId(), req.GetAppIds())
		if err != nil {
			return nil, errf.Errorf(errf.DBOpFailed,
				i18n.T(kt, "list app template bindings by app ids failed, err: %s", err))
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
			item.Spec.LatestTemplateIDs = tools.RemoveDuplicates(tools.MergeAndDeduplicate(templateIdsMap,
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

// 验证套餐并返回套餐数据
func (s *Service) verifyTemplateSetAndReturnData(kt *kit.Kit, tx *gen.QueryTx, bizID uint32, templateSetIds []uint32,
	templateIDs []uint32, additionalQuantity int) ([]*table.TemplateSet, error) {

	tmplSetTmplCnt := getTmplSetTmplCnt(bizID)

	// 1. 查询套餐数据
	templateSets, err := s.dao.TemplateSet().ListByIDsWithTx(kt, tx, templateSetIds)
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "get template set failed, err: %s", err))
	}

	// 2. 验证是否超出套餐限制
	for _, v := range templateSets {
		mergedIDs := tools.MergeAndDeduplicate(v.Spec.TemplateIDs, templateIDs)
		if len(mergedIDs)+additionalQuantity > tmplSetTmplCnt {
			return nil, errf.New(errf.InvalidParameter,
				i18n.T(kt, "the total number of template set %s templates exceeded the limit %d",
					v.Spec.Name, tmplSetTmplCnt))
		}
		v.Spec.TemplateIDs = mergedIDs
	}

	return templateSets, nil
}

// 验证服务下引用的套餐
func (s *Service) verifyAppReferenceTmplSetExceedsLimit(kt *kit.Kit, bizID uint32, bindings []*table.AppTemplateBinding,
	templateSetIDs, templateIDs []uint32, additionalQuantity int) error {

	if bindings == nil {
		return nil
	}

	betectedTemplateSetIDs := make(map[uint32]bool, 0)
	for _, v := range templateSetIDs {
		betectedTemplateSetIDs[v] = true
	}

	appConfigCnt := getAppConfigCnt(bizID)

	// 1. 统计服务下套餐的数量，不包含需要操作的套餐
	configCountWithTemplates, excludedTemplateSetIDs, err := s.countNumberAppTemplateBindings(kt, bizID, bindings,
		betectedTemplateSetIDs, additionalQuantity)
	if err != nil {
		return err
	}

	// 2. 验证服务引用的套餐是否超出限制
	var appIDs []uint32
	for _, v := range bindings {
		appIDs = append(appIDs, v.Attachment.AppID)
	}

	app, err := s.dao.App().ListAppsByIDs(kt, appIDs)
	if err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kt, "list apps by app ids failed, err: %s", err))
	}

	configItemName := make(map[uint32]string, 0)
	for _, v := range app {
		configItemName[v.ID] = v.Spec.Name
	}

	for _, templateBinding := range bindings {
		for _, spec := range templateBinding.Spec.Bindings {
			// 不需要处理的套餐忽略掉
			if !betectedTemplateSetIDs[spec.TemplateSetID] {
				continue
			}
			// 需验证套餐数量是否超出限制
			// 需要验证的套餐配置ID和需要操作的配置文件ID合并，判断是否超出限制
			mergedIDs := tools.MergeAndDeduplicate(excludedTemplateSetIDs[spec.TemplateSetID], templateIDs)
			if len(mergedIDs)+configCountWithTemplates[templateBinding.Attachment.AppID] > appConfigCnt {
				return errf.New(errf.InvalidParameter,
					i18n.T(kt, "the total number of app %s config items(including template and non-template)"+
						"exceeded the limit %d", configItemName[templateBinding.Attachment.AppID], appConfigCnt))
			}
		}
	}

	return nil
}

// 统计服务套餐下配置项数量
// 某个服务下套餐存在betectedTemplateSetIDs中时返回该套餐的模板ID
// 否则统计该服务下的数量
func (s *Service) countNumberAppTemplateBindings(kt *kit.Kit, bizID uint32, bindings []*table.AppTemplateBinding,
	betectedTemplateSetIDs map[uint32]bool, additionalQuantity int) (
	map[uint32]int, map[uint32][]uint32, error) {

	var appIDs []uint32
	for _, v := range bindings {
		appIDs = append(appIDs, v.Attachment.AppID)
	}

	// 获取服务配置数量
	result, err := s.dao.ConfigItem().ListConfigItemCount(kt, bizID, appIDs)
	if err != nil {
		return nil, nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "count the number of app configs failed, err: %s", err))
	}

	configItemCount := make(map[uint32]int, 0)
	for _, v := range result {
		configItemCount[v.AppID] = int(v.Count) + additionalQuantity
	}

	configCountWithTemplates := make(map[uint32]int)
	excludedTemplateSetIDs := make(map[uint32][]uint32)
	for _, templateBinding := range bindings {
		number := 0
		for _, binding := range templateBinding.Spec.Bindings {
			if betectedTemplateSetIDs[binding.TemplateSetID] {
				templateIDs := []uint32{}
				for _, v := range binding.TemplateRevisions {
					templateIDs = append(templateIDs, v.TemplateID)
				}
				excludedTemplateSetIDs[binding.TemplateSetID] = templateIDs
			} else {
				number += len(binding.TemplateRevisions)
			}
		}
		configCountWithTemplates[templateBinding.Attachment.AppID] = number
	}

	// 服务下模板套餐配置数量( 不包含 betectedTemplateSetIDs 套餐下的模板)+非模板配置数量
	for k, v := range configCountWithTemplates {
		configCountWithTemplates[k] = configItemCount[k] + v
	}

	return configCountWithTemplates, excludedTemplateSetIDs, nil
}

// 获取需要更新的模板绑定数据
func (s *Service) handleAppTemplateBindings(kt *kit.Kit, tx *gen.QueryTx, bizID uint32,
	bindings []*table.AppTemplateBinding, templateSetIDs, templateIDs []uint32) (
	[]*table.AppTemplateBinding, error) {
	if bindings == nil {
		return nil, nil
	}

	betectedTemplateSetIDs := make(map[uint32]bool, 0)
	for _, v := range templateSetIDs {
		betectedTemplateSetIDs[v] = true
	}

	// 服务套餐模板绑定
	appTemplateBindings := make([]*table.AppTemplateBinding, 0, len(bindings))
	for _, templateBinding := range bindings {
		// 单个服务绑定的数据
		specBindings := make([]*table.TemplateBinding, 0, len(templateBinding.Spec.Bindings))
		for _, spec := range templateBinding.Spec.Bindings {
			revisions := make([]*table.TemplateRevisionBinding, 0)
			if betectedTemplateSetIDs[spec.TemplateSetID] {
				existingTemplateIds := []uint32{}
				existingTemplateData := map[uint32]*table.TemplateRevisionBinding{}
				for _, v := range spec.TemplateRevisions {
					existingTemplateIds = append(existingTemplateIds, v.TemplateID)
					existingTemplateData[v.TemplateID] = v
				}
				mergedIDs := tools.MergeAndDeduplicate(existingTemplateIds, templateIDs)
				revisionsData, err := s.dao.TemplateRevision().ListLatestGroupByTemplateIdsWithTx(kt, tx, bizID, mergedIDs)
				if err != nil {
					return nil, err
				}
				revisionsId := make(map[uint32]uint32, 0)
				for _, v := range revisionsData {
					revisionsId[v.Attachment.TemplateID] = v.ID
				}
				updateRevisions := func(templateID, revisionsID uint32, isLatest bool) {
					exists := &table.TemplateRevisionBinding{
						TemplateID:         templateID,
						TemplateRevisionID: revisionsID,
						IsLatest:           isLatest,
					}
					revisions = append(revisions, exists)
				}
				// 1. 如果模板ID已存在判断是否是latest，是更新版本号，否则不处理
				// 2. 如果模板ID不存在需要插入到该模板套餐中
				for _, v := range mergedIDs {
					if exists, ok := existingTemplateData[v]; ok {
						if exists.IsLatest {
							updateRevisions(v, revisionsId[v], true)
						} else {
							updateRevisions(v, exists.TemplateRevisionID, exists.IsLatest)
						}
					} else {
						updateRevisions(v, revisionsId[v], true)
					}
				}
			} else {
				revisions = spec.TemplateRevisions
			}
			specBinding := &table.TemplateBinding{
				TemplateSetID:     spec.TemplateSetID,
				TemplateRevisions: revisions,
			}
			specBindings = append(specBindings, specBinding)
		}
		templateIDs, latestTemplateIDs, templateRevisionIDs := []uint32{}, []uint32{}, []uint32{}
		for _, specBinding := range specBindings {
			for _, v := range specBinding.TemplateRevisions {
				templateIDs = append(templateIDs, v.TemplateID)
				if v.IsLatest {
					latestTemplateIDs = append(latestTemplateIDs, v.TemplateID)
				}
				templateRevisionIDs = append(templateRevisionIDs, v.TemplateRevisionID)
			}
		}
		appTemplateBindings = append(appTemplateBindings, &table.AppTemplateBinding{
			ID: templateBinding.ID,
			Spec: &table.AppTemplateBindingSpec{
				TemplateSpaceIDs:    templateBinding.Spec.TemplateSpaceIDs,
				TemplateSetIDs:      templateBinding.Spec.TemplateSetIDs,
				TemplateIDs:         tools.RemoveDuplicates(templateIDs),
				TemplateRevisionIDs: tools.RemoveDuplicates(templateRevisionIDs),
				LatestTemplateIDs:   tools.RemoveDuplicates(latestTemplateIDs),
				Bindings:            specBindings,
			},
			Attachment: templateBinding.Attachment,
			Revision:   templateBinding.Revision,
		})
	}

	return appTemplateBindings, nil
}

// 把某个服务下套餐中的模板移除
func deleteTemplateSetReferencedApp(templateSetId uint32, templateIds []uint32,
	bindings []*table.AppTemplateBinding) []*table.AppTemplateBinding {

	// 待移除的模板ID
	excludedTemplateIds := make(map[uint32]bool)
	for _, v := range templateIds {
		excludedTemplateIds[v] = true
	}

	// 服务套餐模板绑定
	appTemplateBindings := make([]*table.AppTemplateBinding, 0, len(bindings))
	for _, templateBinding := range bindings {
		// 单个服务绑定的数据
		specBindings := make([]*table.TemplateBinding, 0, len(templateBinding.Spec.Bindings))
		for _, spec := range templateBinding.Spec.Bindings {
			revisions := make([]*table.TemplateRevisionBinding, 0)
			// 只需移除指定套餐下的模板
			if templateSetId == spec.TemplateSetID {
				for _, v := range spec.TemplateRevisions {
					// 判断模板ID是否存在
					if !excludedTemplateIds[v.TemplateID] {
						revisions = append(revisions, v)
						continue
					}
				}
			} else {
				revisions = spec.TemplateRevisions
			}
			specBinding := &table.TemplateBinding{
				TemplateSetID:     spec.TemplateSetID,
				TemplateRevisions: revisions,
			}
			specBindings = append(specBindings, specBinding)
		}

		templateIDs, latestTemplateIDs, templateRevisionIDs := []uint32{}, []uint32{}, []uint32{}
		for _, specBinding := range specBindings {
			for _, v := range specBinding.TemplateRevisions {
				templateIDs = append(templateIDs, v.TemplateID)
				if v.IsLatest {
					latestTemplateIDs = append(latestTemplateIDs, v.TemplateID)
				}
				templateRevisionIDs = append(templateRevisionIDs, v.TemplateRevisionID)
			}
		}

		appTemplateBindings = append(appTemplateBindings, &table.AppTemplateBinding{
			ID: templateBinding.ID,
			Spec: &table.AppTemplateBindingSpec{
				TemplateSpaceIDs:    templateBinding.Spec.TemplateSpaceIDs,
				TemplateSetIDs:      templateBinding.Spec.TemplateSetIDs,
				TemplateIDs:         tools.RemoveDuplicates(templateIDs),
				TemplateRevisionIDs: tools.RemoveDuplicates(templateRevisionIDs),
				LatestTemplateIDs:   tools.RemoveDuplicates(latestTemplateIDs),
				Bindings:            specBindings,
			},
			Attachment: templateBinding.Attachment,
			Revision:   templateBinding.Revision,
		})
	}

	return appTemplateBindings
}

// 获取反向操作的ID
func (s *Service) getExclusionOperationID(kt *kit.Kit, bizID, templateSetID, templateSpaceID uint32,
	noSetSpecified bool, templateIds []uint32) ([]uint32, error) {

	exclusiontemplateIDs := []uint32{}
	var err error

	// 指定套餐
	if templateSetID != 0 {
		templateSet, errS := s.dao.TemplateSet().GetByTemplateSetByID(kt, bizID, templateSetID)
		if errS != nil {
			return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "get template set data failed, err: %s", errS))
		}
		exclusiontemplateIDs = templateSet.Spec.TemplateIDs
	}

	// 全部配置文件
	if templateSetID == 0 && !noSetSpecified {
		exclusiontemplateIDs, err = s.dao.Template().ListAllIDs(kt, bizID, templateSpaceID)
		if err != nil {
			return nil, err
		}
	}

	// 未指定套餐
	if templateSetID == 0 && noSetSpecified {
		idsAll, err := s.dao.Template().ListAllIDs(kt, bizID, templateSpaceID)
		if err != nil {
			logs.Errorf("list templates not bound failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		idsBound, err := s.dao.TemplateSet().ListAllTemplateIDs(kt, bizID, templateSpaceID)
		if err != nil {
			logs.Errorf("list templates not bound failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		exclusiontemplateIDs = tools.SliceDiff(idsAll, idsBound)
	}

	return tools.Difference(exclusiontemplateIDs, templateIds), nil
}

func getTmplSetTmplCnt(bizID uint32) int {
	if resLimit, ok := cc.DataService().FeatureFlags.ResourceLimit.Spec[fmt.Sprintf("%d", bizID)]; ok {
		if resLimit.TmplSetTmplCnt > 0 {
			return int(resLimit.TmplSetTmplCnt)
		}
	}
	return int(cc.DataService().FeatureFlags.ResourceLimit.Default.TmplSetTmplCnt)
}

func getAppConfigCnt(bizID uint32) int {
	if resLimit, ok := cc.DataService().FeatureFlags.ResourceLimit.Spec[fmt.Sprintf("%d", bizID)]; ok {
		if resLimit.AppConfigCnt > 0 {
			return int(resLimit.AppConfigCnt)
		}
	}
	return int(cc.DataService().FeatureFlags.ResourceLimit.Default.AppConfigCnt)
}
