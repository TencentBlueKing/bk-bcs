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
	pbtr "bscp.io/pkg/protocol/core/template-revision"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/tools"
)

// CreateTemplateRevision create a template Revision
func (s *Service) CreateTemplateRevision(ctx context.Context, req *pbcs.CreateTemplateRevisionReq) (*pbcs.CreateTemplateRevisionResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateTemplateRevisionResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateRevision, Action: meta.Create,
		ResourceID: req.TemplateId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.CreateTemplateRevisionReq{
		Attachment: &pbtr.TemplateRevisionAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
			TemplateId:      req.TemplateId,
		},
		Spec: &pbtr.TemplateRevisionSpec{
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
	}
	rp, err := s.client.DS.CreateTemplateRevision(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create template Revision failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateTemplateRevisionResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteTemplateRevision delete a template Revision
func (s *Service) DeleteTemplateRevision(ctx context.Context, req *pbcs.DeleteTemplateRevisionReq) (*pbcs.DeleteTemplateRevisionResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteTemplateRevisionResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateRevision, Action: meta.Delete,
		ResourceID: req.TemplateRevisionId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.DeleteTemplateRevisionReq{
		Id: req.TemplateRevisionId,
		Attachment: &pbtr.TemplateRevisionAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
			TemplateId:      req.TemplateId,
		},
	}
	if _, err := s.client.DS.DeleteTemplateRevision(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete template Revision failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListTemplateRevisions list template Revisions
func (s *Service) ListTemplateRevisions(ctx context.Context, req *pbcs.ListTemplateRevisionsReq) (*pbcs.ListTemplateRevisionsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateRevisionsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateRevision, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateRevisionsReq{
		BizId:           grpcKit.BizID,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		SearchFields:    req.SearchFields,
		SearchValue:     req.SearchValue,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTemplateRevisions(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template Revisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateRevisionsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateRevisionsByIDs list template Revisions by ids
func (s *Service) ListTemplateRevisionsByIDs(ctx context.Context, req *pbcs.ListTemplateRevisionsByIDsReq) (*pbcs.
	ListTemplateRevisionsByIDsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateRevisionsByIDsResp)

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

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateRevision, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateRevisionsByIDsReq{
		Ids: req.Ids,
	}

	rp, err := s.client.DS.ListTemplateRevisionsByIDs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template Revisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateRevisionsByIDsResp{
		Details: rp.Details,
	}
	return resp, nil
}
