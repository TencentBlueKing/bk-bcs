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

package common

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func TestGetIngressProtocolLayer(t *testing.T) {
	testCases := []struct {
		title string
		spec  networkextensionv1.IngressSpec
		want  string
	}{
		{
			title: "tcp rule uses transport layer",
			spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{Port: 10001, Protocol: constant.ProtocolTCP},
				},
			},
			want: constant.ProtocolLayerTransport,
		},
		{
			title: "udp rule uses transport layer",
			spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{Port: 53, Protocol: constant.ProtocolUDP},
				},
			},
			want: constant.ProtocolLayerTransport,
		},
		{
			title: "http rule uses application layer",
			spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{Port: 80, Protocol: constant.ProtocolHTTP},
				},
			},
			want: constant.ProtocolLayerApplication,
		},
		{
			title: "https rule uses application layer",
			spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{Port: 443, Protocol: constant.ProtocolHTTPS},
				},
			},
			want: constant.ProtocolLayerApplication,
		},
		{
			title: "multiple tcp rules use transport layer",
			spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{Port: 10001, Protocol: constant.ProtocolTCP},
					{Port: 10080, Protocol: constant.ProtocolTCP},
				},
			},
			want: constant.ProtocolLayerTransport,
		},
		{
			title: "tcp and udp rules use transport layer",
			spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{Port: 10001, Protocol: constant.ProtocolTCP},
					{Port: 10002, Protocol: constant.ProtocolUDP},
				},
			},
			want: constant.ProtocolLayerTransport,
		},
		{
			title: "mixed tcp and http rules use default layer",
			spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{Port: 10001, Protocol: constant.ProtocolTCP},
					{Port: 80, Protocol: constant.ProtocolHTTP},
				},
			},
			want: constant.ProtocolLayerDefault,
		},
		{
			title: "tcp port mapping uses transport layer",
			spec: networkextensionv1.IngressSpec{
				PortMappings: []networkextensionv1.IngressPortMapping{
					{StartPort: 20000, Protocol: constant.ProtocolTCP},
				},
			},
			want: constant.ProtocolLayerTransport,
		},
		{
			title: "empty rules without port mappings use application layer",
			spec:  networkextensionv1.IngressSpec{},
			want:  constant.ProtocolLayerApplication,
		},
		{
			title: "gate zone gaas tcp ingress",
			spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{
						Port:     10001,
						Protocol: constant.ProtocolTCP,
						Services: []networkextensionv1.ServiceRoute{
							{
								ServiceName:      "gate-zone-gaas",
								ServiceNamespace: "37614708-208-live-50",
								ServicePort:      10001,
								IsDirectConnect:  true,
							},
						},
					},
				},
			},
			want: constant.ProtocolLayerTransport,
		},
	}

	for i, tc := range testCases {
		ingress := &networkextensionv1.Ingress{Spec: tc.spec}
		got := GetIngressProtocolLayer(ingress)
		if got != tc.want {
			t.Errorf("case %d %s: got %q, want %q", i, tc.title, got, tc.want)
		}
	}
}

func TestInLayer4Protocol(t *testing.T) {
	for _, protocol := range []string{
		constant.ProtocolTCP,
		constant.ProtocolUDP,
		constant.ProtocolTCPSSL,
		constant.ProtocolQUIC,
		"tcp",
	} {
		if !InLayer4Protocol(protocol) {
			t.Errorf("InLayer4Protocol(%q) = false, want true", protocol)
		}
	}
	for _, protocol := range []string{
		constant.ProtocolHTTP,
		constant.ProtocolHTTPS,
		"unknown",
	} {
		if InLayer4Protocol(protocol) {
			t.Errorf("InLayer4Protocol(%q) = true, want false", protocol)
		}
	}
}

func TestInLayer7Protocol(t *testing.T) {
	for _, protocol := range []string{
		constant.ProtocolHTTP,
		constant.ProtocolHTTPS,
		"http",
	} {
		if !InLayer7Protocol(protocol) {
			t.Errorf("InLayer7Protocol(%q) = false, want true", protocol)
		}
	}
	for _, protocol := range []string{
		constant.ProtocolTCP,
		constant.ProtocolUDP,
		"unknown",
	} {
		if InLayer7Protocol(protocol) {
			t.Errorf("InLayer7Protocol(%q) = true, want false", protocol)
		}
	}
}
