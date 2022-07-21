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

package stringx

import (
	"strings"
)

// SplitString 分割字符串, 允许半角逗号、分号及空格
func SplitString(str string) []string {
	str = strings.Replace(str, ";", ",", -1)
	str = strings.Replace(str, " ", ",", -1)
	return strings.Split(str, ",")
}

// AddString 拼接字符串
func JoinString(str ...string) string {
	var strList []string
	strList = append(strList, str...)
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
