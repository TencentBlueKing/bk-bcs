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
	"fmt"
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
		actual := parseTemplateFileVar(c.input, "")
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

func TestPruneResource(t *testing.T) {
	m := map[string]interface{}{
		"kind":       "Deployment",
		"apiVersion": "apps/v1",
		"metadata": map[string]interface{}{
			"name":       "template-test",
			"generation": 2,
			"labels": map[string]interface{}{
				"app": "template-test",
			},
			"annotations": map[string]interface{}{
				"analysis.crane.io/resource-recommendation": "containers:\n- containerName: main\n  target:\n    " +
					"cpu: 114m\n    memory: \"120586239\"\n",
				"deployment.kubernetes.io/revision": "1",
				"io.tencent.bcs.editFormat":         "form",
				"io.tencent.paas.creator":           "lxf",
				"io.tencent.paas.source_type":       "template",
				"io.tencent.paas.template_name":     "bcs-service",
				"io.tencent.paas.template_version":  "0.0.1",
				"io.tencent.paas.updator":           "lxf",
			},
			"namespace":         "bcs-system",
			"uid":               "b5417ae4-62bf-4d98-96c6-d2d888393ade",
			"resourceVersion":   "120939293",
			"creationTimestamp": "2024-07-05T10:01:06Z",
			"managedFields": []map[string]interface{}{
				{
					"manager":    "__debug_bin534202415",
					"operation":  "Update",
					"apiVersion": "apps/v1",
					"time":       "2024-07-05T10:01:06Z",
					"fieldsType": "FieldsV1",
				},
			},
		},
	}
	result := pruneResource(m)
	fmt.Println(result)
}
