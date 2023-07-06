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
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbas "bscp.io/pkg/protocol/auth-server"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbapp "bscp.io/pkg/protocol/core/app"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/space"
	"bscp.io/pkg/types"
)

// CreateApp create app with options
func (s *Service) CreateApp(ctx context.Context, req *pbcs.CreateAppReq) (*pbcs.CreateAppResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateAppResp)

	if err := req.Validate(); err != nil {
		return nil, err
	}

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Create}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		return nil, err
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
		logs.Errorf("create app failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp = &pbcs.CreateAppResp{Id: rp.Id}
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
		return nil, err
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
		logs.Errorf("update app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

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
		return nil, err
	}

	r := &pbds.DeleteAppReq{
		Id:    req.Id,
		BizId: req.BizId,
	}
	_, err = s.client.DS.DeleteApp(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("delete app failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp, nil
}

// ListApps list apps with filter.
func (s *Service) ListApps(ctx context.Context, req *pbcs.ListAppsReq) (*pbcs.ListAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAppsResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		return nil, err
	}

	if req.Page == nil {
		return nil, errf.New(errf.InvalidParameter, "page is null")
	}

	if err = req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	r := &pbds.ListAppsReq{
		BizId:  req.BizId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListApps(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp = &pbcs.ListAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// GetApp get app with app id
func (s *Service) GetApp(ctx context.Context, req *pbcs.GetAppReq) (*pbapp.App, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbapp.App)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		return nil, err
	}

	r := &pbds.GetAppReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}
	rp, err := s.client.DS.GetApp(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rp, nil
}

// GetAppByName get app by app name
func (s *Service) GetAppByName(ctx context.Context, req *pbcs.GetAppByNameReq) (*pbapp.App, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbapp.App)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		return nil, err
	}

	r := &pbds.GetAppByNameReq{
		BizId:   req.BizId,
		AppName: req.AppName,
	}
	rp, err := s.client.DS.GetAppByName(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return rp, nil
}

// ListAppsRest list apps with rest filter
func (s *Service) ListAppsRest(ctx context.Context, req *pbcs.ListAppsRestReq) (*pbcs.ListAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAppsResp)

	userSpaceResp, err := s.client.AS.ListUserSpace(kt.RpcCtx(), &pbas.ListUserSpaceReq{})
	if err != nil {
		return nil, err
	}

	if len(userSpaceResp.GetItems()) == 0 {
		return nil, errors.New("use have no spaces")
	}

	spaceMap := map[string]*pbas.Space{}
	spaceIdList := []string{}
	for _, s := range userSpaceResp.GetItems() {
		spaceMap[s.SpaceId] = s
		spaceIdList = append(spaceIdList, s.SpaceId)
	}

	r := &pbds.ListAppsRestReq{
		BizId:    strings.Join(spaceIdList, ","),
		Start:    req.Start,
		Limit:    req.Limit,
		Operator: req.Operator,
		Name:     req.Name,
	}
	rp, err := s.client.DS.ListAppsRest(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// 只填写当前页的space
	for _, app := range rp.Details {
		id := strconv.Itoa(int(app.BizId))
		sp, ok := spaceMap[id]
		if !ok {
			app.SpaceId = id
			app.SpaceName = ""
			app.SpaceTypeId = ""
			app.SpaceTypeName = ""
		} else {
			app.SpaceId = id
			app.SpaceName = sp.SpaceName
			app.SpaceTypeId = sp.SpaceTypeId
			app.SpaceTypeName = sp.SpaceTypeName
		}
	}

	resp = &pbcs.ListAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListAppsBySpaceRest list apps with rest filter
func (s *Service) ListAppsBySpaceRest(ctx context.Context, req *pbcs.ListAppsBySpaceRestReq) (*pbcs.ListAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAppsResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.App, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kt, resp, authRes)
	if err != nil {
		return nil, err
	}

	r := &pbds.ListAppsRestReq{
		BizId:    strconv.Itoa(int(req.BizId)),
		Start:    req.Start,
		Limit:    req.Limit,
		Operator: req.Operator,
		Name:     req.Name,
	}
	rp, err := s.client.DS.ListAppsRest(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	spaceUidMap := map[string]struct{}{}
	for _, app := range rp.Details {
		uid := space.BuildSpaceUid(space.BK_CMDB, strconv.Itoa(int(app.BizId)))
		spaceUidMap[uid] = struct{}{}

	}
	querySpaceReq := &pbas.QuerySpaceReq{SpaceUid: []string{}}
	for spaceUid := range spaceUidMap {
		querySpaceReq.SpaceUid = append(querySpaceReq.SpaceUid, spaceUid)
	}

	querySpaceResp, err := s.client.AS.QuerySpace(ctx, querySpaceReq)
	if err != nil {
		return nil, errors.Wrap(err, "QuerySpace")
	}

	spaceMap := map[string]*pbas.Space{}
	for _, s := range querySpaceResp.GetItems() {
		spaceMap[s.SpaceId] = s
	}

	// 只填写当前页的space
	for _, app := range rp.Details {
		id := strconv.Itoa(int(app.BizId))
		sp, ok := spaceMap[id]
		if !ok {
			app.SpaceId = id
			app.SpaceName = ""
			app.SpaceTypeId = ""
			app.SpaceTypeName = ""
		} else {
			app.SpaceId = id
			app.SpaceName = sp.SpaceName
			app.SpaceTypeId = sp.SpaceTypeId
			app.SpaceTypeName = sp.SpaceTypeName
		}
	}

	resp = &pbcs.ListAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
