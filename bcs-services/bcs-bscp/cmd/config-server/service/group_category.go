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
	pbgc "bscp.io/pkg/protocol/core/group-category"
	pbds "bscp.io/pkg/protocol/data-service"
)

// CreateGroupCategory create a group category
func (s *Service) CreateGroupCategory(ctx context.Context, req *pbcs.CreateGroupCategoryReq) (*pbcs.CreateGroupCategoryResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateGroupCategoryResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.GroupCategory, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.CreateGroupCategoryReq{
		Attachment: &pbgc.GroupCategoryAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbgc.GroupCategorySpec{
			Name: req.Name,
		},
	}
	rp, err := s.client.DS.CreateGroupCategory(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create group category failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateGroupCategoryResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteGroupCategory delete a group category
func (s *Service) DeleteGroupCategory(ctx context.Context, req *pbcs.DeleteGroupCategoryReq) (*pbcs.DeleteGroupCategoryResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteGroupCategoryResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.GroupCategory, Action: meta.Delete,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.DeleteGroupCategoryReq{
		Id: req.Id,
		Attachment: &pbgc.GroupCategoryAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	_, err = s.client.DS.DeleteGroupCategory(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("delete group category failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}
