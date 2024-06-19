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
	"fmt"
	"path"
	"reflect"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbcommit "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/commit"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbrci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/released-ci"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateConfigItem create config item.
func (s *Service) CreateConfigItem(ctx context.Context, req *pbds.CreateConfigItemReq) (*pbds.CreateResp, error) { // nolint
	grpcKit := kit.FromGrpcContext(ctx)

	// validates unique key name+path both in table app_template_bindings and config_items
	// validate in table app_template_bindings
	if err := s.ValidateAppTemplateBindingUniqueKey(grpcKit, req.ConfigItemAttachment.BizId,
		req.ConfigItemAttachment.AppId, req.ConfigItemSpec.Name, req.ConfigItemSpec.Path); err != nil {
		return nil, err
	}

	// get all configuration files under this service
	items, err := s.dao.ConfigItem().ListAllByAppID(grpcKit,
		req.ConfigItemAttachment.AppId, req.ConfigItemAttachment.BizId)
	if err != nil {
		return nil, err
	}
	existingPaths := []string{}
	for _, v := range items {
		existingPaths = append(existingPaths, path.Join(v.Spec.Path, v.Spec.Name))
	}

	// validate in table config_items
	if tools.CheckPathConflict(path.Join(req.ConfigItemSpec.Path, req.ConfigItemSpec.Name), existingPaths) {
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

	file1, file2 := make([]tools.CIUniqueKey, 0), make([]tools.CIUniqueKey, 0)
	editingCIMap := make(map[string]*table.ConfigItem)
	newCIMap := make(map[string]*pbds.BatchUpsertConfigItemsReq_ConfigItem)
	for _, ci := range cis {
		editingCIMap[path.Join(ci.Spec.Path, ci.Spec.Name)] = ci
		file1 = append(file1, tools.CIUniqueKey{Name: ci.Spec.Name, Path: ci.Spec.Path})
	}
	for _, item := range req.Items {
		newCIMap[path.Join(item.ConfigItemSpec.Path, item.ConfigItemSpec.Name)] = item
		file2 = append(file2, tools.CIUniqueKey{
			Name: item.GetConfigItemSpec().GetName(), Path: item.GetConfigItemSpec().GetPath(),
		})
	}
	if err = tools.DetectFilePathConflicts(file2, file1); err != nil {
		return nil, err
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
	vars, err := s.checkConfigItemVars(grpcKit, req, req.ReplaceAll)
	if err != nil {
		return nil, err
	}
	if err = s.dao.AppTemplateVariable().UpsertWithTx(grpcKit, tx, vars); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
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

// 检测变量
func (s *Service) checkConfigItemVars(kt *kit.Kit, req *pbds.BatchUpsertConfigItemsReq, replaceAll bool) (
	*table.AppTemplateVariable, error) {
	res := new(table.AppTemplateVariable)
	newVars := make(map[string]*table.TemplateVariableSpec, 0)
	for _, v := range req.GetItems() {
		for _, vars := range v.GetVariables() {
			newVars[vars.Name] = &table.TemplateVariableSpec{
				Name:       vars.Name,
				Type:       table.VariableType(vars.Type),
				DefaultVal: vars.DefaultVal,
				Memo:       vars.Memo,
			}
		}
	}

	variableMap := make([]*table.TemplateVariableSpec, 0)

	for _, item := range newVars {
		variableMap = append(variableMap, item)
	}

	// 获取原有的变量
	variable, err := s.dao.AppTemplateVariable().Get(kt, req.BizId, req.AppId)
	if err != nil {
		return nil, err
	}

	res.Attachment = &table.AppTemplateVariableAttachment{
		BizID: req.BizId,
		AppID: req.AppId,
	}
	if variable != nil {
		res.ID = variable.ID
		res.Revision = &table.Revision{
			Reviser:   kt.User,
			UpdatedAt: time.Now().UTC(),
		}
	} else {
		res.Revision = &table.Revision{
			Reviser:   kt.User,
			Creator:   kt.User,
			CreatedAt: time.Now().UTC(),
		}
	}

	if replaceAll || variable == nil {
		res.Spec = &table.AppTemplateVariableSpec{
			Variables: variableMap,
		}
		return res, nil
	}

	// 覆盖值等信息
	resultMap := make(map[string]*table.TemplateVariableSpec, 0)
	for _, v := range variable.Spec.Variables {
		resultMap[v.Name] = v
	}
	for _, v := range newVars {
		resultMap[v.Name] = v
	}

	for _, item := range resultMap {
		variableMap = append(variableMap, item)
	}

	res.Spec = &table.AppTemplateVariableSpec{
		Variables: variableMap,
	}

	return res, nil
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

		var commits []*table.Commit
		commits, err = s.dao.Commit().ListAppLatestCommits(grpcKit, req.BizId, req.AppId)
		if err != nil {
			logs.Errorf("get commit, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
		configItems = pbrci.PbConfigItemState(details, fileReleased, commits, req.Status)
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

	// 如果有topID则按照topID排最前面
	topId, _ := tools.StrToUint32Slice(req.Ids)
	sort.SliceStable(configItems, func(i, j int) bool {
		iInTopID := tools.Contains(topId, configItems[i].Id)
		jInTopID := tools.Contains(topId, configItems[j].Id)
		if iInTopID && jInTopID {
			return i < j
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

	appIDMap := make(map[uint32]uint32, len(req.AppId))
	for _, id := range req.AppId {
		appIDMap[id] = id
	}

	count, err := s.dao.ConfigItem().GetCount(grpcKit, req.BizId, req.AppId)
	if err != nil {
		logs.Errorf("list config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	var appIds = []uint32{}
	for _, detail := range count {
		delete(appIDMap, detail.AppId)
	}
	if len(appIDMap) > 0 {
		for _, appID := range appIDMap {
			appIds = append(appIds, appID)
		}
		kvDetails, err := s.dao.Kv().GetCount(grpcKit, req.BizId, appIds)
		if err != nil {
			logs.Errorf("list kv failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
		count = append(count, kvDetails...)
	}

	resp := &pbds.ListConfigItemCountResp{
		Details: pbci.PbConfigItemCounts(count, req.AppId),
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

// UnDeleteConfigItem 配置项未命名版本恢复
func (s *Service) UnDeleteConfigItem(ctx context.Context, req *pbds.UnDeleteConfigItemReq) (*pbbase.EmptyResp, error) { // nolint
	grpcKit := kit.FromGrpcContext(ctx)

	// 判断是否需要恢复
	configItem, err := s.dao.ConfigItem().Get(grpcKit, req.GetId(), req.Attachment.BizId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if configItem != nil && configItem.ID != 0 {
		return nil, errors.New("The data has not been deleted")
	}

	// 获取该服务最新发布的 release_id
	release, err := s.dao.Release().GetReleaseLately(grpcKit, req.Attachment.BizId, req.Attachment.AppId)
	if err != nil {
		return nil, err
	}

	// 通过最新发布 release_id + config_item_id 获取需要恢复的数据
	releaseCi, err := s.dao.ReleasedCI().Get(grpcKit, req.Attachment.BizId,
		release.Attachment.AppID, release.ID, req.GetId())
	if err != nil {
		return nil, err
	}

	// 检测文件冲突
	// /a 和 /a/1.txt这类的冲突
	file1 := []tools.CIUniqueKey{{
		Name: releaseCi.ConfigItemSpec.Name,
		Path: releaseCi.ConfigItemSpec.Path,
	}}

	configs, err := s.dao.ConfigItem().ListAllByAppID(grpcKit, req.Attachment.AppId, req.Attachment.BizId)
	if err != nil {
		return nil, err
	}
	file2 := []tools.CIUniqueKey{}
	for _, v := range configs {
		file2 = append(file2, tools.CIUniqueKey{
			Name: v.Spec.Name,
			Path: v.Spec.Path,
		})
	}

	if err = tools.DetectFilePathConflicts(file1, file2); err != nil {
		return nil, err
	}

	ci, err := s.dao.ConfigItem().GetByUniqueKey(grpcKit, req.Attachment.BizId, req.Attachment.AppId,
		releaseCi.ConfigItemSpec.Name, releaseCi.ConfigItemSpec.Path)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	commitID := []uint32{}
	contentID := []uint32{}
	tx := s.dao.GenQuery().Begin()
	// 判断是不是新增的数据
	if ci != nil && ci.ID != 0 {
		rci, errCi := s.dao.ReleasedCI().Get(grpcKit, req.Attachment.BizId,
			release.Attachment.AppID, release.ID, ci.ID)
		if errCi != nil && !errors.Is(errCi, gorm.ErrRecordNotFound) {
			return nil, errCi
		}
		if rci != nil && rci.ID != 0 {
			return nil, errors.New("recovery failed. A file with the same path exists and is not in a new state")
		}

		err = s.dao.ConfigItem().DeleteWithTx(grpcKit, tx, ci)
		if err != nil {
			logs.Errorf("recover config item failed, err: %v, rid: %s", err, grpcKit.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			return nil, err
		}
	}

	// 恢复到最新发布的版本，删除修改的数据
	// 获取大于最新发布版本的记录
	rc, err := s.dao.Commit().ListCommitsByGtID(grpcKit, releaseCi.CommitID, req.Attachment.BizId,
		req.Attachment.AppId, req.Id)
	if err != nil {
		return nil, err
	}
	for _, v := range rc {
		commitID = append(commitID, v.ID)
		contentID = append(contentID, v.Spec.ContentID)
	}

	if err = s.dao.Commit().BatchDeleteWithTx(grpcKit, tx, commitID); err != nil {
		logs.Errorf("undo commit failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	if err = s.dao.Content().BatchDeleteWithTx(grpcKit, tx, contentID); err != nil {
		logs.Errorf("undo content failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	data := &table.ConfigItem{
		ID:         releaseCi.ConfigItemID,
		Spec:       releaseCi.ConfigItemSpec,
		Attachment: releaseCi.Attachment,
		Revision:   releaseCi.Revision,
	}
	if err = s.dao.ConfigItem().RecoverConfigItem(grpcKit, tx, data); err != nil {
		logs.Errorf("recover config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, grpcKit.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}

// UndoConfigItem 撤消配置项
func (s *Service) UndoConfigItem(ctx context.Context, req *pbds.UndoConfigItemReq) (*pbbase.EmptyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// 判断是否存在
	_, err := s.dao.ConfigItem().Get(grpcKit, req.GetId(), req.Attachment.BizId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("data does not exist")
		}
	}

	// 获取该服务最新发布的 release_id
	release, err := s.dao.Release().GetReleaseLately(grpcKit, req.Attachment.BizId, req.Attachment.AppId)
	if err != nil {
		return nil, err
	}

	// 通过最新发布 release_id + config_item_id 获取需要恢复的数据
	releaseCi, err := s.dao.ReleasedCI().Get(grpcKit, req.Attachment.BizId,
		release.Attachment.AppID, release.ID, req.GetId())
	if err != nil {
		return nil, err
	}

	rc, err := s.dao.Commit().ListCommitsByGtID(grpcKit, releaseCi.CommitID, req.Attachment.BizId,
		req.Attachment.AppId, req.Id)
	if err != nil {
		return nil, err
	}

	commitID := []uint32{}
	contentID := []uint32{}
	for _, v := range rc {
		commitID = append(commitID, v.ID)
		contentID = append(contentID, v.Spec.ContentID)
	}

	tx := s.dao.GenQuery().Begin()
	if err = s.dao.Commit().BatchDeleteWithTx(grpcKit, tx, commitID); err != nil {
		logs.Errorf("undo commit failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	if err = s.dao.Content().BatchDeleteWithTx(grpcKit, tx, contentID); err != nil {
		logs.Errorf("undo content failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	data := &table.ConfigItem{
		ID:         releaseCi.ConfigItemID,
		Spec:       releaseCi.ConfigItemSpec,
		Attachment: releaseCi.Attachment,
		Revision:   releaseCi.Revision,
	}

	if err = s.dao.ConfigItem().UpdateWithTx(grpcKit, tx, data); err != nil {
		logs.Errorf("recover config item failed, err: %v, rid: %s", err, grpcKit.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, grpcKit.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}
