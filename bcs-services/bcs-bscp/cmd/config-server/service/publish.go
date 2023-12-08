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

	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbds "bscp.io/pkg/protocol/data-service"
)

// Publish publish a strategy
func (s *Service) Publish(ctx context.Context, req *pbcs.PublishReq) (
	*pbcs.PublishResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Publish, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.PublishReq{
		BizId:           req.BizId,
		AppId:           req.AppId,
		ReleaseId:       req.ReleaseId,
		Memo:            req.Memo,
		All:             req.All,
		GrayPublishMode: req.GrayPublishMode,
		Default:         req.Default,
		Groups:          req.Groups,
		Labels:          req.Labels,
		GroupName:       req.GroupName,
	}
	rp, err := s.client.DS.Publish(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("publish failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.PublishResp{
		Id: rp.PublishedStrategyHistoryId,
	}
	return resp, nil
}

// GenerateReleaseAndPublish generate release and publish
func (s *Service) GenerateReleaseAndPublish(ctx context.Context, req *pbcs.GenerateReleaseAndPublishReq) (
	*pbcs.PublishResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.GenerateRelease, ResourceID: req.AppId}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Publish, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.GenerateReleaseAndPublishReq{
		BizId:           req.BizId,
		AppId:           req.AppId,
		ReleaseName:     req.ReleaseName,
		ReleaseMemo:     req.ReleaseMemo,
		Variables:       req.Variables,
		All:             req.All,
		GrayPublishMode: req.GrayPublishMode,
		Groups:          req.Groups,
		Labels:          req.Labels,
		GroupName:       req.GroupName,
	}
	rp, err := s.client.DS.GenerateReleaseAndPublish(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("generate release and publish failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.PublishResp{
		Id: rp.PublishedStrategyHistoryId,
	}
	return resp, nil
}
