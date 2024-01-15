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

package logrule

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	logv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/apis/bkbcs/v1"
	"k8s.io/klog/v2"

	bklog "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

const (
	spaceUIDFormat        = "bkci__%s"
	bcsLogConfigSeparator = ":"
	bkLogPrefix           = "bklog|"
)

// GetLogRuleResp log rule resp
type GetLogRuleResp struct {
	ID                 string         `json:"id"`
	DisplayName        string         `json:"display_name"`
	Name               string         `json:"name"`
	RuleID             int            `json:"rule_id"`
	RuleName           string         `json:"rule_name"`
	Description        string         `json:"description"`
	FileIndexSetID     int            `json:"file_index_set_id"`
	STDIndexSetID      int            `json:"std_index_set_id"`
	RuleFileIndexSetID int            `json:"rule_file_index_set_id"`
	RuleSTDIndexSetID  int            `json:"rule_std_index_set_id"`
	Config             bklog.LogRule  `json:"rule"`
	CreatedAt          utils.JSONTime `json:"created_at"`
	UpdatedAt          utils.JSONTime `json:"updated_at"`
	Creator            string         `json:"creator"`
	Updator            string         `json:"updator"`
	Old                bool           `json:"old"`
	NewRuleID          string         `json:"new_rule_id"` // 当旧规则转换过新规则后，显示该字段
	Entrypoint         Entrypoint     `json:"entrypoint"`
	Status             string         `json:"status"`
	Message            string         `json:"message"`
}

// GetLogRuleRespSortByUpdateTime sort LogRule by update time
type GetLogRuleRespSortByUpdateTime []*GetLogRuleResp

// Len xxx
func (l GetLogRuleRespSortByUpdateTime) Len() int { return len(l) }

// Less xxx
func (l GetLogRuleRespSortByUpdateTime) Less(i, j int) bool {
	return l[i].UpdatedAt.After(l[j].UpdatedAt.Time)
}

// Swap xxx
func (l GetLogRuleRespSortByUpdateTime) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// GetLogRuleRespSortByName sort LogRule by name
type GetLogRuleRespSortByName []*GetLogRuleResp

// Len xxx
func (l GetLogRuleRespSortByName) Len() int { return len(l) }

// Less xxx
func (l GetLogRuleRespSortByName) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

// Swap xxx
func (l GetLogRuleRespSortByName) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// GetLogRuleRespSortByStatus sort LogRule by status
type GetLogRuleRespSortByStatus []*GetLogRuleResp

// Len xxx
func (l GetLogRuleRespSortByStatus) Len() int { return len(l) }

// Less xxx
func (l GetLogRuleRespSortByStatus) Less(i, j int) bool {
	return l[i].Status == entity.PendingStatus
}

// Swap xxx
func (l GetLogRuleRespSortByStatus) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

// CreateLogRuleReq req
type CreateLogRuleReq struct {
	DisplayName string        `json:"display_name" form:"display_name"`
	Name        string        `json:"name" form:"name" binding:"required" validate:"max=30,min=5,regexp=^[A-Za-z0-9_]+$"`
	RuleName    string        `json:"-" form:"-"`
	Description string        `json:"description"`
	Rule        bklog.LogRule `json:"rule"`
	FromRule    string        `json:"from_rule"`
}

// toEntity convert to entity.LogRule
func (req *CreateLogRuleReq) toEntity(c *rest.Context) *entity.LogRule {
	if req.DisplayName == "" {
		req.DisplayName = req.Name
	}
	return &entity.LogRule{
		DisplayName: req.DisplayName,
		Name:        req.Name,
		RuleName:    req.RuleName,
		Description: req.Description,
		ProjectID:   c.ProjectId,
		ProjectCode: c.ProjectCode,
		ClusterID:   c.ClusterId,
		Rule:        req.Rule,
		Creator:     c.Username,
		Updator:     c.Username,
		Status:      entity.PendingStatus,
		FromRule:    req.FromRule,
	}
}

// toBKLog convert to bklog.CreateBCSCollectorReq
func (req *CreateLogRuleReq) toBKLog(c *rest.Context) *bklog.CreateBCSCollectorReq {
	if req.Rule.LogRuleContainer.DataEncoding == "" {
		req.Rule.LogRuleContainer.DataEncoding = bklog.DefaultEncoding
	}
	matchLabels, matchExpressions := bklog.MergeOutLabels(req.Rule.LogRuleContainer.LabelSelector.MatchLabels,
		req.Rule.LogRuleContainer.LabelSelector.MatchExpressions)
	req.Rule.LogRuleContainer.LabelSelector = bklog.LabelSelector{
		MatchLabels: matchLabels, MatchExpressions: matchExpressions}
	return &bklog.CreateBCSCollectorReq{
		SpaceUID:              GetSpaceID(c.ProjectCode),
		ProjectID:             c.ProjectId,
		CollectorConfigName:   req.RuleName,
		CollectorConfigNameEN: req.RuleName,
		Description:           req.Description,
		BCSClusterID:          c.ClusterId,
		AddPodLabel:           req.Rule.AddPodLabel,
		ExtraLabels:           req.Rule.ExtraLabels,
		LogRuleContainer:      []bklog.LogRuleContainer{req.Rule.LogRuleContainer},
	}
}

// UpdateLogRuleReq req
type UpdateLogRuleReq struct {
	DisplayName string        `json:"display_name" form:"display_name"`
	Description string        `json:"description"`
	Rule        bklog.LogRule `json:"rule"`
}

// toEntity convert to entity.LogRule
func (req *UpdateLogRuleReq) toEntity(username, projectCode, ruleName string) entity.M {
	if req.DisplayName == "" {
		req.DisplayName = ruleName
	}
	return entity.M{
		"displayName":              req.DisplayName,
		"description":              req.Description,
		entity.FieldKeyStatus:      entity.PendingStatus,
		entity.FieldKeyMessage:     "",
		entity.FieldKeyUpdator:     username,
		entity.FieldKeyProjectCode: projectCode,
		entity.FieldKeyRule:        req.Rule,
	}
}

// toBKLog convert to bklog.UpdateBCSCollectorReq
func (req *UpdateLogRuleReq) toBKLog(c *rest.Context, ruleName string) *bklog.UpdateBCSCollectorReq {
	if req.Rule.LogRuleContainer.DataEncoding == "" {
		req.Rule.LogRuleContainer.DataEncoding = bklog.DefaultEncoding
	}
	matchLabels, matchExpressions := bklog.MergeOutLabels(req.Rule.LogRuleContainer.LabelSelector.MatchLabels,
		req.Rule.LogRuleContainer.LabelSelector.MatchExpressions)
	req.Rule.LogRuleContainer.LabelSelector = bklog.LabelSelector{
		MatchLabels: matchLabels, MatchExpressions: matchExpressions}
	return &bklog.UpdateBCSCollectorReq{
		SpaceUID:              GetSpaceID(c.ProjectCode),
		ProjectID:             c.ProjectId,
		CollectorConfigName:   ruleName,
		CollectorConfigNameEN: ruleName,
		Description:           req.Description,
		BCSClusterID:          c.ClusterId,
		AddPodLabel:           req.Rule.AddPodLabel,
		ExtraLabels:           req.Rule.ExtraLabels,
		LogRuleContainer:      []bklog.LogRuleContainer{req.Rule.LogRuleContainer},
	}
}

// GetSpaceID get space id
func GetSpaceID(projectCode string) string {
	return fmt.Sprintf(spaceUIDFormat, projectCode)
}

// This function checks if the given ID string is a valid BCS log configuration ID.
// A BCS log configuration ID is considered valid if it contains a colon (":").
func isBcsLogConfigID(id string) bool {
	return strings.Contains(id, bcsLogConfigSeparator)
}

// The toBcsLogConfigID function takes in two strings 'namespace' and 'name'
// and returns the concatenation of the two strings with a colon in between
// as a single string. This is used as the unique identifier for a logging configuration
// in the BCS.
func toBcsLogConfigID(namespace, name string) string {
	return fmt.Sprintf("%s%s%s", namespace, bcsLogConfigSeparator, name)
}

// getBcsLogConfigNamespaces takes in a string 'id' and splits it into two strings
func getBcsLogConfigNamespaces(id string) (string, string) {
	s := strings.Split(id, bcsLogConfigSeparator)
	if len(s) < 2 {
		return "", ""
	}
	return s[0], s[1]
}

// isBKLogID check if id is bklog id
func isBKLogID(id string) bool {
	return strings.HasPrefix(id, bkLogPrefix)
}

// toBKLogID convert to bklog id
func toBKLogID(name string) string {
	return fmt.Sprintf("%s%s", bkLogPrefix, name)
}

// getBKLogName get bklog name
func getBKLogName(id string) string {
	return strings.TrimPrefix(id, bkLogPrefix)
}

// 转换 bcslogconfig 到通用规则
func (resp *GetLogRuleResp) loadFromBcsLogConfig(logConfig *logv1.BcsLogConfig, logIndexID *entity.LogIndex,
	newRuleID string) {
	id := toBcsLogConfigID(logConfig.Namespace, logConfig.Name)
	resp.ID = id
	resp.DisplayName = logConfig.Name
	resp.Name = logConfig.Name
	resp.CreatedAt = utils.JSONTime(logConfig.CreationTimestamp)
	resp.UpdatedAt = utils.JSONTime(logConfig.CreationTimestamp)
	resp.Creator = "Client"
	resp.Updator = "Client"
	resp.Old = true
	resp.NewRuleID = newRuleID
	resp.Status = bklog.SuccessStatus
	resp.Config.LogRuleContainer.Namespaces = make([]string, 0)
	if logIndexID != nil {
		resp.FileIndexSetID = logIndexID.FileIndexSetID
		resp.STDIndexSetID = logIndexID.STDIndexSetID
		resp.RuleFileIndexSetID = logIndexID.FileIndexSetID
		resp.RuleSTDIndexSetID = logIndexID.STDIndexSetID
		resp.Entrypoint = Entrypoint{
			STDLogURL:  fmt.Sprintf("%s/#/retrieve/%d", config.G.BKLog.Entrypoint, logIndexID.STDIndexSetID),
			FileLogURL: fmt.Sprintf("%s/#/retrieve/%d", config.G.BKLog.Entrypoint, logIndexID.FileIndexSetID),
		}
	}
	resp.Config = bklog.LogRule{
		AddPodLabel: logConfig.Spec.PodLabels,
		ExtraLabels: make([]bklog.Label, 0),
		LogRuleContainer: bklog.LogRuleContainer{
			Namespaces:   make([]string, 0),
			DataEncoding: bklog.DefaultEncoding,
			EnableStdout: logConfig.Spec.Stdout,
			Conditions:   &bklog.Conditions{Type: "match", MatchType: "include"},
			LabelSelector: bklog.LabelSelector{
				MatchLabels:      make([]bklog.Label, 0),
				MatchExpressions: make([]bklog.Label, 0),
			},
			Container: bklog.Container{
				WorkloadType: logConfig.Spec.WorkloadType,
				WorkloadName: logConfig.Spec.WorkloadName,
			},
		},
	}
	resp.Config.LogRuleContainer.Paths = append(resp.Config.LogRuleContainer.Paths, logConfig.Spec.LogPaths...)
	for tagk, tagv := range logConfig.Spec.LogTags {
		resp.Config.ExtraLabels = append(resp.Config.ExtraLabels, bklog.Label{Key: tagk, Value: tagv})
	}
	for tagk, tagv := range logConfig.Spec.Selector.MatchLabels {
		resp.Config.LogRuleContainer.LabelSelector.MatchLabels = append(resp.Config.ExtraLabels, // nolint
			bklog.Label{Key: tagk, Operator: "=", Value: tagv})
	}
	if logConfig.Spec.WorkloadNamespace != "" {
		resp.Config.LogRuleContainer.Namespaces = []string{logConfig.Spec.WorkloadNamespace}
	}
	if len(logConfig.Spec.Selector.MatchExpressions) != 0 {
		for _, v := range logConfig.Spec.Selector.MatchExpressions {
			value := ""
			if len(v.Values) > 0 {
				value = v.Values[0]
			}
			resp.Config.LogRuleContainer.LabelSelector.MatchExpressions = append(
				resp.Config.LogRuleContainer.LabelSelector.MatchExpressions,
				bklog.Label{Key: v.Key, Value: value, Operator: v.Operator})
		}
	}
	if len(logConfig.Spec.ContainerConfs) > 0 {
		names := []string{}
		for _, v := range logConfig.Spec.ContainerConfs {
			names = append(names, v.ContainerName)
			resp.Config.LogRuleContainer.Paths = append(resp.Config.LogRuleContainer.Paths, v.LogPaths...)
			resp.Config.LogRuleContainer.EnableStdout = v.Stdout
			for tagk, tagv := range v.LogTags {
				resp.Config.ExtraLabels = append(resp.Config.ExtraLabels, bklog.Label{
					Key: tagk, Operator: "=", Value: tagv})
			}
		}
		resp.Config.LogRuleContainer.Container.ContainerName = strings.Join(names, ",")
	}
}

// 转换 entity.LogRule 到通用规则
func (resp *GetLogRuleResp) loadFromBkLog(rule bklog.ListBCSCollectorRespData, projectCode string) {
	resp.ID = toBKLogID(rule.CollectorConfigNameEN)
	resp.Name = rule.CollectorConfigName
	resp.DisplayName = rule.CollectorConfigName
	// 从日志平台创建的规则禁止编辑
	resp.RuleID = -1
	resp.RuleName = rule.CollectorConfigNameEN
	resp.Description = rule.Description
	resp.FileIndexSetID = rule.FileIndexSetID
	resp.STDIndexSetID = rule.STDIndexSetID
	resp.RuleFileIndexSetID = rule.RuleFileIndexSetID
	resp.RuleSTDIndexSetID = rule.RuleSTDIndexSetID
	resp.CreatedAt = utils.JSONTime{Time: rule.CreatedAt}
	resp.UpdatedAt = utils.JSONTime{Time: rule.UpdatedAt}
	resp.Creator = rule.Creator
	resp.Updator = rule.Updator
	resp.Status = rule.Status()
	resp.Message = rule.Message()
	resp.Entrypoint = Entrypoint{
		STDLogURL: fmt.Sprintf("%s/#/retrieve/%d?spaceUid=bkci__%s", strings.TrimRight(config.G.BKLog.Entrypoint, "/"),
			rule.STDIndexSetID, projectCode),
		FileLogURL: fmt.Sprintf("%s/#/retrieve/%d?spaceUid=bkci__%s", strings.TrimRight(config.G.BKLog.Entrypoint, "/"),
			rule.FileIndexSetID, projectCode),
	}
	resp.Config = bklog.LogRule{
		ExtraLabels: make([]bklog.Label, 0),
		LogRuleContainer: bklog.LogRuleContainer{
			Paths: make([]string, 0),
		},
	}
	resp.Config = rule.ToLogRule()
	// append bkbase info
	if resp.Config.DataInfo.FileBKDataDataID != 0 {
		resp.Entrypoint.FileBKBaseURL = getBKBaseEntrypoing(config.G.BKLog.BKBaseEntrypoint,
			resp.Config.DataInfo.FileBKDataDataID)
	}
	if resp.Config.DataInfo.StdBKDataDataID != 0 {
		resp.Entrypoint.STDBKBaseURL = getBKBaseEntrypoing(config.G.BKLog.BKBaseEntrypoint,
			resp.Config.DataInfo.StdBKDataDataID)
	}
}

// 转换 entity.LogRule 到通用规则
func (resp *GetLogRuleResp) loadFromEntity(e *entity.LogRule, lcs []bklog.ListBCSCollectorRespData) {
	resp.ID = e.ID.Hex()
	resp.DisplayName = e.DisplayName
	resp.Name = e.Name
	resp.RuleID = e.RuleID
	resp.RuleName = e.RuleName
	resp.Description = e.Description
	resp.FileIndexSetID = e.FileIndexSetID
	resp.STDIndexSetID = e.STDIndexSetID
	resp.RuleFileIndexSetID = e.RuleFileIndexSetID
	resp.RuleSTDIndexSetID = e.RuleSTDIndexSetID
	resp.CreatedAt = e.CreatedAt
	resp.UpdatedAt = e.UpdatedAt
	resp.Creator = e.Creator
	resp.Updator = e.Updator
	resp.Status = e.Status
	resp.Message = e.Message
	resp.Entrypoint = Entrypoint{
		STDLogURL: fmt.Sprintf("%s/#/retrieve/%d?spaceUid=bkci__%s", strings.TrimRight(config.G.BKLog.Entrypoint, "/"),
			e.RuleSTDIndexSetID, e.ProjectCode),
		FileLogURL: fmt.Sprintf("%s/#/retrieve/%d?spaceUid=bkci__%s", strings.TrimRight(config.G.BKLog.Entrypoint, "/"),
			e.RuleFileIndexSetID, e.ProjectCode),
	}
	resp.Config = bklog.LogRule{
		ExtraLabels: make([]bklog.Label, 0),
		LogRuleContainer: bklog.LogRuleContainer{
			Namespaces: e.Rule.LogRuleContainer.Namespaces,
			Paths:      make([]string, 0),
		},
	}

	// append bklog rule
	found := false
	for _, v := range lcs {
		if e.RuleID == v.RuleID {
			found = true
			resp.Config = v.ToLogRule()
			// append bkbase info
			if resp.Config.DataInfo.FileBKDataDataID != 0 {
				resp.Entrypoint.FileBKBaseURL = getBKBaseEntrypoing(config.G.BKLog.BKBaseEntrypoint,
					resp.Config.DataInfo.FileBKDataDataID)
			}
			if resp.Config.DataInfo.StdBKDataDataID != 0 {
				resp.Entrypoint.STDBKBaseURL = getBKBaseEntrypoing(config.G.BKLog.BKBaseEntrypoint,
					resp.Config.DataInfo.StdBKDataDataID)
			}
			if resp.Status == entity.FailedStatus || resp.Status == entity.PendingStatus {
				break
			}
			if v.Status() != "" {
				resp.Status = v.Status()
			}
			if v.Message() != "" {
				resp.Message = v.Message()
			}
			break
		}
	}

	// log is deleted from bklog
	if !found && e.Status == entity.SuccessStatus {
		resp.RuleID = 0
		resp.Status = entity.DeletedStatus
	}
}

// getContainerQueryLogLinks get container query log links
func getContainerQueryLogLinks(containerIDs []string, stdIndexSetID, fileIndexSetID int,
	projectCode string) map[string]Entrypoint {
	defaultEntrypoint := Entrypoint{
		STDLogURL:  fmt.Sprintf("%s/#/retrieve", config.G.BKLog.Entrypoint),
		FileLogURL: fmt.Sprintf("%s/#/retrieve", config.G.BKLog.Entrypoint),
	}
	result := make(map[string]Entrypoint, 0)
	if stdIndexSetID == 0 || fileIndexSetID == 0 {
		for _, v := range containerIDs {
			result[v] = defaultEntrypoint
		}
		return result
	}

	type addition struct {
		Field    string `json:"field"`
		Operator string `json:"operator"`
		Value    string `json:"value"`
	}

	for _, v := range containerIDs {
		id := strings.TrimPrefix(v, "containerd://")
		addition := []addition{{Field: "__ext.container_id", Operator: "=", Value: id}}
		additionData, _ := json.Marshal(addition)
		query := url.Values{}
		query.Add("spaceUid", GetSpaceID(projectCode))
		query.Add("addition", string(additionData))
		result[v] = Entrypoint{
			STDLogURL:  fmt.Sprintf("%s/#/retrieve/%d?%s", config.G.BKLog.Entrypoint, stdIndexSetID, query.Encode()),
			FileLogURL: fmt.Sprintf("%s/#/retrieve/%d?%s", config.G.BKLog.Entrypoint, fileIndexSetID, query.Encode()),
		}
	}
	return result
}

// getRuleIDByNames get rule id by names
func getRuleIDByName(projectID, clusterID, ruleName string) (string, error) {
	store := storage.GlobalStorage
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectID: projectID,
		entity.FieldKeyClusterID: clusterID,
		entity.FieldKeyRuleName:  ruleName,
	})
	_, list, err := store.ListLogRules(context.Background(), cond, &utils.ListOption{})
	if len(list) != 1 {
		return "", err
	}
	return list[0].ID.Hex(), nil
}

// createBKlog create bklog
func createBKLog(req *bklog.CreateBCSCollectorReq) {
	klog.Infof("ready to create bklog, req: %s", req)
	ctx := context.Background()
	store := storage.GlobalStorage
	// get log rule
	ruleID, err := getRuleIDByName(req.ProjectID, req.BCSClusterID, req.CollectorConfigNameEN)
	if err != nil {
		klog.Errorf("can't find log rules with rule_name %s", req.CollectorConfigNameEN)
		return
	}

	// create bk log
	resp, err := bklog.CreateLogCollectors(ctx, req)
	if err != nil {
		klog.Errorf("create bklog error, %s", err.Error())
		// report fail status
		err = store.UpdateLogRule(ctx, ruleID, entity.M{
			entity.FieldKeyStatus:  entity.FailedStatus,
			entity.FieldKeyMessage: err.Error(),
		})
		if err != nil {
			klog.Errorf("UpdateLogRule error, %s", err.Error())
		}
		return
	}

	// update db
	err = store.UpdateLogRule(ctx, ruleID, entity.M{
		entity.FieldKeyRuleID:             resp.RuleID,
		entity.FieldKeyStatus:             entity.SuccessStatus,
		entity.FieldKeyMessage:            "",
		entity.FieldKeyFileIndexSetID:     resp.FileIndexSetID,
		entity.FieldKeyStdIndexSetID:      resp.STDIndexSetID,
		entity.FieldKeyRuleFileIndexSetID: resp.RuleFileIndexSetID,
		entity.FieldKeyRuleStdIndexSetID:  resp.RuleSTDIndexSetID,
	})
	if err != nil {
		klog.Errorf("UpdateLogRule error, %s", err.Error())
	}
}

// updateBKLog update bklog
func updateBKLog(ruleID string, bkRuleID int, req *bklog.UpdateBCSCollectorReq) {
	klog.Infof("ready to update bklog, req: %s", req)
	ctx := context.Background()
	store := storage.GlobalStorage

	// update bk log
	resp, err := bklog.UpdateLogCollectors(ctx, bkRuleID, req)
	if err != nil {
		klog.Errorf("update bklog error, %s", err.Error())
		// report fail status
		err = store.UpdateLogRule(ctx, ruleID, entity.M{
			entity.FieldKeyStatus:  entity.FailedStatus,
			entity.FieldKeyMessage: err.Error(),
		})
		if err != nil {
			klog.Errorf("UpdateLogRule error, %s", err.Error())
		}
		return
	}

	// update db
	err = store.UpdateLogRule(ctx, ruleID, entity.M{
		entity.FieldKeyStatus:             entity.SuccessStatus,
		entity.FieldKeyMessage:            "",
		entity.FieldKeyFileIndexSetID:     resp.FileIndexSetID,
		entity.FieldKeyStdIndexSetID:      resp.STDIndexSetID,
		entity.FieldKeyRuleFileIndexSetID: resp.RuleFileIndexSetID,
		entity.FieldKeyRuleStdIndexSetID:  resp.RuleSTDIndexSetID,
	})
	if err != nil {
		klog.Errorf("UpdateLogRule error, %s", err.Error())
	}
}

// getFromRuleIDFromLogRules get from rule id from log rules
func getFromRuleIDFromLogRules(id string, rules []*entity.LogRule) string {
	for _, v := range rules {
		if v.FromRule == id {
			return v.ID.Hex()
		}
	}
	return ""
}

// getClusterLogRules get cluster log rules
func getClusterLogRules(ctx context.Context, projectID, clusterID string) ([]*entity.LogRule, error) {
	store := storage.GlobalStorage
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectID: projectID,
		entity.FieldKeyClusterID: clusterID,
	})
	listOption := &utils.ListOption{Sort: map[string]interface{}{"updatedAt": 1}}
	_, listInDB, err := store.ListLogRules(ctx, cond, listOption)
	return listInDB, err
}

// getBKBaseEntrypoing get bkbase entrypoint
func getBKBaseEntrypoing(host string, dataID int) string {
	return fmt.Sprintf("%s/#/data-hub-detail/index/%d?data_scenario=custom",
		strings.TrimRight(host, "/"), dataID)
}
