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
	"bscp.io/pkg/protocol/core/strategy-set"
	"bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateStrategySet create a strategy set
func (s *Service) CreateStrategySet(ctx context.Context, req *pbcs.CreateStrategySetReq) (
	*pbcs.CreateStrategySetResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateStrategySetResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.StrategySet, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return resp, nil
	}

	r := &pbds.CreateStrategySetReq{
		Attachment: &pbss.StrategySetAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbss.StrategySetSpec{
			Name: req.Name,
			Memo: req.Memo,
		},
	}
	rp, err := s.client.DS.CreateStrategySet(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("create strategy set failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.CreateStrategySetResp_RespData{
		Id: rp.Id,
	}
	return resp, nil
}

// UpdateStrategySet update a strategy set
func (s *Service) UpdateStrategySet(ctx context.Context, req *pbcs.UpdateStrategySetReq) (
	*pbcs.UpdateStrategySetResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateStrategySetResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.StrategySet, Action: meta.Update,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return resp, nil
	}

	r := &pbds.UpdateStrategySetReq{
		Id: req.Id,
		Attachment: &pbss.StrategySetAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbss.StrategySetSpec{
			Name: req.Name,
			Memo: req.Memo,
		},
	}
	_, err = s.client.DS.UpdateStrategySet(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("update strategy set failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	return resp, nil
}

// PublishStrategySet publish the strategy set
func (s *Service) PublishStrategySet(ctx context.Context, req *pbcs.PublishStrategySetReq) (
	*pbcs.PublishStrategySetResp, error) {

	return nil, nil
}

// FinishPublishStrategySet finish the published strategy set
func (s *Service) FinishPublishStrategySet(ctx context.Context, req *pbcs.FinishPublishStrategySetReq) (
	*pbcs.FinishPublishStrategySetResp, error) {

	return nil, nil
}

// ListStrategySets list strategy set with filter.
func (s *Service) ListStrategySets(ctx context.Context, req *pbcs.ListStrategySetsReq) (
	*pbcs.ListStrategySetsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListStrategySetsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.StrategySet, Action: meta.Find}, BizID: req.BizId}
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

	r := &pbds.ListStrategySetsReq{
		BizId:  req.BizId,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListStrategySets(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("list strategy sets failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.ListStrategySetsResp_RespData{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// DeleteStrategySet delete a strategy set
func (s *Service) DeleteStrategySet(ctx context.Context, req *pbcs.DeleteStrategySetReq) (*pbcs.DeleteStrategySetResp,
	error) {

	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteStrategySetResp)

	if err := req.Validate(); err != nil {
		errf.Error(err).AssignResp(kt, resp)
		logs.Errorf("request validate failed, err: %v, rid: %s", err, kt.Rid)
		return resp, nil
	}

	r := &pbds.DeleteStrategySetReq{
		Id: req.Id,
		Attachment: &pbss.StrategySetAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	_, err := s.client.DS.DeleteStrategySet(kt.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kt, resp)
		logs.Errorf("delete strategy set failed, err: %v, rid: %s", err, kt.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	return resp, nil
}
