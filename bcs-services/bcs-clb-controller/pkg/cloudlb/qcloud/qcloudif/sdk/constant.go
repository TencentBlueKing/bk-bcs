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

package sdk

import (
	cloudlbType "github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
)

const (
	LoadBalancerForwardApplication = 1
	LoadBalancerForwardClassic     = 0

	LoadBalancerNetworkPublic   = "OPEN"
	LoadBalancerNetworkInternal = "INTERNAL"

	ListenerProtocolHTTP  = "HTTP"
	ListenerProtocolHTTPS = "HTTPS"
	ListenerProtocolTCP   = "TCP"
	ListenerProtocolUDP   = "UDP"

	ListenerSSLModeUnidirectional = "UNIDIRECTIONAL"
	ListenerSSLModeMutual         = "MUTUAL"

	LBAlgorithmLeastConn  = "LEAST_CONN"
	LBAlgorithmRoundRobin = "WRR"
	LBAlgorithmIPHash     = "IP_HASH"

	// RequestLimitExceededCode code for request exceeded limit
	RequestLimitExceededCode = "4400"
	// WrongStatusCode code for incorrect status
	WrongStatusCode = "4000"
)

// ProtocolBcs2SDKMap map for bcs protocol type to sdk protocol type
var ProtocolBcs2SDKMap = map[string]string{
	cloudlbType.ClbListenerProtocolHTTP:  ListenerProtocolHTTP,
	cloudlbType.ClbListenerProtocolHTTPS: ListenerProtocolHTTPS,
	cloudlbType.ClbListenerProtocolTCP:   ListenerProtocolTCP,
	cloudlbType.ClbListenerProtocolUDP:   ListenerProtocolUDP,
}

// ProtocolSDK2BcsMap map for sdk protocol type to bcs protocol type
var ProtocolSDK2BcsMap = map[string]string{
	ListenerProtocolHTTP:  cloudlbType.ClbListenerProtocolHTTP,
	ListenerProtocolHTTPS: cloudlbType.ClbListenerProtocolHTTPS,
	ListenerProtocolTCP:   cloudlbType.ClbListenerProtocolTCP,
	ListenerProtocolUDP:   cloudlbType.ClbListenerProtocolUDP,
}

// LBAlgorithmTypeBcs2SDKMap map for bcs lb policy to sdk lb policy
var LBAlgorithmTypeBcs2SDKMap = map[string]string{
	cloudlbType.ClbLBPolicyLeastConn: LBAlgorithmLeastConn,
	cloudlbType.ClbLBPolicyWRR:       LBAlgorithmRoundRobin,
	cloudlbType.ClbLBPolicyIPHash:    LBAlgorithmIPHash,
}

// LBAlgorithmTypeSDK2BcsMap map for sdk lb policy to bcs lb policy
var LBAlgorithmTypeSDK2BcsMap = map[string]string{
	LBAlgorithmLeastConn:  cloudlbType.ClbLBPolicyLeastConn,
	LBAlgorithmRoundRobin: cloudlbType.ClbLBPolicyWRR,
	LBAlgorithmIPHash:     cloudlbType.ClbLBPolicyIPHash,
}

// SSLModeBcs2SDKMap map for bcs ssl mode to sdk ssl mode
var SSLModeBcs2SDKMap = map[string]string{
	cloudlbType.ClbListenerTLSModeUniDirectional: ListenerSSLModeUnidirectional,
	cloudlbType.ClbListenerTLSModeMutual:         ListenerSSLModeMutual,
}

// SSLModeSDK2BcsMap map for sdk ssl mode to bcs ssl mode
var SSLModeSDK2BcsMap = map[string]string{
	ListenerSSLModeUnidirectional: cloudlbType.ClbListenerTLSModeUniDirectional,
	ListenerSSLModeMutual:         cloudlbType.ClbListenerTLSModeMutual,
}
