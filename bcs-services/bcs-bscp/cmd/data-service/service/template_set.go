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
	pbtset "bscp.io/pkg/protocol/core/template-set"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateTemplateSet create template set.
func (s *Service) CreateTemplateSet(ctx context.Context, req *pbds.CreateTemplateSetReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.TemplateSet().GetByUniqueKey(
		kt, req.Attachment.BizId, req.Attachment.TemplateSpaceId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("template set's same name %s already exists", req.Spec.Name)
	}

	TemplateSet := &table.TemplateSet{
		Spec:       req.Spec.TemplateSetSpec(),
		Attachment: req.Attachment.TemplateSetAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	id, err := s.dao.TemplateSet().Create(kt, TemplateSet)
	if err != nil {
		logs.Errorf("create template set failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListTemplateSets list template set.
func (s *Service) ListTemplateSets(ctx context.Context, req *pbds.ListTemplateSetsReq) (*pbds.ListTemplateSetsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	details, count, err := s.dao.TemplateSet().List(kt, req.BizId, req.TemplateSpaceId, opt)

	if err != nil {
		logs.Errorf("list template sets failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplateSetsResp{
		Count:   uint32(count),
		Details: pbtset.PbTemplateSets(details),
	}
	return resp, nil
}

// UpdateTemplateSet update template set.
func (s *Service) UpdateTemplateSet(ctx context.Context, req *pbds.UpdateTemplateSetReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	TemplateSet := &table.TemplateSet{
		ID:         req.Id,
		Spec:       req.Spec.TemplateSetSpec(),
		Attachment: req.Attachment.TemplateSetAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if err := s.dao.TemplateSet().Update(kt, TemplateSet); err != nil {
		logs.Errorf("update template set failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteTemplateSet delete template set.
func (s *Service) DeleteTemplateSet(ctx context.Context, req *pbds.DeleteTemplateSetReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	TemplateSet := &table.TemplateSet{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateSetAttachment(),
	}
	if err := s.dao.TemplateSet().Delete(kt, TemplateSet); err != nil {
		logs.Errorf("delete template set failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
