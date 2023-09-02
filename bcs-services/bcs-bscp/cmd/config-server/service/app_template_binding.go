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

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbatb "bscp.io/pkg/protocol/core/app-template-binding"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// CreateAppTemplateBinding create an app template binding
func (s *Service) CreateAppTemplateBinding(ctx context.Context, req *pbcs.CreateAppTemplateBindingReq) (*pbcs.
	CreateAppTemplateBindingResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateAppTemplateBindingResp)

	templateSetIDs, templateIDs, err := parseBindings(req.Bindings)
	if err != nil {
		logs.Errorf("create app template binding failed, parse bindings err: %v, rid: %s", err, grpcKit.Rid)
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
	if len(templateIDs) > 500 {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [1,500]",
			len(templateIDs))
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
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
	rp, err := s.client.DS.CreateAppTemplateBinding(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create app template binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateAppTemplateBindingResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteAppTemplateBinding delete an app template binding
func (s *Service) DeleteAppTemplateBinding(ctx context.Context, req *pbcs.DeleteAppTemplateBindingReq) (*pbcs.
	DeleteAppTemplateBindingResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteAppTemplateBindingResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Delete,
		ResourceID: req.BindingId}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
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

	return resp, nil
}

// UpdateAppTemplateBinding update an app template binding
func (s *Service) UpdateAppTemplateBinding(ctx context.Context, req *pbcs.UpdateAppTemplateBindingReq) (*pbcs.
	UpdateAppTemplateBindingResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateAppTemplateBindingResp)

	templateSetIDs, templateIDs, err := parseBindings(req.Bindings)
	if err != nil {
		logs.Errorf("update app template binding failed, parse bindings err: %v, rid: %s", err, grpcKit.Rid)
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
	if len(templateIDs) > 500 {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [1,500]",
			len(templateIDs))
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Update,
		ResourceID: req.BindingId}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
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
	if _, err := s.client.DS.UpdateAppTemplateBinding(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update app template binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListAppTemplateBindings list app template bindings
func (s *Service) ListAppTemplateBindings(ctx context.Context, req *pbcs.ListAppTemplateBindingsReq) (*pbcs.
	ListAppTemplateBindingsResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAppTemplateBindingsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Find},
		BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListAppTemplateBindingsReq{
		BizId: req.BizId,
		AppId: req.AppId,
		Start: 0,
		Limit: uint32(types.DefaultMaxPageLimit),
	}

	rp, err := s.client.DS.ListAppTemplateBindings(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list app template bindings failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListAppTemplateBindingsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

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

// ListAppBoundTemplateRevisions list app bound template revisions
func (s *Service) ListAppBoundTemplateRevisions(ctx context.Context, req *pbcs.ListAppBoundTemplateRevisionsReq) (*pbcs.
	ListAppBoundTemplateRevisionsResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAppBoundTemplateRevisionsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Find},
		BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListAppBoundTemplateRevisionsReq{
		BizId:        req.BizId,
		AppId:        req.AppId,
		SearchFields: req.SearchFields,
		SearchValue:  req.SearchValue,
		Start:        req.Start,
		Limit:        req.Limit,
		All:          req.All,
	}

	rp, err := s.client.DS.ListAppBoundTemplateRevisions(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list app template bindings failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListAppBoundTemplateRevisionsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListReleasedAppBoundTemplateRevisions list released app bound template revisions
func (s *Service) ListReleasedAppBoundTemplateRevisions(ctx context.Context,
	req *pbcs.ListReleasedAppBoundTemplateRevisionsReq) (
	*pbcs.ListReleasedAppBoundTemplateRevisionsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListReleasedAppBoundTemplateRevisionsResp)

	if req.ReleaseId <= 0 {
		return nil, fmt.Errorf("invalid release id %d, it must bigger than 0", req.ReleaseId)
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Find},
		BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListReleasedAppBoundTemplateRevisionsReq{
		BizId:        req.BizId,
		AppId:        req.AppId,
		ReleaseId:    req.ReleaseId,
		SearchFields: req.SearchFields,
		SearchValue:  req.SearchValue,
		Start:        req.Start,
		Limit:        req.Limit,
		All:          req.All,
	}

	rp, err := s.client.DS.ListReleasedAppBoundTemplateRevisions(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list released app template bindings failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListReleasedAppBoundTemplateRevisionsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// UpdateAppBoundTemplateRevisions update app bound template revisions
func (s *Service) UpdateAppBoundTemplateRevisions(ctx context.Context, req *pbcs.UpdateAppBoundTemplateRevisionsReq) (
	*pbcs.
		UpdateAppBoundTemplateRevisionsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateAppBoundTemplateRevisionsResp)

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
	if len(templateIDs) > 500 {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [1,500]",
			len(templateIDs))
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Update,
		ResourceID: req.BindingId}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
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
	if _, err := s.client.DS.UpdateAppTemplateBinding(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update app bound template revisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
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

// DeleteAppBoundTemplateSets delete app bound template sets
func (s *Service) DeleteAppBoundTemplateSets(ctx context.Context, req *pbcs.DeleteAppBoundTemplateSetsReq) (
	*pbcs.DeleteAppBoundTemplateSetsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteAppBoundTemplateSetsResp)

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

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Update,
		ResourceID: req.BindingId}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
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
	if _, err := s.client.DS.UpdateAppTemplateBinding(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete app bound template sets failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
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
