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

package main

import (
	"reflect"
	"testing"
)

// TestParseExemptNamespaces checks that comma-separated input is parsed into a set,
// empty / whitespace-only entries are skipped, and empty input returns nil.
func TestParseExemptNamespaces(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want map[string]struct{}
	}{
		{
			name: "empty string returns nil",
			raw:  "",
			want: nil,
		},
		{
			name: "single namespace",
			raw:  "bcs-system",
			want: map[string]struct{}{"bcs-system": {}},
		},
		{
			name: "multiple namespaces",
			raw:  "bcs-system,kube-system,default",
			want: map[string]struct{}{
				"bcs-system":  {},
				"kube-system": {},
				"default":     {},
			},
		},
		{
			name: "trims whitespace around entries",
			raw:  " bcs-system , kube-system ",
			want: map[string]struct{}{
				"bcs-system":  {},
				"kube-system": {},
			},
		},
		{
			name: "skips empty segments",
			raw:  "bcs-system,,kube-system,",
			want: map[string]struct{}{
				"bcs-system":  {},
				"kube-system": {},
			},
		},
		{
			name: "only commas and whitespace returns nil",
			raw:  " , , ",
			want: nil,
		},
		{
			name: "duplicate entries collapse",
			raw:  "bcs-system,bcs-system",
			want: map[string]struct{}{"bcs-system": {}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := parseExemptNamespaces(c.raw)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("parseExemptNamespaces(%q) = %v, want %v", c.raw, got, c.want)
			}
		})
	}
}
