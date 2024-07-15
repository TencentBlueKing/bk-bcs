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

// Package k8s xxx
package k8s

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func printJSON(t *testing.T, obj map[string]interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

// TestDeleteField test function for DeleteField
func TestDeleteField(t *testing.T) {
	tests := []struct {
		before   metav1.Object
		path     []string
		expected metav1.Object
		desc     string
	}{
		{
			before: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"name": "one",
						"labels": map[string]interface{}{
							"app": "my-app",
						},
						"annotations": map[string]interface{}{
							"key-one": "val",
						},
					},
				},
			},
			path: []string{"metadata", "labels", "app"},
			expected: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"name":   "one",
						"labels": map[string]interface{}{},
						"annotations": map[string]interface{}{
							"key-one": "val",
						},
					},
				},
			},
			desc: "remove labels.app",
		},
		{
			before: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"name": "one",
						"labels": map[string]interface{}{
							"app": "my-app",
						},
						"annotations": map[string]interface{}{
							"key-one": map[string]interface{}{
								"subkey-one": "val",
							},
						},
					},
				},
			},
			path: []string{"metadata", "annotations", "key-one", "subkey-one"},
			expected: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"kind": "Pod",
					"metadata": map[string]interface{}{
						"name": "one",
						"labels": map[string]interface{}{
							"app": "my-app",
						},
						"annotations": map[string]interface{}{
							"key-one": map[string]interface{}{},
						},
					},
				},
			},
			desc: "remove nested map key",
		},
		{
			before: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]interface{}{
						"name":      "mysecret",
						"namespace": "default",
					},
					"data": map[string]interface{}{
						"datakey1": "datavalue1",
						"datakey2": "datavalue2",
					},
					"type": "Opaque",
				},
			},
			path: []string{"data"},
			expected: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]interface{}{
						"name":      "mysecret",
						"namespace": "default",
					},
					"type": "Opaque",
				},
			},
			desc: "remove nested data",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Log("Before:")
			printJSON(t, test.before.(*unstructured.Unstructured).Object)

			deleteField(test.before.(*unstructured.Unstructured).Object, test.path)

			t.Log("After:")
			printJSON(t, test.before.(*unstructured.Unstructured).Object)

			afterBytes, _ := json.Marshal(test.before.(*unstructured.Unstructured).Object)
			expectedBytes, _ := json.Marshal(test.expected.(*unstructured.Unstructured).Object)

			if string(afterBytes) != string(expectedBytes) {
				t.Errorf("got:\n%s\nwant:\n%s", string(afterBytes), string(expectedBytes))
			}
		})
	}
}
