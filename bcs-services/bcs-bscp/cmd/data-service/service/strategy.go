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
	"time"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	release "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/release"
	pbstrategy "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/strategy"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// GetLastSelect get last version select publish_type
func (s *Service) GetLastSelect(ctx context.Context, req *pbds.GetLastSelectReq) (*pbds.GetLastSelectResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	resp := &pbds.GetLastSelectResp{
		PublishType: "",
		IsApprove:   false,
	}

	app, err := s.dao.App().Get(grpcKit, req.BizId, req.AppId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, nil
		}
		logs.Errorf("get app failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	resp.IsApprove = app.Spec.IsApprove

	strategy := s.dao.GenQuery().Strategy
	strategyRecord, err := strategy.WithContext(ctx).Where(
		strategy.AppID.Eq(req.AppId), strategy.BizID.Eq(req.BizId)).Last()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, nil
		}
		logs.Errorf("get strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp.PublishType = string(strategyRecord.Spec.PublishType)

	// if approval is required, there will be no immediate type
	if app.Spec.IsApprove && strategyRecord.Spec.PublishType == table.Immediately {
		// default value
		resp.PublishType = string(table.Manually)
	}

	// if approval is not required, it will only be immediately and periodical publish
	if !app.Spec.IsApprove && (strategyRecord.Spec.PublishType == table.Immediately ||
		strategyRecord.Spec.PublishType == table.Periodically) {
		// default value
		resp.PublishType = string(table.Immediately)
	}
	return resp, nil
}

// GetLastPublish get last publish list
func (s *Service) GetLastPublish(ctx context.Context, req *pbds.GetLastPublishReq) (*pbds.GetLastPublishResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	resp := &pbds.GetLastPublishResp{
		IsPublishing:  false,
		VersionName:   "",
		PublishRecord: []*release.PublishRecord{},
	}

	strategyRecord, err := s.dao.Strategy().GetLast(grpcKit, req.BizId, req.AppId, 0, 0)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, nil
		}
		logs.Errorf("get strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// the current service has publishing tasks
	if strategyRecord.Spec.PublishStatus == table.PendApproval ||
		strategyRecord.Spec.PublishStatus == table.PendPublish {
		resp.IsPublishing = true
	}

	releaseRecord, err := s.dao.Release().Get(grpcKit, req.GetBizId(), req.GetAppId(), strategyRecord.Spec.ReleaseID)
	if err != nil {
		logs.Errorf("get release failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp.VersionName = releaseRecord.Spec.Name
	resp.UpdatedAt = strategyRecord.Revision.UpdatedAt.Format(time.DateTime)

	lrs, err := s.dao.Release().ListReleaseStrategies(grpcKit, req.BizId, req.AppId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, nil
		}
		logs.Errorf("list release strategie failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	var publishRecords []*release.PublishRecord
	for _, v := range lrs {
		publishRecords = append(publishRecords, &release.PublishRecord{
			PublishTime:   v.PublishTime,
			Name:          v.Name,
			Scope:         pbstrategy.PbScope(&v.Scope),
			Creator:       v.Creator,
			FullyReleased: v.FullyReleased,
		})
	}
	resp.PublishRecord = publishRecords
	return resp, nil
}

// GetReleasesStatus get last releases status
func (s *Service) GetReleasesStatus(ctx context.Context, req *pbds.GetReleasesStatusReq) (*pbstrategy.Strategy, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	strategy, err := s.dao.Strategy().GetLast(grpcKit, req.BizId, req.AppId, req.ReleaseId, 0)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &pbstrategy.Strategy{
				Spec: &pbstrategy.StrategySpec{},
			}, nil
		}
		logs.Errorf("get strategy last failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	releasedGroups, err := s.dao.ReleasedGroup().ListAllByReleaseID(grpcKit, strategy.Attachment.BizID, req.ReleaseId)
	if err != nil {
		logs.Errorf("list all by release id failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	app, err := s.dao.App().GetByID(grpcKit, req.AppId)
	if err != nil {
		logs.Errorf("get app by id failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// 该版本曾经上过线，后被分组重新上线覆盖了
	if len(releasedGroups) == 0 && strategy.Spec.PublishStatus == table.AlreadyPublish {
		strategy.Spec.PublishStatus = ""
		strategy.Spec.Approver = ""
		strategy.Spec.ApproverProgress = ""
		strategy.Spec.PublishTime = ""
		strategy.Spec.PublishType = ""
	}

	resp := pbstrategy.Strategy{
		Id:         strategy.ID,
		Spec:       pbstrategy.PbStrategySpec(strategy.Spec),
		Status:     pbstrategy.PbStrategyState(strategy.State),
		Attachment: pbstrategy.PbStrategyAttachment(strategy.Attachment),
		Revision:   pbstrategy.PbRevision(strategy.Revision),
		App:        &pbstrategy.AppSpec{Creator: app.Revision.Creator},
	}

	return &resp, nil
}
