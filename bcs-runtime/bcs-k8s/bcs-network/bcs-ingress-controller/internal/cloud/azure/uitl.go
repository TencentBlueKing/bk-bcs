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
	// NOCC:gas/crypto(未用于生成密钥)
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/go-autorest/autorest/to"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func transTransportProtocolPtr(str string) *armnetwork.TransportProtocol {
	var protocol armnetwork.TransportProtocol
	switch strings.ToUpper(str) {
	case AzureProtocolTCP:
		protocol = armnetwork.TransportProtocolTCP
	case AzureProtocolUDP:
		protocol = armnetwork.TransportProtocolUDP
	}

	return &protocol
}

func transAgProtocolPtr(str string) *armnetwork.ApplicationGatewayProtocol {
	var protocol armnetwork.ApplicationGatewayProtocol
	switch strings.ToUpper(str) {
	case AzureProtocolHTTP:
		protocol = armnetwork.ApplicationGatewayProtocolHTTP
	case AzureProtocolHTTPS:
		protocol = armnetwork.ApplicationGatewayProtocolHTTPS
	case AzureProtocolTCP:
		protocol = armnetwork.ApplicationGatewayProtocolTCP
	case AzureProtocolTLS:
		protocol = armnetwork.ApplicationGatewayProtocolTLS
	}

	return &protocol
}

func transProbeProtocolPtr(str string) *armnetwork.ProbeProtocol {
	var protocol armnetwork.ProbeProtocol
	switch strings.ToUpper(str) {
	case AzureProtocolHTTP:
		protocol = armnetwork.ProbeProtocolHTTP
	case AzureProtocolHTTPS:
		protocol = armnetwork.ProbeProtocolHTTPS
	case AzureProtocolTCP, AzureProtocolUDP: // azure不支持UDP协议的健康检查，当使用UDP时，转为TCP
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
	// NOCC:gas/crypto(未用于生成密钥)
	return fmt.Sprintf("%s.%x.%d", listenerName, md5.Sum([]byte(listenerName+domain+path)), listenPort)
}

// listenPort.md5(domain)
func getHttpListenerName(listenPort int, domain string) string {
	// NOCC:gas/crypto(未用于生成密钥)
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

func splitListenersToDiffProtocol(listeners []*networkextensionv1.Listener) [][]*networkextensionv1.Listener {
	retMap := make(map[string][]*networkextensionv1.Listener)
	for _, li := range listeners {
		var listenerList []*networkextensionv1.Listener
		if _, ok := retMap[li.Spec.Protocol]; ok {
			listenerList = retMap[li.Spec.Protocol]
		} else {
			listenerList = make([]*networkextensionv1.Listener, 0)
		}

		if li.Spec.EndPort != 0 {
			listenerList = append(listenerList, splitSegListener([]*networkextensionv1.
				Listener{li})...)
		} else {
			listenerList = append(listenerList, li)
		}

		retMap[li.Spec.Protocol] = listenerList
	}

	retList := make([][]*networkextensionv1.Listener, 0)
	for _, list := range retMap {
		retList = append(retList, list)
	}
	return retList
}

func splitSegListener(listenerList []*networkextensionv1.Listener) []*networkextensionv1.Listener {
	newListenerList := make([]*networkextensionv1.Listener, 0)

	for _, listener := range listenerList {
		if listener.Spec.EndPort == 0 {
			newListenerList = append(newListenerList, listener)
		} else {
			portIndex := 0
			for i := listener.Spec.Port; i <= listener.Spec.EndPort; i++ {
				// generate single port listener to ensure listener
				li := listener.DeepCopy()
				li.Spec.Port = i
				li.Spec.EndPort = 0
				if li.Spec.TargetGroup != nil {
					for j := range li.Spec.TargetGroup.Backends {
						li.Spec.TargetGroup.Backends[j].Port += portIndex
					}
				}
				portIndex++
				newListenerList = append(newListenerList, li)
			}
		}
	}
	return newListenerList
}
