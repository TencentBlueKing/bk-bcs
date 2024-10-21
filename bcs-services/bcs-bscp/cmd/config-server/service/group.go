/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbapp "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/app"
	pbgroup "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/group"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// CreateGroup create a group
func (s *Service) CreateGroup(ctx context.Context, req *pbcs.CreateGroupReq) (*pbcs.CreateGroupResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.CreateGroupReq{
		Attachment: &pbgroup.GroupAttachment{
			BizId: req.BizId,
		},
		Spec: &pbgroup.GroupSpec{
			Name:     req.Name,
			Public:   req.Public,
			BindApps: req.BindApps,
			Mode:     req.Mode,
			Selector: req.Selector,
			Uid:      req.Uid,
		},
	}
	rp, err := s.client.DS.CreateGroup(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create group failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateGroupResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteGroup delete a group
func (s *Service) DeleteGroup(ctx context.Context, req *pbcs.DeleteGroupReq) (*pbcs.DeleteGroupResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteGroupResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.DeleteGroupReq{
		Id: req.GroupId,
		Attachment: &pbgroup.GroupAttachment{
			BizId: req.BizId,
		},
	}
	_, err = s.client.DS.DeleteGroup(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("delete group failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// BatchDeleteGroups batch delete groups
func (s *Service) BatchDeleteGroups(ctx context.Context, req *pbcs.BatchDeleteBizResourcesReq) (
	*pbcs.BatchDeleteResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	if len(req.GetIds()) == 0 {
		return nil, errf.Errorf(errf.InvalidArgument, i18n.T(grpcKit, "id is required"))
	}

	eg, egCtx := errgroup.WithContext(grpcKit.RpcCtx())
	eg.SetLimit(10)

	successfulIDs := []uint32{}
	failedIDs := []uint32{}
	var mux sync.Mutex

	// 使用 data-service 原子接口
	for _, v := range req.GetIds() {
		v := v
		eg.Go(func() error {
			r := &pbds.DeleteGroupReq{
				Id: v,
				Attachment: &pbgroup.GroupAttachment{
					BizId: req.BizId,
				},
			}
			if _, err := s.client.DS.DeleteGroup(egCtx, r); err != nil {
				logs.Errorf("delete group %d failed, err: %v, rid: %s", v, err, grpcKit.Rid)

				// 错误不返回异常，记录错误ID
				mux.Lock()
				failedIDs = append(failedIDs, v)
				mux.Unlock()
				return nil
			}

			mux.Lock()
			successfulIDs = append(successfulIDs, v)
			mux.Unlock()

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		logs.Errorf("batch delete groups failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, errf.Errorf(errf.Aborted, i18n.T(grpcKit, "batch delete groups failed"))
	}

	// 全部失败, 当前API视为失败
	if len(failedIDs) == len(req.Ids) {
		return nil, errf.Errorf(errf.Aborted, i18n.T(grpcKit, "batch delete groups failed"))
	}

	return &pbcs.BatchDeleteResp{SuccessfulIds: successfulIDs, FailedIds: failedIDs}, nil
}

// UpdateGroup update a group
func (s *Service) UpdateGroup(ctx context.Context, req *pbcs.UpdateGroupReq) (*pbcs.UpdateGroupResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateGroupResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.UpdateGroupReq{
		Id: req.GroupId,
		Attachment: &pbgroup.GroupAttachment{
			BizId: req.BizId,
		},
		Spec: &pbgroup.GroupSpec{
			Name:     req.Name,
			Public:   req.Public,
			BindApps: req.BindApps,
			Mode:     req.Mode,
			Selector: req.Selector,
			Uid:      req.Uid,
		},
	}
	_, err = s.client.DS.UpdateGroup(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("update group failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListAllGroups list all groups in biz
// nolint:funlen
func (s *Service) ListAllGroups(ctx context.Context, req *pbcs.ListAllGroupsReq) (*pbcs.ListAllGroupsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAllGroupsResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	// 1. list all groups
	lgResp, err := s.client.DS.ListAllGroups(grpcKit.RpcCtx(), &pbds.ListAllGroupsReq{
		BizId:  req.BizId,
		TopIds: req.TopIds,
	})
	if err != nil {
		logs.Errorf("list groups failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	if len(lgResp.Details) == 0 {
		return resp, nil
	}

	// 2. list apps binded by groups if group is not public
	appMap := make(map[uint32]*pbapp.App)
	for _, group := range lgResp.Details {
		for _, appID := range group.Spec.BindApps {
			appMap[appID] = nil
		}
	}
	appIDs := make([]uint32, 0, len(appMap))
	for appID := range appMap {
		appIDs = append(appIDs, appID)
	}

	if len(appIDs) != 0 {
		laReq := &pbds.ListAppsByIDsReq{
			Ids: appIDs,
		}
		laResp, e := s.client.DS.ListAppsByIDs(grpcKit.RpcCtx(), laReq)
		if e != nil {
			logs.Errorf("list apps failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
		}
		for _, app := range laResp.Details {
			appMap[app.Id] = app
		}
	}

	// 3. caculate published apps count
	groups := make([]uint32, len(lgResp.Details))
	for idx, group := range lgResp.Details {
		groups[idx] = group.Id
	}
	countResp, err := s.client.DS.CountGroupsReleasedApps(grpcKit.RpcCtx(), &pbds.CountGroupsReleasedAppsReq{
		BizId:  req.BizId,
		Groups: groups,
	})
	if err != nil {
		logs.Errorf("count group published apps failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	respData := make([]*pbcs.ListAllGroupsResp_ListAllGroupsData, 0, len(lgResp.Details))
	for _, group := range lgResp.Details {
		apps := make([]*pbcs.ListAllGroupsResp_ListAllGroupsData_BindApp, 0, len(group.Spec.BindApps))
		for _, appID := range group.Spec.BindApps {
			if app, ok := appMap[appID]; ok && app != nil {
				apps = append(apps, &pbcs.ListAllGroupsResp_ListAllGroupsData_BindApp{
					Id:   app.Id,
					Name: app.Spec.Name,
				})
			}
		}
		data := &pbcs.ListAllGroupsResp_ListAllGroupsData{
			Id:       group.Id,
			Name:     group.Spec.Name,
			Public:   group.Spec.Public,
			BindApps: apps,
			Selector: group.Spec.Selector,
		}
		for _, d := range countResp.Data {
			if d.GroupId == group.Id {
				data.ReleasedAppsNum = d.Count
				data.Edited = d.Edited
			}
		}
		respData = append(respData, data)
	}

	resp.Details = respData

	return resp, nil
}

// ListAppGroups list groups in app
func (s *Service) ListAppGroups(ctx context.Context, req *pbcs.ListAppGroupsReq) (*pbcs.ListAppGroupsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAppGroupsResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	lReq := &pbds.ListAppGroupsReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}
	lResp, err := s.client.DS.ListAppGroups(grpcKit.RpcCtx(), lReq)
	if err != nil {
		logs.Errorf("list app groups failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	respData := make([]*pbcs.ListAppGroupsResp_ListAppGroupsData, 0, len(lResp.Details))
	for _, detail := range lResp.Details {
		data := &pbcs.ListAppGroupsResp_ListAppGroupsData{
			GroupId:     detail.GroupId,
			GroupName:   detail.GroupName,
			ReleaseId:   detail.ReleaseId,
			ReleaseName: detail.ReleaseName,
			OldSelector: detail.OldSelector,
			NewSelector: detail.NewSelector,
			Edited:      detail.Edited,
		}
		respData = append(respData, data)
	}
	resp.Details = respData

	return resp, nil
}

// ListGroupReleasedApps list released apps in group
func (s *Service) ListGroupReleasedApps(ctx context.Context, req *pbcs.ListGroupReleasedAppsReq) (
	*pbcs.ListGroupReleasedAppsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListGroupReleasedAppsResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	lReq := &pbds.ListGroupReleasedAppsReq{
		BizId:     req.BizId,
		GroupId:   req.GroupId,
		SearchKey: req.SearchKey,
		Start:     req.Start,
		Limit:     req.Limit,
	}
	lResp, err := s.client.DS.ListGroupReleasedApps(grpcKit.RpcCtx(), lReq)
	if err != nil {
		logs.Errorf("list group released apps failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	data := make([]*pbcs.ListGroupReleasedAppsResp_ListGroupReleasedAppsData, len(lResp.Details))
	for idx, detail := range lResp.Details {
		data[idx] = &pbcs.ListGroupReleasedAppsResp_ListGroupReleasedAppsData{
			AppId:       detail.AppId,
			AppName:     detail.AppName,
			ReleaseId:   detail.ReleaseId,
			ReleaseName: detail.ReleaseName,
			Edited:      detail.Edited,
		}
	}
	resp.Details = data
	resp.Count = lResp.Count
	return resp, nil
}

// GetGroupByName get group by name
func (s *Service) GetGroupByName(ctx context.Context, req *pbcs.GetGroupByNameReq) (*pbgroup.Group, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(kt, res...); err != nil {
		return nil, err
	}

	r := &pbds.GetGroupByNameReq{
		BizId:     req.BizId,
		GroupName: req.GroupName,
	}
	rp, err := s.client.DS.GetGroupByName(kt.RpcCtx(), r)
	if err != nil {
		logs.Errorf("get group by name %s failed, err: %v, rid: %s", req.GroupName, err, kt.Rid)
		return nil, err
	}

	return rp, nil
}
