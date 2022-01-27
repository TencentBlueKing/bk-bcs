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

package util

import (
	"fmt"
	"strings"
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
		return nil, fmt.Errorf("items's type must one of (string, []string), get %v", t)
	}
}

func getItems(obj map[string]interface{}, paths []string) (interface{}, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("items is empty list")
	}
	ret, exists := obj[paths[0]]
	if !exists {
		return nil, fmt.Errorf("key %s not exist", paths[0])
	}
	if len(paths) == 1 {
		return ret, nil
	} else if subMap, ok := obj[paths[0]].(map[string]interface{}); ok {
		return GetItems(subMap, paths[1:])
	}
	return nil, fmt.Errorf("key %s, val not map[string]interface{} type", paths[0])
}
