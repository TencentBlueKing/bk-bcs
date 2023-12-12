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

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbkv "bscp.io/pkg/protocol/core/kv"
	pbrkv "bscp.io/pkg/protocol/core/released-kv"
	released_kv "bscp.io/pkg/protocol/core/released-kv"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// GetReleasedKv get released kv
func (s *Service) GetReleasedKv(ctx context.Context, req *pbds.GetReleasedKvReq) (*released_kv.ReleasedKv, error) {

	kt := kit.FromGrpcContext(ctx)

	rkv, err := s.dao.ReleasedKv().Get(kt, req.BizId, req.AppId, req.ReleaseId, req.Key)
	if err != nil {
		logs.Errorf("get released kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	kvType, value, err := s.getReleasedKv(kt, req.BizId, req.AppId, rkv.Spec.Version, req.ReleaseId, req.Key)
	if err != nil {
		logs.Errorf("get vault released kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &pbrkv.ReleasedKv{
		Id:        rkv.ID,
		ReleaseId: rkv.ReleaseID,
		Spec: &pbkv.KvSpec{
			Key:    rkv.Spec.Key,
			KvType: string(kvType),
			Value:  value,
		},
		Attachment: &pbkv.KvAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Revision: pbbase.PbRevision(rkv.Revision),
	}, nil

}

// ListReleasedKvs list app bound kv revisions.
func (s *Service) ListReleasedKvs(ctx context.Context, req *pbds.ListReleasedKvReq) (*pbds.ListReleasedKvResp, error) {

	kt := kit.FromGrpcContext(ctx)

	page := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	opt := &types.ListRKvOption{
		ReleaseID: req.ReleaseId,
		BizID:     req.BizId,
		AppID:     req.AppId,
		Key:       req.Key,
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
	details, count, err := s.dao.ReleasedKv().List(kt, opt)
	if err != nil {
		logs.Errorf("list released kv failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var rkvs []*pbrkv.ReleasedKv
	for _, detail := range details {
		_, val, err := s.getReleasedKv(kt, req.BizId, req.AppId, detail.Spec.Version, detail.ReleaseID, detail.Spec.Key)
		if err != nil {
			logs.Errorf("get vault released kv failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		rkvs = append(rkvs, pbrkv.PbRKv(detail, val))
	}

	resp := &pbds.ListReleasedKvResp{
		Count:   uint32(count),
		Details: rkvs,
	}
	return resp, nil

}

func (s *Service) getReleasedKv(kt *kit.Kit, bizID, appID, version, releasedID uint32,
	key string) (table.DataType, string, error) {

	opt := &types.GetRKvOption{
		BizID:      bizID,
		AppID:      appID,
		Key:        key,
		Version:    int(version),
		ReleasedID: releasedID,
	}
	return s.vault.GetRKv(kt, opt)
}
