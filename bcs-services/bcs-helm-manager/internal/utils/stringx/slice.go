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

package stringx

// RemoveDuplicateValues 删除重复的值
func RemoveDuplicateValues(strSlice []string) []string {
	keys := make(map[string]bool)
	result := []string{}
	for _, s := range strSlice {
		if _, v := keys[s]; !v {
			keys[s] = true
			result = append(result, s)
		}
	}
	return result
}

// StringInSlice 判断字符串是否存在 Slice 中
func StringInSlice(str string, list []string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}
	return false
}
