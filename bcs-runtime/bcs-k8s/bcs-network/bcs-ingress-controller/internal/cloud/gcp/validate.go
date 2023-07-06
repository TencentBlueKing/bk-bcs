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

package gcp

import (
	"fmt"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// GclbValidater validates the gclb parameters
type GclbValidater struct{}

// NewGclbValidater creates a new gclb validater
func NewGclbValidater() *GclbValidater {
	return &GclbValidater{}
}

// IsIngressValid check bcs ingress parameter
func (g *GclbValidater) IsIngressValid(ingress *networkextensionv1.Ingress) (bool, string) {
	if ingress == nil {
		return false, "ingress cannot be empty"
	}
	for i := range ingress.Spec.Rules {
		if ok, msg := g.validateIngressRule(&ingress.Spec.Rules[i]); !ok {
			return ok, msg
		}
	}

	for i := range ingress.Spec.PortMappings {
		if ok, msg := g.validateListenerMapping(&ingress.Spec.PortMappings[i]); !ok {
			return ok, msg
		}
	}
	return true, ""
}

// CheckNoConflictsInIngress return true, if there is no conflicts in ingress itself
func (g *GclbValidater) CheckNoConflictsInIngress(ingress *networkextensionv1.Ingress) (bool, string) {
	ruleMap := make(map[int]networkextensionv1.IngressRule)
	for index, rule := range ingress.Spec.Rules {
		existedRule, ok := ruleMap[rule.Port]
		if !ok {
			ruleMap[rule.Port] = ingress.Spec.Rules[index]
			continue
		}
		return false, fmt.Sprintf("%+v conflicts with %+v", rule, existedRule)
	}

	for i := 0; i < len(ingress.Spec.PortMappings)-1; i++ {
		mapping := ingress.Spec.PortMappings[i]
		for port, rule := range ruleMap {
			if port >= mapping.StartPort+mapping.StartIndex && port < mapping.StartPort+mapping.EndIndex {
				return false, fmt.Sprintf("%+v port conflicts with %+v", mapping, rule)
			}
		}
		for j := i + 1; j < len(ingress.Spec.PortMappings); j++ {
			tmpMapping := ingress.Spec.PortMappings[j]
			if mapping.StartPort+mapping.StartIndex > tmpMapping.StartPort+tmpMapping.EndIndex ||
				mapping.StartPort+mapping.EndIndex < tmpMapping.StartPort+tmpMapping.StartIndex {
				continue
			}
			return false, fmt.Sprintf("%+v ports conflicts with %+v", mapping, tmpMapping)
		}
	}
	return true, ""
}

// validateIngressRule check ingress rule
func (g *GclbValidater) validateIngressRule(rule *networkextensionv1.IngressRule) (bool, string) {
	if rule.Port <= 0 || rule.Port >= 65536 {
		return false, fmt.Sprintf("invalid port %d, available [1-65535]", rule.Port)
	}
	if rule.Protocol != ProtocolHTTP &&
		rule.Protocol != ProtocolHTTPS &&
		rule.Protocol != ProtocolTCP &&
		rule.Protocol != ProtocolUDP {
		return false, fmt.Sprintf("invalid protocol %s, available [http, https, tcp, udp]", rule.Protocol)
	}
	if rule.Protocol == ProtocolHTTP && (rule.Port != 80 && rule.Port != 8080) {
		return false, fmt.Sprintf("invalid port %d for protocol %s, available [80, 8080]", rule.Port, rule.Protocol)
	}
	if rule.Protocol == ProtocolHTTPS {
		if rule.Port != 443 {
			return false, "https protocol only support 443 port"
		}
		if rule.Certificate == nil {
			return false, "certificate cannot be empty for protocol https"
		}
		if ok, msg := g.validateCertificate(rule.Certificate); !ok {
			return ok, msg
		}
	}

	switch rule.Protocol {
	case ProtocolHTTP, ProtocolHTTPS:
		if rule.ListenerAttribute != nil {
			if ok, msg := g.validateApplicationListenerAttribute(rule.ListenerAttribute); !ok {
				return ok, msg
			}
		}
		for i := range rule.Routes {
			if ok, msg := g.validateListenerRoute(&rule.Routes[i]); !ok {
				return ok, msg
			}
		}
	case ProtocolTCP, ProtocolUDP:
		if rule.ListenerAttribute != nil {
			if ok, msg := g.validateNetworkListenerAttribute(rule.ListenerAttribute); !ok {
				return ok, msg
			}
		}
		for i := range rule.Services {
			if ok, msg := g.validateListenerService(&rule.Services[i]); !ok {
				return ok, msg
			}
		}
	}
	return true, ""
}

// validateAppListenerAttribute check application lb listener attribute
func (g *GclbValidater) validateApplicationListenerAttribute(attr *networkextensionv1.IngressListenerAttribute) (bool, string) {
	// validate health check
	if attr.HealthCheck == nil {
		return true, ""
	}
	if attr.HealthCheck.HealthNum != 0 && (attr.HealthCheck.HealthNum < 1 || attr.HealthCheck.HealthNum > 10) {
		return false, fmt.Sprintf("invalid healthNum %d, available [1, 10]", attr.HealthCheck.HealthNum)
	}
	if attr.HealthCheck.UnHealthNum != 0 &&
		(attr.HealthCheck.UnHealthNum < 1 || attr.HealthCheck.UnHealthNum > 10) {
		return false, fmt.Sprintf("invalid unHealthNum %d, available [1, 10]", attr.HealthCheck.UnHealthNum)
	}
	if attr.HealthCheck.IntervalTime != 0 &&
		(attr.HealthCheck.IntervalTime < 1 || attr.HealthCheck.IntervalTime > 300) {
		return false, fmt.Sprintf("invalid interval time %d, available [1, 300]", attr.HealthCheck.IntervalTime)
	}
	if attr.HealthCheck.Timeout != 0 && attr.HealthCheck.Timeout < 1 {
		return false, fmt.Sprintf("invalid timeout %d, timeout must be an integer bigger than or equal to 1", attr.HealthCheck.Timeout)
	}
	if attr.HealthCheck.Timeout != 0 && attr.HealthCheck.Timeout > attr.HealthCheck.IntervalTime {
		return false, fmt.Sprintf("invalid timeout %d, timeout must be lower than or equal to the interval", attr.HealthCheck.Timeout)
	}
	return true, ""
}

// validateNetworkListenerAttribute check network lb listener attribute
func (g *GclbValidater) validateNetworkListenerAttribute(attr *networkextensionv1.IngressListenerAttribute) (bool, string) {
	return true, ""
}

// validateCertificate check listener certificate
func (g *GclbValidater) validateCertificate(certs *networkextensionv1.IngressListenerCertificate) (bool, string) {
	if len(certs.CertID) == 0 {
		return false, "certID cannot be empty"
	}
	return true, ""
}

// validateListenerMapping check listener mapping
func (g *GclbValidater) validateListenerMapping(mapping *networkextensionv1.IngressPortMapping) (bool, string) {
	// disable hostPort, gcp use service to support lb, hostport is another model to bind lb.
	// maybe support later.
	if mapping.HostPort {
		return false, "hostPort is not support in gcp ingress"
	}
	switch mapping.Protocol {
	case ProtocolHTTP, ProtocolHTTPS:
		if len(mapping.Routes) == 0 {
			return false, fmt.Sprintf("no routes in 7 layer mapping, startPort %d", mapping.StartPort)
		}
		for index := range mapping.Routes {
			if ok, msg := g.validatePortMappingRoute(&mapping.Routes[index]); !ok {
				return ok, msg
			}
		}
		if mapping.Protocol == ProtocolHTTPS {
			if mapping.Certificate == nil {
				return false, "no certificate for https listener"
			}
			if ok, msg := g.validateCertificate(mapping.Certificate); !ok {
				return ok, msg
			}
		}
	case ProtocolTCP, ProtocolUDP:
		if mapping.ListenerAttribute != nil {
			if ok, msg := g.validateNetworkListenerAttribute(mapping.ListenerAttribute); !ok {
				return ok, msg
			}
		}
	default:
		return false, fmt.Sprintf("invalid mapping protocol %s", mapping.Protocol)
	}
	return true, ""
}

func (g *GclbValidater) validateListenerRoute(r *networkextensionv1.Layer7Route) (bool, string) {
	if len(r.Domain) == 0 {
		return false, "domain cannot be empty for 7 layer listener"
	}
	if r.ListenerAttribute != nil {
		if ok, msg := g.validateApplicationListenerAttribute(r.ListenerAttribute); !ok {
			return ok, msg
		}
	}
	for i := range r.Services {
		if ok, msg := g.validateListenerService(&r.Services[i]); !ok {
			return ok, msg
		}
	}
	return true, ""
}

func (g *GclbValidater) validatePortMappingRoute(r *networkextensionv1.IngressPortMappingLayer7Route) (bool, string) {
	if len(r.Domain) == 0 {
		return false, "domain cannot be empty for 7 layer listener"
	}
	if r.ListenerAttribute != nil {
		if ok, msg := g.validateApplicationListenerAttribute(r.ListenerAttribute); !ok {
			return ok, msg
		}
	}
	return true, ""
}

func (g *GclbValidater) validateListenerService(svc *networkextensionv1.ServiceRoute) (bool, string) {
	// disable hostPort, gcp use service to support lb, hostport is another model to bind lb.
	// maybe support later.
	if svc.HostPort {
		return false, "hostPort is not support in gcp ingress"
	}

	// trans nodeport to pod ip:port
	if !svc.IsDirectConnect {
		svc.IsDirectConnect = true
	}
	return true, ""
}
