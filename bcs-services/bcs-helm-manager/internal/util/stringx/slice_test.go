/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package stringx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveDuplicateValues(t *testing.T) {
	// 零元素
	strSlice := make([]string, 1, 1)
	assert.Equal(t, RemoveDuplicateValues(strSlice), strSlice)

	// 单个元素
	strSlice = []string{"test"}
	assert.Equal(t, RemoveDuplicateValues(strSlice), strSlice)

	// 多元素
	strSlice = []string{"test1", "test2", "test1", "test3", "test2"}
	expectedSlice := []string{"test1", "test2", "test3"}
	assert.Equal(t, RemoveDuplicateValues(strSlice), expectedSlice)
}
