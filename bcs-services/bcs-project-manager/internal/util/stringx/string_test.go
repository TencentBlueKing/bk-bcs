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
