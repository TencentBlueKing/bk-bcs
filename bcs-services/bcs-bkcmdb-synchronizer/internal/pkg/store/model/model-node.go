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

// Node node structural description.
type Node struct {
	// ID cluster auto-increment ID in cc
	ID int64 `json:"id,omitempty" gorm:"primaryKey"`
	// BizID the business ID to which the cluster belongs
	BizID int64 `json:"bk_biz_id" gorm:"column:bk_biz_id"`
	// SupplierAccount the supplier account that this resource belongs to.
	SupplierAccount string `json:"bk_supplier_account" gorm:"column:bk_supplier_account"`
	// HostID the node ID to which the host belongs
	HostID int64 `json:"bk_host_id,omitempty" gorm:"column:bk_host_id"`
	// ClusterID the node ID to which the cluster belongs
	ClusterID int64 `json:"bk_cluster_id,omitempty" gorm:"column:bk_cluster_id"`
	// ClusterUID the node ID to which the cluster belongs
	ClusterUID string `json:"cluster_uid" gorm:"column:cluster_uid;index"`
	// NodeFields node base fields
	NodeBaseFields `json:",inline" gorm:",inline"`
	// Revision record this app's revision information
	Revision `json:",inline" gorm:",inline"`
}

// NodeBaseFields node's basic attribute field description.
type NodeBaseFields struct {
	// HasPod this field indicates whether there is a pod in the node.
	// if there is a pod, this field is true. If there is no pod, this
	// field is false. this field is false when node is created by default.
	HasPod           bool            `json:"has_pod,omitempty" gorm:"column:has_pod"`
	Name             string          `json:"name,omitempty" gorm:"column:name;index"`
	Roles            string          `json:"roles,omitempty" gorm:"column:roles"`
	Labels           MapStringString `json:"labels,omitempty" gorm:"column:labels;type:json"`
	Taints           MapStringString `json:"taints,omitempty" gorm:"column:taints;type:json"`
	Unschedulable    bool            `json:"unschedulable,omitempty" gorm:"column:unschedulable"`
	InternalIP       StringSlice     `json:"internal_ip,omitempty" gorm:"column:internal_ip;type:json"`
	ExternalIP       StringSlice     `json:"external_ip,omitempty" gorm:"column:external_ip;type:json"`
	HostName         string          `json:"hostname,omitempty" gorm:"column:hostname"`
	RuntimeComponent string          `json:"runtime_component,omitempty" gorm:"column:runtime_component"`
	KubeProxyMode    string          `json:"kube_proxy_mode,omitempty" gorm:"column:kube_proxy_mode"`
	PodCidr          string          `json:"pod_cidr,omitempty" gorm:"columnpod_cidr"`
}

// NodeMigrate 函数用于自动迁移 Node 结构体到数据库中对应的表
// db 是一个 *gorm.DB 实例，用于执行数据库操作
func NodeMigrate(db *gorm.DB) error {
	// AutoMigrate 方法会自动创建或更新数据库表结构以匹配 Node 结构体
	return db.AutoMigrate(&Node{})
}

// TableName 方法定义了 Node 结构体对应的数据库表名
// 返回值 "node" 即为表名
func (n Node) TableName() string {
	return "node"
}
