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

package check

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlackListCheck(t *testing.T) {
	testCases := []struct {
		name    string
		req     *RequestCheck
		allowed bool
	}{
		{
			name: "check namespace allowed",
			req: &RequestCheck{
				Namespace: "kube-system",
			},
			allowed: true,
		},
		{
			name: "check namespace not allowed",
			req: &RequestCheck{
				Namespace: "foo",
				Kind:      "Deployment",
			},
			allowed: true,
		},
		{
			name: "check hostNetwork not allowed",
			req: &RequestCheck{
				Namespace: "aa",
				Kind:      "Deployment",
				Object:    []byte(`{"spec":{"template":{"spec":{"hostNetwork":true}}}}`),
			},
			allowed: false,
		},
		{
			name: "check hostNetwork allowed",
			req: &RequestCheck{
				Namespace: "aa",
				Kind:      "Deployment",
				Object:    []byte(`{"spec":{"template":{"spec":{"hostNetwork":false}}}}`),
			},
			allowed: true,
		},
		{
			name: "check hostNetwork allowed",
			req: &RequestCheck{
				Namespace: "aa",
				Kind:      "Deployment",
				Object:    []byte(`{"spec":{"template":{"spec":{"aa":false}}}}`),
			},
			allowed: true,
		},
		{
			name: "check hostPath not allowed",
			req: &RequestCheck{
				Namespace: "aa",
				Kind:      "Deployment",
				Object:    []byte(`{"spec":{"template":{"spec":{"hostNetwork":false,"volumes":[{"name":"localPath","hostPath":{"path":"/data"}}]}}}}`),
			},
			allowed: false,
		},
		{
			name: "check hostPath allowed",
			req: &RequestCheck{
				Namespace: "aa",
				Kind:      "Deployment",
				Object:    []byte(`{"spec":{"template":{"spec":{"hostNetwork":false,"volumes":[{"name":"localPath","hostPath":{"path":"/etc/localtime"}}]}}}}`),
			},
			allowed: true,
		},
	}

	blackList, err := NewBlackList("./test-data/test_match_rule.yaml")
	assert.Nil(t, err)

	for _, testCase := range testCases {
		res, err := blackList.Check(testCase.req)
		assert.Nil(t, err)
		assert.Equal(t, res.Allowed, testCase.allowed)
	}
}
