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

package tencentcloud

const (
	// ClbProtocolHTTP clb http protocol
	ClbProtocolHTTP = "HTTP"
	// ClbProtocolHTTPS clb https protocol
	ClbProtocolHTTPS = "HTTPS"
	// ClbProtocolTCP clb tcp protocol
	ClbProtocolTCP = "TCP"
	// ClbProtocolUDP clb udp protocol
	ClbProtocolUDP = "UDP"
	// ClbProtocolTCPSSL clb tcp_ssl protocol
	ClbProtocolTCPSSL = "TCP_SSL"
	// ClbProtocolGRPC clb grpc protocol
	ClbProtocolGRPC = "GRPC"

	// just for v2 api

	// ClbListenerProtocolHTTP clb listener http protocol
	ClbListenerProtocolHTTP = 1
	// ClbListenerProtocolHTTPS clb listener https protocol
	ClbListenerProtocolHTTPS = 4
	// ClbListenerProtocolTCP clb listener tcp protocol
	ClbListenerProtocolTCP = 2
	// ClbListenerProtocolUDP clb listener udp protocol
	ClbListenerProtocolUDP = 3
	// DefaultTencentCloudClbV2Domain default domain for tencent cloud clb
	DefaultTencentCloudClbV2Domain = "lb.api.qcloud.com"

	// DefaultHealthCheckEnabled default value for health check enabled
	DefaultHealthCheckEnabled = 1
	// DefaultHealthCheckIntervalTime default value for health check interval time
	DefaultHealthCheckIntervalTime = 5
	// DefaultHealthCheckTimeout default value for health check timeout
	DefaultHealthCheckTimeout = 2
	// DefaultHealthCheckHealthNum default value for health check health num
	DefaultHealthCheckHealthNum = 3
	// DefaultHealthCheckUnhealthNum default value for healtch check unhealthy num
	DefaultHealthCheckUnhealthNum = 3

	// SystemNameInMetricTencentCloud system name in metric for tencent cloud
	SystemNameInMetricTencentCloud = "tencentcloud"
	// HandlerNameInMetricTencentCloudAPI handler name in metric for tencent cloud api
	HandlerNameInMetricTencentCloudAPI = "api"
	// HandlerNameInMetricTencentCloudSDK handler name in metric for tencent cloud sdk
	HandlerNameInMetricTencentCloudSDK = "sdk"

	// MaxTargetForRegisterEachTime max target number for registering each time
	MaxTargetForRegisterEachTime = 20
	// MaxListenerForDescribeTargetEachTime max listener number for describing targets each time
	MaxListenerForDescribeTargetEachTime = 20

	// MaxListenersForDeleteEachTime max listener number to delete each time
	MaxListenersForDeleteEachTime = 20
	// MaxListenersForCreateEachTime max listener number to create each time
	MaxListenersForCreateEachTime = 50
	// MaxListenersForDescribeEachTime max listener number to describe each time
	MaxListenersForDescribeEachTime = 100
	// MaxTargetForBatchRegisterEachTime max target number for batch registering each time
	MaxTargetForBatchRegisterEachTime = 250
	// MaxLoadBalancersForDescribeHealthStatus max loadbalancers to describe health status
	MaxLoadBalancersForDescribeHealthStatus = 5

	// MaxSegmentListenerCurrentCreateEachTime max segment listener number to create each time,
	MaxSegmentListenerCurrentCreateEachTime = 5

	// ClbBackendAlive alive status of clb backend
	ClbBackendAlive = "Alive"
	// ClbBackendDead dead status of clb backend
	ClbBackendDead = "Dead"
	// ClbBackendUnknown unknown status of clb backend
	ClbBackendUnknown = "Unknown"
)

var (
	// ProtocolTypeBcs2QCloudMap map for translate protocol (just for v2 api)
	ProtocolTypeBcs2QCloudMap = map[string]int{
		ClbProtocolHTTP:  ClbListenerProtocolHTTP,
		ClbProtocolHTTPS: ClbListenerProtocolHTTPS,
		ClbProtocolTCP:   ClbListenerProtocolTCP,
		ClbProtocolUDP:   ClbListenerProtocolUDP,
	}
)
