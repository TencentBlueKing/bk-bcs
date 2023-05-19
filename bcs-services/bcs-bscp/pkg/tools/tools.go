/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package tools

import (
	"strconv"
	"strings"
)

// GetIntList 获取Int列表, 解析BizID时使用
func GetIntList(value string) ([]int, error) {
	items := strings.Split(value, ",")
	result := make([]int, 0, len(items))
	for _, v := range items {
		intValue, err := strconv.Atoi(v)
		if err != nil {
			return []int{}, err
		}
		result = append(result, intValue)
	}
	return result, nil
}

// GetUint32List convert string to uint32 list
func GetUint32List(value string) ([]uint32, error) {
	items := strings.Split(value, ",")
	result := make([]uint32, 0, len(items))
	for _, v := range items {
		intValue, err := strconv.Atoi(v)
		if err != nil {
			return []uint32{}, err
		}
		result = append(result, uint32(intValue))
	}
	return result, nil
}
