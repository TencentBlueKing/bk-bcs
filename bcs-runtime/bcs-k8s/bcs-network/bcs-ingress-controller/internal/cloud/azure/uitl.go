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
 *
 */

package azure

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func transTransportProtocolPtr(str string) *armnetwork.TransportProtocol {
	var protocol armnetwork.TransportProtocol
	switch strings.ToLower(str) {
	case "tcp":
		protocol = armnetwork.TransportProtocolTCP
	case "udp":
		protocol = armnetwork.TransportProtocolUDP
	}

	return &protocol
}

func transAgProtocolPtr(str string) *armnetwork.ApplicationGatewayProtocol {
	var protocol armnetwork.ApplicationGatewayProtocol
	switch strings.ToLower(str) {
	case "http":
		protocol = armnetwork.ApplicationGatewayProtocolHTTP
	case "https":
		protocol = armnetwork.ApplicationGatewayProtocolHTTPS
	case "tcp":
		protocol = armnetwork.ApplicationGatewayProtocolTCP
	case "tls":
		protocol = armnetwork.ApplicationGatewayProtocolTLS
	}

	return &protocol
}

func transProbeProtocolPtr(str string) *armnetwork.ProbeProtocol {
	var protocol armnetwork.ProbeProtocol
	switch strings.ToLower(str) {
	case "http":
		protocol = armnetwork.ProbeProtocolHTTP
	case "https":
		protocol = armnetwork.ProbeProtocolHTTPS
	case "tcp":
		protocol = armnetwork.ProbeProtocolTCP
	}
	return &protocol
}

// transAgProbeMatch translate healthCheck to azure entity
func transAgProbeMatch(healthCheck *networkextensionv1.ListenerHealthCheck) *armnetwork.
	ApplicationGatewayProbeHealthResponseMatch {
	if healthCheck == nil || healthCheck.HTTPCode < 1 || healthCheck.HTTPCode > 31 {
		return nil
	}
	match := &armnetwork.ApplicationGatewayProbeHealthResponseMatch{}
	httpCode := healthCheck.HTTPCode
	cnt := 1
	for httpCode != 0 && cnt <= 5 {
		if httpCode&1 != 0 {
			matchCode := fmt.Sprintf("%d-%d", cnt*100, (cnt+1)*100-1)
			match.StatusCodes = append(match.StatusCodes, to.StringPtr(matchCode))
		}
		httpCode = httpCode >> 1
		cnt++
	}
	return match
}

// listenerName.md5(listenerName+domain+path)
func getRuleTgName(listenerName, domain, path string, listenPort int) string {
	return fmt.Sprintf("%s.%x.%d", listenerName, md5.Sum([]byte(listenerName+domain+path)), listenPort)
}

// listenPort.md5(domain)
func getHttpListenerName(listenPort int, domain string) string {
	return fmt.Sprintf("%d.%x", listenPort, md5.Sum([]byte(domain)))
}

// listenerName.port
func getLBRuleTgName(listenerName string, listenPort int) string {
	return fmt.Sprintf("%s.%d", listenerName, listenPort)
}

// isSamePort check if all backends have same port, return true if all ports are same
func isSamePort(targetGroup *networkextensionv1.ListenerTargetGroup) bool {
	if targetGroup == nil || len(targetGroup.Backends) == 0 {
		return true
	}
	for i := 1; i < len(targetGroup.Backends); i++ {
		if targetGroup.Backends[i].Port != targetGroup.Backends[i-1].Port {
			return false
		}
	}

	return true
}

// isRuleSamePort check backends' port are same in one rule, return true if all ports are same
func isRuleSamePort(listener *networkextensionv1.Listener) bool {
	for _, rule := range listener.Spec.Rules {
		if !isSamePort(rule.TargetGroup) {
			return false
		}
	}
	return true
}

// return 0 if no available priority
func generatePriority(appGateway *armnetwork.ApplicationGateway) int32 {
	// priority 范围： 1～20000
	usedPriority := make([]bool, 20001)
	usedPriority[0] = true
	for _, requestRule := range appGateway.Properties.RequestRoutingRules {
		usedPriority[*requestRule.Properties.Priority] = true
	}

	for i, used := range usedPriority {
		if used == false {
			return int32(i)
		}
	}

	return 0
}

// getBackendPort return targetGroup's backend port, assume all ports are same. return 80 as default
func getBackendPort(targetGroup *networkextensionv1.ListenerTargetGroup) int32 {
	port := 80
	if targetGroup != nil && len(targetGroup.Backends) != 0 {
		port = targetGroup.Backends[0].Port
	}

	return int32(port)
}
