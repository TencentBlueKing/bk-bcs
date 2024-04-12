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
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbhook "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/hook"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// CreateHook create a hook
func (s *Service) CreateHook(ctx context.Context, req *pbcs.CreateHookReq) (*pbcs.CreateHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.CreateHookReq{
		Attachment: &pbhook.HookAttachment{
			BizId: grpcKit.BizID,
		},
		Spec: &pbhook.HookSpec{
			Name:         req.Name,
			Type:         req.Type,
			Tag:          req.Tag,
			RevisionName: req.RevisionName,
			Memo:         req.Memo,
			Content:      req.Content,
		},
	}
	rp, err := s.client.DS.CreateHook(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateHookResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteHook delete a hook
func (s *Service) DeleteHook(ctx context.Context, req *pbcs.DeleteHookReq) (*pbcs.DeleteHookResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteHookResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
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

// BatchDeleteHook batch delete hook
func (s *Service) BatchDeleteHook(ctx context.Context, req *pbcs.BatchDeleteHookReq) (*pbcs.BatchDeleteHookResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	idsLen := len(req.Ids)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, errf.Errorf(errf.InvalidArgument, i18n.T(grpcKit,
			"the length of hook ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit))
	}

	eg, egCtx := errgroup.WithContext(grpcKit.RpcCtx())
	eg.SetLimit(10)

	var successfulIDs, failedIDs []uint32
	var mux sync.Mutex

	// 使用 data-service 原子接口
	for _, v := range req.Ids {
		v := v
		eg.Go(func() error {
			r := &pbds.DeleteHookReq{
				BizId:  req.BizId,
				HookId: v,
				Force:  req.Force,
			}
			if _, err := s.client.DS.DeleteHook(egCtx, r); err != nil {
				logs.Errorf("delete hook failed, err: %v, rid: %s", err, grpcKit.Rid)

				// 错误不返回异常，记录错误ID
				mux.Lock()
				failedIDs = append(failedIDs, v)
				mux.Unlock()
				return nil
			}

			mux.Lock()
			successfulIDs = append(successfulIDs, v)
			mux.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		logs.Errorf("batch delete failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, errf.Errorf(errf.Aborted, i18n.T(grpcKit, "batch delete failed"))
	}

	// 全部失败, 当前API视为失败
	if len(failedIDs) == len(req.Ids) {
		return nil, errf.Errorf(errf.Aborted, i18n.T(grpcKit, "batch delete failed"))
	}

	return &pbcs.BatchDeleteHookResp{SuccessfulIds: successfulIDs, FailedIds: failedIDs}, nil
}

// ListHooks list hooks with filter
func (s *Service) ListHooks(ctx context.Context, req *pbcs.ListHooksReq) (*pbcs.ListHooksResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListHooksReq{
		BizId:     grpcKit.BizID,
		Name:      req.Name,
		Tag:       req.Tag,
		All:       req.All,
		NotTag:    req.NotTag,
		SearchKey: req.SearchKey,
	}

	if !req.All {
		if req.Limit == 0 {
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

	resp := &pbcs.ListHooksResp{
		Count:   rp.Count,
		Details: details,
	}

	return resp, nil
}

// ListHookTags list tag
func (s *Service) ListHookTags(ctx context.Context, req *pbcs.ListHookTagsReq) (*pbcs.ListHookTagsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListHookTagReq{BizId: req.BizId}

	ht, err := s.client.DS.ListHookTags(grpcKit.RpcCtx(), r)
	if err != nil {
		return nil, err
	}

	resp := &pbcs.ListHookTagsResp{
		Details: ht.Details,
	}

	return resp, nil
}

// GetHook get a hook
func (s *Service) GetHook(ctx context.Context, req *pbcs.GetHookReq) (*pbcs.GetHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
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

	resp := &pbcs.GetHookResp{
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

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListHookReferencesReq{
		BizId:     req.BizId,
		HookId:    req.HookId,
		Limit:     req.Limit,
		Start:     req.Start,
		SearchKey: req.SearchKey,
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
			Deprecated:       detail.Deprecated,
		})
	}
	resp := &pbcs.ListHookReferencesResp{
		Count:   rp.Count,
		Details: details,
	}

	return resp, nil

}

// GetReleaseHook get release's pre hook and post hook
func (s *Service) GetReleaseHook(ctx context.Context, req *pbcs.GetReleaseHookReq) (*pbcs.GetReleaseHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
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
	resp := &pbcs.GetReleaseHookResp{
		PreHook:  pre,
		PostHook: post,
	}

	return resp, nil
}
