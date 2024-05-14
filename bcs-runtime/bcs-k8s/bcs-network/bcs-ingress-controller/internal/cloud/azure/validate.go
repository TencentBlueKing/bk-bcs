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

package azure

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// AlbValidater check if azure ingress is valid
type AlbValidater struct{}

// NewAlbValidater new azure validater
func NewAlbValidater() *AlbValidater {
	return &AlbValidater{}
}

// IsIngressValid check bcs ingress parameter
func (a *AlbValidater) IsIngressValid(ingress *networkextensionv1.Ingress) (bool, string) {
	if ingress == nil {
		return false, "ingress cannot be empty"
	}
	layer := common.GetIngressProtocolLayer(ingress)
	if layer != constant.ProtocolLayerApplication && layer != constant.ProtocolLayerTransport {
		return false, fmt.Sprintf("ingress[%s/%s] can only be one protocol layer in azure ingress", ingress.Namespace,
			ingress.Name)
	}
	for _, rule := range ingress.Spec.Rules {
		if ok, msg := a.validateIngressRule(&rule); !ok {
			return false, msg
		}
	}

	for _, mapping := range ingress.Spec.PortMappings {
		if ok, msg := a.validateListenerMapping(&mapping); !ok {
			return false, msg
		}
	}
	return true, ""
}

// validateIngressRule check ingress rule
func (a *AlbValidater) validateIngressRule(rule *networkextensionv1.IngressRule) (bool, string) {
	if rule.Port <= 0 || rule.Port >= 65536 {
		return false, fmt.Sprintf("invalid port %d, available [1-65535]", rule.Port)
	}

	if rule.Protocol == AzureProtocolHTTPS {
		if rule.Certificate == nil {
			return false, fmt.Sprintf("certificate cannot be empty for protocol https")
		}
		if ok, msg := a.validateCertificate(rule.Certificate); !ok {
			return false, msg
		}
	}
	if rule.ListenerAttribute != nil {
		var validateFunc func(*networkextensionv1.IngressListenerAttribute) (bool, string)
		if rule.Protocol == AzureProtocolHTTPS || rule.Protocol == AzureProtocolHTTP {
			validateFunc = a.validateAgListenerAttribute
		} else {
			validateFunc = a.validateLBListenerAttribute
		}
		if ok, msg := validateFunc(rule.ListenerAttribute); !ok {
			return false, msg
		}
	}
	switch rule.Protocol {
	case AzureProtocolHTTP, AzureProtocolHTTPS:
		for _, r := range rule.Routes {
			if ok, msg := a.validateListenerRoute(&r); !ok {
				return false, msg
			}
		}
	case AzureProtocolTCP, AzureProtocolUDP:
		for _, svc := range rule.Services {
			if ok, msg := a.validateListenerService(&svc); !ok {
				return false, msg
			}
		}
	default:
		return false, fmt.Sprintf("invalid protocol %s, available [HTTP, HTTPS, TCP, UDP]", rule.Protocol)
	}
	return true, ""
}

func (a *AlbValidater) validateListenerRoute(r *networkextensionv1.Layer7Route) (bool, string) {
	if r.ListenerAttribute != nil {
		if ok, msg := a.validateAgListenerAttribute(r.ListenerAttribute); !ok {
			return false, msg
		}
	}
	for _, svc := range r.Services {
		if ok, msg := a.validateListenerService(&svc); !ok {
			return false, msg
		}
	}
	return true, ""
}

func (a *AlbValidater) validatePortMappingRoute(r *networkextensionv1.IngressPortMappingLayer7Route) (bool, string) {
	if r.ListenerAttribute != nil {
		if ok, msg := a.validateAgListenerAttribute(r.ListenerAttribute); !ok {
			return false, msg
		}
	}
	return true, ""
}

func (a *AlbValidater) validateListenerService(svc *networkextensionv1.ServiceRoute) (bool, string) {
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

// validateLBListenerAttribute check listener attribute
func (a *AlbValidater) validateLBListenerAttribute(attr *networkextensionv1.IngressListenerAttribute) (bool, string) {
	if attr.SessionTime != 0 && (attr.SessionTime < 4 || attr.SessionTime > 30) {
		return false, fmt.Sprintf("invalid session time %d(minute), available [0, 4-30]", attr.SessionTime)
	}

	if attr.HealthCheck != nil && attr.HealthCheck.Enabled {
		healthCheck := attr.HealthCheck
		if healthCheck.HealthCheckProtocol != "" && healthCheck.
			HealthCheckProtocol != AzureProtocolTCP && healthCheck.HealthCheckProtocol != AzureProtocolHTTP {
			return false, fmt.Sprintf("invalid check protocol %s, available [TCP, HTTP]",
				healthCheck.HealthCheckProtocol)
		}
		if healthCheck.HealthCheckProtocol == AzureProtocolHTTP && healthCheck.HTTPCheckPath == "" {
			return false, fmt.Sprintf("http health check need httpCheckPath")
		}
		if healthCheck.IntervalTime != 0 && (healthCheck.IntervalTime < 5 || healthCheck.IntervalTime > 86400) {
			return false, fmt.Sprintf("invalid interval time %d (seconds), available [5, 86400]",
				attr.HealthCheck.IntervalTime)
		}
		if healthCheck.HealthCheckPort != 0 && (healthCheck.HealthCheckPort < 0 || healthCheck.
			HealthCheckPort > 65535) {
			return false, fmt.Sprintf("invalid healthCheckPort %d, available [0, 65535]",
				attr.HealthCheck.HealthCheckPort)
		}
	}
	return true, ""
}

// validateListenerAttribute check listener attribute
func (a *AlbValidater) validateAgListenerAttribute(attr *networkextensionv1.IngressListenerAttribute) (bool, string) {
	if attr.HealthCheck != nil && attr.HealthCheck.Enabled {
		healthCheck := attr.HealthCheck
		if healthCheck.HealthCheckProtocol != "" && healthCheck.
			HealthCheckProtocol != AzureProtocolHTTP && healthCheck.HealthCheckProtocol != AzureProtocolHTTPS {
			return false, fmt.Sprintf("invalid health check protocol %s, available [HTTP, HTTPS]",
				healthCheck.HealthCheckProtocol)
		}
		if healthCheck.UnHealthNum < 0 || healthCheck.UnHealthNum > 20 {
			return false, fmt.Sprintf("invalid unHealthNum %d, available [0, 20]", healthCheck.UnHealthNum)
		}
		if healthCheck.Timeout != 0 && (healthCheck.Timeout < 1 || healthCheck.Timeout > 86400) {
			return false, fmt.Sprintf("invalid timeout %d (seconds), available [1, 86400]", healthCheck.Timeout)
		}
		if healthCheck.IntervalTime != 0 && (healthCheck.IntervalTime < 1 || healthCheck.
			IntervalTime > 86400) {
			return false, fmt.Sprintf("invalid interval time %d (seconds), available [0, 86400]",
				healthCheck.IntervalTime)
		}
		if healthCheck.HTTPCode != 0 &&
			(healthCheck.HTTPCode < 1 || healthCheck.HTTPCode > 31) {
			return false, fmt.Sprintf("invalid httpCode %d, available [1, 31]", healthCheck.HTTPCode)
		}
		if healthCheck.HTTPCheckPath != "" && !strings.HasPrefix(healthCheck.HTTPCheckPath, "/") {
			return false, fmt.Sprintf("invalid httpCheckPath %s, path need start with '/'", healthCheck.HTTPCheckPath)
		}
	}
	return true, ""
}

// validateListenerMapping check listener mapping
func (a *AlbValidater) validateListenerMapping(mapping *networkextensionv1.IngressPortMapping) (bool, string) {
	switch mapping.Protocol {
	case AzureProtocolHTTP, AzureProtocolHTTPS:
		if len(mapping.Routes) == 0 {
			return false, fmt.Sprintf("no routes in 7 layer mapping, startPort %d", mapping.StartPort)
		}
		for index := range mapping.Routes {
			if ok, msg := a.validatePortMappingRoute(&mapping.Routes[index]); !ok {
				return false, msg
			}
		}
		if mapping.Protocol == AzureProtocolHTTPS {
			if mapping.Certificate == nil {
				return false, fmt.Sprintf("no certificate for https listener")
			}
			if ok, msg := a.validateCertificate(mapping.Certificate); !ok {
				return false, msg
			}
		}
	case AzureProtocolTCP, AzureProtocolUDP:
		if mapping.ListenerAttribute != nil {
			if ok, msg := a.validateLBListenerAttribute(mapping.ListenerAttribute); !ok {
				return false, msg
			}
		}
	default:
		return false, fmt.Sprintf("invalid mapping protocol %s", mapping.Protocol)
	}
	return true, ""
}

// CheckNoConflictsInIngress return true, if there is no conflicts in ingress itself
func (a *AlbValidater) CheckNoConflictsInIngress(ingress *networkextensionv1.Ingress) (bool, string) {
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

// validateCertificate check listener certificate
func (a *AlbValidater) validateCertificate(certs *networkextensionv1.IngressListenerCertificate) (bool, string) {
	if len(certs.CertID) == 0 {
		return false, "certID cannot be empty"
	}
	return true, ""
}
