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

	// 2. create hook release
	release := &table.HookRelease{
		Spec: &table.HookReleaseSpec{
			Name:    req.Spec.ReleaseName,
			Content: req.Spec.Content,
			State:   table.NotDeployedHookReleased,
		},
		Attachment: &table.HookReleaseAttachment{
			BizID:  req.Attachment.BizId,
			HookID: id,
		},
		Revision: res,
	}
	_, err = s.dao.HookRelease().CreateWithTx(kt, tx, release)
	if err != nil {
		logs.Errorf("create hook release failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListHooks list hooks.
func (s *Service) ListHooks(ctx context.Context, req *pbds.ListHooksReq) (*pbds.ListHooksResp, error) {

	kt := kit.FromGrpcContext(ctx)

	page := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	opt := &types.ListHooksOption{
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

	details, count, err := s.dao.Hook().List(kt, opt)
	if err != nil {
		logs.Errorf("list hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if err != nil {
		logs.Errorf("list hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	hooks, err := pbhook.PbHooks(details)
	if err != nil {
		logs.Errorf("get pb hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListHooksResp{
		Count:   uint32(count),
		Details: hooks,
	}
	return resp, nil
}

// DeleteHook delete hook.
func (s *Service) DeleteHook(ctx context.Context, req *pbds.DeleteHookReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	tx := s.dao.GenQuery().Begin()

	// 1. delete hook
	hook := &table.Hook{
		ID:         req.Id,
		Attachment: req.Attachment.HookAttachment(),
	}
	if err := s.dao.Hook().DeleteWithTx(kt, tx, hook); err != nil {
		logs.Errorf("delete hook failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 2. delete hook release
	release := &table.HookRelease{
		Attachment: &table.HookReleaseAttachment{
			BizID:  req.Attachment.BizId,
			HookID: req.Id,
		},
	}
	if err := s.dao.HookRelease().DeleteByHookIDWithTx(kt, tx, release); err != nil {
		logs.Errorf("delete hook release failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
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
		State:  table.ShutdownHookReleased,
	}
	var releaseID uint32
	release, err := s.dao.HookRelease().GetByPubState(kt, opt)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("get hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		releaseID = 0
	} else {
		releaseID = release.ID
	}

	resp := &pbds.GetHookResp{
		Id: h.ID,
		Spec: &pbds.GetHookInfoSpec{
			Name:       h.Spec.Name,
			Type:       string(h.Spec.Type),
			Tag:        h.Spec.Tag,
			Memo:       h.Spec.Memo,
			PublishNum: h.Spec.PublishNum,
			Releases:   &pbds.GetHookInfoSpec_Releases{NotReleaseId: releaseID},
		},
		Attachment: pbhook.PbHookAttachment(h.Attachment),
		Revision:   pbbase.PbRevision(h.Revision),
	}

	return resp, nil
}
