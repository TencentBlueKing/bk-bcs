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
	"bscp.io/pkg/runtime/filter"
	"context"
	"errors"
	"fmt"
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	hr "bscp.io/pkg/protocol/core/hook-release"
	pbhr "bscp.io/pkg/protocol/core/hook-release"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateHookRelease create hook release  with option
func (s *Service) CreateHookRelease(ctx context.Context,
	req *pbds.CreateHookReleaseReq) (*pbds.CreateResp, error) {

	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.Hook().GetByID(kt, req.Attachment.BizId, req.Attachment.HookId); err != nil {
		logs.Errorf("get hook (%d) failed, err: %v, rid: %s", req.Attachment.HookId, err, kt.Rid)
		return nil, err
	}

	if _, err := s.dao.HookRelease().GetByName(kt, req.Attachment.BizId, req.Attachment.HookId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("hook name %s already exists", req.Spec.Name)
	}

	spec, err := req.Spec.HookReleaseSpec()
	spec.State = table.NotDeployedHookReleased
	if err != nil {
		logs.Errorf("get HookReleaseSpec spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	hookRelease := &table.HookRelease{
		Spec:       spec,
		Attachment: req.Attachment.HookReleaseAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}

	id, err := s.dao.HookRelease().Create(kt, hookRelease)
	if err != nil {
		logs.Errorf("create HookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListHookReleases list HookRelease with filter
func (s *Service) ListHookReleases(ctx context.Context,
	req *pbds.ListHookReleasesReq) (*pbds.ListHookReleasesResp, error) {

	kt := kit.FromGrpcContext(ctx)

	page := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	opt := &types.ListHookReleasesOption{
		BizID:     req.BizId,
		HookID:    req.HookId,
		SearchKey: req.SearchKey,
		Page:      page,
		State:     table.HookReleaseStatus(req.State),
	}
	po := &types.PageOption{
		EnableUnlimitedLimit: true,
	}
	if err := opt.Validate(po); err != nil {
		return nil, err
	}

	details, count, err := s.dao.HookRelease().List(kt, opt)
	if err != nil {
		logs.Errorf("list HookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	hookRelease, err := pbhr.PbHookReleaseSpaces(details)
	if err != nil {
		logs.Errorf("get pb hookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListHookReleasesResp{
		Count:   uint32(count),
		Details: hookRelease,
	}
	return resp, nil
}

// GetHookReleaseByID get a release
func (s *Service) GetHookReleaseByID(ctx context.Context,
	req *pbds.GetHookReleaseByIdReq) (*hr.HookRelease, error) {

	kt := kit.FromGrpcContext(ctx)

	hookRelease, err := s.dao.HookRelease().Get(kt, req.BizId, req.HookId, req.Id)
	if err != nil {
		logs.Errorf("get app by id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp, _ := hr.PbHookRelease(hookRelease)
	return resp, nil
}

// DeleteHookRelease delete a HookRelease
func (s *Service) DeleteHookRelease(ctx context.Context,
	req *pbds.DeleteHookReleaseReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	HookRelease := &table.HookRelease{
		ID: req.Id,
		Attachment: &table.HookReleaseAttachment{
			BizID:  req.BizId,
			HookID: req.HookId,
		},
	}

	if err := s.dao.HookRelease().Delete(kt, HookRelease); err != nil {
		logs.Errorf("delete HookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// PublishHookRelease publish a release
func (s *Service) PublishHookRelease(ctx context.Context, req *pbds.PublishHookReleaseReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	r := &table.HookRelease{
		Attachment: &table.HookReleaseAttachment{
			BizID:  req.BizId,
			HookID: req.HookId,
		},
		Spec: &table.HookReleaseSpec{
			State: table.NotDeployedHookReleased,
		},
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}

	tx := s.dao.GenQuery().Begin()

	// 1. 上线的版本下线
	opt := &types.GetByPubStateOption{
		BizID:  req.BizId,
		HookID: req.HookId,
		State:  table.DeployedHookReleased,
	}
	old, err := s.dao.HookRelease().GetByPubState(kt, opt)
	if err == nil {
		r.ID = old.ID
		r.Spec.State = table.ShutdownHookReleased
		if e := s.dao.HookRelease().UpdatePubStateWithTx(kt, tx, r); e != nil {
			logs.Errorf("update HookRelease State failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, e
		}
	}

	// 2. 未上线的版本上线
	r.ID = req.Id
	r.Spec.State = table.DeployedHookReleased
	if e := s.dao.HookRelease().UpdatePubStateWithTx(kt, tx, r); e != nil {
		logs.Errorf("update HookRelease State failed, err: %v, rid: %s", e, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return new(pbbase.EmptyResp), nil

}

// GetHookReleaseByPubState get a HookRelease by State
func (s *Service) GetHookReleaseByPubState(ctx context.Context,
	req *pbds.GetByPubStateReq) (*hr.HookRelease, error) {

	kt := kit.FromGrpcContext(ctx)

	opt := &types.GetByPubStateOption{
		BizID:  req.BizId,
		HookID: req.HookId,
		State:  table.HookReleaseStatus(req.State),
	}

	release, err := s.dao.HookRelease().GetByPubState(kt, opt)
	if err != nil {
		logs.Errorf("get HookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp, _ := hr.PbHookRelease(release)

	return resp, nil
}

func (s *Service) UpdateHookRelease(ctx context.Context, req *pbds.UpdateHookReleaseReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	hookRelease, err := s.dao.HookRelease().Get(kt, req.Attachment.BizId, req.Attachment.HookId, req.Id)
	if err != nil {
		logs.Errorf("update HookRelease spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if hookRelease.Spec.State != table.NotDeployedHookReleased {
		logs.Errorf("update HookRelease spec from pb failed, err: HookRelease state is not not_released, rid: %s", kt.Rid)
		return nil, errors.New("the HookRelease pubState is not not_released, and cannot be updated")
	}

	spec, err := req.Spec.HookReleaseSpec()
	if err != nil {
		logs.Errorf("update HookRelease spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	hookRelease = &table.HookRelease{
		ID:         req.Id,
		Spec:       spec,
		Attachment: req.Attachment.HookReleaseAttachment(),
		Revision: &table.Revision{
			Reviser:   kt.User,
			UpdatedAt: now,
		},
	}
	if e := s.dao.HookRelease().Update(kt, hookRelease); e != nil {
		logs.Errorf("update hookRelease failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}

func (s *Service) ListHookReleasesReferences(ctx context.Context,
	req *pbds.ListHookReleasesReferencesReq) (*pbds.ListHookReleasesReferencesResp, error) {

	kt := kit.FromGrpcContext(ctx)

	page := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	opt := &types.ListHookReleasesReferencesOption{
		BizID:          req.BizId,
		HookID:         req.HookId,
		HookReleasesID: req.ReleasesId,
		Page:           page,
	}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	results, count, err := s.dao.HookRelease().ListHookReleasesReferences(kt, opt)
	if err != nil {
		logs.Errorf("list TemplateSpace failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	gcrs, err := s.dao.ReleasedGroup().List(kt, &types.ListReleasedGroupsOption{
		BizID: req.BizId,
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: []filter.RuleFactory{},
		},
	})
	if err != nil {
		logs.Errorf("list group current releases failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, result := range results {
		status, _ := s.queryPublishStatus(gcrs, result.ConfigReleaseID)
		result.PubSate = status
	}

	if err != nil {
		logs.Errorf("list group current releases failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListHookReleasesReferencesResp{
		Count:   uint32(count),
		Details: pbhr.PbListHookReleasesReferencesDetails(results),
	}

	return resp, nil

}
