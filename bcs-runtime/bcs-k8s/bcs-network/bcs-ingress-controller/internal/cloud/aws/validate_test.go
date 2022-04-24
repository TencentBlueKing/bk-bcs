/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package aws

import "testing"

func TestCheckHTTPCodeValues(t *testing.T) {
	var tests = []struct {
		name string
		code string
		want bool
	}{
		{
			name: "valid http code",
			code: "200",
			want: true,
		},
		{
			name: "out of range http code, 199",
			code: "199",
			want: false,
		},
		{
			name: "out of range http code, 500",
			code: "500",
			want: false,
		},
		{
			name: "invalid http code",
			code: "200a",
			want: false,
		},
		{
			name: "valid multiple http code",
			code: "200,201",
			want: true,
		},
		{
			name: "valid multiple http code #2",
			code: "200,201",
			want: true,
		},
		{
			name: "invalid multiple http code #1",
			code: "200,201,a",
			want: false,
		},
		{
			name: "invalid multiple http code #2",
			code: ",200,201",
			want: false,
		},
		{
			name: "invalid multiple http code #3",
			code: "200,201,199",
			want: false,
		},
		{
			name: "valid range http code",
			code: "200-299",
			want: true,
		},
		{
			name: "invalid range http code #1",
			code: "200-a",
			want: false,
		},
		{
			name: "invalid range http code #2",
			code: "200-",
			want: false,
		},
		{
			name: "invalid range http code #3",
			code: "200-299-399",
			want: false,
		},
		{
			name: "invalid range http code #4",
			code: "299-200",
			want: false,
		},
		{
			name: "invalid range http code #5",
			code: "200-299-",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkHTTPCodeValues(tt.code); got != tt.want {
				t.Errorf("checkHTTPCodeValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
