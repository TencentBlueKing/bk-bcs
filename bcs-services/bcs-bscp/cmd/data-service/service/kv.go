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

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbkv "bscp.io/pkg/protocol/core/kv"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

func (s *Service) CreateKv(ctx context.Context, req *pbds.CreateKvReq) (*pbds.CreateResp, error) {

	kt := kit.FromGrpcContext(ctx)

	kv := &table.Kv{
		Spec:       req.Spec.KvSpec(),
		Attachment: req.Attachment.KvAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	id, err := s.dao.Kv().Create(kt, kv)
	if err != nil {
		logs.Errorf("create kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil

}

func (s *Service) UpdateKv(ctx context.Context, req *pbds.UpdateKvReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	kv := &table.Kv{
		ID:         req.Id,
		Spec:       req.Spec.KvSpec(),
		Attachment: req.Attachment.KvAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if err := s.dao.Kv().Update(kt, kv); err != nil {
		logs.Errorf("update kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil

}

func (s *Service) ListKv(ctx context.Context, req *pbds.ListKvReq) (*pbds.ListKvResp, error) {

	kt := kit.FromGrpcContext(ctx)

	page := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	opt := &types.ListKvOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		ID:        req.Id,
		SearchKey: req.SearchKey,
		All:       req.All,
		Page:      page,
	}

	po := &types.PageOption{
		EnableUnlimitedLimit: true,
	}
	if err := opt.Validate(po); err != nil {
		return nil, err
	}

	details, count, err := s.dao.Kv().List(kt, opt)

	if err != nil {
		logs.Errorf("list kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	kvs := pbkv.PbKvs(details)
	resp := &pbds.ListKvResp{
		Count:   uint32(count),
		Details: kvs,
	}
	return resp, nil

}

func (s *Service) DeleteKv(ctx context.Context, req *pbds.DeleteKvReq) (*pbbase.EmptyResp, error) {

	kt := kit.FromGrpcContext(ctx)

	kv := &table.Kv{
		ID:         req.Id,
		Attachment: req.Attachment.KvAttachment(),
	}
	if err := s.dao.Kv().Delete(kt, kv); err != nil {
		logs.Errorf("delete kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
