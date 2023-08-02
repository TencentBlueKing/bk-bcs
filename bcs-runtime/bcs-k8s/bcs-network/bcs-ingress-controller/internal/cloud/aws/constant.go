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

package aws

const (
	// SystemNameInMetricAWS system name in metric for aws
	SystemNameInMetricAWS = "aws"
	// HandlerNameInMetricAWSSDK handler name in metric for aws sdk
	HandlerNameInMetricAWSSDK = "sdk"
	// HandlerNameInMetricAWSSDKEC2 handler name in metric for aws sdk ec2
	HandlerNameInMetricAWSSDKEC2 = "sdk-ec2"
	// HandlerNameInMetricAWSSDKAGA handler name in metric for aws sdk global accelerator
	HandlerNameInMetricAWSSDKAGA = "sdk-aga"

	// ElbProtocolHTTP elb http protocol
	ElbProtocolHTTP = "HTTP"
	// ElbProtocolHTTPS elb https protocol
	ElbProtocolHTTPS = "HTTPS"
	// ElbProtocolTCP elb tcp protocol
	ElbProtocolTCP = "TCP"
	// ElbProtocolUDP elb udp protocol
	ElbProtocolUDP = "UDP"
)
