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
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	logv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubebkbcs/apis/bkbcs/v1"
	"k8s.io/apimachinery/pkg/util/rand"

	bklog "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// LogCollector log collector
type LogCollector struct {
	ID             string                    `json:"id"`
	Name           string                    `json:"name"`
	RuleName       string                    `json:"rule_name"`
	RuleID         int                       `json:"rule_id"`
	Description    string                    `json:"description"`
	ProjectID      string                    `json:"project_id"`
	ProjectCode    string                    `json:"project_code"`
	ClusterID      string                    `json:"cluster_id"`
	Namespace      string                    `json:"namespace"`
	AddPodLabel    bool                      `json:"add_pod_label"`
	ExtraLabels    map[string]string         `json:"extra_labels"`
	ConfigSelected entity.ConfigSelected     `json:"config_selected"`
	Config         entity.LogCollectorConfig `json:"config"`
	FileIndexSetID int                       `json:"file_index_set_id"`
	STDIndexSetID  int                       `json:"std_index_set_id"`
	CreatedAt      utils.JSONTime            `json:"created_at"`
	UpdatedAt      utils.JSONTime            `json:"updated_at"`
	Creator        string                    `json:"creator"`
	Updator        string                    `json:"updator"`
	Deleted        bool                      `json:"deleted"`
	Old            bool                      `json:"old"`
	Convertibility bool                      `json:"convertibility"` // 是否可以转换为新的日志采集配置
	Entrypoint     Entrypoint                `json:"entrypoint"`
}

func mergeLogCollector(lcs []LogCollector, resp []bklog.ListBCSCollectorRespData) []LogCollector {
	for i := range lcs {
		lcs[i].Deleted = true
		for _, lc := range resp {
			if lcs[i].Name == lc.CollectorConfigNameEN && lcs[i].ClusterID == lc.BCSClusterID {
				lcs[i].Deleted = false
				lcs[i].RuleID = lc.RuleID
				break
			}
		}
	}
	result := make([]LogCollector, 0)
	for i := range lcs {
		lcs[i].Entrypoint = Entrypoint{
			STDLogURL: fmt.Sprintf("%s/#/retrieve/%d?spaceUid=bkci__%s", config.G.BKLog.Entrypoint,
				lcs[i].STDIndexSetID, lcs[i].ProjectCode),
			FileLogURL: fmt.Sprintf("%s/#/retrieve/%d?spaceUid=bkci__%s", config.G.BKLog.Entrypoint,
				lcs[i].FileIndexSetID, lcs[i].ProjectCode),
		}
		result = append(result, lcs[i])
	}
	return result
}

const (
	spaceUIDFormat = "bkci__%s"
)

func (req *CreateLogCollectorReq) toBKLog(c *rest.Context) *bklog.LogCollector {
	lc := &bklog.LogCollector{
		SpaceUID:              fmt.Sprintf(spaceUIDFormat, c.ProjectCode),
		ProjectID:             c.ProjectId,
		CollectorConfigName:   req.RuleName,
		CollectorConfigNameEN: req.RuleName,
		BCSClusterID:          c.ClusterId,
		AddPodLabel:           req.AddPodLabel,
		ExtraLabels:           make([]bklog.Label, 0),
	}
	if req.ExtraLabels != nil {
		for k, v := range req.ExtraLabels {
			lc.ExtraLabels = append(lc.ExtraLabels, bklog.Label{Key: k, Value: v})
		}
	}
	lc.Config = make([]bklog.LogCollectorConfig, 0)
	switch req.ConfigSelected {
	case entity.AllContainers:
		lc.Config = req.Config.AllContainers.ToConfig(req.Namespace)
	case entity.SelectedContainers:
		lc.Config = req.Config.Workload.ToConfig(req.Namespace)
	case entity.SelectedLabels:
		lc.Config = req.Config.LabelSelector.ToConfig(req.Namespace)
	}
	return lc
}

func (req *CreateLogCollectorReq) toEntity(c *rest.Context) *entity.LogCollector {
	if req.ExtraLabels == nil {
		req.ExtraLabels = make(map[string]string, 0)
	}
	lc := &entity.LogCollector{
		Name:           req.Name,
		RuleName:       req.RuleName,
		Description:    req.Description,
		ProjectID:      c.ProjectId,
		ProjectCode:    c.ProjectCode,
		ClusterID:      c.ClusterId,
		Namespace:      req.Namespace,
		AddPodLabel:    req.AddPodLabel,
		ExtraLabels:    req.ExtraLabels,
		ConfigSelected: req.ConfigSelected,
		Config:         req.Config,
		Creator:        c.Username,
		Updator:        c.Username,
	}
	return lc
}

func (req *UpdateLogCollectorReq) toBKLog(c *rest.Context, origin *entity.LogCollector) *bklog.LogCollector {
	lc := &bklog.LogCollector{
		SpaceUID:              fmt.Sprintf(spaceUIDFormat, c.ProjectCode),
		ProjectID:             c.ProjectId,
		CollectorConfigName:   origin.RuleName,
		CollectorConfigNameEN: origin.RuleName,
		BCSClusterID:          c.ClusterId,
		AddPodLabel:           req.AddPodLabel,
		ExtraLabels:           make([]bklog.Label, 0),
	}
	if req.ExtraLabels != nil {
		for k, v := range req.ExtraLabels {
			lc.ExtraLabels = append(lc.ExtraLabels, bklog.Label{Key: k, Value: v})
		}
	}
	lc.Config = make([]bklog.LogCollectorConfig, 0)
	switch req.ConfigSelected {
	case entity.AllContainers:
		lc.Config = req.Config.AllContainers.ToConfig(origin.Namespace)
	case entity.SelectedContainers:
		lc.Config = req.Config.Workload.ToConfig(origin.Namespace)
	case entity.SelectedLabels:
		lc.Config = req.Config.LabelSelector.ToConfig(origin.Namespace)
	}
	return lc
}

func (req *UpdateLogCollectorReq) toEntity(c *rest.Context, origin *entity.LogCollector) entity.M {
	return entity.M{
		"description":          req.Description,
		"add_pod_label":        req.AddPodLabel,
		"extra_labels":         req.ExtraLabels,
		"config":               req.Config,
		entity.FieldKeyUpdator: c.Username,
	}
}

// This function checks if the given ID string is a valid BCS log configuration ID.
// A BCS log configuration ID is considered valid if it contains a hyphen ("-").
func isBcsLogConfigID(id string) bool {
	return strings.Contains(id, "-")
}

// The toBcsLogConfigID function takes in two strings 'namespace' and 'name'
// and returns the concatenation of the two strings with a hyphen in between
// as a single string. This is used as the unique identifier for a logging configuration
// in the BCS.
func toBcsLogConfigID(namespace, name string) string {
	return fmt.Sprintf("%s-%s", namespace, name)
}

func getBcsLogConfigNamespaces(id string) (string, string) {
	s := strings.Split(id, "-")
	if len(s) < 2 {
		return "", ""
	}
	return s[0], s[1]
}

func bcsLogToLogCollector(config logv1.BcsLogConfig, clusterID string, logIndexID *entity.LogIndex) LogCollector {
	lc := LogCollector{
		ID:          toBcsLogConfigID(config.Namespace, config.Name),
		Name:        config.Name,
		ProjectID:   "",
		ProjectCode: "",
		ClusterID:   clusterID,
		Namespace:   config.Namespace,
		AddPodLabel: config.Spec.PodLabels,
		ExtraLabels: config.Spec.LogTags,
		CreatedAt:   utils.JSONTime(config.CreationTimestamp),
		UpdatedAt:   utils.JSONTime(config.CreationTimestamp),
		Creator:     "Client",
		Updator:     "Client",
		Old:         true,
	}
	if logIndexID != nil {
		lc.FileIndexSetID = logIndexID.FileIndexSetID
		lc.STDIndexSetID = logIndexID.STDIndexSetID
		// 没有使用自定义 dataid 的才能一键转换规则，因为使用一键转换规则会重新创建 dataid，而自定义 dataid 一般是有设置清洗规则或者自建
		// ES 集群，不适合一键转换
		if config.Spec.StdDataId == strconv.Itoa(logIndexID.STDDataID) &&
			config.Spec.NonStdDataId == strconv.Itoa(logIndexID.FileDataID) {
			lc.Convertibility = true
		}
		for _, v := range config.Spec.ContainerConfs {
			lc.Convertibility = true
			if v.StdDataId != strconv.Itoa(logIndexID.STDDataID) || v.NonStdDataId != strconv.Itoa(logIndexID.FileDataID) {
				lc.Convertibility = false
				break
			}
		}
	}

	// selectedLables
	if config.Spec.Selector.MatchExpressions != nil || config.Spec.Selector.MatchLabels != nil {
		return bcsSelectedLabelsLogToLogCollector(config, lc)
	}

	// selected containers
	if config.Spec.WorkloadType != "" {
		return bcsSelectedContainersLogToLC(config, lc)
	}

	// all containers
	return bcsAllContainersLogToLogCollector(config, lc)
}

func bcsSelectedLabelsLogToLogCollector(config logv1.BcsLogConfig, lc LogCollector) LogCollector {
	lc.ConfigSelected = entity.SelectedLabels
	matchExpressions := make([]entity.Expression, 0)
	for _, v := range config.Spec.Selector.MatchExpressions {
		if len(v.Values) == 0 {
			continue
		}
		matchExpressions = append(matchExpressions, entity.Expression{
			Key:      v.Key,
			Operator: v.Operator,
			Value:    v.Values[0],
		})
	}
	lc.Config = entity.LogCollectorConfig{
		LabelSelector: &entity.LogCollectorConfigSelector{
			MatchLabels:  config.Spec.Selector.MatchLabels,
			Paths:        config.Spec.LogPaths,
			EnableStdout: config.Spec.Stdout,
		},
	}
	lc.Namespace = config.Spec.WorkloadNamespace
	return lc
}

func bcsSelectedContainersLogToLC(config logv1.BcsLogConfig, lc LogCollector) LogCollector {
	lc.ConfigSelected = entity.SelectedContainers
	containers := make([]entity.LogCollectorConfigContainer, 0)
	for _, v := range config.Spec.ContainerConfs {
		containers = append(containers, entity.LogCollectorConfigContainer{
			ContainerName: v.ContainerName,
			Paths:         v.LogPaths,
			EnableStdout:  v.Stdout,
		})
	}
	lc.Config = entity.LogCollectorConfig{
		Workload: &entity.LogCollectorConfigWorkload{
			Kind:       config.Spec.WorkloadType,
			Name:       config.Spec.WorkloadName,
			Containers: containers,
		},
	}
	lc.Namespace = config.Spec.WorkloadNamespace
	return lc
}

func bcsAllContainersLogToLogCollector(config logv1.BcsLogConfig, lc LogCollector) LogCollector {
	lc.ConfigSelected = entity.AllContainers
	lc.Config = entity.LogCollectorConfig{
		AllContainers: &entity.LogCollectorConfigAllContainers{
			Paths:        config.Spec.LogPaths,
			EnableStdout: config.Spec.Stdout,
		},
	}
	lc.Namespace = config.Spec.WorkloadNamespace
	return lc
}

func logEntityToLogCollector(e entity.LogCollector) LogCollector {
	return LogCollector{
		ID:             e.ID.Hex(),
		Name:           e.Name,
		RuleName:       e.RuleName,
		RuleID:         e.RuleID,
		Description:    e.Description,
		ProjectID:      e.ProjectID,
		ProjectCode:    e.ProjectCode,
		ClusterID:      e.ClusterID,
		Namespace:      e.Namespace,
		AddPodLabel:    e.AddPodLabel,
		ExtraLabels:    e.ExtraLabels,
		ConfigSelected: e.ConfigSelected,
		Config:         e.Config,
		FileIndexSetID: e.FileIndexSetID,
		STDIndexSetID:  e.STDIndexSetID,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
		Creator:        e.Creator,
		Updator:        e.Updator,
	}
}

func (current *LogCollector) toBKLog(c *rest.Context) *bklog.LogCollector {
	lc := &bklog.LogCollector{
		SpaceUID:              fmt.Sprintf(spaceUIDFormat, c.ProjectCode),
		ProjectID:             c.ProjectId,
		CollectorConfigName:   current.RuleName,
		CollectorConfigNameEN: current.RuleName,
		BCSClusterID:          c.ClusterId,
		AddPodLabel:           current.AddPodLabel,
		ExtraLabels:           make([]bklog.Label, 0),
	}
	if current.ExtraLabels != nil {
		for k, v := range current.ExtraLabels {
			lc.ExtraLabels = append(lc.ExtraLabels, bklog.Label{Key: k, Value: v})
		}
	}
	lc.Config = make([]bklog.LogCollectorConfig, 0)
	switch current.ConfigSelected {
	case entity.AllContainers:
		lc.Config = current.Config.AllContainers.ToConfig(current.Namespace)
	case entity.SelectedContainers:
		lc.Config = current.Config.Workload.ToConfig(current.Namespace)
	case entity.SelectedLabels:
		lc.Config = current.Config.LabelSelector.ToConfig(current.Namespace)
	}
	return lc
}

func (current *LogCollector) toEntity() *entity.LogCollector {
	if current.ExtraLabels == nil {
		current.ExtraLabels = make(map[string]string, 0)
	}
	lc := &entity.LogCollector{
		Name:           current.Name,
		RuleName:       current.RuleName,
		Description:    current.Description,
		ProjectID:      current.ProjectID,
		ProjectCode:    current.ProjectCode,
		ClusterID:      current.ClusterID,
		Namespace:      current.Namespace,
		AddPodLabel:    current.AddPodLabel,
		ExtraLabels:    current.ExtraLabels,
		ConfigSelected: current.ConfigSelected,
		Config:         current.Config,
		Creator:        current.Creator,
		Updator:        current.Updator,
	}
	return lc
}

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
		Field    string
		Operator string
		Value    string
	}

	for _, v := range containerIDs {
		addition := []addition{{Field: "container_id", Operator: "is", Value: v}}
		additionData, _ := json.Marshal(addition)
		query := url.Values{}
		query.Add("spaceUid", fmt.Sprintf("bkci__%s", projectCode))
		query.Add("addition", string(additionData))
		result[v] = Entrypoint{
			STDLogURL:  fmt.Sprintf("%s/#/retrieve/%d?%s", config.G.BKLog.Entrypoint, stdIndexSetID, query.Encode()),
			FileLogURL: fmt.Sprintf("%s/#/retrieve/%d?%s", config.G.BKLog.Entrypoint, fileIndexSetID, query.Encode()),
		}
	}
	return result
}

func namespaceNameToRuleName(namespace, name string) (string, string) {
	s := fmt.Sprintf("%s_%s", name, namespace)
	s = strings.ReplaceAll(s, "-", "_")
	ruleName := fmt.Sprintf("%s_%s", s[:30], rand.String(5))
	if len(s) > 30 {
		return ruleName, ruleName
	}
	return s, ruleName
}
