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
	"errors"
	"fmt"
	"time"

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
	if _, err = s.dao.TemplateRelease().CreateWithTx(kt, tx, TemplateRelease); err != nil {
		logs.Errorf("create template release failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 3. add current template to template sets if necessary
	if len(req.TemplateSetIds) > 0 {
		if err = s.dao.TemplateSet().AddTemplateToTemplateSets(kt, id, req.TemplateSetIds); err != nil {
			logs.Errorf("add current template to template sets failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

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

	now := time.Now()
	Template := &table.Template{
		ID:         req.Id,
		Spec:       req.Spec.TemplateSpec(),
		Attachment: req.Attachment.TemplateAttachment(),
		Revision: &table.Revision{
			Reviser:   kt.User,
			UpdatedAt: now,
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

	r := &pbds.ListTemplateBoundCountsReq{
		BizId:           req.Attachment.BizId,
		TemplateSpaceId: req.Attachment.TemplateSpaceId,
		TemplateIds:     []uint32{req.Id},
	}
	boundCnt, err := s.ListTemplateBoundCounts(ctx, r)
	if err != nil {
		logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var hasTmplSet, hasUnnamedApp bool
	if len(boundCnt.Details) > 0 {
		if boundCnt.Details[0].BoundTemplateSetCount > 0 && boundCnt.Details[0].BoundUnnamedAppCount > 0 {
			hasTmplSet, hasUnnamedApp = true, true
			if !req.Force {
				return nil, errors.New("template is bound to template set and unnamed app, please unbind first")
			}
		} else if boundCnt.Details[0].BoundTemplateSetCount > 0 {
			hasTmplSet = true
			if !req.Force {
				return nil, errors.New("template is bound to template set, please unbind first")
			}
		} else if boundCnt.Details[0].BoundUnnamedAppCount > 0 {
			hasUnnamedApp = true
			if !req.Force {
				return nil, errors.New("template is bound to unnamed app, please unbind first")
			}
		}
	}

	tx := s.dao.GenQuery().Begin()

	// 1. delete template
	Template := &table.Template{
		ID:         req.Id,
		Attachment: req.Attachment.TemplateAttachment(),
	}
	if err = s.dao.Template().DeleteWithTx(kt, tx, Template); err != nil {
		logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 2. delete template releases of current template
	if err = s.dao.TemplateRelease().DeleteForTmplWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
		logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback()
		return nil, err
	}

	// 3. delete bound template set if exists
	if hasTmplSet {
		if err = s.dao.TemplateSet().DeleteTmplFromTmplSetsWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
			logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, err
		}
	}

	// 4. delete bound unnamed app if exists
	if hasUnnamedApp {
		if err = s.dao.TemplateBindingRelation().DeleteTmplWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
			logs.Errorf("delete template failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	return new(pbbase.EmptyResp), nil
}

// AddTemplateToTemplateSets add a template to template sets.
func (s *Service) AddTemplateToTemplateSets(ctx context.Context, req *pbds.AddTemplateToTemplateSetsReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := s.dao.Validator().ValidateTemplateExist(kt, req.TemplateId); err != nil {
		return nil, err
	}
	if err := s.dao.Validator().ValidateTemplateSetsExist(kt, req.TemplateSetIds); err != nil {
		return nil, err
	}

	if err := s.dao.TemplateSet().AddTemplateToTemplateSets(kt, req.TemplateId, req.TemplateSetIds); err != nil {
		logs.Errorf(" add template to template sets failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
