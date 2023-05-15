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

package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"

	bklog "github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bk_log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// LogCollector log collector
type LogCollector struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	RuleName       string             `json:"rule_name" bson:"rule_name"`
	Name           string             `json:"name" bson:"name"`
	RuleID         int                `json:"rule_id" bson:"ruleID"`
	Description    string             `json:"description" bson:"description"`
	ProjectID      string             `json:"project_id" bson:"projectID"`
	ProjectCode    string             `json:"project_code" bson:"projectCode"`
	ClusterID      string             `json:"cluster_id" bson:"clusterID"`
	Namespace      string             `json:"namespace" bson:"namespace"`
	AddPodLabel    bool               `json:"add_pod_label" bson:"addPodLabel"`
	ExtraLabels    map[string]string  `json:"extra_labels" bson:"extraLabels"`
	ConfigSelected ConfigSelected     `json:"config_selected" bson:"configSelected"`
	Config         LogCollectorConfig `json:"config" bson:"config"`
	FileIndexSetID int                `json:"file_index_set_id" bson:"fileIndexSetID"`
	STDIndexSetID  int                `json:"std_index_set_id" bson:"stdIndexSetID"`
	CreatedAt      utils.JSONTime     `json:"created_at" bson:"createdAt"`
	UpdatedAt      utils.JSONTime     `json:"updated_at" bson:"updatedAt"`
	Creator        string             `json:"creator" bson:"creator"`
	Updator        string             `json:"updator" bson:"updator"`
}

// LogCollectorConfig log collector config
type LogCollectorConfig struct {
	Workload      *LogCollectorConfigWorkload      `json:"workload,omitempty" bson:"workload"`
	LabelSelector *LogCollectorConfigSelector      `json:"label_selector,omitempty" bson:"labelSelector"`
	AllContainers *LogCollectorConfigAllContainers `json:"all_containers,omitempty" bson:"allContainers"`
}

// LogCollectorConfigWorkload workload
type LogCollectorConfigWorkload struct {
	Kind       string                        `json:"kind" bson:"kind"`
	Name       string                        `json:"name" bson:"name"`
	Containers []LogCollectorConfigContainer `json:"containers" bson:"containers"`
}

// ToConfig to bklog config
func (s *LogCollectorConfigWorkload) ToConfig(namespace string) []bklog.LogCollectorConfig {
	configs := make([]bklog.LogCollectorConfig, 0)
	if s == nil {
		return configs
	}
	for _, v := range s.Containers {
		configs = append(configs, bklog.LogCollectorConfig{
			Namespaces: []string{namespace},
			Container: &bklog.LogCollectorConfigContainer{
				WorkloadType:  s.Kind,
				WorkloadName:  s.Name,
				ContainerName: v.ContainerName,
			},
			Paths:        v.Paths,
			DataEncoding: v.GetDataEncoding(),
			EnableStdout: v.EnableStdout,
		})
	}
	return configs
}

// LogCollectorConfigContainer log collector config container
type LogCollectorConfigContainer struct {
	ContainerName string   `json:"container_name" bson:"containerName"`
	Paths         []string `json:"paths" bson:"paths"`
	DataEncoding  string   `json:"data_encoding" bson:"dataEncoding"`
	EnableStdout  bool     `json:"enable_stdout" bson:"enableStdout"`
}

// GetDataEncoding get data encoding
func (s *LogCollectorConfigContainer) GetDataEncoding() string {
	if s.DataEncoding == "" {
		return DefaultDataEncoding
	}
	return s.DataEncoding
}

// LogCollectorConfigSelector log collector config selector
type LogCollectorConfigSelector struct {
	MatchLabels      map[string]string `json:"match_labels" bson:"matchLabels"`
	MatchExpressions []Expression      `json:"match_expressions" bson:"matchExpressions"`
	Paths            []string          `json:"paths" bson:"paths"`
	DataEncoding     string            `json:"data_encoding" bson:"dataEncoding"`
	EnableStdout     bool              `json:"enable_stdout" bson:"enableStdout"`
}

// Expression expression
type Expression struct {
	Key      string `json:"key" bson:"key"`
	Operator string `json:"operator" bson:"operator"`
	Value    string `json:"value" bson:"value"`
}

// ToBKlogExpression transfer to bklog expression
func (e *Expression) ToBKlogExpression() bklog.Expression {
	return bklog.Expression{
		Key:      e.Key,
		Operator: e.Operator,
		Value:    e.Value,
	}
}

// GetDataEncoding get data encoding
func (s *LogCollectorConfigSelector) GetDataEncoding() string {
	if s.DataEncoding == "" {
		return DefaultDataEncoding
	}
	return s.DataEncoding
}

// GetLabels get labels
func (s *LogCollectorConfigSelector) GetLabels() []bklog.Label {
	labels := make([]bklog.Label, 0)
	for k, v := range s.MatchLabels {
		labels = append(labels, bklog.Label{
			Key:   k,
			Value: v,
		})
	}
	return labels
}

// ToConfig to bklog config
func (s *LogCollectorConfigSelector) ToConfig(namespace string) []bklog.LogCollectorConfig {
	configs := make([]bklog.LogCollectorConfig, 0)
	if s == nil {
		return configs
	}
	cfg := bklog.LogCollectorConfig{
		Namespaces: []string{namespace},
		LabelSelector: &bklog.LogCollectorConfigSelector{
			MatchLabels:      s.GetLabels(),
			MatchExpressions: make([]bklog.Expression, 0),
		},
		Paths:        s.Paths,
		DataEncoding: s.GetDataEncoding(),
		EnableStdout: s.EnableStdout,
	}
	for _, v := range s.MatchExpressions {
		cfg.LabelSelector.MatchExpressions = append(cfg.LabelSelector.MatchExpressions, v.ToBKlogExpression())
	}
	configs = append(configs, cfg)
	return configs
}

// LogCollectorConfigAllContainers log collector all containers
type LogCollectorConfigAllContainers struct {
	Paths        []string `json:"paths" bson:"paths"`
	DataEncoding string   `json:"data_encoding" bson:"dataEncoding"`
	EnableStdout bool     `json:"enable_stdout" bson:"enableStdout"`
}

// ToConfig to bklog config
func (s *LogCollectorConfigAllContainers) ToConfig(namespace string) []bklog.LogCollectorConfig {
	configs := make([]bklog.LogCollectorConfig, 0)
	if s == nil {
		return configs
	}
	conf := bklog.LogCollectorConfig{
		Namespaces:   []string{namespace},
		Paths:        s.Paths,
		DataEncoding: s.GetDataEncoding(),
		EnableStdout: s.EnableStdout,
	}
	configs = append(configs, conf)
	return configs
}

// GetDataEncoding get data encoding
func (s *LogCollectorConfigAllContainers) GetDataEncoding() string {
	if s.DataEncoding == "" {
		return DefaultDataEncoding
	}
	return s.DataEncoding
}

// ConfigSelected xxx
type ConfigSelected string

const (
	// AllContainers all containers
	AllContainers ConfigSelected = "AllContainers"
	// SelectedContainers selected containers
	SelectedContainers ConfigSelected = "SelectedContainers"
	// SelectedLabels selected labels
	SelectedLabels ConfigSelected = "SelectedLabels"

	// DefaultDataEncoding deafult data encoding
	DefaultDataEncoding = "UTF-8"
)

// LogCollectorSortByUpdateTime sort LogCollector by update time
type LogCollectorSortByUpdateTime []*LogCollector

func (l LogCollectorSortByUpdateTime) Len() int { return len(l) }
func (l LogCollectorSortByUpdateTime) Less(i, j int) bool {
	return l[i].UpdatedAt.After(l[j].UpdatedAt.Time)
}
func (l LogCollectorSortByUpdateTime) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
