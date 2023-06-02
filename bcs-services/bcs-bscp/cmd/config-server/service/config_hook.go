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

	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbch "bscp.io/pkg/protocol/core/config-hook"
	pbds "bscp.io/pkg/protocol/data-service"
)

// CreateConfigHook create a ConfigHook
func (s *Service) CreateConfigHook(ctx context.Context, req *pbcs.CreateConfigHookReq) (*pbcs.CreateConfigHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateConfigHookResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Create}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	h := &pbds.CreateConfigHookReq{
		Attachment: &pbch.ConfigHookAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbch.ConfigHookSpec{
			PreHookId:  req.PreHookId,
			PostHookId: req.PostHookId,
		},
	}
	rp, err := s.client.DS.CreateConfigHook(grpcKit.RpcCtx(), h)
	if err != nil {
		logs.Errorf("create ConfigHook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateConfigHookResp{
		Id: rp.Id,
	}
	return resp, nil
}

// UpdateConfigHook update a ConfigHook
func (s *Service) UpdateConfigHook(ctx context.Context, req *pbcs.UpdateConfigHookReq) (*pbcs.UpdateConfigHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateConfigHookResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Update}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.UpdateConfigHookReq{
		Attachment: &pbch.ConfigHookAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbch.ConfigHookSpec{
			PreHookId:  req.PreHookId,
			PostHookId: req.PostHookId,
		},
	}
	if _, err := s.client.DS.UpdateConfigHook(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update ConfigHook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// GetConfigHook get a ConfigHook
func (s *Service) GetConfigHook(ctx context.Context, req *pbcs.GetConfigHookReq) (*pbcs.GetConfigHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.GetConfigHookResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Update}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.GetConfigHookReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}

	hook, err := s.client.DS.GetConfigHook(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get ConfigHook failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.GetConfigHookResp{ConfigHook: hook}, nil

}
