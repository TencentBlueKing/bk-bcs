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
	pbbase "bscp.io/pkg/protocol/core/base"
	pbrelease "bscp.io/pkg/protocol/core/release"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// CreateRelease create a release
func (s *Service) CreateRelease(ctx context.Context, req *pbcs.CreateReleaseReq) (*pbcs.CreateReleaseResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateReleaseResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Release, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
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
	rp, err := s.client.DS.CreateRelease(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateReleaseResp{
		Id: rp.Id,
	}
	return resp, nil
}

// ListReleases list releases with options
func (s *Service) ListReleases(ctx context.Context, req *pbcs.ListReleasesReq) (*pbcs.ListReleasesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListReleasesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Release, Action: meta.Find,
		ResourceID: req.AppId}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	if !req.All {
		if req.Start < 0 {
			return nil, errf.New(errf.InvalidParameter, "start has to be greater than 0")
		}

		if req.Limit < 0 {
			return nil, errf.New(errf.InvalidParameter, "limit has to be greater than 0")
		}
	}

	ft := &filter.Expression{
		Op:    filter.Or,
		Rules: []filter.RuleFactory{},
	}
	if req.SearchKey != "" {
		ft.Rules = append(ft.Rules, &filter.AtomRule{
			Field: "name",
			Op:    filter.ContainsInsensitive.Factory(),
			Value: req.SearchKey,
		}, &filter.AtomRule{
			Field: "memo",
			Op:    filter.ContainsInsensitive.Factory(),
			Value: req.SearchKey,
		}, &filter.AtomRule{
			Field: "creator",
			Op:    filter.ContainsInsensitive.Factory(),
			Value: req.SearchKey,
		})
	}
	ftpb, err := ft.MarshalPB()
	if err != nil {
		return nil, err
	}

	page := &pbbase.BasePage{
		Start: req.Start,
		Limit: req.Limit,
	}

	if req.All {
		page = &pbbase.BasePage{
			Start: 0,
			Limit: 0,
		}
	}

	r := &pbds.ListReleasesReq{
		BizId:      grpcKit.BizID,
		AppId:      req.AppId,
		Filter:     ftpb,
		Page:       page,
		Deprecated: req.Deprecated,
	}
	rp, err := s.client.DS.ListReleases(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list releases failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListReleasesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListReleasedConfigItems list a release's configure items.
func (s *Service) ListReleasedConfigItems(ctx context.Context, req *pbcs.ListReleasedConfigItemsReq) (
	*pbcs.ListReleasedConfigItemsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListReleasedConfigItemsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.ReleasedCI, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	if err = req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		return nil, err
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
		logs.Errorf("list releases failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	r := &pbds.ListReleasedCIsReq{
		BizId:  req.BizId,
		Filter: pbFilter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListReleasedConfigItems(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list released config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListReleasedConfigItemsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
