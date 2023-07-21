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

import (
	"math/rand"
	"regexp"
	"strings"
)

// DefaultCharset 默认字符集（用于生成随机字符串）
const DefaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

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

// Partition 从指定分隔符的第一个位置，将字符串分为两段
func Partition(s string, sep string) (string, string) {
	parts := strings.SplitN(s, sep, 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
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

var sep = regexp.MustCompile("(?:^|\\s*\n)---\\s*")

// SplitManifests takes a string of manifest and returns string slice
func SplitManifests(bigFile string) []string {
	res := make([]string, 0)
	bigFileTmp := strings.TrimSpace(bigFile)
	docs := sep.Split(bigFileTmp, -1)
	for _, d := range docs {
		if d == "" {
			continue
		}

		d = strings.TrimSpace(d)
		res = append(res, d)
	}
	return res
}

// Rand 生成随机字符串，若使用默认字符集，则 charset 传入空字符串即可
func Rand(n int, charset string) string {
	if charset == "" {
		charset = DefaultCharset
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
