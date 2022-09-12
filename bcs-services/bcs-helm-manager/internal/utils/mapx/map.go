/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mapx

import (
	"fmt"
	"strings"
)

// GetItems 获取嵌套定义的 Map 值
func GetItems(obj map[interface{}]interface{}, paths interface{}) (interface{}, error) {
	switch t := paths.(type) {
	case string:
		return getItems(obj, strings.Split(paths.(string), "."))
	case []string:
		return getItems(obj, paths.([]string))
	default:
		return nil, fmt.Errorf("paths's type must one of (string, []string), get %v", t)
	}
}

func getItems(obj map[interface{}]interface{}, paths []string) (interface{}, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("paths is empty list")
	}
	ret, exists := obj[paths[0]]
	if !exists {
		return nil, fmt.Errorf("key %s not exist", paths[0])
	}
	if len(paths) == 1 {
		return ret, nil
	} else if subMap, ok := obj[paths[0]].(map[interface{}]interface{}); ok {
		return getItems(subMap, paths[1:])
	}
	return nil, fmt.Errorf("key %s, val not map[string]interface{} type", paths[0])
}

// SetItems 对嵌套 Map 进行赋值
func SetItems(obj map[interface{}]interface{}, paths interface{}, val interface{}) error {
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
		return fmt.Errorf("paths's type must one of (string, []string), get %v", t)
	}
	return nil
}

func setItems(obj map[interface{}]interface{}, paths []string, val interface{}) error {
	if len(paths) == 0 {
		return fmt.Errorf("paths must is empty list")
	}
	if len(paths) == 1 {
		obj[paths[0]] = val
	} else if subMap, ok := obj[paths[0]].(map[interface{}]interface{}); ok {
		return setItems(subMap, paths[1:], val)
	} else {
		return fmt.Errorf("key %s not exists or obj[key] not map[string]interface{} type", paths[0])
	}
	return nil
}
