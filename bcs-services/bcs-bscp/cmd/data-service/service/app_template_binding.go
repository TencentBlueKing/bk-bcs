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
	"encoding/json"
	"fmt"
	"strings"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbatb "bscp.io/pkg/protocol/core/app-template-binding"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/search"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
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

	id, err := s.dao.AppTemplateBinding().Create(kt, appTemplateBinding)
	if err != nil {
		logs.Errorf("create app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListAppTemplateBindings list app template binding.
func (s *Service) ListAppTemplateBindings(ctx context.Context, req *pbds.ListAppTemplateBindingsReq) (
	*pbds.ListAppTemplateBindingsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
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
func (s *Service) UpdateAppTemplateBinding(ctx context.Context, req *pbds.UpdateAppTemplateBindingReq) (*pbbase.
	EmptyResp,
	error) {
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

	if err := s.dao.AppTemplateBinding().Update(kt, appTemplateBinding); err != nil {
		logs.Errorf("update app template binding failed, err: %v, rid: %s", err, kt.Rid)
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

// ListAppBoundTemplateRevisions list app bound template revisions.
func (s *Service) ListAppBoundTemplateRevisions(ctx context.Context, req *pbds.ListAppBoundTemplateRevisionsReq) (
	*pbds.ListAppBoundTemplateRevisionsResp, error) {
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
		return &pbds.ListAppBoundTemplateRevisionsResp{
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

	/// combine resp details
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
				ByteSize:             d.Spec.ContentSpec.ByteSize,
				Creator:              d.Revision.Creator,
				CreateAt:             d.Revision.CreatedAt.Format(constant.TimeStdFormat),
			})
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
		newDetails := make([]*pbatb.AppBoundTmplRevision, 0)
		for _, detail := range details {
			if (fieldsMap["revision_name"] && strings.Contains(detail.TemplateRevisionName, req.SearchValue)) ||
				(fieldsMap["revision_memo"] && strings.Contains(detail.TemplateRevisionMemo, req.SearchValue)) ||
				(fieldsMap["name"] && strings.Contains(detail.Name, req.SearchValue)) ||
				(fieldsMap["path"] && strings.Contains(detail.Path, req.SearchValue)) ||
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
		return &pbds.ListAppBoundTemplateRevisionsResp{
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

	resp := &pbds.ListAppBoundTemplateRevisionsResp{
		Count:   totalCnt,
		Details: details,
	}
	return resp, nil
}

// ListReleasedAppBoundTemplateRevisions list app bound template revisions.
func (s *Service) ListReleasedAppBoundTemplateRevisions(ctx context.Context,
	req *pbds.ListReleasedAppBoundTemplateRevisionsReq) (
	*pbds.ListReleasedAppBoundTemplateRevisionsResp, error) {
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

	details, count, err := s.dao.ReleasedAppTemplate().List(kt, req.BizId, req.AppId, req.ReleaseId, searcher, opt)
	if err != nil {
		logs.Errorf("list template spaces failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListReleasedAppBoundTemplateRevisionsResp{
		Count:   uint32(count),
		Details: pbatb.PbAppBoundTmplRevisions(details),
	}
	return resp, nil
}

// CascadeUpdateATB update app template binding in cascaded way.
// Only called by bscp system itself, no need to validate the input, but need the uniqueness verification.
/*
在模版/套餐有被服务引用的情况下，如下场景需要级联更新应用模版绑定数据：
1.对套餐添加/移出模板 （更新套餐接口、添加模版到套餐接口、从套餐移出模版接口）
2.删除套餐（删除套餐接口）
3.删除模版（删除模版接口、批量删除模版接口）
4.创建模版版本（创建模版版本接口）
5.删除模版版本（删除模版版本接口，暂不开放该接口）
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

	if err := s.validateATBUniqueKey(kt, tmplRevisions, atb.Attachment.BizID, atb.Attachment.AppID); err != nil {
		return err
	}

	s.fillATBTmplSpace(kt, atb, tmplRevisions)
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
				// only append non latest template revisions at beginning
				pbs.TemplateRevisionIDs = append(pbs.TemplateRevisionIDs, r.TemplateRevisionID)
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
		// get all latest template ids
		for _, id := range ts.Spec.TemplateIDs {
			if latestTmplMap[ts.ID][id] {
				pbs.LatestTemplateIDs = append(pbs.LatestTemplateIDs, id)
				continue
			}
			if _, ok := nonLatestRevisionMap[ts.ID][id]; !ok {
				pbs.LatestTemplateIDs = append(pbs.LatestTemplateIDs, id)
				latestTmplMap[ts.ID][id] = true
			}
		}
	}
	if err := s.validateTmplForATBWithTx(kt, tx, pbs.TemplateIDs); err != nil {
		logs.Errorf("validate template for app template binding failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// get all latest revisions of latest templates
	latestTmplRevisions, err := s.dao.TemplateRevision().ListByTemplateIDsWithTx(kt, tx, kt.BizID,
		pbs.LatestTemplateIDs)
	if err != nil {
		logs.Errorf("list template revision names by template ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	// template id => the latest template revision id
	latestRevisionMap := getLatestTmplRevisions(latestTmplRevisions)

	for tID, rID := range latestRevisionMap {
		pbs.TemplateRevisionIDs = append(pbs.TemplateRevisionIDs, rID)
		allTmplRevisionMap[tID] = rID
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

// validateTmplForATB validate template to avoid same templates are bound to one app
func (s *Service) validateTmplForATB(kt *kit.Kit, tmplIDs []uint32) error {
	if len(tmplIDs) == 0 {
		return nil
	}

	if repeated := tools.SliceRepeatedElements(tmplIDs); len(repeated) > 0 {
		// get template details
		tmpls, err := s.dao.Template().ListByIDs(kt, repeated)
		if err != nil {
			logs.Errorf("list template by ids failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		type tmplT struct {
			ID   uint32 `json:"id"`
			Name string `json:"name"`
			Path string `json:"path"`
		}
		details := make([]tmplT, len(tmpls))
		for idx, t := range tmpls {
			details[idx] = tmplT{
				ID:   t.ID,
				Name: t.Spec.Name,
				Path: t.Spec.Path,
			}
		}
		detailsJs, err := json.Marshal(details)
		if err != nil {
			logs.Errorf("marshal template details failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		return fmt.Errorf("same template id in %v can't be bound to the same app, template details: %s",
			repeated, detailsJs)
	}

	return nil
}

// validateTmplForATBWithTx validate template with transaction to avoid same templates are bound to one app
func (s *Service) validateTmplForATBWithTx(kt *kit.Kit, tx *gen.QueryTx, tmplIDs []uint32) error {
	if len(tmplIDs) == 0 {
		return nil
	}

	if repeated := tools.SliceRepeatedElements(tmplIDs); len(repeated) > 0 {
		// get template details
		tmpls, err := s.dao.Template().ListByIDsWithTx(kt, tx, repeated)
		if err != nil {
			logs.Errorf("list template by ids failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		type tmplT struct {
			ID   uint32 `json:"id"`
			Name string `json:"name"`
			Path string `json:"path"`
		}
		details := make([]tmplT, len(tmpls))
		for idx, t := range tmpls {
			details[idx] = tmplT{
				ID:   t.ID,
				Name: t.Spec.Name,
				Path: t.Spec.Path,
			}
		}
		detailsJs, err := json.Marshal(details)
		if err != nil {
			logs.Errorf("marshal template details failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		return fmt.Errorf("same template id in %v can't be bound to the same app, template details: %s",
			repeated, detailsJs)
	}

	return nil
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

	if err := s.validateTmplForATB(kt, pbs.TemplateIDs); err != nil {
		return err
	}

	if err := s.dao.Validator().ValidateTemplateRevisionsExist(kt, pbs.TemplateRevisionIDs); err != nil {
		return err
	}
	tmplRevisions, err := s.dao.TemplateRevision().ListByIDs(kt, pbs.TemplateRevisionIDs)
	if err != nil {
		return err
	}

	if err := s.validateATBUniqueKey(kt, tmplRevisions, atb.Attachment.BizID, atb.Attachment.AppID); err != nil {
		return err
	}

	s.fillATBTmplSpace(kt, atb, tmplRevisions)
	s.fillATBModel(kt, atb, pbs)

	return nil
}

// ValidateAppTemplateBindingUniqueKey validate the unique key name+path for an app.
// if the unique key name+path exists in table app_template_binding for the app, return error.
func (s *Service) ValidateAppTemplateBindingUniqueKey(kt *kit.Kit, bizID, appID uint32, name,
	path string) error {
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
		if name == tr.Spec.Name && path == tr.Spec.Path {
			return fmt.Errorf("config item's same name %s and path %s already exists", name, path)
		}
	}

	return nil
}

// fillATBModel fill model AppTemplateBinding's template space ids field
func (s *Service) fillATBTmplSpace(kit *kit.Kit, g *table.AppTemplateBinding,
	tmplRevisions []*table.TemplateRevision) error {
	tmplSpaceMap := make(map[uint32]struct{})
	for _, tr := range tmplRevisions {
		tmplSpaceMap[tr.Attachment.TemplateSpaceID] = struct{}{}
	}
	g.Spec.TemplateSpaceIDs = convertToSlice(tmplSpaceMap)

	return nil
}

// fillATBModel fill model AppTemplateBinding's fields
func (s *Service) fillATBModel(kit *kit.Kit, g *table.AppTemplateBinding, pbs *parsedBindings) {
	g.Spec.TemplateSetIDs = pbs.TemplateSetIDs
	g.Spec.TemplateRevisionIDs = pbs.TemplateRevisionIDs
	g.Spec.LatestTemplateIDs = pbs.LatestTemplateIDs
	g.Spec.TemplateIDs = pbs.TemplateIDs
	g.Spec.Bindings = pbs.TemplateBindings
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
func (s *Service) fillUnspecifiedTemplates(kit *kit.Kit, pbs *parsedBindings) error {
	for i := range pbs.TemplateBindings {
		b := pbs.TemplateBindings[i]
		var templateIDs []uint32
		for _, r := range b.TemplateRevisions {
			templateIDs = append(templateIDs, r.TemplateID)
		}

		if err := s.dao.Validator().ValidateTemplatesBelongToTemplateSet(kit, templateIDs, b.TemplateSetID); err != nil {
			return err
		}

		// get all the templates belong to the template set, then get the unspecified templates
		templateSets, err := s.dao.TemplateSet().ListByIDs(kit, []uint32{b.TemplateSetID})
		if err != nil {
			logs.Errorf("fill unspecified templates failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
		unspecified := tools.SliceDiff(templateSets[0].Spec.TemplateIDs, templateIDs)

		// get all latest revisions and update pbs's unspecified templates
		if len(unspecified) > 0 {
			pbs.TemplateIDs = append(pbs.TemplateIDs, unspecified...)

			templateRevisions, err := s.ListTemplateRevisionNamesByTemplateIDs(
				kit.Ctx,
				&pbds.ListTemplateRevisionNamesByTemplateIDsReq{
					BizId:       kit.BizID,
					TemplateIds: unspecified,
				})
			if err != nil {
				logs.Errorf("fill unspecified templates failed, err: %v, rid: %s", err, kit.Rid)
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

func convertToSlice(m map[uint32]struct{}) []uint32 {
	var keys []uint32
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// validateUpsert validate for create or update operation of app template binding
func (s *Service) validateATBUpsert(kit *kit.Kit, b *parsedBindings) error {
	if err := s.dao.Validator().ValidateTemplateSetsExist(kit, b.TemplateSetIDs); err != nil {
		return err
	}

	if err := s.validateATBLatestRevisions(kit, b); err != nil {
		return err
	}

	return nil
}

// validateATBLatestRevisions validate whether the latest revisions specified by user is latest
func (s *Service) validateATBLatestRevisions(kit *kit.Kit, b *parsedBindings) error {
	if len(b.TemplateIDs) == 0 {
		return nil
	}

	// the method will validate whether template ids exist as well
	templateRevisions, err := s.ListTemplateRevisionNamesByTemplateIDs(
		kit.Ctx,
		&pbds.ListTemplateRevisionNamesByTemplateIDsReq{
			BizId:       kit.BizID,
			TemplateIds: b.TemplateIDs,
		})
	if err != nil {
		logs.Errorf("validate the latest template revision failed, err: %v, rid: %s", err, kit.Rid)
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

// validateATBUniqueKey validate unique key for app template binding
func (s *Service) validateATBUniqueKey(
	kit *kit.Kit, tmplRevisions []*table.TemplateRevision, bizID, appID uint32) error {
	// validates unique key name+path both in table app_template_bindings and config_items
	// validate the input is equivalent to validate in table app_template_bindings
	if err := validateUniqueKeyOfInput(tmplRevisions); err != nil {
		return err
	}
	// validate in table config_items
	for _, tr := range tmplRevisions {
		if _, err := s.dao.ConfigItem().GetByUniqueKey(kit, bizID, appID, tr.Spec.Name, tr.Spec.Path); err == nil {
			return fmt.Errorf("config item's same name %s and path %s already exists", tr.Spec.Name, tr.Spec.Path)
		}
	}

	return nil
}

// validateUniqueKeyOfInput validates unique key which is name+path of input only
func validateUniqueKeyOfInput(tmplRevisions []*table.TemplateRevision) error {
	var uids []uid
	for _, tr := range tmplRevisions {
		uids = append(uids, uid{
			Name: tr.Spec.Name,
			Path: tr.Spec.Path,
		})
	}
	repeated := findRepeatedElements(uids)
	if len(repeated) > 0 {
		js, _ := json.Marshal(repeated)
		return fmt.Errorf("config item's name and path must be unique, these are repeated: %s", js)
	}

	return nil
}

// validateUniqueKeyForApp validates unique key which is name+path for an app
func validateUniqueKeyForApp(tmplRevisions []*table.TemplateRevision, name, path string) error {
	for _, tr := range tmplRevisions {
		if name == tr.Spec.Name && path == tr.Spec.Path {
			return fmt.Errorf("config item's same name %s and path %s already exists", name, path)
		}
	}

	return nil
}

type uid struct {
	Name string
	Path string
}

func findRepeatedElements(slice []uid) []uid {
	frequencyMap := make(map[uid]int)
	var repeatedElements []uid

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
