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

const (
	// SystemNameInMetricAzure system name in metric for azure
	SystemNameInMetricAzure = "azure"
	// HandlerNameInMetricAzureSDK handler name in metric for azure sdk
	HandlerNameInMetricAzureSDK = "sdk"

	// AzureProtocolHTTP elb http protocol
	AzureProtocolHTTP = "HTTP"
	// AzureProtocolHTTPS elb https protocol
	AzureProtocolHTTPS = "HTTPS"
	// AzureProtocolTCP elb tcp protocol
	AzureProtocolTCP = "TCP"
	// AzureProtocolUDP elb udp protocol
	AzureProtocolUDP = "UDP"
	// AzureProtocolTLS elb tls protocol
	AzureProtocolTLS = "TLS"

	// DefaultRequestTimeout seconds that application wait for backend's response
	DefaultRequestTimeout = 20

	// DefaultLoadBalancerProbeInterval seconds that do probe
	DefaultLoadBalancerProbeInterval = 5
	// DefaultLoadBalancerProbeNumberOfProbes  The number of probes where if no response,
	//  will result in stopping further traffic from being delivered to the endpoint.
	//	This values allows endpoints to be taken out of rotation faster or slower than
	//	the typical times used in Azure.
	DefaultLoadBalancerProbeNumberOfProbes = 1

	// MaxRoutingRulePriority max priority allowed by azure application gateway request routing rule
	MaxRoutingRulePriority int32 = 20000

	// Azure itself never assigns a priority to a new rule, the field is mandatory since
	// api-version 2021-08-01. These starting points mirror AGIC, microsoft's own application
	// gateway ingress controller, whose code states the scheme follows how the gateway populates
	// priorities internally: a multi site rule is evaluated before a basic one, and the low
	// numbers stay free for the values users declare themselves.
	// https://github.com/Azure/application-gateway-kubernetes-ingress/blob/master/pkg/appgw/requestroutingrules.go

	// MultiSiteRulePriorityStart first auto assigned priority of a rule carrying a domain
	MultiSiteRulePriorityStart int32 = 19000
	// BasicRulePriorityStart first auto assigned priority of a rule without domain
	BasicRulePriorityStart int32 = 19500
	// RulePriorityJump step between two auto assigned priorities
	RulePriorityJump int32 = 5

	// MaxAzureResourceNameLen max length azure allows for a network child resource name, the limit
	// is the same for application gateway and load balancer children
	MaxAzureResourceNameLen = 80

	// DefaultProbeUnhealthyThreshold default retry count of application gateway probe.
	// Azure only accepts 1~20, so an unset value cannot be sent as is.
	DefaultProbeUnhealthyThreshold = 3

	// agResourceNameSep separator between the listener name and the rest of a generated
	// child resource name. It makes listener ownership checks unambiguous.
	agResourceNameSep = "."

	// DefaultBackendPoolName name for default backend address pool
	DefaultBackendPoolName = "bkbcs-default-backendaddresspool"
	// DefaultBackendSettingName name for default backend setting
	DefaultBackendSettingName = "bkbcs-default-backendsetting"

	// CreateGoroutineLimit define how much goroutines can be used to create resource each time
	CreateGoroutineLimit = 10
	// DeleteGoroutineLimit  define how much goroutines can be used to delete resource each time
	DeleteGoroutineLimit = 10

	// osEnvSep sep of os env
	osEnvSep = ","
)
