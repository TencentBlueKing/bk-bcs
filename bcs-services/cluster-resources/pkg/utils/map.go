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

package utils

import (
	"errors"
	"fmt"
)

// GetItems 获取嵌套定义的 Map 值
func GetItems(m map[string]interface{}, items []string) (interface{}, error) {
	switch {
	case len(items) == 0:
		return nil, errors.New("items is empty list")
	case len(items) == 1:
		return m[items[0]], nil
	default:
		if subMap, ok := m[items[0]].(map[string]interface{}); ok {
			return GetItems(subMap, items[1:])
		}
		return nil, errors.New(fmt.Sprintf("key %s, val not map[string]interface{} type!", items[0]))
	}
}
