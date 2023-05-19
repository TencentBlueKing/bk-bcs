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

// LogCollector log collector
type LogCollector struct {
	SpaceUID              string               `json:"space_uid"`
	ProjectID             string               `json:"project_id"`
	CollectorConfigName   string               `json:"collector_config_name"`
	CollectorConfigNameEN string               `json:"collector_config_name_en"`
	Description           string               `json:"description"`
	BCSClusterID          string               `json:"bcs_cluster_id"`
	AddPodLabel           bool                 `json:"add_pod_label"`
	ExtraLabels           []Label              `json:"extra_labels,omitempty"`
	Config                []LogCollectorConfig `json:"config"`
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

// LogCollectorConfig log collector config
type LogCollectorConfig struct {
	Namespaces    []string                     `json:"namespaces"`
	Container     *LogCollectorConfigContainer `json:"container,omitempty"`
	LabelSelector *LogCollectorConfigSelector  `json:"label_selector,omitempty"`
	Paths         []string                     `json:"paths"`
	DataEncoding  string                       `json:"data_encoding"`
	EnableStdout  bool                         `json:"enable_stdout"`
}

// LogCollectorConfigContainer log collector config container
type LogCollectorConfigContainer struct {
	WorkloadType  string `json:"workload_type"`
	WorkloadName  string `json:"workload_name"`
	ContainerName string `json:"container_name"`
}

// LogCollectorConfigSelector log collector config selector
type LogCollectorConfigSelector struct {
	MatchLabels      []Label      `json:"match_labels"`
	MatchExpressions []Expression `json:"match_expressions"`
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
	RuleID                int    `json:"rule_id"`
	CollectorConfigName   string `json:"collector_config_name"`
	CollectorConfigNameEN string `json:"collector_config_name_en"`
	BCSClusterID          string `json:"bcs_cluster_id"`
	FileIndexSetID        int    `json:"file_index_set_id"`
	STDIndexSetID         int    `json:"std_index_set_id"`
}

// CreateBCSCollectorResp xxx
type CreateBCSCollectorResp struct {
	BaseResp
	Data CreateBCSCollectorRespData `json:"data"`
}

// CreateBCSCollectorRespData xxx
type CreateBCSCollectorRespData struct {
	RuleID         int `json:"rule_id"`
	FileIndexSetID int `json:"file_index_set_id"`
	STDIndexSetID  int `json:"std_index_set_id"`
	BKDataID       int `json:"bk_data_id"`
	STDOUTConf     struct {
		BKDataID int `json:"bk_data_id"`
	} `json:"stdout_conf"`
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
	STDOUTConf     struct {
		BKDataID int `json:"bk_data_id"`
	} `json:"stdout_conf"`
}

// GetRuleID get rule id, rule id is int or string
func (u *UpdateBCSCollectoRespData) GetRuleID() int {
	c, err := u.RuleID.Int64()
	if err != nil {
		return 0
	}
	return int(c)
}
