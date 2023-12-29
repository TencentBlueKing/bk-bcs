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
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbcontent "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	pbtemplate "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template"
	pbtr "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-revision"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// CreateTemplate create a template
func (s *Service) CreateTemplate(ctx context.Context, req *pbcs.CreateTemplateReq) (*pbcs.CreateTemplateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	// validate input param
	idsLen := len(req.TemplateSetIds)
	if idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template set ids is %d, it must be within the range of [0,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	r := &pbds.CreateTemplateReq{
		Attachment: &pbtemplate.TemplateAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
		},
		Spec: &pbtemplate.TemplateSpec{
			Name: req.Name,
			Path: req.Path,
			Memo: req.Memo,
		},
		TrSpec: &pbtr.TemplateRevisionSpec{
			RevisionMemo: req.RevisionMemo,
			Name:         req.Name,
			Path:         req.Path,
			FileType:     req.FileType,
			FileMode:     req.FileMode,
			Permission: &pbci.FilePermission{
				User:      req.User,
				UserGroup: req.UserGroup,
				Privilege: req.Privilege,
			},
			ContentSpec: &pbcontent.ContentSpec{
				Signature: req.Sign,
				ByteSize:  req.ByteSize,
			},
		},
		TemplateSetIds: req.TemplateSetIds,
	}
	rp, err := s.client.DS.CreateTemplate(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create template failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateTemplateResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteTemplate delete a template
func (s *Service) DeleteTemplate(ctx context.Context, req *pbcs.DeleteTemplateReq) (*pbcs.DeleteTemplateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.DeleteTemplateReq{
		Id: req.TemplateId,
		Attachment: &pbtemplate.TemplateAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
		},
		Force: req.Force,
	}
	if _, err := s.client.DS.DeleteTemplate(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete template failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.DeleteTemplateResp{}, nil
}

// BatchDeleteTemplate delete templates in batch
func (s *Service) BatchDeleteTemplate(ctx context.Context, req *pbcs.BatchDeleteTemplateReq) (
	*pbcs.BatchDeleteTemplateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	templateIDs, err := tools.GetUint32List(req.TemplateIds)
	if err != nil {
		return nil, fmt.Errorf("invalid template ids, %s", err)
	}
	idsLen := len(templateIDs)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.BatchDeleteTemplateReq{
		Ids: templateIDs,
		Attachment: &pbtemplate.TemplateAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
		},
		Force: req.Force,
	}
	if _, err := s.client.DS.BatchDeleteTemplate(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("batch delete template failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.BatchDeleteTemplateResp{}, nil
}

// UpdateTemplate update a template
func (s *Service) UpdateTemplate(ctx context.Context, req *pbcs.UpdateTemplateReq) (*pbcs.UpdateTemplateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.UpdateTemplateReq{
		Id: req.TemplateId,
		Attachment: &pbtemplate.TemplateAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
		},
		Spec: &pbtemplate.TemplateSpec{
			Memo: req.Memo,
		},
	}
	if _, err := s.client.DS.UpdateTemplate(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update template failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UpdateTemplateResp{}, nil
}

// ListTemplates list templates
func (s *Service) ListTemplates(ctx context.Context, req *pbcs.ListTemplatesReq) (*pbcs.ListTemplatesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplatesReq{
		BizId:           grpcKit.BizID,
		TemplateSpaceId: req.TemplateSpaceId,
		SearchFields:    req.SearchFields,
		SearchValue:     req.SearchValue,
		Start:           req.Start,
		Limit:           req.Limit,
		Ids:             req.Ids,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTemplates(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list templates failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTemplatesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// AddTmplsToTmplSets add templates to template sets
func (s *Service) AddTmplsToTmplSets(ctx context.Context, req *pbcs.AddTmplsToTmplSetsReq) (
	*pbcs.AddTmplsToTmplSetsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	idsLen := len(req.TemplateIds)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	idsLen2 := len(req.TemplateSetIds)
	if idsLen2 == 0 || idsLen2 > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template set ids is %d, it must be within the range of [1,%d]",
			idsLen2, constant.ArrayInputLenLimit)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.AddTmplsToTmplSetsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateIds:     req.TemplateIds,
		TemplateSetIds:  req.TemplateSetIds,
	}

	if _, err := s.client.DS.AddTmplsToTmplSets(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update template failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.AddTmplsToTmplSetsResp{}, nil
}

// DeleteTmplsFromTmplSets delete templates from template sets
func (s *Service) DeleteTmplsFromTmplSets(ctx context.Context, req *pbcs.DeleteTmplsFromTmplSetsReq) (
	*pbcs.DeleteTmplsFromTmplSetsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	idsLen := len(req.TemplateIds)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	idsLen2 := len(req.TemplateSetIds)
	if idsLen2 == 0 || idsLen2 > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template set ids is %d, it must be within the range of [1,%d]",
			idsLen2, constant.ArrayInputLenLimit)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.DeleteTmplsFromTmplSetsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateIds:     req.TemplateIds,
		TemplateSetIds:  req.TemplateSetIds,
	}

	if _, err := s.client.DS.DeleteTmplsFromTmplSets(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update template failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.DeleteTmplsFromTmplSetsResp{}, nil
}

// ListTemplatesByIDs list templates by ids
func (s *Service) ListTemplatesByIDs(ctx context.Context, req *pbcs.ListTemplatesByIDsReq) (
	*pbcs.ListTemplatesByIDsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	ids := tools.SliceRepeatedElements(req.Ids)
	if len(ids) > 0 {
		return nil, fmt.Errorf("repeated ids: %v, id must be unique", ids)
	}
	idsLen := len(req.Ids)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplatesByIDsReq{
		Ids: req.Ids,
	}

	rp, err := s.client.DS.ListTemplatesByIDs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list templates failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTemplatesByIDsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplatesNotBound list templates not bound
func (s *Service) ListTemplatesNotBound(ctx context.Context, req *pbcs.ListTemplatesNotBoundReq) (
	*pbcs.ListTemplatesNotBoundResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplatesNotBoundReq{
		BizId:           grpcKit.BizID,
		TemplateSpaceId: req.TemplateSpaceId,
		SearchFields:    req.SearchFields,
		SearchValue:     req.SearchValue,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTemplatesNotBound(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list templates not bound failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTemplatesNotBoundResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplsOfTmplSet list templates of template set
func (s *Service) ListTmplsOfTmplSet(ctx context.Context, req *pbcs.ListTmplsOfTmplSetReq) (
	*pbcs.ListTmplsOfTmplSetResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplsOfTmplSetReq{
		BizId:           grpcKit.BizID,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateSetId:   req.TemplateSetId,
		SearchFields:    req.SearchFields,
		SearchValue:     req.SearchValue,
		Start:           req.Start,
		Limit:           req.Limit,
		Ids:             req.Ids,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTmplsOfTmplSet(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list templates of template set failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplsOfTmplSetResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateByTuple 按照多个字段in查询
func (s *Service) ListTemplateByTuple(ctx context.Context, req *pbcs.ListTemplateByTupleReq) (
	*pbcs.ListTemplateByTupleResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}
	data := []*pbds.ListTemplateByTupleReq_Item{}

	for _, item := range req.Items {
		data = append(data, &pbds.ListTemplateByTupleReq_Item{
			BizId:           req.BizId,
			TemplateSpaceId: req.TemplateSpaceId,
			Name:            item.Name,
			Path:            item.Path,
		})
	}
	tuple, err := s.client.DS.ListTemplateByTuple(grpcKit.RpcCtx(), &pbds.ListTemplateByTupleReq{Items: data})
	if err != nil {
		return nil, err
	}
	templatesData := []*pbcs.ListTemplateByTupleResp_Item{}
	for _, item := range tuple.Items {
		templatesData = append(templatesData,
			&pbcs.ListTemplateByTupleResp_Item{
				Template:         item.Template,
				TemplateRevision: item.TemplateRevision,
			})
	}
	resp := &pbcs.ListTemplateByTupleResp{Items: templatesData}
	return resp, nil
}

// BatchUpsertTemplates batch upsert templates.
func (s *Service) BatchUpsertTemplates(ctx context.Context, req *pbcs.BatchUpsertTemplatesReq) (
	*pbcs.BatchUpsertTemplatesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}
	var g errgroup.Group
	g.SetLimit(constant.MaxConcurrentUpload)
	for _, item := range req.Items {
		sign := item.Sign
		g.Go(func() error {
			// validate if file content uploaded.
			if err := s.validateContentExist(grpcKit, req.BizId, sign); err != nil {
				logs.Errorf("validate file content uploaded failed, err: %v, rid: %s", err, grpcKit.Rid)
				return err
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	items := make([]*pbds.BatchUpsertTemplatesReq_Item, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &pbds.BatchUpsertTemplatesReq_Item{
			Template: &pbtemplate.Template{
				Id: item.Id,
				Spec: &pbtemplate.TemplateSpec{
					Name: item.Name,
					Path: item.Path,
					Memo: item.Memo,
				},
				Attachment: &pbtemplate.TemplateAttachment{
					BizId:           req.BizId,
					TemplateSpaceId: req.TemplateSpaceId,
				},
			},
			TemplateRevision: &pbtr.TemplateRevision{
				Spec: &pbtr.TemplateRevisionSpec{
					Name:     item.Name,
					Path:     item.Path,
					FileType: item.FileType,
					FileMode: item.FileMode,
					Permission: &pbci.FilePermission{
						User:      item.User,
						UserGroup: item.UserGroup,
						Privilege: item.Privilege,
					},
					ContentSpec: &pbcontent.ContentSpec{
						Signature: item.Sign,
						ByteSize:  item.ByteSize,
					},
				},
				Attachment: &pbtr.TemplateRevisionAttachment{
					BizId:           req.BizId,
					TemplateSpaceId: req.TemplateSpaceId,
				},
			},
		})
	}
	in := &pbds.BatchUpsertTemplatesReq{Items: items}
	data, err := s.client.DS.BatchUpsertTemplates(grpcKit.RpcCtx(), in)
	if err != nil {
		return nil, err
	}
	resp := &pbcs.BatchUpsertTemplatesResp{Ids: data.Ids}
	return resp, nil
}
