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

package bklog

import (
	"encoding/json"
	"time"
)

const (
	// SuccessStatus log rule is health
	SuccessStatus = "SUCCESS"

	// DefaultEncoding default data encoding
	DefaultEncoding = "UTF-8"
)

// LogRule log rule
type LogRule struct {
	AddPodLabel      bool             `json:"add_pod_label"`
	ExtraLabels      []Label          `json:"extra_labels"`
	LogRuleContainer LogRuleContainer `json:"config"`
	DataInfo         DataInfo         `json:"data_info"`
}

// LogRuleContainer log rule container config
type LogRuleContainer struct {
	Namespaces    []string      `json:"namespaces"`
	Paths         []string      `json:"paths"`
	DataEncoding  string        `json:"data_encoding"`
	EnableStdout  bool          `json:"enable_stdout"`
	Conditions    *Conditions   `json:"conditions,omitempty"`
	LabelSelector LabelSelector `json:"label_selector,omitempty"`
	Container     Container     `json:"container,omitempty"`
	Multiline     *Multiline    `json:"multiline,omitempty"`
}

// LabelSelector label selector
type LabelSelector struct {
	MatchLabels      []Label `json:"match_labels,omitempty"`
	MatchExpressions []Label `json:"match_expressions,omitempty"`
}

// Container container config
type Container struct {
	WorkloadType  string `json:"workload_type"`
	WorkloadName  string `json:"workload_name"`
	ContainerName string `json:"container_name"`
}

// MergeInLabels merge labels
// matchExpressions 合并到 matchLabels
func MergeInLabels(matchLabels, matchExpressions []Label) []Label {
	matchLabels = FilterLabels(matchLabels)
	matchExpressions = FilterLabels(matchExpressions)
	matchLabels = append(matchLabels, matchExpressions...)
	return matchLabels
}

// MergeOutLabels merge labels
// matchLabels 分开非 = 的表达式到 matchExpressions
func MergeOutLabels(matchLabels, matchExpressions []Label) ([]Label, []Label) {
	matchLabels = FilterLabels(matchLabels)
	matchExpressions = FilterLabels(matchExpressions)
	newLabels := make([]Label, 0)
	for _, v := range matchLabels {
		if v.Operator != "=" {
			matchExpressions = append(matchExpressions, v)
			continue
		}
		newLabels = append(newLabels, v)
	}
	return newLabels, matchExpressions
}

// FilterLabels filter labels
func FilterLabels(labels []Label) []Label {
	newLabels := make([]Label, 0)
	for _, v := range labels {
		switch v.Operator {
		case "", "=":
			newLabels = append(newLabels, Label{Key: v.Key, Operator: "=", Value: v.Value})
		case "In":
			newLabels = append(newLabels, Label{Key: v.Key, Operator: "In", Value: v.Value})
		case "NotIn":
			newLabels = append(newLabels, Label{Key: v.Key, Operator: "NotIn", Value: v.Value})
		case "Exists":
			newLabels = append(newLabels, Label{Key: v.Key, Operator: "Exists", Value: v.Value})
		case "DoesNotExist":
			newLabels = append(newLabels, Label{Key: v.Key, Operator: "DoesNotExist", Value: v.Value})
		}
	}
	return newLabels
}

// Label label
type Label struct {
	Key      string `json:"key"`
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value"`
}

// Conditions content condition match
type Conditions struct {
	Type             string             `json:"type"`       // match, separator
	MatchType        string             `json:"match_type"` // include, exclude(开发中，暂不支持)
	MatchContent     string             `json:"match_content"`
	Separator        string             `json:"separator"`                   // 分隔符，| 等
	SeparatorFilters []SeparatorFilters `json:"separator_filters,omitempty"` // 分隔符过滤条件
}

// SeparatorFilters 分隔符过滤条件
type SeparatorFilters struct {
	FieldIndex json.Number `json:"fieldindex"` // 匹配项所在列
	LogicOp    string      `json:"logic_op"`   // and, or
	Op         string      `json:"op"`         // 匹配方式，目前只有 =
	Word       string      `json:"word"`       // 匹配值
}

// IntFieldIndex get int field index
func (s SeparatorFilters) IntFieldIndex() int {
	i, err := s.FieldIndex.Int64()
	if err != nil {
		return 1
	}
	return int(i)
}

// Multiline multi line
type Multiline struct {
	MultilinePattern  string      `json:"multiline_pattern"`
	MultilineMaxLines json.Number `json:"multiline_max_lines"`
	MultilineTimeout  json.Number `json:"multiline_timeout"`
}

// DataInfo data info
type DataInfo struct {
	FileDataID       int `json:"file_data_id"`        // 文件日志 dataid
	StdDataID        int `json:"std_data_id"`         // 标准输出日志 dataid
	FileBKDataDataID int `json:"file_bkdata_data_id"` // 计算平台文件日志 dataid
	StdBKDataDataID  int `json:"std_bkdata_data_id"`  // 计算平台标准输出日志 dataid
}

// BaseResp base resp
type BaseResp struct {
	Code      json.Number `json:"code"`
	Result    bool        `json:"result"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
}

// GetCode get resp code
func (b *BaseResp) GetCode() int {
	c, err := b.Code.Int64()
	if err != nil {
		return 1
	}
	return int(c)
}

// IsSuccess  method returns true if code == 0 and result is true, otherwise false.
func (b *BaseResp) IsSuccess() bool {
	return b.GetCode() == 0 && b.Result
}

// ListBCSCollectorResp xxx
type ListBCSCollectorResp struct {
	BaseResp
	Data []ListBCSCollectorRespData `json:"data"`
}

// ListBCSCollectorRespData xxx
type ListBCSCollectorRespData struct {
	RuleID                int               `json:"rule_id"`
	CollectorConfigName   string            `json:"collector_config_name"`
	CollectorConfigNameEN string            `json:"collector_config_name_en"`
	Description           string            `json:"description"`
	BCSClusterID          string            `json:"bcs_cluster_id"`
	FileIndexSetID        int               `json:"file_index_set_id"`
	STDIndexSetID         int               `json:"std_index_set_id"`
	RuleFileIndexSetID    int               `json:"rule_file_index_set_id"`
	RuleSTDIndexSetID     int               `json:"rule_std_index_set_id"`
	IsSTDDeleted          bool              `json:"is_std_deleted"`
	IsFileDeleted         bool              `json:"is_file_deleted"`
	Creator               string            `json:"created_by"`
	Updator               string            `json:"updated_by"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
	AddPodLabel           bool              `json:"add_pod_label"`
	ExtraLabels           []Label           `json:"extra_labels"`
	ContainerConfig       []ContainerConfig `json:"container_config"`
	FromBKLog             bool              `json:"from_bklog"`
}

// ToLogRule trans resp to log rule
func (resp *ListBCSCollectorRespData) ToLogRule() LogRule {
	rule := LogRule{
		AddPodLabel: resp.AddPodLabel,
		ExtraLabels: make([]Label, 0),
		LogRuleContainer: LogRuleContainer{
			Namespaces: make([]string, 0),
			Paths:      make([]string, 0),
		},
	}
	if resp.ExtraLabels != nil {
		rule.ExtraLabels = resp.ExtraLabels
	}
	if len(resp.ContainerConfig) > 0 {
		conf := resp.ContainerConfig[0]
		namespaces := make([]string, 0)
		paths := make([]string, 0)
		var multiline *Multiline
		if conf.Namespaces != nil {
			namespaces = conf.Namespaces
		}
		if conf.Params.Paths != nil {
			paths = conf.Params.Paths
		}
		if conf.Params.MultilinePattern != "" {
			multiline = &Multiline{MultilinePattern: conf.Params.MultilinePattern,
				MultilineMaxLines: conf.Params.MultilineMaxLines, MultilineTimeout: conf.Params.MultilineTimeout}
		}
		if conf.Params.Conditions != nil {
			if conf.Params.Conditions.MatchType == "" {
				conf.Params.Conditions.MatchType = "include"
			}
			if conf.Params.Conditions.Type == "" {
				conf.Params.Conditions.Type = "match"
			}
		}
		labelSelector := LabelSelector{}
		labelSelector.MatchLabels = MergeInLabels(conf.LabelSelector.MatchLabels, conf.LabelSelector.MatchExpressions)
		rule.LogRuleContainer = LogRuleContainer{
			Namespaces:    namespaces,
			Paths:         paths,
			DataEncoding:  conf.DataEncoding,
			EnableStdout:  conf.EnableStdout,
			Conditions:    conf.Params.Conditions,
			LabelSelector: labelSelector,
			Container:     conf.Container,
			Multiline:     multiline,
		}
		// append data info
		rule.DataInfo = DataInfo{
			FileDataID:       conf.BkDataID,
			StdDataID:        conf.StdoutConf.BkDataID,
			FileBKDataDataID: conf.BKDataDataID,
			StdBKDataDataID:  conf.StdoutConf.BKDataDataID,
		}
	}
	return rule
}

// Status return log status
func (resp *ListBCSCollectorRespData) Status() string {
	var status string
	if len(resp.ContainerConfig) > 0 {
		status = resp.ContainerConfig[0].Status
	}
	// bklog RUNNING status is PENDING status
	if status == "RUNNING" {
		status = "PENDING"
	}
	return status
}

// Message return log message
func (resp *ListBCSCollectorRespData) Message() string {
	if len(resp.ContainerConfig) > 0 {
		return resp.ContainerConfig[0].Message
	}
	return ""
}

// ContainerConfig container config
type ContainerConfig struct {
	ID            int           `json:"id"`
	BkDataID      int           `json:"bk_data_id"`
	BKDataDataID  int           `json:"bkdata_data_id"` // 计算平台 dataid
	Namespaces    []string      `json:"namespaces"`
	AnyNamespace  bool          `json:"any_namespace"`
	DataEncoding  string        `json:"data_encoding"`
	Params        Params        `json:"params"`
	Container     Container     `json:"container"`
	LabelSelector LabelSelector `json:"label_selector"`
	AllContainer  bool          `json:"all_container"`
	Status        string        `json:"status"`
	Message       string        `json:"status_detail"`
	EnableStdout  bool          `json:"enable_stdout"`
	StdoutConf    StdoutConf    `json:"stdout_conf"`
}

// Params container config params
type Params struct {
	Paths      []string    `json:"paths"`
	Conditions *Conditions `json:"conditions"`
	Multiline
}

// StdoutConf stdout config
type StdoutConf struct {
	BkDataID     int `json:"bk_data_id"`
	BKDataDataID int `json:"bkdata_data_id"` // 计算平台 dataid
}

// CreateBCSCollectorReq req
type CreateBCSCollectorReq struct {
	SpaceUID              string             `json:"space_uid"`
	ProjectID             string             `json:"project_id"`
	CollectorConfigName   string             `json:"collector_config_name"`
	CollectorConfigNameEN string             `json:"collector_config_name_en"`
	Description           string             `json:"description"`
	BCSClusterID          string             `json:"bcs_cluster_id"`
	AddPodLabel           bool               `json:"add_pod_label"`
	ExtraLabels           []Label            `json:"extra_labels,omitempty"`
	LogRuleContainer      []LogRuleContainer `json:"config"`
	Username              string             `json:"-"`
}

// String get req json string
func (req *CreateBCSCollectorReq) String() string {
	b, err := json.Marshal(req)
	if err != nil {
		return ""
	}
	return string(b)
}

// CreateBCSCollectorResp xxx
type CreateBCSCollectorResp struct {
	BaseResp
	Data CreateBCSCollectorRespData `json:"data"`
}

// CreateBCSCollectorRespData xxx
type CreateBCSCollectorRespData struct {
	RuleID             int        `json:"rule_id"`
	FileIndexSetID     int        `json:"file_index_set_id"`
	STDIndexSetID      int        `json:"std_index_set_id"`
	RuleFileIndexSetID int        `json:"rule_file_index_set_id"`
	RuleSTDIndexSetID  int        `json:"rule_std_index_set_id"`
	BKDataID           int        `json:"bk_data_id"`
	StdoutConf         StdoutConf `json:"stdout_conf"`
}

// UpdateBCSCollectorReq req
type UpdateBCSCollectorReq struct {
	SpaceUID              string             `json:"space_uid"`
	ProjectID             string             `json:"project_id"`
	CollectorConfigName   string             `json:"collector_config_name"`
	CollectorConfigNameEN string             `json:"collector_config_name_en"`
	Description           string             `json:"description"`
	BCSClusterID          string             `json:"bcs_cluster_id"`
	AddPodLabel           bool               `json:"add_pod_label"`
	ExtraLabels           []Label            `json:"extra_labels"`
	LogRuleContainer      []LogRuleContainer `json:"config"`
	Username              string             `json:"-"`
}

// String get req json string
func (req *UpdateBCSCollectorReq) String() string {
	b, err := json.Marshal(req)
	if err != nil {
		return ""
	}
	return string(b)
}

// UpdateBCSCollectorResp xxx
type UpdateBCSCollectorResp struct {
	BaseResp
	Data UpdateBCSCollectoRespData `json:"data"`
}

// UpdateBCSCollectoRespData xxx
type UpdateBCSCollectoRespData struct {
	RuleID             json.Number `json:"rule_id"`
	FileIndexSetID     int         `json:"file_index_set_id"`
	STDIndexSetID      int         `json:"std_index_set_id"`
	RuleFileIndexSetID int         `json:"rule_file_index_set_id"`
	RuleSTDIndexSetID  int         `json:"rule_std_index_set_id"`
	BKDataID           int         `json:"bk_data_id"`
	StdoutConf         StdoutConf  `json:"stdout_conf"`
}

// GetRuleID get rule id, rule id is int or string
func (u *UpdateBCSCollectoRespData) GetRuleID() int {
	c, err := u.RuleID.Int64()
	if err != nil {
		return 0
	}
	return int(c)
}

// QueryLogResp query log resp
type QueryLogResp struct {
	BaseResp
	Data struct {
		Hits struct {
			Total int `json:"total"`
		} `json:"hits"`
	} `json:"data"`
}

// GetStorageClustersReq xxx
type GetStorageClustersReq struct {
	ProjectId string `in:"path=projectId;required"`
	ClusterId string `in:"path=clusterId;required"`
}

// GetStorageClustersResp xxx
type GetStorageClustersResp struct {
	BaseResp
	Data []GetStorageClustersRespData `json:"data"`
}

// GetStorageClustersRespData xxx
type GetStorageClustersRespData struct {
	StorageClusterID   int     `json:"storage_cluster_id"`
	StorageClusterName string  `json:"storage_cluster_name"`
	StorageVersion     string  `json:"storage_version"`
	StorageUsage       int     `json:"storage_usage"`
	StorageTotal       float64 `json:"storage_total"`
	IsPlatform         bool    `json:"is_platform"`
	IsSelected         bool    `json:"is_selected"`
	Description        string  `json:"description"`
}

// GetBcsCollectorStorageResp xxx
type GetBcsCollectorStorageResp struct {
	BaseResp
	Data json.Number `json:"data"`
}
