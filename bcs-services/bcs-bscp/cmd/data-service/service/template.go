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
	pbtemplate "bscp.io/pkg/protocol/core/template"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateTemplate create template.
func (s *Service) CreateTemplate(ctx context.Context, req *pbds.CreateTemplateReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.Template().GetByUniqueKey(kt, req.Attachment.BizId, req.Attachment.TemplateSpaceId,
		req.Spec.Name, req.Spec.Path); err == nil {
		return nil, fmt.Errorf("template's same name %s and path %s already exists", req.Spec.Name, req.Spec.Path)
	}

	tx := s.dao.GenQuery().Begin()

	// 1. create template
	template := &table.Template{
		Spec:       req.Spec.TemplateSpec(),
		Attachment: req.Attachment.TemplateAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	id, err := s.dao.Template().CreateWithTx(kt, tx, template)
	if err != nil {
		logs.Errorf("create template failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 2. create template release
	TemplateRelease := &table.TemplateRelease{
		Spec: req.TrSpec.TemplateReleaseSpec(),
		Attachment: &table.TemplateReleaseAttachment{
			BizID:           template.Attachment.BizID,
			TemplateSpaceID: template.Attachment.TemplateSpaceID,
			TemplateID:      id,
		},
		Revision: &table.CreatedRevision{
			Creator: kt.User,
		},
	}
	_, err = s.dao.TemplateRelease().CreateWithTx(kt, tx, TemplateRelease)
	if err != nil {
		logs.Errorf("create template release failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListTemplates list templates.
func (s *Service) ListTemplates(ctx context.Context, req *pbds.ListTemplatesReq) (*pbds.ListTemplatesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	details, count, err := s.dao.Template().List(kt, req.BizId, req.TemplateSpaceId, opt)

	if err != nil {
		logs.Errorf("list template failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListTemplatesResp{
		Count:   uint32(count),
		Details: pbtemplate.PbTemplates(details),
	}
	return resp, nil
}

// UpdateTemplate update template.
func (s *Service) UpdateTemplate(ctx context.Context, req *pbds.UpdateTemplateReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	Template := &table.Template{
		ID:         req.Id,
		Spec:       req.Spec.TemplateSpec(),
		Attachment: req.Attachment.TemplateAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if err := s.dao.Template().Update(kt, Template); err != nil {
		logs.Errorf("update template failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteTemplate delete template.
func (s *Service) DeleteTemplate(ctx context.Context, req *pbds.DeleteTemplateReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	Template := &table.Template{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateAttachment(),
	}
	if err := s.dao.Template().Delete(kt, Template); err != nil {
		logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
