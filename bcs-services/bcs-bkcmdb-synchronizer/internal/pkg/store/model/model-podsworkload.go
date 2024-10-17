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

import "gorm.io/gorm"

// PodsWorkload define the pods workload struct.
type PodsWorkload struct {
	WorkloadBase    `json:",inline" gorm:",inline"`
	Labels          MapStringString `json:"labels,omitempty" gorm:"column:labels;type:json"`
	Selector        LabelSelector   `json:"selector,omitempty" gorm:"column:selector;type:json"`
	Replicas        int64           `json:"replicas,omitempty" gorm:"column:replicas"`
	MinReadySeconds int64           `json:"min_ready_seconds,omitempty" gorm:"column:min_ready_seconds"`
}

// PodsWorkloadMigrate 函数用于自动迁移数据库表结构，确保PodsWorkload模型对应的表是最新的。
// 参数db是gorm.DB类型的数据库连接实例。
func PodsWorkloadMigrate(db *gorm.DB) error {
	// 调用AutoMigrate方法自动迁移PodsWorkload模型对应的表结构
	return db.AutoMigrate(&PodsWorkload{})
}

// TableName 方法返回PodsWorkload模型对应的数据库表名。
func (p PodsWorkload) TableName() string {
	// 返回表名"podsworkload"
	return "podsworkload"
}
