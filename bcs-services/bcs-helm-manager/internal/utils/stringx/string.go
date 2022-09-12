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

import "strings"

// SplitString 分割字符串, 允许半角逗号、分号及空格
func SplitString(str string) []string {
	str = strings.Replace(str, ";", ",", -1)
	str = strings.Replace(str, " ", ",", -1)
	return strings.Split(str, ",")
}

// JoinString 拼接字符串
func JoinString(str ...string) string {
	var strList []string
	strList = append(strList, str...)
	return strings.Join(strList, ",")
}

// JoinStringBySeparator 通过分割符拼装字符串
func JoinStringBySeparator(strList []string, separator string, addSep bool) string {
	// 如果分隔符为空，则以 \n---\n 分割
	if separator == "" {
		separator = "\n---\n"
	}
	// NOTE: 前面追加一个，以便于分割
	joinedStr := strings.Join(strList, separator)
	if addSep {
		return separator + joinedStr
	}
	return joinedStr
}

// Errs2String error array to string
func Errs2String(errs []error) string {
	var strList []string
	for _, err := range errs {
		strList = append(strList, err.Error())
	}
	return strings.Join(strList, ",")
}

// SplitYaml2Array 分割 yaml 为数组 string
// separator 为空时，设置为以 "---\n" 分割
func SplitYaml2Array(yamlStr string, separator string) []string {
	if separator == "" {
		separator = "---\n"
	}
	var splitedStrArr []string
	for _, s := range strings.Split(yamlStr, separator) {
		// 当为空或\n时，忽略
		if s == "" || s == "\n" {
			continue
		}
		splitedStrArr = append(splitedStrArr, s)
	}
	return splitedStrArr
}
