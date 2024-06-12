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

package v1

import (
	"fmt"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Detect 告警检测配置
type Detect struct {
	Algorithm *Algorithm `json:"algorithm" yaml:"algorithm"`               // 告警检测算法
	Nodata    *Nodata    `json:"nodata,omitempty" yaml:"nodata,omitempty"` // 无数据告警配置
	Trigger   string     `json:"trigger" yaml:"trigger"`                   // nolint 触发配置，如 1/5/6表示5个周期内满足1次则告警，连续6个周期内不满足条件则表示恢复
}

// Algorithm 告警检测算法
type Algorithm struct {
	Fatal    []*AlgorithmConfig `json:"fatal,omitempty" yaml:"fatal,omitempty"`
	Remind   []*AlgorithmConfig `json:"remind,omitempty" yaml:"remind,omitempty"`
	Warning  []*AlgorithmConfig `json:"warning,omitempty" yaml:"warning,omitempty"`
	Operator string             `json:"operator,omitempty" yaml:"operator,omitempty"`
}

// AlgorithmConfig 告警检测算法配置
type AlgorithmConfig struct {
	ConfigStr string                `json:"config,omitempty" yaml:"-"`            // 告警检测配置，仅当检测算法为Threshold时使用
	ConfigObj AlgorithmConfigStruct `json:"configObj,omitempty" yaml:"-"`         // 告警检测配置
	Type      string                `json:"type,omitempty" yaml:"type,omitempty"` // nolint 告警检测算法， 如Threshold/ RingRatioAmplitude/ YearRoundRange...
}

// AlgorithmConfigStruct 检测算法详细配置
type AlgorithmConfigStruct struct {
	Ceil          int    `json:"ceil,omitempty" yaml:"ceil,omitempty"`
	CeilInterval  int    `json:"ceil_interval,omitempty" yaml:"ceil_interval,omitempty"`
	Floor         int    `json:"floor,omitempty" yaml:"floor,omitempty"`
	FloorInterval int    `json:"floor_interval,omitempty" yaml:"floor_interval,omitempty"`
	Ratio         int    `json:"ratio,omitempty" yaml:"ratio,omitempty"`
	Shock         int    `json:"shock,omitempty" yaml:"shock,omitempty"`
	Days          int    `json:"days,omitempty" yaml:"days,omitempty"`
	Method        string `json:"method,omitempty" yaml:"method,omitempty"`
	Threshold     int    `json:"threshold,omitempty" yaml:"threshold,omitempty"`
}

type yamlNode struct {
	Config interface{} `yaml:"config"`
	Type   string      `yaml:"type"`
}

// UnmarshalYAML unmarshal from yaml format
func (a *AlgorithmConfig) UnmarshalYAML(node *yaml.Node) error {
	var yNode yamlNode
	err := node.Decode(&yNode)
	if err != nil {
		return err
	}

	a.Type = yNode.Type
	switch v := yNode.Config.(type) {
	case string:
		a.ConfigStr = v
	case map[string]interface{}:
		if ceil, ok := v["ceil"].(int); ok {
			a.ConfigObj.Ceil = ceil
		}
		if ceilInterval, ok := v["ceil_interval"].(int); ok {
			a.ConfigObj.CeilInterval = ceilInterval
		}
		if floor, ok := v["floor"].(int); ok {
			a.ConfigObj.Floor = floor
		}
		if floorInterval, ok := v["floor_interval"].(int); ok {
			a.ConfigObj.FloorInterval = floorInterval
		}
		if ratio, ok := v["ratio"].(int); ok {
			a.ConfigObj.Ratio = ratio
		}
		if shock, ok := v["shock"].(int); ok {
			a.ConfigObj.Shock = shock
		}
		if days, ok := v["days"].(int); ok {
			a.ConfigObj.Days = days
		}
		if threshold, ok := v["threshold"].(int); ok {
			a.ConfigObj.Threshold = threshold
		}
		if method, ok := v["method"].(string); ok {
			a.ConfigObj.Method = method
		}

	}

	return nil
}

// MarshalYAML transfer to yaml format
func (a AlgorithmConfig) MarshalYAML() (interface{}, error) {
	yNode := yamlNode{}
	if a.ConfigStr != "" {
		yNode.Config = a.ConfigStr
	} else {
		yNode.Config = a.ConfigObj
	}
	yNode.Type = a.Type
	return yNode, nil
}

// Nodata 无数据告警配置
type Nodata struct {
	Continuous int    `json:"continuous,omitempty" yaml:"continuous,omitempty"` // 数据连续丢失n个周期时触发告警
	Level      string `json:"level,omitempty" yaml:"level,omitempty"`           // 告警级别, fatal/remind/warning
}

// Notice 告警通知配置
type Notice struct {
	// 分派配置 可用值 only_notice(默认通知) by_rule（基于分派）
	AssignMode  []string       `json:"assign_mode,omitempty" yaml:"assign_mode,omitempty"`
	Signal      []string       `json:"signal,omitempty" yaml:"signal,omitempty"`             // 告警时机
	UserGroups  []string       `json:"user_groups,omitempty" yaml:"user_groups,omitempty"`   // 告警用户组
	NoiseReduce NoiseReduce    `json:"noise_reduce,omitempty" yaml:"noise_reduce,omitempty"` // 降噪配置
	Interval    int            `json:"interval,omitempty" yaml:"interval,omitempty"`         // 通知间隔（分钟），默认120
	Template    NoticeTemplate `json:"template,omitempty" yaml:"template,omitempty"`         // 告警通知模板，非必填，默认使用默认模板 nolint
}

// NoiseReduce 降噪配置
type NoiseReduce struct {
	Enabled       bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`       // 是否开启
	Dimensions    []string `json:"dimensions,omitempty" yaml:"dimensions,omitempty"` // 降噪维度
	AbnormalRatio int      `json:"abnormal_ratio" yaml:"abnormal_ratio,omitempty"`   // 降噪阈值，百分比 1～100
}

// TemplateContent 通知模板
type TemplateContent struct {
	Title   string `json:"title,omitempty" yaml:"title,omitempty"`     // 告警标题，非必填，默认使用默认模板
	Content string `json:"content,omitempty" yaml:"content,omitempty"` // 告警通知模板，非必填，默认使用默认模板
}

// NoticeTemplate 告警通知模板，非必填，默认使用默认模板
type NoticeTemplate struct {
	Abnormal  TemplateContent `json:"abnormal,omitempty" yaml:"abnormal,omitempty"`
	Recovered TemplateContent `json:"recovered,omitempty" yaml:"recovered,omitempty"`
	Closed    TemplateContent `json:"closed,omitempty" yaml:"closed,omitempty"`
}

// Query 指标采集
type Query struct {
	// +kubebuilder:default=bk_monitor
	DataSource string `json:"data_source" yaml:"data_source"` // 数据源
	// +kubebuilder:default=time_series
	DataType     string         `json:"data_type" yaml:"data_type"`                       // 数据类型
	Expression   string         `json:"expression,omitempty" yaml:"expression,omitempty"` // 计算表达式
	QueryConfigs []*QueryConfig `json:"query_configs" yaml:"query_configs"`               // 指标采集配置
}

// QueryConfig 指标采集配置
type QueryConfig struct {
	Alias     string   `json:"alias,omitempty" yaml:"alias,omitempty"`         // 别名
	Interval  int      `json:"interval" yaml:"interval"`                       // 聚合周期
	Method    string   `json:"method" yaml:"method"`                           // 聚合方法
	GroupBy   []string `json:"group_by,omitempty" yaml:"group_by,omitempty"`   // 聚合维度
	Metric    string   `json:"metric" yaml:"metric"`                           // 指标名
	Where     string   `json:"where,omitempty" yaml:"where,omitempty"`         // 匹配规则
	Functions []string `json:"functions,omitempty" yaml:"functions,omitempty"` // 函数
}

// MonitorRuleDetail 告警规则配置
type MonitorRuleDetail struct {
	Name           string   `json:"name,omitempty" yaml:"name,omitempty"`
	Labels         []string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Enabled        *bool    `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	ActiveTime     string   `json:"active_time,omitempty" yaml:"active_time,omitempty"`
	ActiveCalendar []int    `json:"active_calendar,omitempty" yaml:"active_calendar,omitempty"`
	Detect         *Detect  `json:"detect,omitempty" yaml:"detect,omitempty" patchStrategy:"replace" `
	Notice         *Notice  `json:"notice,omitempty" yaml:"notice,omitempty" patchStrategy:"replace"`
	Query          *Query   `json:"query,omitempty" yaml:"query,omitempty" patchStrategy:"replace"`
}

// WhereAdd 追加告警策略匹配规则
func (m *MonitorRuleDetail) WhereAdd(addQuery string) {
	if addQuery != "" {
		for _, query := range m.Query.QueryConfigs {
			if query.Where == "" {
				query.Where = addQuery
			} else {
				query.Where = fmt.Sprintf("%s and %s", query.Where, addQuery)
			}
		}
	}
}

// WhereOr 追加告警策略匹配规则
func (m *MonitorRuleDetail) WhereOr(addQuery string) {
	if addQuery != "" {
		for _, query := range m.Query.QueryConfigs {
			if query.Where == "" {
				query.Where = addQuery
			} else {
				query.Where = fmt.Sprintf("%s or %s", query.Where, addQuery)
			}
		}
	}
}

// IsEnabled return true if ptr is null
func (m *MonitorRuleDetail) IsEnabled() bool {
	if m.Enabled == nil {
		return true
	}

	return *m.Enabled
}

// MonitorRuleSpec defines the desired state of MonitorRule
type MonitorRuleSpec struct {
	BizID    string `json:"bizID" yaml:"-"`
	BizToken string `json:"bizToken,omitempty" yaml:"-"`
	// if true, controller will ignore resource's change
	IgnoreChange bool `json:"ignoreChange,omitempty"`
	// 是否覆盖同名配置，默认为false
	Override bool `json:"override,omitempty"`

	// +kubebuilder:default=AUTO_MERGE
	// +kubebuilder:validation:Enum=AUTO_MERGE;LOCAL_FIRST
	ConflictHandle string `json:"conflictHandle,omitempty"`

	Scenario string               `json:"scenario,omitempty"`
	Rules    []*MonitorRuleDetail `json:"rules" yaml:"rules" patchStrategy:"merge" patchMergeKey:"name"`
}

// MonitorRuleStatus defines the observed state of MonitorRule
type MonitorRuleStatus struct {
	SyncStatus SyncStatus `json:"syncStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.syncStatus.state`

// MonitorRule is the Schema for the monitorrules API
type MonitorRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonitorRuleSpec   `json:"spec,omitempty"`
	Status MonitorRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MonitorRuleList contains a list of MonitorRule
type MonitorRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MonitorRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MonitorRule{}, &MonitorRuleList{})
}
