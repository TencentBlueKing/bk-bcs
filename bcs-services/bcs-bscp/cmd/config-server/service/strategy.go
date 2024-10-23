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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbstrategy "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/strategy"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// GetLastSelect get last version select publish_type
func (s *Service) GetLastSelect(ctx context.Context, req *pbcs.GetLastSelectReq) (
	*pbcs.GetLastSelectResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.GetLastSelectReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}
	rp, err := s.client.DS.GetLastSelect(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("publish failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.GetLastSelectResp{
		PublishType: rp.PublishType,
		IsApprove:   rp.IsApprove,
	}
	return resp, nil
}

// GetLastPublish get last publish list
func (s *Service) GetLastPublish(ctx context.Context, req *pbcs.GetLastPublishReq) (
	*pbcs.GetLastPublishResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.GetLastPublishReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}
	rp, err := s.client.DS.GetLastPublish(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("publish failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.GetLastPublishResp{
		IsPublishing:  rp.IsPublishing,
		VersionName:   rp.VersionName,
		UpdatedAt:     rp.UpdatedAt,
		PublishRecord: rp.PublishRecord,
	}
	return resp, nil
}

// GetReleasesStatus get releases status
func (s *Service) GetReleasesStatus(ctx context.Context, req *pbcs.GetReleasesStatusReq) (
	*pbstrategy.Strategy, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.GetReleasesStatusReq{
		BizId:     req.BizId,
		AppId:     req.AppId,
		ReleaseId: req.ReleaseId,
	}
	rp, err := s.client.DS.GetReleasesStatus(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("publish failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return rp, nil
}
