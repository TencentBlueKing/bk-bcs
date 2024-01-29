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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/cache-service"
	pbrci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/released-ci"
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

// BenchReleasedCI list released config item.
func (s *Service) BenchReleasedCI(ctx context.Context, req *pbcs.BenchReleasedCIReq) (*pbcs.BenchReleasedCIResp,
	error) {

	kt := kit.FromGrpcContext(ctx)

	if err := req.Validate(); err != nil {
		return nil, err
	}

	cancel := kt.CtxWithTimeoutMS(500)
	defer cancel()

	list, err := s.dao.ReleasedCI().ListAll(kt, req.BizId)
	if err != nil {
		logs.Errorf("benchmark list released config item failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbcs.BenchReleasedCIResp{
		Meta: pbrci.PbReleasedConfigItems(list),
	}

	return resp, nil
}
