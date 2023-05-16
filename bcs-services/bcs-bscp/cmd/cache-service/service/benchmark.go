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

	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/cache-service"
	pbrci "bscp.io/pkg/protocol/core/released-ci"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// Node: the interface of this file is only used for internal basic interface stress testing.

// BenchAppMeta list app meta info.
func (s *Service) BenchAppMeta(ctx context.Context, req *pbcs.BenchAppMetaReq) (*pbcs.BenchAppMetaResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := req.Validate(); err != nil {
		return nil, err
	}

	metaMap, err := s.dao.App().ListAppMetaForCache(kt, req.BizId, req.AppIds)
	if err != nil {
		logs.Errorf("benchmark list app meta failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbcs.BenchAppMetaResp{
		Meta: pbcs.PbAppMetaMap(metaMap),
	}

	return resp, nil
}

// BenchAppCRIMeta list app current released instance meta.
func (s *Service) BenchAppCRIMeta(ctx context.Context, req *pbcs.BenchAppCRIMetaReq) (*pbcs.BenchAppCRIMetaResp,
	error) {

	kt := kit.FromGrpcContext(ctx)

	if err := req.Validate(); err != nil {
		return nil, err
	}

	metaMap, err := s.dao.CRInstance().ListAppCRIMeta(kt, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("benchmark list app current released instance failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbcs.BenchAppCRIMetaResp{
		Meta: pbcs.PbAppCRIMetas(metaMap),
	}

	return resp, nil
}

// BenchReleasedCI list released config item.
func (s *Service) BenchReleasedCI(ctx context.Context, req *pbcs.BenchReleasedCIReq) (*pbcs.BenchReleasedCIResp,
	error) {

	kt := kit.FromGrpcContext(ctx)

	if err := req.Validate(); err != nil {
		return nil, err
	}

	opts := &types.ListReleasedCIsOption{
		BizID: req.BizId,
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "release_id",
					Op:    filter.Equal.Factory(),
					Value: req.ReleaseId,
				},
			},
		},
		// use unlimited page.
		Page: &types.BasePage{Start: 0, Limit: 0},
	}

	cancel := kt.CtxWithTimeoutMS(500)
	defer cancel()

	detail, err := s.dao.ReleasedCI().List(kt, opts)
	if err != nil {
		logs.Errorf("benchmark list released config item failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbcs.BenchReleasedCIResp{
		Meta: pbrci.PbReleasedConfigItems(detail.Details),
	}

	return resp, nil
}

// BenchAppCPS list app current publish strategy.
func (s *Service) BenchAppCPS(ctx context.Context, req *pbcs.BenchAppCPSReq) (*pbcs.BenchAppCPSResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if err := req.Validate(); err != nil {
		return nil, err
	}

	opts := &types.GetAppCPSOption{
		BizID: req.BizId,
		AppID: req.AppId,
		Page: &types.BasePage{
			Count: false,
			Start: 0,
			Limit: types.GetCPSMaxPageLimit,
		},
	}
	appStrategies := make([]*types.PublishedStrategyCache, 0)
	for start := uint32(0); ; start += types.GetCPSMaxPageLimit {
		opts.Page.Start = start

		list, err := s.dao.Publish().GetAppCPStrategies(kt, opts)
		if err != nil {
			logs.Errorf("benchmark list app current published strategy failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		appStrategies = append(appStrategies, list...)

		if len(list) < types.GetCPSMaxPageLimit {
			break
		}
	}

	strategies, err := pbcs.PbPublishedStrategies(appStrategies)
	if err != nil {
		logs.Errorf("benchmark convert pb publish strategies failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	resp := &pbcs.BenchAppCPSResp{
		Meta: strategies,
	}

	return resp, nil
}
