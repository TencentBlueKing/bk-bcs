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
	pbstrategy "bscp.io/pkg/protocol/core/strategy"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateStrategy create strategy.
func (s *Service) CreateStrategy(ctx context.Context, req *pbds.CreateStrategyReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	spec, err := req.Spec.StrategySpec()
	if err != nil {
		logs.Errorf("get strategy spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	strategy := &table.Strategy{
		Spec: spec,
		State: &table.StrategyState{
			PubState: table.Unpublished,
		},
		Attachment: req.Attachment.StrategyAttachment(),
		Revision: &table.Revision{
			Creator:   kt.User,
			Reviser:   kt.User,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	id, err := s.dao.Strategy().Create(kt, strategy)
	if err != nil {
		logs.Errorf("create strategy failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListStrategies list strategies.
func (s *Service) ListStrategies(ctx context.Context, req *pbds.ListStrategiesReq) (*pbds.ListStrategiesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	query := &types.ListStrategiesOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.Strategy().List(kt, query)
	if err != nil {
		logs.Errorf("list strategy failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	strategies, err := pbstrategy.PbStrategies(details.Details)
	if err != nil {
		logs.Errorf("get pb strategy failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListStrategiesResp{
		Count:   details.Count,
		Details: strategies,
	}
	return resp, nil
}

// UpdateStrategy update strategy.
func (s *Service) UpdateStrategy(ctx context.Context, req *pbds.UpdateStrategyReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	spec, err := req.Spec.StrategySpec()
	if err != nil {
		logs.Errorf("get strategy spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	strategy := &table.Strategy{
		ID:         req.Id,
		Spec:       spec,
		State:      req.Status.StrategyState(),
		Attachment: req.Attachment.StrategyAttachment(),
		Revision: &table.Revision{
			Reviser:   kt.User,
			UpdatedAt: now,
		},
	}
	if err := s.dao.Strategy().Update(kt, strategy); err != nil {
		logs.Errorf("update strategy failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteStrategy delete strategy.
func (s *Service) DeleteStrategy(ctx context.Context, req *pbds.DeleteStrategyReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	strategy := &table.Strategy{
		ID:         req.Id,
		Attachment: req.Attachment.StrategyAttachment(),
	}
	if err := s.dao.Strategy().Delete(kt, strategy); err != nil {
		logs.Errorf("delete strategy failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
