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

// Package stringx xxx
package stringx

import (
	"errors"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// DefaultCharset 默认字符集（用于生成随机字符串）
const DefaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// LowerCharset 小写字符集（用于生成随机字符串）
const LowerCharset = "abcdefghijklmnopqrstuvwxyz"

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
		// NOCC:gosec/crypto(误报)
		// nolint
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		b[i] = charset[r.Intn(len(charset))]
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

// GetInt64 string转换成int64
func GetInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// TrimStringToRuneCount 裁剪字符串使其不超过指定的字符数（rune count）
func TrimStringToRuneCount(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	trimmed := make([]rune, 0, maxRunes)
	for i, r := range s {
		if i >= maxRunes {
			break
		}
		trimmed = append(trimmed, r)
	}
	return string(trimmed)
}

// ReplaceIllegalChars 将不合法的用户名进行k8s label 规则转换
func ReplaceIllegalChars(username string) string {
	// 只允许 A-Za-z0-9-_.，其他的全部替换成 "_"
	re := regexp.MustCompile(`[^A-Za-z0-9\-_.]`)
	return re.ReplaceAllString(username, "_")
}

// templatePattern 文件夹及文件名称不支持字符：\ / : * ? " < > |
var templatePattern = regexp.MustCompile(`^[^\\/:*?"<>|]+$`)

// ParseTemplateName 解析模板文件夹及模板名称
func ParseTemplateName(name string) (string, string, error) {
	// 校验路径：文件路径不能以 / 开头或结尾，也不能包含连续的 //
	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") || strings.Contains(name, "//") {
		return "", "", errors.New("文件路径不能以 / 开头或结尾，也不能包含连续的 //")
	}
	path := strings.Split(name, "/")
	if len(path) < 2 {
		return "", "", errors.New("无效的模板文件路径名称")
	}

	// 校验文件夹及文件名称是否符合要求
	for i := 0; i < len(path); i++ {
		// 去除右边包含空格，点号(.)
		n := strings.TrimRight(path[i], " .")
		if n == "" {
			return "", "", errors.New(`文件夹或者文件名称不能为空`)
		}
		if !templatePattern.MatchString(n) {
			return "", "", errors.New(`文件夹或者文件名称不能包含字符：\/:*?"<>|`)
		}
		path[i] = n
	}
	// 需要把父文件夹和子文件夹及文件名称分开
	subTemplateaName := strings.Join(path[1:], "/")
	return path[0], subTemplateaName, nil
}
