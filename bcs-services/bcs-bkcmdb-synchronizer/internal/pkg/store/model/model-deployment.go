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

// Package model is the model of bcs cmdb
package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// Deployment define the deployment struct.
type Deployment struct {
	WorkloadBase          `json:",inline" gorm:",inline"`
	Labels                MapStringString         `json:"labels,omitempty" gorm:"column:labels;type:json"`
	Selector              LabelSelector           `json:"selector,omitempty" gorm:"column:selector;type:json"`
	Replicas              int64                   `json:"replicas,omitempty" gorm:"column:replicas"`
	MinReadySeconds       int64                   `json:"min_ready_seconds,omitempty" gorm:"column:min_ready_seconds"`
	StrategyType          DeploymentStrategyType  `json:"strategy_type,omitempty" gorm:"column:strategy_type"`
	RollingUpdateStrategy RollingUpdateDeployment `json:"rolling_update_strategy,omitempty" gorm:"column:rolling_update_strategy;type:json"` //nolint
}

// NamespaceSpec describes the common attributes of namespace, it is used by the structure below it.
type NamespaceSpec struct {
	ClusterSpec `json:",inline" gorm:",inline"`

	// NamespaceID namespace id in cc
	NamespaceID int64 `json:"bk_namespace_id,omitempty" gorm:"column:bk_namespace_id"`

	// Namespace namespace name in third party platform
	Namespace string `json:"namespace,omitempty" gorm:"column:namespace;index"`
}

// ClusterSpec describes the common attributes of cluster, it is used by the structure below it.
type ClusterSpec struct {
	// BizID business id in cc
	BizID int64 `json:"bk_biz_id,omitempty" gorm:"column:bk_biz_id"`

	// ClusterID cluster id in cc
	ClusterID int64 `json:"bk_cluster_id,omitempty" gorm:"column:bk_cluster_id"`

	// ClusterUID cluster id in third party platform
	ClusterUID string `json:"cluster_uid,omitempty" gorm:"column:cluster_uid;index"`
}

// LabelSelectorRequirement a label selector requirement is a selector that contains values, a key,
// and an operator that relates the key and values.
type LabelSelectorRequirement struct {
	// key is the label key that the selector applies to.
	Key string `json:"key" bson:"key"`
	// operator represents a key's relationship to a set of values.
	// Valid operators are In, NotIn, Exists and DoesNotExist.
	Operator LabelSelectorOperator `json:"operator" bson:"operator"`
	// Values is an array of string values. If the operator is In or NotIn,
	// values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty.
	Values []string `json:"values" bson:"values"`
}

// LabelSelectorOperator a label selector operator is the set of operators that can be used in a selector requirement.
type LabelSelectorOperator string

// DeploymentStrategyType deployment strategy type
type DeploymentStrategyType string

// RollingUpdateDeployment spec to control the desired behavior of rolling update.
type RollingUpdateDeployment struct {
	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxSurge is 0.
	MaxUnavailable *IntOrString `json:"max_unavailable" bson:"max_unavailable"`

	// The maximum number of pods that can be scheduled above the desired number of pods.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxUnavailable is 0.
	MaxSurge *IntOrString `json:"max_surge" bson:"max_surge"`
}

// Type represents the stored type of IntOrString.
type Type int64

// Value 方法将 RollingUpdateDeployment 结构体转换为 JSON 格式的 []byte，以便存储到数据库中。
func (r RollingUpdateDeployment) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan 方法将从数据库读取的 JSON 格式的 []byte 反序列化为 RollingUpdateDeployment 结构体。
func (r *RollingUpdateDeployment) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(b, r)
}

// DeploymentMigrate 函数使用 GORM 的 AutoMigrate 方法自动迁移 Deployment 结构体对应的数据库表。
func DeploymentMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Deployment{})
}

// TableName 方法指定 Deployment 结构体对应的数据库表名为 "deployment"。
func (d Deployment) TableName() string {
	return "deployment"
}
