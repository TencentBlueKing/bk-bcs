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
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbts "bscp.io/pkg/protocol/core/template-space"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateTemplateSpace create TemplateSpace.
func (s *Service) CreateTemplateSpace(ctx context.Context, req *pbds.CreateTemplateSpaceReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.TemplateSpace().GetByName(kt, req.Attachment.BizId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("templateSpace name %s already exists", req.Spec.Name)
	}

	spec, err := req.Spec.TemplateSpaceSpec()
	if err != nil {
		logs.Errorf("get TemplateSpace spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	TemplateSpace := &table.TemplateSpace{
		Spec:       spec,
		Attachment: req.Attachment.TemplateSpaceAttachment(),
		Revision: &table.Revision{
			Creator:   kt.User,
			Reviser:   kt.User,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	id, err := s.dao.TemplateSpace().Create(kt, TemplateSpace)
	if err != nil {
		logs.Errorf("create TemplateSpace failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListTemplateSpaces list TemplateSpaces.
func (s *Service) ListTemplateSpaces(ctx context.Context, req *pbds.ListTemplateSpacesReq) (*pbds.ListTemplateSpacesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	details, count, err := s.dao.TemplateSpace().List(kt, req.BizId, opt)

	if err != nil {
		logs.Errorf("list TemplateSpace failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	TemplateSpaces, err := pbts.PbTemplateSpaces(details)
	if err != nil {
		logs.Errorf("get pb TemplateSpace failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplateSpacesResp{
		Count:   uint32(count),
		Details: TemplateSpaces,
	}
	return resp, nil
}

// UpdateTemplateSpace update TemplateSpace.
func (s *Service) UpdateTemplateSpace(ctx context.Context, req *pbds.UpdateTemplateSpaceReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	spec, err := req.Spec.TemplateSpaceSpec()
	if err != nil {
		logs.Errorf("get TemplateSpace spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	TemplateSpace := &table.TemplateSpace{
		ID:         req.Id,
		Spec:       spec,
		Attachment: req.Attachment.TemplateSpaceAttachment(),
		Revision: &table.Revision{
			Reviser:   kt.User,
			UpdatedAt: now,
		},
	}
	if err := s.dao.TemplateSpace().Update(kt, TemplateSpace); err != nil {
		logs.Errorf("update TemplateSpace failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteTemplateSpace delete TemplateSpace.
func (s *Service) DeleteTemplateSpace(ctx context.Context, req *pbds.DeleteTemplateSpaceReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	TemplateSpace := &table.TemplateSpace{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateSpaceAttachment(),
	}
	if err := s.dao.TemplateSpace().Delete(kt, TemplateSpace); err != nil {
		logs.Errorf("delete TemplateSpace failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
