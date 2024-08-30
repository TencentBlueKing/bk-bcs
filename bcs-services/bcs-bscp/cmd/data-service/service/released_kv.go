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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbcontent "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	pbkv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/kv"
	pbrkv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/released-kv"
	released_kv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/released-kv"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
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
		Revision:    pbbase.PbRevision(rkv.Revision),
		ContentSpec: pbcontent.PbContentSpec(rkv.ContentSpec),
	}, nil

}

// ListReleasedKvs list app bound kv revisions.
func (s *Service) ListReleasedKvs(ctx context.Context, req *pbds.ListReleasedKvReq) (*pbds.ListReleasedKvResp, error) {

	kt := kit.FromGrpcContext(ctx)

	if len(req.Sort) == 0 {
		req.Sort = "key"
	}
	page := &types.BasePage{
		Start: req.Start,
		Limit: uint(req.Limit),
		Sort:  req.Sort,
		Order: types.Order(req.Order),
	}
	opt := &types.ListRKvOption{
		ReleaseID: req.ReleaseId,
		BizID:     req.BizId,
		AppID:     req.AppId,
		Key:       req.Key,
		SearchKey: req.SearchKey,
		All:       req.All,
		Page:      page,
		KvType:    req.KvType,
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
		// value 是否隐藏
		if detail.Spec.SecretHidden {
			val = i18n.T(kt, "sensitive data is not visible, unable to view actual content")
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
