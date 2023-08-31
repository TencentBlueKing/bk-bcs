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
	pbrelease "bscp.io/pkg/protocol/core/release"
	pbds "bscp.io/pkg/protocol/data-service"
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

	r := &pbds.ListReleasesReq{
		BizId:      grpcKit.BizID,
		AppId:      req.AppId,
		Deprecated: req.Deprecated,
		Start:      req.Start,
		Limit:      req.Limit,
		All:        req.All,
		SearchKey:  req.SearchKey,
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

// GetReleaseByName get release by name
func (s *Service) GetReleaseByName(ctx context.Context, req *pbcs.GetReleaseByNameReq) (*pbrelease.Release, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbrelease.Release)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Release, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		return nil, err
	}

	r := &pbds.GetReleaseByNameReq{
		BizId:       req.BizId,
		AppId:       req.AppId,
		ReleaseName: req.ReleaseName,
	}
	rp, err := s.client.DS.GetReleaseByName(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get release by name %s failed, err: %v, rid: %s", req.ReleaseName, err, kt.Rid)
		return nil, err
	}

	return rp, nil
}
