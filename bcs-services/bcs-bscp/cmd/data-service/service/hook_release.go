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
	"fmt"

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
	}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
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
			PubState: table.FullReleased,
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
		State:  table.PartialReleased,
	}
	old, err := s.dao.HookRelease().GetByPubState(kt, opt)
	if err == nil {
		r.ID = old.ID
		if e := s.dao.HookRelease().UpdatePubStateWithTx(kt, tx, r); e != nil {
			logs.Errorf("update HookRelease PubState failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, e
		}
	}

	// 2. 未上线的版本上线
	r.ID = req.Id
	r.Spec.PubState = table.PartialReleased
	if e := s.dao.HookRelease().UpdatePubStateWithTx(kt, tx, r); e != nil {
		logs.Errorf("update HookRelease PubState failed, err: %v, rid: %s", e, kt.Rid)
		return nil, err
	}

	tx.Commit()

	return new(pbbase.EmptyResp), nil

}

// GetHookReleaseByPubState get a HookRelease by PubState
func (s *Service) GetHookReleaseByPubState(ctx context.Context,
	req *pbds.GetByPubStateReq) (*hr.HookRelease, error) {

	kt := kit.FromGrpcContext(ctx)

	opt := &types.GetByPubStateOption{
		BizID:  req.BizId,
		HookID: req.HookId,
		State:  table.ReleaseStatus(req.PubState),
	}

	release, err := s.dao.HookRelease().GetByPubState(kt, opt)
	if err != nil {
		logs.Errorf("get HookRelease failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp, _ := hr.PbHookRelease(release)

	return resp, nil
}
