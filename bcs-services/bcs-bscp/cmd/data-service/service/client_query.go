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
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"gorm.io/gorm"

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

	if req.SearchName != "" {
		query, errQ := s.dao.ClientQuery().GetBySearchName(grpcKit, req.GetBizId(), req.GetAppId(),
			grpcKit.User, req.SearchName)
		if errQ != nil && !errors.Is(errQ, gorm.ErrRecordNotFound) {
			return nil, errQ
		}
		if query != nil {
			return nil, errors.New("search name already exists")
		}
	}

	data, err := s.dao.ClientQuery().ListBySearchCondition(grpcKit, req.GetBizId(), req.GetAppId(),
		grpcKit.User, req.SearchType, string(searchCondition))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	oldData := new(table.ClientQuery)
	if len(data) > 0 {
		// 对比两个json串是否一致
		var obj1 map[string]interface{}
		_ = json.Unmarshal(searchCondition, &obj1)
		for _, v := range data {
			var obj2 map[string]interface{}
			_ = json.Unmarshal([]byte(v.Spec.SearchCondition), &obj2)
			if reflect.DeepEqual(obj1, obj2) {
				oldData = v
			}
		}
	}

	var id uint32
	// 更新
	if oldData.ID > 0 {
		err = s.dao.ClientQuery().Update(grpcKit, &table.ClientQuery{
			ID:         oldData.ID,
			Attachment: oldData.Attachment,
			Spec: &table.ClientQuerySpec{
				Creator:         grpcKit.User,
				SearchName:      req.SearchName,
				SearchType:      table.SearchType(req.SearchType),
				SearchCondition: string(searchCondition),
				UpdatedAt:       time.Now().UTC(),
			},
		})
		if err != nil {
			return nil, err
		}
		id = oldData.ID
	} else {
		id, err = s.dao.ClientQuery().Create(grpcKit, &table.ClientQuery{
			Spec: &table.ClientQuerySpec{
				Creator:         grpcKit.User,
				SearchName:      req.SearchName,
				SearchType:      table.SearchType(req.SearchType),
				SearchCondition: string(searchCondition),
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			},
			Attachment: &table.ClientQueryAttachment{
				BizID: req.BizId,
				AppID: req.AppId,
			},
		})
		if err != nil {
			return nil, err
		}
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

	if req.SearchName != "" {
		query, errQ := s.dao.ClientQuery().GetBySearchName(grpcKit, req.GetBizId(), req.GetAppId(),
			grpcKit.User, req.SearchName)
		if errQ != nil && !errors.Is(errQ, gorm.ErrRecordNotFound) {
			return nil, errQ
		}
		if query != nil && query.ID != req.Id {
			return nil, errors.New("search name already exists")
		}
	}

	err = s.dao.ClientQuery().Update(grpcKit, &table.ClientQuery{
		ID: req.Id,
		Spec: &table.ClientQuerySpec{
			Creator:         grpcKit.User,
			SearchName:      req.SearchName,
			SearchCondition: string(searchCondition),
			UpdatedAt:       time.Now().UTC(),
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
