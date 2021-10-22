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

package api

import (
	loadbalance "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"
)

const (
	//Clb type

	//ClbPublic face internet
	ClbPublic = 2
	//ClbPrivate face internal network
	ClbPrivate = 3
	//ClbApplicationType application type lb
	ClbApplicationType = 1
	//ClbTraditionalType traditional type lb (never used)
	ClbTraditionalType = 0

	//ClbListenerProtocolHTTP clb listener http protocol
	ClbListenerProtocolHTTP = 1
	//ClbListenerProtocolHTTPS clb listener https protocol
	ClbListenerProtocolHTTPS = 4
	//ClbListenerProtocolTCP clb listener tcp protocol
	ClbListenerProtocolTCP = 2
	//ClbListenerProtocolUDP clb listener udp protocol
	ClbListenerProtocolUDP = 3

	//ClbInstanceRunningStatus clb instance running normally
	ClbInstanceRunningStatus = 1
	//ClbInstanceCreatingStatus clb instance is under creating
	ClbInstanceCreatingStatus = 0

	//ClbSecurityGroupDefaultFromPort clb security group default begin port
	ClbSecurityGroupDefaultFromPort = 31000
	//ClbSecurityGroupDefaultEndPort clb security group default end port
	ClbSecurityGroupDefaultEndPort = 32000

	//ClbConfigPath clb config file path
	ClbConfigPath = "conf/clbConf/clbConfig.conf"

	//ClbSecurityGroupPolicyIndex 修改安全组的第0条规则
	ClbSecurityGroupPolicyIndex = 0
	//ClbVIPSgPolicyIndex 对VIP开放的安全策略位置，在第一条规则上创建，排在端口策略后面
	ClbVIPSgPolicyIndex = 1

	//LBAlgorithmIPHash IP Hash
	LBAlgorithmIPHash = "ip_hash"
	//LBAlgorithmLeastConn least conn
	LBAlgorithmLeastConn = "least_conn"
	//LBAlgorithmRoundRobin round robin
	LBAlgorithmRoundRobin = "wrr"

	// HealthSwitchOn 开启健康检查
	HealthSwitchOn = 1
	// HealthSwitchOff 关闭健康检查
	HealthSwitchOff = 0

	// SecurityGroupPolicyProtocolUDP security policy for udp
	SecurityGroupPolicyProtocolUDP = "udp"
	// SecurityGroupPolicyProtocolTCP security policy for tcp
	SecurityGroupPolicyProtocolTCP = "tcp"

	// TaskResultStatusSuccess result status for success
	TaskResultStatusSuccess = 0
	// TaskResultStatusFailed result status for failed
	TaskResultStatusFailed = 1
	// TaskResultStatusDealing result status for dealing
	TaskResultStatusDealing = 2

	// QCloudLBURL tencent cloud lb url
	QCloudLBURL = "https://lb.api.qcloud.com/v2/index.php"
	// QCloudCVMURL tencent cloud cvm v2 url
	QCloudCVMURL = "https://cvm.api.qcloud.com/v2/index.php"
	// QCloudCVMURLV3 tencent cloud cvm v3 url
	QCloudCVMURLV3 = "https://cvm.tencentcloudapi.com/"
	// QCloudDfwURL tencent cloud dfw url
	QCloudDfwURL = "https://dfw.api.qcloud.com/v2/index.php"

	//ClbMaxTimeout 异步clb接口超时时间，之所以定三分钟是因为之前遇到过最久的接口是一分多钟
	ClbMaxTimeout = 180

	// RequestLimitExceededCode code for request exceeded limit
	RequestLimitExceededCode = 4400
	// RequestLimitExceededMessage message for request exceeded limit
	RequestLimitExceededMessage = "RequestLimitExceeded"
	// WrongStatusCode code for incorrect status
	WrongStatusCode = 4000
	// WrongStatusMessage message for incorrect status
	WrongStatusMessage = "IncorrectStatus.LBWrongStatus"
)

var (
	// LBAlgorithmTypeBcs2QCloudMap map for algorithm name mapping from bcs to qcloud
	LBAlgorithmTypeBcs2QCloudMap = map[string]string{
		loadbalance.ClbLBPolicyLeastConn: LBAlgorithmLeastConn,
		loadbalance.ClbLBPolicyWRR:       LBAlgorithmRoundRobin,
		loadbalance.ClbLBPolicyIPHash:    LBAlgorithmIPHash,
	}
	// LBAlgorithmTypeQCloud2BcsMap map for algorithm name mapping from qcloud to bcs
	LBAlgorithmTypeQCloud2BcsMap = map[string]string{
		LBAlgorithmLeastConn:  loadbalance.ClbLBPolicyLeastConn,
		LBAlgorithmRoundRobin: loadbalance.ClbLBPolicyWRR,
		LBAlgorithmIPHash:     loadbalance.ClbLBPolicyIPHash,
	}
	// NetworkTypeBcs2QCloudMap map for network type from bcs to qcloud
	NetworkTypeBcs2QCloudMap = map[string]int{
		loadbalance.ClbNetworkTypePrivate: ClbPrivate,
		loadbalance.ClbNetworkTypePublic:  ClbPublic,
	}
	// NetworkTypeQCloud2BcsMap map for network type from qcloud to bcs
	NetworkTypeQCloud2BcsMap = map[int]string{
		ClbPrivate: loadbalance.ClbNetworkTypePrivate,
		ClbPublic:  loadbalance.ClbNetworkTypePublic,
	}
	// ProtocolTypeBcs2QCloudMap map for protocol type from bcs to qcloud
	ProtocolTypeBcs2QCloudMap = map[string]int{
		loadbalance.ClbListenerProtocolHTTP:  ClbListenerProtocolHTTP,
		loadbalance.ClbListenerProtocolHTTPS: ClbListenerProtocolHTTPS,
		loadbalance.ClbListenerProtocolTCP:   ClbListenerProtocolTCP,
		loadbalance.ClbListenerProtocolUDP:   ClbListenerProtocolUDP,
	}
	// ProtocolTypeQCloud2BcsMap map for protocol type from qcloud to bcs
	ProtocolTypeQCloud2BcsMap = map[int]string{
		ClbListenerProtocolHTTP:  loadbalance.ClbListenerProtocolHTTP,
		ClbListenerProtocolHTTPS: loadbalance.ClbListenerProtocolHTTPS,
		ClbListenerProtocolTCP:   loadbalance.ClbListenerProtocolTCP,
		ClbListenerProtocolUDP:   loadbalance.ClbListenerProtocolUDP,
	}
)
