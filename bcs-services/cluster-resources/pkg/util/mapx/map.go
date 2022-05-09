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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// GetItems 获取嵌套定义的 Map 值
// paths 参数支持 []string 类型，如 []string{"metadata", "namespace"}
// 或 string 类型（以 '.' 为分隔符），如 "spec.template.spec.containers"
func GetItems(obj map[string]interface{}, paths interface{}) (interface{}, error) {
	switch t := paths.(type) {
	case string:
		return getItems(obj, strings.Split(paths.(string), "."))
	case []string:
		return getItems(obj, paths.([]string))
	default:
		return nil, errorx.New(errcode.General, "paths's type must one of (string, []string), get %v", t)
	}
}

func getItems(obj map[string]interface{}, paths []string) (interface{}, error) {
	if len(paths) == 0 {
		return nil, errorx.New(errcode.General, "paths is empty list")
	}
	ret, exists := obj[paths[0]]
	if !exists {
		return nil, errorx.New(errcode.General, "key %s not exist", paths[0])
	}
	if len(paths) == 1 {
		return ret, nil
	} else if subMap, ok := obj[paths[0]].(map[string]interface{}); ok {
		return getItems(subMap, paths[1:])
	}
	return nil, errorx.New(errcode.General, "key %s, val not map[string]interface{} type", paths[0])
}

// Get 若指定值不存在，则返回默认值
func Get(obj map[string]interface{}, paths interface{}, _default interface{}) interface{} {
	ret, err := GetItems(obj, paths)
	if err != nil {
		return _default
	}
	return ret
}

// GetBool 获取 Bool 类型快捷方法，默认值为 false
func GetBool(obj map[string]interface{}, paths interface{}) bool {
	return Get(obj, paths, false).(bool)
}

// GetInt64 获取 int64 类型快捷方法，默认值为 int64(0)
func GetInt64(obj map[string]interface{}, paths interface{}) int64 {
	return Get(obj, paths, int64(0)).(int64)
}

// GetStr 获取 string 类型快捷方法，默认值为 ""
func GetStr(obj map[string]interface{}, paths interface{}) string {
	return Get(obj, paths, "").(string)
}

// GetList 获取 []interface{} 类型快捷方法，默认值为 []interface{}{}
func GetList(obj map[string]interface{}, paths interface{}) []interface{} {
	return Get(obj, paths, []interface{}{}).([]interface{})
}

// SetItems 对嵌套 Map 进行赋值
// paths 参数支持 []string 类型，如 []string{"metadata", "namespace"}
// 或 string 类型（以 '.' 为分隔符），如 "spec.template.spec.containers"
func SetItems(obj map[string]interface{}, paths interface{}, val interface{}) error {
	// 检查 paths 类型
	switch t := paths.(type) {
	case string:
		if err := setItems(obj, strings.Split(paths.(string), "."), val); err != nil {
			return err
		}
	case []string:
		if err := setItems(obj, paths.([]string), val); err != nil {
			return err
		}
	default:
		return errorx.New(errcode.General, "paths's type must one of (string, []string), get %v", t)
	}
	return nil
}

func setItems(obj map[string]interface{}, paths []string, val interface{}) error {
	if len(paths) == 0 {
		return errorx.New(errcode.General, "paths is empty list")
	}
	if len(paths) == 1 {
		obj[paths[0]] = val
	} else if subMap, ok := obj[paths[0]].(map[string]interface{}); ok {
		return setItems(subMap, paths[1:], val)
	} else {
		return errorx.New(errcode.General, "key %s not exists or obj[key] not map[string]interface{} type", paths[0])
	}
	return nil
}

// RemoveZeroSubItem 对 Map 中子项均为零值的项进行清理，返回值表示 Map 是否为空
// TODO 可以考虑支持 options，支持忽略布尔值，空列表等...
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
