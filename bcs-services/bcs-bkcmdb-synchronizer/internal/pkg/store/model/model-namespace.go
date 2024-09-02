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

// Namespace define the namespace struct.
type Namespace struct {
	ClusterSpec     `json:",inline" gorm:",inline"`
	ID              int64           `json:"id,omitempty" gorm:"primaryKey"`
	Name            string          `json:"name,omitempty" gorm:"column:name"`
	Labels          MapStringString `json:"labels,omitempty" gorm:"column:labels;type:json"`
	ResourceQuotas  []ResourceQuota `json:"resource_quotas,omitempty" gorm:"column:resource_quotas;type:json"`
	SupplierAccount string          `json:"bk_supplier_account" gorm:"column:bk_supplier_account"`
	// Revision record this app's revision information
	Revision `json:",inline" gorm:",inline"`
}

// ResourceQuota defines the desired hard limits to enforce for Quota.
type ResourceQuota struct {
	Hard          map[string]string    `json:"hard" bson:"hard"`
	Scopes        []ResourceQuotaScope `json:"scopes" bson:"scopes"`
	ScopeSelector *ScopeSelector       `json:"scope_selector" bson:"scope_selector"`
}

// ResourceQuotaScope defines a filter that must match each object tracked by a quota
type ResourceQuotaScope string

// ScopeSelector a scope selector represents the AND of the selectors represented
// by the scoped-resource selector requirements.
type ScopeSelector struct {
	// MatchExpressions a list of scope selector requirements by scope of the resources.
	MatchExpressions []ScopedResourceSelectorRequirement `json:"match_expressions" bson:"match_expressions"`
}

// ScopedResourceSelectorRequirement a scoped-resource selector requirement is a selector that
// contains values, a scope name, and an operator that relates the scope name and values.
type ScopedResourceSelectorRequirement struct {
	// ScopeName The name of the scope that the selector applies to.
	ScopeName ResourceQuotaScope `json:"scope_name" bson:"scope_name"`
	// Represents a scope's relationship to a set of values.
	// Valid operators are In, NotIn, Exists, DoesNotExist.
	Operator ScopeSelectorOperator `json:"operator" bson:"operator"`
	// Values An array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty.
	Values []string `json:"values" bson:"values"`
}

// ScopeSelectorOperator a scope selector operator is the set of operators
// that can be used in a scope selector requirement.
type ScopeSelectorOperator string

// Scan 将数据库中的JSON字段扫描到ResourceQuota结构体中
func (r *ResourceQuota) Scan(value interface{}) error {
	bytes, ok := value.([]byte) // 尝试将value转换为字节切片
	if !ok {                    // 如果转换失败
		return fmt.Errorf("failed to scan JSON field: %v", value) // 返回错误信息
	}
	return json.Unmarshal(bytes, r) // 使用json.Unmarshal将字节切片反序列化为ResourceQuota结构体
}

// Value 将ResourceQuota结构体转换为数据库能存储的driver.Value类型
func (r ResourceQuota) Value() (driver.Value, error) {
	return json.Marshal(r) // 使用json.Marshal将ResourceQuota结构体序列化为JSON格式的字节切片
}

// NamespaceMigrate 自动迁移Namespace模型对应的数据库表结构
func NamespaceMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Namespace{}) // 调用GORM的AutoMigrate方法进行自动迁移
}

// TableName 返回Namespace模型对应的数据库表名
func (n Namespace) TableName() string {
	return "namespace" // 返回表名"namespace"
}
