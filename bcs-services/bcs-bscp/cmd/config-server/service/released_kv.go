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

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbds "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// GetReleasedKv get released kv
func (s *Service) GetReleasedKv(ctx context.Context, req *pbcs.GetReleasedKvReq) (*pbcs.GetReleasedKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	grciReq := &pbds.GetReleasedKvReq{
		BizId:     req.BizId,
		AppId:     req.AppId,
		ReleaseId: req.ReleaseId,
		Key:       req.Key,
	}
	releasedKv, err := s.client.DS.GetReleasedKv(grpcKit.RpcCtx(), grciReq)
	if err != nil {
		logs.Errorf("get released kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.GetReleasedKvResp{
		Kv: releasedKv,
	}
	return resp, nil

}

// ListReleasedKvs list released kvs
func (s *Service) ListReleasedKvs(ctx context.Context, req *pbcs.ListReleasedKvsReq) (*pbcs.ListReleasedKvsResp,
	error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.AppId}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListReleasedKvReq{
		BizId:     grpcKit.BizID,
		AppId:     req.AppId,
		Start:     req.Start,
		Limit:     req.Limit,
		All:       req.All,
		ReleaseId: req.ReleaseId,
		SearchKey: req.SearchKey,
		Key:       req.Key,
		KvType:    req.KvType,
		Sort:      req.Sort,
		Order:     req.Order,
	}
	rkv, err := s.client.DS.ListReleasedKvs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListReleasedKvsResp{
		Count:   rkv.Count,
		Details: rkv.Details,
	}
	return resp, nil
}
