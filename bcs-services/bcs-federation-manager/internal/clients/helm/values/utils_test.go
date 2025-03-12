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

// Package values xxx
package values

import (
	"testing"
)

func TestMergeValues(t *testing.T) {
	// case1
	values1 := []string{
		`
a: 1
b: 2
c:
  d: 3
`,
		`
b: 4
c:
  e: 5
`,
	}

	expectedResult1 := `a: 1
b: 4
c:
  d: 3
  e: 5
`

	mergedValues1, err := MergeValues(values1...)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(mergedValues1) != 1 {
		t.Fatalf("Expected merged values length to be 1, got: %d", len(mergedValues1))
	}

	if mergedValues1[0] != expectedResult1 {
		t.Fatalf("Expected merged values to be:\n%s\nBut got:\n%s", expectedResult1, mergedValues1[0])
	}

	// case2
	values2 := []string{
		`
a: 1
b: 2
c:
  d: 
    - 1
    - 2
`,
		`
b: 4
c:
  d: 
    - 3
    - 4
  e: 5
`,
	}

	expectedResult2 := `a: 1
b: 4
c:
  d:
  - 3
  - 4
  e: 5
`

	mergedValues2, err := MergeValues(values2...)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(mergedValues2) != 1 {
		t.Fatalf("Expected merged values length to be 1, got: %d", len(mergedValues2))
	}

	if mergedValues2[0] != expectedResult2 {
		t.Fatalf("Expected merged values to be:\n%s\nBut got:\n%s", expectedResult2, mergedValues2[0])
	}
}
