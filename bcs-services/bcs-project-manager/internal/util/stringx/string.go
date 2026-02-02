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

package stringx

import (
	"strconv"
	"strings"
)

// SplitString 分割字符串, 允许半角逗号、分号及空格
func SplitString(str string) []string {
	str = strings.ReplaceAll(str, ";", ",")
	str = strings.ReplaceAll(str, " ", ",")
	return strings.Split(str, ",")
}

// Partition 从指定分隔符的第一个位置，将字符串分为两段
func Partition(s string, sep string) (string, string) {
	parts := strings.SplitN(s, sep, 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// JoinString xxx
// AddString 拼接字符串
func JoinString(str ...string) string {
	var strList []string
	strList = append(strList, str...)
	return strings.Join(strList, ",")
}

// JoinStringWithQuote 拼接字符串，并且每个字符串用双引号包裹
func JoinStringWithQuote(str ...string) string {
	var strList []string
	for _, s := range str {
		strList = append(strList, "\""+s+"\"")
	}
	return strings.Join(strList, ",")
}

// Errs2String error array to string
func Errs2String(errs []error) string {
	var strList []string
	for _, err := range errs {
		strList = append(strList, err.Error())
	}
	return strings.Join(strList, ",")
}

// StringToUint32 字符串转换成 uint32，如果为空或者转换失败，返回 0
func StringToUint32(str string) uint32 {
	if str == "" {
		return 0
	}
	result, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return uint32(result)
}
