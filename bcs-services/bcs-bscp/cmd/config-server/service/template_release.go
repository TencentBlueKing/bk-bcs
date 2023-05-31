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
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbcontent "bscp.io/pkg/protocol/core/content"
	pbtr "bscp.io/pkg/protocol/core/template-release"
	pbds "bscp.io/pkg/protocol/data-service"
)

// CreateTemplateRelease create a template release
func (s *Service) CreateTemplateRelease(ctx context.Context, req *pbcs.CreateTemplateReleaseReq) (*pbcs.CreateTemplateReleaseResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateTemplateReleaseResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateRelease, Action: meta.Create,
		ResourceID: req.TemplateId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.CreateTemplateReleaseReq{
		Attachment: &pbtr.TemplateReleaseAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
			TemplateId:      req.TemplateId,
		},
		Spec: &pbtr.TemplateReleaseSpec{
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
	}
	rp, err := s.client.DS.CreateTemplateRelease(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create template release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateTemplateReleaseResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteTemplateRelease delete a template release
func (s *Service) DeleteTemplateRelease(ctx context.Context, req *pbcs.DeleteTemplateReleaseReq) (*pbcs.DeleteTemplateReleaseResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteTemplateReleaseResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateRelease, Action: meta.Delete,
		ResourceID: req.TemplateReleaseId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.DeleteTemplateReleaseReq{
		Id: req.TemplateReleaseId,
		Attachment: &pbtr.TemplateReleaseAttachment{
			BizId:           grpcKit.BizID,
			TemplateSpaceId: req.TemplateSpaceId,
			TemplateId:      req.TemplateId,
		},
	}
	if _, err := s.client.DS.DeleteTemplateRelease(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete template release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListTemplateReleases list template releases
func (s *Service) ListTemplateReleases(ctx context.Context, req *pbcs.ListTemplateReleasesReq) (*pbcs.ListTemplateReleasesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateReleasesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateRelease, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateReleasesReq{
		BizId:           grpcKit.BizID,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		Start:           req.Start,
		Limit:           req.Limit,
	}

	rp, err := s.client.DS.ListTemplateReleases(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template releases failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateReleasesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
