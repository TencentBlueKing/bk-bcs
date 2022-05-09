/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pbstruct_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
)

var deploySpec = map[string]interface{}{
	"testKey":              "testValue",
	"replicas":             3,
	"revisionHistoryLimit": 10,
	"intKey4SetItem":       8,
	"selector": map[string]interface{}{
		"matchLabels": map[string]interface{}{
			"app": "nginx",
		},
	},
	"strategy": map[string]interface{}{
		"rollingUpdate": map[string]interface{}{
			"maxSurge":       "25%",
			"maxUnavailable": "25%",
		},
		"type": "RollingUpdate",
	},
}

var map4pbStruct = map[string]interface{}{
	"nil":     nil,
	"ready":   true,
	"int":     1,
	"int32":   int32(2),
	"int64":   int64(3),
	"uint":    uint(4),
	"uint32":  uint32(5),
	"uint64":  uint64(6),
	"float32": float32(7),
	"float64": float64(8),
	"bytes":   []byte{},
	"map[string]interface{}": map[string]interface{}{
		"ready": true,
		"int":   1,
		"int32": int32(2),
	},
	"[]interface": []interface{}{"str1", 1, uint(2)},
	"[]map[string]interface{}": []map[string]interface{}{
		{
			"ready":   true,
			"uint64":  uint64(6),
			"float32": float32(7),
		},
	},
	"[]string": []string{"str1", "str2", "str3"},
}

func TestUnstructured2pbStruct(t *testing.T) {
	utd := unstructured.Unstructured{Object: deploySpec}
	pbStruct := pbstruct.Unstructured2pbStruct(&utd)

	assert.Equal(t, "testValue", pbStruct.AsMap()["testKey"])
	// 转换后数字类型都会变成 float64
	assert.Equal(t, float64(3), pbStruct.AsMap()["replicas"])
}

func TestMapSlice2ListValue(t *testing.T) {
	slice := []map[string]interface{}{deploySpec}
	listValue, err := pbstruct.MapSlice2ListValue(slice)
	assert.Nil(t, err)

	spec, ok := listValue.AsSlice()[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "testValue", spec["testKey"])
	// 转换后数字类型都会变成 float64
	assert.Equal(t, float64(3), spec["replicas"])
}

func TestMap2pbStruct(t *testing.T) {
	pbStruct, _ := pbstruct.Map2pbStruct(deploySpec)
	assert.Equal(t, "testValue", pbStruct.AsMap()["testKey"])
	// 转换后数字类型都会变成 float64
	assert.Equal(t, float64(10), pbStruct.AsMap()["revisionHistoryLimit"])

	// 特殊类型的情况
	_, err := pbstruct.Map2pbStruct(map4pbStruct)
	assert.Nil(t, err)

	// 暂不支持的情况（[]int）
	_, err = pbstruct.Map2pbStruct(
		map[string]interface{}{
			"[]int": []int{1, 2, 3},
		},
	)
	assert.NotNil(t, err)
}
