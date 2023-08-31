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

	"gorm.io/gorm"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbhook "bscp.io/pkg/protocol/core/hook"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// CreateHook create hook.
func (s *Service) CreateHook(ctx context.Context, req *pbds.CreateHookReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.Hook().GetByName(kt, req.Attachment.BizId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("hook name %s already exists", req.Spec.Name)
	}

	spec, err := req.Spec.HookSpec()
	if err != nil {
		logs.Errorf("get hook spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	res := &table.Revision{
		Creator: kt.User,
		Reviser: kt.User,
	}

	tx := s.dao.GenQuery().Begin()

	// 1. create hook
	hook := &table.Hook{
		Spec:       spec,
		Attachment: req.Attachment.HookAttachment(),
		Revision:   res,
	}

	id, err := s.dao.Hook().CreateWithTx(kt, tx, hook)
	if err != nil {
		logs.Errorf("create hook failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 2. create hook revision
	revision := &table.HookRevision{
		Spec: &table.HookRevisionSpec{
			Name:    tools.GenerateRevisionName(),
			Content: req.Spec.Content,
			Memo:    req.Spec.Memo,
			State:   table.HookRevisionStatusDeployed,
		},
		Attachment: &table.HookRevisionAttachment{
			BizID:  req.Attachment.BizId,
			HookID: id,
		},
		Revision: res,
	}
	_, err = s.dao.HookRevision().CreateWithTx(kt, tx, revision)
	if err != nil {
		logs.Errorf("create hook revision failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListHooks list hooks.
func (s *Service) ListHooks(ctx context.Context, req *pbds.ListHooksReq) (*pbds.ListHooksResp, error) {

	kt := kit.FromGrpcContext(ctx)

	page := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	opt := &types.ListHooksWithReferOption{
		BizID:  req.BizId,
		Name:   req.Name,
		Tag:    req.Tag,
		All:    req.All,
		NotTag: req.NotTag,
		Page:   page,
	}

	po := &types.PageOption{
		EnableUnlimitedLimit: true,
	}
	if err := opt.Validate(po); err != nil {
		return nil, err
	}

	details, count, err := s.dao.Hook().ListWithRefer(kt, opt)
	if err != nil {
		logs.Errorf("list hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if err != nil {
		logs.Errorf("list hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := []*pbds.ListHooksResp_Detail{}
	for _, detail := range details {
		result = append(result, &pbds.ListHooksResp_Detail{
			Hook:                pbhook.PbHook(detail.Hook),
			BoundNum:            uint32(detail.ReferCount),
			ConfirmDelete:       detail.BoundEditingRelease,
			PublishedRevisionId: detail.PublishedRevisionID,
		})
	}

	resp := &pbds.ListHooksResp{
		Count:   uint32(count),
		Details: result,
	}
	return resp, nil
}

// DeleteHook delete hook.
func (s *Service) DeleteHook(ctx context.Context, req *pbds.DeleteHookReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	tx := s.dao.GenQuery().Begin()

	// 1. check if hook was bound to an editing release
	count, err := s.dao.ReleasedHook().CountByHookIDAndReleaseID(kt, req.BizId, req.HookId, 0)
	if err != nil {
		logs.Errorf("count hook bound editing releases failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}
	if count > 0 && !req.Force {
		tx.Rollback()
		return nil, fmt.Errorf("hook was bound to %d editing releases, "+
			"set force=true to delete hook with references, rid: %s", count, kt.Rid)
	}
	// 2. delete released hook that release_id = 0
	if err := s.dao.ReleasedHook().DeleteByHookIDAndReleaseIDWithTx(kt, tx, req.BizId, req.HookId, 0); err != nil {
		logs.Errorf("delete released hook failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}
	// 3. delete all hook revisions by hook id
	if err := s.dao.HookRevision().DeleteByHookIDWithTx(kt, tx, req.HookId, req.BizId); err != nil {
		logs.Errorf("delete hook revision failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}
	// 4. delete hook
	hook := &table.Hook{
		ID: req.HookId,
		Attachment: &table.HookAttachment{
			BizID: req.BizId,
		},
	}
	if err := s.dao.Hook().DeleteWithTx(kt, tx, hook); err != nil {
		logs.Errorf("delete hook failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return new(pbbase.EmptyResp), nil
}

// ListHookTags list tag
func (s *Service) ListHookTags(ctx context.Context, req *pbds.ListHookTagReq) (*pbds.ListHookTagResp, error) {

	kt := kit.FromGrpcContext(ctx)

	ht, err := s.dao.Hook().CountHookTag(kt, req.BizId)
	if err != nil {
		logs.Errorf("list hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListHookTagResp{}

	for _, count := range ht {
		resp.Details = append(resp.Details, &pbhook.CountHookTags{
			Tag:    count.Tag,
			Counts: count.Counts,
		})
	}

	return resp, nil
}

// GetHook get a hook
func (s *Service) GetHook(ctx context.Context, req *pbds.GetHookReq) (*pbds.GetHookResp, error) {

	kt := kit.FromGrpcContext(ctx)

	h, err := s.dao.Hook().GetByID(kt, req.BizId, req.HookId)
	if err != nil {
		logs.Errorf("list hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	opt := &types.GetByPubStateOption{
		BizID:  req.BizId,
		HookID: req.HookId,
		State:  table.HookRevisionStatusNotDeployed,
	}
	var revisionID uint32
	revision, err := s.dao.HookRevision().GetByPubState(kt, opt)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("get hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		revisionID = 0
	} else {
		revisionID = revision.ID
	}

	resp := &pbds.GetHookResp{
		Id: h.ID,
		Spec: &pbds.GetHookInfoSpec{
			Name:     h.Spec.Name,
			Type:     string(h.Spec.Type),
			Tag:      h.Spec.Tag,
			Memo:     h.Spec.Memo,
			Releases: &pbds.GetHookInfoSpec_Releases{NotReleaseId: revisionID},
		},
		Attachment: pbhook.PbHookAttachment(h.Attachment),
		Revision:   pbbase.PbRevision(h.Revision),
	}

	return resp, nil
}

// ListHookReferences ..
func (s *Service) ListHookReferences(ctx context.Context,
	req *pbds.ListHookReferencesReq) (*pbds.ListHookReferencesResp, error) {

	kt := kit.FromGrpcContext(ctx)

	page := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	opt := &types.ListHookReferencesOption{
		BizID:  req.BizId,
		HookID: req.HookId,
		Page:   page,
	}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	results, count, err := s.dao.Hook().ListHookReferences(kt, opt)
	if err != nil {
		logs.Errorf("list hook references failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]*pbds.ListHookReferencesResp_Detail, 0, len(results))
	for _, result := range results {
		details = append(details, &pbds.ListHookReferencesResp_Detail{
			HookRevisionId:   result.HookRevisionID,
			HookRevisionName: result.HookRevisionName,
			AppId:            result.AppID,
			AppName:          result.AppName,
			ReleaseId:        result.ReleaseID,
			ReleaseName:      result.ReleaseName,
			Type:             result.HookType,
		})
	}

	if err != nil {
		logs.Errorf("list group current releases failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListHookReferencesResp{
		Count:   uint32(count),
		Details: details,
	}

	return resp, nil
}

// GetReleaseHook get release's pre hook and post hook
func (s *Service) GetReleaseHook(ctx context.Context, req *pbds.GetReleaseHookReq) (*pbds.GetReleaseHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	preHook, err := s.dao.ReleasedHook().Get(grpcKit, req.BizId, req.AppId, req.ReleaseId, table.PreHook)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("get pre hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	postHook, err := s.dao.ReleasedHook().Get(grpcKit, req.BizId, req.AppId, req.ReleaseId, table.PostHook)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("get post hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	var pre, post *pbds.GetReleaseHookResp_Hook
	if preHook != nil {
		pre = &pbds.GetReleaseHookResp_Hook{
			HookId:           preHook.HookID,
			HookName:         preHook.HookName,
			HookRevisionId:   preHook.HookRevisionID,
			HookRevisionName: preHook.HookRevisionName,
			Type:             preHook.ScriptType.String(),
			Content:          preHook.Content,
		}
	}
	if postHook != nil {
		post = &pbds.GetReleaseHookResp_Hook{
			HookId:           postHook.HookID,
			HookName:         postHook.HookName,
			HookRevisionId:   postHook.HookRevisionID,
			HookRevisionName: postHook.HookRevisionName,
			Type:             postHook.ScriptType.String(),
			Content:          postHook.Content,
		}
	}
	resp := &pbds.GetReleaseHookResp{
		PreHook:  pre,
		PostHook: post,
	}

	return resp, nil
}
