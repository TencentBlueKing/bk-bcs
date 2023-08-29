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
	pbtv "bscp.io/pkg/protocol/core/template-variable"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/search"
	"bscp.io/pkg/types"
)

// CreateTemplateVariable create template variable.
func (s *Service) CreateTemplateVariable(ctx context.Context, req *pbds.CreateTemplateVariableReq) (*pbds.CreateResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.TemplateVariable().GetByUniqueKey(kt, req.Attachment.BizId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("template variable's same name %s already exists", req.Spec.Name)
	}

	templateVariable := &table.TemplateVariable{
		Spec:       req.Spec.TemplateVariableSpec(),
		Attachment: req.Attachment.TemplateVariableAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	id, err := s.dao.TemplateVariable().Create(kt, templateVariable)
	if err != nil {
		logs.Errorf("create template variable failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListTemplateVariables list template variable.
func (s *Service) ListTemplateVariables(ctx context.Context, req *pbds.ListTemplateVariablesReq) (*pbds.
	ListTemplateVariablesResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.TemplateVariable)
	if err != nil {
		return nil, err
	}

	details, count, err := s.dao.TemplateVariable().List(kt, req.BizId, searcher, opt)

	if err != nil {
		logs.Errorf("list template variables failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplateVariablesResp{
		Count:   uint32(count),
		Details: pbtv.PbTemplateVariables(details),
	}
	return resp, nil
}

// UpdateTemplateVariable update template variable.
func (s *Service) UpdateTemplateVariable(ctx context.Context, req *pbds.UpdateTemplateVariableReq) (*pbbase.EmptyResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	templateVariable := &table.TemplateVariable{
		ID:         req.Id,
		Spec:       req.Spec.TemplateVariableSpec(),
		Attachment: req.Attachment.TemplateVariableAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if err := s.dao.TemplateVariable().Update(kt, templateVariable); err != nil {
		logs.Errorf("update template variable failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteTemplateVariable delete template variable.
func (s *Service) DeleteTemplateVariable(ctx context.Context, req *pbds.DeleteTemplateVariableReq) (*pbbase.EmptyResp,
	error) {
	kt := kit.FromGrpcContext(ctx)

	templateVariable := &table.TemplateVariable{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateVariableAttachment(),
	}
	if err := s.dao.TemplateVariable().Delete(kt, templateVariable); err != nil {
		logs.Errorf("delete template variable failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
