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
	"testing"
)

// TestSetData test set data for customobjectdefinition
func BenchmarkSetData(t *testing.B) {
	sod := NewStorageObjectDefinition()
	data := map[string]interface{}{
		"objectType": "TestType",
		"fields": []map[string]interface{}{
			{
				"name": "field1",
				"type": "fieldtype1",
			},
		},
	}
	for i := 0; i < 1000000; i++ {
		sod.SetData(data)
	}
	t.Logf("%+v", sod)
}

// BenchmarkJSONSetData test json set
func BenchmarkJSONSetData(t *testing.B) {
	sod := NewStorageObjectDefinition()
	data := map[string]interface{}{
		"objectType": "TestType",
		"fields": []map[string]interface{}{
			{
				"name": "field1",
				"type": "fieldtype1",
			},
		},
	}
	for i := 0; i < 1000000; i++ {
		var spec StorageObjectDefinitionSpec
		bytes, _ := json.Marshal(data)
		json.Unmarshal(bytes, &spec)
		sod.Data = spec
	}
	t.Logf("%+v", sod)
}

// TestGetData test get data
func TestGetData(t *testing.T) {
	sod := NewStorageObjectDefinition()
	sod.Data = StorageObjectDefinitionSpec{
		ObjectType: "TestType",
		Fields: []Field{
			{
				Name:     "name1",
				Type:     "int",
				MaxBytes: 64,
			},
		},
	}
	obj := sod.GetData()
	t.Logf("%+v", obj)
	t.Error()
}
