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

package slice

// StringInSlice 判断字符串是否存在 Slice 中
func StringInSlice(str string, list []string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}
	return false
}

// MatchKVInSlice 对 MapList 中每项进行检查，
// 若存在某项的 key 的 value 为指定值，则返回 true
func MatchKVInSlice(list []interface{}, key, value string) bool {
	for _, item := range list {
		if it, ok := item.(map[string]interface{}); ok {
			if v, ok := it[key]; ok {
				if s, ok := v.(string); ok && s == value {
					return true
				}
			}
		}
	}
	return false
}

// FilterMatchKVFromSlice 对 MapList 中每项进行检查，
// 若存在某项的 key 的 value 为指定值，添加到返回的列表中
func FilterMatchKVFromSlice(list []interface{}, key, value string) []interface{} {
	ret := []interface{}{}
	for _, item := range list {
		if it, ok := item.(map[string]interface{}); ok {
			if v, ok := it[key]; ok {
				if s, ok := v.(string); ok && s == value {
					ret = append(ret, it)
				}
			}
		}
	}
	return ret
}
