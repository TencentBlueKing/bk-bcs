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

// Package mapx xxx
package mapx

import (
	"fmt"
	"strings"
)

// ExistsKey 判断 key 是否存在于 map 中
func ExistsKey(obj map[string]interface{}, key string) bool {
	_, ok := obj[key]
	return ok
}

// CleanUpMap 清理 map 中的值
func CleanUpMap(m map[string]interface{}) map[string]interface{} {
	for k, v := range m {
		m[k] = CleanUpMapValue(v)
	}
	return m
}

// CleanUpMapValue 清理 map 中的值
func CleanUpMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case map[interface{}]interface{}:
		newMap := make(map[string]interface{}, len(v))
		for kk, vv := range v {
			newMap[fmt.Sprintf("%v", kk)] = CleanUpMapValue(vv)
		}
		return newMap
	case []interface{}:
		newMap := make([]interface{}, len(v))
		for kk, vv := range v {
			newMap[kk] = CleanUpMapValue(vv)
		}
		return newMap
	default:
		return v
	}
}

// ReplaceMapValue 替换 map 中的值
func ReplaceMapValue(m map[string]interface{}, old, new string) map[string]interface{} {
	for k := range m {
		m[k] = ReplaceMapValueString(m[k], old, new)
	}
	return m
}

// ReplaceMapValueString 替换 map 中的值
func ReplaceMapValueString(v interface{}, old, new string) interface{} {
	switch v := v.(type) {
	case map[interface{}]interface{}:
		newMap := make(map[string]interface{}, len(v))
		for kk, vv := range v {
			newMap[fmt.Sprintf("%v", kk)] = ReplaceMapValueString(vv, old, new)
		}
		return newMap
	case []interface{}:
		newMap := make([]interface{}, len(v))
		for kk, vv := range v {
			newMap[kk] = ReplaceMapValueString(vv, old, new)
		}
		return newMap
	case string:
		return strings.ReplaceAll(v, old, new)
	case map[string]interface{}:
		newMap := make(map[string]interface{}, len(v))
		for kk, vv := range v {
			newMap[fmt.Sprintf("%v", kk)] = ReplaceMapValueString(vv, old, new)
		}
		return newMap
	default:
		return v
	}
}
