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

package aws

import (
	"fmt"
	"strconv"
	"strings"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// ElbValidater validates the elb parameters
type ElbValidater struct{}

// NewELbValidater creates a new elb validater
func NewELbValidater() *ElbValidater {
	return &ElbValidater{}
}

// IsIngressValid check bcs ingress parameter
func (e *ElbValidater) IsIngressValid(ingress *networkextensionv1.Ingress) (bool, string) {
	if ingress == nil {
		return false, "ingress cannot be empty"
	}
	for _, rule := range ingress.Spec.Rules {
		if ok, msg := e.validateIngressRule(&rule); !ok {
			return ok, msg
		}
	}

	for _, mapping := range ingress.Spec.PortMappings {
		if ok, msg := e.validateListenerMapping(&mapping); !ok {
			return ok, msg
		}
	}
	return true, ""
}

// CheckNoConflictsInIngress return true, if there is no conflicts in ingress itself
func (e *ElbValidater) CheckNoConflictsInIngress(ingress *networkextensionv1.Ingress) (bool, string) {
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
func (e *ElbValidater) validateIngressRule(rule *networkextensionv1.IngressRule) (bool, string) {
	if rule.Port <= 0 || rule.Port >= 65536 {
		return false, fmt.Sprintf("invalid port %d, available [1-65535]", rule.Port)
	}
	if rule.Protocol != ElbProtocolHTTP &&
		rule.Protocol != ElbProtocolHTTPS &&
		rule.Protocol != ElbProtocolTCP &&
		rule.Protocol != ElbProtocolUDP {
		return false, fmt.Sprintf("invalid protocol %s, available [http, https, tcp, udp]", rule.Protocol)
	}
	if rule.Protocol == ElbProtocolHTTPS {
		if rule.Certificate == nil {
			return false, "certificate cannot be empty for protocol https"
		}
		if ok, msg := e.validateCertificate(rule.Certificate); !ok {
			return ok, msg
		}
	}

	switch rule.Protocol {
	case ElbProtocolHTTP, ElbProtocolHTTPS:
		if rule.ListenerAttribute != nil {
			if ok, msg := e.validateApplicationListenerAttribute(rule.ListenerAttribute); !ok {
				return ok, msg
			}
		}
		for _, r := range rule.Routes {
			if ok, msg := e.validateListenerRoute(&r); !ok {
				return ok, msg
			}
		}
	case ElbProtocolTCP, ElbProtocolUDP:
		if rule.ListenerAttribute != nil {
			if ok, msg := e.validateNetworkListenerAttribute(rule.ListenerAttribute); !ok {
				return ok, msg
			}
		}
		for _, svc := range rule.Services {
			if ok, msg := e.validateListenerService(&svc); !ok {
				return ok, msg
			}
		}
	}
	return true, ""
}

// validateAppListenerAttribute check aws application lb listener attribute
// aws validater only check the HealthCheck and AWSAttribute
func (e *ElbValidater) validateApplicationListenerAttribute(attr *networkextensionv1.IngressListenerAttribute) (bool, string) {
	// check HealthCheck
	if attr.HealthCheck != nil {
		if attr.HealthCheck.HealthNum != 0 && (attr.HealthCheck.HealthNum < 2 || attr.HealthCheck.HealthNum > 10) {
			return false, fmt.Sprintf("invalid healthNum %d, available [2, 10]", attr.HealthCheck.HealthNum)
		}
		if attr.HealthCheck.UnHealthNum != 0 &&
			(attr.HealthCheck.UnHealthNum < 2 || attr.HealthCheck.UnHealthNum > 10) {
			return false, fmt.Sprintf("invalid unHealthNum %d, available [2, 10]", attr.HealthCheck.UnHealthNum)
		}
		if attr.HealthCheck.Timeout != 0 &&
			(attr.HealthCheck.Timeout < 2 || attr.HealthCheck.Timeout > 120) {
			return false, fmt.Sprintf("invalid timeout %d, available [2, 120]", attr.HealthCheck.Timeout)
		}
		if attr.HealthCheck.IntervalTime != 0 &&
			(attr.HealthCheck.IntervalTime < 5 || attr.HealthCheck.IntervalTime > 300) {
			return false, fmt.Sprintf("invalid interval time %d, available [5, 300]", attr.HealthCheck.IntervalTime)
		}
		// check code
		if len(attr.HealthCheck.HTTPCodeValues) != 0 && !checkHTTPCodeValues(attr.HealthCheck.HTTPCodeValues) {
			return false, fmt.Sprintf(`invalid http code values %s, you can specify values between 200 and 499, and the default value is 200. You can specify multiple values (for example,"200,202") or a range of values (for example, "200-299")`,
				attr.HealthCheck.HTTPCodeValues)
		}
	}

	// check AWSAttribute
	for _, attributes := range attr.AWSAttributes {
		if attributes.Key == "" || attributes.Value == "" {
			return false, "aws attributes's key and value cannot be empty"
		}
	}
	return true, ""
}

func checkHTTPCodeValues(httpCode string) bool {
	if strings.Contains(httpCode, ",") {
		values := strings.Split(httpCode, ",")
		for _, v := range values {
			if i, err := strconv.Atoi(v); err != nil || i < 200 || i > 499 {
				return false
			}
		}
		return true
	}
	if strings.Contains(httpCode, "-") {
		values := strings.Split(httpCode, "-")
		if len(values) != 2 {
			return false
		}
		a, err := strconv.Atoi(values[0])
		if err != nil || a < 200 || a > 499 {
			return false
		}
		b, err := strconv.Atoi(values[1])
		if err != nil || b < 200 || b > 499 {
			return false
		}
		if a >= b {
			return false
		}
		return true
	}
	if i, err := strconv.Atoi(httpCode); err != nil || i < 200 || i > 499 {
		return false
	}
	return true
}

// validateNetworkListenerAttribute check aws network lb listener attribute
func (e *ElbValidater) validateNetworkListenerAttribute(attr *networkextensionv1.IngressListenerAttribute) (bool, string) {
	// check HealthCheck
	if attr.HealthCheck != nil {
		if attr.HealthCheck.HealthNum != 0 && (attr.HealthCheck.HealthNum < 2 || attr.HealthCheck.HealthNum > 10) {
			return false, fmt.Sprintf("invalid healthNum %d, available [2, 10]", attr.HealthCheck.HealthNum)
		}
		if attr.HealthCheck.IntervalTime != 0 &&
			(attr.HealthCheck.IntervalTime != 10 && attr.HealthCheck.IntervalTime != 30) {
			return false, fmt.Sprintf("invalid interval time %d, available 10 or 30", attr.HealthCheck.IntervalTime)
		}
	}

	// check AWSAttribute
	for _, attributes := range attr.AWSAttributes {
		if attributes.Key == "" || attributes.Value == "" {
			return false, "aws attributes's key and value cannot be empty"
		}
	}
	return true, ""
}

// validateCertificate check listener certificate
func (e *ElbValidater) validateCertificate(certs *networkextensionv1.IngressListenerCertificate) (bool, string) {
	if len(certs.CertID) == 0 {
		return false, "certID cannot be empty"
	}
	return true, ""
}

// validateListenerMapping check listener mapping
func (e *ElbValidater) validateListenerMapping(mapping *networkextensionv1.IngressPortMapping) (bool, string) {
	switch mapping.Protocol {
	case ElbProtocolHTTP, ElbProtocolHTTPS:
		if len(mapping.Routes) == 0 {
			return false, fmt.Sprintf("no routes in 7 layer mapping, startPort %d", mapping.StartPort)
		}
		for index := range mapping.Routes {
			if ok, msg := e.validatePortMappingRoute(&mapping.Routes[index]); !ok {
				return ok, msg
			}
		}
		if mapping.Protocol == ElbProtocolHTTPS {
			if mapping.Certificate == nil {
				return false, "no certificate for https listener"
			}
			if ok, msg := e.validateCertificate(mapping.Certificate); !ok {
				return ok, msg
			}
		}
	case ElbProtocolTCP, ElbProtocolUDP:
		if mapping.ListenerAttribute != nil {
			if ok, msg := e.validateNetworkListenerAttribute(mapping.ListenerAttribute); !ok {
				return ok, msg
			}
		}
	default:
		return false, fmt.Sprintf("invalid mapping protocol %s", mapping.Protocol)
	}
	return true, ""
}

func (e *ElbValidater) validateListenerRoute(r *networkextensionv1.Layer7Route) (bool, string) {
	if len(r.Domain) == 0 {
		return false, "domain cannot be empty for 7 layer listener"
	}
	if r.ListenerAttribute != nil {
		if ok, msg := e.validateApplicationListenerAttribute(r.ListenerAttribute); !ok {
			return ok, msg
		}
	}
	for _, svc := range r.Services {
		if ok, msg := e.validateListenerService(&svc); !ok {
			return ok, msg
		}
	}
	return true, ""
}

func (e *ElbValidater) validatePortMappingRoute(r *networkextensionv1.IngressPortMappingLayer7Route) (bool, string) {
	if len(r.Domain) == 0 {
		return false, "domain cannot be empty for 7 layer listener"
	}
	if r.ListenerAttribute != nil {
		if ok, msg := e.validateApplicationListenerAttribute(r.ListenerAttribute); !ok {
			return ok, msg
		}
	}
	return true, ""
}

func (e *ElbValidater) validateListenerService(svc *networkextensionv1.ServiceRoute) (bool, string) {
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
