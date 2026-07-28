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

package webhookserver

import (
	"strings"
	"testing"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func TestCheckSNINotDisabledOnUpdate(t *testing.T) {
	attrOn := func() *networkextensionv1.IngressListenerAttribute {
		return &networkextensionv1.IngressListenerAttribute{SniSwitch: 1}
	}
	attrOff := func() *networkextensionv1.IngressListenerAttribute {
		return &networkextensionv1.IngressListenerAttribute{SniSwitch: 0}
	}
	ruleHTTPS := func(port int, attr *networkextensionv1.IngressListenerAttribute) networkextensionv1.IngressRule {
		return networkextensionv1.IngressRule{
			Port:              port,
			Protocol:          "HTTPS",
			ListenerAttribute: attr,
		}
	}

	testCases := []struct {
		name    string
		old     *networkextensionv1.Ingress
		new     *networkextensionv1.Ingress
		wantOK  bool
		wantSub string
	}{
		{
			name:   "nil old ingress",
			old:    nil,
			new:    &networkextensionv1.Ingress{Spec: networkextensionv1.IngressSpec{Rules: []networkextensionv1.IngressRule{ruleHTTPS(443, attrOff())}}},
			wantOK: true,
		},
		{
			name: "disable sni on existing port rejected",
			old: &networkextensionv1.Ingress{Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{ruleHTTPS(443, attrOn())},
			}},
			new: &networkextensionv1.Ingress{Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{ruleHTTPS(443, attrOff())},
			}},
			wantOK:  false,
			wantSub: "cannot disable SNI on port 443",
		},
		{
			name: "remove sniSwitch field treated as disable rejected",
			old: &networkextensionv1.Ingress{Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{ruleHTTPS(443, attrOn())},
			}},
			new: &networkextensionv1.Ingress{Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{ruleHTTPS(443, nil)},
			}},
			wantOK:  false,
			wantSub: "cannot disable SNI on port 443",
		},
		{
			name: "keep sni on allowed",
			old: &networkextensionv1.Ingress{Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{ruleHTTPS(443, attrOn())},
			}},
			new: &networkextensionv1.Ingress{Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{ruleHTTPS(443, attrOn())},
			}},
			wantOK: true,
		},
		{
			name: "enable sni on new port allowed",
			old: &networkextensionv1.Ingress{Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{ruleHTTPS(443, attrOff())},
			}},
			new: &networkextensionv1.Ingress{Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{ruleHTTPS(8443, attrOn())},
			}},
			wantOK: true,
		},
	}

	for _, tc := range testCases {
		ok, msg := checkSNINotDisabledOnUpdate(tc.old, tc.new)
		if ok != tc.wantOK {
			t.Errorf("%s: ok = %v, want %v (msg %q)", tc.name, ok, tc.wantOK, msg)
		}
		if !tc.wantOK && tc.wantSub != "" && !strings.Contains(msg, tc.wantSub) {
			t.Errorf("%s: msg %q should contain %q", tc.name, msg, tc.wantSub)
		}
	}
}
