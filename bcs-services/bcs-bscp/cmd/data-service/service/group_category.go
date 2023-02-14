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
	pbgc "bscp.io/pkg/protocol/core/group-category"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateGroupCategory create group category.
func (s *Service) CreateGroupCategory(ctx context.Context, req *pbds.CreateGroupCategoryReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	now := time.Now()
	category := &table.GroupCategory{
		Spec:       req.Spec.GroupCategorySpec(),
		Attachment: req.Attachment.GroupCategoryAttachment(),
		Revision: &table.CreatedRevision{
			Creator:   kt.User,
			CreatedAt: now,
		},
	}
	id, err := s.dao.GroupCategory().Create(kt, category)
	if err != nil {
		logs.Errorf("create group category failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListGroupCategories list group categories.
func (s *Service) ListGroupCategories(ctx context.Context, req *pbds.ListGroupCategoriesReq) (*pbds.ListGroupCategoriesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	query := &types.ListGroupCategoriesOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.GroupCategory().List(kt, query)
	if err != nil {
		logs.Errorf("list group category failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListGroupCategoriesResp{
		Count:   details.Count,
		Details: pbgc.PbGroupCategories(details.Details),
	}
	return resp, nil
}

// DeleteGroupCategory delete group category.
func (s *Service) DeleteGroupCategory(ctx context.Context, req *pbds.DeleteGroupCategoryReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	category := &table.GroupCategory{
		ID:         req.Id,
		Attachment: req.Attachment.GroupCategoryAttachment(),
	}
	if err := s.dao.GroupCategory().Delete(kt, category); err != nil {
		logs.Errorf("delete group category failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
