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

// Package logrule log rule
package logrule

import (
	"context"
	"fmt"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/rand"

	bklog "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// GetEntrypointsReq params
type GetEntrypointsReq struct {
	ProjectId    string   `json:"projectId" in:"path=projectId" validate:"required"`
	ClusterId    string   `json:"clusterId" in:"path=clusterId" validate:"required"`
	ContainerIDs []string `json:"container_ids" form:"container_ids"`
}

// Entrypoint entrypoint
type Entrypoint struct {
	STDLogURL     string `json:"std_log_url"`
	FileLogURL    string `json:"file_log_url"`
	STDBKBaseURL  string `json:"std_bk_base_url"`  // 跳转到数据平台地址
	FileBKBaseURL string `json:"file_bk_base_url"` // 跳转到数据平台地址
}

// GetEntrypoints 获取容器日志查询入口
// @Summary 获取容器日志查询入口
// @Tags    LogCollectors
// @Produce json
// @Success 200 {object} map[string]Entrypoint
// @Router  /log_collector/entrypoints [post]
func GetEntrypoints(c context.Context, req *GetEntrypointsReq) (*map[string]Entrypoint, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	return getContainerQueryLogLinks(req.ContainerIDs, rctx.ProjectCode, rctx.ClusterId), nil
}

// ListLogCollectors 获取日志采集规则列表
// @Summary 获取日志采集规则列表
// @Tags    LogCollectors
// @Produce json
// @Success 200 {array} GetLogRuleResp
// @Router  /log_collector/rules [get]
func ListLogCollectors(c context.Context, req *ListLogCollectorsReq) (*[]*GetLogRuleResp, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	listInDB, err := getClusterLogRules(c, rctx.ProjectId, req.ClusterId)
	if err != nil {
		return nil, err
	}

	// 从 bk-log 获取规则数据
	lcs, err := bklog.ListLogCollectors(c, req.ClusterId, GetSpaceID(rctx.ProjectCode))
	if err != nil {
		return nil, err
	}

	result := make([]*GetLogRuleResp, 0)
	for _, rule := range listInDB {
		lrr := &GetLogRuleResp{}
		lrr.loadFromEntity(rule, lcs)
		result = append(result, lrr)
	}

	// 从日志平台获取非 bcs 创建的日志规则
	for _, rule := range lcs {
		for _, v := range result {
			if v.RuleName == rule.CollectorConfigNameEN {
				continue
			}
		}
		if !rule.FromBKLog {
			continue
		}
		lrr := &GetLogRuleResp{}
		lrr.loadFromBkLog(rule, rctx.ProjectCode)
		result = append(result, lrr)
	}

	sort.Sort(GetLogRuleRespSortByName(result))

	return &result, nil
}

// GetLogRule 获取日志采集规则详情
// @Summary 获取日志采集规则详情
// @Tags    LogCollectors
// @Produce json
// @Success 200 object GetLogRuleResp
// @Router  /log_collector/rules/:id [get]
func GetLogRule(c context.Context, req *GetLogRuleReq) (*GetLogRuleResp, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	id := req.ID
	store := storage.GlobalStorage

	// 从 bk-log 获取规则数据
	lcs, err := bklog.ListLogCollectors(c, req.ClusterId, GetSpaceID(rctx.ProjectCode))
	if err != nil {
		return nil, err
	}

	if isBKLogID(id) {
		ruleName := getBKLogName(id)
		for _, rule := range lcs {
			if rule.CollectorConfigNameEN == ruleName {
				result := &GetLogRuleResp{}
				result.loadFromBkLog(rule, rctx.ProjectCode)
				return result, nil
			}
		}
		return nil, errors.Errorf("not found %s", id)
	}

	// 从数据库获取规则数据
	lcInDB, err := store.GetLogRule(c, id)
	if err != nil {
		return nil, err
	}

	result := &GetLogRuleResp{}
	result.loadFromEntity(lcInDB, lcs)
	return result, nil
}

// CreateLogRule 创建日志采集规则
// @Summary 创建日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules [post]
func CreateLogRule(c context.Context, req *CreateLogRuleReq) (*string, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	req.RuleName = fmt.Sprintf("%s_%s", req.Name, rand.String(5))

	// check rule is exist
	store := storage.GlobalStorage
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectID: rctx.ProjectId,
		entity.FieldKeyClusterID: req.ClusterId,
		entity.FieldKeyName:      req.Name,
	})
	count, _, err := store.ListLogRules(c, cond, &utils.ListOption{})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.Errorf("%s is exist", req.Name)
	}

	id, err := store.CreateLogRule(c, req.toEntity(rctx))
	if err != nil {
		return nil, err
	}

	// 创建 bklog 规则耗时比较长，异步调用
	go createBKLog(utils.GetTenantIDFromContext(c), req.toBKLog(rctx))
	return &id, nil
}

// UpdateLogRule 更新日志采集规则
// @Summary 更新日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules/:id [put]
func UpdateLogRule(c context.Context, req *UpdateLogRuleReq) (*string, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	id := req.ID
	if isBcsLogConfigID(id) || isBKLogID(id) {
		return nil, fmt.Errorf("id is invalid")
	}

	// check rule is exist
	store := storage.GlobalStorage
	rule, err := store.GetLogRule(c, id)
	if err != nil {
		return nil, err
	}

	if rule.RuleID == 0 {
		return nil, errors.Errorf("invalid rule id")
	}

	err = store.UpdateLogRule(c, id, req.toEntity(rctx.Username, rctx.ProjectCode, rule.Name))
	if err != nil {
		return nil, err
	}

	// 更新 bklog 规则耗时比较长，异步调用
	go updateBKLog(utils.GetTenantIDFromContext(c), rule.ID.Hex(), rule.RuleID, req.toBKLog(rctx, rule.RuleName))
	return &id, nil
}

// DeleteLogRule 删除日志采集规则
// @Summary 删除日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules/:id [delete]
func DeleteLogRule(c context.Context, req *GetLogRuleReq) (*any, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	id := req.ID

	// 从数据库获取规则数据
	store := storage.GlobalStorage
	lc, err := store.GetLogRule(c, id)
	if err != nil {
		return nil, err
	}

	if lc.ProjectCode != rctx.ProjectCode || lc.ClusterID != req.ClusterId {
		return nil, errors.New("invalid id")
	}

	if lc.RuleID != 0 {
		// 删除 bklog 数据
		err = bklog.DeleteLogCollectors(c, lc.RuleID)
		if err != nil {
			// 有可能日志平台侧已经删除，这时删除会失败，可以忽略报错，继续删除数据库记录
			blog.Errorf("delete bklog rule error, %s", err.Error())
		}
	}

	err = store.DeleteLogRule(c, id)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// RetryLogRule 重试日志采集规则
// @Summary 重试日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules/:id/retry [post]
func RetryLogRule(c context.Context, req *GetLogRuleReq) (*any, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	id := req.ID
	if isBcsLogConfigID(id) {
		return nil, nil
	}

	// 从数据库获取规则数据
	store := storage.GlobalStorage
	rule, err := store.GetLogRule(c, id)
	if err != nil {
		return nil, err
	}

	data := entity.M{
		entity.FieldKeyStatus:  entity.PendingStatus,
		entity.FieldKeyMessage: "",
		entity.FieldKeyUpdator: rctx.Username,
	}
	// 重新创建
	if rule.RuleID == 0 {
		// 创建 bklog 规则耗时比较长，异步调用
		ruleName := fmt.Sprintf("%s_%s", rule.Name, rand.String(5))
		data.Update(entity.FieldKeyRuleName, ruleName)
		// 更新状态
		err = store.UpdateLogRule(c, id, data)
		if err != nil {
			return nil, err
		}
		matchLabels, matchExpressions := bklog.MergeOutLabels(rule.Rule.LogRuleContainer.LabelSelector.MatchLabels,
			rule.Rule.LogRuleContainer.LabelSelector.MatchExpressions)
		rule.Rule.LogRuleContainer.LabelSelector = bklog.LabelSelector{
			MatchLabels: matchLabels, MatchExpressions: matchExpressions}
		go createBKLog(utils.GetTenantIDFromContext(c), &bklog.CreateBCSCollectorReq{
			SpaceUID:              GetSpaceID(rctx.ProjectCode),
			ProjectID:             rctx.ProjectId,
			CollectorConfigName:   rule.DisplayName,
			CollectorConfigNameEN: ruleName,
			Description:           rule.Description,
			BCSClusterID:          req.ClusterId,
			AddPodLabel:           rule.Rule.AddPodLabel,
			ExtraLabels:           rule.Rule.ExtraLabels,
			LogRuleContainer:      []bklog.LogRuleContainer{rule.Rule.LogRuleContainer},
			Username:              rctx.Username,
		})
		return nil, nil
	}

	// 重试 bklog collector
	err = bklog.RetryLogCollectors(c, rule.RuleID, rctx.Username)
	if err != nil {
		data.Update(entity.FieldKeyStatus, entity.FailedStatus)
		data.Update(entity.FieldKeyMessage, err.Error())
		store.UpdateLogRule(c, id, data) // nolint
		return nil, err
	}

	// 更新状态
	err = store.UpdateLogRule(c, id, data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// EnableLogRule 启用日志采集规则
// @Summary 启用日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules/:id/enable [post]
func EnableLogRule(c context.Context, req *GetLogRuleReq) (*any, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	id := req.ID
	if isBcsLogConfigID(id) {
		return nil, nil
	}

	// 从数据库获取规则数据
	store := storage.GlobalStorage
	rule, err := store.GetLogRule(c, id)
	if err != nil {
		return nil, err
	}

	if rule.RuleID == 0 {
		return nil, errors.Errorf("invalid rule id, please recreate log rule")
	}

	data := entity.M{
		entity.FieldKeyStatus:  entity.SuccessStatus,
		entity.FieldKeyMessage: "",
		entity.FieldKeyUpdator: rctx.Username,
	}

	// 开启 bklog collector
	err = bklog.StartLogCollectors(c, rule.RuleID, rctx.Username)
	if err != nil {
		data.Update(entity.FieldKeyStatus, entity.FailedStatus)
		data.Update(entity.FieldKeyMessage, err.Error())
		store.UpdateLogRule(c, id, data) // nolint
		return nil, err
	}

	// 更新状态
	err = store.UpdateLogRule(c, id, data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// DisableLogRule 停用日志采集规则
// @Summary 停用日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules/:id/disable [post]
func DisableLogRule(c context.Context, req *GetLogRuleReq) (*any, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	id := req.ID
	if isBcsLogConfigID(id) {
		return nil, nil
	}

	// 从数据库获取规则数据
	store := storage.GlobalStorage
	rule, err := store.GetLogRule(c, id)
	if err != nil {
		return nil, err
	}

	if rule.RuleID == 0 {
		return nil, errors.Errorf("invalid rule id, please recreate log rule")
	}

	data := entity.M{
		entity.FieldKeyStatus:  entity.SuccessStatus,
		entity.FieldKeyMessage: "",
		entity.FieldKeyUpdator: rctx.Username,
	}

	// 停止 bklog collector
	err = bklog.StopLogCollectors(c, rule.RuleID, rctx.Username)
	if err != nil {
		data.Update(entity.FieldKeyStatus, entity.FailedStatus)
		data.Update(entity.FieldKeyMessage, err.Error())
		store.UpdateLogRule(c, id, data) // nolint
		return nil, err
	}

	// 更新状态
	err = store.UpdateLogRule(c, id, data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// GetStorageClusters 获取 ES 存储集群
// @Summary 获取 ES 存储集群
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/storages/cluster_groups [get]
func GetStorageClusters(
	c context.Context, req *bklog.GetStorageClustersReq) (*[]bklog.GetStorageClustersRespData, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}
	data, err := bklog.GetStorageClusters(c, GetSpaceID(rctx.ProjectCode))
	if err != nil {
		return nil, err
	}
	selectCluster, err := bklog.GetBcsCollectorStorage(c, GetSpaceID(rctx.ProjectCode), req.ClusterId)
	if err != nil {
		return nil, err
	}
	for i := range data {
		if data[i].StorageClusterID == selectCluster {
			data[i].IsSelected = true
		}
	}

	return &data, nil
}

// SwitchStorageReq 切换 ES 存储集群请求
type SwitchStorageReq struct {
	ProjectId        string `json:"projectId" in:"path=projectId" validate:"required"`
	ClusterId        string `json:"clusterId" in:"path=clusterId" validate:"required"`
	StorageClusterID int    `json:"storage_cluster_id"`
}

// SwitchStorage 切换 ES 存储集群
// @Summary 切换 ES 存储集群
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/storages/switch_storage [post]
func SwitchStorage(c context.Context, req *SwitchStorageReq) (*any, error) {
	rctx, err := rest.GetRestContext(c)
	if err != nil {
		return nil, err
	}

	if err := bklog.SwitchStorage(c,
		GetSpaceID(rctx.ProjectCode), rctx.ClusterId, req.StorageClusterID, rctx.Username); err != nil {
		return nil, err
	}

	return nil, nil
}
