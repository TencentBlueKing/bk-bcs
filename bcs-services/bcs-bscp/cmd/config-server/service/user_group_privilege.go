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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// DeleteUserGroupPrivilege 删除用户组权限数据
func (s *Service) DeleteUserGroupPrivilege(ctx context.Context, req *pbcs.DeleteUserPrivilegesReq) (
	*pbcs.DeleteUserPrivilegesResp, error) {
	// FromGrpcContext used only to obtain Kit through grpc context.
	kit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Delete, ResourceID: req.AppId}, BizID: req.BizId},
	}
	// Authorize authorize if user has permission to the resources.
	// If user is unauthorized, assign apply url and resources into error.
	if err := s.authorizer.Authorize(kit, res...); err != nil {
		return nil, err
	}

	_, err := s.client.DS.DeleteUserGroupPrivilege(kit.RpcCtx(), &pbds.DeleteUserPrivilegesReq{
		Id:              req.Id,
		BizId:           req.BizId,
		AppId:           req.AppId,
		TemplateSpaceId: req.TemplateSpaceId,
	})

	if err != nil {
		return nil, err
	}

	return &pbcs.DeleteUserPrivilegesResp{}, nil
}

// ListUserGroupPrivileges 获取用户组数据列表
func (s *Service) ListUserGroupPrivileges(ctx context.Context, req *pbcs.ListUserPrivilegesReq) (
	*pbcs.ListUserPrivilegesResp, error) {
	// FromGrpcContext used only to obtain Kit through grpc context.
	kit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	// Authorize authorize if user has permission to the resources.
	// If user is unauthorized, assign apply url and resources into error.
	if err := s.authorizer.Authorize(kit, res...); err != nil {
		return nil, err
	}

	resp, err := s.client.DS.ListUserGroupPrivileges(kit.RpcCtx(), &pbds.ListUserPrivilegesReq{
		BizId:           req.BizId,
		AppId:           req.AppId,
		TemplateSpaceId: req.TemplateSpaceId,
		Name:            req.Name,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	})

	if err != nil {
		return nil, err
	}

	items := make([]*pbcs.ListUserPrivilegesResp_Detail, 0, len(resp.GetDetails()))
	for _, v := range resp.GetDetails() {
		items = append(items, &pbcs.ListUserPrivilegesResp_Detail{
			Id:            v.GetId(),
			Name:          v.GetName(),
			PrivilegeType: v.GetPrivilegeType(),
			ReadOnly:      v.GetReadOnly(),
			Pid:           v.GetPid(),
		})
	}

	return &pbcs.ListUserPrivilegesResp{Details: items, Count: resp.Count}, nil
}
