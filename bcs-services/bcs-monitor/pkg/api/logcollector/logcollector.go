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

package logcollector

import (
	"fmt"

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

// CreateLogCollectorReq create log collector request
type CreateLogCollectorReq struct {
	Name           string                    `json:"name" form:"name" binding:"required" validate:"max=32,min=5,regexp=^[A-Za-z0-9_]+$"`
	RuleName       string                    `json:"-" form:"-"`
	Namespace      string                    `json:"namespace" form:"namespace" binding:"required" validate:"max=63,min=1"`
	Description    string                    `json:"description" form:"description"`
	AddPodLabel    bool                      `json:"add_pod_label" form:"add_pod_label"`
	ExtraLabels    map[string]string         `json:"extra_labels" form:"extra_labels"`
	ConfigSelected entity.ConfigSelected     `json:"config_selected" form:"config_selected" binding:"required"`
	Config         entity.LogCollectorConfig `json:"config" form:"config" binding:"required"`
}

// UpdateLogCollectorReq update log collector request
type UpdateLogCollectorReq struct {
	Description    string                    `json:"description" form:"description"`
	AddPodLabel    bool                      `json:"add_pod_label"`
	ExtraLabels    map[string]string         `json:"extra_labels"`
	ConfigSelected entity.ConfigSelected     `json:"config_selected" form:"config_selected" binding:"required"`
	Config         entity.LogCollectorConfig `json:"config" form:"config" binding:"required"`
}

// GetEntrypointsReq params
type GetEntrypointsReq struct {
	ContainerIDs []string `json:"container_ids" form:"container_ids"`
}

// Entrypoint entrypoint
type Entrypoint struct {
	STDLogURL  string `json:"std_log_url"`
	FileLogURL string `json:"file_log_url"`
}

// GetEntrypoints 获取日志采集日志查询入口
// @Summary 获取日志采集规则列表
// @Tags    LogCollectors
// @Produce json
// @Success 200 {object} Entrypoint
// @Router  /log_collector/entrypoints [get]
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
// @Success 200 {array} entity.LogCollector
// @Router  /log_collector/rules [get]
func ListLogCollectors(c *rest.Context) (interface{}, error) {
	// 从数据库获取规则数据
	store := storage.GlobalStorage
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectID: c.ProjectId,
		entity.FieldKeyClusterID: c.ClusterId,
	})
	listOption := &utils.ListOption{Sort: map[string]interface{}{"updatedAt": 1}}
	_, listInDB, err := store.ListLogCollectors(c.Request.Context(), cond, listOption)
	if err != nil {
		return nil, err
	}
	list := make([]LogCollector, 0)
	for _, v := range listInDB {
		list = append(list, logEntityToLogCollector(*v))
	}

	// 从 bk-log 获取规则数据
	lcs, err := bklog.ListLogCollectors(c.Request.Context(), c.ClusterId, fmt.Sprintf(spaceUIDFormat, c.ProjectCode))
	if err != nil {
		return nil, err
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
		for _, v := range bcsLogConfigs {
			list = append(list, bcsLogToLogCollector(v, c.ClusterId, logIndex))
		}
	}

	// 合并数据
	result := mergeLogCollector(list, lcs)
	return result, nil
}

// GetCollector 获取日志采集规则详情
// @Summary 获取日志采集规则详情
// @Tags    LogCollectors
// @Produce json
// @Success 200 object entity.LogCollector
// @Router  /log_collector/rules/:id [get]
func GetCollector(c *rest.Context) (interface{}, error) {
	id := c.Param("id")
	store := storage.GlobalStorage
	if isBcsLogConfigID(id) {
		ns, name := getBcsLogConfigNamespaces(id)
		blc, err := k8sclient.GetBcsLogConfig(c.Request.Context(), c.ClusterId, ns, name)
		if err != nil {
			return nil, err
		}
		logIndex, err := store.GetOldIndexSetID(c.Request.Context(), c.ProjectId)
		if err != nil {
			return nil, err
		}
		return bcsLogToLogCollector(*blc, c.ClusterId, logIndex), nil
	}

	// 从数据库获取规则数据
	lcInDB, err := store.GetLogCollector(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}

	// 从 bk-log 获取规则数据
	lcs, err := bklog.ListLogCollectors(c.Request.Context(), c.ClusterId, fmt.Sprintf(spaceUIDFormat, c.ProjectCode))
	if err != nil {
		return nil, err
	}

	lc := logEntityToLogCollector(*lcInDB)
	// 合并数据
	lc.Deleted = true
	for _, v := range lcs {
		if v.CollectorConfigName == lc.RuleName {
			lc.RuleID = v.RuleID
			lc.Deleted = false
			break
		}
	}
	return lc, nil
}

// CreateLogCollector 创建日志采集规则
// @Summary 创建日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules [post]
func CreateLogCollector(c *rest.Context) (interface{}, error) {
	req := &CreateLogCollectorReq{}
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
	count, _, err := store.ListLogCollectors(c.Request.Context(), cond, &utils.ListOption{})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.Errorf("%s is exist", req.Name)
	}

	resp, err := bklog.CreateLogCollectors(c.Request.Context(), req.toBKLog(c))
	if err != nil {
		return nil, err
	}

	e := req.toEntity(c)
	e.FileIndexSetID = resp.FileIndexSetID
	e.STDIndexSetID = resp.STDIndexSetID
	e.RuleID = resp.RuleID
	err = store.CreateLogCollector(c.Request.Context(), e)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// UpdateLogCollector 更新日志采集规则
// @Summary 更新日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules/:id [put]
func UpdateLogCollector(c *rest.Context) (interface{}, error) {
	id := c.Param("id")
	if isBcsLogConfigID(id) {
		return nil, fmt.Errorf("can't update bcslogconfig")
	}
	req := &UpdateLogCollectorReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		klog.Errorf("UpdateLogCollector bind req json error, %s", err.Error())
		return nil, err
	}

	// check rule is exist
	store := storage.GlobalStorage
	lc, err := store.GetLogCollector(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}

	resp, err := bklog.UpdateLogCollectors(c.Request.Context(), lc.RuleID, req.toBKLog(c, lc))
	if err != nil {
		return nil, err
	}

	e := req.toEntity(c, lc)
	e.Update("file_index_set_id", resp.FileIndexSetID)
	e.Update("std_index_set_id", resp.STDIndexSetID)
	e.Update("rule_id", resp.GetRuleID())
	err = store.UpdateLogCollector(c.Request.Context(), id, e)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// DeleteLogCollector 删除日志采集规则
// @Summary 删除日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules/:id [delete]
func DeleteLogCollector(c *rest.Context) (interface{}, error) {
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
	lc, err := store.GetLogCollector(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}

	if lc.RuleID != 0 {
		// 删除 bklog 数据
		err = bklog.DeleteLogCollectors(c.Request.Context(), lc.RuleID)
		if err != nil {
			return nil, err
		}
	}

	err = store.DeleteLogCollector(c.Request.Context(), id)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ConvertLogCollector 转换日志采集规则
// @Summary 转换日志采集规则
// @Tags    LogCollectors
// @Produce json
// @Success 200
// @Router  /log_collector/rules/:id/conversion [post]
func ConvertLogCollector(c *rest.Context) (interface{}, error) {
	id := c.Param("id")
	if !isBcsLogConfigID(id) {
		return nil, fmt.Errorf("can't convert current log config")
	}

	// get bcs log config
	ns, name := getBcsLogConfigNamespaces(id)
	bcsLog, err := k8sclient.GetBcsLogConfig(c.Request.Context(), c.ClusterId, ns, name)
	if err != nil {
		return nil, err
	}

	// convert to bk log config and create bk log config
	lc := bcsLogToLogCollector(*bcsLog, c.ClusterId, nil)
	lc.Name, lc.RuleName = namespaceNameToRuleName(lc.Namespace, lc.Name)
	create := lc.toBKLog(c)
	resp, err := bklog.CreateLogCollectors(c.Request.Context(), create)
	if err != nil {
		return nil, err
	}

	// store in db
	e := lc.toEntity()
	e.FileIndexSetID = resp.FileIndexSetID
	e.STDIndexSetID = resp.STDIndexSetID
	e.RuleID = resp.RuleID
	store := storage.GlobalStorage
	err = store.CreateLogCollector(c.Request.Context(), e)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
