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

import (
	"fmt"
	"net/url"
)

//DescribeLBInput query LoadBalance instance info in clb
type DescribeLBInput struct {
	APIMeta         `url:",inline"`
	Forward         int    `url:"forward,omitempty"`
	LoadBalanceID   string `url:"loadBalancerIds.0,omitempty"`
	LoadBalanceName string `url:"loadBalancerName,omitempty"`
	LoadBalanceType int    `url:"loadBalancerType,omitempty"`
	ProjectID       int    `url:"projectId,omitempty"`
}

//DescribeLBOutput DescribeLB result
type DescribeLBOutput struct {
	Response     `json:",inline"`
	TotalCount   int               `json:"totalCount"`
	LoadBalances []LoadBalanceInfo `json:"loadBalancerSet"`
}

//LoadBalanceInfo sub info of DescribeLBResp
type LoadBalanceInfo struct {
	CreateTime       string   `json:"createTime"`
	LoadBalancerID   string   `json:"loadBalancerId"`
	LoadBalancerName string   `json:"loadBalancerName"`
	UnLoadBalancerID string   `json:"unLoadBalancerId"`
	LoadBalancerType int      `json:"loadBalancerType"`
	Forward          int      `json:"forward"`
	Domain           string   `json:"domain"`
	ProjectID        int      `json:"projectId"`
	VpcID            int      `json:"vpcId"`
	SubnetID         int      `json:"subnetId"`
	Isolation        int      `json:"isolation"`
	Snat             bool     `json:"snat"`
	LoadBalancerVips []string `json:"loadBalancerVips"`
	Status           int      `json:"status"`
	StatusTime       string   `json:"statusTime"`
}

//DescribeListenerInput query LoadBalance listener info in clb
type DescribeListenerInput struct {
	APIMeta          `url:",inline"`
	ListenerID       string `url:"listenerIds.0,omitempty"`
	LoadBalanceID    string `url:"loadBalancerId,omitempty"`
	LoadBalancerPort int    `url:"loadBalancerPort,omitempty"`
	Protocol         int    `url:"protocol,omitempty"`
}

//DescribeListenerOutput DescribeListener result
type DescribeListenerOutput struct {
	Response  `json:",inline"`
	Listeners []ListenerInfo `json:"listenerSet"`
}

// DescribeForwardLBListenersInput input for query forward lb listeners
type DescribeForwardLBListenersInput struct {
	APIMeta          `url:",inline"`
	ListenerID       string `url:"listenerIds.0,omitempty"`
	LoadBalanceID    string `url:"loadBalancerId,omitempty"`
	LoadBalancerPort int    `url:"loadBalancerPort,omitempty"`
	Protocol         int    `url:"protocol,omitempty"`
}

// DescribeForwardLBListenersOutput output for query forward lb listeners
type DescribeForwardLBListenersOutput struct {
	Response  `json:",inline"`
	Listeners []ListenerInfo `json:"listenerSet"`
}

//ListenerInfo sub info of DescribeListenerResp
type ListenerInfo struct {
	LoadBalancerPort int        `json:"loadBalancerPort"`
	EndPort          int        `json:"endPort"`
	Protocol         int        `json:"protocol"`
	ProtocolType     string     `json:"protocolType"`
	ListenerID       string     `json:"listenerId"`
	SSLMode          string     `json:"SSLMode"`
	CertID           string     `json:"certId"`
	CertCaID         string     `json:"certCaId"`
	Rules            []RuleInfo `json:"rules"`
}

//RuleInfo sub info of ListenerInfo
type RuleInfo struct {
	LocationID      string `json:"locationId"`
	Domain          string `json:"domain"`
	URL             string `json:"url"`
	HTTPHash        string `json:"httpHash"`
	SessionExpire   int    `json:"sessionExpire"`
	HealthSwitch    int    `json:"healthSwitch"`
	TimeOut         int    `json:"timeOut"`
	IntervalTime    int    `json:"intervalTime"`
	HealthNum       int    `json:"healthNum"`
	UnhealthNum     int    `json:"unhealthNum"`
	HTTPCode        int    `json:"httpCode"`
	HTTPCheckPath   string `json:"httpCheckPath"`
	HTTPCheckMethod string `json:"httpCheckMethod"`
}

//CreateLBInput create clb instance info
type CreateLBInput struct {
	APIMeta          `url:",inline"`
	Forward          int    `url:"forward,omitempty"`
	LoadBalancerName string `url:"loadBalancerName,omitempty"`
	LoadBalancerType int    `url:"loadBalancerType,omitempty"`
	ProjectID        int    `url:"projectId,omitempty"`
	SubnetID         string `url:"subnetId,omitempty"`
	VpcID            string `url:"vpcId,omitempty"`
}

//CreateLBOutput create lb result
type CreateLBOutput struct {
	Response          `json:",inline"`
	UnLoadBalancerIds map[string][]string `json:"unLoadBalancerIds"`
	RequestID         int                 `json:"requestId"`
	DealIds           []string            `json:"dealIds"`
}

//CreateSeventhLayerListenerInput create 7 layer listener info
type CreateSeventhLayerListenerInput struct {
	APIMeta                   `url:",inline"`
	ListenersCertID           string `url:"listeners.0.certId,omitempty"`
	ListenersCertCaID         string `url:"listeners.0.certCaId,omitempty"`
	ListenersCertCaContent    string `url:"listeners.0.certCaContent,omitempty"`
	ListenersCertCaName       string `url:"listeners.0.certCaName,omitempty"`
	ListenersCertContent      string `url:"listeners.0.certContent,omitempty"`
	ListenersCertKey          string `url:"listeners.0.certContent,omitempty"`
	ListenersCertName         string `url:"listeners.0.certName,omitempty"`
	ListenersListenerName     string `url:"listeners.0.listenerName,omitempty"`
	ListenersLoadBalancerPort int    `url:"listeners.0.loadBalancerPort,omitempty"`
	ListenersProtocol         int    `url:"listeners.0.protocol,omitempty"`
	ListenersSSLMode          string `url:"listeners.0.SSLMode,omitempty"`
	LoadBalanceID             string `url:"loadBalancerId,omitempty"`
}

//CreateSeventhLayerListenerOutput create 7 layer listener response
type CreateSeventhLayerListenerOutput struct {
	Response    `json:",inline"`
	ListenerIds []string `json:"listenerIds"`
}

//CreateForwardLBFourthLayerListenersInput create 4 lay listener info
//4层协议没有加密证书
//文档地址 https://cloud.tencent.com/document/api/214/9001
//没有摘取的值都是用默认值
type CreateForwardLBFourthLayerListenersInput struct {
	APIMeta                   `url:",inline"`
	ListenersListenerName     string `url:"listeners.0.listenerName"`
	ListenersLoadBalancerPort int    `url:"listeners.0.loadBalancerPort"`
	EndPort                   int    `url:"listeners.0.endPort"`
	ListenersProtocol         int    `url:"listeners.0.protocol"`
	ListenerExpireTime        int    `url:"listeners.0.sessionExpire,omitempty"`
	ListenerHealthSwitch      int    `url:"listeners.0.healthSwitch"`
	ListenerIntervalTime      int    `url:"listeners.0.intervalTime,omitempty"`
	ListenerTimeout           int    `url:"listeners.0.timeOut,omitempty"`
	ListenerHealthNum         int    `url:"listeners.0.healthNum,omitempty"`
	ListenerUnHealthNum       int    `url:"listeners.0.unhealthNum,omitempty"`
	ListenerScheduler         string `url:"listeners.0.scheduler,omitempty"`
	LoadBalanceID             string `url:"loadBalancerId,omitempty"`
}

// CreateForwardLBFourthLayerListenersOutput CreateForwardLBFourthLayerListeners result
type CreateForwardLBFourthLayerListenersOutput struct {
	AsynchronousBaseResponse `json:",inline"`
	UListenerIds             []string `json:"uListenerIds"`
	ListenerIds              []string `json:"listenerIds"`
}

// ModifyForwardLBFourthListenerInput ModifyForwardLBFourthListener input
type ModifyForwardLBFourthListenerInput struct {
	APIMeta       `url:",inline"`
	LoadBalanceID string `url:"loadBalancerId"`
	ListenerID    string `url:"listenerId"`
	ListenerName  string `url:"listenerName,omitempty"`
	SessionExpire int    `url:"sessionExpire,omitempty"`
	HealthSwitch  int    `url:"healthSwitch"`
	Timeout       int    `url:"timeOut,omitempty"`
	IntervalTime  int    `url:"intervalTime,omitempty"`
	HealthNum     int    `url:"healthNum,omitempty"`
	UnHealthNum   int    `url:"unhealthNum,omitempty"`
	Scheduler     string `url:"scheduler,omitempty"`
}

// ModifyForwardLBFourthListenerOutput ModifyForwardLBFourthListener output
type ModifyForwardLBFourthListenerOutput struct {
	AsynchronousBaseResponse `json:",inline"`
}

// ModifyForwardLBSeventhListenerInput ModifyForwardLBSeventhListener input
type ModifyForwardLBSeventhListenerInput struct {
	APIMeta       `url:",inline"`
	LoadBalanceID string `url:"loadBalancerId"`
	ListenerID    string `url:"listenerId"`
	ListenerName  string `url:"listenerName,omitempty"`
	SSLMode       string `url:"SSLMode,omitempty"`
	CertID        string `url:"certId,omitempty"`
	CertCaID      string `url:"certCaId,omitempty"`
	CertCaContent string `url:"certCaContent,omitempty"`
	CertCaName    string `url:"certCaName,omitempty"`
	CertContent   string `url:"certContent,omitempty"`
	CertKey       string `url:"certKey,omitempty"`
	CertName      string `url:"certName,omitempty"`
}

// ModifyForwardLBSeventhListenerOutput ModifyForwardLBSeventhListener output
type ModifyForwardLBSeventhListenerOutput struct {
	AsynchronousBaseResponse `json:",inline"`
}

// ModifyLoadBalancerRulesProbeInput ModifyLoadBalancerRulesProbe input
type ModifyLoadBalancerRulesProbeInput struct {
	APIMeta       `url:",inline"`
	LoadBalanceID string `url:"loadBalancerId"`
	ListenerID    string `url:"listenerId"`
	LocationID    string `url:"locationId"`
	URL           string `url:"url,omitempty"`
	SessionExpire int    `url:"sessionExpire,omitempty"`
	HealthSwitch  int    `url:"healthSwitch"`
	Timeout       int    `url:"timeOut,omitempty"`
	IntervalTime  int    `url:"intervalTime,omitempty"`
	HealthNum     int    `url:"healthNum,omitempty"`
	UnHealthNum   int    `url:"unhealthNum,omitempty"`
	HTTPHash      string `url:"httpHash,omitempty"`
	HTTPCode      int    `url:"httpCode,omitempty"`
	HTTPCheckPath string `url:"httpCheckPath,omitempty"`
}

// ModifyLoadBalancerRulesProbeOutput ModifyLoadBalancerRulesProbe output
type ModifyLoadBalancerRulesProbeOutput struct {
	AsynchronousBaseResponse `json:",inline"`
}

// ModifyForwardLBRulesDomainInput ModifyForwardLBRulesDomain input
type ModifyForwardLBRulesDomainInput struct {
	APIMeta       `url:",inline"`
	LoadBalanceID string `url:"loadBalancerId"`
	ListenerID    string `url:"listenerId"`
	LocationID    string `url:"locationId"`
	Domain        string `url:"domain"`
	NewDomain     string `url:"newDomain"`
}

// ModifyForwardLBRulesDomainOutput ModifyForwardLBRulesDomain output
type ModifyForwardLBRulesDomainOutput struct {
	AsynchronousBaseResponse `json:",inline"`
}

//RegisterInstancesWithForwardLBSeventhListenerInput register instance info 7 layer
type RegisterInstancesWithForwardLBSeventhListenerInput struct {
	APIMeta       `url:",inline"`
	Backends      BackendTargetList
	ListenerID    string `url:"listenerId,omitempty"`
	LoadBalanceID string `url:"loadBalancerId,omitempty"`
	LocationID    string `url:"locationIds.0,omitempty"`
	URL           string `url:"url,omitempty"`
}

//RegisterInstancesWithForwardLBSeventhListenerOutput register instance response
type RegisterInstancesWithForwardLBSeventhListenerOutput struct {
	AsynchronousBaseResponse
}

//BackendTarget backend to be registered with loadbalance
type BackendTarget struct {
	BackendsInstanceID string
	BackendsIP         string
	BackendsPort       int
	BackendsWeight     int
}

//BackendTargetList backend list
type BackendTargetList []BackendTarget

//EncodeValues encode backend target info into url format
func (btl BackendTargetList) EncodeValues(key string, urlv *url.Values) error {
	for i, v := range btl {
		if len(v.BackendsInstanceID) != 0 {
			urlv.Set(fmt.Sprintf("backends.%d.instanceId", i), fmt.Sprintf("%v", v.BackendsInstanceID))
		} else if len(v.BackendsIP) != 0 {
			urlv.Set(fmt.Sprintf("backends.%d.eniIp", i), fmt.Sprintf("%v", v.BackendsIP))
		}
		urlv.Set(fmt.Sprintf("backends.%d.port", i), fmt.Sprintf("%v", v.BackendsPort))
		urlv.Set(fmt.Sprintf("backends.%d.weight", i), fmt.Sprintf("%v", v.BackendsWeight))

	}
	return nil
}

//RegisterInstancesWithForwardLBFourthListenerInput regitster instance info 4 layer
type RegisterInstancesWithForwardLBFourthListenerInput struct {
	APIMeta       `url:",inline"`
	Backends      BackendTargetList
	ListenerID    string `url:"listenerId,omitempty"`
	LoadBalanceID string `url:"loadBalancerId,omitempty"`
}

//RegisterInstancesWithForwardLBFourthListenerOutput register 4 layer result
type RegisterInstancesWithForwardLBFourthListenerOutput struct {
	AsynchronousBaseResponse
}

//DeregisterInstancesFromForwardLBFourthListenerInput deregister 4 layer info
type DeregisterInstancesFromForwardLBFourthListenerInput struct {
	RegisterInstancesWithForwardLBFourthListenerInput
}

//DeregisterInstancesFromForwardLBFourthListenerOutput deregister 4 layer result
type DeregisterInstancesFromForwardLBFourthListenerOutput struct {
	AsynchronousBaseResponse
}

//DeregisterInstancesFromForwardLBSeventhListenerInput deregister 7 layer target info
type DeregisterInstancesFromForwardLBSeventhListenerInput struct {
	APIMeta       `url:",inline"`
	Backends      BackendTargetList
	ListenerID    string `url:"listenerId,omitempty"`
	LoadBalanceID string `url:"loadBalancerId,omitempty"`
	LocationID    string `url:"locationIds.0,omitempty"`
}

//DeregisterInstancesFromForwardLBSeventhListenerOutput deregister 7 layer result
type DeregisterInstancesFromForwardLBSeventhListenerOutput struct {
	AsynchronousBaseResponse
}

//ModifyForwardFourthBackendsInput modify 4 layer backends weight input
type ModifyForwardFourthBackendsInput struct {
	LoadbalanceID string `json:"loadBalancerId"`
	ListenerID    string `json:"listenerId"`
	Backends      BackendTargetList
}

//ModifyForwardFourthBackendsOutput modify 4 layer backends weight output
type ModifyForwardFourthBackendsOutput struct {
	AsynchronousBaseResponse
}

//ModifyForwardSeventhBackendsInput modify 7 layer backends weight input
type ModifyForwardSeventhBackendsInput struct {
	LoadbalanceID string `json:"loadBalancerId"`
	ListenerID    string `json:"listenerId"`
	LocationID    string `json:"locationIds.0"`
	Backends      BackendTargetList
}

//ModifyForwardSeventhBackendsOutput modify 7 layer backends weight output
type ModifyForwardSeventhBackendsOutput struct {
	AsynchronousBaseResponse
}

//AsynchronousBaseResponse base async response
type AsynchronousBaseResponse struct {
	Response  `json:",inline"`
	RequestID int `json:"requestId"`
}

//DescribeLoadBalancersTaskResultInput describe lb task result input
type DescribeLoadBalancersTaskResultInput struct {
	APIMeta   `url:",inline"`
	RequestID int `url:"requestId,omitempty"`
}

//DescribeLoadBalancersTaskResultOutput describe lb task result output
type DescribeLoadBalancersTaskResultOutput struct {
	Response `json:",inline"`
	Data     TaskStatus `json:"data"`
}

//TaskStatus task status of lb task
type TaskStatus struct {
	Status int `json:"status"`
}

//CreateForwardLBListenerRulesInput input of create forward loadbalance listener
type CreateForwardLBListenerRulesInput struct {
	APIMeta       `url:",inline"`
	ListenerID    string `url:"listenerId,omitempty"`
	LoadBalanceID string `url:"loadBalancerId,omitempty"`
	Rules         RuleCreateInfoList
	// RuleDomain        string `url:"rules.0.domain,omitempty"`
	// RuleHealthSwitch  int    `url:"rules.0.healthSwitch,omitempty"`
	// RuleHTTPCheckPath string `url:"rules.0.httpCheckPath,omitempty"`
	// RuleHTTPHash      string `url:"rules.0.httpHash,omitempty"`
	// RuleSessionExpire int    `url:"rules.0.sessionExpire,omitempty"`
	// RuleURL           string `url:"rules.0.url,omitempty"`
}

//RuleCreateInfo rule info for creation
type RuleCreateInfo struct {
	RuleDomain        string
	RuleURL           string
	RuleHealthSwitch  int
	RuleHTTPCheckPath string
	RuleHTTPHash      string
	RuleHTTPCode      int
	RuleIntervalTime  int
	RuleHealthNum     int
	RuleUnhealthNum   int
	RuleSessionExpire int
}

//RuleCreateInfoList rule list
type RuleCreateInfoList []RuleCreateInfo

//EncodeValues encode rule info list
func (list RuleCreateInfoList) EncodeValues(key string, urlv *url.Values) error {
	for i, v := range list {
		urlv.Set(fmt.Sprintf("rules.%d.domain", i), fmt.Sprintf("%v", v.RuleDomain))
		urlv.Set(fmt.Sprintf("rules.%d.url", i), fmt.Sprintf("%v", v.RuleURL))
		urlv.Set(fmt.Sprintf("rules.%d.healthSwitch", i), fmt.Sprintf("%v", v.RuleHealthSwitch))
		if len(v.RuleHTTPCheckPath) != 0 {
			urlv.Set(fmt.Sprintf("rules.%d.httpCheckPath", i), fmt.Sprintf("%v", v.RuleHTTPCheckPath))
		}
		if len(v.RuleHTTPHash) != 0 {
			urlv.Set(fmt.Sprintf("rules.%d.httpHash", i), fmt.Sprintf("%v", v.RuleHTTPHash))
		}
		urlv.Set(fmt.Sprintf("rules.%d.intervalTime", i), fmt.Sprintf("%v", v.RuleIntervalTime))
		urlv.Set(fmt.Sprintf("rules.%d.healthNum", i), fmt.Sprintf("%v", v.RuleHealthNum))
		urlv.Set(fmt.Sprintf("rules.%d.unhealthNum", i), fmt.Sprintf("%v", v.RuleUnhealthNum))
		urlv.Set(fmt.Sprintf("rules.%d.httpCode", i), fmt.Sprintf("%v", v.RuleHTTPCode))
		urlv.Set(fmt.Sprintf("rules.%d.sessionExpire", i), fmt.Sprintf("%v", v.RuleSessionExpire))
	}
	return nil
}

//CreateForwardLBListenerRulesOutput create forward lb listener rule result
type CreateForwardLBListenerRulesOutput struct {
	AsynchronousBaseResponse
}

//DeleteForwardLBListenerRulesInput input of delete loadbalance listener rule
type DeleteForwardLBListenerRulesInput struct {
	APIMeta        `url:",inline"`
	LoadBalanceID  string `url:"loadBalancerId"`
	ListenerID     string `url:"listenerId"`
	LocationIDList RuleIDList
	Domain         string `url:"domain,omitempty"`
	Url            string `url:"url,omitempty"`
}

//RuleID rule id
type RuleID string

//RuleIDList rule id list
type RuleIDList []RuleID

//EncodeValues encode rule id list
func (list RuleIDList) EncodeValues(key string, urlv *url.Values) error {
	for i, v := range list {
		urlv.Set(fmt.Sprintf("locationIds.%d", i), string(v))
	}
	return nil
}

//DeleteForwardLBListenerRulesOutput output of delete loadbalance listener rule
type DeleteForwardLBListenerRulesOutput struct {
	AsynchronousBaseResponse
}

//CreateSecurityGroupInput info to create security group
type CreateSecurityGroupInput struct {
	APIMeta   `url:",inline"`
	SgName    string `url:"sgName,omitempty"`
	SgRemark  string `url:"sgRemark,omitempty"`
	ProjectID int    `url:"projectId,omitempty"`
}

//CreateSecurityGroupOutput create security group api response
type CreateSecurityGroupOutput struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	SgInfo  SecureGroupInfo `json:"data"`
}

//SecureGroupInfo security group info
type SecureGroupInfo struct {
	SgID     string `json:"sgId"`
	SgName   string `json:"sgName"`
	SgRemark string `json:"sgRemark"`
}

//DescribeSecurityGroupPolicysInput describe security group policys info
type DescribeSecurityGroupPolicysInput struct {
	APIMeta `url:",inline"`
	SgID    string `url:"sgId,omitempty"`
}

//DescribeSecurityGroupPolicysOutput get describe security group policy result
type DescribeSecurityGroupPolicysOutput struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    IngressEgressData `json:"data"`
}

//IngressEgressData include ingress and egress information
type IngressEgressData struct {
	SgID         string                   `json:"sgId"`
	IngressInfos []SecurityGroupGressInfo `json:"ingress"`
	EgressInfos  []SecurityGroupGressInfo `json:"egress"`
}

//SecurityGroupGressInfo ingress/egress information
type SecurityGroupGressInfo struct {
	Index         int    `json:"index"`
	AddressModule string `json:"addressModule"`
	IPProtocol    string `json:"ipProtocol"`
	CidrIP        string `json:"cidrIp"`
	SgID          string `json:"sgId"`
	PortRange     string `json:"portRange"`
	ServiceModule string `json:"serviceModule"`
	Desc          string `json:"desc"`
	Action        string `json:"action"`
	Version       int    `json:"version"`
}

//CreateSecurityGroupPolicyInput create security group policy input info
type CreateSecurityGroupPolicyInput struct {
	APIMeta             `url:",inline"`
	Direction           string `url:"direction,omitempty"`
	Index               string `url:"index,omitempty"`
	PolicyAction        string `url:"policys.0.action,omitempty"`
	PolicyAddressModule string `url:"policys.0.addressModule,omitempty"`
	PolicyCidrIP        string `url:"policys.0.cidrIp,omitempty"`
	PolicyDesc          string `url:"policys.0.desc,omitempty"`
	PolicyIPProtocol    string `url:"policys.0.ipProtocol,omitempty"`
	PolicyPortRange     string `url:"policys.0.portRange,omitempty"`
	PolicyServiceModule string `url:"policys.0.serviceModule,omitempty"`
	PolicySgID          string `url:"policys.0.sgId,omitempty"`
	SgID                string `url:"sgId,omitempty"`
}

//CreateSecurityGroupPolicyOutput create security group policy result
type CreateSecurityGroupPolicyOutput struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//ModifySingleSecurityGroupPolicyInput modify security group policy input
type ModifySingleSecurityGroupPolicyInput struct {
	APIMeta          `url:",inline"`
	Direction        string `url:"direction,omitempty"`
	Index            int    `url:"index"`
	PolicyAction     string `url:"policys.action,omitempty"`
	PolicyCidrIp     string `url:"policys.cidrIp,omitempty"`
	PolicyDesc       string `url:"policys.desc,omitempty"`
	PolicyIpProtocol string `url:"policys.ipProtocol,omitempty"`
	PolicyPortRange  string `url:"policys.portRange,omitempty"`
	SgId             string `url:"sgId,omitempty"`
}

//ModifySingleSecurityGroupPolicyOutput modify security group policy output
type ModifySingleSecurityGroupPolicyOutput struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//DeleteForwardLBListenerInput delete listener api input
type DeleteForwardLBListenerInput struct {
	APIMeta       `url:",inline"`
	ListenerID    string `url:"listenerId,omitempty"`
	LoadBalanceID string `url:"loadBalancerId,omitempty"`
}

//DeleteForwardLBListenerOutput delete listener api output
type DeleteForwardLBListenerOutput struct {
	AsynchronousBaseResponse
}

//ModifySecurityGroupsOfInstanceInput modify security group of cvm instance
//doc: https://cloud.tencent.com/document/api/213/9381
//SecurityGroups1 string `url:"SecurityGroups.0,omitempty"`
//SecurityGroups2 string `url:"SecurityGroups.1,omitempty"`
type ModifySecurityGroupsOfInstanceInput struct {
	APIMeta        `url:",inline"`
	InstanceID     string `url:"InstanceIds.0,omitempty"`
	SecurityGroups SecurityGroupList
	Version        string `url:"Version,omitempty"`
}

//SecurityGroupList security group list
type SecurityGroupList []string

//EncodeValues encode security group list info into url format
func (sgs SecurityGroupList) EncodeValues(key string, urlv *url.Values) error {
	for i, v := range sgs {
		primary := fmt.Sprintf("SecurityGroups.%d", i)
		urlv.Set(primary, fmt.Sprintf("%s", v))
	}
	return nil
}

//DescribeForwardLBBackendsInput list clb targets
type DescribeForwardLBBackendsInput struct {
	APIMeta          `url:",inline"`
	ListenerID       string `url:"listenerIds.0,omitempty"`
	LoadBalanceID    string `url:"loadBalancerId,omitempty"`
	LoadBalancerPort int    `url:"loadBalancerPort,omitempty"`
	Protocol         int    `url:"protocol,omitempty"`
}

//DescribeForwardLBBackendsOutput describe forward lb backend response
type DescribeForwardLBBackendsOutput struct {
	Response
	Data []ListenerDetail `json:"data"`
}

//CommonInfo common information
type CommonInfo struct {
	LoadBalancerPort int    `json:"loadBalancerPort"`
	Protocol         int    `json:"protocol"`
	ProtocolType     string `json:"protocolType"`
	ListenerID       string `json:"listenerId"`
}

//ListenerDetail one listener datail
type ListenerDetail struct {
	CommonInfo
	Rules    []RuleDetail     `json:"rules,omitempty"`
	Backends []InstanceDetail `json:"backends,omitempty"`
}

//RuleDetail describe one listener rule info
type RuleDetail struct {
	LocationID string           `json:"locationId"`
	Domain     string           `json:"domain"`
	URL        string           `json:"url"`
	Backends   []InstanceDetail `json:"backends"`
}

//InstanceDetail one backend detail information
type InstanceDetail struct {
	InstanceName   string   `json:"instanceName"`
	LanIP          string   `json:"lanIp"`
	WanIPSet       []string `json:"wanIpSet"`
	InstanceStatus int      `json:"instanceStatus"`
	Port           int      `json:"port"`
	Weight         int      `json:"weight"`
	UnInstanceID   string   `json:"unInstanceId"`
	UUID           string   `json:"uuid"`
	AddTimestamp   string   `json:"addTimestamp"`
}
