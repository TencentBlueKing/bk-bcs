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

package mapx

import (
	"reflect"
)

// RemoveZeroSubItem 对 Map 中子项均为零值的项进行清理，返回值表示 Map 是否为空
// 可以考虑支持 options，支持忽略布尔值，空列表等...
func RemoveZeroSubItem(raw map[string]interface{}) bool { // nolint:cyclop
	for key, val := range raw {
		switch v := val.(type) {
		case map[string]interface{}:
			if RemoveZeroSubItem(v) {
				delete(raw, key)
			}
		case []interface{}:
			if len(v) == 0 {
				delete(raw, key)
				continue
			}
			newList := []interface{}{}
			for _, item := range v {
				if it, ok := item.(map[string]interface{}); ok {
					if !RemoveZeroSubItem(it) {
						newList = append(newList, it)
					}
				} else if !isZeroVal(item) {
					newList = append(newList, item)
				}
			}
			raw[key] = newList
		default:
			if isZeroVal(v) {
				delete(raw, key)
			}
		}
	}
	return len(raw) == 0
}

// isZeroVal 检查某个值是否为零值
func isZeroVal(val interface{}) bool {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}
