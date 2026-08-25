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

package tencentcloud

import (
	"testing"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// TestValidateSSLProtocolsSNI test SNI related validation gated by protocol
func TestValidateSSLProtocolsSNI(t *testing.T) {
	cv := NewClbValidater()
	cert := &networkextensionv1.IngressListenerCertificate{Mode: "UNIDIRECTIONAL", CertID: "cert-A"}

	testCases := []struct {
		name   string
		rule   *networkextensionv1.IngressRule
		wantOK bool
	}{
		{
			name: "https sni off with listener cert",
			rule: &networkextensionv1.IngressRule{
				Port: 443, Protocol: ClbProtocolHTTPS, Certificate: cert,
			},
			wantOK: true,
		},
		{
			name: "https sni off without cert rejected",
			rule: &networkextensionv1.IngressRule{
				Port: 443, Protocol: ClbProtocolHTTPS,
			},
			wantOK: false,
		},
		{
			name: "https sni on with route certs",
			rule: &networkextensionv1.IngressRule{
				Port:              443,
				Protocol:          ClbProtocolHTTPS,
				ListenerAttribute: &networkextensionv1.IngressListenerAttribute{SniSwitch: 1},
				Routes: []networkextensionv1.Layer7Route{
					{Domain: "a.example.com", Certificate: cert},
				},
			},
			wantOK: true,
		},
		{
			name: "https sni on without route cert rejected",
			rule: &networkextensionv1.IngressRule{
				Port:              443,
				Protocol:          ClbProtocolHTTPS,
				ListenerAttribute: &networkextensionv1.IngressListenerAttribute{SniSwitch: 1},
				Routes: []networkextensionv1.Layer7Route{
					{Domain: "a.example.com"},
				},
			},
			wantOK: false,
		},
		{
			name: "tcp_ssl sni on rejected",
			rule: &networkextensionv1.IngressRule{
				Port:              443,
				Protocol:          ClbProtocolTCPSSL,
				Certificate:       cert,
				ListenerAttribute: &networkextensionv1.IngressListenerAttribute{SniSwitch: 1},
			},
			wantOK: false,
		},
		{
			name: "tcp_ssl sni off with cert ok",
			rule: &networkextensionv1.IngressRule{
				Port: 443, Protocol: ClbProtocolTCPSSL, Certificate: cert,
			},
			wantOK: true,
		},
	}

	for _, tc := range testCases {
		ok, msg := cv.validateSSLProtocols(tc.rule)
		if ok != tc.wantOK {
			t.Errorf("%s: validateSSLProtocols ok = %v (msg %q), want %v", tc.name, ok, msg, tc.wantOK)
		}
	}
}
