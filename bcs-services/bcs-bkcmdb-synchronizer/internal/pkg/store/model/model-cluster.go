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

// Cluster container cluster table structure
type Cluster struct {
	// ID cluster auto-increment ID in cc
	ID int64 `json:"id" gorm:"primaryKey"`
	// BizID the business ID to which the cluster belongs
	BizID int64 `json:"bk_biz_id" gorm:"column:bk_biz_id"`
	// SupplierAccount the supplier account that this resource belongs to.
	SupplierAccount string `json:"bk_supplier_account" gorm:"column:bk_supplier_account"`
	// ClusterFields cluster base fields
	ClusterBaseFields `json:",inline" gorm:",inline"`
	// Revision record this app's revision information
	Revision `json:",inline" gorm:",inline"`
}

// ClusterBaseFields basic description fields for container clusters.
type ClusterBaseFields struct {
	Name string `json:"name" gorm:"column:name"`
	// SchedulingEngine scheduling engines, such as k8s, tke, etc.
	SchedulingEngine string `json:"scheduling_engine" gorm:"column:scheduling_engine"`
	// Uid ID of the cluster itself
	Uid string `json:"uid" gorm:"column:uid"`
	// Xid The underlying cluster ID it depends on
	Xid string `json:"xid" gorm:"column:xid"`
	// Version cluster version
	Version string `json:"version" gorm:"column:version"`
	// NetworkType network type, such as overlay or underlay
	NetworkType string `json:"network_type" gorm:"column:network_type"`
	// Region the region where the cluster is located
	Region string `json:"region" gorm:"column:region"`
	// Vpc vpc network
	Vpc string `json:"vpc" gorm:"column:vpc"`
	// NetWork global routing network address (container overlay network) For example: ["1.1.1.0/21"]
	NetWork StringSlice `json:"network" gorm:"column:network;type:json"`
	// Type cluster network type, e.g. public clusters, private clusters, etc.
	Type string `json:"type" gorm:"column:type"`
}

// ClusterMigrate 使用GORM的AutoMigrate功能自动迁移Cluster模型到数据库
func ClusterMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Cluster{}) // 自动迁移操作，确保数据库表结构与模型匹配
}

// TableName 定义Cluster模型对应的数据库表名
func (d Cluster) TableName() string {
	return "cluster" // 返回表名"cluster"
}
