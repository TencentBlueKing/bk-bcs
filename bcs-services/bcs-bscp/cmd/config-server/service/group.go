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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbgroup "bscp.io/pkg/protocol/core/group"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
)

// CreateGroup create a group
func (s *Service) CreateGroup(ctx context.Context, req *pbcs.CreateGroupReq) (*pbcs.CreateGroupResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateGroupResp)

	bizID, err := strconv.Atoi(grpcKit.SpaceID)
	if err != nil {
		return nil, err
	}
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Create,
		ResourceID: req.AppId}, BizID: uint32(bizID)}
	err = s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.CreateGroupReq{
		Attachment: &pbgroup.GroupAttachment{
			BizId:           uint32(bizID),
			AppId:           req.AppId,
			GroupCategoryId: req.GroupCategoryId,
		},
		Spec: &pbgroup.GroupSpec{
			Name:     req.Name,
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

	bizID, err := strconv.Atoi(grpcKit.SpaceID)
	if err != nil {
		return nil, err
	}
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Delete,
		ResourceID: req.AppId}, BizID: uint32(bizID)}
	err = s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.DeleteGroupReq{
		Id: req.GroupId,
		Attachment: &pbgroup.GroupAttachment{
			BizId: uint32(bizID),
			AppId: req.AppId,
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

	bizID, err := strconv.Atoi(grpcKit.SpaceID)
	if err != nil {
		return nil, err
	}
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Update,
		ResourceID: req.AppId}, BizID: uint32(bizID)}
	err = s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.UpdateGroupReq{
		Id: req.GroupId,
		Attachment: &pbgroup.GroupAttachment{
			BizId: uint32(bizID),
			AppId: req.AppId,
		},
		Spec: &pbgroup.GroupSpec{
			Name:     req.Name,
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

// ListGroups list groups with filter
func (s *Service) ListGroups(ctx context.Context, req *pbcs.ListGroupsReq) (*pbcs.ListGroupsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListGroupsResp)

	bizID, err := strconv.Atoi(grpcKit.SpaceID)
	if err != nil {
		return nil, err
	}
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Find}, BizID: uint32(bizID)}
	err = s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	if req.Start < 0 {
		return nil, errf.New(errf.InvalidParameter, "start has to be greater than 0")
	}

	if req.Limit < 0 {
		return nil, errf.New(errf.InvalidParameter, "limit has to be greater than 0")
	}

	ft := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "group_category_id",
				Op:    filter.Equal.Factory(),
				Value: req.GroupCategoryId,
			},
			&filter.AtomRule{
				Field: "mode",
				Op:    filter.Equal.Factory(),
				Value: req.Mode,
			},
		},
	}
	ftpb, err := ft.MarshalPB()
	if err != nil {
		return nil, err
	}

	r := &pbds.ListGroupsReq{
		BizId:  uint32(bizID),
		AppId:  req.AppId,
		Filter: ftpb,
		Page: &pbbase.BasePage{
			Start: req.Start,
			Limit: req.Limit,
		},
	}
	rp, err := s.client.DS.ListGroups(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list groups failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListGroupsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListAllGroups list all groups
func (s *Service) ListAllGroups(ctx context.Context, req *pbcs.ListAllGroupsReq) (*pbcs.ListAllGroupsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAllGroupsResp)

	bizID, err := strconv.Atoi(grpcKit.SpaceID)
	if err != nil {
		return nil, err
	}
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Find}, BizID: uint32(bizID)}
	err = s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	ft := &filter.Expression{
		Op:    filter.And,
		Rules: []filter.RuleFactory{},
	}
	ftpb, err := ft.MarshalPB()
	if err != nil {
		return nil, err
	}

	lgcReq := &pbds.ListGroupCategoriesReq{
		BizId:  uint32(bizID),
		AppId:  req.AppId,
		Filter: ftpb,
		Page:   &pbbase.BasePage{},
	}
	lgcResp, err := s.client.DS.ListGroupCategories(grpcKit.RpcCtx(), lgcReq)
	if err != nil {
		logs.Errorf("list all group categories failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	ft = &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "mode",
				Op:    filter.Equal.Factory(),
				Value: req.Mode,
			},
		},
	}
	ftpb, err = ft.MarshalPB()
	if err != nil {
		return nil, err
	}

	lgReq := &pbds.ListGroupsReq{
		BizId:  uint32(bizID),
		AppId:  req.AppId,
		Filter: ftpb,
		Page:   &pbbase.BasePage{},
	}

	lgResp, err := s.client.DS.ListGroups(grpcKit.RpcCtx(), lgReq)
	if err != nil {
		logs.Errorf("list all groups failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	groupsMap := make(map[uint32][]*pbgroup.Group, len(lgcResp.Details))
	for _, detail := range lgResp.Details {
		if _, exists := groupsMap[detail.Attachment.GroupCategoryId]; exists {
			groupsMap[detail.Attachment.GroupCategoryId] = append(groupsMap[detail.Attachment.GroupCategoryId], detail)
		} else {
			groupsMap[detail.Attachment.GroupCategoryId] = []*pbgroup.Group{detail}
		}
	}
	respData := []*pbcs.ListAllGroupsResp_ListAllGroupsData{}

	for _, detail := range lgcResp.Details {
		respData = append(respData, &pbcs.ListAllGroupsResp_ListAllGroupsData{
			GroupCategoryId:   detail.Id,
			GroupCategoryName: detail.Spec.Name,
			Groups:            groupsMap[detail.Id],
		})
	}

	resp = &pbcs.ListAllGroupsResp{
		Details: respData,
	}
	return resp, nil
}
