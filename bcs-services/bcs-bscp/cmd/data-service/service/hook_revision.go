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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	hr "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/hook-revision"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateHookRevision create hook revision with option
func (s *Service) CreateHookRevision(ctx context.Context,
	req *pbds.CreateHookRevisionReq) (*pbds.CreateResp, error) {

	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.Hook().GetByID(kt, req.Attachment.BizId, req.Attachment.HookId); err != nil {
		logs.Errorf("get hook (%d) failed, err: %v, rid: %s", req.Attachment.HookId, err, kt.Rid)
		return nil, err
	}

	if _, err := s.dao.HookRevision().GetByName(kt, req.Attachment.BizId, req.Attachment.HookId,
		req.Spec.Name); err == nil {
		return nil, fmt.Errorf("hook name %s already exists", req.Spec.Name)
	}

	spec, err := req.Spec.HookRevisionSpec()
	// if no revision name is specified, generate it by system
	if spec.Name == "" {
		spec.Name = tools.GenerateRevisionName()
	}
	spec.State = table.HookRevisionStatusNotDeployed
	if err != nil {
		logs.Errorf("get HookRevisionSpec spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	hookRevision := &table.HookRevision{
		Spec:       spec,
		Attachment: req.Attachment.HookRevisionAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}

	id, err := s.dao.HookRevision().Create(kt, hookRevision)
	if err != nil {
		logs.Errorf("create HookRevision failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListHookRevisions list HookRevision with filter
func (s *Service) ListHookRevisions(ctx context.Context,
	req *pbds.ListHookRevisionsReq) (*pbds.ListHookRevisionsResp, error) {

	kt := kit.FromGrpcContext(ctx)

	page := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	opt := &types.ListHookRevisionsOption{
		BizID:     req.BizId,
		HookID:    req.HookId,
		SearchKey: req.SearchKey,
		Page:      page,
		State:     table.HookRevisionStatus(req.State),
	}
	po := &types.PageOption{
		EnableUnlimitedLimit: true,
	}
	if err := opt.Validate(po); err != nil {
		return nil, err
	}

	details, count, err := s.dao.HookRevision().ListWithRefer(kt, opt)
	if err != nil {
		logs.Errorf("list HookRevision failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]*pbds.ListHookRevisionsResp_ListHookRevisionsData, 0, len(details))
	for _, detail := range details {
		result = append(result, &pbds.ListHookRevisionsResp_ListHookRevisionsData{
			HookRevision:  hr.PbHookRevision(detail.HookRevision),
			BoundNum:      uint32(detail.ReferCount),
			ConfirmDelete: detail.BoundEditingRelease,
		})
	}

	resp := &pbds.ListHookRevisionsResp{
		Count:   uint32(count),
		Details: result,
	}
	return resp, nil
}

// GetHookRevisionByID get a release
func (s *Service) GetHookRevisionByID(ctx context.Context,
	req *pbds.GetHookRevisionByIdReq) (*hr.HookRevision, error) {

	kt := kit.FromGrpcContext(ctx)

	hookRevision, err := s.dao.HookRevision().Get(kt, req.BizId, req.HookId, req.Id)
	if err != nil {
		logs.Errorf("get app by id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return hr.PbHookRevision(hookRevision), nil
}

// DeleteHookRevision delete a HookRevision
func (s *Service) DeleteHookRevision(ctx context.Context,
	req *pbds.DeleteHookRevisionReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	tx := s.dao.GenQuery().Begin()

	// 1. check if hook was bound to an editing release
	count, err := s.dao.ReleasedHook().CountByHookRevisionIDAndReleaseID(kt, req.BizId, req.HookId, req.RevisionId, 0)
	if err != nil {
		logs.Errorf("count hook revision bound editing releases failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}
	if count > 0 && !req.Force {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, fmt.Errorf("hook revision was bound to %d editing releases, "+
			"set force=true to delete hook revision with references, rid: %s", count, kt.Rid)
	}

	// 2. delete released hook that release_id = 0
	if e := s.dao.ReleasedHook().DeleteByHookRevisionIDAndReleaseIDWithTx(kt, tx,
		req.BizId, req.HookId, req.RevisionId, 0); e != nil {
		logs.Errorf("delete released hook failed, err: %v, rid: %s", e, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}
	// 3. delete hook revision
	HookRevision := &table.HookRevision{
		ID: req.RevisionId,
		Attachment: &table.HookRevisionAttachment{
			BizID:  req.BizId,
			HookID: req.HookId,
		},
	}

	if e := s.dao.HookRevision().DeleteWithTx(kt, tx, HookRevision); e != nil {
		logs.Errorf("delete HookRevision failed, err: %v, rid: %s", e, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}

// PublishHookRevision publish a release
func (s *Service) PublishHookRevision(ctx context.Context, req *pbds.PublishHookRevisionReq) (
	*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	tx := s.dao.GenQuery().Begin()

	// 1. 上线的版本下线
	opt := &types.GetByPubStateOption{
		BizID:  req.BizId,
		HookID: req.HookId,
		State:  table.HookRevisionStatusDeployed,
	}
	old, err := s.dao.HookRevision().GetByPubState(kt, opt)
	if err == nil {
		old.Spec.State = table.HookRevisionStatusShutdown
		if e := s.dao.HookRevision().UpdatePubStateWithTx(kt, tx, old); e != nil {
			logs.Errorf("update HookRevision State failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, e
		}
	}

	// 2. 上线脚本版本
	hookRevision, err := s.dao.HookRevision().Get(kt, req.BizId, req.HookId, req.Id)
	if err != nil {
		logs.Errorf("get HookRevision failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}
	hookRevision.Revision.Reviser = kt.User
	hookRevision.Spec.State = table.HookRevisionStatusDeployed
	if e := s.dao.HookRevision().UpdatePubStateWithTx(kt, tx, hookRevision); e != nil {
		logs.Errorf("update HookRevision State failed, err: %v, rid: %s", e, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}

	// 3. 修改未命名版本绑定的脚本版本为上线版本
	if e := s.dao.ReleasedHook().UpdateHookRevisionByReleaseIDWithTx(kt, tx,
		req.BizId, 0, req.HookId, hookRevision); e != nil {
		logs.Errorf("update released hook failed, err: %v, rid: %s", e, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, e
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil

}

// GetHookRevisionByPubState get a HookRevision by State
func (s *Service) GetHookRevisionByPubState(ctx context.Context,
	req *pbds.GetByPubStateReq) (*hr.HookRevision, error) {

	kt := kit.FromGrpcContext(ctx)

	opt := &types.GetByPubStateOption{
		BizID:  req.BizId,
		HookID: req.HookId,
		State:  table.HookRevisionStatus(req.State),
	}

	release, err := s.dao.HookRevision().GetByPubState(kt, opt)
	if err != nil {
		logs.Errorf("get HookRevision failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return hr.PbHookRevision(release), nil
}

// UpdateHookRevision ..
func (s *Service) UpdateHookRevision(ctx context.Context, req *pbds.UpdateHookRevisionReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	hookRevision, err := s.dao.HookRevision().Get(kt, req.Attachment.BizId, req.Attachment.HookId, req.Id)
	if err != nil {
		logs.Errorf("update HookRevision spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if hookRevision.Spec.State != table.HookRevisionStatusNotDeployed {
		logs.Errorf("update HookRevision spec from pb failed, err: HookRevision state is not not_released, rid: %s", kt.Rid)
		return nil, errors.New("the HookRevision pubState is not not_released, and cannot be updated")
	}

	spec, err := req.Spec.HookRevisionSpec()
	if err != nil {
		logs.Errorf("update HookRevision spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	hookRevision = &table.HookRevision{
		ID:         req.Id,
		Spec:       spec,
		Attachment: req.Attachment.HookRevisionAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if e := s.dao.HookRevision().Update(kt, hookRevision); e != nil {
		logs.Errorf("update hookRevision failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}

// ListHookRevisionReferences ..
func (s *Service) ListHookRevisionReferences(ctx context.Context,
	req *pbds.ListHookRevisionReferencesReq) (*pbds.ListHookRevisionReferencesResp, error) {

	kt := kit.FromGrpcContext(ctx)

	page := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	opt := &types.ListHookRevisionReferencesOption{
		BizID:           req.BizId,
		HookID:          req.HookId,
		HookRevisionsID: req.RevisionId,
		SearchKey:       req.SearchKey,
		Page:            page,
	}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	results, count, err := s.dao.HookRevision().ListHookRevisionReferences(kt, opt)
	if err != nil {
		logs.Errorf("list hook references failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	details := make([]*pbds.ListHookRevisionReferencesResp_Detail, 0, len(results))
	for _, result := range results {
		details = append(details, &pbds.ListHookRevisionReferencesResp_Detail{
			RevisionId:   result.RevisionID,
			RevisionName: result.RevisionName,
			AppId:        result.AppID,
			AppName:      result.AppName,
			ReleaseId:    result.ReleaseID,
			ReleaseName:  result.ReleaseName,
			Type:         result.HookType,
			Deprecated:   result.Deprecated,
		})
	}

	if err != nil {
		logs.Errorf("list group current releases failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListHookRevisionReferencesResp{
		Count:   uint32(count),
		Details: details,
	}

	return resp, nil

}
