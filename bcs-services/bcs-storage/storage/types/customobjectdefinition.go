/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"encoding/json"

	"github.com/mitchellh/mapstructure"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

const (
	// ObjectTypeStorageObjectDefinition type name for StorageObjectDefinition
	ObjectTypeStorageObjectDefinition = "StorageObjectDefinition"
)

// FieldType type of field
type FieldType string

// Field field
type Field struct {
	Name     string    `json:"name" bson:"name"`
	Type     FieldType `json:"type" bson:"type"`
	MaxBytes int64     `json:"maxBytes" bson:"maxBytes"`
}

// FieldIndex field index
type FieldIndex struct {
	Keys   []string `json:"keys" bson:"keys"`
	Name   string   `json:"name" bson:"name"`
	Unique bool     `json:"unique" bson:"unique"`
}

// StorageObjectDefinitionSpec spec of StorageObjectDefinition
type StorageObjectDefinitionSpec struct {
	ObjectType string       `json:"objectType" bson:"objectType"`
	Fields     []Field      `json:"fields" bson:"fields"`
	Indexes    []FieldIndex `json:"indexes" bson:"indexes"`
}

// StorageObjectDefinition object definition for storage
type StorageObjectDefinition struct {
	Meta `json:",inline" bson:",inline"`
	Data StorageObjectDefinitionSpec `json:"data" bson:"data"`
}

// NewStorageObjectDefinition create storage object definition
func NewStorageObjectDefinition() *StorageObjectDefinition {
	return &StorageObjectDefinition{
		Meta: Meta{
			Type: ObjectTypeStorageObjectDefinition,
		},
	}
}

// GetData get data
func (sod *StorageObjectDefinition) GetData() map[string]interface{} {
	bytes, _ := json.Marshal(sod.Data)
	ms := make(map[string]interface{})
	err := json.Unmarshal(bytes, &ms)
	if err != nil {
		blog.Errorf("unmarshal %s to map[string]interface{} failed, err %s", string(bytes), err.Error())
	}
	return ms
}

// SetData set data
func (sod *StorageObjectDefinition) SetData(data map[string]interface{}) error {
	var spec StorageObjectDefinitionSpec
	err := mapstructure.Decode(data, &spec)
	if err != nil {
		return err
	}
	sod.Data = spec
	return nil
}
