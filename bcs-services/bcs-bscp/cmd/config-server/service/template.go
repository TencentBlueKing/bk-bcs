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

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbcontent "bscp.io/pkg/protocol/core/content"
	pbtemplate "bscp.io/pkg/protocol/core/template"
	pbtr "bscp.io/pkg/protocol/core/template-release"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/tools"
)

// CreateTemplate create a template
func (s *Service) CreateTemplate(ctx context.Context, req *pbcs.CreateTemplateReq) (*pbcs.CreateTemplateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateTemplateResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Create,
		ResourceID: req.TemplateSpaceId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	idsLen := len(req.TemplateSetIds)
	if idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template release ids is %d, it must be within the range of [1,%d]",
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
		TrSpec: &pbtr.TemplateReleaseSpec{
			ReleaseName: req.ReleaseName,
			ReleaseMemo: req.ReleaseMemo,
			Name:        req.Name,
			Path:        req.Path,
			FileType:    req.FileType,
			FileMode:    req.FileMode,
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

	resp = &pbcs.CreateTemplateResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteTemplate delete a template
func (s *Service) DeleteTemplate(ctx context.Context, req *pbcs.DeleteTemplateReq) (*pbcs.DeleteTemplateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteTemplateResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Delete,
		ResourceID: req.TemplateId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
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

	return resp, nil
}

// UpdateTemplate update a template
func (s *Service) UpdateTemplate(ctx context.Context, req *pbcs.UpdateTemplateReq) (*pbcs.UpdateTemplateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateTemplateResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Update,
		ResourceID: req.TemplateId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
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

	return resp, nil
}

// ListTemplates list templates
func (s *Service) ListTemplates(ctx context.Context, req *pbcs.ListTemplatesReq) (*pbcs.ListTemplatesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplatesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplatesReq{
		BizId:           grpcKit.BizID,
		TemplateSpaceId: req.TemplateSpaceId,
		SearchKey:       req.SearchKey,
		Start:           req.Start,
		Limit:           req.Limit,
	}

	rp, err := s.client.DS.ListTemplates(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list templates failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplatesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// AddTemplateToTemplateSets add template to template sets
func (s *Service) AddTemplateToTemplateSets(ctx context.Context, req *pbcs.AddTemplateToTemplateSetsReq) (
	*pbcs.AddTemplateToTemplateSetsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.AddTemplateToTemplateSetsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Update,
		ResourceID: req.TemplateId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	idsLen := len(req.TemplateSetIds)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template release ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	r := &pbds.AddTemplateToTemplateSetsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		TemplateSetIds:  req.TemplateSetIds,
	}

	if _, err := s.client.DS.AddTemplateToTemplateSets(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update template failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListTemplatesByIDs list templates by ids
func (s *Service) ListTemplatesByIDs(ctx context.Context, req *pbcs.ListTemplatesByIDsReq) (*pbcs.
	ListTemplatesByIDsResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplatesByIDsResp)

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

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
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

	resp = &pbcs.ListTemplatesByIDsResp{
		Details: rp.Details,
	}
	return resp, nil
}
