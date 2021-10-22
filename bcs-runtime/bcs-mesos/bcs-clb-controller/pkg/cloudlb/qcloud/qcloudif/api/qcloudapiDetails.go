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
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	loadbalance "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"

	qcloud "github.com/Tencent/bk-bcs/bcs-common/pkg/qcloud/clbv2"
)

func (clb *ClbAPI) waitForTaskResult(id int) error {
	count := 0
	for ; count <= ClbMaxTimeout; count++ {
		status, err := clb.describeLoadBalancersTaskResult(id)
		if err != nil {
			blog.Errorf("describe task %d result failed, err %s", id, err.Error())
			return fmt.Errorf("describe task %d result failed, err %s", id, err.Error())
		}
		if status == TaskResultStatusDealing {
			blog.Warn("clb is dealing")
			time.Sleep(time.Duration(clb.WaitPeriodLBDealing) * time.Second)
			continue
		}
		if status != TaskResultStatusSuccess {
			return fmt.Errorf("describe clb task result failed, invalid status code %d", status)
		}
		blog.Infof("clb task %d is done", id)
		return nil
	}
	blog.Errorf("wait for task %d result timeout", id)
	return fmt.Errorf("wait for task %d result timeout", id)
}

//describeLoadBalancersTaskResult query asynchronous clb api result
//status 1 for failed, 0 for successful, 2 for dealing
func (clb *ClbAPI) describeLoadBalancersTaskResult(requestID int) (int, error) {
	desc := new(qcloud.DescribeLoadBalancersTaskResultInput)
	desc.Action = "DescribeLoadBalancersTaskResult"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.RequestID = requestID

	output, err := clb.api.DescribeLoadBalanceTaskResult(desc)
	if err != nil {
		return TaskResultStatusFailed, fmt.Errorf("describe clb task result failed, input %v, err %s",
			desc, err.Error())
	}

	//when request exceeds limits, should have a rest, we treated it as dealing status but with sleeping 5 seconds more
	/*
		{
			"code":4400,
			"message":"请求超过限额配置，请稍后重试。",
			"codeDesc":"RequestLimitExceeded"
		}
	*/
	if output.Code == RequestLimitExceededCode && output.CodeDesc == RequestLimitExceededMessage {
		blog.Warn("clb request exceed limit, need to have a rest")
		time.Sleep(time.Duration(clb.WaitPeriodExceedLimit) * time.Second)
		return TaskResultStatusDealing, nil
	}
	/*
		{
			"code": 4000,
			"message": "(12003)该负载均衡在执行其他操作",
			"codeDesc": "IncorrectStatus.LBWrongStatus"
		}
	*/
	if output.Code == WrongStatusCode && output.CodeDesc == WrongStatusMessage {
		blog.Warn("clb request lb busy status, lb is dealing another action")
		return TaskResultStatusDealing, nil
	}
	if output.Code != 0 {
		blog.Errorf("describe clb task result returned code %d invalid, msg %s", output.Code, output.Message)
		return TaskResultStatusFailed, fmt.Errorf("DescribeLoadBalancersTaskResultOutput invalid")
	}

	blog.Infof("clb task %d done", requestID)
	return output.Data.Status, nil
}

//describeLoadBalance
//return (nil, nil) when lb does not exit bu no error happened
//**CAUTION** DescribeLoadBalancers for application lb must set Forward to 1, default Forward is 0
func (clb *ClbAPI) doDescribeLoadBalance(name string) (*qcloud.DescribeLBOutput, error) {
	desc := new(qcloud.DescribeLBInput)
	desc.Action = "DescribeLoadBalancers"
	desc.Forward = ClbApplicationType
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.LoadBalanceName = name

	output, err := clb.api.DescribeLoadBalance(desc)
	if err != nil {
		blog.Errorf("describeLoadBalance %s request failed, %s", name, err.Error())
		return nil, fmt.Errorf("describeLoadBalance %s request failed, %s", name, err.Error())
	}
	if output.Code != 0 && output.Code != 5000 {
		blog.Errorf("describeLoadBalance %s failed, invalid code %d, code desc %s", name, output.Code, output.CodeDesc)
		return output, fmt.Errorf("describeLoadBalance %s failed, invalid code %d, code desc %s",
			name, output.Code, output.CodeDesc)
	} else if output.Code == 5000 {
		blog.Warnf("DescribeLoadBalancer lb %s does not exist", name)
		return output, nil
	}

	blog.Infof("describeLoadBalance done")
	return output, nil
}

//create7LayerListener
func (clb *ClbAPI) create7LayerListener(listener *loadbalance.CloudListener) (string, error) {
	desc := new(qcloud.CreateSeventhLayerListenerInput)
	desc.Action = "CreateForwardLBSeventhLayerListeners"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	if listener.Spec.Protocol == loadbalance.ClbListenerProtocolHTTPS {
		if listener.Spec.TLS == nil {
			return "", fmt.Errorf("tls config is necessary for https listener")
		}
		desc.ListenersSSLMode = listener.Spec.TLS.Mode
		desc.ListenersCertID = listener.Spec.TLS.CertID
		desc.ListenersCertCaID = listener.Spec.TLS.CertCaID
		desc.ListenersCertCaContent = listener.Spec.TLS.CertClientCaContent
		desc.ListenersCertCaName = listener.Spec.TLS.CertClientCaName
		desc.ListenersCertContent = listener.Spec.TLS.CertServerContent
		desc.ListenersCertKey = listener.Spec.TLS.CertServerKey
		desc.ListenersCertName = listener.Spec.TLS.CertServerName
	}
	desc.ListenersListenerName = listener.GetName()
	desc.ListenersLoadBalancerPort = listener.Spec.ListenPort
	protocol, _ := ProtocolTypeBcs2QCloudMap[listener.Spec.Protocol]
	desc.ListenersProtocol = protocol
	desc.LoadBalanceID = listener.Spec.LoadBalancerID
	count := 0
	for ; count <= ClbMaxTimeout; count++ {
		output, err := clb.api.Create7LayerListener(desc)
		if err != nil {
			return "", fmt.Errorf("create 7 layer listener name %s protocol %d port %d failed, err %s",
				desc.ListenersListenerName, desc.ListenersProtocol, desc.ListenersLoadBalancerPort, err.Error())
		}
		if output.Code == WrongStatusCode && output.CodeDesc == WrongStatusMessage {
			blog.Warn("LB is dealing another action will wait a second")
			time.Sleep(time.Duration(clb.WaitPeriodLBDealing) * time.Second)
			continue
		}
		if output.Code == RequestLimitExceededCode && output.CodeDesc == RequestLimitExceededMessage {
			time.Sleep(time.Duration(clb.WaitPeriodExceedLimit) * time.Second)
			blog.Errorf("clb request exceed limit, create 7 layer listener failed")
			return "", fmt.Errorf("clb request exceed limit, create 7 layer listener failed")
		}
		if output.Code != 0 {
			blog.Errorf("create 7 layer listener (name %s,  protocol %d,  port %d) failed, code %d, code desc %s",
				desc.ListenersListenerName, desc.ListenersProtocol,
				desc.ListenersLoadBalancerPort, output.Code, output.CodeDesc)
			return "", fmt.Errorf(
				"create 7 layer listener (name %s,  protocol %d,  port %d) failed, code %d, code desc %s",
				desc.ListenersListenerName, desc.ListenersProtocol,
				desc.ListenersLoadBalancerPort, output.Code, output.CodeDesc)
		}
		if len(output.ListenerIds) != 1 {
			blog.Errorf("create 7 layer listener (name %s,  protocol %d,  port %d) failed, invalid ids length %d",
				desc.ListenersListenerName, desc.ListenersProtocol,
				desc.ListenersLoadBalancerPort, len(output.ListenerIds))
			return "", fmt.Errorf(
				"create 7 layer listener (name %s,  protocol %d,  port %d) failed, invalid ids length %d",
				desc.ListenersListenerName, desc.ListenersProtocol,
				desc.ListenersLoadBalancerPort, len(output.ListenerIds))
		}

		blog.Infof("create 7 layer listener (name %s,  protocol %d,  port %d) successfully",
			desc.ListenersListenerName, desc.ListenersProtocol, desc.ListenersLoadBalancerPort)
		return output.ListenerIds[0], nil
	}
	return "", fmt.Errorf("create 7 layer listener (name %s,  protocol %d,  port %d) timeout",
		desc.ListenersListenerName, desc.ListenersProtocol, desc.ListenersLoadBalancerPort)
}

//create4LayerListener
func (clb *ClbAPI) create4LayerListener(listener *loadbalance.CloudListener) (string, error) {
	desc := new(qcloud.CreateForwardLBFourthLayerListenersInput)
	desc.Action = "CreateForwardLBFourthLayerListeners"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.ListenersListenerName = listener.GetName()
	desc.ListenersLoadBalancerPort = listener.Spec.ListenPort
	desc.LoadBalanceID = listener.Spec.LoadBalancerID
	//we will validate the field in upper function
	protocol, _ := ProtocolTypeBcs2QCloudMap[listener.Spec.Protocol]
	desc.ListenersProtocol = protocol
	if listener.Spec.TargetGroup != nil {
		desc.ListenerExpireTime = listener.Spec.TargetGroup.SessionExpire
		desc.ListenerScheduler = listener.Spec.TargetGroup.LBPolicy
		if listener.Spec.TargetGroup.HealthCheck != nil {
			desc.ListenerHealthSwitch = listener.Spec.TargetGroup.HealthCheck.Enabled
			desc.ListenerIntervalTime = listener.Spec.TargetGroup.HealthCheck.IntervalTime
			desc.ListenerTimeout = listener.Spec.TargetGroup.HealthCheck.Timeout
			desc.ListenerHealthNum = listener.Spec.TargetGroup.HealthCheck.HealthNum
			desc.ListenerUnHealthNum = listener.Spec.TargetGroup.HealthCheck.UnHealthNum
		}
	}

	count := 0
	var err error
	var output *qcloud.CreateForwardLBFourthLayerListenersOutput
	for ; count <= ClbMaxTimeout; count++ {
		if count == ClbMaxTimeout {
			blog.Errorf("create 4 layer listener (name %s,  protocol %d,  port %d) timeout",
				desc.ListenersListenerName, desc.ListenersProtocol, desc.ListenersLoadBalancerPort)
			return "", fmt.Errorf("create 4 layer listener (name %s,  protocol %d,  port %d) timeout",
				desc.ListenersListenerName, desc.ListenersProtocol, desc.ListenersLoadBalancerPort)
		}
		output, err = clb.api.Create4LayerListener(desc)
		if err != nil {
			return "", fmt.Errorf("create 4 layer listener (name %s,  protocol %d,  port %d) failed, err %s",
				desc.ListenersListenerName, desc.ListenersProtocol, desc.ListenersLoadBalancerPort, err.Error())
		}
		if output.Code == WrongStatusCode && output.CodeDesc == WrongStatusMessage {
			blog.Warn("LB is dealing another action will wait a second")
			time.Sleep(time.Duration(clb.WaitPeriodLBDealing) * time.Second)
			continue
		}
		if output.Code == RequestLimitExceededCode && output.CodeDesc == RequestLimitExceededMessage {
			time.Sleep(time.Duration(clb.WaitPeriodExceedLimit) * time.Second)
			blog.Errorf("clb request exceed limit, create 4 layer listener failed")
			return "", fmt.Errorf("clb request exceed limit, create 4 layer listener failed")
		}
		if output.Code != 0 {
			blog.Errorf("create 4 layer listener (name %s,  protocol %d,  port %d) failed, code %d, code desc %s",
				desc.ListenersListenerName, desc.ListenersProtocol,
				desc.ListenersLoadBalancerPort, output.Code, output.CodeDesc)
			return "", fmt.Errorf(
				"create 4 layer listener (name %s,  protocol %d,  port %d) failed, code %d, code desc %s",
				desc.ListenersListenerName, desc.ListenersProtocol,
				desc.ListenersLoadBalancerPort, output.Code, output.CodeDesc)
		}
		if len(output.ListenerIds) != 1 {
			blog.Errorf("create 4 layer listener (name %s,  protocol %d,  port %d) failed, invalid ids length %d",
				desc.ListenersListenerName, desc.ListenersProtocol,
				desc.ListenersLoadBalancerPort, len(output.ListenerIds))
			return "", fmt.Errorf(
				"create 4 layer listener (name %s,  protocol %d,  port %d) failed, invalid ids length %d",
				desc.ListenersListenerName, desc.ListenersProtocol,
				desc.ListenersLoadBalancerPort, len(output.ListenerIds))
		}
		blog.Infof("create 4 layer listener (name %s,  protocol %d,  port %d) request done",
			desc.ListenersListenerName, desc.ListenersProtocol, desc.ListenersLoadBalancerPort)
		break
	}

	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return "", fmt.Errorf("wait for task result failed, err %s", err.Error())
	}

	return output.ListenerIds[0], nil
}

//doDescribeListener describe clb listener
//return (nil, nil) when listener does not existed
func (clb *ClbAPI) doDescribeListener(loadBalanceID, listenerID string) (*qcloud.ListenerInfo, error) {
	desc := new(qcloud.DescribeListenerInput)
	desc.Action = "DescribeForwardLBListeners"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.ListenerID = listenerID
	desc.LoadBalanceID = loadBalanceID

	output, err := clb.api.DescribeListener(desc)
	if err != nil {
		return nil, fmt.Errorf("describe clb %s listener %s failed, err %s", loadBalanceID, listenerID, err.Error())
	}
	if output.Code != 0 && output.Code != 5000 {
		blog.Errorf("describe clb %s listener %s returned code %d invalid:%s",
			loadBalanceID, listenerID, output.Code, output.CodeDesc)
		return nil, fmt.Errorf("describe clb %s listener %s returned code %d invalid:%s",
			loadBalanceID, listenerID, output.Code, output.CodeDesc)
	} else if output.Code == 5000 {
		blog.Warnf("described clb %s listener %s is not existed", loadBalanceID, listenerID)
		return nil, nil
	}
	if len(output.Listeners) == 0 {
		blog.Warnf("described clb %s listener %s return zero result length", loadBalanceID, listenerID)
		return nil, nil
	} else if len(output.Listeners) != 1 {
		blog.Errorf("described clb %s listener %s return invalid result length %d",
			loadBalanceID, listenerID, len(output.Listeners))
		return nil, fmt.Errorf("described clb %s listener %s return invalid result length %d",
			loadBalanceID, listenerID, len(output.Listeners))
	}
	blog.Infof("describe clb %s listener %s done", loadBalanceID, listenerID)
	return &output.Listeners[0], nil
}

//doDescribeListenerByPort describe listener by port
func (clb *ClbAPI) doDescribeListenerByPort(loadBalanceID string, port int) (*qcloud.ListenerInfo, error) {
	desc := new(qcloud.DescribeListenerInput)
	desc.Action = "DescribeForwardLBListeners"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.LoadBalanceID = loadBalanceID
	desc.LoadBalancerPort = port

	output, err := clb.api.DescribeListener(desc)
	if err != nil {
		blog.Errorf("describe clb %s listener by port %d failed, err %s", loadBalanceID, port, err.Error())
		return nil, fmt.Errorf("describe clb %s listener by port %d failed, err %s", loadBalanceID, port, err.Error())
	}
	if output.Code != 0 && output.Code != 5000 {
		blog.Errorf("describe clb %s listener by port %d, return code %d invalid, code desc %s",
			loadBalanceID, port, output.Code, output.CodeDesc)
		return nil, fmt.Errorf("describe clb %s listener by port %d, return code %d invalid, code desc %s",
			loadBalanceID, port, output.Code, output.CodeDesc)
	} else if output.Code == 5000 {
		return nil, nil
	}
	if len(output.Listeners) == 0 {
		blog.Warnf("described clb %s listener by port %d return zero result length", loadBalanceID, port)
		return nil, nil
	} else if len(output.Listeners) != 1 {
		blog.Errorf("described clb %s listener by port %d return invalid result length %d",
			loadBalanceID, port, len(output.Listeners))
		return nil, fmt.Errorf("described clb %s listener by port %d return invalid result length %d",
			loadBalanceID, port, len(output.Listeners))
	}
	blog.Infof("describe clb %s listener by port %d done", loadBalanceID, port)

	return &output.Listeners[0], nil
}

//doDeleteListener delete listener
func (clb *ClbAPI) doDeleteListener(loadBalanceID, listenerID string) error {
	desc := new(qcloud.DeleteForwardLBListenerInput)
	desc.Action = "DeleteForwardLBListener"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.ListenerID = listenerID
	desc.LoadBalanceID = loadBalanceID
	blog.Infof("start delete listener %s", listenerID)

	output, err := clb.api.DeleteListener(desc)
	if err != nil {
		return fmt.Errorf("request delete listener failed, err %s", err.Error())
	}
	if output.Code != 0 {
		return fmt.Errorf("delete listener failed, code %d, message %s", output.Code, output.Message)
	}

	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return fmt.Errorf("wait for task result %d failed, err %s", output.RequestID, err.Error())
	}

	blog.Infof("DeleteListener listenerId %s success", listenerID)
	return nil
}

func (clb *ClbAPI) doModify7LayerListenerAttribute(listener *loadbalance.CloudListener) error {
	desc := new(qcloud.ModifyForwardLBSeventhListenerInput)
	desc.Action = "ModifyForwardLBSeventhListener"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.LoadBalanceID = listener.Spec.LoadBalancerID
	desc.ListenerID = listener.Spec.ListenerID

	desc.SSLMode = listener.Spec.TLS.Mode
	desc.CertID = listener.Spec.TLS.CertID
	desc.CertCaID = listener.Spec.TLS.CertCaID
	desc.CertCaContent = listener.Spec.TLS.CertClientCaContent
	desc.CertCaName = listener.Spec.TLS.CertClientCaName
	desc.CertContent = listener.Spec.TLS.CertServerContent
	desc.CertKey = listener.Spec.TLS.CertServerKey
	desc.CertName = listener.Spec.TLS.CertServerName

	blog.Infof("start modify 7 layer listener with %v", desc)
	output, err := clb.api.Modify7LayerListener(desc)
	if err != nil {
		blog.Errorf("modify 7 layer listener attr failed, err %s", err.Error())
		return fmt.Errorf("modify 7 layer listener attr failed, err %s", err.Error())
	}
	if output.Code != 0 {
		return fmt.Errorf("modify 7 layer listener failed, code %d, message %s", output.Code, output.Message)
	}
	blog.Infof("modify 7 layer listener done, wait for task result")
	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return fmt.Errorf("wait for result failed, err %s", err.Error())
	}
	blog.Infof("modify 7 layer listener %s with port %d of %s successfully",
		listener.Spec.ListenerID, listener.Spec.ListenPort, listener.Spec.LoadBalancerID)

	return nil
}

func (clb *ClbAPI) doModify4LayerListenerAttribute(listener *loadbalance.CloudListener) error {
	desc := new(qcloud.ModifyForwardLBFourthListenerInput)
	desc.Action = "ModifyForwardLBFourthListener"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.LoadBalanceID = listener.Spec.LoadBalancerID
	desc.ListenerID = listener.Spec.ListenerID
	desc.SessionExpire = listener.Spec.TargetGroup.SessionExpire
	scheduler, _ := LBAlgorithmTypeBcs2QCloudMap[listener.Spec.TargetGroup.LBPolicy]
	desc.Scheduler = scheduler
	desc.HealthSwitch = listener.Spec.TargetGroup.HealthCheck.Enabled
	if desc.HealthSwitch == HealthSwitchOn {
		desc.Timeout = listener.Spec.TargetGroup.HealthCheck.Timeout
		desc.IntervalTime = listener.Spec.TargetGroup.HealthCheck.IntervalTime
		desc.HealthNum = listener.Spec.TargetGroup.HealthCheck.HealthNum
		desc.UnHealthNum = listener.Spec.TargetGroup.HealthCheck.UnHealthNum
	}

	blog.Infof("start modify 4 layer listener with %v", desc)
	output, err := clb.api.Modify4LayerListener(desc)
	if err != nil {
		blog.Errorf("modify 4 layer listener attr failed, err %s", err.Error())
		return fmt.Errorf("modify 4 layer listener attr failed, err %s", err.Error())
	}
	if output.Code != 0 {
		return fmt.Errorf("modify 4 layer listener failed, code %d, message %s", output.Code, output.Message)
	}
	blog.Infof("modify 4 layer listener done, wait for task result")
	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return fmt.Errorf("wait for result failed, err %s", err.Error())
	}
	blog.Infof("modify 4 layer listener %s with port %d of %s successfully",
		listener.Spec.ListenerID, listener.Spec.ListenPort, listener.Spec.LoadBalancerID)

	return nil
}

//doCreateRule create rule
func (clb *ClbAPI) doCreateRules(loadBalanceID, listenerID string, rules loadbalance.RuleList) error {
	desc := new(qcloud.CreateForwardLBListenerRulesInput)
	desc.Action = "CreateForwardLBListenerRules"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.ListenerID = listenerID
	desc.LoadBalanceID = loadBalanceID

	var ruleCreateList qcloud.RuleCreateInfoList
	for _, rule := range rules {
		//**CAUTION** domains like lol.qq.com:8080 is invalid, domain name should not contain port info
		var domain string
		if strings.Contains(rule.Domain, ":") {
			validDomains := strings.Split(rule.Domain, ":")
			domain = validDomains[0]
		} else {
			domain = rule.Domain
		}
		newRuleCreateInfo := qcloud.RuleCreateInfo{}
		if len(rule.TargetGroup.LBPolicy) != 0 {
			httpHash, _ := LBAlgorithmTypeBcs2QCloudMap[rule.TargetGroup.LBPolicy]
			newRuleCreateInfo.RuleHTTPHash = httpHash
		} else {
			newRuleCreateInfo.RuleHTTPHash = LBAlgorithmRoundRobin
		}
		newRuleCreateInfo.RuleDomain = domain
		newRuleCreateInfo.RuleURL = rule.URL
		newRuleCreateInfo.RuleHealthSwitch = rule.TargetGroup.HealthCheck.Enabled
		newRuleCreateInfo.RuleHTTPCheckPath = rule.TargetGroup.HealthCheck.HTTPCheckPath
		newRuleCreateInfo.RuleIntervalTime = rule.TargetGroup.HealthCheck.IntervalTime
		newRuleCreateInfo.RuleHealthNum = rule.TargetGroup.HealthCheck.HealthNum
		newRuleCreateInfo.RuleUnhealthNum = rule.TargetGroup.HealthCheck.UnHealthNum
		newRuleCreateInfo.RuleHTTPCode = rule.TargetGroup.HealthCheck.HTTPCode
		newRuleCreateInfo.RuleSessionExpire = rule.TargetGroup.SessionExpire
		ruleCreateList = append(ruleCreateList, newRuleCreateInfo)
	}
	desc.Rules = ruleCreateList

	blog.Infof("start create rules %v for listener %s", ruleCreateList, listenerID)
	output, err := clb.api.CreateRules(desc)
	if err != nil {
		return fmt.Errorf("create rules failed, err %s", err.Error())
	}
	if output.Code != 0 {
		return fmt.Errorf("create rules failed, code %d, message %s", output.Code, output.Message)
	}
	blog.Infof("request create rules done, wait for task result")

	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return fmt.Errorf("wait for result failed, err %s", err.Error())
	}
	return nil
}

//doDeleteRule delete rules
func (clb *ClbAPI) doDeleteRule(loadBalanceID, listenerID, domain, url string) error {
	desc := new(qcloud.DeleteForwardLBListenerRulesInput)
	desc.Action = "DeleteForwardLBListenerRules"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.ListenerID = listenerID
	desc.LoadBalanceID = loadBalanceID
	desc.Domain = domain
	desc.Url = url

	blog.Infof("start delete path %s, domain %s for listener %s",
		url, domain, listenerID)
	output, err := clb.api.DeleteRules(desc)
	if err != nil {
		return fmt.Errorf("delete rules failed, err %s", err.Error())
	}
	if output.Code != 0 {
		return fmt.Errorf("delete rules failed, code %d, message %s", output.Code, output.Message)
	}
	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return fmt.Errorf("wait for task %d result failed, err %s", output.RequestID, err.Error())
	}
	blog.Infof("delete rule with domain %s, path %s of listener %s successfully", domain, url, listenerID)
	return nil
}

//doModifyRule()
func (clb *ClbAPI) doModifyRule(loadBalanceID, listenerID string, rule *loadbalance.Rule) error {
	desc := new(qcloud.ModifyLoadBalancerRulesProbeInput)
	desc.Action = "ModifyLoadBalancerRulesProbe"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.ListenerID = listenerID
	desc.LoadBalanceID = loadBalanceID
	desc.LocationID = rule.ID
	desc.SessionExpire = rule.TargetGroup.SessionExpire
	desc.HealthSwitch = rule.TargetGroup.HealthCheck.Enabled
	httpHash, _ := LBAlgorithmTypeQCloud2BcsMap[rule.TargetGroup.LBPolicy]
	desc.HTTPHash = httpHash
	desc.Timeout = rule.TargetGroup.HealthCheck.Timeout
	desc.IntervalTime = rule.TargetGroup.HealthCheck.IntervalTime
	desc.HealthNum = rule.TargetGroup.HealthCheck.HealthNum
	desc.UnHealthNum = rule.TargetGroup.HealthCheck.UnHealthNum
	desc.HTTPCode = rule.TargetGroup.HealthCheck.HTTPCode
	desc.HTTPCheckPath = rule.TargetGroup.HealthCheck.HTTPCheckPath

	blog.Infof("start modify rule with %v", desc)
	output, err := clb.api.ModifyRuleProbe(desc)
	if err != nil {
		return fmt.Errorf("modify rule failed, err %s", err.Error())
	}
	if output.Code != 0 {
		return fmt.Errorf("modify rule failed, code %d, message %s", output.Code, output.Message)
	}
	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return fmt.Errorf("wait for task %d result failed, err %s", output.RequestID, err.Error())
	}
	blog.Infof("modify rule (lb %s, listener %s, domain %s, path %s) successfully",
		loadBalanceID, listenerID, rule.Domain, rule.URL)

	return nil
}

//registerInsWith7thLayerListener
func (clb *ClbAPI) registerInsWith7thLayerListener(
	loadBalanceID, listenerID, locationID string, bdTargets qcloud.BackendTargetList) error {
	desc := new(qcloud.RegisterInstancesWithForwardLBSeventhListenerInput)
	desc.Action = "RegisterInstancesWithForwardLBSeventhListener"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.Backends = bdTargets
	desc.ListenerID = listenerID
	desc.LoadBalanceID = loadBalanceID
	desc.LocationID = locationID

	blog.Infof("start register instance with 7 layer listener %v for listener %s rule %s",
		bdTargets, listenerID, locationID)

	output, err := clb.api.RegInstancesWith7LayerListener(desc)
	if err != nil {
		return fmt.Errorf("register instance with 7 layer listener failed, err %s", err.Error())
	}
	if output.Code != 0 {
		return fmt.Errorf("register instance with 7 layer listener failed, code %d, message %s",
			output.Code, output.Message)
	}
	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return fmt.Errorf("wait for task %d result failed, err %s", output.RequestID, err.Error())
	}
	blog.Infof("register successfully")
	return nil
}

//registerInstanceWith4thLayerListener
func (clb *ClbAPI) registerInsWith4thLayerListener(
	loadBalanceID, listenerID string, bdTargets qcloud.BackendTargetList) error {
	desc := new(qcloud.RegisterInstancesWithForwardLBFourthListenerInput)
	desc.Action = "RegisterInstancesWithForwardLBFourthListener"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.Backends = bdTargets
	desc.ListenerID = listenerID
	desc.LoadBalanceID = loadBalanceID

	blog.Infof("start register instance with 4 layer listener %v for listener %s", bdTargets, listenerID)

	output, err := clb.api.RegInstancesWith4LayerListener(desc)
	if err != nil {
		return fmt.Errorf("register instance with 4 layer listener failed, err %s", err.Error())
	}
	if output.Code != 0 {
		return fmt.Errorf("register instance with 4 layer listener failed, code %d, message %s",
			output.Code, output.Message)
	}
	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return fmt.Errorf("wait for task %d result failed, err %s", output.RequestID, err.Error())
	}
	blog.Infof("register successfully")
	return nil
}

//deRegisterInstances7thListener
func (clb *ClbAPI) deRegisterInstances7thListener(
	loadBalanceID, listenerID, ruleID string, bdTargets qcloud.BackendTargetList) error {
	desc := new(qcloud.DeregisterInstancesFromForwardLBSeventhListenerInput)
	desc.Action = "DeregisterInstancesFromForwardLB"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.Backends = bdTargets
	desc.ListenerID = listenerID
	desc.LoadBalanceID = loadBalanceID
	desc.LocationID = ruleID
	blog.Infof("start deRegister instances %v of lb %s, listener %s, rule %s",
		bdTargets, loadBalanceID, listenerID, ruleID)

	output, err := clb.api.DeRegInstancesWith7LayerListener(desc)
	if err != nil {
		return fmt.Errorf("deRegister instance with 7 layer listener failed, err %s", err.Error())
	}
	if output.Code != 0 {
		return fmt.Errorf("deRegister instance with 7 layer listener failed, code %d, message %s",
			output.Code, output.Message)
	}
	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return fmt.Errorf("wait for task %d result failed, err %s", output.RequestID, err.Error())
	}
	blog.Infof("deRegister instances successfully")
	return nil
}

//deRegisterInstances4thListener
func (clb *ClbAPI) deRegisterInstances4thListener(
	loadBalanceID, listenerID string, bdTargets qcloud.BackendTargetList) error {
	desc := new(qcloud.DeregisterInstancesFromForwardLBFourthListenerInput)
	desc.Action = "DeregisterInstancesFromForwardLBFourthListener"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.Backends = bdTargets
	desc.ListenerID = listenerID
	desc.LoadBalanceID = loadBalanceID

	blog.Infof("start deRegister instances %v of lb %s, listener %s", bdTargets, loadBalanceID, listenerID)

	output, err := clb.api.DeRegInstancesWith4LayerListener(desc)
	if err != nil {
		return fmt.Errorf("deRegister instance with 4 layer listener failed, err %s", err.Error())
	}
	if output.Code != 0 {
		return fmt.Errorf("deRegister instance with 4 layer listener failed, code %d, message %s",
			output.Code, output.Message)
	}
	err = clb.waitForTaskResult(output.RequestID)
	if err != nil {
		return fmt.Errorf("wait for task %d result failed, err %s", output.RequestID, err.Error())
	}
	blog.Infof("deRegister instances successfully")
	return nil
}

func (clb *ClbAPI) getCVMInstanceIDs(backends loadbalance.BackendList) ([]string, error) {
	var lanIPs []string
	for _, back := range backends {
		lanIPs = append(lanIPs, back.IP)
	}
	output, err := clb.describeCVMInstanceV3(lanIPs)
	if err != nil {
		return nil, fmt.Errorf("describe cvm instance v3 failed, err %s", err.Error())
	}
	var ids []string
	for _, instance := range output.CVMInfos.CVMInfo {
		ids = append(ids, instance.InstanceID)
	}
	if len(backends) != len(ids) {
		blog.Errorf("length of instance ids %v is not equal to length of backends %v", ids, backends)
		return nil, fmt.Errorf("length of instance ids %v is not equal to length of backends %v", ids, backends)
	}
	return ids, nil
}

//describeCVMInstanceV3 v3 api
//https://cloud.tencent.com/document/api/213/9388
func (clb *ClbAPI) describeCVMInstanceV3(lanIPs []string) (*qcloud.DescribeCVMInstanceV3Output, error) {
	desc := new(qcloud.DescribeCVMInstanceV3Input)
	desc.Action = "DescribeInstances"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.FilterIPName = "private-ip-address"

	var ipList qcloud.FilterIPValueFieldList
	for _, ip := range lanIPs {
		ipList = append(ipList, qcloud.FilterIPValueField{
			IP: ip,
		})
	}
	desc.FilterIPValues = ipList
	desc.Version = "2017-03-12"

	blog.Infof("describe cvm instance v3 by ips %v", lanIPs)

	output, err := clb.api.DescribeCVMInstanceV3(desc)
	if err != nil {
		return nil, fmt.Errorf("describe cvm instance v3 failed, err %s", err.Error())
	}

	return output, nil
}

/*

//getSecurityGroupPortRange
func (clb *ClbAPI) getSecurityGroupPortRange(index int, groupId string) (string, error) {
	desc := new(qcloud.DescribeSecurityGroupPolicysInput)
	desc.Action = "DescribeSecurityGroupPolicys"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.SgID = groupId
	c := qcloud.Client{
		URL:       QCloudDfwURL,
		SecretKey: clb.SecretKey,
	}
	dataBytes, err := c.GetRequest(desc)
	if err != nil {
		blog.Errorf("getSecurityGroupPortRange DescribeSecurityGroupPolicys failed, %s", err)
		return "", fmt.Errorf("getSecurityGroupPortRange DescribeSecurityGroupPolicys failed, %s", err)
	}
	blog.Infof("getSecurityGroupPortRange DescribeSecurityGroupPolicys result:%s", dataBytes)
	response := &qcloud.DescribeSecurityGroupPolicysOutput{}
	if err := json.Unmarshal(dataBytes, response); err != nil {
		blog.Errorf("getSecurityGroupPortRange DescribeSecurityGroupPolicys parse TaskResponse failed, %s", err)
		return "", fmt.Errorf("getSecurityGroupPortRange DescribeSecurityGroupPolicys parse TaskResponse failed, %s", err)
	}
	if response.Code != 0 {
		blog.Errorf("getSecurityGroupPortRange DescribeSecurityGroupPolicys result invalid:%s", response.Message)
		return "", fmt.Errorf("getSecurityGroupPortRange DescribeSecurityGroupPolicys result invalid")
	}
	blog.Infof("getSecurityGroupPortRange DescribeSecurityGroupPolicys result parse success:%v", response)
	for i, sgInfo := range response.Data.IngressInfos {
		if i == index {
			blog.Infof("getSecurityGroupPortRange found sgId %s portRange %s older", groupId, sgInfo.PortRange)
			return sgInfo.PortRange, nil
		}
	}
	blog.Infof("getSecurityGroupPortRange can not found sgId %s index %d", groupId, index)
	return "", fmt.Errorf("getSecurityGroupPortRange invalid index for SecurityGroup %s", groupId)
}

//modifySingleSecurityGroupPolicy
func (clb *ClbAPI) modifySingleSecurityGroupPolicy(sgID, protocol, portRange string) error {
	olderPortRange, err := clb.getSecurityGroupPortRange(qcloud.ClbSecurityGroupPolicyIndex, sgID)
	if err != nil {
		blog.Errorf("getSecurityGroupPortRange failed:%s", err.Error())
		return err
	}
	newPortRange := olderPortRange + "," + portRange
	desc := new(qcloud.ModifySingleSecurityGroupPolicyInput)
	desc.Action = "ModifySingleSecurityGroupPolicy"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.Direction = "ingress"
	desc.Index = 0
	desc.PolicyAction = "accept"
	//http/https/tcp -> tcp
	//udp -> udp
	desc.PolicyIpProtocol = protocol
	desc.PolicyCidrIp = clb.CidrIP
	desc.PolicyPortRange = newPortRange
	desc.PolicyDesc = "clb policy"
	desc.SgId = sgID
	c := qcloud.Client{
		URL:       QCloudDfwURL,
		SecretKey: clb.SecretKey,
	}
	dataBytes, err := c.GetRequest(desc)
	if err != nil {
		blog.Errorf("modifySingleSecurityGroupPolicy failed, %s", err)
		return err
	}
	blog.Infof("modifySingleSecurityGroupPolicy result:%s", dataBytes)
	response := &qcloud.ModifySingleSecurityGroupPolicyOutput{}
	if err := json.Unmarshal(dataBytes, response); err != nil {
		blog.Errorf("modifySingleSecurityGroupPolicy parse TaskResponse failed, %s", err)
		return err
	}
	if response.Code != 0 {
		blog.Errorf("ModifySingleSecurityGroupPolicyOutput result invalid:%s", response.Message)
		return fmt.Errorf("ModifySingleSecurityGroupPolicyOutput result invalid")
	}
	blog.Infof("modify portRange %s for sucurityGroup %s success", portRange, sgID)
	return nil
}

//modifySecurityGroupsOfInstance
func (clb *ClbAPI) modifySecurityGroupsOfInstance(instanceID string, securityGroups []string) error {
	desc := new(qcloud.ModifySecurityGroupsOfInstanceInput)
	desc.Action = "ModifyInstancesAttribute"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.InstanceID = instanceID
	desc.SecurityGroups = securityGroups
	desc.Version = "2017-03-12"
	c := qcloud.Client{
		URL:       QCloudCVMURL,
		SecretKey: clb.SecretKey,
	}
	dataBytes, err := c.GetRequest(desc)
	if err != nil {
		blog.Errorf("ModifyInstancesAttribute failed, %s", err)
		return err
	}
	blog.Infof("ModifyInstancesAttribute result:%s", dataBytes)
	response := &qcloud.DescribeCVMInstanceV3Output{}
	if err := json.Unmarshal(dataBytes, response); err != nil {
		blog.Errorf("DescribeCVMInstanceV3 parse TaskResponse failed, %s", err)
		return err
	}
	blog.Infof("DescribeCVMInstanceV3 result parse success:%v", response)

	return nil
}

func (clb *ClbAPI) securityGroupRuleCheck(data []byte, port int64, protocol string) error {
	var sgInfoStruct qcloud.IngressEgressData
	if err := json.Unmarshal([]byte(data), &sgInfoStruct); err != nil {
		blog.Errorf("json.Unmarshal %s failed:%s", data, err.Error())
		return err
	}
	for _, IpPermission := range sgInfoStruct.IngressInfos {
		if IpPermission.CidrIP != clb.CidrIP {
			continue
		}
		portRange := strings.Split(IpPermission.PortRange, ",")
		for _, comPort := range portRange {
			if comPort == "all" {
				//get it
				return nil
			}
			portInt, err := strconv.Atoi(comPort)
			if err != nil {
				blog.Errorf("invalid portRange %d failed, %s", port, err.Error())
				return err
			}

			if portInt == int(port) {
				blog.Infof("port %d already in securityGoup %s, cidr %s", port, sgInfoStruct.SgID, IpPermission.CidrIP)
				return nil
			}
		}
	}
	//go there, need to add port to security group
	if err := clb.modifySingleSecurityGroupPolicy(sgInfoStruct.SgID, protocol, strconv.Itoa(int(port))); err != nil {
		blog.Errorf("ModifySingleSecurityGroupPolicy for port %d for securitygroup %s failed:%s",
			port, sgInfoStruct.SgID, err.Error())
	}

	return nil
}

//ModifyHostSecurityGroups 追加主机安全组
//没有办法批量增加一堆IP安全组，因为这一堆IP原来的安全组可能不一样，没有办法统一覆盖
func (clb *ClbAPI) ModifyHostSecurityGroups(securityGroupID, hostIP string) error {
	//查询hostIP原来的安全组
	cvmInfo, err := clb.describeCVMInstanceV3(hostIP)
	if err != nil {
		blog.Errorf("DescribeCVMInstanceV3 failed:%s", err.Error())
		return err
	}
	//调用接口全量覆盖主机关联的安全组
	for _, cvm := range cvmInfo.CVMInfos.CVMInfo {
		sgIDs := cvm.SecurityGroupIds
		sgIDs = append(sgIDs, securityGroupID)
		if err = clb.modifySecurityGroupsOfInstance(cvm.InstanceID, sgIDs); err != nil {
			blog.Errorf("ModifySecurityGroupsOfInstance failed:%s", err.Error())
			return err
		}
		blog.Infof("modify instance %s security group %v success", cvm.InstanceID, sgIDs)
	}

	return nil
}

func (clb *ClbAPI) describeSecurityGroup(groupId *string) ([]byte, error) {
	desc := new(qcloud.DescribeSecurityGroupPolicysInput)
	desc.Action = "DescribeSecurityGroupPolicys"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.SgID = *groupId
	c := qcloud.Client{
		URL:       QCloudDfwURL,
		SecretKey: clb.SecretKey,
	}
	dataBytes, err := c.GetRequest(desc)
	if err != nil {
		blog.Errorf("DescribeSecurityGroupPolicys failed, %s", err.Error())
		return nil, err
	}
	blog.Infof("DescribeSecurityGroupPolicys result:%s", dataBytes)
	response := &qcloud.DescribeSecurityGroupPolicysOutput{}
	if err := json.Unmarshal(dataBytes, response); err != nil {
		blog.Errorf("DescribeSecurityGroupPolicys parse TaskResponse failed, %s", err)
		return nil, err
	}
	if response.Code != 0 {
		blog.Errorf("DescribeSecurityGroupPolicys result invalid:%s", response.Message)
		return nil, fmt.Errorf("DescribeSecurityGroupPolicys result invalid")
	}
	blog.Infof("DescribeSecurityGroupPolicys result parse success:%v", response)
	rst := &qcloud.IngressEgressData{}
	rst.SgID = *groupId
	rst.IngressInfos = response.Data.IngressInfos
	rst.EgressInfos = response.Data.EgressInfos

	dataRst, err := json.Marshal(rst)
	if err != nil {
		blog.Errorf("json.Marshal %v failed:%s", rst, err.Error())
		return dataRst, fmt.Errorf("json.Marshal %v failed:%s", rst, err.Error())
	}
	return dataRst, nil
}

func (clb *ClbAPI) createSecurityGroupPolicy(index, sgID, protocol, portRange, cidrIP string) error {
	desc := new(qcloud.CreateSecurityGroupPolicyInput)
	desc.Action = "CreateSecurityGroupPolicy"
	desc.Nonce = uint(rand.Uint32())
	desc.Region = clb.Region
	desc.SecretID = clb.SecretID
	desc.Timestamp = uint(time.Now().Unix())
	desc.Index = index
	desc.Direction = "ingress"
	desc.PolicyAction = "accept"
	if strings.ToLower(protocol) == "udp" {
		desc.PolicyIPProtocol = "udp"
	} else {
		//http/https/tcp
		desc.PolicyIPProtocol = "tcp"
	}

	//不再按照范围开通策略，只是留了一个口子
	desc.PolicyPortRange = portRange

	if strings.ToLower(desc.PolicyPortRange) == "all" {
		desc.PolicyIPProtocol = "all"
	}

	//来源IP范围
	desc.PolicyCidrIP = cidrIP
	desc.PolicyDesc = "clb policy"
	desc.SgID = sgID
	c := qcloud.Client{
		URL:       QCloudDfwURL,
		SecretKey: clb.SecretKey,
	}
	dataBytes, err := c.GetRequest(desc)
	if err != nil {
		blog.Errorf("CreateSecurityGroupPolicy failed, %s", err)
		return err
	}
	blog.Infof("CreateSecurityGroupPolicy result:%s", dataBytes)
	response := &qcloud.CreateSecurityGroupPolicyOutput{}
	if err := json.Unmarshal(dataBytes, response); err != nil {
		blog.Errorf("CreateSecurityGroupPolicy parse TaskResponse failed, %s", err)
		return err
	}
	if response.Code != 0 {
		blog.Errorf("CreateSecurityGroupPolicy result invalid:%s", response.Message)
		return fmt.Errorf("CreateSecurityGroupPolicy result invalid")
	}
	blog.Infof("CreateSecurityGroupPolicy result parse success:%v", response)

	return nil
}

*/
