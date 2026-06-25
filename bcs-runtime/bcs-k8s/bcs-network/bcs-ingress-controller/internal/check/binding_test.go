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

package check

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func certBlock(certID, mode, certCaID string) *networkextensionv1.IngressListenerCertificate {
	return &networkextensionv1.IngressListenerCertificate{
		CertID:   certID,
		Mode:     mode,
		CertCaID: certCaID,
	}
}

func TestExpandBindings(t *testing.T) {
	ingresses := []networkextensionv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "https-rule"},
			Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{
						Port:        443,
						Protocol:    constant.ProtocolHTTPS,
						Certificate: certBlock("cert-server", certModeMutual, "cert-ca"),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "unidirectional"},
			Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{
						Port:        443,
						Protocol:    constant.ProtocolHTTPS,
						Certificate: certBlock("cert-only", "UNIDIRECTIONAL", "ignored-ca"),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns1", Name: "sni-route"},
			Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{
						Port:     443,
						Protocol: constant.ProtocolHTTPS,
						Routes: []networkextensionv1.Layer7Route{
							{
								Domain:      "example.com",
								Certificate: certBlock("cert-route", "", ""),
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns2", Name: "port-mapping"},
			Spec: networkextensionv1.IngressSpec{
				PortMappings: []networkextensionv1.IngressPortMapping{
					{
						StartPort:   8443,
						Protocol:    constant.ProtocolTCPSSL,
						Certificate: certBlock("cert-pm", "", ""),
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns3", Name: "tcp-skip"},
			Spec: networkextensionv1.IngressSpec{
				Rules: []networkextensionv1.IngressRule{
					{Port: 80, Protocol: constant.ProtocolTCP, Certificate: certBlock("skip", "", "")},
				},
			},
		},
	}

	bindings := expandBindings(ingresses)
	if len(bindings) != 5 {
		t.Fatalf("expected 5 bindings, got %d", len(bindings))
	}

	found := map[string]CertificateBinding{}
	for _, b := range bindings {
		found[b.BindingKey()] = b
	}

	ruleServer := CertificateBinding{
		OwnerKind: constant.KindIngress, OwnerNamespace: "default", OwnerName: "https-rule", CertID: "cert-server",
		CertRole: CertRoleServer, CertScope: CertScopeRule, Protocol: constant.ProtocolHTTPS,
		Port: "443", Domain: "",
	}
	ruleCA := CertificateBinding{
		OwnerKind: constant.KindIngress, OwnerNamespace: "default", OwnerName: "https-rule", CertID: "cert-ca",
		CertRole: CertRoleClientCA, CertScope: CertScopeRule, Protocol: constant.ProtocolHTTPS,
		Port: "443", Domain: "",
	}
	if _, ok := found[ruleServer.BindingKey()]; !ok {
		t.Fatalf("missing MUTUAL server binding")
	}
	if _, ok := found[ruleCA.BindingKey()]; !ok {
		t.Fatalf("missing MUTUAL client_ca binding")
	}

	uni := CertificateBinding{
		OwnerKind: constant.KindIngress, OwnerNamespace: "default", OwnerName: "unidirectional", CertID: "cert-only",
		CertRole: CertRoleServer, CertScope: CertScopeRule, Protocol: constant.ProtocolHTTPS,
		Port: "443", Domain: "",
	}
	if _, ok := found[uni.BindingKey()]; !ok {
		t.Fatalf("missing UNIDIRECTIONAL server binding")
	}
	for _, b := range bindings {
		if b.OwnerName == "unidirectional" && b.CertRole == CertRoleClientCA {
			t.Fatalf("UNIDIRECTIONAL must not produce client_ca binding")
		}
	}

	route := CertificateBinding{
		OwnerKind: constant.KindIngress, OwnerNamespace: "ns1", OwnerName: "sni-route", CertID: "cert-route",
		CertRole: CertRoleServer, CertScope: CertScopeRoute, Protocol: constant.ProtocolHTTPS,
		Port: "443", Domain: "example.com",
	}
	if got, ok := found[route.BindingKey()]; !ok || got.Domain != "example.com" {
		t.Fatalf("missing SNI route binding with domain")
	}

	pm := CertificateBinding{
		OwnerKind: constant.KindIngress, OwnerNamespace: "ns2", OwnerName: "port-mapping", CertID: "cert-pm",
		CertRole: CertRoleServer, CertScope: CertScopePortMapping, Protocol: constant.ProtocolTCPSSL,
		Port: "8443", Domain: "",
	}
	if _, ok := found[pm.BindingKey()]; !ok {
		t.Fatalf("missing port_mapping binding")
	}

	for _, b := range bindings {
		if b.OwnerName == "tcp-skip" {
			t.Fatalf("non-SSL protocol must not produce bindings")
		}
	}
}

func TestCollectUniqueCertIDs(t *testing.T) {
	bindings := []CertificateBinding{
		{CertID: "a"}, {CertID: "b"}, {CertID: "a"},
	}
	ids := collectUniqueCertIDs(bindings)
	if len(ids) != 2 {
		t.Fatalf("expected 2 unique cert IDs, got %d", len(ids))
	}
}
