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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbgroup "bscp.io/pkg/protocol/core/group"
	pbstrategy "bscp.io/pkg/protocol/core/strategy"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// CreateStrategy create a strategy
func (s *Service) CreateStrategy(ctx context.Context, req *pbcs.CreateStrategyReq) (*pbcs.CreateStrategyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateStrategyResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Strategy, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	groups, err := s.queryGroups(grpcKit.RpcCtx(), req.BizId, req.AppId, req.Groups)
	if err != nil {
		return nil, err
	}

	r := &pbds.CreateStrategyReq{
		Attachment: &pbstrategy.StrategyAttachment{
			BizId:         req.BizId,
			AppId:         req.AppId,
			StrategySetId: req.StrategySetId,
		},
		Spec: &pbstrategy.StrategySpec{
			Name:      req.Name,
			ReleaseId: req.ReleaseId,
			AsDefault: req.AsDefault,
			Scope: &pbstrategy.Scope{
				Groups: groups,
			},
			Namespace: req.Namespace,
			Memo:      req.Memo,
		},
	}
	rp, err := s.client.DS.CreateStrategy(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateStrategyResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteStrategy delete a strategy
func (s *Service) DeleteStrategy(ctx context.Context, req *pbcs.DeleteStrategyReq) (*pbcs.DeleteStrategyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteStrategyResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Strategy, Action: meta.Delete,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.DeleteStrategyReq{
		Id: req.Id,
		Attachment: &pbstrategy.StrategyAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	_, err = s.client.DS.DeleteStrategy(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("delete strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// UpdateStrategy update a strategy
func (s *Service) UpdateStrategy(ctx context.Context, req *pbcs.UpdateStrategyReq) (*pbcs.UpdateStrategyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateStrategyResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Strategy, Action: meta.Update,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	groups, err := s.queryGroups(grpcKit.RpcCtx(), req.BizId, req.AppId, req.Groups)
	if err != nil {
		return nil, err
	}

	r := &pbds.UpdateStrategyReq{
		Id: req.Id,
		Attachment: &pbstrategy.StrategyAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbstrategy.StrategySpec{
			Name:      req.Name,
			ReleaseId: req.ReleaseId,
			AsDefault: req.AsDefault,
			Scope: &pbstrategy.Scope{
				Groups: groups,
			},
			Memo: req.Memo,
		},
	}
	_, err = s.client.DS.UpdateStrategy(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("update strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListStrategies list strategies with filter
func (s *Service) ListStrategies(ctx context.Context, req *pbcs.ListStrategiesReq) (*pbcs.ListStrategiesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListStrategiesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Strategy, Action: meta.Find}, BizID: req.BizId}
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

	r := &pbds.ListStrategiesReq{
		BizId:  req.BizId,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListStrategies(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list strategies failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListStrategiesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

func (s *Service) queryGroups(ctx context.Context, bizID, appID uint32, groupIDs []uint32) ([]*pbgroup.Group, error) {

	exp := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "id",
				Op:    filter.In.Factory(),
				Value: groupIDs,
			},
		},
	}
	ft, err := exp.MarshalPB()
	if err != nil {
		return nil, err
	}
	in := &pbds.ListGroupsReq{
		BizId:  bizID,
		AppId:  appID,
		Filter: ft,
		Page: &pbbase.BasePage{
			Count: false,
			Start: 0,
			Limit: 0,
			Order: string(types.Ascending.Order()),
		},
	}

	resp, err := s.client.DS.ListGroups(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp.Details, nil

}
