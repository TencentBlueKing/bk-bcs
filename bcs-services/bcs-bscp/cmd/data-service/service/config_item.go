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
	"fmt"
	"path"
	"reflect"
	"sort"
	"strings"
	"time"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbcommit "bscp.io/pkg/protocol/core/commit"
	pbci "bscp.io/pkg/protocol/core/config-item"
	pbrci "bscp.io/pkg/protocol/core/released-ci"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/search"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// CreateConfigItem create config item.
func (s *Service) CreateConfigItem(ctx context.Context, req *pbds.CreateConfigItemReq) (*pbds.CreateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validates unique key name+path both in table app_template_bindings and config_items
	// validate in table app_template_bindings
	if err := s.ValidateAppTemplateBindingUniqueKey(grpcKit, req.ConfigItemAttachment.BizId,
		req.ConfigItemAttachment.AppId, req.ConfigItemSpec.Name, req.ConfigItemSpec.Path); err != nil {
		return nil, err
	}
	// validate in table config_items
	if _, err := s.dao.ConfigItem().GetByUniqueKey(grpcKit, req.ConfigItemAttachment.BizId,
		req.ConfigItemAttachment.AppId, req.ConfigItemSpec.Name, req.ConfigItemSpec.Path); err == nil {
		return nil, fmt.Errorf("config item's same name %s and path %s already exists",
			req.ConfigItemSpec.Name, req.ConfigItemSpec.Path)
	}

	tx := s.dao.GenQuery().Begin()
	// 1. create config item.
	ci := &table.ConfigItem{
		Spec:       req.ConfigItemSpec.ConfigItemSpec(),
		Attachment: req.ConfigItemAttachment.ConfigItemAttachment(),
		Revision: &table.Revision{
			Creator: grpcKit.User,
			Reviser: grpcKit.User,
		},
	}
	ciID, err := s.dao.ConfigItem().CreateWithTx(grpcKit, tx, ci)
	if err != nil {
		logs.Errorf("create config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	// validate config items count.
	if e := s.dao.ConfigItem().ValidateAppCINumber(grpcKit, tx, req.ConfigItemAttachment.BizId,
		req.ConfigItemAttachment.AppId); e != nil {
		logs.Errorf("validate config items count failed, err: %v, rid: %s", e, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, e
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
			Creator: grpcKit.User,
		},
	}
	contentID, err := s.dao.Content().CreateWithTx(grpcKit, tx, content)
	if err != nil {
		logs.Errorf("create content failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
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
			Creator: grpcKit.User,
		},
	}
	_, err = s.dao.Commit().CreateWithTx(grpcKit, tx, commit)
	if err != nil {
		logs.Errorf("create commit failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, grpcKit.Rid)
		return nil, e
	}
	return &pbds.CreateResp{Id: ciID}, nil
}

// BatchUpsertConfigItems batch upsert config items.
func (s *Service) BatchUpsertConfigItems(ctx context.Context, req *pbds.BatchUpsertConfigItemsReq) (
	*pbds.BatchUpsertConfigItemsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	// 1. list all editing config items.
	cis, err := s.dao.ConfigItem().ListAllByAppID(grpcKit, req.AppId, req.BizId)
	if err != nil {
		logs.Errorf("list editing config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	editingCIMap := make(map[string]*table.ConfigItem)
	newCIMap := make(map[string]*pbds.BatchUpsertConfigItemsReq_ConfigItem)
	for _, ci := range cis {
		editingCIMap[path.Join(ci.Spec.Path, ci.Spec.Name)] = ci
	}
	for _, item := range req.Items {
		newCIMap[path.Join(item.ConfigItemSpec.Path, item.ConfigItemSpec.Name)] = item
	}
	// 2. check if config item is already exists in editing config items list.
	toCreate, toUpdateSpec, toUpdateContent, toDelete, err := s.checkConfigItems(grpcKit, req, editingCIMap, newCIMap)
	if err != nil {
		logs.Errorf("check and compare config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	now := time.Now().UTC()
	tx := s.dao.GenQuery().Begin()
	createId, e := s.doBatchCreateConfigItems(grpcKit, tx, toCreate, now, req.BizId, req.AppId)
	if e != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, e
	}
	updateId, e := s.doBatchUpdateConfigItemSpec(grpcKit, tx, toUpdateSpec, now,
		req.BizId, req.AppId, editingCIMap)
	if e != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, e
	}
	if e := s.doBatchUpdateConfigItemContent(grpcKit, tx, toUpdateContent, now,
		req.BizId, req.AppId, editingCIMap); e != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, e
	}
	if req.ReplaceAll {
		// if replace all,delete config items not in batch upsert request.
		if e := s.doBatchDeleteConfigItems(grpcKit, tx, toDelete, req.BizId, req.AppId); e != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			return nil, e
		}
	}
	// validate config items count.
	if e := s.dao.ConfigItem().ValidateAppCINumber(grpcKit, tx, req.BizId, req.AppId); e != nil {
		logs.Errorf("validate config items count failed, err: %v, rid: %s", e, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, e
	}
	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, grpcKit.Rid)
		return nil, e
	}
	// 返回创建和更新的ID
	mergedID := append(createId, updateId...) // nolint
	return &pbds.BatchUpsertConfigItemsResp{Ids: mergedID}, nil
}

func (s *Service) checkConfigItems(kt *kit.Kit, req *pbds.BatchUpsertConfigItemsReq,
	editingCIMap map[string]*table.ConfigItem, newCIMap map[string]*pbds.BatchUpsertConfigItemsReq_ConfigItem) (
	toCreate []*pbds.BatchUpsertConfigItemsReq_ConfigItem, toUpdateSpec []*pbds.BatchUpsertConfigItemsReq_ConfigItem,
	toUpdateContent []*pbds.BatchUpsertConfigItemsReq_ConfigItem, toDelete []uint32, err error) {
	// 1. list all config items' latest commit.
	ids := make([]uint32, 0, len(editingCIMap))
	for _, ci := range editingCIMap {
		ids = append(ids, ci.ID)
	}
	commits, err := s.dao.Commit().BatchListLatestCommits(kt, req.BizId, req.AppId, ids)
	commitMap := make(map[uint32]*table.Commit)
	for _, commit := range commits {
		commitMap[commit.Attachment.ConfigItemID] = commit
	}
	if err != nil {
		logs.Errorf("list latest commits failed, err: %v, rid: %s", err, kt.Rid)
		return nil, nil, nil, nil, err
	}
	for _, item := range req.Items {
		if editing, exists := editingCIMap[path.Join(item.ConfigItemSpec.Path, item.ConfigItemSpec.Name)]; exists {
			// 2.1 if config item already exists, need compare and update.
			specDiff, contentDiff, cErr := s.compareConfigItem(kt, item, editing, commitMap)
			if cErr != nil {
				logs.Errorf("compare config item failed, err: %v, rid: %s", err, kt.Rid)
				return nil, nil, nil, nil, cErr
			}
			if specDiff || contentDiff {
				toUpdateSpec = append(toUpdateSpec, item)
			}
			if contentDiff {
				toUpdateContent = append(toUpdateContent, item)
			}
		} else {
			// 2.2 if not exists, create new config item.
			toCreate = append(toCreate, item)
		}
	}
	// 3. delete config items not in batch upsert request.
	for _, ci := range editingCIMap {
		if newCIMap[path.Join(ci.Spec.Path, ci.Spec.Name)] == nil {
			// if config item not in batch upsert request, delete it.
			toDelete = append(toDelete, ci.ID)
		}
	}
	return //nolint
}

func (s *Service) doBatchCreateConfigItems(kt *kit.Kit, tx *gen.QueryTx,
	toCreate []*pbds.BatchUpsertConfigItemsReq_ConfigItem, now time.Time, bizID, appID uint32) ([]uint32, error) {
	createId := []uint32{}
	toCreateConfigItems := []*table.ConfigItem{}
	for _, item := range toCreate {
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
		toCreateConfigItems = append(toCreateConfigItems, ci)
	}
	if err := s.dao.ConfigItem().BatchCreateWithTx(kt, tx, bizID, appID, toCreateConfigItems); err != nil {
		logs.Errorf("batch create config items failed, err: %v, rid: %s", err, kt.Rid)
		return createId, err
	}
	toCreateContent := []*table.Content{}
	for i, item := range toCreate {
		toCreateContent = append(toCreateContent, &table.Content{
			Spec: item.ContentSpec.ContentSpec(),
			Attachment: &table.ContentAttachment{
				BizID:        bizID,
				AppID:        appID,
				ConfigItemID: toCreateConfigItems[i].ID,
			},
			Revision: &table.CreatedRevision{
				Creator: kt.User,
			},
		})
	}
	if err := s.dao.Content().BatchCreateWithTx(kt, tx, toCreateContent); err != nil {
		logs.Errorf("batch create config items failed, err: %v, rid: %s", err, kt.Rid)
		return createId, err
	}
	toCreateCommit := []*table.Commit{}
	for i := range toCreateContent {
		toCreateCommit = append(toCreateCommit, &table.Commit{
			Spec: &table.CommitSpec{
				ContentID: toCreateContent[i].ID,
				Content:   toCreateContent[i].Spec,
			},
			Attachment: &table.CommitAttachment{
				BizID:        bizID,
				AppID:        appID,
				ConfigItemID: toCreateConfigItems[i].ID,
			},
			Revision: &table.CreatedRevision{
				Creator: kt.User,
			},
		})
	}
	if err := s.dao.Commit().BatchCreateWithTx(kt, tx, toCreateCommit); err != nil {
		logs.Errorf("batch create commits failed, err: %v, rid: %s", err, kt.Rid)
		return createId, err
	}

	// 返回创建ID
	for _, item := range toCreateConfigItems {
		createId = append(createId, item.ID)
	}

	return createId, nil
}

func (s *Service) doBatchUpdateConfigItemSpec(kt *kit.Kit, tx *gen.QueryTx,
	toUpdate []*pbds.BatchUpsertConfigItemsReq_ConfigItem, now time.Time, _, _ uint32,
	ciMap map[string]*table.ConfigItem) ([]uint32, error) {
	updateId := []uint32{}
	configItems := []*table.ConfigItem{}
	for _, item := range toUpdate {
		ci := &table.ConfigItem{
			ID:         ciMap[path.Join(item.ConfigItemSpec.Path, item.ConfigItemSpec.Name)].ID,
			Spec:       item.ConfigItemSpec.ConfigItemSpec(),
			Attachment: item.ConfigItemAttachment.ConfigItemAttachment(),
			Revision: &table.Revision{
				Creator:   ciMap[path.Join(item.ConfigItemSpec.Path, item.ConfigItemSpec.Name)].Revision.Creator,
				Reviser:   kt.User,
				CreatedAt: ciMap[path.Join(item.ConfigItemSpec.Path, item.ConfigItemSpec.Name)].Revision.CreatedAt,
				UpdatedAt: now,
			},
		}
		configItems = append(configItems, ci)
	}
	if err := s.dao.ConfigItem().BatchUpdateWithTx(kt, tx, configItems); err != nil {
		logs.Errorf("batch update config items failed, err: %v, rid: %s", err, kt.Rid)
		return updateId, err
	}
	// 返回编辑ID
	for _, item := range configItems {
		updateId = append(updateId, item.ID)
	}

	return updateId, nil
}

func (s *Service) doBatchUpdateConfigItemContent(kt *kit.Kit, tx *gen.QueryTx,
	toUpdate []*pbds.BatchUpsertConfigItemsReq_ConfigItem, now time.Time,
	bizID, appID uint32, ciMap map[string]*table.ConfigItem) error {
	toCreateContents := []*table.Content{}
	for _, item := range toUpdate {
		content := &table.Content{
			Spec: item.ContentSpec.ContentSpec(),
			Attachment: &table.ContentAttachment{
				BizID:        bizID,
				AppID:        appID,
				ConfigItemID: ciMap[path.Join(item.ConfigItemSpec.Path, item.ConfigItemSpec.Name)].ID,
			},
			Revision: &table.CreatedRevision{
				Creator:   kt.User,
				CreatedAt: now,
			},
		}
		toCreateContents = append(toCreateContents, content)
	}
	if err := s.dao.Content().BatchCreateWithTx(kt, tx, toCreateContents); err != nil {
		logs.Errorf("batch create contents failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	toCreateCommits := []*table.Commit{}
	for i, item := range toUpdate {
		commit := &table.Commit{
			Spec: &table.CommitSpec{
				ContentID: toCreateContents[i].ID,
				Content:   item.ContentSpec.ContentSpec(),
			},
			Attachment: &table.CommitAttachment{
				BizID:        bizID,
				AppID:        appID,
				ConfigItemID: ciMap[path.Join(item.ConfigItemSpec.Path, item.ConfigItemSpec.Name)].ID,
			},
			Revision: &table.CreatedRevision{
				Creator:   kt.User,
				CreatedAt: now,
			},
		}
		toCreateCommits = append(toCreateCommits, commit)
	}
	if err := s.dao.Commit().BatchCreateWithTx(kt, tx, toCreateCommits); err != nil {
		logs.Errorf("batch create commits failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (s *Service) doBatchDeleteConfigItems(kt *kit.Kit, tx *gen.QueryTx, toDelete []uint32, bizID, appID uint32) error {
	if err := s.dao.ConfigItem().BatchDeleteWithTx(kt, tx, toDelete, bizID, appID); err != nil {
		logs.Errorf("batch create contents failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// nolint: unused
func (s *Service) createNewConfigItem(kt *kit.Kit, tx *gen.QueryTx, bizID, appID uint32,
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
			Creator: kt.User,
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
			Creator: kt.User,
		},
	}
	_, err = s.dao.Commit().CreateWithTx(kt, tx, commit)
	if err != nil {
		logs.Errorf("create commit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// compareConfigItem compare config item
// return specDiff, contentDiff, error
func (s *Service) compareConfigItem(_ *kit.Kit, new *pbds.BatchUpsertConfigItemsReq_ConfigItem,
	editing *table.ConfigItem, commitMap map[uint32]*table.Commit) (specDiff bool, contentDiff bool, err error) {
	// 1. compare config item spec.
	if !reflect.DeepEqual(new.ConfigItemSpec.ConfigItemSpec(), editing.Spec) {
		specDiff = true
	}
	// 2. compare content.
	// 2.1 get latest commit.
	commit, exists := commitMap[editing.ID]
	if !exists {
		// ! config item should have at least one commit.
		logs.Errorf("[SHOULD-NOT-HAPPEN] latest commit for config item %d not found", editing.ID)
		return false, false, fmt.Errorf("[SHOULD-NOT-HAPPEN] latest commit for config item %d not found", editing.ID)
	}
	// 2.2 compare content spec.
	if new.ContentSpec.Signature != commit.Spec.Content.Signature {
		contentDiff = true
	}
	return //nolint
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
			Reviser: grpcKit.User,
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
func (s *Service) ListConfigItems(ctx context.Context, req *pbds.ListConfigItemsReq) (*pbds.ListConfigItemsResp, // nolint
	error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate the page params
	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	// search all editing config items
	details, err := s.dao.ConfigItem().ListAllByAppID(grpcKit, req.AppId, req.BizId)
	if err != nil {
		logs.Errorf("list editing config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	configItems := make([]*pbci.ConfigItem, 0)
	// if WithStatus is true, the config items includes the deleted ones and file state, else  without these data
	if req.WithStatus {
		var fileReleased []*table.ReleasedConfigItem
		fileReleased, err = s.dao.ReleasedCI().GetReleasedLately(grpcKit, req.BizId, req.AppId)
		if err != nil {
			logs.Errorf("get released failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
		configItems = pbrci.PbConfigItemState(details, fileReleased)
	} else {
		for _, ci := range details {
			configItems = append(configItems, pbci.PbConfigItem(ci, ""))
		}
	}

	if err = s.setCommitSpecForCIs(grpcKit, configItems); err != nil {
		logs.Errorf("set commit spec for config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// search by logic
	if req.SearchValue != "" {
		var searcher search.Searcher
		searcher, err = search.NewSearcher(req.SearchFields, req.SearchValue, search.ConfigItem)
		if err != nil {
			return nil, err
		}
		fields := searcher.SearchFields()
		fieldsMap := make(map[string]bool)
		for _, f := range fields {
			fieldsMap[f] = true
		}
		cis := make([]*pbci.ConfigItem, 0)
		for _, ci := range configItems {
			if (fieldsMap["name"] && strings.Contains(ci.Spec.Name, req.SearchValue)) ||
				(fieldsMap["path"] && strings.Contains(ci.Spec.Path, req.SearchValue)) ||
				(fieldsMap["memo"] && strings.Contains(ci.Spec.Memo, req.SearchValue)) ||
				(fieldsMap["creator"] && strings.Contains(ci.Revision.Creator, req.SearchValue)) ||
				(fieldsMap["reviser"] && strings.Contains(ci.Revision.Reviser, req.SearchValue)) {
				cis = append(cis, ci)
			}
		}
		configItems = cis
	}

	// page by logic
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
	topId, _ := tools.StrToUint32Slice(req.Ids)
	sort.SliceStable(configItems, func(i, j int) bool {
		// 检测模板id是否在topId中
		iInTopID := tools.Contains(topId, configItems[i].Id)
		jInTopID := tools.Contains(topId, configItems[j].Id)
		// 两者都在则先path排再name排
		// 不管topID有没有都要先path排再name排
		if iInTopID && jInTopID || len(topId) == 0 {
			if configItems[i].GetSpec().GetPath() != configItems[j].GetSpec().GetPath() {
				return configItems[i].GetSpec().GetPath() < configItems[j].GetSpec().GetPath()
			}
			return configItems[i].GetSpec().GetName() < configItems[j].GetSpec().GetName()
		}
		if iInTopID {
			return true
		}
		if jInTopID {
			return false
		}
		return i < j
	})
	resp := &pbds.ListConfigItemsResp{
		Count:   uint32(len(configItems)),
		Details: configItems[start:end],
	}
	return resp, nil
}

// setCommitSpecForCIs set commit spec for config items
func (s *Service) setCommitSpecForCIs(kt *kit.Kit, cis []*pbci.ConfigItem) error {
	ids := make([]uint32, len(cis))
	for i, ci := range cis {
		ids[i] = ci.Id
	}

	commits, err := s.dao.Commit().BatchListLatestCommits(kt, kt.BizID, kt.AppID, ids)
	if err != nil {
		logs.Errorf("batch list latest commits failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	commitMap := make(map[uint32]*table.CommitSpec, len(commits))
	for _, c := range commits {
		commitMap[c.Attachment.ConfigItemID] = c.Spec
	}

	for _, ci := range cis {
		ci.CommitSpec = pbcommit.PbCommitSpec(commitMap[ci.Id])
	}

	return nil
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

// ListConfigItemByTuple 按照多个字段in查询
func (s *Service) ListConfigItemByTuple(ctx context.Context, req *pbds.ListConfigItemByTupleReq) (
	*pbds.ListConfigItemByTupleResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	data := [][]interface{}{}
	for _, item := range req.Items {
		data = append(data, []interface{}{item.BizId, item.AppId, item.Name, item.Path})
	}
	tuple, err := s.dao.ConfigItem().ListConfigItemByTuple(grpcKit, data)
	if err != nil {
		logs.Errorf("list config item by tuple failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	configItems := []*pbci.ConfigItem{}
	for _, item := range tuple {
		configItems = append(configItems, pbci.PbConfigItem(item, ""))
	}
	resp := &pbds.ListConfigItemByTupleResp{ConfigItems: configItems}
	return resp, nil
}
