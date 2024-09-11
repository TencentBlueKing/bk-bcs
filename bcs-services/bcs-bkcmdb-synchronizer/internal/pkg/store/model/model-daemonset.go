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

// DaemonSet define the daemonSet struct.
type DaemonSet struct {
	WorkloadBase          `json:",inline" bson:",inline"`
	Labels                MapStringString             `json:"labels,omitempty" gorm:"column:labels;type:json"`
	Selector              LabelSelector               `json:"selector,omitempty" gorm:"column:selector;type:json"`
	Replicas              int64                       `json:"replicas,omitempty" gorm:"column:replicas"`
	MinReadySeconds       int64                       `json:"min_ready_seconds,omitempty" gorm:"column:min_ready_seconds"`
	StrategyType          DaemonSetUpdateStrategyType `json:"strategy_type,omitempty" gorm:"column:strategy_type"`
	RollingUpdateStrategy *RollingUpdateDaemonSet     `json:"rolling_update_strategy,omitempty" gorm:"column:rolling_update_strategy"` //nolint
}

// DaemonSetUpdateStrategyType is a strategy according to which a daemon set gets updated.
type DaemonSetUpdateStrategyType string

// RollingUpdateDaemonSet spec to control the desired behavior of rolling update.
type RollingUpdateDaemonSet struct {
	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxSurge is 0.
	MaxUnavailable *IntOrString `json:"max_unavailable" bson:"max_unavailable"`

	// The maximum number of pods that can be scheduled above the desired number of pods.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxUnavailable is 0.
	MaxSurge *IntOrString `json:"max_surge" bson:"max_surge"`
}

// Value RollingUpdateDaemonSet 类型的 Value 方法将其转换为 JSON 格式的 driver.Value
func (r RollingUpdateDaemonSet) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan RollingUpdateDaemonSet 类型的 Scan 方法将 JSON 格式的 driver.Value 反序列化到 RollingUpdateDaemonSet 实例中
func (r *RollingUpdateDaemonSet) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(b, r)
}

// DaemonSetMigrate 函数使用 GORM 自动迁移 DaemonSet 模型到数据库
func DaemonSetMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&DaemonSet{})
}

// TableName DaemonSet 类型的 TableName 方法返回数据库中的表名
func (d DaemonSet) TableName() string {
	return "daemonset"
}
