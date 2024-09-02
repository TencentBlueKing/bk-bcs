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
)

// Revision resource revision information.
type Revision struct {
	Creator    string `gorm:"column:creator" json:"creator,omitempty"`
	Modifier   string `gorm:"column:modifier" json:"modifier,omitempty"`
	CreateTime int64  `gorm:"column:create_time" json:"create_time,omitempty"`
	LastTime   int64  `gorm:"column:last_time" json:"last_time,omitempty"`
}

// MapStringString is a custom type for map[string]string
type MapStringString map[string]string

// Value Implement the Value method for Valuer interface
func (m MapStringString) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan Implement the Scan method for Scanner interface
func (m *MapStringString) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSON field: %v", value)
	}
	return json.Unmarshal(bytes, m)
}

// StringSlice is a custom type for []string
type StringSlice []string

// Value Implement the Value method for Valuer interface
func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan Implement the Scan method for Scanner interface
func (s *StringSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSON field: %v", value)
	}
	return json.Unmarshal(bytes, s)
}

// WorkloadBase define the workload common struct, k8s workload attributes are placed in their respective structures,
// except for very public variables, please do not put them in.
type WorkloadBase struct {
	NamespaceSpec   `json:",inline" bson:",inline"`
	ID              int64  `json:"id,omitempty" gorm:"primaryKey"`
	Name            string `json:"name,omitempty" gorm:"column:name;index"`
	SupplierAccount string `json:"bk_supplier_account,omitempty" gorm:"column:bk_supplier_account"`
	// Revision record this app's revision information
	Revision `json:",inline" gorm:",inline"`
}

// LabelSelector a label selector is a label query over a set of resources.
// the result of matchLabels and matchExpressions are ANDed. An empty label
// selector matches all objects. A null label selector matches no objects.
type LabelSelector struct {
	// MatchLabels is a map of {key,value} pairs.
	MatchLabels map[string]string `json:"match_labels" bson:"match_labels"`
	// MatchExpressions is a list of label selector requirements. The requirements are ANDed.
	MatchExpressions []LabelSelectorRequirement `json:"match_expressions" bson:"match_expressions"`
}

// Value 方法将 LabelSelector 对象转换为 JSON 格式的字节数组，以便存储到数据库中。
func (l LabelSelector) Value() (driver.Value, error) {
	return json.Marshal(l) // 将 LabelSelector 对象序列化为 JSON 字节数组
}

// Scan 方法从数据库读取 JSONB 类型的值，并将其反序列化为 LabelSelector 对象。
func (l *LabelSelector) Scan(value interface{}) error {
	b, ok := value.([]byte) // 尝试将 value 转换为字节数组
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value) // 如果转换失败，返回错误
	}

	return json.Unmarshal(b, l) // 将字节数组反序列化为 LabelSelector 对象
}

// IntOrString is a type that can hold an int32 or a string.
type IntOrString struct {
	Type   Type   `json:"type" bson:"type"`
	IntVal int32  `json:"int_val" bson:"int_val"`
	StrVal string `json:"str_val" bson:"str_val"`
}
