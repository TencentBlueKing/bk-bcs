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
	"errors"
	"fmt"
	"time"

	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbstrategy "bscp.io/pkg/protocol/core/strategy"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// Publish exec publish strategy.
func (s *Service) Publish(ctx context.Context, req *pbds.PublishReq) (*pbds.PublishResp, error) {

	kt := kit.FromGrpcContext(ctx)

	opt := &types.PublishOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		ReleaseID: req.ReleaseId,
		All:       req.All,
		Default:   req.Default,
		Memo:      req.Memo,
		Groups:    req.Groups,
		Revision: &table.CreatedRevision{
			Creator:   kt.User,
			CreatedAt: time.Now(),
		},
	}

	if err := s.validatePublishGroups(kt, opt); err != nil {
		return nil, err
	}

	pshID, err := s.dao.Publish().Publish(kt, opt)
	if err != nil {
		logs.Errorf("publish strategy failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.PublishResp{PublishedStrategyHistoryId: pshID}
	return resp, nil
}

// GenerateReleaseAndPublish generate release and publish.
func (s *Service) GenerateReleaseAndPublish(ctx context.Context, req *pbds.GenerateReleaseAndPublishReq) (
	*pbds.PublishResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	// step1: validate and query group ids.
	groupIDs := make([]uint32, 0)
	if !req.All {
		if len(req.Groups) == 0 {
			return nil, fmt.Errorf("groups can't be empty when publish not all")
		}
		for _, name := range req.Groups {
			group, e := s.dao.Group().GetByName(grpcKit, req.BizId, name)
			if e != nil {
				return nil, fmt.Errorf("group %s not exist", name)
			}
			groupIDs = append(groupIDs, group.ID)
		}
	}

	releasedCIs := make([]*table.ReleasedConfigItem, 0)
	// TODO: need to change batch operator to query config item and it's commit.
	// step2: query app's all config items.
	cfgItems, err := s.queryAppConfigItemList(grpcKit, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("query app config item list failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// if no config item, return directly.
	if len(cfgItems) == 0 {
		return nil, errors.New("app config items is empty")
	}

	// step3: query config item newest commit
	now := time.Now()
	for _, item := range cfgItems {
		commit, e := s.queryCILatestCommit(grpcKit, req.BizId, req.AppId, item.ID)
		if e != nil {
			logs.Errorf("query config item latest commit failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
		}

		releasedCIs = append(releasedCIs, &table.ReleasedConfigItem{
			CommitID:       commit.ID,
			CommitSpec:     commit.Spec,
			ConfigItemID:   item.ID,
			ConfigItemSpec: item.Spec,
			Attachment:     item.Attachment,
			Revision:       item.Revision,
		})
	}

	if _, e := s.dao.Release().GetByName(grpcKit, req.BizId, req.AppId, req.ReleaseName); e == nil {
		return nil, fmt.Errorf("release name %s already exists", req.ReleaseName)
	}

	// step4: begin transaction to create release and released config item.
	tx, err := s.dao.BeginTx(grpcKit, req.BizId)
	if err != nil {
		return nil, err
	}
	// step5: create release, and create release and released config item need to begin tx.
	release := &table.Release{
		// Spec:       req.Spec.ReleaseSpec(),
		Spec: &table.ReleaseSpec{
			Name: req.ReleaseName,
			Memo: req.ReleaseMemo,
		},
		Attachment: &table.ReleaseAttachment{
			BizID: req.BizId,
			AppID: req.AppId,
		},
		Revision: &table.CreatedRevision{
			Creator:   grpcKit.User,
			CreatedAt: now,
		},
	}
	releaseID, err := s.dao.Release().CreateWithTx(grpcKit, tx, release)
	if err != nil {
		logs.Errorf("create release failed, err: %v, rid: %s", err, grpcKit.Rid)
		tx.Rollback(grpcKit)
		return nil, err
	}

	// step6: create released config item.
	for _, rci := range releasedCIs {
		rci.ReleaseID = releaseID
	}

	if err = s.dao.ReleasedCI().BulkCreateWithTx(grpcKit, tx, releasedCIs); err != nil {
		logs.Errorf("bulk create released config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		tx.Rollback(grpcKit)
		return nil, err
	}

	// step7: publish with transaction.

	kt := kit.FromGrpcContext(ctx)

	opt := &types.PublishOption{
		BizID:     req.BizId,
		AppID:     req.AppId,
		ReleaseID: releaseID,
		All:       req.All,
		Memo:      req.ReleaseMemo,
		Groups:    groupIDs,
		Revision: &table.CreatedRevision{
			Creator:   kt.User,
			CreatedAt: time.Now(),
		},
	}
	if e := s.validatePublishGroups(kt, opt); e != nil {
		return nil, e
	}
	pshID, err := s.dao.Publish().PublishWithTx(kt, tx, opt)
	if err != nil {
		logs.Errorf("publish strategy failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback(kt)
		return nil, err
	}

	// step8: commit transaction.
	if err = tx.Commit(kt); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.PublishResp{PublishedStrategyHistoryId: pshID}
	return resp, nil
}

// FinishPublish finish publish strategy.
func (s *Service) FinishPublish(ctx context.Context, req *pbds.FinishPublishReq) (
	*pbbase.EmptyResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	opt := &types.FinishPublishOption{
		BizID:      req.BizId,
		AppID:      req.AppId,
		StrategyID: req.StrategyId,
	}
	err := s.dao.Publish().FinishPublish(grpcKit, opt)
	if err != nil {
		logs.Errorf("finish publish strategy failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// ListPublishedStrategyHistories list published strategy histories.
func (s *Service) ListPublishedStrategyHistories(ctx context.Context, req *pbds.ListPubStrategyHistoriesReq) (
	*pbds.ListPubStrategyHistoriesResp, error) {

	kt := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	ft, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	query := &types.ListPSHistoriesOption{
		BizID:  req.BizId,
		AppID:  req.AppId,
		Filter: ft,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.Publish().ListPSHistory(kt, query)
	if err != nil {
		logs.Errorf("list published strategy history failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	strategies, err := pbstrategy.PbPubStrategyHistories(details.Details)
	if err != nil {
		logs.Errorf("get pb strategy histories failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListPubStrategyHistoriesResp{
		Count:   details.Count,
		Details: strategies,
	}
	return resp, nil
}

func (s *Service) validatePublishGroups(kt *kit.Kit, opt *types.PublishOption) error {
	for _, groupID := range opt.Groups {
		// frontend would set groupID 0 as default.
		if groupID == 0 {
			opt.Default = true
			continue
		}
		group, e := s.dao.Group().Get(kt, groupID, opt.BizID)
		if e != nil {
			if e == orm.ErrRecordNotFound {
				return fmt.Errorf("group %d not exists", groupID)
			}
			return e
		}
		if group.Spec.Public {
			continue
		}
		if _, e := s.dao.GroupAppBind().Get(kt, groupID, opt.AppID, opt.BizID); e != nil {
			if e == orm.ErrRecordNotFound {
				return fmt.Errorf("group %d not bind app %d", groupID, opt.AppID)
			}
			return e
		}
	}
	return nil
}
