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
	"errors"
	"fmt"
)

// GetItems 获取嵌套定义的 Map 值
func GetItems(obj map[string]interface{}, items []string) (interface{}, error) {
	if len(items) == 0 {
		return nil, errors.New("items is empty list")
	}
	ret, exists := obj[items[0]]
	if !exists {
		return nil, errors.New(fmt.Sprintf("key %s not exist", items[0]))
	}
	if len(items) == 1 {
		return ret, nil
	} else if subMap, ok := obj[items[0]].(map[string]interface{}); ok {
		return GetItems(subMap, items[1:])
	}
	return nil, errors.New(fmt.Sprintf("key %s, val not map[string]interface{} type", items[0]))
}
