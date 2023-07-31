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

	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbhook "bscp.io/pkg/protocol/core/hook"
	pbds "bscp.io/pkg/protocol/data-service"
)

// CreateHook create a hook
func (s *Service) CreateHook(ctx context.Context, req *pbcs.CreateHookReq) (*pbcs.CreateHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateHookResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Hook, Action: meta.Create,
		ResourceID: req.BizId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.CreateHookReq{
		Attachment: &pbhook.HookAttachment{
			BizId: grpcKit.BizID,
		},
		Spec: &pbhook.HookSpec{
			Name:    req.Name,
			Type:    req.Type,
			Tag:     req.Tag,
			Memo:    req.Memo,
			Content: req.Content,
		},
	}
	rp, err := s.client.DS.CreateHook(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateHookResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteHook delete a hook
func (s *Service) DeleteHook(ctx context.Context, req *pbcs.DeleteHookReq) (*pbcs.DeleteHookResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteHookResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Hook, Action: meta.Delete,
		ResourceID: req.HookId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.DeleteHookReq{
		BizId:  req.BizId,
		HookId: req.HookId,
		Force:  req.Force,
	}
	if _, err := s.client.DS.DeleteHook(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListHooks list hooks with filter
func (s *Service) ListHooks(ctx context.Context, req *pbcs.ListHooksReq) (*pbcs.ListHooksResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListHooksResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListHooksReq{
		BizId:  grpcKit.BizID,
		Name:   req.Name,
		Tag:    req.Tag,
		All:    req.All,
		NotTag: req.NotTag,
	}

	if !req.All {
		if req.Start < 0 {
			return nil, errors.New("start has to be greater than 0")
		}

		if req.Limit < 0 {
			return nil, errors.New("limit has to be greater than 0")
		}

		r.Start = req.Start
		r.Limit = req.Limit
	}

	rp, err := s.client.DS.ListHooks(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	details := make([]*pbcs.ListHooksResp_Detail, 0, len(rp.Details))
	for _, detail := range rp.Details {
		details = append(details, &pbcs.ListHooksResp_Detail{
			Hook:                detail.Hook,
			BoundNum:            detail.BoundNum,
			ConfirmDelete:       detail.ConfirmDelete,
			PublishedRevisionId: detail.PublishedRevisionId,
		})
	}

	resp = &pbcs.ListHooksResp{
		Count:   rp.Count,
		Details: details,
	}

	return resp, nil
}

// ListHookTags list tag
func (s *Service) ListHookTags(ctx context.Context, req *pbcs.ListHookTagsReq) (*pbcs.ListHookTagsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListHookTagsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Hook, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListHookTagReq{BizId: req.BizId}

	ht, err := s.client.DS.ListHookTags(grpcKit.RpcCtx(), r)
	if err != nil {
		return nil, err
	}

	resp = &pbcs.ListHookTagsResp{
		Details: ht.Details,
	}

	return resp, nil
}

// GetHook get a hook
func (s *Service) GetHook(ctx context.Context, req *pbcs.GetHookReq) (*pbcs.GetHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.GetHookResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.GetHookReq{
		BizId:  req.BizId,
		HookId: req.HookId,
	}

	hook, err := s.client.DS.GetHook(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.GetHookResp{
		Id: hook.Id,
		Spec: &pbcs.GetHookInfoSpec{
			Name:     hook.Spec.Name,
			Type:     hook.Spec.Type,
			Tag:      hook.Spec.Tag,
			Memo:     hook.Spec.Memo,
			Releases: &pbcs.GetHookInfoSpec_Releases{NotReleaseId: hook.Spec.Releases.NotReleaseId},
		},
		Attachment: &pbhook.HookAttachment{BizId: hook.Attachment.BizId},
		Revision: &pbbase.Revision{
			Creator:  hook.Revision.CreateAt,
			Reviser:  hook.Revision.Reviser,
			CreateAt: hook.Revision.Creator,
			UpdateAt: hook.Revision.UpdateAt,
		},
	}

	return resp, nil
}

// ListHookReferences 查询hook版本被引用列表
func (s *Service) ListHookReferences(ctx context.Context,
	req *pbcs.ListHookReferencesReq) (*pbcs.ListHookReferencesResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListHookReferencesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListHookReferencesReq{
		BizId:  req.BizId,
		HookId: req.HookId,
		Limit:  req.Limit,
		Start:  req.Start,
	}

	rp, err := s.client.DS.ListHookReferences(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list TemplateSpaces failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	details := []*pbcs.ListHookReferencesResp_Detail{}

	for _, detail := range rp.Details {
		details = append(details, &pbcs.ListHookReferencesResp_Detail{
			HookRevisionId:   detail.HookRevisionId,
			HookRevisionName: detail.HookRevisionName,
			AppId:            detail.AppId,
			AppName:          detail.AppName,
			ReleaseId:        detail.ReleaseId,
			ReleaseName:      detail.ReleaseName,
			Type:             detail.Type,
		})
	}
	resp = &pbcs.ListHookReferencesResp{
		Count:   rp.Count,
		Details: details,
	}

	return resp, nil

}

// GetReleaseHook get release's pre hook and post hook
func (s *Service) GetReleaseHook(ctx context.Context, req *pbcs.GetReleaseHookReq) (*pbcs.GetReleaseHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.GetReleaseHookResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Release, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.GetReleaseHookReq{
		BizId:     req.BizId,
		AppId:     req.AppId,
		ReleaseId: req.ReleaseId,
	}

	grhResp, err := s.client.DS.GetReleaseHook(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	var pre, post *pbcs.GetReleaseHookResp_Hook
	if grhResp.PreHook != nil {
		pre = &pbcs.GetReleaseHookResp_Hook{
			HookId:           grhResp.PreHook.HookId,
			HookName:         grhResp.PreHook.HookName,
			HookRevisionId:   grhResp.PreHook.HookRevisionId,
			HookRevisionName: grhResp.PreHook.HookRevisionName,
			Type:             grhResp.PreHook.Type,
			Content:          grhResp.PreHook.Content,
		}
	}
	if grhResp.PostHook != nil {
		post = &pbcs.GetReleaseHookResp_Hook{
			HookId:           grhResp.PostHook.HookId,
			HookName:         grhResp.PostHook.HookName,
			HookRevisionId:   grhResp.PostHook.HookRevisionId,
			HookRevisionName: grhResp.PostHook.HookRevisionName,
			Type:             grhResp.PostHook.Type,
			Content:          grhResp.PostHook.Content,
		}
	}
	resp = &pbcs.GetReleaseHookResp{
		PreHook:  pre,
		PostHook: post,
	}

	return resp, nil
}
