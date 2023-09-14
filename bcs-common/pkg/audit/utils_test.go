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
 *
 */

package audit

import (
	"reflect"
	"testing"
)

func TestSplitSlice(t *testing.T) {
	testCasesInt := []struct {
		name     string
		input    []int
		length   int
		expected [][]int
	}{
		{
			name:     "Slice length of 5",
			input:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			length:   5,
			expected: [][]int{{1, 2, 3, 4, 5}, {6, 7, 8, 9, 10}},
		},
		{
			name:     "Slice length of 3",
			input:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			length:   3,
			expected: [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}},
		},
		{
			name:     "Slice length greater than input",
			input:    []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			length:   20,
			expected: [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		},
		{
			name:     "zero",
			input:    []int{},
			length:   1,
			expected: nil,
		},
		{
			name:     "nil",
			input:    nil,
			length:   1,
			expected: nil,
		},
	}
	testCasesString := []struct {
		name     string
		input    []string
		length   int
		expected [][]string
	}{
		{
			name:     "Slice strings length of 3",
			input:    []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			length:   3,
			expected: [][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7", "8", "9"}, {"10"}},
		},
	}

	for _, tc := range testCasesInt {
		t.Run(tc.name, func(t *testing.T) {
			result := SplitSlice(tc.input, tc.length)

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
	for _, tc := range testCasesString {
		t.Run(tc.name, func(t *testing.T) {
			result := SplitSlice(tc.input, tc.length)

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, but got %v", tc.expected, result)
			}
		})
	}
}
