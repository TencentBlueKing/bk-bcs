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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitString(t *testing.T) {
	// 以逗号分隔
	srcStr := "str,str1"
	splitList := SplitString(srcStr)
	assert.Equal(t, []string{"str", "str1"}, splitList)

	// 以分号分隔
	srcStr = "str,str1"
	splitList = SplitString(srcStr)
	assert.Equal(t, []string{"str", "str1"}, splitList)

	// 以空格分隔
	srcStr = "str str1"
	splitList = SplitString(srcStr)
	assert.Equal(t, []string{"str", "str1"}, splitList)
}

func TestJoinString(t *testing.T) {
	str1, str2 := "str1", "str2"
	joinedStr := JoinString(str1, str2)
	assert.Equal(t, "str1,str2", joinedStr)
}

func TestJoinStringBySeparator(t *testing.T) {
	// 不带分割符
	srcList := []string{"a", "b", "c"}
	joinedStr := JoinStringBySeparator(srcList, "", true)
	assert.Equal(t, "\n---\na\n---\nb\n---\nc", joinedStr)
	// 带有分隔符
	srcList = []string{"a", "b", "c"}
	joinedStr = JoinStringBySeparator(srcList, "yyy", false)
	assert.Equal(t, "ayyybyyyc", joinedStr)
	// 带有分隔符并左侧追加
	srcList = []string{"a", "b", "c"}
	joinedStr = JoinStringBySeparator(srcList, "yyy", true)
	assert.Equal(t, "yyyayyybyyyc", joinedStr)
}

func TestReplaceIllegalChars(t *testing.T) {
	users := map[string]string{
		"admin@123":   "admin_123",
		"admin.123":   "admin.123",
		"admin_123":   "admin_123",
		"admin-123":   "admin-123",
		"a&dmin@123":  "a_dmin_123",
		"a/dmin@123":  "a_dmin_123",
		"a+dmin@123":  "a_dmin_123",
		"ad`min@123":  "ad_min_123",
		"_ad`min@123": "_ad_min_123",
	}

	for key, data := range users {
		user := ReplaceIllegalChars(key)
		assert.Equal(t, data, user)
	}
}
