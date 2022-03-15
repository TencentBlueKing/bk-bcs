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

package stringx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

func TestSplit(t *testing.T) {
	// 空字符串的情况
	ret := stringx.Split("")
	assert.Equal(t, []string{""}, ret)

	// 正常情况，分隔符为 ","
	ret = stringx.Split("str1,str2,str3")
	assert.Equal(t, []string{"str1", "str2", "str3"}, ret)

	// 正常情况，分隔符为 ";"
	ret = stringx.Split("str4;str5;str6")
	assert.Equal(t, []string{"str4", "str5", "str6"}, ret)

	// 混合分隔符的情况
	ret = stringx.Split("str7;str8,str9 str0")
	assert.Equal(t, []string{"str7", "str8", "str9", "str0"}, ret)
}

func TestPartition(t *testing.T) {
	pre, post := stringx.Partition("key=val", "=")
	assert.Equal(t, []string{"key", "val"}, []string{pre, post})

	pre, post = stringx.Partition("key=val=", "=")
	assert.Equal(t, []string{"key", "val="}, []string{pre, post})

	pre, post = stringx.Partition("key=", "=")
	assert.Equal(t, []string{"key", ""}, []string{pre, post})

	pre, post = stringx.Partition("key", "=")
	assert.Equal(t, []string{"key", ""}, []string{pre, post})

	pre, post = stringx.Partition("key^val", "=")
	assert.Equal(t, []string{"key^val", ""}, []string{pre, post})

	pre, post = stringx.Partition("key^val", "^")
	assert.Equal(t, []string{"key", "val"}, []string{pre, post})

	pre, post = stringx.Partition("key^^val", "^")
	assert.Equal(t, []string{"key", "^val"}, []string{pre, post})
}

func TestDecapitalize(t *testing.T) {
	assert.Equal(t, "pod", stringx.Decapitalize("Pod"))
	assert.Equal(t, "aClaim", stringx.Decapitalize("AClaim"))
	assert.Equal(t, "deploySpec", stringx.Decapitalize("deploySpec"))
	assert.Equal(t, "status", stringx.Decapitalize("status"))
}

func TestRand(t *testing.T) {
	assert.Equal(t, 10, len(stringx.Rand(10, "")))
	assert.Equal(t, 15, len(stringx.Rand(15, "abcd1234")))
	assert.Equal(t, "aaa", stringx.Rand(3, "a"))
}
