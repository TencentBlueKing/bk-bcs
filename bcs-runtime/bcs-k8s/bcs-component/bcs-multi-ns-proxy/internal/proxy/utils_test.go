/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package proxy

import (
	"net/http"
	"testing"
)

// TestGetNamespaceFromRequest test function getNamespaceFromRequest
func TestGetNamespaceFromRequest(t *testing.T) {
	testCases := []struct {
		method            string
		url               string
		expectedNamespace string
	}{
		// resource paths
		{"GET", "/api/v1/namespaces", ""},
		{"GET", "/api/v1/namespaces/other", "other"},

		{"GET", "/api/v1/namespaces/other/pods", "other"},
		{"GET", "/api/v1/namespaces/other/pods/foo", "other"},
		{"HEAD", "/api/v1/namespaces/other/pods/foo", "other"},
		{"GET", "/api/v1/pods", ""},

		// special verbs
		{"GET", "/api/v1/proxy/namespaces/other/pods/foo", "other"},
		{"GET", "/api/v1/proxy/namespaces/other/pods/foo/subpath/not/a/subresource", "other"},
		{"GET", "/api/v1/watch/pods", ""},
		{"GET", "/api/v1/pods?watch=true", ""},
		{"GET", "/api/v1/pods?watch=false", ""},
		{"GET", "/api/v1/watch/namespaces/other/pods", "other"},
		{"GET", "/api/v1/namespaces/other/pods?watch=1", "other"},
		{"GET", "/api/v1/namespaces/other/pods?watch=0", "other"},

		// subresource identification
		{"GET", "/api/v1/namespaces/other/pods/foo/status", "other"},
		{"GET", "/api/v1/namespaces/other/pods/foo/proxy/subpath", "other"},
		{"PUT", "/api/v1/namespaces/other/finalize", "other"},
		{"PUT", "/api/v1/namespaces/other/status", "other"},

		// verb identification
		{"PATCH", "/api/v1/namespaces/other/pods/foo", "other"},

		// api group identification
		{"POST", "/apis/extensions/v1/namespaces/other/pods", "other"},

		// api version identification
		{"POST", "/apis/extensions/v1beta3/namespaces/other/pods", "other"},
	}
	for _, test := range testCases {
		req, _ := http.NewRequest(test.method, test.url, nil)
		ns, _ := getNamespaceFromRequest(req)
		if ns != test.expectedNamespace {
			t.Errorf("expected %s but get %s", test.expectedNamespace, ns)
		}
	}
}
