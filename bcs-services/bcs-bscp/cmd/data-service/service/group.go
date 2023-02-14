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
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbgroup "bscp.io/pkg/protocol/core/group"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateGroup create group.
func (s *Service) CreateGroup(ctx context.Context, req *pbds.CreateGroupReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	spec, err := req.Spec.GroupSpec()
	if err != nil {
		logs.Errorf("get group spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	group := &table.Group{
		Spec:       spec,
		Attachment: req.Attachment.GroupAttachment(),
		Revision: &table.Revision{
			Creator:   kt.User,
			Reviser:   kt.User,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	id, err := s.dao.Group().Create(kt, group)
	if err != nil {
		logs.Errorf("create group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListGroups list groups.
func (s *Service) ListGroups(ctx context.Context, req *pbds.ListGroupsReq) (*pbds.ListGroupsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	query := &types.ListGroupsOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.Group().List(kt, query)
	if err != nil {
		logs.Errorf("list group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	groups, err := pbgroup.PbGroups(details.Details)
	if err != nil {
		logs.Errorf("get pb group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListGroupsResp{
		Count:   details.Count,
		Details: groups,
	}
	return resp, nil
}

// UpdateGroup update group.
func (s *Service) UpdateGroup(ctx context.Context, req *pbds.UpdateGroupReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	spec, err := req.Spec.GroupSpec()
	if err != nil {
		logs.Errorf("get group spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	group := &table.Group{
		ID:         req.Id,
		Spec:       spec,
		Attachment: req.Attachment.GroupAttachment(),
		Revision: &table.Revision{
			Reviser:   kt.User,
			UpdatedAt: now,
		},
	}
	if err := s.dao.Group().Update(kt, group); err != nil {
		logs.Errorf("update group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteGroup delete group.
func (s *Service) DeleteGroup(ctx context.Context, req *pbds.DeleteGroupReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	group := &table.Group{
		ID:         req.Id,
		Attachment: req.Attachment.GroupAttachment(),
	}
	if err := s.dao.Group().Delete(kt, group); err != nil {
		logs.Errorf("delete group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
