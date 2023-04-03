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
	pbapp "bscp.io/pkg/protocol/core/app"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbgroup "bscp.io/pkg/protocol/core/group"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
)

// CreateGroup create a group
func (s *Service) CreateGroup(ctx context.Context, req *pbcs.CreateGroupReq) (*pbcs.CreateGroupResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateGroupResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Create,
		ResourceID: req.BizId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
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

	resp = &pbcs.CreateGroupResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteGroup delete a group
func (s *Service) DeleteGroup(ctx context.Context, req *pbcs.DeleteGroupReq) (*pbcs.DeleteGroupResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteGroupResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Delete,
		ResourceID: req.GroupId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
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

// UpdateGroup update a group
func (s *Service) UpdateGroup(ctx context.Context, req *pbcs.UpdateGroupReq) (*pbcs.UpdateGroupResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateGroupResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Update,
		ResourceID: req.GroupId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
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
func (s *Service) ListAllGroups(ctx context.Context, req *pbcs.ListAllGroupsReq) (*pbcs.ListAllGroupsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAllGroupsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	// 1. list groups
	lgft := &filter.Expression{
		Op:    filter.And,
		Rules: []filter.RuleFactory{},
	}
	lgftpb, err := lgft.MarshalPB()
	if err != nil {
		return nil, err
	}

	lgReq := &pbds.ListGroupsReq{
		BizId:  req.BizId,
		Filter: lgftpb,
		Page:   &pbbase.BasePage{},
	}
	lgResp, err := s.client.DS.ListGroups(grpcKit.RpcCtx(), lgReq)
	if err != nil {
		logs.Errorf("list groups failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	if lgResp.Count == 0 {
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
		laft := &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{&filter.AtomRule{
				Field: "id",
				Op:    filter.In.Factory(),
				Value: appIDs,
			}},
		}
		laftpb, err := laft.MarshalPB()
		if err != nil {
			return nil, err
		}
		laReq := &pbds.ListAppsReq{
			BizId:  req.BizId,
			Filter: laftpb,
			Page:  &pbbase.BasePage{},
		}
		laResp, err := s.client.DS.ListApps(grpcKit.RpcCtx(), laReq)
		if err != nil {
			logs.Errorf("list apps failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
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
	countResp, err := s.client.DS.CountGroupsPublishedApps(grpcKit.RpcCtx(), &pbds.CountGroupsPublishedAppsReq{
		BizId:  req.BizId,
		Groups: groups,
	})
	if err != nil {
		logs.Errorf("count group published apps failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	respData := make([]*pbcs.ListAllGroupsResp_ListAllGroupsData, 0, len(lgResp.Details))
	for _, group := range lgResp.Details {
		apps := make([]string, 0, len(group.Spec.BindApps))
		for _, appID := range group.Spec.BindApps {
			if app, ok := appMap[appID]; ok {
				apps = append(apps, app.Spec.Name)
			}
		}
		respData = append(respData, &pbcs.ListAllGroupsResp_ListAllGroupsData{
			Id:              group.Id,
			Name:            group.Spec.Name,
			Public:          group.Spec.Public,
			BindApps:        apps,
			Selector:        group.Spec.Selector,
			ReleasedAppsNum: countResp.Counts[group.Id],
		})
	}
	resp.Details = respData

	return resp, nil
}
