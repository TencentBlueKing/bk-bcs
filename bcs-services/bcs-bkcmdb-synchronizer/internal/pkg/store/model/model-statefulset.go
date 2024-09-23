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

// StatefulSet define the statefulSet struct.
type StatefulSet struct {
	WorkloadBase          `json:",inline" bson:",inline"`
	Labels                MapStringString                   `json:"labels,omitempty" gorm:"column:labels;type:json"`
	Selector              LabelSelector                     `json:"selector,omitempty" gorm:"column:selector;type:json"`
	Replicas              int64                             `json:"replicas,omitempty" gorm:"column:replicas"`
	MinReadySeconds       int64                             `json:"min_ready_seconds,omitempty" gorm:"column:min_ready_seconds"` //nolint
	StrategyType          *StatefulSetUpdateStrategyType    `json:"strategy_type,omitempty" gorm:"column:strategy_type"`
	RollingUpdateStrategy *RollingUpdateStatefulSetStrategy `json:"rolling_update_strategy,omitempty" gorm:"column:rolling_update_strategy"` //nolint
}

// StatefulSetUpdateStrategyType is a strategy according to which a StatefulSet gets updated.
type StatefulSetUpdateStrategyType string

// RollingUpdateStatefulSetStrategy spec to control the desired behavior of rolling update.
type RollingUpdateStatefulSetStrategy struct {
	// Partition indicates the ordinal at which the StatefulSet should be partitioned for updates.
	Partition *int32 `json:"partition" bson:"partition"`

	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxSurge is 0.
	MaxUnavailable *IntOrString `json:"max_unavailable" bson:"max_unavailable"`
}

// Value RollingUpdateStatefulSetStrategy 类型的 Value 方法将结构体转换为 JSON 字符串
func (r RollingUpdateStatefulSetStrategy) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan RollingUpdateStatefulSetStrategy 类型的 Scan 方法将 JSON 字符串反序列化为结构体
func (r *RollingUpdateStatefulSetStrategy) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(b, r)
}

// StatefulSetMigrate 函数用于自动迁移 StatefulSet 结构体到数据库
func StatefulSetMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&StatefulSet{})
}

// TableName StatefulSet 结构体的 TableName 方法返回数据库表名
func (d StatefulSet) TableName() string {
	return "statefulset"
}
