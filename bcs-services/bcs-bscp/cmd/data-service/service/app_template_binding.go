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
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbatb "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app-template-binding"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbrci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/released-ci"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateAppTemplateBinding create app template binding.
func (s *Service) CreateAppTemplateBinding(ctx context.Context, req *pbds.CreateAppTemplateBindingReq) (
	*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	appTemplateBinding := &table.AppTemplateBinding{
		Spec:       req.Spec.AppTemplateBindingSpec(),
		Attachment: req.Attachment.AppTemplateBindingAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}

	if err := s.genFinalATB(kt, appTemplateBinding); err != nil {
		logs.Errorf("create app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()

	id, err := s.dao.AppTemplateBinding().CreateWithTx(kt, tx, appTemplateBinding)
	if err != nil {
		logs.Errorf("create app template binding failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	// validate config items count.
	if err = s.dao.ConfigItem().ValidateAppCINumber(kt, tx, req.Attachment.BizId, req.Attachment.AppId); err != nil {
		logs.Errorf("validate config items count failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListAppTemplateBindings list app template binding.
func (s *Service) ListAppTemplateBindings(ctx context.Context, req *pbds.ListAppTemplateBindingsReq) (
	*pbds.ListAppTemplateBindingsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
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
func (s *Service) UpdateAppTemplateBinding(ctx context.Context, req *pbds.UpdateAppTemplateBindingReq) (
	*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	appTemplateBinding := &table.AppTemplateBinding{
		ID:         req.Id,
		Spec:       req.Spec.AppTemplateBindingSpec(),
		Attachment: req.Attachment.AppTemplateBindingAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}

	if err := s.genFinalATB(kt, appTemplateBinding); err != nil {
		logs.Errorf("update app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()

	if err := s.dao.AppTemplateBinding().UpdateWithTx(kt, tx, appTemplateBinding); err != nil {
		logs.Errorf("update app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// validate config items count.
	if err := s.dao.ConfigItem().ValidateAppCINumber(kt, tx, req.Attachment.BizId, req.Attachment.AppId); err != nil {
		logs.Errorf("validate config items count failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteAppTemplateBinding delete app template binding.
func (s *Service) DeleteAppTemplateBinding(ctx context.Context, req *pbds.DeleteAppTemplateBindingReq) (
	*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	appTemplateBinding := &table.AppTemplateBinding{
		ID:         req.Id,
		Attachment: req.Attachment.AppTemplateBindingAttachment(),
	}
	if err := s.dao.AppTemplateBinding().Delete(kt, appTemplateBinding); err != nil {
		logs.Errorf("delete app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// ListAppBoundTmplRevisions list app bound template revisions.
//
//nolint:funlen,gocyclo
func (s *Service) ListAppBoundTmplRevisions(ctx context.Context,
	req *pbds.ListAppBoundTmplRevisionsReq) (*pbds.ListAppBoundTmplRevisionsResp, error) {

	kt := kit.FromGrpcContext(ctx)

	// validate the page params
	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	// get app template binding
	atb, _, err := s.dao.AppTemplateBinding().List(kt, req.BizId, req.AppId, &types.BasePage{All: true})
	if err != nil {
		logs.Errorf("list app bound template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(atb) == 0 {
		return &pbds.ListAppBoundTmplRevisionsResp{
			Count:   0,
			Details: []*pbatb.AppBoundTmplRevision{},
		}, nil
	}

	var (
		tmplSpaces      []*table.TemplateSpace
		tmplSets        []*table.TemplateSet
		tmpls           []*table.Template
		tmplRevisions   []*table.TemplateRevision
		tmplSpaceMap    = make(map[uint32]*table.TemplateSpace)
		tmplSetMap      = make(map[uint32]*table.TemplateSet)
		tmplMap         = make(map[uint32]*table.Template)
		tmplRevisionMap = make(map[uint32]*table.TemplateRevision)
	)

	// get template space details
	if tmplSpaces, err = s.dao.TemplateSpace().ListByIDs(kt, atb[0].Spec.TemplateSpaceIDs); err != nil {
		logs.Errorf("list app bound template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	for _, r := range tmplSpaces {
		tmplSpaceMap[r.ID] = r
	}

	// get template set details
	if tmplSets, err = s.dao.TemplateSet().ListByIDs(kt, atb[0].Spec.TemplateSetIDs); err != nil {
		logs.Errorf("list app bound template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	for _, t := range tmplSets {
		tmplSetMap[t.ID] = t
	}

	// get template details
	if tmpls, err = s.dao.Template().ListByIDs(kt, atb[0].Spec.TemplateIDs); err != nil {
		logs.Errorf("list app bound template revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	for _, t := range tmpls {
		tmplMap[t.ID] = t
	}

	// get template revision details
	tmplRevisions, err = s.dao.TemplateRevision().ListByIDs(kt, atb[0].Spec.TemplateRevisionIDs)
	if err != nil {
		logs.Errorf("list app bound template revisions failed err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	for _, t := range tmplRevisions {
		tmplRevisionMap[t.ID] = t
	}

	// combine resp details
	details := make([]*pbatb.AppBoundTmplRevision, 0)
	for _, b := range atb[0].Spec.Bindings {
		for _, r := range b.TemplateRevisions {
			d := tmplRevisionMap[r.TemplateRevisionID]
			details = append(details, &pbatb.AppBoundTmplRevision{
				TemplateSpaceId:      d.Attachment.TemplateSpaceID,
				TemplateSpaceName:    tmplSpaceMap[d.Attachment.TemplateSpaceID].Spec.Name,
				TemplateSetId:        b.TemplateSetID,
				TemplateSetName:      tmplSetMap[b.TemplateSetID].Spec.Name,
				TemplateId:           d.Attachment.TemplateID,
				Name:                 tmplMap[d.Attachment.TemplateID].Spec.Name,
				Path:                 tmplMap[d.Attachment.TemplateID].Spec.Path,
				TemplateRevisionId:   r.TemplateRevisionID,
				IsLatest:             r.IsLatest,
				TemplateRevisionName: d.Spec.RevisionName,
				TemplateRevisionMemo: d.Spec.RevisionMemo,
				FileType:             string(d.Spec.FileType),
				FileMode:             string(d.Spec.FileMode),
				User:                 d.Spec.Permission.User,
				UserGroup:            d.Spec.Permission.UserGroup,
				Privilege:            d.Spec.Permission.Privilege,
				Signature:            d.Spec.ContentSpec.Signature,
				Md5:                  d.Spec.ContentSpec.Md5,
				ByteSize:             d.Spec.ContentSpec.ByteSize,
				Creator:              d.Revision.Creator,
				CreateAt:             d.Revision.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	if req.WithStatus {
		if details, err = s.setFileState(kt, details); err != nil {
			logs.Errorf("set file state for app bound template config items failed err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	existingPaths := []string{}
	for _, v := range details {
		if v.FileState != constant.FileStateDelete {
			existingPaths = append(existingPaths, path.Join(v.Path, v.Name))
		}
	}

	conflictPaths, err := s.compareNonTemplateConfigConflicts(kt, req.BizId, req.AppId, existingPaths)
	if err != nil {
		return nil, err
	}

	for _, v := range details {
		if v.FileState != constant.FileStateDelete {
			v.IsConflict = conflictPaths[path.Join(v.Path, v.Name)]
		}
	}

	// search by logic
	if req.SearchValue != "" {
		searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.TemplateRevision)
		if err != nil {
			return nil, err
		}
		fields := searcher.SearchFields()
		fieldsMap := make(map[string]bool)
		for _, f := range fields {
			fieldsMap[f] = true
		}
		fieldsMap["combinedPathName"] = true
		newDetails := make([]*pbatb.AppBoundTmplRevision, 0)
		for _, detail := range details {
			combinedPathName := path.Join(detail.Path, detail.Name)
			if (fieldsMap["revision_name"] && strings.Contains(detail.TemplateRevisionName, req.SearchValue)) ||
				(fieldsMap["revision_memo"] && strings.Contains(detail.TemplateRevisionMemo, req.SearchValue)) ||
				(fieldsMap["combinedPathName"] && strings.Contains(combinedPathName, req.SearchValue)) ||
				(fieldsMap["creator"] && strings.Contains(detail.Creator, req.SearchValue)) {
				newDetails = append(newDetails, detail)
			}
		}
		details = newDetails
	}

	// totalCnt is all data count
	totalCnt := uint32(len(details))

	if req.All {
		// return all data
		return &pbds.ListAppBoundTmplRevisionsResp{
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

	resp := &pbds.ListAppBoundTmplRevisionsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// 对比非模板配置冲突
func (s *Service) compareNonTemplateConfigConflicts(kt *kit.Kit, bizID, appID uint32,
	existingPaths []string) (map[string]bool, error) {

	configItemDetails, err := s.dao.ConfigItem().ListAllByAppID(kt, appID, bizID)
	if err != nil {
		logs.Errorf("list editing config items failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	var fileReleased []*table.ReleasedConfigItem
	fileReleased, err = s.dao.ReleasedCI().GetReleasedLately(kt, bizID, appID)
	if err != nil {
		logs.Errorf("get released failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var commits []*table.Commit
	commits, err = s.dao.Commit().ListAppLatestCommits(kt, bizID, appID)
	if err != nil {
		logs.Errorf("get commit, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// 过滤被删除的非模板配置
	configItems := pbrci.PbConfigItemState(configItemDetails, fileReleased, commits,
		[]string{constant.FileStateUnchange, constant.FileStateAdd, constant.FileStateRevise})

	for _, v := range configItems {
		if v.FileState != constant.FileStateDelete {
			existingPaths = append(existingPaths, path.Join(v.Spec.Path, v.Spec.Name))
		}
	}

	_, conflictPaths := checkExistingPathConflict(existingPaths)

	return conflictPaths, nil
}

// setFileState set file state for template config items.
func (s *Service) setFileState(kt *kit.Kit, unreleased []*pbatb.AppBoundTmplRevision) ([]*pbatb.AppBoundTmplRevision,
	error) {
	if len(unreleased) == 0 {
		return []*pbatb.AppBoundTmplRevision{}, nil
	}

	released, err := s.dao.ReleasedAppTemplate().GetReleasedLately(kt, kt.BizID, kt.AppID)
	if err != nil {
		logs.Errorf("get released app templates lately failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	releasedMap := make(map[uint32]*table.ReleasedAppTemplate, len(released))
	for _, r := range released {
		releasedMap[r.Spec.TemplateID] = r
	}

	for _, ci := range unreleased {
		if len(releasedMap) == 0 {
			ci.FileState = constant.FileStateAdd
			continue
		}

		if _, ok := releasedMap[ci.TemplateId]; ok {
			if ci.TemplateRevisionId == releasedMap[ci.TemplateId].Spec.TemplateRevisionID {
				ci.FileState = constant.FileStateUnchange
			} else {
				ci.FileState = constant.FileStateRevise
			}
			delete(releasedMap, ci.TemplateId)
			continue
		}

		ci.FileState = constant.FileStateAdd
	}

	result := unreleased
	if len(releasedMap) > 0 {
		releasedTmpls := make([]*table.ReleasedAppTemplate, 0)
		for _, r := range releasedMap {
			releasedTmpls = append(releasedTmpls, r)
		}
		deleted := pbatb.PbAppBoundTmplRevisionsFromReleased(releasedTmpls)
		for _, d := range deleted {
			d.FileState = constant.FileStateDelete
		}
		//nolint:gocritic
		result = append(unreleased, deleted...)

	}

	return result, nil
}

// ListReleasedAppBoundTmplRevisions list app bound template revisions.
func (s *Service) ListReleasedAppBoundTmplRevisions(ctx context.Context,
	req *pbds.ListReleasedAppBoundTmplRevisionsReq) (
	*pbds.ListReleasedAppBoundTmplRevisionsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// validate the page params
	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.ReleasedAppTemplate)
	if err != nil {
		return nil, err
	}

	details, count, err := s.dao.ReleasedAppTemplate().List(kt, req.BizId,
		req.AppId, req.ReleaseId, searcher, opt, req.SearchValue)
	if err != nil {
		logs.Errorf("list released app bound templates revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListReleasedAppBoundTmplRevisionsResp{
		Count:   uint32(count),
		Details: pbatb.PbReleasedAppBoundTmplRevisions(details),
	}
	return resp, nil
}

// GetReleasedAppBoundTmplRevision get app bound template revision.
func (s *Service) GetReleasedAppBoundTmplRevision(ctx context.Context,
	req *pbds.GetReleasedAppBoundTmplRevisionReq) (
	*pbds.GetReleasedAppBoundTmplRevisionResp, error) {
	kt := kit.FromGrpcContext(ctx)

	detail, err := s.dao.ReleasedAppTemplate().Get(kt, req.BizId, req.AppId, req.ReleaseId, req.TemplateRevisionId)
	if err != nil {
		logs.Errorf("get released app bound template revision failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.GetReleasedAppBoundTmplRevisionResp{
		Detail: pbatb.PbReleasedAppBoundTmplRevision(detail),
	}
	return resp, nil
}

// CheckAppTemplateBinding check conflicts of app template binding.
func (s *Service) CheckAppTemplateBinding(ctx context.Context, req *pbds.CheckAppTemplateBindingReq) (
	*pbds.CheckAppTemplateBindingResp, error) {
	kt := kit.FromGrpcContext(ctx)

	atb := &table.AppTemplateBinding{
		Spec:       req.Spec.AppTemplateBindingSpec(),
		Attachment: req.Attachment.AppTemplateBindingAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}

	conflicts, err := s.getConflictsOfATB(kt, atb)
	if err != nil {
		logs.Errorf("check app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &pbds.CheckAppTemplateBindingResp{
		Details: conflicts,
	}, nil
}

// getConflictsOfATB get conflicts of app template binding.
func (s *Service) getConflictsOfATB(kt *kit.Kit, atb *table.AppTemplateBinding) ([]*pbatb.Conflict, error) {
	pbs := parseBindings(atb.Spec.Bindings)

	if err := s.fillUnspecifiedTemplates(kt, pbs); err != nil {
		return nil, err
	}

	if repeated := tools.SliceRepeatedElements(pbs.TemplateIDs); len(repeated) > 0 {
		return s.getConflictDetailsOfATB(kt, pbs, repeated)
	}

	tmplRevisions, err := s.dao.TemplateRevision().ListByIDs(kt, pbs.TemplateRevisionIDs)
	if err != nil {
		return nil, err
	}

	duplicated, _, err := s.getDuplicatedCIs(kt, tmplRevisions)
	if err != nil {
		logs.Errorf("get duplicated config items failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return s.getConflictDetailsOfATB(kt, pbs, duplicated)
}

func (s *Service) getConflictDetailsOfATB(kt *kit.Kit, pbs *parsedBindings, tmplIDs []uint32) ([]*pbatb.Conflict,
	error) {
	if len(tmplIDs) == 0 {
		return []*pbatb.Conflict{}, nil
	}

	// get template set details
	tmplSets, err := s.dao.TemplateSet().ListByIDs(kt, pbs.TemplateSetIDs)
	if err != nil {
		logs.Errorf("list template set details by ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplSetMap := make(map[uint32]*table.TemplateSet, len(tmplSets))
	for _, t := range tmplSets {
		tmplSetMap[t.ID] = t
	}

	// get template details
	tmpls, err := s.dao.Template().ListByIDs(kt, tmplIDs)
	if err != nil {
		logs.Errorf("list template details by ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	tmplMap := make(map[uint32]*table.Template, len(tmpls))
	for _, t := range tmpls {
		tmplMap[t.ID] = t
	}

	conflicts := make([]*pbatb.Conflict, 0)
	for _, b := range pbs.TemplateBindings {
		for _, r := range b.TemplateRevisions {
			if _, ok := tmplMap[r.TemplateID]; ok {
				conflicts = append(conflicts, &pbatb.Conflict{
					TemplateSetId:   b.TemplateSetID,
					TemplateSetName: tmplSetMap[b.TemplateSetID].Spec.Name,
					TemplateId:      r.TemplateID,
					TemplateName:    tmplMap[r.TemplateID].Spec.Name,
				})
			}
		}
	}

	return conflicts, nil
}

// CascadeUpdateATB update app template binding in cascaded way.
// Only called by bscp system itself, no need to validate the input, but need the uniqueness verification.
/*
在模版/套餐有被服务引用的情况下，如下场景需要级联更新应用模版绑定数据：
1.对套餐添加/移出模板 （更新套餐接口、添加模版到套餐接口、从套餐移出模版接口）
2.删除套餐（删除套餐接口）
3.创建模版时指定了套餐（创建模版接口）
4.删除模版（删除模版接口、批量删除模版接口）
5.创建模版版本（创建模版版本接口）
6.删除模版版本（删除模版版本接口，暂不开放该接口）
*/
func (s *Service) CascadeUpdateATB(kt *kit.Kit, tx *gen.QueryTx, atb *table.AppTemplateBinding) error {
	if err := s.genFinalATBForCascade(kt, tx, atb); err != nil {
		logs.Errorf("cascade update app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err := s.dao.AppTemplateBinding().UpdateWithTx(kt, tx, atb); err != nil {
		logs.Errorf("cascade update app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	// validate config items count.
	if err := s.dao.ConfigItem().ValidateAppCINumber(kt, tx, atb.Attachment.BizID, atb.Attachment.AppID); err != nil {
		logs.Errorf("validate config items count failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// genFinalATBForCascade generate the final app template binding for cascade update operation.
// 因为服务是引用套餐下的所有模版，主要关注套餐变化，对于模版和模版版本相关操作，基于原有atb重新生成即可（包括更新latest版本）
func (s *Service) genFinalATBForCascade(kt *kit.Kit, tx *gen.QueryTx, atb *table.AppTemplateBinding) error {
	pbs, err := s.getPBSForCascade(kt, tx, atb.Spec.Bindings)
	if err != nil {
		return err
	}

	tmplRevisions, err := s.dao.TemplateRevision().ListByIDsWithTx(kt, tx, pbs.TemplateRevisionIDs)
	if err != nil {
		return err
	}

	if e := s.fillATBTmplSpace(kt, atb, tmplRevisions); e != nil {
		return e
	}
	s.fillATBModel(kt, atb, pbs)

	return nil
}

// getPBSForCascade get parsed bindings for cascade update operation.
func (s *Service) getPBSForCascade(kt *kit.Kit, tx *gen.QueryTx, bindings []*table.TemplateBinding) (*parsedBindings,
	error) {
	pbs := new(parsedBindings)
	if len(bindings) == 0 {
		return pbs, nil
	}

	// tmplSetID => [tmplID => tmplRevisionID]
	nonLatestRevisionMap := make(map[uint32]map[uint32]uint32)
	// tmplSetID => [tmplID => isLatest]
	latestTmplMap := make(map[uint32]map[uint32]bool)
	// tmplID => tmplRevisionID
	allTmplRevisionMap := make(map[uint32]uint32)
	for _, b := range bindings {
		pbs.TemplateSetIDs = append(pbs.TemplateSetIDs, b.TemplateSetID)
		nonLatestRevisionMap[b.TemplateSetID] = make(map[uint32]uint32)
		latestTmplMap[b.TemplateSetID] = make(map[uint32]bool)
		for _, r := range b.TemplateRevisions {
			if r.IsLatest {
				latestTmplMap[b.TemplateSetID][r.TemplateID] = true
			} else {
				nonLatestRevisionMap[b.TemplateSetID][r.TemplateID] = r.TemplateRevisionID
				allTmplRevisionMap[r.TemplateID] = r.TemplateRevisionID
			}
		}
	}

	// get all the templates of the template set
	templateSets, err := s.dao.TemplateSet().ListByIDsWithTx(kt, tx, pbs.TemplateSetIDs)
	if err != nil {
		logs.Errorf("list template set by ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// tmplSetID => [tmplID...]
	allTmplMap := make(map[uint32][]uint32)
	for _, ts := range templateSets {
		allTmplMap[ts.ID] = ts.Spec.TemplateIDs
		pbs.TemplateIDs = append(pbs.TemplateIDs, ts.Spec.TemplateIDs...)
		// get all latest template ids and all non latest template revision ids
		for _, id := range ts.Spec.TemplateIDs {
			if latestTmplMap[ts.ID][id] {
				pbs.LatestTemplateIDs = append(pbs.LatestTemplateIDs, id)
				continue
			}
			if _, ok := nonLatestRevisionMap[ts.ID][id]; ok {
				pbs.TemplateRevisionIDs = append(pbs.TemplateRevisionIDs, nonLatestRevisionMap[ts.ID][id])
			} else {
				pbs.LatestTemplateIDs = append(pbs.LatestTemplateIDs, id)
				latestTmplMap[ts.ID][id] = true
			}
		}
	}

	// get all latest revisions of latest templates
	latestTmplRevisions, err := s.dao.TemplateRevision().ListByTemplateIDsWithTx(kt, tx, kt.BizID,
		pbs.LatestTemplateIDs)
	if err != nil {
		logs.Errorf("list template revision names by template ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// template id => the latest template revision
	latestRevisionMap := getLatestTmplRevisions(latestTmplRevisions)

	for tID, r := range latestRevisionMap {
		pbs.TemplateRevisionIDs = append(pbs.TemplateRevisionIDs, r.ID)
		allTmplRevisionMap[tID] = r.ID
	}

	for tsID, tmpls := range allTmplMap {
		b := new(table.TemplateBinding)
		b.TemplateSetID = tsID
		for _, tID := range tmpls {
			b.TemplateRevisions = append(b.TemplateRevisions, &table.TemplateRevisionBinding{
				TemplateID:         tID,
				TemplateRevisionID: allTmplRevisionMap[tID],
				IsLatest:           latestTmplMap[tsID][tID],
			})
		}
		pbs.TemplateBindings = append(pbs.TemplateBindings, b)
	}

	return pbs, nil
}

// genFinalATB generate the final app template binding.
func (s *Service) genFinalATB(kt *kit.Kit, atb *table.AppTemplateBinding) error {
	pbs := parseBindings(atb.Spec.Bindings)

	if err := s.validateATBUpsert(kt, pbs); err != nil {
		return err
	}

	if err := s.fillUnspecifiedTemplates(kt, pbs); err != nil {
		return err
	}

	if err := s.dao.Validator().ValidateTmplRevisionsExist(kt, pbs.TemplateRevisionIDs); err != nil {
		return err
	}
	tmplRevisions, err := s.dao.TemplateRevision().ListByIDs(kt, pbs.TemplateRevisionIDs)
	if err != nil {
		return err
	}

	if e := s.fillATBTmplSpace(kt, atb, tmplRevisions); e != nil {
		return e
	}
	s.fillATBModel(kt, atb, pbs)

	return nil
}

// ValidateAppTemplateBindingUniqueKey validate the unique key name+path for an app.
// if the unique key name+path exists in table app_template_binding for the app, return error.
func (s *Service) ValidateAppTemplateBindingUniqueKey(kt *kit.Kit, bizID, appID uint32, name,
	dir string) error {
	opt := &types.BasePage{All: true}
	details, _, err := s.dao.AppTemplateBinding().List(kt, bizID, appID, opt)
	if err != nil {
		logs.Errorf("validate app template binding unique key failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// so far, no any template config item exists for the app
	if len(details) == 0 {
		return nil
	}

	templateRevisions, err := s.dao.TemplateRevision().ListByIDs(kt, details[0].Spec.TemplateRevisionIDs)
	if err != nil {
		logs.Errorf("validate app template binding unique key failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	for _, tr := range templateRevisions {
		if name == tr.Spec.Name && dir == tr.Spec.Path {
			return errf.Errorf(errf.InvalidRequest, i18n.T(kt,
				"the config file %s already exists in this space and cannot be created again", path.Join(dir, name)))
		}
	}

	return nil
}

// fillATBModel fill model AppTemplateBinding's template space ids field
func (s *Service) fillATBTmplSpace(_ *kit.Kit, g *table.AppTemplateBinding,
	tmplRevisions []*table.TemplateRevision) error {
	tmplSpaceIDs := make([]uint32, 0)
	for _, tr := range tmplRevisions {
		tmplSpaceIDs = append(tmplSpaceIDs, tr.Attachment.TemplateSpaceID)
	}
	g.Spec.TemplateSpaceIDs = tools.RemoveDuplicates(tmplSpaceIDs)
	return nil
}

// fillATBModel fill model AppTemplateBinding's fields
func (s *Service) fillATBModel(_ *kit.Kit, g *table.AppTemplateBinding, pbs *parsedBindings) {
	g.Spec.TemplateSetIDs = pbs.TemplateSetIDs
	g.Spec.TemplateRevisionIDs = pbs.TemplateRevisionIDs
	g.Spec.LatestTemplateIDs = pbs.LatestTemplateIDs
	g.Spec.TemplateIDs = pbs.TemplateIDs
	g.Spec.Bindings = pbs.TemplateBindings

	// set for empty slice to ensure the data in db is not `null` but `[]`
	if len(g.Spec.TemplateSetIDs) == 0 {
		g.Spec.TemplateSetIDs = []uint32{}
	}
	if len(g.Spec.TemplateRevisionIDs) == 0 {
		g.Spec.TemplateRevisionIDs = []uint32{}
	}
	if len(g.Spec.LatestTemplateIDs) == 0 {
		g.Spec.LatestTemplateIDs = []uint32{}
	}
	if len(g.Spec.TemplateIDs) == 0 {
		g.Spec.TemplateIDs = []uint32{}
	}
	if len(g.Spec.Bindings) == 0 {
		g.Spec.Bindings = []*table.TemplateBinding{}
	}
}

// parseBindings parse the input into the target object
// no need to validate the bindings here, it is already done in config server
func parseBindings(bindings []*table.TemplateBinding) *parsedBindings {
	pbs := new(parsedBindings)
	for _, b := range bindings {
		b2 := &table.TemplateBinding{
			TemplateSetID:     b.TemplateSetID,
			TemplateRevisions: make([]*table.TemplateRevisionBinding, 0, len(b.TemplateRevisions)),
		}

		pbs.TemplateSetIDs = append(pbs.TemplateSetIDs, b.TemplateSetID)
		for _, r := range b.TemplateRevisions {
			pbs.TemplateRevisionIDs = append(pbs.TemplateRevisionIDs, r.TemplateRevisionID)
			pbs.TemplateIDs = append(pbs.TemplateIDs, r.TemplateID)
			if r.IsLatest {
				pbs.LatestTemplateIDs = append(pbs.LatestTemplateIDs, r.TemplateID)
				pbs.LatestTemplateRevisionIDs = append(pbs.LatestTemplateRevisionIDs, r.TemplateRevisionID)
			}

			b2.TemplateRevisions = append(b2.TemplateRevisions, &table.TemplateRevisionBinding{
				TemplateID:         r.TemplateID,
				TemplateRevisionID: r.TemplateRevisionID,
				IsLatest:           r.IsLatest,
			})
		}

		pbs.TemplateBindings = append(pbs.TemplateBindings, b2)
	}

	return pbs
}

// fillUnspecifiedTemplates update the pbs's unspecified templates and revisions
func (s *Service) fillUnspecifiedTemplates(kt *kit.Kit, pbs *parsedBindings) error {
	for i := range pbs.TemplateBindings {
		b := pbs.TemplateBindings[i]
		var templateIDs []uint32
		for _, r := range b.TemplateRevisions {
			templateIDs = append(templateIDs, r.TemplateID)
		}

		if err := s.dao.Validator().ValidateTmplsBelongToTmplSet(kt, templateIDs, b.TemplateSetID); err != nil {
			return err
		}

		// get all the templates belong to the template set, then get the unspecified templates
		templateSets, err := s.dao.TemplateSet().ListByIDs(kt, []uint32{b.TemplateSetID})
		if err != nil {
			logs.Errorf("fill unspecified templates failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		unspecified := tools.SliceDiff(templateSets[0].Spec.TemplateIDs, templateIDs)

		// get all latest revisions and update pbs's unspecified templates
		if len(unspecified) > 0 {
			pbs.TemplateIDs = append(pbs.TemplateIDs, unspecified...)

			templateRevisions, err := s.ListTmplRevisionNamesByTmplIDs(
				kt.Ctx,
				&pbds.ListTmplRevisionNamesByTmplIDsReq{
					BizId:       kt.BizID,
					TemplateIds: unspecified,
				})
			if err != nil {
				logs.Errorf("fill unspecified templates failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}

			// for unspecified template, use the latest revision and set is_latest true
			for _, t := range templateRevisions.Details {
				b.TemplateRevisions = append(b.TemplateRevisions, &table.TemplateRevisionBinding{
					TemplateID:         t.TemplateId,
					TemplateRevisionID: t.LatestTemplateRevisionId,
					IsLatest:           true,
				})

				pbs.TemplateRevisionIDs = append(pbs.TemplateRevisionIDs, t.LatestTemplateRevisionId)
				pbs.LatestTemplateIDs = append(pbs.LatestTemplateIDs, t.TemplateId)
			}
		}
	}

	return nil
}

// parsedBindings is parsed bindings which suits to save in db
type parsedBindings struct {
	TemplateIDs               []uint32
	TemplateSetIDs            []uint32
	TemplateRevisionIDs       []uint32
	LatestTemplateIDs         []uint32
	LatestTemplateRevisionIDs []uint32
	TemplateBindings          []*table.TemplateBinding
}

// validateUpsert validate for create or update operation of app template binding
func (s *Service) validateATBUpsert(kt *kit.Kit, b *parsedBindings) error {
	if err := s.dao.Validator().ValidateTmplSetsExist(kt, b.TemplateSetIDs); err != nil {
		return err
	}

	if err := s.validateATBLatestRevisions(kt, b); err != nil {
		return err
	}

	return nil
}

// validateATBLatestRevisions validate whether the latest revisions specified by user is latest
func (s *Service) validateATBLatestRevisions(kt *kit.Kit, b *parsedBindings) error {
	if len(b.TemplateIDs) == 0 {
		return nil
	}

	// the method will validate whether template ids exist as well
	templateRevisions, err := s.ListTmplRevisionNamesByTmplIDs(
		kt.Ctx,
		&pbds.ListTmplRevisionNamesByTmplIDsReq{
			BizId:       kt.BizID,
			TemplateIds: b.TemplateIDs,
		})
	if err != nil {
		logs.Errorf("validate the latest template revision failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	latestMap := make(map[uint32]bool, len(templateRevisions.Details))
	for _, t := range templateRevisions.Details {
		latestMap[t.LatestTemplateRevisionId] = true
	}

	// validate whether the latest revision specified by user is latest
	nonLatest := make([]uint32, 0)
	for _, id := range b.LatestTemplateRevisionIDs {
		if !latestMap[id] {
			nonLatest = append(nonLatest, id)

		}
	}

	if len(nonLatest) > 0 {
		return fmt.Errorf("template revision id in %v is not the latest revision, please confirm it carefully, "+
			"refresh the page to get the latest revision if you are using with browser", nonLatest)
	}

	return nil
}

// getDuplicatedCIs get duplicated config items whose unique keys `name+path` are same
func (s *Service) getDuplicatedCIs(kt *kit.Kit, tmplRevisions []*table.TemplateRevision) (tmplIDs []uint32,
	uKeys []types.CIUniqueKey, err error) {
	var addKeys []types.CIUniqueKey
	tmplMap := make(map[types.CIUniqueKey][]uint32)
	for _, tr := range tmplRevisions {
		k := types.CIUniqueKey{
			Name: tr.Spec.Name,
			Path: tr.Spec.Path,
		}
		addKeys = append(addKeys, k)
		tmplMap[k] = append(tmplMap[k], tr.Attachment.TemplateID)
	}

	existKeys, err := s.dao.ConfigItem().GetUniqueKeys(kt, kt.BizID, kt.AppID)
	if err != nil {
		logs.Errorf("get config items unique keys failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, err
	}

	//nolint:gocritic
	allKeys := append(existKeys, addKeys...)
	uKeys = findRepeatedElements(allKeys)
	for _, k := range uKeys {
		tmplIDs = append(tmplIDs, tmplMap[k]...)
	}
	tmplIDs = tools.RemoveDuplicates(tmplIDs)

	return tmplIDs, uKeys, nil
}

func findRepeatedElements(slice []types.CIUniqueKey) []types.CIUniqueKey {
	frequencyMap := make(map[types.CIUniqueKey]int)
	var repeatedElements []types.CIUniqueKey

	// Count the frequency of each uID in the slice
	for _, key := range slice {
		frequencyMap[key]++
	}

	// Check if any uID appears more than once
	for key, count := range frequencyMap {
		if count > 1 {
			repeatedElements = append(repeatedElements, key)
		}
	}

	return repeatedElements
}

// ImportFromTemplateSetToApp 从配置模板导入到服务
func (s *Service) ImportFromTemplateSetToApp(ctx context.Context, req *pbds.ImportFromTemplateSetToAppReq) (
	*pbbase.EmptyResp, error) {
	kit := kit.FromGrpcContext(ctx)
	templateIds, templateRevisionIds, latestTemplateIds, templateSetIds, templateSpaceIds :=
		[]uint32{}, []uint32{}, []uint32{}, []uint32{}, []uint32{}
	validatedTemplateSetNames := map[uint32]string{}
	validatedTemplateSpaceNames := map[uint32]string{}
	bindings := make([]*table.TemplateBinding, 0)
	for _, binding := range req.GetBindings() {
		templateSetIds = append(templateSetIds, binding.GetTemplateSetId())
		templateSpaceIds = append(templateSpaceIds, binding.GetTemplateSpaceId())
		validatedTemplateSetNames[binding.GetTemplateSetId()] = binding.GetTemplateSetName()
		validatedTemplateSpaceNames[binding.GetTemplateSpaceId()] = binding.GetTemplateSpaceName()
		revisions := make([]*table.TemplateRevisionBinding, 0)
		for _, v := range binding.GetTemplateRevisions() {
			templateIds = append(templateIds, v.TemplateId)
			templateRevisionIds = append(templateRevisionIds, v.TemplateRevisionId)
			if v.IsLatest {
				latestTemplateIds = append(latestTemplateIds, v.TemplateId)
			}
			revisions = append(revisions, &table.TemplateRevisionBinding{
				TemplateID:         v.TemplateId,
				TemplateRevisionID: v.TemplateRevisionId,
				IsLatest:           v.IsLatest,
			})
		}
		bindings = append(bindings, &table.TemplateBinding{
			TemplateSetID:     binding.TemplateSetId,
			TemplateRevisions: revisions,
		})
	}

	templateSets, err := s.dao.TemplateSet().ListByIDs(kit, templateSetIds)
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kit, "list template sets by template set ids failed, err: %v", err))
	}

	// 1. 验证模板空间
	if err := s.verifyTemplateSpaces(kit, templateSpaceIds, validatedTemplateSpaceNames); err != nil {
		return nil, err
	}

	// 2. 验证模板套餐
	if err := s.verifyTemplateSets(kit, templateSets, validatedTemplateSetNames); err != nil {
		return nil, err
	}

	// 3. 验证模板文件和模板版本数据
	if err := s.verifyTemplateSetAndRevisions(kit, validatedTemplateSetNames, req.GetBindings()); err != nil {
		return nil, err
	}

	// 4. 验证是否超出服务限制
	if err := s.verifyTemplateSetBoundTemplatesNumber(kit, templateSets, req); err != nil {
		return nil, err
	}

	tx := s.dao.GenQuery().Begin()
	appTemplateBinding := &table.AppTemplateBinding{
		Spec: &table.AppTemplateBindingSpec{
			TemplateSpaceIDs:    tools.RemoveDuplicates(templateSpaceIds),
			TemplateSetIDs:      tools.RemoveDuplicates(templateSetIds),
			TemplateIDs:         tools.RemoveDuplicates(templateIds),
			TemplateRevisionIDs: tools.RemoveDuplicates(templateRevisionIds),
			LatestTemplateIDs:   tools.RemoveDuplicates(latestTemplateIds),
			Bindings:            bindings,
		},
		Attachment: &table.AppTemplateBindingAttachment{
			BizID: req.BizId,
			AppID: req.AppId,
		},
		Revision: &table.Revision{
			Creator:   kit.User,
			Reviser:   kit.User,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	if err := s.dao.AppTemplateBinding().UpsertWithTx(kit, tx, appTemplateBinding); err != nil {
		logs.Errorf("update app template binding failed, err: %v, rid: %s", err, kit.Rid)
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "update app template binding failed, err: %v", err))
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	return &pbbase.EmptyResp{}, nil
}

// 验证模板空间
func (s *Service) verifyTemplateSpaces(kit *kit.Kit, templateSpaceIds []uint32,
	templateSpaceNames map[uint32]string) error {

	templateSpaces, err := s.dao.TemplateSpace().ListByIDs(kit, templateSpaceIds)
	if err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "list template spaces failed, err: %v", err))
	}

	existingTemplateSpaces := map[uint32]bool{}
	for _, v := range templateSpaces {
		existingTemplateSpaces[v.ID] = true
	}

	for k, v := range templateSpaceNames {
		if !existingTemplateSpaces[k] {
			return errf.Errorf(errf.NotFound, i18n.T(kit, "template space %s not found", v))
		}
	}

	return nil
}

// 验证模板套餐
func (s *Service) verifyTemplateSets(kit *kit.Kit, templateSets []*table.TemplateSet,
	validatedTemplateSetNames map[uint32]string) error {

	existsTemplateSet := map[uint32]bool{}
	for _, v := range templateSets {
		existsTemplateSet[v.ID] = true
	}

	for k, v := range validatedTemplateSetNames {
		if !existsTemplateSet[k] {
			return errf.Errorf(errf.NotFound, i18n.T(kit, "template set %s not found", v))
		}
	}

	return nil
}

// 验证模板配置是否存在
// 验证模板版本是否存在
// 验证待提交latest版本是不是latest的
func (s *Service) verifyTemplateSetAndRevisions(kit *kit.Kit, validatedTemplateSetNames map[uint32]string,
	bindings []*pbds.ImportFromTemplateSetToAppReq_Binding) error {

	// 模板套餐下的模板配置id 和 模板配置下的模板版本id
	templateSetAndTemplateIds, templateAndRevisionIds := map[uint32][]uint32{}, map[uint32][]uint32{}
	// 模板配置名称 和 模板版本名称
	templateNames, revisionNames := map[uint32]string{}, map[uint32]string{}
	// 待验证的latest版本
	validatedLatest := map[uint32]uint32{}
	latestTemplateIds, templateIds, templateRevisionIds := []uint32{}, []uint32{}, []uint32{}
	for _, binding := range bindings {
		for _, v := range binding.GetTemplateRevisions() {
			templateIds = append(templateIds, v.TemplateId)
			templateRevisionIds = append(templateRevisionIds, v.TemplateRevisionId)
			templateSetAndTemplateIds[binding.TemplateSetId] = append(templateSetAndTemplateIds[binding.TemplateSetId],
				v.TemplateId)
			templateAndRevisionIds[v.TemplateId] = append(templateAndRevisionIds[v.TemplateId],
				v.TemplateRevisionId)
			templateNames[v.TemplateId] = v.TemplateName
			revisionNames[v.TemplateRevisionId] = v.TemplateRevisionName
			if v.IsLatest {
				latestTemplateIds = append(latestTemplateIds, v.TemplateId)
				validatedLatest[v.TemplateId] = v.TemplateRevisionId
			}
		}
	}

	// 获取指定的模板配置
	existsTemplateIds := map[uint32]bool{}
	templates, err := s.dao.Template().ListByIDs(kit, templateIds)
	if err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "list templates of template set failed, err: %v", err))
	}
	for _, v := range templates {
		existsTemplateIds[v.ID] = true
	}

	for sId, tId := range templateSetAndTemplateIds {
		for _, v := range tId {
			if !existsTemplateIds[v] {
				return errors.New(i18n.T(kit, `the template file %s in the template set 
				%s has been removed. Please import the set again`,
					validatedTemplateSetNames[sId], templateNames[v]))
			}
		}
	}

	// 获取指定的模板版本
	existsRevisionsIds := map[uint32]bool{}
	revisions, err := s.dao.TemplateRevision().ListByIDs(kit, templateRevisionIds)
	if err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "list template revisions failed, err: %v", err))
	}
	for _, v := range revisions {
		existsRevisionsIds[v.ID] = true
	}

	for tId, rId := range templateAndRevisionIds {
		for _, v := range rId {
			if !existsRevisionsIds[v] {
				return errors.New(i18n.T(kit, `template version %s in template file %s 
				has been removed. Please import the set again`,
					revisionNames[v], templateNames[tId]))
			}
		}
	}

	// 获取指定模板最新的模板版本
	latest, err := s.dao.TemplateRevision().ListLatestRevisionsGroupByTemplateIds(kit, latestTemplateIds)
	if err != nil {
		return err
	}
	for _, v := range latest {
		if rid, ok := validatedLatest[v.Attachment.TemplateID]; ok && rid == v.ID {
			continue
		}
		return errors.New(i18n.T(kit, `the version number %s in the template file %s is not the 
		latest version. Please import the set again`,
			templateNames[v.Attachment.TemplateID], revisionNames[v.ID]))
	}

	return nil
}

// 模板套餐导入服务时验证是否超出服务限制
// nolint:goconst
func (s *Service) verifyTemplateSetBoundTemplatesNumber(kt *kit.Kit, templateSets []*table.TemplateSet,
	req *pbds.ImportFromTemplateSetToAppReq) error {

	var templateCount int
	for _, v := range templateSets {
		templateCount += len(v.Spec.TemplateIDs)
	}

	// 1. 获取未命名版本配置文件数量
	configItemCount, err := s.dao.ConfigItem().GetConfigItemCount(kt, req.BizId, req.AppId)
	if err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kt, "count the number of service configurations failed, err: %s", err))
	}

	// 获取app信息
	app, err := s.dao.App().GetByID(kt, req.AppId)
	if err != nil {
		return errf.Errorf(errf.AppNotExists, i18n.T(kt, "app %d not found", req.AppId))
	}

	// 只有文件类型的才能绑定套餐
	if app.Spec.ConfigType != table.File {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kt, "app %s is not file type", app.Spec.Name))
	}

	appConfigCnt := getAppConfigCnt(req.BizId)

	if int(configItemCount)+templateCount > appConfigCnt {
		return errf.New(errf.InvalidParameter,
			i18n.T(kt, "the total number of app %s config items(including template and non-template)"+
				"exceeded the limit %d", app.Spec.Name, appConfigCnt))
	}

	return nil
}
