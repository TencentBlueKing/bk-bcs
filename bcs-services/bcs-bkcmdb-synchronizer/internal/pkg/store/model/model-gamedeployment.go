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

// GameDeployment define the gameDeployment struct.
type GameDeployment struct {
	WorkloadBase          `json:",inline" bson:",inline"`
	Labels                MapStringString                  `json:"labels,omitempty" gorm:"column:labels;type:json"`
	Selector              LabelSelector                    `json:"selector,omitempty" gorm:"column:selector;type:json"`
	Replicas              int64                            `json:"replicas,omitempty" gorm:"column:replicas"`
	MinReadySeconds       int64                            `json:"min_ready_seconds,omitempty" gorm:"column:min_ready_seconds"` //nolint
	StrategyType          GameDeploymentUpdateStrategyType `json:"strategy_type,omitempty" gorm:"column:strategy_type"`
	RollingUpdateStrategy RollingUpdateGameDeployment      `json:"rolling_update_strategy,omitempty" gorm:"column:rolling_update_strategy"` //nolint
}

// GameDeploymentUpdateStrategyType defines strategies for pods in-place update.
type GameDeploymentUpdateStrategyType string

// RollingUpdateGameDeployment gameDeployment update strategy
type RollingUpdateGameDeployment struct {
	// Partition is the desired number of pods in old revisions. It means when partition
	// is set during pods updating, (replicas - partition) number of pods will be updated.
	Partition *int32 `json:"partition" bson:"partition"`

	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxSurge is 0.
	MaxUnavailable *IntOrString `json:"max_unavailable" bson:"max_unavailable"`

	// The maximum number of pods that can be scheduled above the desired number of pods.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// This can not be 0 if MaxUnavailable is 0.
	MaxSurge *IntOrString `json:"max_surge" bson:"max_surge"`
}

// Value 方法将 RollingUpdateGameDeployment 结构体转换为 JSON 格式的 driver.Value
func (r RollingUpdateGameDeployment) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan 方法将 JSON 格式的 driver.Value 反序列化为 RollingUpdateGameDeployment 结构体
func (r *RollingUpdateGameDeployment) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}
	return json.Unmarshal(b, r)
}

// GameDeploymentMigrate 函数用于自动迁移 GameDeployment 结构体对应的数据库表
func GameDeploymentMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&GameDeployment{})
}

// TableName 方法返回 GameDeployment 结构体对应的数据库表名
func (d GameDeployment) TableName() string {
	return "gamedeployment"
}
