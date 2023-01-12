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
	pbapp "bscp.io/pkg/protocol/core/app"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateApp create app with options
func (s *Service) CreateApp(ctx context.Context, req *pbcs.CreateAppReq) (*pbcs.CreateAppResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateAppResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Create}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		return resp, nil
	}

	r := &pbds.CreateAppReq{
		BizId: req.BizId,
		Spec: &pbapp.AppSpec{
			Name:       req.Name,
			ConfigType: req.ConfigType,
			Mode:       req.Mode,
			Memo:       req.Memo,
			Reload: &pbapp.Reload{
				ReloadType: req.ReloadType,
				FileReloadSpec: &pbapp.FileReloadSpec{
					ReloadFilePath: req.ReloadFilePath,
				},
			},
		},
	}
	rp, err := s.client.DS.CreateApp(kt.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kt, resp)
		logs.Errorf("create app failed, err: %v, rid: %s", err, kt.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.CreateAppResp_RespData{
		Id: rp.Id,
	}
	return resp, nil
}

// UpdateApp update app with options
func (s *Service) UpdateApp(ctx context.Context, req *pbcs.UpdateAppReq) (*pbcs.UpdateAppResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateAppResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Update, ResourceID: req.Id},
		BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, authRes)
	if err != nil {
		return resp, nil
	}

	r := &pbds.UpdateAppReq{
		Id:    req.Id,
		BizId: req.BizId,
		Spec: &pbapp.AppSpec{
			Name: req.Name,
			Memo: req.Memo,
			Reload: &pbapp.Reload{
				ReloadType: req.ReloadType,
				FileReloadSpec: &pbapp.FileReloadSpec{
					ReloadFilePath: req.ReloadFilePath,
				},
			},
		},
	}
	_, err = s.client.DS.UpdateApp(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("update app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	return resp, nil
}

// DeleteApp delete app with options
func (s *Service) DeleteApp(ctx context.Context, req *pbcs.DeleteAppReq) (*pbcs.DeleteAppResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteAppResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Delete, ResourceID: req.Id},
		BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		return resp, nil
	}

	r := &pbds.DeleteAppReq{
		Id:    req.Id,
		BizId: req.BizId,
	}
	_, err = s.client.DS.DeleteApp(kt.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kt, resp)
		logs.Errorf("delete app failed, err: %v, rid: %s", err, kt.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	return resp, nil
}

// ListApps list apps with filter.
func (s *Service) ListApps(ctx context.Context, req *pbcs.ListAppsReq) (*pbcs.ListAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAppsResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		return resp, nil
	}

	if req.Page == nil {
		errf.Error(errf.New(errf.InvalidParameter, "page is null")).AssignResp(kt, resp)
		return resp, nil
	}

	if err = req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		errf.Error(err).AssignResp(kt, resp)
		return resp, nil
	}

	r := &pbds.ListAppsReq{
		BizId:  req.BizId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListApps(kt.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kt, resp)
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.ListAppsResp_RespData{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListAppsRest list apps with rest filter
func (s *Service) ListAppsRest(ctx context.Context, req *pbcs.ListAppsRestReq) (*pbcs.ListAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAppsResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		return resp, nil
	}

	r := &pbds.ListAppsRestReq{
		BizId:    req.BizId,
		Start:    req.Start,
		Limit:    req.Limit,
		Operator: req.Operator,
		Name:     req.Name,
	}
	rp, err := s.client.DS.ListAppsRest(kt.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kt, resp)
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.ListAppsResp_RespData{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
