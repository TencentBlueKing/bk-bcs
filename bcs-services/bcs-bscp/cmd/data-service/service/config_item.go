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
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbrci "bscp.io/pkg/protocol/core/released-ci"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// CreateConfigItem create config item.
func (s *Service) CreateConfigItem(ctx context.Context, req *pbds.CreateConfigItemReq) (*pbds.CreateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	tx, err := s.dao.BeginTx(grpcKit, req.ConfigItemAttachment.BizId)
	if err != nil {
		logs.Errorf("create config item, begin transaction failed, err: %v", err)
		return nil, err
	}
	now := time.Now()
	// 1. create config item.
	ci := &table.ConfigItem{
		Spec:       req.ConfigItemSpec.ConfigItemSpec(),
		Attachment: req.ConfigItemAttachment.ConfigItemAttachment(),
		Revision: &table.Revision{
			Creator:   grpcKit.User,
			Reviser:   grpcKit.User,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	ciID, err := s.dao.ConfigItem().CreateWithTx(grpcKit, tx, ci)
	if err != nil {
		logs.Errorf("create config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		tx.Rollback(grpcKit)
		return nil, err
	}
	// 2. create content.
	content := &table.Content{
		Spec: req.ContentSpec.ContentSpec(),
		Attachment: &table.ContentAttachment{
			BizID:        req.ConfigItemAttachment.BizId,
			AppID:        req.ConfigItemAttachment.AppId,
			ConfigItemID: ciID,
		},
		Revision: &table.CreatedRevision{
			Creator:   grpcKit.User,
			CreatedAt: now,
		},
	}
	contentID, err := s.dao.Content().CreateWithTx(grpcKit, tx, content)
	if err != nil {
		logs.Errorf("create content failed, err: %v, rid: %s", err, grpcKit.Rid)
		tx.Rollback(grpcKit)
		return nil, err
	}
	// 3. create commit.
	commit := &table.Commit{
		Spec: &table.CommitSpec{
			ContentID: contentID,
			Content:   content.Spec,
		},
		Attachment: &table.CommitAttachment{
			BizID:        req.ConfigItemAttachment.BizId,
			AppID:        req.ConfigItemAttachment.AppId,
			ConfigItemID: ciID,
		},
		Revision: &table.CreatedRevision{
			Creator:   grpcKit.User,
			CreatedAt: now,
		},
	}
	_, err = s.dao.Commit().CreateWithTx(grpcKit, tx, commit)
	if err != nil {
		logs.Errorf("create commit failed, err: %v, rid: %s", err, grpcKit.Rid)
		tx.Rollback(grpcKit)
		return nil, err
	}
	tx.Commit(grpcKit)

	resp := &pbds.CreateResp{Id: ciID}
	return resp, nil
}

// BatchUpsertConfigItems batch upsert config items.
func (s *Service) BatchUpsertConfigItems(ctx context.Context, req *pbds.BatchUpsertConfigItemsReq) (
	*pbbase.EmptyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	tx, err := s.dao.BeginTx(grpcKit, req.BizId)
	if err != nil {
		logs.Errorf("create config item, begin transaction failed, err: %v", err)
		return nil, err
	}
	now := time.Now()
	// 1. truncate app config items.
	if e := s.dao.ConfigItem().TruncateWithTx(grpcKit, tx, req.BizId, req.AppId); e != nil {
		logs.Errorf("truncate app config items failed, err: %v, rid: %s", e, grpcKit.Rid)
		tx.Rollback(grpcKit)
		return nil, e
	}
	// 2. create config items.
	for _, item := range req.Items {
		// 2.1 create config item.
		ci := &table.ConfigItem{
			Spec:       item.ConfigItemSpec.ConfigItemSpec(),
			Attachment: item.ConfigItemAttachment.ConfigItemAttachment(),
			Revision: &table.Revision{
				Creator:   grpcKit.User,
				Reviser:   grpcKit.User,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
		ciID, err := s.dao.ConfigItem().CreateWithTx(grpcKit, tx, ci)
		if err != nil {
			logs.Errorf("create config item failed, err: %v, rid: %s", err, grpcKit.Rid)
			tx.Rollback(grpcKit)
			return nil, err
		}
		// 2. create content.
		content := &table.Content{
			Spec: item.ContentSpec.ContentSpec(),
			Attachment: &table.ContentAttachment{
				BizID:        req.BizId,
				AppID:        req.AppId,
				ConfigItemID: ciID,
			},
			Revision: &table.CreatedRevision{
				Creator:   grpcKit.User,
				CreatedAt: now,
			},
		}
		contentID, err := s.dao.Content().CreateWithTx(grpcKit, tx, content)
		if err != nil {
			logs.Errorf("create content failed, err: %v, rid: %s", err, grpcKit.Rid)
			tx.Rollback(grpcKit)
			return nil, err
		}
		// 3. create commit.
		commit := &table.Commit{
			Spec: &table.CommitSpec{
				ContentID: contentID,
				Content:   content.Spec,
			},
			Attachment: &table.CommitAttachment{
				BizID:        req.BizId,
				AppID:        req.AppId,
				ConfigItemID: ciID,
			},
			Revision: &table.CreatedRevision{
				Creator:   grpcKit.User,
				CreatedAt: now,
			},
		}
		_, err = s.dao.Commit().CreateWithTx(grpcKit, tx, commit)
		if err != nil {
			logs.Errorf("create commit failed, err: %v, rid: %s", err, grpcKit.Rid)
			tx.Rollback(grpcKit)
			return nil, err
		}
	}
	tx.Commit(grpcKit)

	return new(pbbase.EmptyResp), nil
}

// UpdateConfigItem update config item.
func (s *Service) UpdateConfigItem(ctx context.Context, req *pbds.UpdateConfigItemReq) (
	*pbbase.EmptyResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	ci := &table.ConfigItem{
		ID:         req.Id,
		Spec:       req.Spec.ConfigItemSpec(),
		Attachment: req.Attachment.ConfigItemAttachment(),
		Revision: &table.Revision{
			Reviser:   grpcKit.User,
			UpdatedAt: time.Now(),
		},
	}
	if err := s.dao.ConfigItem().Update(grpcKit, ci); err != nil {
		logs.Errorf("update config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// DeleteConfigItem delete config item.
func (s *Service) DeleteConfigItem(ctx context.Context, req *pbds.DeleteConfigItemReq) (*pbbase.EmptyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	ci := &table.ConfigItem{
		ID:         req.Id,
		Attachment: req.Attachment.ConfigItemAttachment(),
	}
	if err := s.dao.ConfigItem().Delete(grpcKit, ci); err != nil {
		logs.Errorf("delete config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// GetConfigItem get config item detail
func (s *Service) GetConfigItem(ctx context.Context, req *pbds.GetConfigItemReq) (*pbci.ConfigItem, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	configItem, err := s.dao.ConfigItem().Get(grpcKit, req.Id, req.BizId)
	if err != nil {
		logs.Errorf("get config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	resp := pbci.PbConfigItem(configItem, "")
	return resp, nil
}

// ListConfigItems list config items by query condition.
func (s *Service) ListConfigItems(ctx context.Context, req *pbds.ListConfigItemsReq) (*pbds.ListConfigItemsResp,
	error) {

	grpcKit := kit.FromGrpcContext(ctx)

	if req.ReleaseId == 0 {
		// list editing config items
		query := &types.ListConfigItemsOption{
			BizID: req.BizId,
			AppID: req.AppId,
			Filter: &filter.Expression{
				Op:    filter.Or,
				Rules: []filter.RuleFactory{},
			},
			Page: &types.BasePage{
				Start: req.Start,
				Limit: uint(req.Limit),
			},
		}
		if req.SearchKey != "" {
			query.Filter.Rules = append(query.Filter.Rules, &filter.AtomRule{
				Field: "name",
				Op:    filter.ContainsInsensitive.Factory(),
				Value: req.SearchKey,
			}, &filter.AtomRule{
				Field: "creator",
				Op:    filter.ContainsInsensitive.Factory(),
				Value: req.SearchKey,
			}, &filter.AtomRule{
				Field: "reviser",
				Op:    filter.ContainsInsensitive.Factory(),
				Value: req.SearchKey,
			})
		}
		if req.All {
			query.Page.Start = 0
			query.Page.Limit = 0
		}
		details, err := s.dao.ConfigItem().List(grpcKit, query)
		if err != nil {
			logs.Errorf("list editing config items failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}

		fileReleased, err := s.dao.ReleasedCI().GetReleasedLately(grpcKit, req.AppId, req.BizId, req.SearchKey)
		if err != nil {
			logs.Errorf("get released failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
		configItems, count := s.queryConfigItemsWithDeleted(details, fileReleased, req.Start, req.Limit, req.All)
		resp := &pbds.ListConfigItemsResp{
			Count:   count,
			Details: configItems,
		}
		return resp, nil
	}
	// list released config items
	query := &types.ListReleasedCIsOption{
		BizID: req.BizId,
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{Field: "release_id", Op: filter.Equal.Factory(), Value: req.ReleaseId},
			},
		},
		Page: &types.BasePage{
			Start: req.Start,
			Limit: uint(req.Limit),
		},
	}
	details, err := s.dao.ReleasedCI().List(grpcKit, query)
	if err != nil {
		logs.Errorf("list released config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	resp := &pbds.ListConfigItemsResp{
		Count:   details.Count,
		Details: pbrci.PbConfigItems(details.Details),
	}
	return resp, nil

}

// queryAppConfigItemList query config item list under specific app.
func (s *Service) queryAppConfigItemList(kit *kit.Kit, bizID, appID uint32) ([]*table.ConfigItem, error) {
	cfgItems := make([]*table.ConfigItem, 0)
	f := &filter.Expression{
		Op:    filter.And,
		Rules: []filter.RuleFactory{},
	}

	const step = 200
	for start := uint32(0); ; start += step {
		opt := &types.ListConfigItemsOption{
			BizID:  bizID,
			AppID:  appID,
			Filter: f,
			Page: &types.BasePage{
				Count: false,
				Start: start,
				Limit: step,
				Sort:  "id",
			},
		}

		details, err := s.dao.ConfigItem().List(kit, opt)
		if err != nil {
			return nil, err
		}

		cfgItems = append(cfgItems, details.Details...)

		if len(details.Details) < step {
			break
		}
	}

	return cfgItems, nil
}

// queryAppConfigItemList query config item list under specific app.
func (s *Service) queryConfigItemsWithDeleted(details *types.ListConfigItemDetails,
	released []*table.ReleasedConfigItem, start, limit uint32, all bool) ([]*pbci.ConfigItem, uint32) {
	configItems, deleted := pbrci.PbConfigItemState(details.Details, released)
	count := details.Count + uint32(len(deleted))
	// if all, return configItems and all deleted
	if all {
		configItems = append(configItems, deleted...)
		return configItems, count
	}
	// 1. req.Start > details.Count
	if start > uint32(details.Count) {
		deletedStart := len(deleted)
		deletedEnd := len(deleted)
		if start-details.Count < uint32(len(deleted)) {
			deletedStart = int(start - details.Count)
		}
		if deletedStart+int(limit) < len(deleted) {
			deletedEnd = deletedStart + int(limit)
		}
		deleted = deleted[deletedStart:deletedEnd]
	} else {
		// 2 req.Start < details.Count
		// 2.1 req.Start+req.Limit > details.Count
		if start+limit < uint32(details.Count) {
			deleted = deleted[:0]
		} else {
			// 2.2 req.Start+req.Limit < details.Count
			deletedStart := 0
			deletedEnd := len(deleted)
			if start+limit-uint32(details.Count) < uint32(len(deleted)) {
				deletedEnd = int(start + limit - uint32(details.Count))
			}
			deleted = deleted[deletedStart:deletedEnd]
		}
	}
	configItems = append(configItems, deleted...)
	return configItems, count
}

// ListConfigItemCount list config items count.
func (s *Service) ListConfigItemCount(ctx context.Context, req *pbds.ListConfigItemCountReq) (
	*pbds.ListConfigItemCountResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	details, err := s.dao.ConfigItem().GetCount(grpcKit, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("list editing config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.ListConfigItemCountResp{
		Details: pbci.PbConfigItemCounts(details, req.AppId),
	}
	return resp, nil

}
