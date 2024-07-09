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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbatb "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app-template-binding"
	pbtset "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-set"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/natsort"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// CreateAppTemplateBinding create an app template binding
func (s *Service) CreateAppTemplateBinding(ctx context.Context, req *pbcs.CreateAppTemplateBindingReq) (
	*pbcs.CreateAppTemplateBindingResp, error) {
	// FromGrpcContext used only to obtain Kit through grpc context.
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	templateSetIDs, _, err := parseBindings(req.Bindings)
	if err != nil {
		logs.Errorf("create app template binding failed, parse bindings err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	// SliceRepeatedElements get the repeated elements in a slice, and the keep the sequence of result
	repeatedTmplSetIDs := tools.SliceRepeatedElements(templateSetIDs)
	if len(repeatedTmplSetIDs) > 0 {
		return nil, fmt.Errorf("repeated template set ids: %v, id must be unique", repeatedTmplSetIDs)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	// Authorize authorize if user has permission to the resources.
	// If user is unauthorized, assign apply url and resources into error.
	if err = s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.CreateAppTemplateBindingReq{
		Attachment: &pbatb.AppTemplateBindingAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbatb.AppTemplateBindingSpec{
			Bindings: req.Bindings,
		},
	}
	var rp *pbds.CreateResp
	rp, err = s.client.DS.CreateAppTemplateBinding(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create app template binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateAppTemplateBindingResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteAppTemplateBinding delete an app template binding
func (s *Service) DeleteAppTemplateBinding(ctx context.Context, req *pbcs.DeleteAppTemplateBindingReq) (
	*pbcs.DeleteAppTemplateBindingResp, error) {
	// FromGrpcContext used only to obtain Kit through grpc context.
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.DeleteAppTemplateBindingReq{
		Id: req.BindingId,
		Attachment: &pbatb.AppTemplateBindingAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	if _, err := s.client.DS.DeleteAppTemplateBinding(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete app template binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.DeleteAppTemplateBindingResp{}, nil
}

// UpdateAppTemplateBinding update an app template binding
func (s *Service) UpdateAppTemplateBinding(ctx context.Context, req *pbcs.UpdateAppTemplateBindingReq) (
	*pbcs.UpdateAppTemplateBindingResp, error) {
	// FromGrpcContext used only to obtain Kit through grpc context.
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	templateSetIDs, _, err := parseBindings(req.Bindings)
	if err != nil {
		logs.Errorf("update app template binding failed, parse bindings err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	// SliceRepeatedElements get the repeated elements in a slice, and the keep the sequence of result
	repeatedTmplSetIDs := tools.SliceRepeatedElements(templateSetIDs)
	if len(repeatedTmplSetIDs) > 0 {
		return nil, fmt.Errorf("repeated template set ids: %v, id must be unique", repeatedTmplSetIDs)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err = s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.UpdateAppTemplateBindingReq{
		Id: req.BindingId,
		Attachment: &pbatb.AppTemplateBindingAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbatb.AppTemplateBindingSpec{
			Bindings: req.Bindings,
		},
	}
	if _, err = s.client.DS.UpdateAppTemplateBinding(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update app template binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UpdateAppTemplateBindingResp{}, nil
}

// ListAppTemplateBindings list app template bindings
func (s *Service) ListAppTemplateBindings(ctx context.Context, req *pbcs.ListAppTemplateBindingsReq) (
	*pbcs.ListAppTemplateBindingsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListAppTemplateBindingsReq{
		BizId: req.BizId,
		AppId: req.AppId,
		All:   true,
	}

	rp, err := s.client.DS.ListAppTemplateBindings(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list app template bindings failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListAppTemplateBindingsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// parseBindings parse the bindings param
func parseBindings(bindings []*pbatb.TemplateBinding) (templateSetIDs, templateIDs []uint32, err error) {
	if len(bindings) == 0 {
		return nil, nil, errors.New("bindings can't be empty")
	}
	for _, b := range bindings {
		if b.TemplateSetId <= 0 {
			return nil, nil, fmt.Errorf("invalid template set id of bindings member: %d", b.TemplateSetId)
		}
		templateSetIDs = append(templateSetIDs, b.TemplateSetId)
		for _, r := range b.TemplateRevisions {
			if r.TemplateId <= 0 {
				return nil, nil, fmt.Errorf("invalid template id of bindings member: %d", r.TemplateId)
			}
			templateIDs = append(templateIDs, r.TemplateId)
			if r.TemplateRevisionId <= 0 {
				return nil, nil, fmt.Errorf("invalid template revision id of bindings member: %d", r.TemplateRevisionId)
			}
		}
	}

	return templateSetIDs, templateIDs, nil
}

// ListAppBoundTmplRevisions list app bound template revisions
func (s *Service) ListAppBoundTmplRevisions(ctx context.Context, req *pbcs.ListAppBoundTmplRevisionsReq) ( // nolint
	*pbcs.ListAppBoundTmplRevisionsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	tmplSetInfo, err := s.getAllAppTmplSets(grpcKit, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("get all app template sets failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	r := &pbds.ListAppBoundTmplRevisionsReq{
		BizId:        req.BizId,
		AppId:        req.AppId,
		SearchFields: req.SearchFields,
		SearchValue:  req.SearchValue,
		All:          true,
		WithStatus:   req.WithStatus,
	}

	var rp *pbds.ListAppBoundTmplRevisionsResp
	rp, err = s.client.DS.ListAppBoundTmplRevisions(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list app template revisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// group by template set
	tmplSetMap := make(map[uint32][]*pbatb.AppBoundTmplRevision)
	for _, d := range rp.Details {
		tmplSetMap[d.TemplateSetId] = append(tmplSetMap[d.TemplateSetId], d)
	}

	// 对比非模板配置, 检测是否存在冲突
	ci, err := s.client.DS.ListConfigItems(grpcKit.RpcCtx(), &pbds.ListConfigItemsReq{
		BizId: req.BizId,
		AppId: req.AppId,
		All:   true,
	})
	if err != nil {
		logs.Errorf("list config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	existingPaths := []string{}
	for _, v := range ci.GetDetails() {
		if v.FileState != constant.FileStateDelete {
			existingPaths = append(existingPaths, path.Join(v.Spec.Path, v.Spec.Name))
		}
	}

	// 所有套餐之间的冲突检测
	for _, tmplSet := range tmplSetInfo {
		revisions := tmplSetMap[tmplSet.TemplateSetId]
		for _, revision := range revisions {
			existingPaths = append(existingPaths, path.Join(revision.Path, revision.Name))
		}
	}
	_, conflictPaths := checkExistingPathConflict(existingPaths)

	details := make([]*pbatb.AppBoundTmplRevisionGroupBySet, 0)
	for _, tmplSet := range tmplSetInfo {
		group := &pbatb.AppBoundTmplRevisionGroupBySet{
			TemplateSpaceId:   tmplSet.TemplateSpaceId,
			TemplateSpaceName: tmplSet.TemplateSpaceName,
			TemplateSetId:     tmplSet.TemplateSetId,
			TemplateSetName:   tmplSet.TemplateSetName,
		}
		revisions := tmplSetMap[tmplSet.TemplateSetId]
		// 根据path+name排序
		sort.SliceStable(revisions, func(i, j int) bool {
			iPath := path.Join(revisions[i].Path, revisions[i].Name)
			jPath := path.Join(revisions[j].Path, revisions[j].Name)
			return iPath < jPath
		})
		for _, r := range revisions {
			var isConflict bool
			if r.FileState != constant.FileStateDelete {
				isConflict = conflictPaths[path.Join(r.Path, r.Name)]
			}
			group.TemplateRevisions = append(group.TemplateRevisions,
				&pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail{
					TemplateId:           r.TemplateId,
					Name:                 r.Name,
					Path:                 r.Path,
					TemplateRevisionId:   r.TemplateRevisionId,
					IsLatest:             r.IsLatest,
					TemplateRevisionName: r.TemplateRevisionName,
					TemplateRevisionMemo: r.TemplateRevisionMemo,
					FileType:             r.FileType,
					FileMode:             r.FileMode,
					User:                 r.User,
					UserGroup:            r.UserGroup,
					Privilege:            r.Privilege,
					Signature:            r.Signature,
					ByteSize:             r.ByteSize,
					Creator:              r.Creator,
					CreateAt:             r.CreateAt,
					FileState:            r.FileState,
					Md5:                  r.Md5,
					IsConflict:           isConflict,
				})
		}
		if req.WithStatus {
			sortFileStateInGroup(group, req.Status)
		}
		details = append(details, group)
	}

	// 自然排序
	sort.SliceStable(details, func(i, j int) bool {
		if details[i].TemplateSpaceName == details[j].TemplateSpaceName {
			return natsort.NaturalLess(details[i].TemplateSetName, details[j].TemplateSetName)
		}
		return natsort.NaturalLess(details[i].TemplateSpaceName, details[j].TemplateSpaceName)
	})

	resp := &pbcs.ListAppBoundTmplRevisionsResp{
		Details: details,
	}
	return resp, nil
}

// getAllAppTmplSets get all the template sets for the app, including empty template set which has not templates
func (s *Service) getAllAppTmplSets(grpcKit *kit.Kit, bizID, appID uint32) ([]*pbtset.TemplateSetBriefInfo, error) {
	atbReq := &pbds.ListAppTemplateBindingsReq{
		BizId: bizID,
		AppId: appID,
		All:   true,
	}

	atbRsp, err := s.client.DS.ListAppTemplateBindings(grpcKit.RpcCtx(), atbReq)
	if err != nil {
		logs.Errorf("list app template bindings failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	if len(atbRsp.Details) == 0 {
		return []*pbtset.TemplateSetBriefInfo{}, nil
	}
	tmplSetIDs := make([]uint32, 0)
	for _, b := range atbRsp.Details[0].Spec.Bindings {
		tmplSetIDs = append(tmplSetIDs, b.TemplateSetId)
	}

	var tsbRsp *pbds.ListTemplateSetBriefInfoByIDsResp
	tsbRsp, err = s.client.DS.ListTemplateSetBriefInfoByIDs(grpcKit.RpcCtx(), &pbds.ListTemplateSetBriefInfoByIDsReq{
		Ids: tmplSetIDs,
	})
	if err != nil {
		logs.Errorf("list template set brief info by ids failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return tsbRsp.Details, nil
}

// sortFileStateInGroup sort as add > revise > delete > unchange
func sortFileStateInGroup(g *pbatb.AppBoundTmplRevisionGroupBySet, status []string) {
	if len(g.TemplateRevisions) <= 1 {
		return
	}

	result := make([]*pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail, 0)
	add := make([]*pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail, 0)
	del := make([]*pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail, 0)
	revise := make([]*pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail, 0)
	unchange := make([]*pbatb.AppBoundTmplRevisionGroupBySetTemplateRevisionDetail, 0)
	for _, ci := range g.TemplateRevisions {
		switch ci.FileState {
		case constant.FileStateAdd:
			add = append(add, ci)
		case constant.FileStateDelete:
			del = append(del, ci)
		case constant.FileStateRevise:
			revise = append(revise, ci)
		case constant.FileStateUnchange:
			unchange = append(unchange, ci)
		}
	}

	if len(status) == 0 {
		result = append(result, add...)
		result = append(result, revise...)
		result = append(result, del...)
		result = append(result, unchange...)
	} else {
		for _, v := range status {
			switch strings.ToUpper(v) {
			case constant.FileStateAdd:
				result = append(result, add...)
			case constant.FileStateRevise:
				result = append(result, revise...)
			case constant.FileStateDelete:
				result = append(result, del...)
			case constant.FileStateUnchange:
				result = append(result, unchange...)
			}
		}
	}
	g.TemplateRevisions = result
}

// ListReleasedAppBoundTmplRevisions list released app bound template revisions
func (s *Service) ListReleasedAppBoundTmplRevisions(ctx context.Context,
	req *pbcs.ListReleasedAppBoundTmplRevisionsReq) (*pbcs.ListReleasedAppBoundTmplRevisionsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	if req.ReleaseId <= 0 {
		return nil, fmt.Errorf("invalid release id %d, it must bigger than 0", req.ReleaseId)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListReleasedAppBoundTmplRevisionsReq{
		BizId:        req.BizId,
		AppId:        req.AppId,
		ReleaseId:    req.ReleaseId,
		SearchFields: req.SearchFields,
		SearchValue:  req.SearchValue,
		All:          true,
	}

	rp, err := s.client.DS.ListReleasedAppBoundTmplRevisions(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list released app template revisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// group by template set
	tmplSetMap := make(map[uint32][]*pbatb.ReleasedAppBoundTmplRevision)
	for _, d := range rp.Details {
		tmplSetMap[d.TemplateSetId] = append(tmplSetMap[d.TemplateSetId], d)
	}

	details := make([]*pbatb.ReleasedAppBoundTmplRevisionGroupBySet, 0)
	for id, revisions := range tmplSetMap {
		group := &pbatb.ReleasedAppBoundTmplRevisionGroupBySet{
			TemplateSpaceId:   revisions[0].TemplateSpaceId,
			TemplateSpaceName: revisions[0].TemplateSpaceName,
			TemplateSetId:     id,
			TemplateSetName:   revisions[0].TemplateSetName,
		}
		for _, r := range revisions {
			group.TemplateRevisions = append(group.TemplateRevisions,
				&pbatb.ReleasedAppBoundTmplRevisionGroupBySetTemplateRevisionDetail{
					TemplateId:           r.TemplateId,
					Name:                 r.Name,
					Path:                 r.Path,
					TemplateRevisionId:   r.TemplateRevisionId,
					IsLatest:             r.IsLatest,
					TemplateRevisionName: r.TemplateRevisionName,
					TemplateRevisionMemo: r.TemplateRevisionMemo,
					FileType:             r.FileType,
					FileMode:             r.FileMode,
					User:                 r.User,
					UserGroup:            r.UserGroup,
					Privilege:            r.Privilege,
					Signature:            r.Signature,
					ByteSize:             r.ByteSize,
					OriginSignature:      r.OriginSignature,
					OriginByteSize:       r.OriginByteSize,
					Creator:              r.Creator,
					Reviser:              r.Reviser,
					CreateAt:             r.CreateAt,
					UpdateAt:             r.UpdateAt,
					Md5:                  r.Md5,
				})
		}
		details = append(details, group)
	}

	sort.SliceStable(details, func(i, j int) bool {
		if details[i].TemplateSpaceName == details[j].TemplateSpaceName {
			return natsort.NaturalLess(details[i].TemplateSetName, details[j].TemplateSetName)
		}
		return natsort.NaturalLess(details[i].TemplateSpaceName, details[j].TemplateSpaceName)
	})

	resp := &pbcs.ListReleasedAppBoundTmplRevisionsResp{
		Details: details,
	}
	return resp, nil
}

// GetReleasedAppBoundTmplRevision get released app bound template revision
func (s *Service) GetReleasedAppBoundTmplRevision(ctx context.Context,
	req *pbcs.GetReleasedAppBoundTmplRevisionReq) (*pbcs.GetReleasedAppBoundTmplRevisionResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	if req.ReleaseId <= 0 {
		return nil, fmt.Errorf("invalid release id %d, it must bigger than 0", req.ReleaseId)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.GetReleasedAppBoundTmplRevisionReq{
		BizId:              req.BizId,
		AppId:              req.AppId,
		ReleaseId:          req.ReleaseId,
		TemplateRevisionId: req.TemplateRevisionId,
	}

	rp, err := s.client.DS.GetReleasedAppBoundTmplRevision(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get released app template revision failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.GetReleasedAppBoundTmplRevisionResp{
		Detail: rp.Detail,
	}
	return resp, nil
}

// UpdateAppBoundTmplRevisions update app bound template revisions
func (s *Service) UpdateAppBoundTmplRevisions(ctx context.Context, req *pbcs.UpdateAppBoundTmplRevisionsReq) (
	*pbcs.UpdateAppBoundTmplRevisionsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	templateSetIDs, templateIDs, err := parseBindings(req.Bindings)
	if err != nil {
		logs.Errorf("update app bound template revisions failed, parse bindings err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	repeatedTmplSetIDs := tools.SliceRepeatedElements(templateSetIDs)
	if len(repeatedTmplSetIDs) > 0 {
		return nil, fmt.Errorf("repeated template set ids: %v, id must be unique", repeatedTmplSetIDs)
	}
	repeatedTmplRevisionIDs := tools.SliceRepeatedElements(templateIDs)
	if len(repeatedTmplRevisionIDs) > 0 {
		return nil, fmt.Errorf("repeated template ids: %v, id must be unique", repeatedTmplRevisionIDs)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err = s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	var (
		bResp    *pbcs.ListAppTemplateBindingsResp
		bindings []*pbatb.TemplateBinding
	)

	if bResp, err = s.ListAppTemplateBindings(ctx, &pbcs.ListAppTemplateBindingsReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}); err != nil {
		logs.Errorf("update app bound template revisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	if len(bResp.Details) == 0 {
		return nil, fmt.Errorf("app template binding not found, biz id: %d, app id: %d", req.BizId, req.AppId)
	}

	if bindings, err = getBindingsAfterUpdate(bResp.Details[0].Spec.Bindings, req.Bindings); err != nil {
		return nil, err
	}

	r := &pbds.UpdateAppTemplateBindingReq{
		Id: req.BindingId,
		Attachment: &pbatb.AppTemplateBindingAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbatb.AppTemplateBindingSpec{
			Bindings: bindings,
		},
	}
	if _, err = s.client.DS.UpdateAppTemplateBinding(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update app bound template revisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UpdateAppBoundTmplRevisionsResp{}, nil
}

// getBindingsAfterUpdate get the final template bindings after update
func getBindingsAfterUpdate(origin, update []*pbatb.TemplateBinding) ([]*pbatb.TemplateBinding, error) {
	// revisionsMap is the existent template revisions
	// map: template set id -> (map: template id -> template revision binding)
	revisionsMap := make(map[uint32]map[uint32]*pbatb.TemplateRevisionBinding)
	for _, b := range origin {
		if _, ok := revisionsMap[b.TemplateSetId]; !ok {
			revisionsMap[b.TemplateSetId] = make(map[uint32]*pbatb.TemplateRevisionBinding)
		}
		for _, t := range b.TemplateRevisions {
			revisionsMap[b.TemplateSetId][t.TemplateId] = t
		}
	}

	// update existent template revisions with the new template revisions
	for _, b := range update {
		if _, ok := revisionsMap[b.TemplateSetId]; !ok {
			return nil, fmt.Errorf("template set id %d is not existent for the app bound templates", b.TemplateSetId)
		}
		for _, t := range b.TemplateRevisions {
			if _, ok := revisionsMap[b.TemplateSetId][t.TemplateId]; !ok {
				return nil, fmt.Errorf("template id %d is not existent for the app bound templates", t.TemplateId)
			}
			revisionsMap[b.TemplateSetId][t.TemplateId] = t
		}
	}

	// final is the final template bindings after update
	final := make([]*pbatb.TemplateBinding, 0, len(revisionsMap))
	for tmplSetID, tmplRevisionMap := range revisionsMap {
		tmplRevisions := make([]*pbatb.TemplateRevisionBinding, 0, len(tmplRevisionMap))
		for _, tmplRevision := range tmplRevisionMap {
			tmplRevisions = append(tmplRevisions, tmplRevision)
		}
		final = append(final, &pbatb.TemplateBinding{
			TemplateSetId:     tmplSetID,
			TemplateRevisions: tmplRevisions,
		})
	}

	return final, nil
}

// DeleteAppBoundTmplSets delete app bound template sets
func (s *Service) DeleteAppBoundTmplSets(ctx context.Context, req *pbcs.DeleteAppBoundTmplSetsReq) (
	*pbcs.DeleteAppBoundTmplSetsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	templateSetIDs, err := tools.GetUint32List(req.TemplateSetIds)
	if err != nil {
		return nil, fmt.Errorf("invalid template set ids, %s", err)
	}
	idsLen := len(templateSetIDs)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template set ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}
	templateSetIDs = tools.RemoveDuplicates(templateSetIDs)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err = s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	var (
		bResp    *pbcs.ListAppTemplateBindingsResp
		bindings []*pbatb.TemplateBinding
	)

	if bResp, err = s.ListAppTemplateBindings(ctx, &pbcs.ListAppTemplateBindingsReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}); err != nil {
		logs.Errorf("delete app bound template sets failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	if len(bResp.Details) == 0 {
		return nil, fmt.Errorf("app template binding not found, biz id: %d, app id: %d", req.BizId, req.AppId)
	}

	if bindings, err = getBindingsAfterDelete(bResp.Details[0].Spec.Bindings, templateSetIDs); err != nil {
		return nil, err
	}

	r := &pbds.UpdateAppTemplateBindingReq{
		Id: req.BindingId,
		Attachment: &pbatb.AppTemplateBindingAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbatb.AppTemplateBindingSpec{
			Bindings: bindings,
		},
	}
	if _, err = s.client.DS.UpdateAppTemplateBinding(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete app bound template sets failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.DeleteAppBoundTmplSetsResp{}, nil
}

// getBindingsAfterDelete get the final template bindings after delete
func getBindingsAfterDelete(origin []*pbatb.TemplateBinding, deletedTmplSetIDs []uint32) (
	[]*pbatb.TemplateBinding, error) {
	// revisionsMap is the existent template revisions
	// map: template set id -> (map: template id -> template revision binding)
	revisionsMap := make(map[uint32]map[uint32]*pbatb.TemplateRevisionBinding)
	for _, b := range origin {
		if _, ok := revisionsMap[b.TemplateSetId]; !ok {
			revisionsMap[b.TemplateSetId] = make(map[uint32]*pbatb.TemplateRevisionBinding)
		}
		for _, t := range b.TemplateRevisions {
			revisionsMap[b.TemplateSetId][t.TemplateId] = t
		}
	}

	// delete existent template revisions with the new template revisions
	for _, id := range deletedTmplSetIDs {
		if _, ok := revisionsMap[id]; !ok {
			return nil, fmt.Errorf("template set id %d is not existent for the app bound templates", id)
		}
		delete(revisionsMap, id)
	}

	// final is the final template bindings after delete
	final := make([]*pbatb.TemplateBinding, 0, len(revisionsMap))
	for tmplSetID, tmplRevisionMap := range revisionsMap {
		tmplRevisions := make([]*pbatb.TemplateRevisionBinding, 0, len(tmplRevisionMap))
		for _, tmplRevision := range tmplRevisionMap {
			tmplRevisions = append(tmplRevisions, tmplRevision)
		}
		final = append(final, &pbatb.TemplateBinding{
			TemplateSetId:     tmplSetID,
			TemplateRevisions: tmplRevisions,
		})
	}

	return final, nil
}

// CheckAppTemplateBinding check conflicts of app template binding.
func (s *Service) CheckAppTemplateBinding(ctx context.Context, req *pbcs.CheckAppTemplateBindingReq) (
	*pbcs.CheckAppTemplateBindingResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	_, templateIDs, err := parseBindings(req.Bindings)
	if err != nil {
		logs.Errorf("create app template binding failed, parse bindings err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	if len(templateIDs) > 500 {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [1,500]",
			len(templateIDs))
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err = s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.CheckAppTemplateBindingReq{
		Attachment: &pbatb.AppTemplateBindingAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbatb.AppTemplateBindingSpec{
			Bindings: req.Bindings,
		},
	}
	var rp *pbds.CheckAppTemplateBindingResp
	rp, err = s.client.DS.CheckAppTemplateBinding(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create app template binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CheckAppTemplateBindingResp{
		Details: rp.Details,
	}
	return resp, nil
}
