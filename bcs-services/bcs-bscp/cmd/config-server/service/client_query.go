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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// CreateClientQuery create client query
func (s *Service) CreateClientQuery(ctx context.Context, req *pbcs.CreateClientQueryReq) (
	*pbcs.CreateClientQueryResp, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	search, err := s.client.DS.CreateClientQuery(kt.RpcCtx(), &pbds.CreateClientQueryReq{
		BizId:           req.BizId,
		AppId:           req.AppId,
		SearchType:      req.SearchType,
		SearchName:      req.SearchName,
		SearchCondition: req.SearchCondition,
	})
	if err != nil {
		return nil, err
	}

	resp := &pbcs.CreateClientQueryResp{
		Id: search.Id,
	}
	return resp, nil
}

// ListClientQuerys list client querys
func (s *Service) ListClientQuerys(ctx context.Context, req *pbcs.ListClientQuerysReq) (
	*pbcs.ListClientQuerysResp, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.View}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	items, err := s.client.DS.ListClientQuerys(kt.RpcCtx(), &pbds.ListClientQuerysReq{
		BizId:      req.BizId,
		AppId:      req.AppId,
		SearchType: req.SearchType,
		Start:      req.Start,
		Limit:      req.Limit,
		All:        req.All,
	})
	if err != nil {
		return nil, err
	}

	resp := &pbcs.ListClientQuerysResp{
		Details: items.Details,
		Count:   items.Count,
	}

	return resp, nil
}

// UpdateClientQuery update client query
func (s *Service) UpdateClientQuery(ctx context.Context, req *pbcs.UpdateClientQueryReq) (
	*pbcs.UpdateClientQueryResp, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.Id}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	_, err = s.client.DS.UpdateClientQuery(kt.RpcCtx(), &pbds.UpdateClientQueryReq{
		Id:              req.Id,
		BizId:           req.BizId,
		AppId:           req.AppId,
		SearchName:      req.SearchName,
		SearchCondition: req.SearchCondition,
	})

	if err != nil {
		return nil, err
	}

	return &pbcs.UpdateClientQueryResp{}, nil
}

// DeleteClientQuery delete client query
func (s *Service) DeleteClientQuery(ctx context.Context, req *pbcs.DeleteClientQueryReq) (
	*pbcs.DeleteClientQueryResp, error) {
	kt := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.App, Action: meta.View, ResourceID: req.Id}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(kt, res...)
	if err != nil {
		return nil, err
	}

	_, err = s.client.DS.DeleteClientQuery(kt.RpcCtx(), &pbds.DeleteClientQueryReq{
		Id:    req.Id,
		BizId: req.BizId,
		AppId: req.AppId,
	})
	if err != nil {
		return nil, err
	}

	return &pbcs.DeleteClientQueryResp{}, nil
}
