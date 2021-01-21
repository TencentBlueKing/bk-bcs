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

package qcloud

//Parameters from environment variable
const (

	// ConfigBcsClbRegion lb region from environment variable
	ConfigBcsClbRegion = "BCS_CLB_REGION"
	// ConfigBcsClbSecretID secret id from environment variable
	ConfigBcsClbSecretID = "BCS_CLB_SECRETID"
	// ConfigBcsClbSecretKey secret key from environment variable
	ConfigBcsClbSecretKey = "BCS_CLB_SECRETKEY"
	// ConfigBcsClbProjectID project id from environment variable
	ConfigBcsClbProjectID = "BCS_CLB_PROJECTID"
	// ConfigBcsClbCertID cert id from environment variable
	ConfigBcsClbCertID = "BCS_CLB_CERTID"
	// ConfigBcsClbVpcID vpc id from environment variable
	ConfigBcsClbVpcID = "BCS_CLB_VPCID"
	// ConfigBcsClbSubnet subnet from environment variable
	ConfigBcsClbSubnet = "BCS_CLB_SUBNET"
	// ConfigBcsClbSecurity security from environment variable
	ConfigBcsClbSecurity = "BCS_CLB_SECURITY"
	// ConfigBcsClbNetworkType networktype from environment variable
	ConfigBcsClbNetworkType = "BCS_CLB_NETWORKTYPE"
	// ConfigBcsClbBackendMode backend mode, CVM or ENI
	ConfigBcsClbBackendMode    = "BCS_CLB_BACKENDMODE"
	ConfigBcsClbBackendModeCVM = "cvm"
	ConfigBcsClbBackendModeENI = "eni"
	// ConfigBcsClbCidrIP source cidr ip
	ConfigBcsClbCidrIP = "BCS_CLB_CIDRIP"
	// ConfigBcsClbExpireTime expire time
	ConfigBcsClbExpireTime = "BCS_CLB_EXPIRETIME"
	// DefaultClbCidrIP default clb cidr ip(默认开通科兴网段)
	DefaultClbCidrIP = ""

	// ConfigBcsClbMaxTimeout max retry times
	ConfigBcsClbMaxTimeout = "BCS_CLB_MAX_TIMEOUT"
	// DefaultClbMaxTimeout
	DefaultClbMaxTimeout = 180
	// ConfigBcsClbWaitPeriodExceedLimit wait seconds for api exceed limit
	ConfigBcsClbWaitPeriodExceedLimit = "BCS_CLB_WAIT_PERIOD_EXCEED_LIMIT"
	// DefaultClbWaitPeriodExceedLimit default wait seconds for api exceed limit
	DefaultClbWaitPeriodExceedLimit = 10
	// ConfigBcsClbWaitPeriodDealing wait secondes for lb busy
	ConfigBcsClbWaitPeriodDealing = "BCS_CLB_WAIT_PERIOD_DEALING"
	// DefaultClbWaitPeriodDealing default wait seconds for when lb is busy
	DefaultClbWaitPeriodDealing = 3

	// ConfigBcsClbImplement clb implement type
	ConfigBcsClbImplement = "BCS_CLB_IMPLEMENT"
	// ConfigBcsClbImplementAPI use api
	ConfigBcsClbImplementAPI = "api"
	// ConfigBcsClbImplementSDK use official sdk
	ConfigBcsClbImplementSDK = "sdk"

	// LimitationMaxListenerNum clb limitation
	LimitationMaxListenerNum = 50
	// LimitationMaxRulePerListener clb max rule for per listener
	LimitationMaxRulePerListener = 50
	// LimitationMaxBackendNumPerRule max backend number for per rule
	LimitationMaxBackendNumPerRule = 100
	// LimitationMaxBackendNumEachBind max backend number for each binding operation
	LimitationMaxBackendNumEachBind = 20
)

// CheckRegion validate region field
func CheckRegion(region string) bool {
	if len(region) == 0 {
		return false
	}
	return true
}
