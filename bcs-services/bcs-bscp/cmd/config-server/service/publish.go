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
	"errors"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
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
		Id:              rp.PublishedStrategyHistoryId,
		HaveCredentials: rp.HaveCredentials,
		HavePull:        rp.HavePull,
	}
	return resp, nil
}

// SubmitPublishApprove submit publish a strategy
func (s *Service) SubmitPublishApprove(ctx context.Context, req *pbcs.SubmitPublishApproveReq) (
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

	r := &pbds.SubmitPublishApproveReq{
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
		PublishType:     req.PublishType,
		PublishTime:     req.PublishTime,
		IsCompare:       req.IsCompare,
	}
	rp, err := s.client.DS.SubmitPublishApprove(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("publish failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.PublishResp{
		Id:              rp.PublishedStrategyHistoryId,
		HaveCredentials: rp.HaveCredentials,
		HavePull:        rp.HavePull,
	}
	return resp, nil
}

// Approve publish approve
func (s *Service) Approve(ctx context.Context, req *pbcs.ApproveReq) (*pbcs.ApproveResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.Publish, ResourceID: req.AppId}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.ApproveReq{
		BizId:         req.BizId,
		AppId:         req.AppId,
		ReleaseId:     req.ReleaseId,
		PublishStatus: req.PublishStatus,
		Reason:        req.Reason,
	}
	rp, err := s.client.DS.Approve(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("approve failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ApproveResp{
		HaveCredentials: rp.HaveCredentials,
		Code:            0,
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

	// 创建版本前验证非模板配置和模板配置是否存在冲突
	ci, err := s.ListConfigItems(grpcKit.RpcCtx(), &pbcs.ListConfigItemsReq{
		BizId: req.BizId,
		AppId: req.AppId,
		All:   true,
	})
	if err != nil {
		return nil, err
	}
	if ci.ConflictNumber > 0 {
		logs.Errorf("generate release and publish failed there is a file conflict, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, errors.New("generate release and publish failed there is a file conflict")
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
