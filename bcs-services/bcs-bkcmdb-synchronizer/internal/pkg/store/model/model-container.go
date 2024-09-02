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

// Package model define model for synchronizer
package model

import "gorm.io/gorm"

// Container container details
type Container struct {
	// cc的自增主键
	ID    int64 `gorm:"primaryKey" json:"id,omitempty"`
	PodID int64 `gorm:"column:bk_pod_id;index" json:"bk_pod_id,omitempty"`
	// ClusterID cluster id in cc
	ClusterID           int64 `json:"bk_cluster_id,omitempty" gorm:"column:bk_cluster_id;index"`
	ContainerBaseFields `json:",inline" gorm:",inline"`
	// Revision record this app's revision information
	Revision `json:",inline" gorm:",inline"`
}

// ContainerBaseFields container core details
type ContainerBaseFields struct {
	Name        string `gorm:"column:name" json:"name,omitempty"`
	Image       string `json:"image,omitempty" gorm:"column:image"`
	ContainerID string `gorm:"column:container_uid" json:"container_uid,omitempty"`
}

// ContainerMigrate 函数用于自动迁移数据库表结构以匹配 Container 结构体
func ContainerMigrate(db *gorm.DB) error {
	// 自动迁移，创建或更新表结构以匹配 Container 结构体
	return db.AutoMigrate(&Container{})
}

// TableName 方法定义了 Container 结构体对应的数据库表名
func (c Container) TableName() string {
	// 返回表名为 "container"
	return "container"
}
