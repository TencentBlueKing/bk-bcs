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
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbcq "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client-query"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateClientQuery create client query
func (s *Service) CreateClientQuery(ctx context.Context, req *pbds.CreateClientQueryReq) (
	*pbds.CreateClientQueryResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	searchCondition, err := req.GetSearchCondition().MarshalJSON()
	if err != nil {
		return nil, err
	}

	id, err := s.dao.ClientQuery().Create(grpcKit, &table.ClientQuery{
		Spec: &table.ClientQuerySpec{
			Creator:         grpcKit.User,
			SearchName:      req.SearchName,
			SearchType:      table.SearchType(req.SearchType),
			SearchCondition: string(searchCondition),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		Attachment: &table.ClientQueryAttachment{
			BizID: req.BizId,
			AppID: req.AppId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &pbds.CreateClientQueryResp{
		Id: id,
	}, nil
}

// ListClientQuerys list client querys
func (s *Service) ListClientQuerys(ctx context.Context, req *pbds.ListClientQuerysReq) (
	*pbds.ListClientQuerysResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	items, count, err := s.dao.ClientQuery().List(grpcKit, req.BizId, req.AppId, grpcKit.User, req.SearchType,
		&types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All})
	if err != nil {
		return nil, err
	}

	resp := &pbds.ListClientQuerysResp{
		Count:   uint32(count),
		Details: pbcq.PbClientQuerys(items),
	}
	return resp, nil
}

// UpdateClientQuery update client query
func (s *Service) UpdateClientQuery(ctx context.Context, req *pbds.UpdateClientQueryReq) (
	*pbbase.EmptyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	searchCondition, err := req.GetSearchCondition().MarshalJSON()
	if err != nil {
		return nil, err
	}

	err = s.dao.ClientQuery().Update(grpcKit, &table.ClientQuery{
		ID: req.Id,
		Spec: &table.ClientQuerySpec{
			Creator:         grpcKit.User,
			SearchName:      req.SearchName,
			SearchCondition: string(searchCondition),
			UpdatedAt:       time.Now(),
		},
		Attachment: &table.ClientQueryAttachment{
			BizID: req.BizId,
			AppID: req.AppId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &pbbase.EmptyResp{}, nil
}

// DeleteClientQuery delete client query
func (s *Service) DeleteClientQuery(ctx context.Context, req *pbds.DeleteClientQueryReq) (
	*pbbase.EmptyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	err := s.dao.ClientQuery().Delete(grpcKit, &table.ClientQuery{
		ID: req.Id,
		Attachment: &table.ClientQueryAttachment{
			BizID: req.BizId,
			AppID: req.AppId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &pbbase.EmptyResp{}, nil
}
