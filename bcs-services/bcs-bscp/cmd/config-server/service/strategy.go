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
	"bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/protocol/core/strategy"
	"bscp.io/pkg/protocol/data-service"
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
		return resp, nil
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
			Scope:     req.Scope,
			Namespace: req.Namespace,
			Memo:      req.Memo,
		},
	}
	rp, err := s.client.DS.CreateStrategy(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("create strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.CreateStrategyResp_RespData{
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
		return resp, nil
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
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("delete strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
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
		return resp, nil
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
			Scope:     req.Scope,
			Memo:      req.Memo,
		},
	}
	_, err = s.client.DS.UpdateStrategy(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("update strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	return resp, nil
}

// ListStrategies list strategies with filter
func (s *Service) ListStrategies(ctx context.Context, req *pbcs.ListStrategiesReq) (*pbcs.ListStrategiesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListStrategiesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Strategy, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return resp, nil
	}

	if req.Page == nil {
		errf.Error(errf.New(errf.InvalidParameter, "page is null")).AssignResp(grpcKit, resp)
		return resp, nil
	}

	if err = req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		return resp, nil
	}

	r := &pbds.ListStrategiesReq{
		BizId:  req.BizId,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListStrategies(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("list strategies failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.ListStrategiesResp_RespData{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
