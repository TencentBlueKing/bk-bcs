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

// Pod pod structural description.
type Pod struct {
	// cc的自增主键
	ID            int64 `gorm:"primaryKey" json:"id,omitempty"`
	SysSpec       `json:",inline" gorm:",inline"`
	PodBaseFields `json:",inline" gorm:",inline"`
	// Revision record this app's revision information
	Revision `json:",inline" gorm:",inline"`
}

// SysSpec sys spec
type SysSpec struct {
	BizID           int64  `gorm:"column:bk_biz_id" json:"bk_biz_id,omitempty"`
	SupplierAccount string `json:"bk_supplier_account" gorm:"column:bk_supplier_account"`
	ClusterID       int64  `gorm:"column:bk_cluster_id" json:"bk_cluster_id,omitempty"`
	// redundant cluster id
	ClusterUID  string `gorm:"column:cluster_uid;index" json:"cluster_uid,omitempty"`
	NameSpaceID int64  `gorm:"column:bk_namespace_id" json:"bk_namespace_id,omitempty"`
	// redundant namespace names
	NameSpace string `gorm:"column:namespace;index" json:"namespace,omitempty"`
	Workload  Ref    `gorm:"column:ref;type:json" json:"ref,omitempty"`
	HostID    int64  `gorm:"column:bk_host_id" json:"bk_host_id,omitempty"`
	NodeID    int64  `gorm:"column:bk_node_id" json:"bk_node_id,omitempty"`
	// redundant node names
	Node string `gorm:"column:node_name" json:"node_name,omitempty"`
}

// PodBaseFields pod core details
type PodBaseFields struct {
	Name   string          `gorm:"column:name;index" json:"name,omitempty"`
	Labels MapStringString `json:"labels,omitempty"  gorm:"column:labels;type:json"`
}

// Ref 结构体用于表示一个引用，包含类型、名称和ID
type Ref struct {
	Kind string `json:"kind"` // 引用的类型
	// redundant workload names
	Name string `json:"name,omitempty"` // 引用的名称，如果为空则不显示在JSON中
	// ID workload ID in cc
	ID int64 `json:"id,omitempty"` // 引用的ID，如果为空则不显示在JSON中
}

// Scan 方法用于将数据库扫描到的值解析到Ref结构体中
func (r *Ref) Scan(value interface{}) error {
	bytes, ok := value.([]byte) // 尝试将值转换为字节切片
	if !ok {
		return fmt.Errorf("failed to scan JSON field: %v", value) // 转换失败时返回错误
	}
	return json.Unmarshal(bytes, r) // 使用json.Unmarshal解析字节切片到结构体
}

// Value 方法用于将Ref结构体转换为数据库能接受的值
func (r Ref) Value() (driver.Value, error) {
	return json.Marshal(r) // 使用json.Marshal将结构体转换为字节切片
}

// PodMigrate 函数用于自动迁移Pod模型到数据库
func PodMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Pod{}) // 执行自动迁移
}

// TableName 方法定义了Pod模型对应的数据库表名
func (p Pod) TableName() string {
	return "pod" // 返回表名
}
