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

package utils_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/utils"
)

func TestSplitString(t *testing.T) {
	var excepted []string

	// 空字符串的情况
	excepted = []string{""}
	if ret := utils.SplitString(""); !cmp.Equal(excepted, ret) {
		t.Errorf("Excepted: %v, Result：%v", excepted, ret)
	}

	// 正常情况，分隔符为 ","
	excepted = []string{"str1", "str2", "str3"}
	if ret := utils.SplitString("str1,str2,str3"); !cmp.Equal(excepted, ret) {
		t.Errorf("Excepted: %v, Result：%v", excepted, ret)
	}

	// 正常情况，分隔符为 ";"
	excepted = []string{"str4", "str5", "str6"}
	if ret := utils.SplitString("str4;str5;str6"); !cmp.Equal(excepted, ret) {
		t.Errorf("Excepted: %v, Result：%v", excepted, ret)
	}

	// 混合分隔符的情况
	excepted = []string{"str7", "str8", "str9", "str0"}
	if ret := utils.SplitString("str7;str8,str9 str0"); !cmp.Equal(excepted, ret) {
		t.Errorf("Excepted: %v, Result：%v", excepted, ret)
	}
}
