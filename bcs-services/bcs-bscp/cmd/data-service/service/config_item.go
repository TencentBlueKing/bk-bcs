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
	"fmt"
	"path"
	"reflect"
	"time"

	"bscp.io/pkg/dal/sharding"
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

	now := time.Now()
	// 1. list all editing config items.
	listOpts := &types.ListConfigItemsOption{
		BizID: req.BizId,
		AppID: req.AppId,
		Filter: &filter.Expression{
			Op:    filter.And,
			Rules: []filter.RuleFactory{},
		},
		Page: &types.BasePage{
			// set start to 0, limit to 0, to get all editing config items.
			Start: 0,
			Limit: 0,
		},
	}
	cis, err := s.dao.ConfigItem().List(grpcKit, listOpts)
	if err != nil {
		logs.Errorf("list editing config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	editingCIMap := make(map[string]*table.ConfigItem)
	newCIMap := make(map[string]*pbds.BatchUpsertConfigItemsReq_ConfigItem)
	for _, ci := range cis.Details {
		editingCIMap[path.Join(ci.Spec.Path, ci.Spec.Name)] = ci
	}
	for _, item := range req.Items {
		newCIMap[path.Join(item.ConfigItemSpec.Path, item.ConfigItemSpec.Name)] = item
	}
	tx, err := s.dao.BeginTx(grpcKit, req.BizId)
	if err != nil {
		logs.Errorf("create config item, begin transaction failed, err: %v", err)
		return nil, err
	}
	// 2. check if config item is already exists in editing config items list.
	for _, item := range req.Items {
		if editing, exists := editingCIMap[path.Join(item.ConfigItemSpec.Path, item.ConfigItemSpec.Name)]; exists {
			// 2.1 if config item already exists, compare and update.
			if err := s.compareAndUpdateConfigItem(grpcKit, tx, req.BizId, req.AppId, now, item, editing); err != nil {
				tx.Rollback(grpcKit)
				return nil, err
			}
		} else {
			// 2.2 if not exists, create new config item.
			if err := s.createNewConfigItem(grpcKit, tx, req.BizId, req.AppId, now, item); err != nil {
				tx.Rollback(grpcKit)
				return nil, err
			}
		}
	}
	// 3. delete config items not in batch upsert request.
	for _, ci := range cis.Details {
		if newCIMap[path.Join(ci.Spec.Path, ci.Spec.Name)] == nil {
			// if config item not in batch upsert request, delete it.
			err := s.dao.ConfigItem().DeleteWithTx(grpcKit, tx, &table.ConfigItem{ID: ci.ID,
				Attachment: &table.ConfigItemAttachment{BizID: req.BizId, AppID: req.AppId}})
			if err != nil {
				logs.Errorf("delete config item %d failed, err: %v, rid: %s", ci.ID, err, grpcKit.Rid)
				tx.Rollback(grpcKit)
				return nil, err
			}
		}
	}
	tx.Commit(grpcKit)

	return new(pbbase.EmptyResp), nil
}

func (s *Service) createNewConfigItem(kt *kit.Kit, tx *sharding.Tx, bizID, appID uint32,
	now time.Time, item *pbds.BatchUpsertConfigItemsReq_ConfigItem) error {
	// 1. create config item.
	ci := &table.ConfigItem{
		Spec:       item.ConfigItemSpec.ConfigItemSpec(),
		Attachment: item.ConfigItemAttachment.ConfigItemAttachment(),
		Revision: &table.Revision{
			Creator:   kt.User,
			Reviser:   kt.User,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	ciID, err := s.dao.ConfigItem().CreateWithTx(kt, tx, ci)
	if err != nil {
		logs.Errorf("create config item failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// 2. create content.
	content := &table.Content{
		Spec: item.ContentSpec.ContentSpec(),
		Attachment: &table.ContentAttachment{
			BizID:        bizID,
			AppID:        appID,
			ConfigItemID: ciID,
		},
		Revision: &table.CreatedRevision{
			Creator:   kt.User,
			CreatedAt: now,
		},
	}
	contentID, err := s.dao.Content().CreateWithTx(kt, tx, content)
	if err != nil {
		logs.Errorf("create content failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// 3. create commit.
	commit := &table.Commit{
		Spec: &table.CommitSpec{
			ContentID: contentID,
			Content:   content.Spec,
		},
		Attachment: &table.CommitAttachment{
			BizID:        bizID,
			AppID:        appID,
			ConfigItemID: ciID,
		},
		Revision: &table.CreatedRevision{
			Creator:   kt.User,
			CreatedAt: now,
		},
	}
	_, err = s.dao.Commit().CreateWithTx(kt, tx, commit)
	if err != nil {
		logs.Errorf("create commit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (s *Service) compareAndUpdateConfigItem(kt *kit.Kit, tx *sharding.Tx, bizID, appID uint32,
	now time.Time, new *pbds.BatchUpsertConfigItemsReq_ConfigItem, editing *table.ConfigItem) error {
	// compare spec and content.
	specDiff, contentDiff, err := s.compareConfigItem(kt, new, editing)
	if err != nil {
		logs.Errorf("compare config item failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	// if any(spec/content) diff, update config item.
	if specDiff || contentDiff {
		ci := &table.ConfigItem{
			ID:         editing.ID,
			Spec:       new.ConfigItemSpec.ConfigItemSpec(),
			Attachment: new.ConfigItemAttachment.ConfigItemAttachment(),
			Revision: &table.Revision{
				Reviser:   kt.User,
				UpdatedAt: now,
			},
		}
		if err := s.dao.ConfigItem().Update(kt, ci); err != nil {
			logs.Errorf("update config item failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	// if content diff, create new content and commit.
	if contentDiff {
		// 1. create content.
		content := &table.Content{
			Spec: new.ContentSpec.ContentSpec(),
			Attachment: &table.ContentAttachment{
				BizID:        bizID,
				AppID:        appID,
				ConfigItemID: editing.ID,
			},
			Revision: &table.CreatedRevision{
				Creator:   kt.User,
				CreatedAt: now,
			},
		}
		contentID, err := s.dao.Content().CreateWithTx(kt, tx, content)
		if err != nil {
			logs.Errorf("create content failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
		// 2. create commit.
		commit := &table.Commit{
			Spec: &table.CommitSpec{
				ContentID: contentID,
				Content:   content.Spec,
			},
			Attachment: &table.CommitAttachment{
				BizID:        bizID,
				AppID:        appID,
				ConfigItemID: editing.ID,
			},
			Revision: &table.CreatedRevision{
				Creator:   kt.User,
				CreatedAt: now,
			},
		}
		_, err = s.dao.Commit().CreateWithTx(kt, tx, commit)
		if err != nil {
			logs.Errorf("create commit failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	// no diff, do nothing.
	return nil
}

// compareConfigItem compare config item
// return specDiff, contentDiff, error
func (s *Service) compareConfigItem(kt *kit.Kit, new *pbds.BatchUpsertConfigItemsReq_ConfigItem,
	editing *table.ConfigItem) (specDiff bool, contentDiff bool, err error) {
	// 1. compare config item spec.
	if !reflect.DeepEqual(new.ConfigItemSpec.ConfigItemSpec(), editing.Spec) {
		specDiff = true
	}
	// 2. compare content.
	// 2.1 get latest commit.
	commit, err := s.queryCILatestCommit(kt, editing.Attachment.BizID, editing.Attachment.AppID, editing.ID)
	if err != nil {
		return false, false, fmt.Errorf("query config item %d latest commit failed, err: %v", editing.ID, err)
	}
	// 2.2 compare content spec.
	if new.ContentSpec.Signature != commit.Spec.Content.Signature {
		contentDiff = true
	}
	return
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
		// list all editing config items
		query := &types.ListConfigItemsOption{
			BizID: req.BizId,
			AppID: req.AppId,
			Filter: &filter.Expression{
				Op:    filter.Or,
				Rules: []filter.RuleFactory{},
			},
			Page: &types.BasePage{
				// set start to 0, limit to 0, to get all editing config items.
				Start: 0,
				Limit: 0,
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
		configItems := pbrci.PbConfigItemState(details.Details, fileReleased)
		var start, end uint32 = 0, uint32(len(configItems))
		if !req.All {
			if req.Start < uint32(len(configItems)) {
				start = req.Start
			}
			if req.Start+req.Limit < uint32(len(configItems)) {
				end = req.Start + req.Limit
			} else {
				end = uint32(len(configItems))
			}
		}
		resp := &pbds.ListConfigItemsResp{
			Count:   uint32(len(configItems)),
			Details: configItems[start:end],
		}
		return resp, nil
	}
	// list released config items
	query := &types.ListReleasedCIsOption{
		BizID:     req.BizId,
		ReleaseID: req.ReleaseId,
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
