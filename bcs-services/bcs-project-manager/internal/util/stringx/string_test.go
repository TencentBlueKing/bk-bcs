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
	"fmt"
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

func TestErrs2String(t *testing.T) {
	var errs []error
	for i := range []int{0, 1, 2} {
		errs = append(errs, fmt.Errorf("error %v", i))
	}
	assert.Equal(t, Errs2String(errs), "error 0,error 1,error 2")
}

func TestStringToUint32(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint32
	}{
		// 空字符串
		{"空字符串", "", 0},

		// 正常整数
		{"正常整数", "5", 5},
		{"零", "0", 0},
		{"大整数", "12345", 12345},
		{"带前导零", "007", 7},

		// 浮点数字符串
		{"浮点数 5.00", "5.00", 5},
		{"浮点数 0.00", "0.00", 0},
		{"浮点数 5.5 截断", "5.5", 5},
		{"浮点数 5.99 截断", "5.99", 5},
		{"浮点数 100.00", "100.00", 100},

		// 非法字符串
		{"非数字字符串", "abc", 0},
		{"混合字符串", "5abc", 0},
		{"空格字符串", " ", 0},
		{"带空格的数字", " 5 ", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, StringToUint32(tt.input))
		})
	}
}
