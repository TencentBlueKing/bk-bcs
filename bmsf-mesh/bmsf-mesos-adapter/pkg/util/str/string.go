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
 *
 */

package str

import "strings"

// ReplaceSpecialCharForAppName replace _ to - for app name
func ReplaceSpecialCharForAppName(str string) string {
	str = strings.Replace(str, "_", "-", -1)
	return str
}

// ReplaceSpecialCharForLabelKey replace special char to "-" for label key
func ReplaceSpecialCharForLabelKey(str string) string {
	str = strings.Replace(str, "@", "-", -1)
	str = strings.Replace(str, "\"", "-", -1)
	str = strings.Replace(str, "'", "-", -1)
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "{", "", -1)
	str = strings.Replace(str, "}", "", -1)
	str = strings.Replace(str, "_", "-", -1)
	if strings.HasPrefix(str, "io.tencent.paas") {
		if len(str) > 63 {
			str = str[0:63]
		}
	}
	return str
}

// ReplaceSpecialCharForLabelValue replace special char to "-" for label value
func ReplaceSpecialCharForLabelValue(str string) string {
	str = strings.Replace(str, "@", "-", -1)
	str = strings.Replace(str, "/", "-", -1)
	str = strings.Replace(str, "\"", "-", -1)
	str = strings.Replace(str, "\\", "-", -1)
	str = strings.Replace(str, "'", "-", -1)
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "{", "", -1)
	str = strings.Replace(str, "}", "", -1)
	if len(str) > 63 {
		str = str[0:63]
	}
	return str
}

// ReplaceSpecialCharForLabel replace special char for label
func ReplaceSpecialCharForLabel(ss map[string]string) map[string]string {
	ret := make(map[string]string)
	for key, value := range ss {
		newKey := ReplaceSpecialCharForLabelKey(key)
		newValue := ReplaceSpecialCharForLabelValue(value)
		ret[newKey] = newValue
	}
	return ret
}
