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
	pbgroup "bscp.io/pkg/protocol/core/group"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateGroup create a group
func (s *Service) CreateGroup(ctx context.Context, req *pbcs.CreateGroupReq) (*pbcs.CreateGroupResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateGroupResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return resp, nil
	}

	r := &pbds.CreateGroupReq{
		Attachment: &pbgroup.GroupAttachment{
			BizId:           req.BizId,
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
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("create group failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.CreateGroupResp_RespData{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteGroup delete a group
func (s *Service) DeleteGroup(ctx context.Context, req *pbcs.DeleteGroupReq) (*pbcs.DeleteGroupResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteGroupResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Delete,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return resp, nil
	}

	r := &pbds.DeleteGroupReq{
		Id: req.Id,
		Attachment: &pbgroup.GroupAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	_, err = s.client.DS.DeleteGroup(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("delete group failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	return resp, nil
}

// UpdateGroup update a group
func (s *Service) UpdateGroup(ctx context.Context, req *pbcs.UpdateGroupReq) (*pbcs.UpdateGroupResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateGroupResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Update,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return resp, nil
	}

	r := &pbds.UpdateGroupReq{
		Id: req.Id,
		Attachment: &pbgroup.GroupAttachment{
			BizId: req.BizId,
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
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("update group failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	return resp, nil
}

// ListGroups list groups with filter
func (s *Service) ListGroups(ctx context.Context, req *pbcs.ListGroupsReq) (*pbcs.ListGroupsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListGroupsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Group, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return resp, nil
	}

	if req.Page == nil {
		errf.Error(errf.New(errf.InvalidParameter, "page is null")).AssignResp(grpcKit, resp)
		return resp, nil
	}

	if err = req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		return resp, nil
	}

	r := &pbds.ListGroupsReq{
		BizId:  req.BizId,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListGroups(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("list groups failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.ListGroupsResp_RespData{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
