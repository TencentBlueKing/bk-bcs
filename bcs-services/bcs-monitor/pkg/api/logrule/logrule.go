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
 *
 */

package logrule

import (
	"context"
	"fmt"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog/v2"

	bklog "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// GetEntrypointsReq params
type GetEntrypointsReq struct {
	ContainerIDs []string `json:"container_ids" form:"container_ids"`
}

// Entrypoint entrypoint
type Entrypoint struct {
	STDLogURL  string `json:"std_log_url"`
	FileLogURL string `json:"file_log_url"`
}

// GetEntrypoints 获取容器日志查询入口
// @Summary 获取容器日志查询入口
// @Tags    LogCollectors
// @Produce json
// @Success 200 {object} map[string]Entrypoint
// @Router  /log_collector/entrypoints [post]
func GetEntrypoints(c *rest.Context) (interface{}, error) {
	req := &GetEntrypointsReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		klog.Errorf("GetEntrypoints bind req json error, %s", err.Error())
		return nil, err
	}

	// 从数据库获取规则数据
	store := storage.GlobalStorage
	stdIndexSetID, fileIndexSetID, err := store.GetIndexSetID(c.Request.Context(), c.ProjectId, c.ClusterId)
	if err != nil {
		return nil, err
	}

	return getContainerQueryLogLinks(req.ContainerIDs, stdIndexSetID, fileIndexSetID, c.ProjectCode), nil
}

// ListLogCollectors 获取日志采集规则列表
// @Summary 获取日志采集规则列表
// @Tags    LogCollectors
// @Produce json
// @Success 200 {array} GetLogRuleResp
// @Router  /log_collector/rules [get]
func ListLogCollectors(c *rest.Context) (interface{}, error) {
	// 从数据库获取规则数据
	store := storage.GlobalStorage
	listInDB, err := getClusterLogRules(context.Background(), c.ProjectId, c.ClusterId)
	if err != nil {
		return nil, err
	}

	// 从 bk-log 获取规则数据
	lcs, err := bklog.ListLogCollectors(c.Request.Context(), c.ClusterId, GetSpaceID(c.ProjectCode))
	if err != nil {
		return nil, err
	}

	result := make([]*GetLogRuleResp, 0)
	for _, rule := range listInDB {
		lrr := &GetLogRuleResp{}
		lrr.loadFromEntity(rule, lcs)
		result = append(result, lrr)
	}

	// 从 bcslogconfigs 获取数据
	bcsLogConfigs, err := k8sclient.ListBcsLogConfig(c.Request.Context(), c.ClusterId)
	if err != nil {
		return nil, err
	}

	// 添加 bcslogconfigs 至列表
	if len(bcsLogConfigs) > 0 {
		// get old default data id
		logIndex, err := store.GetOldIndexSetID(c.Request.Context(), c.ProjectId)
		if err != nil {
			return nil, err
		}
		for i := range bcsLogConfigs {
			bcsLogID := toBcsLogConfigID(bcsLogConfigs[i].Namespace, bcsLogConfigs[i].Name)
			newRuleID := getFromRuleIDFromLogRules(bcsLogID, listInDB)
			lrr := &GetLogRuleResp{}
			lrr.loadFromBcsLogConfig(&bcsLogConfigs[i], logIndex, newRuleID)
			result = append(result, lrr)
		}
	}

	sort.Sort(GetLogRuleRespSortByUpdateTime(result))
	return result, nil
}

// GetLogRule 获取日志采集规则详情
// @Summary 获取日志采集规则详情
// @Tags    LogCollectors
// @Produce json
// @Success 200 object GetLogRuleResp
// @Router  /log_collector/rules/:id [get]
func GetLogRule(c *rest.Context) (interface{}, error) {
	id := c.Param("id")
	store := storage.GlobalStorage

	// check bcslogconfig
	if isBcsLogConfigID(id) {
		ns, name := getBcsLogConfigNamespaces(id)
		blc, err := k8sclient.GetBcsLogConfig(c.Request.Context(), c.ClusterId, ns, name)
		if err != nil {
			return nil, err
		}
		logIndex, err := store.GetOldIndexSetID(c.Request.Context(), c.ProjectId)
		if err != nil {
			klog.Warningf("not found %s index set id", c.ProjectId)
		}
		listInDB, err := getClusterLogRules(c.Request.Context(), c.ProjectId, c.ClusterId)
		if err != nil {
			return nil, err
		}
		newRuleID := getFromRuleIDFromLogRules(id, listInDB)
		result := &GetLogRuleResp{}
		result.loadFromBcsLogConfig(blc, logIndex, newRuleID)
		return result, nil
	}

	// 从数据库获取规则数据
	lcInDB, err := store.GetLogRule(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}

	// 从 bk-log 获取规则数据
	lcs, err := bklog.ListLogCollectors(c.Request.Context(), c.ClusterId, GetSpaceID(c.ProjectCode))
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
func CreateLogRule(c *rest.Context) (interface{}, error) {
	req := &CreateLogRuleReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		klog.Errorf("CreateLogCollector bind req json error, %s", err.Error())
		return nil, err
	}
	req.RuleName = fmt.Sprintf("%s_%s", req.Name, rand.String(5))

	// check rule is exist
	store := storage.GlobalStorage
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectID: c.ProjectId,
		entity.FieldKeyClusterID: c.ClusterId,
		entity.FieldKeyName:      req.Name,
	})
	count, _, err := store.ListLogRules(c.Request.Context(), cond, &utils.ListOption{})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.Errorf("%s is exist", req.Name)
	}

	err = store.CreateLogRule(c.Request.Context(), req.toEntity(c))
	if err != nil {
		return nil, err
	}

	// 创建 bklog 规则耗时比较长，异步调用
	go createBKLog(req.toBKLog(c))
	return nil, nil
}

// UpdateLogRule 更新日志采集规则
// @Summary 更新日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules/:id [put]
func UpdateLogRule(c *rest.Context) (interface{}, error) {
	id := c.Param("id")
	if isBcsLogConfigID(id) {
		return nil, fmt.Errorf("can't update bcslogconfig")
	}
	req := &UpdateLogRuleReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		klog.Errorf("UpdateLogCollector bind req json error, %s", err.Error())
		return nil, err
	}

	// check rule is exist
	store := storage.GlobalStorage
	rule, err := store.GetLogRule(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}

	if rule.RuleID == 0 {
		return nil, errors.Errorf("invalid rule id")
	}

	err = store.UpdateLogRule(c.Request.Context(), id, req.toEntity(c.Username, c.ProjectCode))
	if err != nil {
		return nil, err
	}

	// 更新 bklog 规则耗时比较长，异步调用
	go updateBKLog(rule.ID.Hex(), rule.RuleID, req.toBKLog(c, rule.RuleName))
	return nil, nil
}

// DeleteLogRule 删除日志采集规则
// @Summary 删除日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules/:id [delete]
func DeleteLogRule(c *rest.Context) (interface{}, error) {
	id := c.Param("id")
	if isBcsLogConfigID(id) {
		ns, name := getBcsLogConfigNamespaces(id)
		err := k8sclient.DeleteBcsLogConfig(c.Request.Context(), c.ClusterId, ns, name)
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	// 从数据库获取规则数据
	store := storage.GlobalStorage
	lc, err := store.GetLogRule(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}

	if lc.ProjectCode != c.ProjectCode || lc.ClusterID != c.ClusterId {
		return nil, errors.New("invalid id")
	}

	if lc.RuleID != 0 {
		// 删除 bklog 数据
		err = bklog.DeleteLogCollectors(c.Request.Context(), lc.RuleID)
		if err != nil {
			return nil, err
		}
	}

	err = store.DeleteLogRule(c.Request.Context(), id)
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
func RetryLogRule(c *rest.Context) (interface{}, error) {
	id := c.Param("id")
	if isBcsLogConfigID(id) {
		return nil, nil
	}

	// 从数据库获取规则数据
	store := storage.GlobalStorage
	rule, err := store.GetLogRule(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}

	if rule.RuleID == 0 {
		return nil, errors.Errorf("invalid rule id, please recreate log rule")
	}

	data := entity.M{
		entity.FieldKeyStatus:  entity.SuccessStatus,
		entity.FieldKeyMessage: "",
		entity.FieldKeyUpdator: c.Username,
	}

	// 重试 bklog collector
	err = bklog.RetryLogCollectors(c.Request.Context(), rule.RuleID)
	if err != nil {
		data.Update(entity.FieldKeyStatus, entity.FailedStatus)
		data.Update(entity.FieldKeyMessage, err.Error())
		return nil, err
	}

	// 更新状态
	err = store.UpdateLogRule(c.Context, id, data)
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
func EnableLogRule(c *rest.Context) (interface{}, error) {
	id := c.Param("id")
	if isBcsLogConfigID(id) {
		return nil, nil
	}

	// 从数据库获取规则数据
	store := storage.GlobalStorage
	rule, err := store.GetLogRule(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}

	if rule.RuleID == 0 {
		return nil, errors.Errorf("invalid rule id, please recreate log rule")
	}

	data := entity.M{
		entity.FieldKeyStatus:  entity.SuccessStatus,
		entity.FieldKeyMessage: "",
		entity.FieldKeyUpdator: c.Username,
	}

	// 开启 bklog collector
	err = bklog.StartLogCollectors(c.Request.Context(), rule.RuleID)
	if err != nil {
		data.Update(entity.FieldKeyStatus, entity.FailedStatus)
		data.Update(entity.FieldKeyMessage, err.Error())
		return nil, err
	}

	// 更新状态
	err = store.UpdateLogRule(c.Context, id, data)
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
func DisableLogRule(c *rest.Context) (interface{}, error) {
	id := c.Param("id")
	if isBcsLogConfigID(id) {
		return nil, nil
	}

	// 从数据库获取规则数据
	store := storage.GlobalStorage
	rule, err := store.GetLogRule(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}

	if rule.RuleID == 0 {
		return nil, errors.Errorf("invalid rule id, please recreate log rule")
	}

	data := entity.M{
		entity.FieldKeyStatus:  entity.SuccessStatus,
		entity.FieldKeyMessage: "",
		entity.FieldKeyUpdator: c.Username,
	}

	// 停止 bklog collector
	err = bklog.StopLogCollectors(c.Request.Context(), rule.RuleID)
	if err != nil {
		data.Update(entity.FieldKeyStatus, entity.FailedStatus)
		data.Update(entity.FieldKeyMessage, err.Error())
		return nil, err
	}

	// 更新状态
	err = store.UpdateLogRule(c.Context, id, data)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
