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
