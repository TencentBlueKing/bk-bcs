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

// Package k8s xxx
package k8s

import (
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Masker mask data for given namespace and data path
type Masker struct {
	Namespace string
	Path      []string
}

// MaskData mask data for given namespace and given data path
func (m *Masker) MaskData(dMeta *unstructured.Unstructured) {
	// empty namespace, then mask all namespaces
	if len(m.Namespace) == 0 || m.Namespace == "" {
		deleteField(dMeta.Object, m.Path)
	}
	// mask data only for given namespace
	if dMeta.GetNamespace() != m.Namespace {
		return
	}
	deleteField(dMeta.Object, m.Path)
}

// support map[string]interface{}['fieldName'], not support []string[index]
func deleteField(obj map[string]interface{}, path []string) {
	if len(path) == 0 {
		return
	}
	if len(path) > 1 {
		key := path[0] // 第一个路径元素设为键
		// 如果此键在对象内部并且是一个 map[string]interface{} 类型，
		if subObj, ok := obj[key].(map[string]interface{}); ok {
			// 递归地在此对象的子对象上调用 deleteField
			deleteField(subObj, path[1:])
		} else if arr, ok := obj[key].([]interface{}); ok {
			// 如果此键在对象内部并且是一个 []interface{} 类型（即数组），
			// 并假设路径的下一个元素是数组的索引
			if i, err := strconv.Atoi(path[1]); err == nil && i >= 0 && i < len(arr) {
				if subObj, ok := arr[i].(map[string]interface{}); ok {
					// 递归地在此对象的子对象上调用 deleteField
					deleteField(subObj, path[2:])
				}
			}
		}
	} else if len(path) == 1 {
		delete(obj, path[0]) // 直接删除数组的元素或对象的字段
	}
}
