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

package bklog

import (
	"encoding/json"
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
}

// LabelSelector label selector
type LabelSelector struct {
	MatchLabels      []Label      `json:"match_labels,omitempty"`
	MatchExpressions []Expression `json:"match_expressions,omitempty"`
}

// Container container config
type Container struct {
	WorkloadType  string `json:"workload_type"`
	WorkloadName  string `json:"workload_name"`
	ContainerName string `json:"container_name"`
}

// Label label
type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Expression expression
type Expression struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
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
	BCSClusterID          string            `json:"bcs_cluster_id"`
	FileIndexSetID        int               `json:"file_index_set_id"`
	STDIndexSetID         int               `json:"std_index_set_id"`
	AddPodLabel           bool              `json:"add_pod_label"`
	ExtraLabels           []Label           `json:"extra_labels"`
	ContainerConfig       []ContainerConfig `json:"container_config"`
}

// ToLogRule trans resp to log rule
func (resp *ListBCSCollectorRespData) ToLogRule() LogRule {
	rule := LogRule{
		AddPodLabel: resp.AddPodLabel,
		ExtraLabels: make([]Label, 0),
	}
	if resp.ExtraLabels != nil {
		rule.ExtraLabels = resp.ExtraLabels
	}
	if len(resp.ContainerConfig) > 0 {
		conf := resp.ContainerConfig[0]
		namespaces := make([]string, 0)
		paths := make([]string, 0)
		if conf.Namespaces != nil {
			namespaces = conf.Namespaces
		}
		if conf.Params.Paths != nil {
			paths = conf.Params.Paths
		}
		rule.LogRuleContainer = LogRuleContainer{
			Namespaces:    namespaces,
			Paths:         paths,
			DataEncoding:  conf.DataEncoding,
			EnableStdout:  conf.EnableStdout,
			Conditions:    conf.Params.Conditions,
			LabelSelector: conf.LabelSelector,
			Container:     conf.Container,
		}
	}
	return rule
}

// Status return log status
func (resp *ListBCSCollectorRespData) Status() string {
	if len(resp.ContainerConfig) > 0 {
		return resp.ContainerConfig[0].Status
	}
	return ""
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
}

// StdoutConf stdout config
type StdoutConf struct {
	BkDataID int `json:"bk_data_id"`
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
	RuleID         int        `json:"rule_id"`
	FileIndexSetID int        `json:"file_index_set_id"`
	STDIndexSetID  int        `json:"std_index_set_id"`
	BKDataID       int        `json:"bk_data_id"`
	StdoutConf     StdoutConf `json:"stdout_conf"`
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
	RuleID         json.Number `json:"rule_id"`
	FileIndexSetID int         `json:"file_index_set_id"`
	STDIndexSetID  int         `json:"std_index_set_id"`
	BKDataID       int         `json:"bk_data_id"`
	StdoutConf     StdoutConf  `json:"stdout_conf"`
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
