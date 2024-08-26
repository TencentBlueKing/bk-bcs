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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbcommit "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/commit"
	pbci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbrci "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/released-ci"
	pbtset "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-set"
	pbtv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-variable"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateConfigItem create config item.
func (s *Service) CreateConfigItem(ctx context.Context, req *pbds.CreateConfigItemReq) (*pbds.CreateResp, error) { // nolint
	grpcKit := kit.FromGrpcContext(ctx)

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
		return nil, errf.Errorf(errf.InvalidRequest,
			i18n.T(grpcKit, "the config file %s under this service already exists and cannot be created again",
				path.Join(req.ConfigItemSpec.Path, req.ConfigItemSpec.Name)))
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
// nolint:funlen
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
	if err = tools.DetectFilePathConflicts(grpcKit, file2, file1); err != nil {
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

	// 3. 清空草稿区
	if err = s.clearDraftArea(grpcKit, tx, req.GetBizId(), req.GetAppId(), toDelete, req.GetReplaceAll()); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	// 4. 验证是否超出该服务限制
	binding, err := s.dao.AppTemplateBinding().GetAppTemplateBindingByAppIDWithTx(grpcKit, tx, req.BizId, req.AppId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(grpcKit, fmt.Sprintf("get template info for service binding failed, err: %s", err.Error())))
	}
	if err = s.verifyAppLimitHasBeenExceeded(grpcKit, tx, req.BizId, req.AppId, binding,
		req.GetBindings(), len(toCreate)); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}

	// 5. 处理变量以及新增或编辑
	vars, err := s.checkConfigItemVars(grpcKit, tx, req.BizId, req.AppId, req.GetVariables())
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	if vars != nil {
		if err = s.dao.AppTemplateVariable().UpsertWithTx(grpcKit, tx, vars); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			return nil, err
		}
	}

	// 6. 处理服务和模板绑定的关系以及编辑或新增绑定关系
	mergedTemplateSet, err := s.mergeCurrentWithHistoryTemplateSet(grpcKit, req.GetBizId(), req.GetAppId(),
		binding, req.GetBindings())
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, err
	}
	if mergedTemplateSet != nil {
		if err = s.dao.AppTemplateBinding().UpsertWithTx(grpcKit, tx, mergedTemplateSet); err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
			}
			return nil, err
		}
	}

	// 7. 添加非模板配置文件以及文件内容
	createIds, e := s.doBatchCreateConfigItems(grpcKit, tx, toCreate, now, req.BizId, req.AppId)
	if e != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, grpcKit.Rid)
		}
		return nil, e
	}
	updateIds, e := s.doBatchUpdateConfigItemSpec(grpcKit, tx, toUpdateSpec, now,
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

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, grpcKit.Rid)
		return nil, e
	}

	return &pbds.BatchUpsertConfigItemsResp{Ids: tools.MergeAndDeduplicate(createIds, updateIds)}, nil
}

// 清空草稿区
func (s *Service) clearDraftArea(kit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32,
	deleteConfigItemIDs []uint32, replaceAll bool) error {
	if !replaceAll {
		return nil
	}

	// 1. 删除模板绑定关系
	if err := s.dao.AppTemplateBinding().DeleteByAppIDWithTx(kit, tx, bizID, appID); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kit.Rid)
		}
		return errf.Errorf(errf.DBOpFailed,
			i18n.T(kit, "delete one app template binding instance by app id failed, err: %s", err))
	}

	// 2. 删除配置项数据
	if e := s.doBatchDeleteConfigItems(kit, tx, deleteConfigItemIDs, bizID, appID); e != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kit.Rid)
		}
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "batch create contents failed, err: %s", e))
	}

	// 3. 删除变量数据
	if e := s.dao.AppTemplateVariable().DeleteWithTx(kit, tx, bizID, appID); e != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kit.Rid)
		}
		return errf.Errorf(errf.DBOpFailed,
			i18n.T(kit, "delete one app template variable failed, err: %s", e))
	}

	return nil
}

// 验证是否已超过服务限制
// nolint:goconst
// currentBinding 表示当前未命名版本绑定的套餐数据
// bindings 待提交的套餐数据
// createConfigItemNum 待创建的配置项数量
func (s *Service) verifyAppLimitHasBeenExceeded(kit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32,
	currentBinding *table.AppTemplateBinding, bindings []*pbds.BatchUpsertConfigItemsReq_TemplateBinding,
	createConfigItemNum int) error {

	// 待提交的套餐数据
	betectedTemplateSetIDs := map[uint32]bool{}
	templateSetIDs := []uint32{}
	for _, v := range bindings {
		templateSetIDs = append(templateSetIDs, v.GetTemplateBinding().GetTemplateSetId())
		betectedTemplateSetIDs[v.GetTemplateBinding().GetTemplateSetId()] = true
	}

	// 从历史版本或者其他版本导入时关联的套餐
	templateSet, err := s.dao.TemplateSet().ListByIDs(kit, templateSetIDs)
	if err != nil {
		return err
	}
	historyTemplateSetNum := 0
	for _, v := range templateSet {
		historyTemplateSetNum += len(v.Spec.TemplateIDs)
	}

	// 当前未命名版本的关联套餐数，但不包含历史或其他历史版本套餐数
	currentTemplateSetNum := 0
	if currentBinding != nil {
		for _, binding := range currentBinding.Spec.Bindings {
			if !betectedTemplateSetIDs[binding.TemplateSetID] {
				currentTemplateSetNum += len(binding.TemplateRevisions)
			}
		}
	}

	// 获取当前未命名版本配置项数量
	configItemCount, err := s.dao.ConfigItem().GetConfigItemCountWithTx(kit, tx, bizID, appID)
	if err != nil {
		return err
	}

	appConfigCnt := getAppConfigCnt(bizID)

	app, err := s.dao.App().GetByID(kit, appID)
	if err != nil {
		return errf.Errorf(errf.AppNotExists, i18n.T(kit, "app %d not found", appID))
	}

	// 只有文件类型的才能绑定套餐
	if app.Spec.ConfigType != table.File {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "app %s is not file type", app.Spec.Name))
	}

	// 非模板配置数+当前套餐配置数+历史套餐配置数+待创建的配置项数
	if int(configItemCount)+currentTemplateSetNum+historyTemplateSetNum+createConfigItemNum > appConfigCnt {
		return errf.New(errf.InvalidParameter,
			i18n.T(kit, "the total number of app %s config items(including template and non-template)"+
				"exceeded the limit %d", app.Spec.Name, appConfigCnt))
	}

	return nil
}

// 检测变量
func (s *Service) checkConfigItemVars(kt *kit.Kit, tx *gen.QueryTx, bizID, appID uint32,
	variables []*pbtv.TemplateVariableSpec) (*table.AppTemplateVariable, error) {

	// 获取当前未命名版本中的变量
	currentVariable, err := s.dao.AppTemplateVariable().GetTemplateVariableWithTx(kt, tx, bizID, appID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if len(variables) == 0 {
		return currentVariable, nil
	}

	res := &table.AppTemplateVariable{}

	currentTemplateVars := make(map[string]*table.TemplateVariableSpec, 0)

	if currentVariable != nil {
		res.ID = currentVariable.ID
		res.Attachment = currentVariable.Attachment
		res.Revision = &table.Revision{
			Creator:   currentVariable.Revision.Creator,
			Reviser:   kt.User,
			CreatedAt: currentVariable.Revision.CreatedAt,
			UpdatedAt: time.Now().UTC(),
		}

		for _, v := range currentVariable.Spec.Variables {
			currentTemplateVars[v.Name] = v
		}
	} else {
		res.Attachment = &table.AppTemplateVariableAttachment{
			BizID: bizID,
			AppID: appID,
		}
		res.Revision = &table.Revision{
			Creator:   kt.User,
			Reviser:   kt.User,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
	}

	// 待提交的变量
	newVars := make(map[string]*table.TemplateVariableSpec, 0)
	for _, vars := range variables {
		newVars[vars.Name] = &table.TemplateVariableSpec{
			Name:       vars.Name,
			Type:       table.VariableType(vars.Type),
			DefaultVal: vars.DefaultVal,
			Memo:       vars.Memo,
		}
	}

	mergeTemplateVars := mergeTemplateVariables(currentTemplateVars, newVars)

	variableMap := make([]*table.TemplateVariableSpec, 0)
	for _, item := range mergeTemplateVars {
		variableMap = append(variableMap, item)
	}

	res.Spec = &table.AppTemplateVariableSpec{
		Variables: variableMap,
	}

	return res, nil
}

func mergeTemplateVariables(map1, map2 map[string]*table.TemplateVariableSpec) map[string]*table.TemplateVariableSpec {
	// 创建一个新map存储合并结果
	mergedMap := make(map[string]*table.TemplateVariableSpec)

	// 先将第一组数据拷贝到合并结果中
	for k, v := range map1 {
		mergedMap[k] = v
	}

	// 再将第二组数据拷贝到合并结果中，如果键相同则覆盖
	for k, v := range map2 {
		mergedMap[k] = v
	}

	return mergedMap
}

// 将当前模板与历史模板集合并
func (s *Service) mergeCurrentWithHistoryTemplateSet(kit *kit.Kit, bizID, appID uint32,
	currentBinding *table.AppTemplateBinding, bindings []*pbds.BatchUpsertConfigItemsReq_TemplateBinding) (
	*table.AppTemplateBinding, error) {

	if len(bindings) == 0 {
		return currentBinding, nil
	}

	appTemplateBinding := &table.AppTemplateBinding{
		Spec: &table.AppTemplateBindingSpec{},
	}

	// 当前未命名版本存在套餐关系时更新否则新增
	if currentBinding != nil {
		appTemplateBinding.ID = currentBinding.ID
		appTemplateBinding.Attachment = currentBinding.Attachment
		appTemplateBinding.Revision = &table.Revision{
			Creator:   currentBinding.Revision.Creator,
			Reviser:   kit.User,
			CreatedAt: currentBinding.Revision.CreatedAt,
			UpdatedAt: time.Now().UTC(),
		}
	} else {
		appTemplateBinding.Attachment = &table.AppTemplateBindingAttachment{
			BizID: bizID,
			AppID: appID,
		}
		appTemplateBinding.Revision = &table.Revision{
			Creator:   kit.User,
			Reviser:   kit.User,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
	}

	// 格式化待提交的数据把 pb 结构转换成 table 结构， 方便做对比
	templateSpaceIDs, templateSetIDs := []uint32{}, []uint32{}
	templateRevisions := make(table.TemplateBindings, 0)
	for _, v := range bindings {
		templateSpaceIDs = append(templateSpaceIDs, v.GetTemplateSpaceId())
		templateSetIDs = append(templateSetIDs, v.GetTemplateBinding().GetTemplateSetId())
		revision := make([]*table.TemplateRevisionBinding, 0)
		for _, binding := range v.GetTemplateBinding().GetTemplateRevisions() {
			revision = append(revision, &table.TemplateRevisionBinding{
				TemplateID:         binding.GetTemplateId(),
				TemplateRevisionID: binding.GetTemplateRevisionId(),
				IsLatest:           binding.GetIsLatest(),
			})
		}
		templateRevisions = append(templateRevisions, &table.TemplateBinding{
			TemplateSetID:     v.GetTemplateBinding().GetTemplateSetId(),
			TemplateRevisions: revision,
		})
	}

	// 未命名版本有关联套餐时处理当前套餐和待提交套餐的数据，否则只处理待提交数据
	if currentBinding != nil {
		templateSetIDsExist := make(map[uint32]bool)
		for _, v := range currentBinding.Spec.Bindings {
			templateSetIDsExist[v.TemplateSetID] = true
		}
		currentTemplateSetID := []uint32{}
		for _, v := range templateRevisions {
			if !templateSetIDsExist[v.TemplateSetID] {
				currentTemplateSetID = append(currentTemplateSetID, v.TemplateSetID)
			}
		}
		unBindingTemplateSets, err := s.getUnBindingTemplateSets(kit, currentTemplateSetID)
		if err != nil {
			return nil, err
		}
		currentBinding.Spec.Bindings = append(currentBinding.Spec.Bindings, unBindingTemplateSets...)
		appTemplateBinding.Spec = mergeTemplateSets(templateRevisions, currentBinding.Spec.Bindings)
		appTemplateBinding.Spec.TemplateSpaceIDs = tools.MergeAndDeduplicate(tools.RemoveDuplicates(templateSpaceIDs),
			tools.RemoveDuplicates(currentBinding.Spec.TemplateSpaceIDs))
	} else {
		// 获取待提交套餐下的当前套餐数据
		// 待提交版本以前关联的套餐数据可能被更新过
		unBindingTemplateSets, err := s.getUnBindingTemplateSets(kit, templateSetIDs)
		if err != nil {
			return nil, err
		}
		appTemplateBinding.Spec = mergeTemplateSets(templateRevisions, unBindingTemplateSets)
		appTemplateBinding.Spec.TemplateSpaceIDs = tools.RemoveDuplicates(templateSpaceIDs)
	}

	return appTemplateBinding, nil
}

// 获取未关联的模板套餐
func (s *Service) getUnBindingTemplateSets(kt *kit.Kit, templateSetID []uint32) (table.TemplateBindings, error) {
	// 查询未关联的套餐
	templateSet, err := s.dao.TemplateSet().ListByIDs(kt, templateSetID)
	if err != nil {
		return nil, err
	}

	unBindingTemplateSets := make(table.TemplateBindings, 0)
	for _, v := range templateSet {
		// 获取每个套餐下所有配置的最新版本
		tmplRevisions, err := s.dao.TemplateRevision().
			ListLatestRevisionsGroupByTemplateIds(kt, tools.RemoveDuplicates(v.Spec.TemplateIDs))
		if err != nil {
			return nil, err
		}
		revisions := make([]*table.TemplateRevisionBinding, 0)
		for _, revision := range tmplRevisions {
			revisions = append(revisions, &table.TemplateRevisionBinding{
				TemplateID:         revision.Attachment.TemplateID,
				TemplateRevisionID: revision.ID,
				IsLatest:           true,
			})
		}
		unBindingTemplateSets = append(unBindingTemplateSets, &table.TemplateBinding{
			TemplateSetID:     v.ID,
			TemplateRevisions: revisions,
		})
	}

	return unBindingTemplateSets, nil
}

// 把原有的模板空间、套餐、模板文件合并现有的数据
func mergeTemplateSets(a, b table.TemplateBindings) *table.AppTemplateBindingSpec {
	mergedMap := make(map[uint32]*table.TemplateBinding)

	// 把所有的 a 元素添加到 mergedMap
	for _, aItem := range a {
		mergedMap[aItem.TemplateSetID] = aItem
	}

	// 把元素 b 合并添加到 mergedMap
	for _, bItem := range b {
		if existing, exists := mergedMap[bItem.TemplateSetID]; exists {
			// 合并 template revisions
			revisionMap := make(map[uint32]*table.TemplateRevisionBinding)
			for _, rev := range existing.TemplateRevisions {
				revisionMap[rev.TemplateID] = rev
			}
			for _, rev := range bItem.TemplateRevisions {
				if existingRev, exists := revisionMap[rev.TemplateID]; exists {
					// 如果template_id存在，请检查is_latest
					if existingRev.IsLatest {
						if rev.IsLatest {
							// 如果两者都是最新的，请比较TemplateRevisionID
							if existingRev.TemplateRevisionID < rev.TemplateRevisionID {
								revisionMap[rev.TemplateID] = rev
							}
						}
					} else {
						// 如果existingRev不是最新版本，请保持existingRev
						revisionMap[rev.TemplateID] = existingRev
					}
				} else {
					revisionMap[rev.TemplateID] = rev
				}
			}
			// 将 map 转换回切片
			mergedRevisions := []*table.TemplateRevisionBinding{}
			for _, rev := range revisionMap {
				mergedRevisions = append(mergedRevisions, rev)
			}
			existing.TemplateRevisions = mergedRevisions
			mergedMap[bItem.TemplateSetID] = existing
		} else {
			mergedMap[bItem.TemplateSetID] = bItem
		}
	}

	// 将 map 转换回切片
	merged := []*table.TemplateBinding{}
	for _, item := range mergedMap {
		merged = append(merged, item)
	}

	templateSetIDs, templateIDs, templateRevisionIDs, latestTemplateIDs :=
		[]uint32{}, []uint32{}, []uint32{}, []uint32{}
	for _, v := range merged {
		templateSetIDs = append(templateSetIDs, v.TemplateSetID)
		for _, d := range v.TemplateRevisions {
			templateIDs = append(templateIDs, d.TemplateID)
			templateRevisionIDs = append(templateRevisionIDs, d.TemplateRevisionID)
			if d.IsLatest {
				latestTemplateIDs = append(latestTemplateIDs, d.TemplateID)
			}
		}
	}

	return &table.AppTemplateBindingSpec{
		TemplateSetIDs:      tools.RemoveDuplicates(templateSetIDs),
		TemplateIDs:         tools.RemoveDuplicates(templateIDs),
		TemplateRevisionIDs: tools.RemoveDuplicates(templateRevisionIDs),
		LatestTemplateIDs:   tools.RemoveDuplicates(latestTemplateIDs),
		Bindings:            merged,
	}
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
	createIds := []uint32{}
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
		return nil, err
	}
	toCreateContent := []*table.Content{}
	for i, item := range toCreate {
		createIds = append(createIds, toCreateConfigItems[i].ID)
		toCreateContent = append(toCreateContent, &table.Content{
			Spec: item.GetContentSpec().ContentSpec(),
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
		return nil, err
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
		return nil, err
	}

	return createIds, nil
}

func (s *Service) doBatchUpdateConfigItemSpec(kt *kit.Kit, tx *gen.QueryTx,
	toUpdate []*pbds.BatchUpsertConfigItemsReq_ConfigItem, now time.Time, _, _ uint32,
	ciMap map[string]*table.ConfigItem) ([]uint32, error) {
	updateIds := []uint32{}
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
		return nil, err
	}
	// 返回编辑ID
	for _, item := range configItems {
		updateIds = append(updateIds, item.ID)
	}

	return updateIds, nil
}

func (s *Service) doBatchUpdateConfigItemContent(kt *kit.Kit, tx *gen.QueryTx,
	toUpdate []*pbds.BatchUpsertConfigItemsReq_ConfigItem, now time.Time,
	bizID, appID uint32, ciMap map[string]*table.ConfigItem) error {
	toCreateContents := []*table.Content{}
	for _, item := range toUpdate {
		content := &table.Content{
			Spec: item.GetContentSpec().ContentSpec(),
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
// nolint:funlen,gocyclo
func (s *Service) ListConfigItems(ctx context.Context, req *pbds.ListConfigItemsReq) (*pbds.ListConfigItemsResp,
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

	existingPaths := []string{}
	for _, v := range configItems {
		if v.FileState != constant.FileStateDelete {
			existingPaths = append(existingPaths, path.Join(v.Spec.Path, v.Spec.Name))
		}
	}

	conflictNums, conflictPaths, err := s.compareTemplateConfConflicts(grpcKit, req.BizId, req.AppId, existingPaths)
	if err != nil {
		return nil, err
	}

	for _, v := range configItems {
		if v.FileState != constant.FileStateDelete {
			v.IsConflict = conflictPaths[path.Join(v.Spec.Path, v.Spec.Name)]
		}
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
		fieldsMap["combinedPathName"] = true
		cis := make([]*pbci.ConfigItem, 0)
		for _, ci := range configItems {
			combinedPathName := path.Join(ci.Spec.Path, ci.Spec.Name)
			if (fieldsMap["combinedPathName"] && strings.Contains(combinedPathName, req.SearchValue)) ||
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
	sort.SliceStable(configItems, func(i, j int) bool {
		iInTopID := tools.Contains(req.Ids, configItems[i].Id)
		jInTopID := tools.Contains(req.Ids, configItems[j].Id)
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
		Count:          uint32(len(configItems)),
		Details:        configItems[start:end],
		ConflictNumber: conflictNums,
	}
	return resp, nil
}

// 检测冲突，非模板配置之间对比、非模板配置对比套餐模板配置、空间套餐之间的对比
// 1. 先把非配置模板 path+name 添加到 existingPaths 中
// 2. 把所有关联的空间套餐配置都添加到 existingPaths 中
func (s *Service) compareTemplateConfConflicts(grpcKit *kit.Kit, bizID, appID uint32, existingPaths []string) (
	uint32, map[string]bool, error) {

	tmplRevisions, err := s.ListAppBoundTmplRevisions(grpcKit.RpcCtx(), &pbds.ListAppBoundTmplRevisionsReq{
		BizId:      bizID,
		AppId:      appID,
		All:        true,
		WithStatus: true,
	})
	if err != nil {
		return 0, nil, err
	}

	for _, revision := range tmplRevisions.GetDetails() {
		if revision.FileState != constant.FileStateDelete {
			existingPaths = append(existingPaths, path.Join(revision.Path, revision.Name))
		}
	}

	conflictNums, conflictPaths := checkExistingPathConflict(existingPaths)

	return conflictNums, conflictPaths, nil
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

	if err = tools.DetectFilePathConflicts(grpcKit, file1, file2); err != nil {
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

// CompareConfigItemConflicts compare config item version conflicts
func (s *Service) CompareConfigItemConflicts(ctx context.Context, req *pbds.CompareConfigItemConflictsReq) (
	*pbds.CompareConfigItemConflictsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	nonTemplateConfig, err := s.handleNonTemplateConfig(grpcKit, req.GetBizId(), req.GetAppId(),
		req.GetOtherAppId(), req.GetReleaseId())
	if err != nil {
		return nil, err
	}

	templateConfig, err := s.handleTemplateConfig(grpcKit, req.GetBizId(), req.GetAppId(),
		req.GetOtherAppId(), req.GetReleaseId())
	if err != nil {
		return nil, err
	}

	return &pbds.CompareConfigItemConflictsResp{
		NonTemplateConfigs: nonTemplateConfig,
		TemplateConfigs:    templateConfig,
	}, nil
}

// 处理非模板配置
func (s *Service) handleNonTemplateConfig(grpcKit *kit.Kit, bizID, appID, otherAppId, releaseId uint32) (
	[]*pbds.CompareConfigItemConflictsResp_NonTemplateConfig, error) {

	nonTemplateConfigs := make([]*pbds.CompareConfigItemConflictsResp_NonTemplateConfig, 0)

	// 获取未命名版本配置文件
	ci, err := s.dao.ConfigItem().ListAllByAppID(grpcKit, appID, bizID)
	if err != nil {
		logs.Errorf("list config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	conflicts := make(map[string]bool)
	for _, v := range ci {
		conflicts[path.Join(v.Spec.Path, v.Spec.Name)] = true
	}

	// 获取已发布版本的配置文件
	rci, count, err := s.dao.ReleasedCI().List(grpcKit, bizID, otherAppId, releaseId, nil, &types.BasePage{
		All: true,
	}, "")
	if err != nil {
		logs.Errorf("list released config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	if count == 0 {
		return nonTemplateConfigs, nil
	}

	configItems := make(map[string]bool)
	for _, v := range rci {
		configItems[path.Join(v.ConfigItemSpec.Path, v.ConfigItemSpec.Name)] = true
	}

	vars, err := s.getReleasedNonTemplateConfigVariables(grpcKit, bizID, otherAppId, releaseId)
	if err != nil {
		return nil, err
	}

	for _, v := range rci {
		nonTemplateConfigs = append(nonTemplateConfigs, &pbds.CompareConfigItemConflictsResp_NonTemplateConfig{
			Id: v.ConfigItemID,
			ConfigItemSpec: &pbci.ConfigItemSpec{
				Name:     v.ConfigItemSpec.Name,
				Path:     v.ConfigItemSpec.Path,
				FileType: string(v.ConfigItemSpec.FileType),
				FileMode: string(v.ConfigItemSpec.FileMode),
				Memo:     v.ConfigItemSpec.Memo,
				Permission: &pbci.FilePermission{
					User:      v.ConfigItemSpec.Permission.User,
					UserGroup: v.ConfigItemSpec.Permission.UserGroup,
					Privilege: v.ConfigItemSpec.Permission.Privilege,
				},
			},
			Variables: vars[path.Join(v.ConfigItemSpec.Path, v.ConfigItemSpec.Name)],
			IsExist:   conflicts[path.Join(v.ConfigItemSpec.Path, v.ConfigItemSpec.Name)],
			Signature: v.CommitSpec.Content.OriginSignature,
			ByteSize:  v.CommitSpec.Content.OriginByteSize,
			Md5:       v.CommitSpec.Content.Md5,
		})
	}

	return nonTemplateConfigs, nil
}

// 处理模板套餐配置
func (s *Service) handleTemplateConfig(grpcKit *kit.Kit, bizID, appID, otherAppId, releaseId uint32) (
	[]*pbds.CompareConfigItemConflictsResp_TemplateConfig, error) {
	templateConfigs := make([]*pbds.CompareConfigItemConflictsResp_TemplateConfig, 0)

	// 获取已发布版本的空间、套餐、配置文件
	rp, count, err := s.dao.ReleasedAppTemplate().List(grpcKit, bizID, otherAppId, releaseId, nil, &types.BasePage{
		All: true,
	}, "")
	if err != nil {
		logs.Errorf("list released app template revisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	if count == 0 {
		return templateConfigs, nil
	}

	noNamespacePackage, err := s.getConfigTemplateSet(grpcKit, bizID, appID)
	if err != nil {
		return nil, err
	}

	releaseTemplateSpaceIds, releaseTemplateSetIds, releaseTemplateIds := []uint32{}, []uint32{}, []uint32{}
	releaseTemplateSpaceIdsExist, releaseTemplateSetIdsExist := make(map[uint32]bool), make(map[uint32]bool)
	tmplSetMap := make(map[uint32][]*table.ReleasedAppTemplate)
	for _, v := range rp {
		tmplSetMap[v.Spec.TemplateSetID] = append(tmplSetMap[v.Spec.TemplateSetID], v)
		if !releaseTemplateSpaceIdsExist[v.Spec.TemplateSpaceID] {
			releaseTemplateSpaceIds = append(releaseTemplateSpaceIds, v.Spec.TemplateSpaceID)
			releaseTemplateSpaceIdsExist[v.Spec.TemplateSpaceID] = true
		}
		if !releaseTemplateSetIdsExist[v.Spec.TemplateSetID] {
			releaseTemplateSetIds = append(releaseTemplateSetIds, v.Spec.TemplateSetID)
			releaseTemplateSetIdsExist[v.Spec.TemplateSetID] = true
		}
		releaseTemplateIds = append(releaseTemplateIds, v.Spec.TemplateID)
	}

	templateSpaceExist, templateSetExist, currentSpaceSetTemplateExist, templateExist, templateSetTemplateExist, err :=
		s.getTemplateSpaceSetfile(grpcKit, releaseTemplateSpaceIds, releaseTemplateSetIds, releaseTemplateIds)
	if err != nil {
		return nil, err
	}

	nonExistentTemplateIds := make(map[uint32]bool)
	for _, v := range releaseTemplateIds {
		if !templateExist[v] {
			nonExistentTemplateIds[v] = true
		}
	}

	vars, err := s.getReleasedTemplateConfigVariables(grpcKit, bizID, otherAppId, releaseId, nonExistentTemplateIds)
	if err != nil {
		return nil, err
	}

	for id, revisions := range tmplSetMap {
		group := &pbds.CompareConfigItemConflictsResp_TemplateConfig{
			TemplateSpaceId:    revisions[0].Spec.TemplateSpaceID,
			TemplateSpaceName:  revisions[0].Spec.TemplateSpaceName,
			TemplateSetId:      id,
			TemplateSetName:    revisions[0].Spec.TemplateSetName,
			TemplateSpaceExist: templateSpaceExist[revisions[0].Spec.TemplateSpaceID],
			TemplateSetExist:   templateSetExist[id],
			IsExist:            noNamespacePackage[fmt.Sprintf("%d-%d", revisions[0].Spec.TemplateSpaceID, id)],
			TemplateSetIsEmpty: templateSetTemplateExist[id],
		}
		for _, r := range revisions {
			// 历史套餐模板文件被删除了
			if !templateExist[r.Spec.TemplateID] {
				continue
			}
			// 历史套餐模板不在现有套餐模板中, 被移走了
			if !currentSpaceSetTemplateExist[fmt.Sprintf("%d-%d-%d", r.Spec.TemplateSpaceID,
				r.Spec.TemplateSetID, r.Spec.TemplateID)] {
				continue
			}
			group.TemplateRevisions = append(group.TemplateRevisions,
				&pbds.CompareConfigItemConflictsResp_TemplateConfig_TemplateRevisionDetail{
					TemplateId:         r.Spec.TemplateID,
					TemplateRevisionId: r.Spec.TemplateRevisionID,
					IsLatest:           r.Spec.IsLatest,
					Variables:          vars[path.Join(r.Spec.Path, r.Spec.Name)],
				})
		}
		templateConfigs = append(templateConfigs, group)
	}

	return templateConfigs, nil
}

// 返回空间、套餐、模板配置数据
func (s *Service) getTemplateSpaceSetfile(grpcKit *kit.Kit, templateSpaceIds, templateSetIds, templateIds []uint32) (
	map[uint32]bool, map[uint32]bool, map[string]bool, map[uint32]bool, map[uint32]bool, error) {
	// 获取空间
	templateSpace, err := s.dao.TemplateSpace().ListByIDs(grpcKit, templateSpaceIds)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	templateSpaceExist := make(map[uint32]bool)
	for _, v := range templateSpace {
		templateSpaceExist[v.ID] = true
	}

	// 获取套餐
	templateSet, err := s.dao.TemplateSet().ListByIDs(grpcKit, templateSetIds)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	templateSetExist := make(map[uint32]bool)
	currentSpaceSetTemplateExist := make(map[string]bool)
	templateSetTemplateExist := make(map[uint32]bool)
	for _, v := range templateSet {
		templateSetTemplateExist[v.ID] = false
		if len(v.Spec.TemplateIDs) == 0 {
			templateSetTemplateExist[v.ID] = true
		}
		templateSetExist[v.ID] = true
		for _, tid := range v.Spec.TemplateIDs {
			currentSpaceSetTemplateExist[fmt.Sprintf("%d-%d-%d", v.Attachment.TemplateSpaceID, v.ID, tid)] = true
		}
	}

	// 获取模板
	template, err := s.dao.Template().ListByIDs(grpcKit, templateIds)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	templateExist := make(map[uint32]bool)
	for _, v := range template {
		templateExist[v.ID] = true
	}

	return templateSpaceExist, templateSetExist, currentSpaceSetTemplateExist, templateExist, templateSetTemplateExist, nil
}

// 获取未命名版本的模板套餐
func (s *Service) getConfigTemplateSet(grpcKit *kit.Kit, bizID, appID uint32) (
	map[string]bool, error) {

	noNamespacePackage := make(map[string]bool)

	tmplSetInfo, count, err := s.dao.AppTemplateBinding().List(grpcKit, bizID, appID, &types.BasePage{All: true})
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return noNamespacePackage, nil
	}

	tmplSets, err := s.dao.TemplateSet().ListByIDs(grpcKit, tmplSetInfo[0].Spec.TemplateSetIDs)
	if err != nil {
		logs.Errorf("list template sets failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	tmplSetMap := make(map[uint32]*table.TemplateSet)
	tmplSpaceIDs := make([]uint32, 0)
	for _, ts := range tmplSets {
		tmplSetMap[ts.ID] = ts
		tmplSpaceIDs = append(tmplSpaceIDs, ts.Attachment.TemplateSpaceID)
	}
	tmplSpaceIDs = tools.RemoveDuplicates(tmplSpaceIDs)

	// template space details
	tmplSpaces, err := s.dao.TemplateSpace().ListByIDs(grpcKit, tmplSpaceIDs)
	if err != nil {
		logs.Errorf("list template spaces failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	tmplSpaceMap := make(map[uint32]*table.TemplateSpace)
	for _, ts := range tmplSpaces {
		tmplSpaceMap[ts.ID] = ts
	}

	details := make([]*pbtset.TemplateSetBriefInfo, len(tmplSets))
	for idx, t := range tmplSets {
		details[idx] = &pbtset.TemplateSetBriefInfo{
			TemplateSpaceId:   t.Attachment.TemplateSpaceID,
			TemplateSpaceName: tmplSpaceMap[t.Attachment.TemplateSpaceID].Spec.Name,
			TemplateSetId:     t.ID,
			TemplateSetName:   tmplSetMap[t.ID].Spec.Name,
		}
	}

	for _, tmplSet := range details {
		noNamespacePackage[fmt.Sprintf("%d-%d", tmplSet.TemplateSpaceId, tmplSet.TemplateSetId)] = true
	}

	return noNamespacePackage, nil
}

// 获取已发布的模板配置变量
func (s *Service) getReleasedTemplateConfigVariables(grpcKit *kit.Kit, bizID, otherAppId, releaseId uint32,
	templateIds map[uint32]bool) (map[string][]*pbtv.TemplateVariableSpec, error) {
	varsMap := make(map[string][]*pbtv.TemplateVariableSpec, 0)

	releasedTmpls, count, err := s.dao.ReleasedAppTemplate().List(grpcKit, bizID, otherAppId, releaseId,
		nil, &types.BasePage{All: true}, "")
	if err != nil {
		logs.Errorf("list released app templates failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	if count == 0 {
		return varsMap, nil
	}

	tmplRevisions := getTmplRevisionsFromReleased(releasedTmpls)
	tmplRevisions = filterSizeForTmplRevisions(tmplRevisions)

	newTmplRevisions := make([]*table.TemplateRevision, 0)
	for _, v := range tmplRevisions {
		if templateIds[v.Attachment.TemplateID] {
			continue
		}
		newTmplRevisions = append(newTmplRevisions, v)
	}

	refs, err := s.getVariableReferences(grpcKit, newTmplRevisions, nil)
	if err != nil {
		logs.Errorf("get variable references failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resfMap := make(map[string][]string, 0)
	for _, v := range refs {
		for _, ref := range v.GetReferences() {
			filePath := path.Join(ref.Path, ref.Name)
			resfMap[filePath] = append(resfMap[filePath], v.GetVariableName())
		}
	}

	vars, err := s.dao.ReleasedAppTemplateVariable().ListVariables(grpcKit, bizID, otherAppId, releaseId)
	if err != nil {
		logs.Errorf("list released app template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	for _, v := range vars {
		for key, name := range resfMap {
			for _, n := range name {
				if v.Name == n {
					varsMap[key] = append(varsMap[key], &pbtv.TemplateVariableSpec{
						Name: n, Type: string(v.Type), DefaultVal: v.DefaultVal, Memo: v.Memo,
					})
				}
			}
		}
	}

	return varsMap, nil
}

// 获取已发布的非配置配置变量
func (s *Service) getReleasedNonTemplateConfigVariables(grpcKit *kit.Kit, bizID, otherAppId, releaseId uint32) (
	map[string][]*pbtv.TemplateVariableSpec, error) {
	varsMap := make(map[string][]*pbtv.TemplateVariableSpec, 0)

	releasedCIs, _, err := s.dao.ReleasedCI().List(grpcKit, bizID, otherAppId, releaseId, nil,
		&types.BasePage{All: true}, "")
	if err != nil {
		logs.Errorf("list released config items failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	cis := getPbConfigItemsFromReleased(releasedCIs)
	cis = filterSizeForConfigItems(cis)

	refs, err := s.getVariableReferences(grpcKit, nil, cis)
	if err != nil {
		logs.Errorf("get variable references failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resfMap := make(map[string][]string, 0)
	for _, v := range refs {
		for _, ref := range v.GetReferences() {
			filePath := path.Join(ref.Path, ref.Name)
			resfMap[filePath] = append(resfMap[filePath], v.GetVariableName())
		}
	}

	vars, err := s.dao.ReleasedAppTemplateVariable().ListVariables(grpcKit, bizID, otherAppId, releaseId)
	if err != nil {
		logs.Errorf("list released app template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	for _, v := range vars {
		for key, name := range resfMap {
			for _, n := range name {
				if v.Name == n {
					varsMap[key] = append(varsMap[key], &pbtv.TemplateVariableSpec{
						Name: n, Type: string(v.Type), DefaultVal: v.DefaultVal, Memo: v.Memo,
					})
				}
			}
		}
	}

	return varsMap, nil
}

// GetTemplateAndNonTemplateCICount 获取模板和非模板配置项数量
func (s *Service) GetTemplateAndNonTemplateCICount(ctx context.Context, req *pbds.GetTemplateAndNonTemplateCICountReq) (
	*pbds.GetTemplateAndNonTemplateCICountResp, error) {
	kt := kit.FromGrpcContext(ctx)

	count, err := s.dao.ConfigItem().GetConfigItemCount(kt, req.BizId, req.AppId)
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt, "obtain the number of configuration items"))
	}

	binding, err := s.dao.AppTemplateBinding().GetAppTemplateBindingByAppID(kt, req.BizId, req.AppId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kt,
			"get template binding relationships through business and service IDs failed, err: %s", err))
	}
	templateConfigItemCount := 0
	if binding != nil {
		templateConfigItemCount = len(binding.Spec.TemplateIDs)
	}

	return &pbds.GetTemplateAndNonTemplateCICountResp{
		ConfigItemCount:         uint64(count),
		TemplateConfigItemCount: uint64(templateConfigItemCount),
	}, nil
}

// RemoveAppBoundTmplSet 移除服务绑定的套餐
func (s *Service) RemoveAppBoundTmplSet(ctx context.Context, req *pbds.RemoveAppBoundTmplSetReq) (
	*pbbase.EmptyResp, error) {

	kit := kit.FromGrpcContext(ctx)

	// 1. 查询该服务下的引用套餐
	binding, err := s.dao.AppTemplateBinding().GetAppTemplateBindingByAppID(kit, req.BizId, req.AppId)
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kit, "get reference template set under this app failed, err: %s", err))
	}

	specBindings := make([]*table.TemplateBinding, 0, len(binding.Spec.Bindings))
	for _, spec := range binding.Spec.Bindings {
		revisions := make([]*table.TemplateRevisionBinding, 0)
		// 只需移除指定的模板套餐
		if spec.TemplateSetID == req.TemplateSetId {
			continue
		}
		revisions = append(revisions, spec.TemplateRevisions...)
		specBinding := &table.TemplateBinding{
			TemplateSetID:     spec.TemplateSetID,
			TemplateRevisions: revisions,
		}
		specBindings = append(specBindings, specBinding)
	}

	templateIDs, latestTemplateIDs, templateRevisionIDs, templateSetIDs :=
		[]uint32{}, []uint32{}, []uint32{}, []uint32{}
	for _, specBinding := range specBindings {
		for _, v := range specBinding.TemplateRevisions {
			templateIDs = append(templateIDs, v.TemplateID)
			if v.IsLatest {
				latestTemplateIDs = append(latestTemplateIDs, v.TemplateID)
			}
			templateRevisionIDs = append(templateRevisionIDs, v.TemplateRevisionID)
		}
		templateSetIDs = append(templateSetIDs, specBinding.TemplateSetID)
	}

	// 根据套餐id获取空间id
	templateSets, err := s.dao.TemplateSet().ListByIDs(kit, templateSetIDs)
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kit, "list template sets by template set ids failed, err: %s", err))
	}

	templateSpaceIDs := []uint32{}
	for _, v := range templateSets {
		templateSpaceIDs = append(templateSpaceIDs, v.Attachment.TemplateSpaceID)
	}

	appTemplateBinding := &table.AppTemplateBinding{
		ID: binding.ID,
		Spec: &table.AppTemplateBindingSpec{
			TemplateSpaceIDs:    templateSpaceIDs,
			TemplateSetIDs:      templateSetIDs,
			TemplateIDs:         tools.RemoveDuplicates(templateIDs),
			TemplateRevisionIDs: tools.RemoveDuplicates(templateRevisionIDs),
			LatestTemplateIDs:   tools.RemoveDuplicates(latestTemplateIDs),
			Bindings:            specBindings,
		},
		Attachment: binding.Attachment,
		Revision: &table.Revision{
			Creator:   binding.Revision.Creator,
			Reviser:   kit.User,
			CreatedAt: binding.Revision.CreatedAt,
			UpdatedAt: time.Now().UTC(),
		},
	}

	if err = s.dao.AppTemplateBinding().Update(kit, appTemplateBinding); err != nil {
		return nil, errf.Errorf(errf.DBOpFailed,
			i18n.T(kit, "remove the template set bound to the app failed, err: %s", err))
	}

	return &pbbase.EmptyResp{}, nil
}

// checkExistingPathConflict Check existing path collections for conflicts.
func checkExistingPathConflict(existing []string) (uint32, map[string]bool) {
	conflictPaths := make(map[string]bool, len(existing))
	var conflictNums uint32
	conflictMap := make(map[string]bool, 0)
	// 遍历每一个路径
	for i := 0; i < len(existing); i++ {
		// 检查当前路径与后续路径之间是否存在冲突
		for j := i + 1; j < len(existing); j++ {
			if strings.HasPrefix(existing[j]+"/", existing[i]+"/") || strings.HasPrefix(existing[i]+"/", existing[j]+"/") {
				// 相等也算冲突
				if len(existing[j]) == len(existing[i]) {
					conflictNums++
				} else if len(existing[j]) < len(existing[i]) {
					conflictMap[existing[j]] = true
				} else {
					conflictMap[existing[i]] = true
				}

				conflictPaths[existing[i]] = true
				conflictPaths[existing[j]] = true
			}
		}
	}

	return uint32(len(conflictMap)) + conflictNums, conflictPaths
}
