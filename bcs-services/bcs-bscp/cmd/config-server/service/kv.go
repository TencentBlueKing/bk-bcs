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

	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbkv "bscp.io/pkg/protocol/core/kv"
	pbds "bscp.io/pkg/protocol/data-service"
)

func (s *Service) CreateKv(ctx context.Context, req *pbcs.CreateKvReq) (*pbcs.CreateKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	r := &pbds.CreateKvReq{
		Attachment: &pbkv.KvAttachment{
			BizId: grpcKit.BizID,
			AppId: req.AppId,
		},
		Spec: &pbkv.KvSpec{
			Name:   req.Name,
			KvType: req.Type,
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

func (s *Service) UpdateKv(ctx context.Context, req *pbcs.UpdateKvReq) (*pbcs.UpdateKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	r := &pbds.UpdateKvReq{
		Id: req.KvId,
		Attachment: &pbkv.KvAttachment{
			BizId: grpcKit.BizID,
			AppId: req.AppId,
		},
		Spec: &pbkv.KvSpec{
			Name:  req.Name,
			Value: req.Value,
		},
	}
	if _, err := s.client.DS.UpdateKv(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.UpdateKvResp{}, nil

}

func (s *Service) ListKv(ctx context.Context, req *pbcs.ListKvReq) (*pbcs.ListKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	r := &pbds.ListKvReq{
		BizId:     grpcKit.BizID,
		AppId:     grpcKit.AppID,
		Id:        req.Id,
		Start:     req.Start,
		Limit:     req.Limit,
		All:       req.All,
		SearchKey: req.SearchKey,
	}

	rp, err := s.client.DS.ListKv(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListKvResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil

}

func (s *Service) DeleteKv(ctx context.Context, req *pbcs.DeleteKvReq) (*pbcs.DeleteKvResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	r := &pbds.DeleteKvReq{
		Id: req.KvId,
		Attachment: &pbkv.KvAttachment{
			BizId: grpcKit.BizID,
			AppId: grpcKit.AppID,
		},
	}
	if _, err := s.client.DS.DeleteKv(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete kv failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return &pbcs.DeleteKvResp{}, nil

}
