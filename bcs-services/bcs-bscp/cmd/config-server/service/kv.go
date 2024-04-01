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
	"errors"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbkv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/kv"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// CreateKv is used to create key-value data.
func (s *Service) CreateKv(ctx context.Context, req *pbcs.CreateKvReq) (*pbcs.CreateKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.CreateKvReq{
		Attachment: &pbkv.KvAttachment{
			BizId: grpcKit.BizID,
			AppId: req.AppId,
		},
		Spec: &pbkv.KvSpec{
			Key:    req.Key,
			Memo:   req.Memo,
			KvType: req.KvType,
			Value:  req.Value,
		},
	}
	rp, err := s.client.DS.CreateKv(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateKvResp{
		Id: rp.Id,
	}
	return resp, nil
}

// UpdateKv is used to update key-value data.
func (s *Service) UpdateKv(ctx context.Context, req *pbcs.UpdateKvReq) (*pbcs.UpdateKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.UpdateKvReq{
		Attachment: &pbkv.KvAttachment{
			BizId: grpcKit.BizID,
			AppId: req.AppId,
		},
		Spec: &pbkv.KvSpec{
			Key:   req.Key,
			Value: req.Value,
			Memo:  req.Memo,
		},
	}
	if _, err := s.client.DS.UpdateKv(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UpdateKvResp{}, nil

}

// ListKvs is used to list key-value data.
func (s *Service) ListKvs(ctx context.Context, req *pbcs.ListKvsReq) (*pbcs.ListKvsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListKvsReq{
		BizId:      req.BizId,
		AppId:      req.AppId,
		Key:        req.Key,
		Start:      req.Start,
		Limit:      req.Limit,
		All:        req.All,
		SearchKey:  req.SearchKey,
		WithStatus: req.WithStatus,
		KvType:     req.KvType,
		Sort:       req.Sort,
		Order:      req.Order,
		TopIds:     req.TopIds,
		Status:     req.Status,
	}
	if !req.All {
		if req.Limit == 0 {
			return nil, errors.New("limit has to be greater than 0")
		}
		r.Start = req.Start
		r.Limit = req.Limit
	}

	rp, err := s.client.DS.ListKvs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListKvsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil

}

// DeleteKv is used to delete key-value data.
func (s *Service) DeleteKv(ctx context.Context, req *pbcs.DeleteKvReq) (*pbcs.DeleteKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.DeleteKvReq{
		Id: req.Id,
		Attachment: &pbkv.KvAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	if _, err := s.client.DS.DeleteKv(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.DeleteKvResp{}, nil

}

// BatchDeleteKv is used to batch delete key-value data.
func (s *Service) BatchDeleteKv(ctx context.Context, req *pbcs.BatchDeleteKvReq) (*pbcs.BatchDeleteKvResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	if len(req.GetIds()) == 0 {
		return nil, errf.Errorf(grpcKit, errf.InvalidArgument, "ids is empty")
	}

	errGroup, errCtx := errgroup.WithContext(ctx)
	errGroup.SetLimit(10)

	successfulIDs := []uint32{}
	failedIDs := []uint32{}
	var mux sync.Mutex

	for _, v := range req.GetIds() {
		v := v
		errGroup.Go(func() error {
			r := &pbds.DeleteKvReq{
				Id: v,
				Attachment: &pbkv.KvAttachment{
					BizId: req.BizId,
					AppId: req.AppId,
				},
			}
			if _, err := s.client.DS.DeleteKv(errCtx, r); err != nil {
				logs.Errorf("delete kv failed, err: %v, rid: %s", err, grpcKit.Rid)

				// 错误不返回异常，记录错误ID
				mux.Lock()
				failedIDs = append(failedIDs, v)
				mux.Unlock()
				return nil
			}

			mux.Lock()
			successfulIDs = append(successfulIDs, v)
			mux.Unlock()

			return nil
		})
	}

	// 全部失败, 当前API视为失败
	if len(failedIDs) == len(req.Ids) {
		return nil, errf.Errorf(grpcKit, errf.Aborted, "batch delete failed")
	}

	return &pbcs.BatchDeleteKvResp{SuccessfulIds: successfulIDs, FailedIds: failedIDs}, nil
}

// BatchUpsertKvs is used to insert or update key-value data in bulk.
func (s *Service) BatchUpsertKvs(ctx context.Context, req *pbcs.BatchUpsertKvsReq) (*pbcs.BatchUpsertKvsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	kvs := make([]*pbds.BatchUpsertKvsReq_Kv, 0, len(req.Kvs))
	for _, kv := range req.Kvs {
		kvs = append(kvs, &pbds.BatchUpsertKvsReq_Kv{
			KvAttachment: &pbkv.KvAttachment{
				BizId: req.BizId,
				AppId: req.AppId,
			},
			KvSpec: &pbkv.KvSpec{
				Key:    kv.Key,
				KvType: kv.KvType,
				Value:  kv.Value,
			},
		})
	}

	r := &pbds.BatchUpsertKvsReq{
		BizId:      req.BizId,
		AppId:      req.AppId,
		Kvs:        kvs,
		ReplaceAll: true,
	}
	data, err := s.client.DS.BatchUpsertKvs(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("batch upsert kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.BatchUpsertKvsResp{Ids: data.Ids}, nil
}

// UnDeleteKv reverses the deletion of a key-value pair by reverting the current kvType and value to the previous
// version.
func (s *Service) UnDeleteKv(ctx context.Context, req *pbcs.UnDeleteKvReq) (*pbcs.UnDeleteKvResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.UnDeleteKvReq{
		Spec: &pbkv.KvSpec{
			Key: req.Key,
		},
		Attachment: &pbkv.KvAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	if _, err := s.client.DS.UnDeleteKv(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UnDeleteKvResp{}, nil
}
