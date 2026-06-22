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
	"fmt"
	"strconv"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

const (
	// CertRoleServer is the server certificate binding role.
	CertRoleServer = "server"
	// CertRoleClientCA is the mutual-TLS client CA certificate binding role.
	CertRoleClientCA = "client_ca"

	// CertScopeRule is certificate configured on an ingress rule.
	CertScopeRule = "rule"
	// CertScopeRoute is certificate configured on a layer-7 route.
	CertScopeRoute = "route"
	// CertScopePortMapping is certificate configured on a port mapping.
	CertScopePortMapping = "port_mapping"

	certModeMutual = "MUTUAL"
)

// CertificateBinding represents one SSL certificate mount on an owner resource.
type CertificateBinding struct {
	OwnerKind      string
	OwnerNamespace string
	OwnerName      string
	CertID         string
	CertRole       string
	CertScope      string
	Protocol       string
	Port           string
	Domain         string
}

// BindingKey returns the unique key for this binding used in metric cleanup.
func (b CertificateBinding) BindingKey() string {
	return fmt.Sprintf("%s/%s/%s|%s|%s|%s|%s|%s|%s",
		b.OwnerKind, b.OwnerNamespace, b.OwnerName, b.CertID, b.CertRole, b.CertScope, b.Protocol, b.Port, b.Domain)
}

// LabelValues returns the Prometheus label values for this binding.
func (b CertificateBinding) LabelValues() []string {
	return []string{
		b.OwnerKind, b.OwnerNamespace, b.OwnerName, b.CertID, b.CertRole,
		b.CertScope, b.Protocol, b.Port, b.Domain,
	}
}

func isSSLProtocol(protocol string) bool {
	return protocol == constant.ProtocolHTTPS ||
		protocol == constant.ProtocolTCPSSL ||
		protocol == constant.ProtocolQUIC
}

// expandBindings expands SSL certificate bindings from ingress list.
func expandBindings(ingressList []networkextensionv1.Ingress) []CertificateBinding {
	var bindings []CertificateBinding
	for i := range ingressList {
		ing := &ingressList[i]
		bindings = append(bindings, expandIngressBindings(ing)...)
	}
	return bindings
}

func expandIngressBindings(ing *networkextensionv1.Ingress) []CertificateBinding {
	var bindings []CertificateBinding
	ns := ing.GetNamespace()
	name := ing.GetName()

	for _, rule := range ing.Spec.Rules {
		if !isSSLProtocol(rule.Protocol) {
			continue
		}
		port := strconv.Itoa(rule.Port)
		bindings = append(bindings, expandCertBlock(constant.KindIngress, ns, name, rule.Protocol, port, "", CertScopeRule, rule.Certificate)...)
		for _, route := range rule.Routes {
			domain := route.Domain
			bindings = append(bindings, expandCertBlock(constant.KindIngress, ns, name, rule.Protocol, port, domain, CertScopeRoute, route.Certificate)...)
		}
	}

	for _, mapping := range ing.Spec.PortMappings {
		if !isSSLProtocol(mapping.Protocol) {
			continue
		}
		port := strconv.Itoa(mapping.StartPort)
		bindings = append(bindings, expandCertBlock(constant.KindIngress, ns, name, mapping.Protocol, port, "", CertScopePortMapping, mapping.Certificate)...)
	}

	return bindings
}

func expandCertBlock(ownerKind, ns, name, protocol, port, domain, scope string,
	cert *networkextensionv1.IngressListenerCertificate) []CertificateBinding {
	if cert == nil {
		return nil
	}
	var bindings []CertificateBinding
	if cert.CertID != "" {
		bindings = append(bindings, CertificateBinding{
			OwnerKind:      ownerKind,
			OwnerNamespace: ns,
			OwnerName:      name,
			CertID:         cert.CertID,
			CertRole:       CertRoleServer,
			CertScope:      scope,
			Protocol:       protocol,
			Port:           port,
			Domain:         domain,
		})
	}
	if cert.Mode == certModeMutual && cert.CertCaID != "" {
		bindings = append(bindings, CertificateBinding{
			OwnerKind:      ownerKind,
			OwnerNamespace: ns,
			OwnerName:      name,
			CertID:         cert.CertCaID,
			CertRole:       CertRoleClientCA,
			CertScope:      scope,
			Protocol:       protocol,
			Port:           port,
			Domain:         domain,
		})
	}
	return bindings
}

// collectUniqueCertIDs returns deduplicated certificate IDs from bindings.
func collectUniqueCertIDs(bindings []CertificateBinding) []string {
	seen := make(map[string]struct{}, len(bindings))
	var ids []string
	for _, b := range bindings {
		if b.CertID == "" {
			continue
		}
		if _, ok := seen[b.CertID]; ok {
			continue
		}
		seen[b.CertID] = struct{}{}
		ids = append(ids, b.CertID)
	}
	return ids
}
