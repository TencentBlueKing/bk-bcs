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
	"time"

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

	spec, err := req.Spec.HookSpec()
	if err != nil {
		logs.Errorf("get hook spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	hook := &table.Hook{
		Spec:       spec,
		Attachment: req.Attachment.HookAttachment(),
		Revision: &table.Revision{
			Creator:   kt.User,
			Reviser:   kt.User,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	id, err := s.dao.Hook().Create(kt, hook)
	if err != nil {
		logs.Errorf("create hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListHooks list hooks.
func (s *Service) ListHooks(ctx context.Context, req *pbds.ListHooksReq) (*pbds.ListHooksResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	query := &types.ListHooksOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.Hook().List(kt, query)
	if err != nil {
		logs.Errorf("list hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	hooks, err := pbhook.PbHooks(details.Details)
	if err != nil {
		logs.Errorf("get pb hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListHooksResp{
		Count:   details.Count,
		Details: hooks,
	}
	return resp, nil
}

// UpdateHook update hook.
func (s *Service) UpdateHook(ctx context.Context, req *pbds.UpdateHookReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	spec, err := req.Spec.HookSpec()
	if err != nil {
		logs.Errorf("get hook spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	hook := &table.Hook{
		ID:         req.Id,
		Spec:       spec,
		Attachment: req.Attachment.HookAttachment(),
		Revision: &table.Revision{
			Reviser:   kt.User,
			UpdatedAt: now,
		},
	}
	if err := s.dao.Hook().Update(kt, hook); err != nil {
		logs.Errorf("update hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteHook delete hook.
func (s *Service) DeleteHook(ctx context.Context, req *pbds.DeleteHookReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	hook := &table.Hook{
		ID:         req.Id,
		Attachment: req.Attachment.HookAttachment(),
	}
	if err := s.dao.Hook().Delete(kt, hook); err != nil {
		logs.Errorf("delete hook failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
