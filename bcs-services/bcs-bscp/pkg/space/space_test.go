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

package space

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSpaceMap(t *testing.T) {
	type test struct {
		input   []string
		want    map[string][]string
		wantErr bool
	}

	tests := []test{
		{input: []string{"a__b"}, want: map[string][]string{"a": {"b"}}, wantErr: false},
		{input: []string{"a"}, want: nil, wantErr: true},
		{input: []string{"a__b__c"}, want: map[string][]string{"a": {"b__c"}}, wantErr: true},
	}

	for _, tc := range tests {
		r, err := buildSpaceMap(tc.input)
		if tc.wantErr {
			assert.True(t, err != nil)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tc.want, r)
		}
	}
}
