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
	"encoding/base64"
	"errors"

	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbhr "bscp.io/pkg/protocol/core/hook-release"
	pbds "bscp.io/pkg/protocol/data-service"
)

// CreateHookRelease create hook release  with option
func (s *Service) CreateHookRelease(ctx context.Context,
	req *pbcs.CreateHookReleaseReq) (*pbcs.CreateHookReleaseResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateHookReleaseResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Create,
		ResourceID: req.BizId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	contentBase64 := base64.StdEncoding.EncodeToString([]byte(req.Content))
	r := &pbds.CreateHookReleaseReq{
		Attachment: &pbhr.HookReleaseAttachment{
			BizId:  grpcKit.BizID,
			HookId: req.HookId,
		},
		Spec: &pbhr.HookReleaseSpec{
			Name:    req.Name,
			Content: contentBase64,
			Memo:    req.Memo,
		},
	}

	rp, err := s.client.DS.CreateHookRelease(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create HookRelease failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateHookReleaseResp{
		Id: rp.Id,
	}
	return resp, nil
}

// ListHookRelease list HookRelease with filter
func (s *Service) ListHookRelease(ctx context.Context, req *pbcs.ListHookReleaseReq) (*pbcs.ListHookReleaseResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListHookReleaseResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListHookReleasesReq{
		HookId:    req.HookId,
		SearchKey: req.SearchKey,
		BizId:     grpcKit.BizID,
		All:       req.All,
		State:     req.State,
	}

	if !req.All {
		if req.Start < 0 {
			return nil, errors.New("start has to be greater than 0")
		}

		if req.Limit < 0 {
			return nil, errors.New("limit has to be greater than 0")
		}

		r.Start = req.Start
		r.Limit = req.Limit
	}

	rp, err := s.client.DS.ListHookReleases(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list HookReleases failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	for _, detail := range rp.Details {
		content, e := base64.StdEncoding.DecodeString(detail.Spec.Content)
		if e != nil {
			logs.Errorf("base64 decode release content failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
		}
		detail.Spec.Content = string(content)
	}

	resp = &pbcs.ListHookReleaseResp{
		Count:   rp.Count,
		Details: rp.Details,
	}

	return resp, nil
}

// DeleteHookRelease delete a HookRelease
func (s *Service) DeleteHookRelease(ctx context.Context,
	req *pbcs.DeleteHookReleaseReq) (*pbcs.DeleteHookReleaseResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteHookReleaseResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Delete,
		ResourceID: req.ReleaseId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.DeleteHookReleaseReq{
		BizId:  req.BizId,
		HookId: req.HookId,
		Id:     req.ReleaseId,
	}

	if _, err := s.client.DS.DeleteHookRelease(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete HookRelease failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// PublishHookRelease publish a release
func (s *Service) PublishHookRelease(ctx context.Context, req *pbcs.PublishHookReleaseReq) (*pbcs.PublishHookReleaseResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.PublishHookReleaseResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Update},
		BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.PublishHookReleaseReq{
		BizId:  req.BizId,
		HookId: req.HookId,
		Id:     req.ReleaseId,
	}

	if _, err := s.client.DS.PublishHookRelease(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("publish HookRelease failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// GetHookRelease get a hookRelease
func (s *Service) GetHookRelease(ctx context.Context, req *pbcs.GetHookReleaseReq) (*pbhr.HookRelease, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbhr.HookRelease)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Update},
		BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.GetHookReleaseByIdReq{
		BizId:  req.BizId,
		HookId: req.HookId,
		Id:     req.ReleaseId,
	}

	release, err := s.client.DS.GetHookReleaseByID(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get HookRelease failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	content, err := base64.StdEncoding.DecodeString(release.Spec.Content)
	if err != nil {
		logs.Errorf("base64 decode release content failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	release.Spec.Content = string(content)

	return release, nil
}

// UpdateHookRelease update a HookRelease
func (s *Service) UpdateHookRelease(ctx context.Context, req *pbcs.UpdateHookReleaseReq) (*pbcs.UpdateHookReleaseResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateHookReleaseResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Update,
		ResourceID: req.ReleaseId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	contentBase64 := base64.StdEncoding.EncodeToString([]byte(req.Content))
	r := &pbds.UpdateHookReleaseReq{
		Id: req.ReleaseId,
		Attachment: &pbhr.HookReleaseAttachment{
			BizId:  req.BizId,
			HookId: req.HookId,
		},
		Spec: &pbhr.HookReleaseSpec{
			Name:    req.Name,
			Content: contentBase64,
			Memo:    req.Memo,
		},
	}
	if _, err := s.client.DS.UpdateHookRelease(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update HookRelease failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil

}

// ListHookReleasesReferences 查询hook版本被引用列表
func (s *Service) ListHookReleasesReferences(ctx context.Context,
	req *pbcs.ListHookReleasesReferencesReq) (*pbcs.ListHookReleasesReferencesResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListHookReleasesReferencesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateSpace, Action: meta.Find}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListHookReleasesReferencesReq{
		BizId:      req.BizId,
		HookId:     req.HookId,
		ReleasesId: req.ReleaseId,
		Limit:      req.Limit,
		Start:      req.Start,
	}

	rp, err := s.client.DS.ListHookReleasesReferences(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list TemplateSpaces failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListHookReleasesReferencesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}

	return resp, nil

}
