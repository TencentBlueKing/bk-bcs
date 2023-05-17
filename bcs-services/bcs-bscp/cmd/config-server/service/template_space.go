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

	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbts "bscp.io/pkg/protocol/core/template-space"
	pbds "bscp.io/pkg/protocol/data-service"
)

// CreateTemplateSpace create a TemplateSpace
func (s *Service) CreateTemplateSpace(ctx context.Context, req *pbcs.CreateTemplateSpaceReq) (*pbcs.CreateTemplateSpaceResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateTemplateSpaceResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Create,
		ResourceID: req.BizId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.CreateTemplateSpaceReq{
		Attachment: &pbts.TemplateSpaceAttachment{
			BizId: grpcKit.BizID,
		},
		Spec: &pbts.TemplateSpaceSpec{
			Name: req.Name,
			Memo: req.Memo,
		},
	}
	rp, err := s.client.DS.CreateTemplateSpace(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create TemplateSpace failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateTemplateSpaceResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteTemplateSpace delete a TemplateSpace
func (s *Service) DeleteTemplateSpace(ctx context.Context, req *pbcs.DeleteTemplateSpaceReq) (*pbcs.DeleteTemplateSpaceResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteTemplateSpaceResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Delete,
		ResourceID: req.TemplateSpaceId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.DeleteTemplateSpaceReq{
		Id: req.TemplateSpaceId,
		Attachment: &pbts.TemplateSpaceAttachment{
			BizId: grpcKit.BizID,
		},
	}
	if _, err := s.client.DS.DeleteTemplateSpace(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete TemplateSpace failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// UpdateTemplateSpace update a TemplateSpace
func (s *Service) UpdateTemplateSpace(ctx context.Context, req *pbcs.UpdateTemplateSpaceReq) (*pbcs.UpdateTemplateSpaceResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateTemplateSpaceResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Update,
		ResourceID: req.TemplateSpaceId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.UpdateTemplateSpaceReq{
		Id: req.TemplateSpaceId,
		Attachment: &pbts.TemplateSpaceAttachment{
			BizId: grpcKit.BizID,
		},
		Spec: &pbts.TemplateSpaceSpec{
			Name: req.Name,
			Memo: req.Memo,
		},
	}
	if _, err := s.client.DS.UpdateTemplateSpace(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update TemplateSpace failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListTemplateSpaces list TemplateSpaces with filter
func (s *Service) ListTemplateSpaces(ctx context.Context, req *pbcs.ListTemplateSpacesReq) (*pbcs.ListTemplateSpacesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateSpacesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateSpacesReq{
		BizId: grpcKit.BizID,
		Start: req.Start,
		Limit: req.Limit,
	}

	rp, err := s.client.DS.ListTemplateSpaces(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list TemplateSpaces failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateSpacesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
