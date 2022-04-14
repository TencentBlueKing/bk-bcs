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

package slice_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

func TestStringInSlice(t *testing.T) {
	// 存在，位置在首中尾
	assert.True(t, slice.StringInSlice("str1", []string{"str1", "str2", "str3"}))
	assert.True(t, slice.StringInSlice("str2", []string{"str1", "str2", "str3"}))
	assert.True(t, slice.StringInSlice("str3", []string{"str1", "str2", "str3"}))

	// 不存在的情况
	assert.False(t, slice.StringInSlice("str4", []string{"str1", "str2"}))
	assert.False(t, slice.StringInSlice("str1", []string{}))
	assert.False(t, slice.StringInSlice("", []string{"str1"}))
}

var typeMapList = []interface{}{
	map[string]interface{}{
		"type": "a",
	},
	map[string]interface{}{
		"type": "a",
		"kind": "c",
	},
	map[string]interface{}{
		"type": "b",
		"kind": "d",
	},
	map[string]interface{}{
		"type": 1,
	},
	"k-v",
}

func TestMatchKVInSlice(t *testing.T) {
	// 存在
	assert.True(t, slice.MatchKVInSlice(typeMapList, "type", "a"))
	assert.True(t, slice.MatchKVInSlice(typeMapList, "type", "b"))
	assert.True(t, slice.MatchKVInSlice(typeMapList, "kind", "c"))

	// 不存在的情况
	assert.False(t, slice.MatchKVInSlice(typeMapList, "type", "v"))
	assert.False(t, slice.MatchKVInSlice(typeMapList, "type", "1"))
	assert.False(t, slice.MatchKVInSlice(typeMapList, "kind", "a"))
	assert.False(t, slice.MatchKVInSlice(typeMapList, "k", "v"))
}

func TestFilterMatchKVFormSlice(t *testing.T) {
	mapList := slice.FilterMatchKVFromSlice(typeMapList, "type", "a")
	assert.Equal(t, len(mapList), 2)
	assert.True(t, slice.MatchKVInSlice(mapList, "type", "a"))

	mapList = slice.FilterMatchKVFromSlice(typeMapList, "type", "b")
	assert.Equal(t, len(mapList), 1)
	assert.True(t, slice.MatchKVInSlice(mapList, "type", "b"))

	mapList = slice.FilterMatchKVFromSlice(typeMapList, "kind", "c")
	assert.Equal(t, len(mapList), 1)
	assert.True(t, slice.MatchKVInSlice(mapList, "kind", "c"))
}
