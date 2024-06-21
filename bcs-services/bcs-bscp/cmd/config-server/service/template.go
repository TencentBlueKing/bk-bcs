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
	"sync"

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
	// FromGrpcContext used only to obtain Kit through grpc context.
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	// Authorize authorize if user has permission to the resources.
	// If user is unauthorized, assign apply url and resources into error.
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	// validate input param
	idsLen := len(req.TemplateSetIds)
	if idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template set ids is %d, it must be within the range of [0,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	// validate if file content uploaded.
	metadata, err := s.client.provider.Metadata(grpcKit, req.Sign)
	if err != nil {
		logs.Errorf("validate file content uploaded failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
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
			RevisionName: req.RevisionName,
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
				Md5:       metadata.Md5,
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
	// Authorize authorize if user has permission to the resources.
	// If user is unauthorized, assign apply url and resources into error.
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
	// FromGrpcContext used only to obtain Kit through grpc context.
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	// Authorize authorize if user has permission to the resources.
	// If user is unauthorized, assign apply url and resources into error.
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
	// FromGrpcContext used only to obtain Kit through grpc context.
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	// Authorize authorize if user has permission to the resources.
	// If user is unauthorized, assign apply url and resources into error.
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
	items := make([]*pbds.BatchUpsertTemplatesReq_Item, 0, len(req.Items))
	var mu sync.Mutex
	var g errgroup.Group
	g.SetLimit(constant.MaxConcurrentUpload)
	for _, item := range req.Items {
		i := item
		g.Go(func() error {
			// validate if file content uploaded.
			metadata, err := s.client.provider.Metadata(grpcKit, i.Sign)
			if err != nil {
				logs.Errorf("validate file content uploaded failed, err: %v, rid: %s", err, grpcKit.Rid)
				return err
			}
			mu.Lock()
			items = append(items, &pbds.BatchUpsertTemplatesReq_Item{
				Template: &pbtemplate.Template{
					Id: i.Id,
					Spec: &pbtemplate.TemplateSpec{
						Name: i.Name,
						Path: i.Path,
						Memo: i.Memo,
					},
					Attachment: &pbtemplate.TemplateAttachment{
						BizId:           req.BizId,
						TemplateSpaceId: req.TemplateSpaceId,
					},
				},
				TemplateRevision: &pbtr.TemplateRevision{
					Spec: &pbtr.TemplateRevisionSpec{
						Name:     i.Name,
						Path:     i.Path,
						FileType: i.FileType,
						FileMode: i.FileMode,
						Permission: &pbci.FilePermission{
							User:      i.User,
							UserGroup: i.UserGroup,
							Privilege: i.Privilege,
						},
						ContentSpec: &pbcontent.ContentSpec{
							Signature: i.Sign,
							ByteSize:  i.ByteSize,
							Md5:       metadata.Md5,
						},
					},
					Attachment: &pbtr.TemplateRevisionAttachment{
						BizId:           req.BizId,
						TemplateSpaceId: req.TemplateSpaceId,
					},
				},
			})
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	in := &pbds.BatchUpsertTemplatesReq{Items: items}
	data, err := s.client.DS.BatchUpsertTemplates(grpcKit.RpcCtx(), in)
	if err != nil {
		return nil, err
	}
	resp := &pbcs.BatchUpsertTemplatesResp{Ids: data.Ids}
	return resp, nil
}

// BatchUpdateTemplatePermissions 批量更新模板权限
func (s *Service) BatchUpdateTemplatePermissions(ctx context.Context, req *pbcs.BatchUpdateTemplatePermissionsReq) (
	*pbcs.BatchUpdateTemplatePermissionsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}

	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	if req.User == "" && req.UserGroup == "" && req.Privilege == "" {
		return nil, nil
	}

	resp, err := s.client.DS.BatchUpdateTemplatePermissions(grpcKit.RpcCtx(), &pbds.BatchUpdateTemplatePermissionsReq{
		BizId:       req.BizId,
		TemplateIds: req.TemplateIds,
		User:        req.User,
		UserGroup:   req.UserGroup,
		Privilege:   req.Privilege,
	})

	if err != nil {
		return nil, err
	}

	return &pbcs.BatchUpdateTemplatePermissionsResp{Ids: resp.Ids}, nil
}
