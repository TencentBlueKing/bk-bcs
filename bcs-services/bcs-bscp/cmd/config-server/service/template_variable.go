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
	"strings"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbtv "bscp.io/pkg/protocol/core/template-variable"
	pbds "bscp.io/pkg/protocol/data-service"
)

// CreateTemplateVariable create a template variable
func (s *Service) CreateTemplateVariable(ctx context.Context, req *pbcs.CreateTemplateVariableReq) (
	*pbcs.CreateTemplateVariableResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	if !strings.HasPrefix(strings.ToLower(req.Name), constant.TemplateVariablePrefix) {
		return nil, errf.Errorf(nil, errf.InvalidArgument, "template variable name must start with %s",
			constant.TemplateVariablePrefix)
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

	resp := &pbcs.CreateTemplateVariableResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteTemplateVariable delete a template variable
func (s *Service) DeleteTemplateVariable(ctx context.Context, req *pbcs.DeleteTemplateVariableReq) (
	*pbcs.DeleteTemplateVariableResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
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

	return &pbcs.DeleteTemplateVariableResp{}, nil
}

// UpdateTemplateVariable update a template variable
func (s *Service) UpdateTemplateVariable(ctx context.Context, req *pbcs.UpdateTemplateVariableReq) (
	*pbcs.UpdateTemplateVariableResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
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

	return &pbcs.UpdateTemplateVariableResp{}, nil
}

// ListTemplateVariables list template variables
func (s *Service) ListTemplateVariables(ctx context.Context, req *pbcs.ListTemplateVariablesReq) (
	*pbcs.ListTemplateVariablesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
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

	resp := &pbcs.ListTemplateVariablesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ImportTemplateVariables import template variables
func (s *Service) ImportTemplateVariables(ctx context.Context, req *pbcs.ImportTemplateVariablesReq) (
	*pbcs.ImportTemplateVariablesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	// validate params
	const whiteSpace string = "white-space"
	if req.Variables == "" {
		return nil, errors.New("variables can't be empty")
	}
	if strings.Contains(req.Separator, "\n") {
		return nil, errors.New("separator can't contain char '\\n'")
	}
	if req.Separator == "" {
		req.Separator = whiteSpace
	}

	vars := make([]*pbtv.TemplateVariableSpec, 0)
	lines := strings.Split(req.Variables, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var fields []string
		if req.Separator == whiteSpace {
			fields = strings.Fields(line)
		} else {
			fields = strings.Split(line, req.Separator)
		}
		// validate variables content
		if len(fields) != 3 && len(fields) != 4 {
			return nil, fmt.Errorf("the line [%s] is not valid, which must be 3 or 4 fields", line)
		}
		if !strings.HasPrefix(strings.ToLower(fields[0]), constant.TemplateVariablePrefix) {
			return nil, fmt.Errorf("template variable name must start with %s", constant.TemplateVariablePrefix)
		}

		v := &pbtv.TemplateVariableSpec{
			Name:       strings.TrimSpace(fields[0]),
			Type:       strings.TrimSpace(fields[1]),
			DefaultVal: strings.TrimSpace(fields[2]),
		}
		if len(fields) == 4 {
			v.Memo = strings.TrimSpace(fields[3])
		}
		vars = append(vars, v)
	}

	r := &pbds.ImportTemplateVariablesReq{
		BizId: req.BizId,
		Specs: vars,
	}
	rp, err := s.client.DS.ImportTemplateVariables(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create template variable failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ImportTemplateVariablesResp{
		VariableCount: rp.VariableCount,
	}
	return resp, nil
}
