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
	pbds "bscp.io/pkg/protocol/data-service"
)

// ExtractAppTemplateVariables extract app template variables
func (s *Service) ExtractAppTemplateVariables(ctx context.Context, req *pbcs.ExtractAppTemplateVariablesReq) (
	*pbcs.ExtractAppTemplateVariablesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ExtractAppTemplateVariablesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Find},
		BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ExtractAppTemplateVariablesReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}

	rp, err := s.client.DS.ExtractAppTemplateVariables(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("extract app template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ExtractAppTemplateVariablesResp{
		Details: rp.Details,
	}
	return resp, nil
}

// GetAppTemplateVariableReferences get app template variable references
func (s *Service) GetAppTemplateVariableReferences(ctx context.Context, req *pbcs.GetAppTemplateVariableReferencesReq) (
	*pbcs.GetAppTemplateVariableReferencesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.GetAppTemplateVariableReferencesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Find},
		BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.GetAppTemplateVariableReferencesReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}

	rp, err := s.client.DS.GetAppTemplateVariableReferences(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get app template variable references failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.GetAppTemplateVariableReferencesResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListAppTemplateVariables list app template variables
func (s *Service) ListAppTemplateVariables(ctx context.Context, req *pbcs.ListAppTemplateVariablesReq) (
	*pbcs.ListAppTemplateVariablesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAppTemplateVariablesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Find},
		BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListAppTemplateVariablesReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}

	rp, err := s.client.DS.ListAppTemplateVariables(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list app template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListAppTemplateVariablesResp{
		Details: rp.Details,
	}
	return resp, nil
}
