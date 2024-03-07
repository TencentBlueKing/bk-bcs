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

package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"golang.org/x/exp/slices"
)

// TruncateString 字符串截断
func TruncateString(s string, maxLength, lastLen int) string {
	if len(s) <= maxLength {
		return s
	}
	if lastLen < maxLength {
		prefix := s[0 : maxLength-lastLen]
		afterTruncate := truncateString(s[maxLength-lastLen:], lastLen)
		return prefix + afterTruncate
	}
	return truncateString(s, maxLength)
}

func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}

	h := sha256.New()
	_, _ = io.WriteString(h, s)
	hash := fmt.Sprintf("%x", h.Sum(nil))

	// Ensure the hash fits into the maxLength.
	if maxLength < len(hash) {
		hash = hash[:maxLength]
	}

	// Truncate the original string to make room for the hash.
	s = s[:maxLength-len(hash)]

	return s + hash
}

// ParseGitRepoName 解析 Git 仓库名称
func ParseGitRepoName(repo string) string {
	slice := strings.Split(repo, "/")
	if len(slice) == 0 {
		return ""
	}
	return strings.TrimSuffix(slice[len(slice)-1], ".git")
}

// MapEqualExceptKey 确认 Map 是否一致
func MapEqualExceptKey(m1, m2 map[string]string, excludeKey []string) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v1 := range m1 {
		if slices.Contains[[]string, string](excludeKey, k) {
			continue
		}
		v2, ok := m2[k]
		if !ok || v1 != v2 {
			return false
		}
	}
	return true
}
