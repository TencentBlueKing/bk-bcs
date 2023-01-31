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
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbstrategy "bscp.io/pkg/protocol/core/strategy"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// PublishStrategy exec publish strategy.
func (s *Service) PublishStrategy(ctx context.Context, req *pbds.PublishStrategyReq) (
	*pbds.PublishStrategyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	opt := &types.PublishStrategyOption{
		BizID:      req.BizId,
		AppID:      req.AppId,
		StrategyID: req.StrategyId,
		Revision: &table.CreatedRevision{
			Creator:   kt.User,
			CreatedAt: time.Now(),
		},
	}
	pshID, err := s.dao.Publish().PublishStrategy(kt, opt)
	if err != nil {
		logs.Errorf("publish strategy failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.PublishStrategyResp{PublishedStrategyHistoryId: pshID}
	return resp, nil
}

// FinishPublishStrategy finish publish strategy.
func (s *Service) FinishPublishStrategy(ctx context.Context, req *pbds.FinishPublishStrategyReq) (
	*pbbase.EmptyResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	opt := &types.FinishPublishOption{
		BizID:      req.BizId,
		AppID:      req.AppId,
		StrategyID: req.StrategyId,
	}
	err := s.dao.Publish().FinishPublish(grpcKit, opt)
	if err != nil {
		logs.Errorf("finish publish strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// ListPublishedStrategyHistories list published strategy histories.
func (s *Service) ListPublishedStrategyHistories(ctx context.Context, req *pbds.ListPubStrategyHistoriesReq) (
	*pbds.ListPubStrategyHistoriesResp, error) {

	kt := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	ft, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	query := &types.ListPSHistoriesOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: ft,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.Publish().ListPSHistory(kt, query)
	if err != nil {
		logs.Errorf("list published strategy history failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	strategies, err := pbstrategy.PbPubStrategyHistories(details.Details)
	if err != nil {
		logs.Errorf("get pb strategy histories failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListPubStrategyHistoriesResp{
		Count:   details.Count,
		Details: strategies,
	}
	return resp, nil
}
