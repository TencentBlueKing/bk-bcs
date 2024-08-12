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

package tools

import (
	"testing"
)

func TestMatchConfigItem(t *testing.T) {
	tests := []struct {
		scope string
		path  string
		name  string
		want  bool
	}{
		{"foo/*", "foo", "bar", true},
		{"foo/*", "foo", "baz", true},
		{"foo/*", "bar", "baz", false},
		{"foo/bar", "foo", "bar", true},
		{"foo/bar", "foo", "baz", false},
		{"*/bar", "foo", "bar", true},
		{"*/foo/bar", "bar", "foo/bar", true},
		{"*/**", "bar", "foo/bar", true},
		{"**", "bar", "foo/bar", true},
	}

	for _, tt := range tests {
		got, err := MatchConfigItem(tt.scope, tt.path, tt.name)
		if err != nil {
			t.Errorf("MatchConfigItem() error = %v", err)
			continue
		}
		if got != tt.want {
			t.Errorf("MatchConfigItem() = %v, want %v, tt:%+v", got, tt.want, tt)
		}
	}
}

func TestMatchAppConfigItem(t *testing.T) {
	tests := []struct {
		scope string
		app   string
		path  string
		name  string
		want  bool
	}{
		{"app1/*/bar", "app1", "foo", "bar", true},
		{"app1/*/baz", "app1", "foo", "bar", false},
		{"app1/foo/*", "app1", "foo", "bar", true},
		{"app2/foo/*", "app1", "foo", "bar", false},
		{"app1/*/bar", "app2", "foo", "bar", false},
		{"app1/*/foo/bar", "app1", "bar", "foo/bar", true},
		{"app1/foo/*/bar", "app1", "foo", "bar/bar", true},
		{"app1/foo/**", "app1", "foo", "bar/bar", true},
		{"app1/**", "app1", "foo", "bar/bar", true},
	}

	for _, tt := range tests {
		got, err := MatchAppConfigItem(tt.scope, tt.app, tt.path, tt.name)
		if err != nil {
			t.Errorf("MatchAppConfigItem() error = %v", err)
			continue
		}
		if got != tt.want {
			t.Errorf("MatchAppConfigItem() = %v, want %v, tt:%+v", got, tt.want, tt)
		}
	}
}
