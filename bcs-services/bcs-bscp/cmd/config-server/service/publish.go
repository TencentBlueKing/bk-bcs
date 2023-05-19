/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
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
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// Publish publish a strategy
func (s *Service) Publish(ctx context.Context, req *pbcs.PublishReq) (
	*pbcs.PublishResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.PublishResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Strategy, Action: meta.Publish,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.PublishReq{
		BizId:     req.BizId,
		AppId:     req.AppId,
		ReleaseId: req.ReleaseId,
		Memo:      req.Memo,
		All:       req.All,
		Default:   req.Default,
		Groups:    req.Groups,
	}
	rp, err := s.client.DS.Publish(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("publish failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.PublishResp{
		Id: rp.PublishedStrategyHistoryId,
	}
	return resp, nil
}

// GenerateReleaseAndPublish generate release and publish
func (s *Service) GenerateReleaseAndPublish(ctx context.Context, req *pbcs.GenerateReleaseAndPublishReq) (
	*pbcs.PublishResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.PublishResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Strategy, Action: meta.Publish,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.GenerateReleaseAndPublishReq{
		BizId:       req.BizId,
		AppId:       req.AppId,
		ReleaseName: req.ReleaseName,
		ReleaseMemo: req.ReleaseMemo,
		All:         req.All,
		Groups:      req.Groups,
	}
	rp, err := s.client.DS.GenerateReleaseAndPublish(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("generate release and publish failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.PublishResp{
		Id: rp.PublishedStrategyHistoryId,
	}
	return resp, nil
}

// FinishPublish finish the published strategy
func (s *Service) FinishPublish(ctx context.Context, req *pbcs.FinishPublishReq) (
	*pbcs.FinishPublishResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.FinishPublishResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Strategy, Action: meta.FinishPublish,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.FinishPublishReq{
		BizId:      req.BizId,
		AppId:      req.AppId,
		StrategyId: req.Id,
	}
	_, err = s.client.DS.FinishPublish(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("finish publish strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListPublishedStrategyHistories list the published strategy histories.
func (s *Service) ListPublishedStrategyHistories(ctx context.Context, req *pbcs.ListPubStrategyHistoriesReq) (
	*pbcs.ListPubStrategyHistoriesResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListPubStrategyHistoriesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.PSH, Action: meta.Find,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	if req.Page == nil {
		return nil, errf.New(errf.InvalidParameter, "page is null")
	}

	if err = req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	r := &pbds.ListPubStrategyHistoriesReq{
		BizId:  req.BizId,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListPublishedStrategyHistories(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list published strategy histories failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListPubStrategyHistoriesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
