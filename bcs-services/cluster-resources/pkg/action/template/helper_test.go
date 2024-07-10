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

package template

import (
	"reflect"
	"testing"
)

func TestParseTemplateFileVar(t *testing.T) {
	cases := []struct {
		input  string
		expect []string
	}{
		{
			input:  "",
			expect: []string{},
		},
		{
			input:  "abc",
			expect: []string{},
		},
		{
			input:  "abc{{",
			expect: []string{},
		},
		{
			input:  "abc{{}}",
			expect: []string{},
		},
		{
			input:  "{{var}}",
			expect: []string{"var"},
		},
		{
			input:  "\"{{var}}\"",
			expect: []string{"var"},
		},
		{
			input:  "abc{{var}}",
			expect: []string{"var"},
		},
		{
			input:  "abc{{ var}}",
			expect: []string{"var"},
		},
		{
			input:  "abc{{var }}",
			expect: []string{"var"},
		},
		{
			input:  "abc{{ var }}",
			expect: []string{"var"},
		},
		{
			input:  "abc{{ v ar }}",
			expect: []string{"v ar"},
		},
		{
			input:  "abc{{var}}def",
			expect: []string{"var"},
		},
		{
			input:  "abc{{var}}def{{var2}}",
			expect: []string{"var", "var2"},
		},
	}
	for _, c := range cases {
		actual := parseTemplateFileVar(c.input)
		if !reflect.DeepEqual(actual, c.expect) {
			t.Errorf("input: %s, expect: %v, actual: %v", c.input, c.expect, actual)
		}
	}
}

func TestReplaceTemplateFileVar(t *testing.T) {
	cases := []struct {
		input  string
		values map[string]string
		expect string
	}{
		{
			input:  "",
			values: map[string]string{},
			expect: "",
		},
		{
			input:  "abc",
			values: map[string]string{"var": "val"},
			expect: "abc",
		},
		{
			input:  "abc{{var}}",
			values: map[string]string{"var": "val"},
			expect: "abcval",
		},
		{
			input:  "abc{{var}}",
			values: map[string]string{"var": ""},
			expect: "abc",
		},
		{
			input:  "abc{{var}}",
			values: map[string]string{"var": "val", "var2": "val2"},
			expect: "abcval",
		},
		{
			input:  "abc{{var}}def{{var2}}",
			values: map[string]string{"var": "val", "var2": "val2"},
			expect: "abcvaldefval2",
		},
		{
			input:  "abc{{var}}def{{var2}}",
			values: map[string]string{"var": "val"},
			expect: "abcvaldef",
		},
		{
			input:  "abc{{var}}def{{var2}}",
			values: map[string]string{"var": "val", "var2": ""},
			expect: "abcvaldef",
		},
		{
			input:  "abc{{ var}}def",
			values: map[string]string{"var": "val", "var2": "val2"},
			expect: "abcvaldef",
		},
		{
			input:  "abc{{var }}def",
			values: map[string]string{"var": "val", "var2": "val2"},
			expect: "abcvaldef",
		},
		{
			input:  "abc{{ var }}def",
			values: map[string]string{"var": "val", "var2": "val2"},
			expect: "abcvaldef",
		},
	}
	for _, c := range cases {
		actual := replaceTemplateFileVar(c.input, c.values)
		if actual != c.expect {
			t.Errorf("input: %s, expect: %v, actual: %v", c.input, c.expect, actual)
		}
	}
}
