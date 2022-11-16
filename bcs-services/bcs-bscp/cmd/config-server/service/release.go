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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbrelease "bscp.io/pkg/protocol/core/release"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// CreateRelease create a release
func (s *Service) CreateRelease(ctx context.Context, req *pbcs.CreateReleaseReq) (*pbcs.CreateReleaseResp, error) {
	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateReleaseResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Release, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, res)
	if err != nil {
		return resp, nil
	}

	r := &pbds.CreateReleaseReq{
		Attachment: &pbrelease.ReleaseAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbrelease.ReleaseSpec{
			Name: req.Name,
			Memo: req.Memo,
		},
	}
	rp, err := s.client.DS.CreateRelease(kit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("create release failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.CreateReleaseResp_RespData{
		Id: rp.Id,
	}
	return resp, nil
}

// ListReleases list releases with options
func (s *Service) ListReleases(ctx context.Context, req *pbcs.ListReleasesReq) (*pbcs.ListReleasesResp, error) {
	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListReleasesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Release, Action: meta.Find,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, res)
	if err != nil {
		return resp, nil
	}

	if req.Page == nil {
		errf.Error(errf.New(errf.InvalidParameter, "page is null")).AssignResp(kit, resp)
		return resp, nil
	}

	if err := req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		errf.Error(err).AssignResp(kit, resp)
		return resp, nil
	}

	r := &pbds.ListReleasesReq{
		BizId:  req.BizId,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListReleases(kit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("list releases failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.ListReleasesResp_RespData{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListReleasedConfigItems list a release's configure items.
func (s *Service) ListReleasedConfigItems(ctx context.Context, req *pbcs.ListReleasedConfigItemsReq) (
	*pbcs.ListReleasedConfigItemsResp, error) {
	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListReleasedConfigItemsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ReleasedCI, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, res)
	if err != nil {
		return resp, nil
	}

	if err := req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		errf.Error(err).AssignResp(kit, resp)
		return resp, nil
	}

	// build query filter.
	ft := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "release_id",
				Op:    filter.Equal.Factory(),
				Value: req.ReleaseId,
			},
		},
	}
	pbFilter, err := ft.MarshalPB()
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("list releases failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	r := &pbds.ListReleasedCIsReq{
		BizId:  req.BizId,
		Filter: pbFilter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListReleasedConfigItems(kit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("list released config items failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.ListReleasedConfigItemsResp_RespData{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
