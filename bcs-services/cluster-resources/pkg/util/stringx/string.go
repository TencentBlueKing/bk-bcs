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
	"math/rand"
	"net"
	"strings"
	"unicode"
)

// DefaultCharset 默认字符集（用于生成随机字符串）
const DefaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// Split 分割字符串，支持 " ", ";", "," 分隔符
func Split(originStr string) []string {
	originStr = strings.ReplaceAll(originStr, ";", ",")
	originStr = strings.ReplaceAll(originStr, " ", ",")
	return strings.FieldsFunc(originStr, func(c rune) bool { return c == ',' })
}

// Partition 从指定分隔符的第一个位置，将字符串分为两段
func Partition(s string, sep string) (string, string) {
	parts := strings.SplitN(s, sep, 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// Decapitalize 首字母转小写（暂不考虑去除空白字符）
func Decapitalize(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
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

// ExtractNumberPrefix 提取字符串中的数字前缀
func ExtractNumberPrefix(s string) string {
	for idx, c := range s {
		if !unicode.IsDigit(c) {
			return s[:idx]
		}
	}
	return s
}

// IsIPv4 是否合法的ipv4地址
func IsIPv4(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && !ip.IsUnspecified() && strings.Contains(s, ".")
}

// IsIPv6 是否合法的ipv6地址
func IsIPv6(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && !ip.IsUnspecified() && strings.Contains(s, ":")
}
