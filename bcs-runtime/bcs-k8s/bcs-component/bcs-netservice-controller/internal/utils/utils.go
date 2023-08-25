/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

// StringInSlice split string to slice
func StringInSlice(strs []string, str string) bool {
	for _, item := range strs {
		if str == item {
			return true
		}
	}
	return false
}

// RemoveStringInSlice remove string from slice
func RemoveStringInSlice(strs []string, str string) []string {
	var newSlice []string
	for _, s := range strs {
		if s != str {
			newSlice = append(newSlice, s)
		}
	}
	return newSlice
}
