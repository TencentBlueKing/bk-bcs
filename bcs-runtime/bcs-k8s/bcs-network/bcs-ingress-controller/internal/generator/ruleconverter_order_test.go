/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package generator

import (
	"testing"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func TestListenerRuleOrder(t *testing.T) {
	routes := []networkextensionv1.Layer7Route{
		{Domain: "z.example.com", Path: "/api/specific/*"},
		{Domain: "z.example.com", Path: "/*"},
		{Domain: "a.example.com", Path: "/configured-last"},
	}

	converter := &RuleConverter{
		rule: &networkextensionv1.IngressRule{Protocol: "HTTPS"},
	}
	rules, err := converter.generateListenerRule(routes)
	if err != nil {
		t.Fatalf("generate listener rules failed: %v", err)
	}
	if len(rules) != len(routes) {
		t.Fatalf("got %d listener rules, want %d", len(rules), len(routes))
	}
	for i := range routes {
		if rules[i].Domain != routes[i].Domain || rules[i].Path != routes[i].Path {
			t.Errorf("rule %d changed order: got %s%s, want %s%s",
				i, rules[i].Domain, rules[i].Path, routes[i].Domain, routes[i].Path)
		}
	}
}
