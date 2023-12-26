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

package slice_test

import (
	"fmt"
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

func TestAllInt64Equal(t *testing.T) {
	// 所有元素都相等
	assert.True(t, slice.AllInt64Equal([]int64{}))
	assert.True(t, slice.AllInt64Equal([]int64{1}))
	assert.True(t, slice.AllInt64Equal([]int64{2, 2}))
	assert.True(t, slice.AllInt64Equal([]int64{0, 0, 0}))

	// 存在不相等
	assert.False(t, slice.AllInt64Equal([]int64{1, 2}))
	assert.False(t, slice.AllInt64Equal([]int64{0, 1, 1}))
	assert.False(t, slice.AllInt64Equal([]int64{0, 0, 2, 2}))
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

func TestRemoveDuplicates(t *testing.T) {
	var tests = []struct {
		s    []string
		want []string
	}{
		{[]string{"apple", "banana", "apple", "orange", "banana", "orange", "apple"}, []string{"apple", "banana", "orange"}},
		{[]string{"a", "a", "b", "b", "c", "c"}, []string{"a", "b", "c"}},
		{[]string{}, []string{}},
		{[]string{"a"}, []string{"a"}},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%v", tt.s)
		t.Run(testname, func(t *testing.T) {
			got := slice.RemoveDuplicateValues(tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}
