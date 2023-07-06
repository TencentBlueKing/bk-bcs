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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbhook "bscp.io/pkg/protocol/core/hook"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateHook create a hook
func (s *Service) CreateHook(ctx context.Context, req *pbcs.CreateHookReq) (*pbcs.CreateHookResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateHookResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Hook, Action: meta.Create,
		ResourceID: req.AppId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.CreateHookReq{
		Attachment: &pbhook.HookAttachment{
			BizId:     grpcKit.BizID,
			AppId:     req.AppId,
			ReleaseId: req.ReleaseId,
		},
		Spec: &pbhook.HookSpec{
			Name:     req.Name,
			PreType:  req.PreType,
			PreHook:  req.PreHook,
			PostType: req.PostType,
			PostHook: req.PostHook,
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
		Id: req.HookId,
		Attachment: &pbhook.HookAttachment{
			BizId: grpcKit.BizID,
			AppId: req.AppId,
		},
	}
	if _, err := s.client.DS.DeleteHook(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// UpdateHook update a hook
func (s *Service) UpdateHook(ctx context.Context, req *pbcs.UpdateHookReq) (*pbcs.UpdateHookResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateHookResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Hook, Action: meta.Update,
		ResourceID: req.HookId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.UpdateHookReq{
		Id: req.HookId,
		Attachment: &pbhook.HookAttachment{
			BizId:     grpcKit.BizID,
			AppId:     req.AppId,
			ReleaseId: req.ReleaseId,
		},
		Spec: &pbhook.HookSpec{
			Name:     req.Name,
			PreType:  req.PreType,
			PreHook:  req.PreHook,
			PostType: req.PostType,
			PostHook: req.PostHook,
		},
	}
	if _, err := s.client.DS.UpdateHook(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update hook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListHooks list hooks with filter
func (s *Service) ListHooks(ctx context.Context, req *pbcs.ListHooksReq) (*pbcs.ListHooksResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListHooksResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Hook, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	if req.Page == nil {
		return nil, errf.New(errf.InvalidParameter, "page is null")
	}

	if err := req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	r := &pbds.ListHooksReq{
		BizId:  grpcKit.BizID,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}

	rp, err := s.client.DS.ListHooks(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list hooks failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListHooksResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
