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
	"strings"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbtv "bscp.io/pkg/protocol/core/template-variable"
	pbds "bscp.io/pkg/protocol/data-service"
)

// CreateTemplateVariable create a template variable
func (s *Service) CreateTemplateVariable(ctx context.Context, req *pbcs.CreateTemplateVariableReq) (*pbcs.
	CreateTemplateVariableResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateTemplateVariableResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateVariable, Action: meta.Create,
		ResourceID: req.BizId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	if !strings.HasPrefix(strings.ToLower(req.Name), constant.TemplateVariablePrefix) {
		return nil, fmt.Errorf("template variable name must start with %s", constant.TemplateVariablePrefix)
	}

	r := &pbds.CreateTemplateVariableReq{
		Attachment: &pbtv.TemplateVariableAttachment{
			BizId: grpcKit.BizID,
		},
		Spec: &pbtv.TemplateVariableSpec{
			Name:       req.Name,
			Type:       req.Type,
			DefaultVal: req.DefaultVal,
			Memo:       req.Memo,
		},
	}
	rp, err := s.client.DS.CreateTemplateVariable(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create template variable failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateTemplateVariableResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteTemplateVariable delete a template variable
func (s *Service) DeleteTemplateVariable(ctx context.Context, req *pbcs.DeleteTemplateVariableReq) (*pbcs.
	DeleteTemplateVariableResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteTemplateVariableResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateVariable, Action: meta.Delete,
		ResourceID: req.TemplateVariableId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.DeleteTemplateVariableReq{
		Id: req.TemplateVariableId,
		Attachment: &pbtv.TemplateVariableAttachment{
			BizId: grpcKit.BizID,
		},
	}
	if _, err := s.client.DS.DeleteTemplateVariable(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete template variable failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// UpdateTemplateVariable update a template variable
func (s *Service) UpdateTemplateVariable(ctx context.Context, req *pbcs.UpdateTemplateVariableReq) (*pbcs.
	UpdateTemplateVariableResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateTemplateVariableResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateVariable, Action: meta.Update,
		ResourceID: req.TemplateVariableId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.UpdateTemplateVariableReq{
		Id: req.TemplateVariableId,
		Attachment: &pbtv.TemplateVariableAttachment{
			BizId: grpcKit.BizID,
		},
		Spec: &pbtv.TemplateVariableSpec{
			DefaultVal: req.DefaultVal,
			Memo:       req.Memo,
		},
	}
	if _, err := s.client.DS.UpdateTemplateVariable(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update template variable failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListTemplateVariables list template variables
func (s *Service) ListTemplateVariables(ctx context.Context, req *pbcs.ListTemplateVariablesReq) (*pbcs.
	ListTemplateVariablesResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateVariablesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateVariable, Action: meta.Find},
		BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateVariablesReq{
		BizId:        grpcKit.BizID,
		SearchFields: req.SearchFields,
		SearchValue:  req.SearchValue,
		Start:        req.Start,
		Limit:        req.Limit,
		All:          req.All,
	}

	rp, err := s.client.DS.ListTemplateVariables(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateVariablesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
