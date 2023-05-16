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
	"bscp.io/pkg/protocol/core/base"
	"bscp.io/pkg/protocol/core/strategy-set"
	"bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateStrategySet create strategy set.
func (s *Service) CreateStrategySet(ctx context.Context, req *pbds.CreateStrategySetReq) (*pbds.CreateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	now := time.Now()
	ss := &table.StrategySet{
		Spec:       req.Spec.StrategySetSpec(),
		Attachment: req.Attachment.StrategySetAttachment(),
		State: &table.StrategySetState{
			Status: table.Enabled,
		},
		Revision: &table.Revision{
			Creator:   grpcKit.User,
			Reviser:   grpcKit.User,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	id, err := s.dao.StrategySet().Create(grpcKit, ss)
	if err != nil {
		logs.Errorf("create strategy set failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListStrategySets list strategy sets by query condition.
func (s *Service) ListStrategySets(ctx context.Context, req *pbds.ListStrategySetsReq) (
	*pbds.ListStrategySetsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	query := &types.ListStrategySetsOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.StrategySet().List(grpcKit, query)
	if err != nil {
		logs.Errorf("list strategy set failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.ListStrategySetsResp{
		Count:   details.Count,
		Details: pbss.PbStrategySets(details.Details),
	}
	return resp, nil
}

// UpdateStrategySet update strategy set.
func (s *Service) UpdateStrategySet(ctx context.Context, req *pbds.UpdateStrategySetReq) (*pbbase.EmptyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	ss := &table.StrategySet{
		ID:         req.Id,
		Spec:       req.Spec.StrategySetSpec(),
		Attachment: req.Attachment.StrategySetAttachment(),
		State:      req.State.StrategySetState(),
		Revision: &table.Revision{
			Reviser:   grpcKit.User,
			UpdatedAt: time.Now(),
		},
	}
	if err := s.dao.StrategySet().Update(grpcKit, ss); err != nil {
		logs.Errorf("update strategy set failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteStrategySet delete strategy set.
func (s *Service) DeleteStrategySet(ctx context.Context, req *pbds.DeleteStrategySetReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	strategy := &table.StrategySet{
		ID:         req.Id,
		Attachment: req.Attachment.StrategySetAttachment(),
	}
	if err := s.dao.StrategySet().Delete(kt, strategy); err != nil {
		logs.Errorf("delete strategy set failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
