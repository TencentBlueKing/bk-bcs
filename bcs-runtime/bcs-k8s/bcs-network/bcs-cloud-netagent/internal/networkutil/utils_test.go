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

package networkutil

import "testing"

// TestCleanEniRtTable test clean eni route table content function
func TestCleanEniRtTable(t *testing.T) {
	tests := []struct {
		before string
		after  string
	}{
		{
			before: `
#
# reserved values
#
255	local
254	main
253	default
0	unspec
#
# local
#
#1	inr.ruhep
100 eni0
101 eni1
`,
			after: `
#
# reserved values
#
255	local
254	main
253	default
0	unspec
#
# local
#
#1	inr.ruhep
`,
		},
		{
			before: `
#
# reserved values
#
255	local
254	main
253	default
0	unspec
#
# local
#
#1	inr.ruhep
`,
			after: `
#
# reserved values
#
255	local
254	main
253	default
0	unspec
#
# local
#
#1	inr.ruhep
`,
		},
	}
	for index, test := range tests {
		tmpAfter := cleanEniRtTable(test.before)
		if tmpAfter != test.after {
			t.Errorf("[test %d]: failed, expect %s, but get %s", index, test.after, tmpAfter)
		}
	}

}
