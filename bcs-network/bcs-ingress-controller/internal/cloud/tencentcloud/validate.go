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

package tencentcloud

import (
	"fmt"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// ClbValidater valiadater for clb parameters
type ClbValidater struct{}

// NewClbValidater create clb validater
func NewClbValidater() *ClbValidater {
	return &ClbValidater{}
}

// IsIngressValid check bcs ingress parameter
func (cv *ClbValidater) IsIngressValid(ingress *networkextensionv1.Ingress) (bool, string) {
	if ingress == nil {
		return false, "ingress cannot be empty"
	}
	for _, rule := range ingress.Spec.Rules {
		if ok, msg := cv.validateIngressRule(&rule); !ok {
			return ok, msg
		}
	}

	for _, mapping := range ingress.Spec.PortMappings {
		if ok, msg := cv.validateListenerMapping(&mapping); !ok {
			return ok, msg
		}
	}
	return true, ""
}

// validateIngressRule check ingress rule
func (cv *ClbValidater) validateIngressRule(rule *networkextensionv1.IngressRule) (bool, string) {
	if rule.Port <= 0 || rule.Port >= 65536 {
		return false, fmt.Sprintf("invalid port %d, available [1-65535]", rule.Port)
	}
	if rule.Protocol != ClbProtocolHTTP &&
		rule.Protocol != ClbProtocolHTTPS &&
		rule.Protocol != ClbProtocolTCP &&
		rule.Protocol != ClbProtocolUDP {
		return false, fmt.Sprintf("invalid protocol %s, available [http, https, tcp, udp]", rule.Protocol)
	}
	if rule.Protocol == ClbProtocolHTTPS {
		if rule.Certificate == nil {
			return false, fmt.Sprintf("certificate cannot be empty for protocol https")
		}
		if ok, msg := cv.validateCertificate(rule.Certificate); !ok {
			return ok, msg
		}
	}
	if rule.ListenerAttribute != nil {
		if ok, msg := cv.validateListenerAttribute(rule.ListenerAttribute); !ok {
			return ok, msg
		}
	}
	switch rule.Protocol {
	case ClbProtocolHTTP, ClbProtocolHTTPS:
		for _, r := range rule.Routes {
			if ok, msg := cv.validateListenerRoute(&r); !ok {
				return ok, msg
			}
		}
	case ClbProtocolTCP, ClbProtocolUDP:
		for _, svc := range rule.Services {
			if ok, msg := cv.validateListenerService(&svc); !ok {
				return ok, msg
			}
		}
	}
	return true, ""
}

func (cv *ClbValidater) validateListenerRoute(r *networkextensionv1.Layer7Route) (bool, string) {
	if len(r.Domain) == 0 {
		return false, "domain cannot be empty for 7 layer listener"
	}
	if r.ListenerAttribute != nil {
		if ok, msg := cv.validateListenerAttribute(r.ListenerAttribute); !ok {
			return ok, msg
		}
	}
	for _, svc := range r.Services {
		if ok, msg := cv.validateListenerService(&svc); !ok {
			return ok, msg
		}
	}
	return true, ""
}

func (cv *ClbValidater) validatePortMappingRoute(r *networkextensionv1.IngressPortMappingLayer7Route) (bool, string) {
	if len(r.Domain) == 0 {
		return false, "domain cannot be empty for 7 layer listener"
	}
	if r.ListenerAttribute != nil {
		if ok, msg := cv.validateListenerAttribute(r.ListenerAttribute); !ok {
			return ok, msg
		}
	}
	return true, ""
}

func (cv *ClbValidater) validateListenerService(svc *networkextensionv1.ServiceRoute) (bool, string) {
	if svc.Weight != nil && (svc.Weight.Value < 0 || svc.Weight.Value > 100) {
		return false, fmt.Sprintf("invalid weight value %d, avaialbe [0-100]", svc.Weight.Value)
	}
	for _, set := range svc.Subsets {
		if set.Weight != nil && (set.Weight.Value < 0 || set.Weight.Value > 100) {
			return false, fmt.Sprintf("invalid weight value %d, avaialbe [0-100]", svc.Weight.Value)
		}
	}
	return true, ""
}

// validateListenerAttribute check listener attribute
func (cv *ClbValidater) validateListenerAttribute(attr *networkextensionv1.IngressListenerAttribute) (bool, string) {
	if attr.SessionTime != 0 && (attr.SessionTime < 30 || attr.SessionTime > 3600) {
		return false, fmt.Sprintf("invalid session time %d, available [0, 30-3600]", attr.SessionTime)
	}
	if len(attr.LbPolicy) != 0 {
		if attr.LbPolicy != "WRR" &&
			attr.LbPolicy != "LEAST_CONN" {
			return false, fmt.Sprintf("invalid lb policy %s, available [WRR, LEAST_CONN]", attr.LbPolicy)
		}
	}
	if attr.HealthCheck != nil && attr.HealthCheck.Enabled {
		if attr.HealthCheck.HealthNum != 0 && (attr.HealthCheck.HealthNum < 2 || attr.HealthCheck.HealthNum > 10) {
			return false, fmt.Sprintf("invalid healthNum %d, available [2, 10]", attr.HealthCheck.HealthNum)
		}
		if attr.HealthCheck.UnHealthNum != 0 &&
			(attr.HealthCheck.UnHealthNum < 2 || attr.HealthCheck.UnHealthNum > 10) {
			return false, fmt.Sprintf("invalid unHealthNum %d, available [2, 10]", attr.HealthCheck.UnHealthNum)
		}
		if attr.HealthCheck.Timeout != 0 &&
			(attr.HealthCheck.Timeout < 2 || attr.HealthCheck.Timeout > 60) {
			return false, fmt.Sprintf("invalid timeout %d, available [2, 60]", attr.HealthCheck.Timeout)
		}
		if attr.HealthCheck.IntervalTime != 0 &&
			(attr.HealthCheck.IntervalTime < 5 || attr.HealthCheck.IntervalTime > 300) {
			return false, fmt.Sprintf("invalid interval time %d, available [5, 300]", attr.HealthCheck.IntervalTime)
		}
		if attr.HealthCheck.HTTPCode != 0 &&
			(attr.HealthCheck.HTTPCode < 1 || attr.HealthCheck.HTTPCode > 31) {
			return false, fmt.Sprintf("invalid httpCode %d, available [1, 31]", attr.HealthCheck.HTTPCode)
		}
	}
	return true, ""
}

// validateCertificate check listener certificate
func (cv *ClbValidater) validateCertificate(certs *networkextensionv1.IngressListenerCertificate) (bool, string) {
	if certs.Mode != "UNIDIRECTIONAL" && certs.Mode != "MUTUAL" {
		return false, fmt.Sprintf("invalid tls mod %s, available [UNIDIRECTIONAL, MUTUAL]", certs.Mode)
	}
	if len(certs.CertID) == 0 {
		return false, "certID cannot be empty"
	}
	if certs.Mode == "MUTUAL" && len(certs.CertCaID) == 0 {
		return false, "certCaID cannot be empty"
	}
	return true, ""
}

// validateListenerMapping check listener mapping
func (cv *ClbValidater) validateListenerMapping(mapping *networkextensionv1.IngressPortMapping) (bool, string) {
	switch mapping.Protocol {
	case ClbProtocolHTTP, ClbProtocolHTTPS:
		if len(mapping.Routes) == 0 {
			return false, fmt.Sprintf("no routes in 7 layer mapping, startPort %d", mapping.StartPort)
		}
		for index := range mapping.Routes {
			if ok, msg := cv.validatePortMappingRoute(&mapping.Routes[index]); !ok {
				return ok, msg
			}
		}
		if mapping.Protocol == ClbProtocolHTTPS {
			if mapping.Certificate == nil {
				return false, fmt.Sprintf("no certificate for https listener")
			}
			if ok, msg := cv.validateCertificate(mapping.Certificate); !ok {
				return ok, msg
			}
		}
	case ClbProtocolTCP, ClbProtocolUDP:
		if mapping.ListenerAttribute != nil {
			if ok, msg := cv.validateListenerAttribute(mapping.ListenerAttribute); !ok {
				return ok, msg
			}
		}
	default:
		return false, fmt.Sprintf("invalid mapping protocol %s", mapping.Protocol)
	}
	return true, ""
}

// CheckNoConflictsInIngress return true, if there is no conflicts in ingress itself
func (cv *ClbValidater) CheckNoConflictsInIngress(ingress *networkextensionv1.Ingress) (bool, string) {
	ruleMap := make(map[int]networkextensionv1.IngressRule)
	portReuseMap := make(map[int]struct{})
	for index, rule := range ingress.Spec.Rules {
		existedRule, ok := ruleMap[rule.Port]
		if !ok {
			ruleMap[rule.Port] = ingress.Spec.Rules[index]
			continue
		}
		// for tencent cloud clb, udp and tcp listener can use the same port with different protocol
		if (rule.Protocol == ClbProtocolTCP && existedRule.Protocol == ClbProtocolUDP) ||
			(existedRule.Protocol == ClbProtocolTCP && rule.Protocol == ClbProtocolUDP) {
			_, ok := portReuseMap[rule.Port]
			if !ok {
				portReuseMap[rule.Port] = struct{}{}
				continue
			}
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
