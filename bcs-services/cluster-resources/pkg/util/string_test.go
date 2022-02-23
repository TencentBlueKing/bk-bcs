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

package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

func TestSplitString(t *testing.T) {
	// 空字符串的情况
	ret := util.SplitString("")
	assert.Equal(t, []string{""}, ret)

	// 正常情况，分隔符为 ","
	ret = util.SplitString("str1,str2,str3")
	assert.Equal(t, []string{"str1", "str2", "str3"}, ret)

	// 正常情况，分隔符为 ";"
	ret = util.SplitString("str4;str5;str6")
	assert.Equal(t, []string{"str4", "str5", "str6"}, ret)

	// 混合分隔符的情况
	ret = util.SplitString("str7;str8,str9 str0")
	assert.Equal(t, []string{"str7", "str8", "str9", "str0"}, ret)
}

func TestPartition(t *testing.T) {
	pre, post := util.Partition("key=val", "=")
	assert.Equal(t, []string{"key", "val"}, []string{pre, post})

	pre, post = util.Partition("key=val=", "=")
	assert.Equal(t, []string{"key", "val="}, []string{pre, post})

	pre, post = util.Partition("key=", "=")
	assert.Equal(t, []string{"key", ""}, []string{pre, post})

	pre, post = util.Partition("key", "=")
	assert.Equal(t, []string{"key", ""}, []string{pre, post})

	pre, post = util.Partition("key^val", "=")
	assert.Equal(t, []string{"key^val", ""}, []string{pre, post})

	pre, post = util.Partition("key^val", "^")
	assert.Equal(t, []string{"key", "val"}, []string{pre, post})

	pre, post = util.Partition("key^^val", "^")
	assert.Equal(t, []string{"key", "^val"}, []string{pre, post})
}

func TestDecapitalize(t *testing.T) {
	assert.Equal(t, "pod", util.Decapitalize("Pod"))
	assert.Equal(t, "aClaim", util.Decapitalize("AClaim"))
	assert.Equal(t, "deploySpec", util.Decapitalize("deploySpec"))
	assert.Equal(t, "status", util.Decapitalize("status"))
}

func TestGenRandStr(t *testing.T) {
	assert.Equal(t, 10, len(util.GenRandStr(10, "")))
	assert.Equal(t, 15, len(util.GenRandStr(15, "abcd1234")))
	assert.Equal(t, "aaa", util.GenRandStr(3, "a"))
}
