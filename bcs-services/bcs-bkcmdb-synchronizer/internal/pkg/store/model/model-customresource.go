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

package model

import (
	"gorm.io/gorm"
)

// CustomResource defines the custom resource struct.
type CustomResource struct {
	WorkloadBase          `json:",inline" gorm:",inline"`
	Labels                MapStringString         `json:"labels,omitempty" gorm:"column:labels;type:json"`
	Selector              LabelSelector           `json:"selector,omitempty" gorm:"column:selector;type:json"`
	Replicas              int64                   `json:"replicas,omitempty" gorm:"column:replicas"`
	MinReadySeconds       int64                   `json:"min_ready_seconds,omitempty" gorm:"column:min_ready_seconds"`
	StrategyType          DeploymentStrategyType  `json:"strategy_type,omitempty" gorm:"column:strategy_type"`
	RollingUpdateStrategy RollingUpdateDeployment `json:"rolling_update_strategy,omitempty" gorm:"column:rolling_update_strategy;type:json"` //nolint
	// CRKind is the kind of custom resource (e.g., "CronJob", "Task", etc.)
	CRKind string `json:"cr_kind,omitempty" gorm:"column:cr_kind;index"`
	// CRApiVersion is the api version of custom resource (e.g., "batch.tkestack.io/v1")
	CRApiVersion string `json:"cr_api_version,omitempty" gorm:"column:cr_api_version;index"`
}

// CustomResourceMigrate function uses GORM's AutoMigrate method to automatically migrate
// the database table corresponding to the CustomResource struct.
func CustomResourceMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&CustomResource{})
}

// TableName method specifies the database table name for CustomResource struct as "customresource".
func (c CustomResource) TableName() string {
	return "customresource"
}
