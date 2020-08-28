/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tencentcloud

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/qcloud"
)

func (c *Clb) createListner(region string, listener *networkextensionv1.Listener) (string, error) {
	switch listener.Spec.Protocol {
	case ClbProtocolHTTP, ClbProtocolHTTPS:
		return c.create7LayerListener(region, listener)
	case ClbProtocolTCP, ClbProtocolUDP:
		return c.create4LayerListener(region, listener)
	default:
		blog.Errorf("invalid protocol %s", listener.Spec.Protocol)
		return "", fmt.Errorf("invalid protocol %s", listener.Spec.Protocol)
	}
}

func (c *Clb) create4LayerListener(region string, listener *networkextensionv1.Listener) (string, error) {
	req := tclb.NewCreateListenerRequest()
	req.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadbalancerID)
	req.Ports = []*int64{
		tcommon.Int64Ptr(int64(listener.Spec.Port)),
	}
	req.ListenerNames = tcommon.StringPtrs([]string{listener.GetName()})
	req.Protocol = tcommon.StringPtr(listener.Spec.Protocol)
	if listener.Spec.ListenerAttribute != nil {
		if listener.Spec.ListenerAttribute.SessionTime != 0 {
			req.SessionExpireTime = tcommon.Int64Ptr(int64(listener.Spec.ListenerAttribute.SessionTime))
		}
		if len(listener.Spec.ListenerAttribute.LbPolicy) != 0 {
			req.Scheduler = tcommon.StringPtr(listener.Spec.ListenerAttribute.LbPolicy)
		}
		req.HealthCheck = transIngressHealtchCheck(listener.Spec.ListenerAttribute.HealthCheck)
	}

	ctime := time.Now()
	listenerID, err := c.sdkWrapper.CreateListener(region, req)
	if err != nil {
		cloud.StatRequest("CreateListener", cloud.MetricAPIFailed, ctime, time.Now())
		return "", err
	}
	cloud.StatRequest("CreateListener", cloud.MetricAPISuccess, ctime, time.Now())

	if listener.Spec.TargetGroup != nil && len(listener.Spec.TargetGroup.Backends) != 0 {
		var tgs []*tclb.Target
		for _, backend := range listener.Spec.TargetGroup.Backends {
			tgs = append(tgs, &tclb.Target{
				EniIp:  tcommon.StringPtr(backend.IP),
				Port:   tcommon.Int64Ptr(int64(backend.Port)),
				Weight: tcommon.Int64Ptr(int64(backend.Weight)),
			})
		}
		req := tclb.NewRegisterTargetsRequest()
		req.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadbalancerID)
		req.ListenerId = tcommon.StringPtr(listenerID)
		req.Targets = tgs
		ctime := time.Now()
		err := c.sdkWrapper.RegisterTargets(region, req)
		if err != nil {
			cloud.StatRequest("RegisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
			return "", err
		}
		cloud.StatRequest("RegisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
	}
	return listenerID, nil
}

func (c *Clb) create7LayerListener(region string, listener *networkextensionv1.Listener) (string, error) {
	req := tclb.NewCreateListenerRequest()
	req.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadbalancerID)
	req.Ports = []*int64{
		tcommon.Int64Ptr(int64(listener.Spec.Port)),
	}
	req.ListenerNames = tcommon.StringPtrs([]string{listener.GetName()})
	req.Protocol = tcommon.StringPtr(listener.Spec.Protocol)
	req.Certificate = transIngressCertificate(listener.Spec.Certificate)

	ctime := time.Now()
	listenerID, err := c.sdkWrapper.CreateListener(region, req)
	if err != nil {
		cloud.StatRequest("CreateListener", cloud.MetricAPIFailed, ctime, time.Now())
		return "", err
	}
	cloud.StatRequest("CreateListener", cloud.MetricAPISuccess, ctime, time.Now())

	for _, rule := range listener.Spec.Rules {
		err := c.addListenerRule(region, listener.Spec.LoadbalancerID, listenerID, rule)
		if err != nil {
			return "", err
		}
	}
	return listenerID, nil
}

func (c *Clb) getListenerInfoByPort(region, lbID string, port int) (*networkextensionv1.Listener, error) {
	req := tclb.NewDescribeListenersRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.Port = tcommon.Int64Ptr(int64(port))

	ctime := time.Now()
	resp, err := c.sdkWrapper.DescribeListeners(region, req)
	if err != nil {
		cloud.StatRequest("DescribeListeners", cloud.MetricAPIFailed, ctime, time.Now())
		return nil, err
	}
	cloud.StatRequest("DescribeListeners", cloud.MetricAPISuccess, ctime, time.Now())

	if len(resp.Response.Listeners) == 0 {
		blog.Errorf("listener with port %d of clb %s not found", port, lbID)
		return nil, cloud.ErrListenerNotFound
	}
	if len(resp.Response.Listeners) > 1 {
		blog.Errorf("DescribeListeners response invalid, more than one listener, resp: %s",
			resp.ToJsonString())
		return nil, fmt.Errorf("DescribeListeners response invalid, more than one listener, resp: %s",
			resp.ToJsonString())
	}

	respListener := resp.Response.Listeners[0]
	li := &networkextensionv1.Listener{}
	li.Spec.LoadbalancerID = lbID
	li.Spec.Port = port
	li.Spec.Protocol = strings.ToLower(*respListener.Protocol)
	li.Spec.Certificate = convertCertificate(respListener.Certificate)
	li.Spec.ListenerAttribute = convertListenerAttribute(respListener)

	rules, tg, err := c.getListenerBackendsByPort(region, lbID, port)
	if err != nil {
		return nil, err
	}
	li.Spec.Rules = rules
	li.Spec.TargetGroup = tg
	li.Status.ListenerID = *respListener.ListenerId
	return li, nil
}

func (c *Clb) getListenerBackendsByPort(region, lbID string, port int) (
	[]networkextensionv1.ListenerRule, *networkextensionv1.ListenerTargetGroup, error) {

	req := tclb.NewDescribeTargetsRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.Port = tcommon.Int64Ptr(int64(port))

	ctime := time.Now()
	resp, err := c.sdkWrapper.DescribeTargets(region, req)
	if err != nil {
		cloud.StatRequest("DescribeTargets", cloud.MetricAPIFailed, ctime, time.Now())
		return nil, nil, err
	}
	cloud.StatRequest("DescribeTargets", cloud.MetricAPISuccess, ctime, time.Now())

	if len(resp.Response.Listeners) == 0 {
		return nil, nil, nil
	}
	if len(resp.Response.Listeners) > 1 {
		blog.Errorf("DescribeTargets response invalid, more than one listener, resp: %s",
			resp.ToJsonString())
		return nil, nil, fmt.Errorf("DescribeTargets response invalid, more than one listener, resp: %s",
			resp.ToJsonString())
	}

	respListener := resp.Response.Listeners[0]

	switch *respListener.Protocol {
	case ClbProtocolHTTP, ClbProtocolHTTPS:
		var rules []networkextensionv1.ListenerRule
		for _, retRule := range respListener.Rules {
			rules = append(rules, networkextensionv1.ListenerRule{
				RuleID:      *retRule.LocationId,
				Domain:      *retRule.Domain,
				Path:        *retRule.Url,
				TargetGroup: convertClbBackends(retRule.Targets),
			})
		}
		return rules, nil, nil

	case ClbProtocolTCP, ClbProtocolUDP:
		tg := convertClbBackends(respListener.Targets)
		return nil, tg, nil
	default:
		blog.Errorf("invalid protocol %s listener", *respListener.Protocol)
		return nil, nil, fmt.Errorf("invalid protocol %s listener", *respListener.Protocol)
	}
}

func (c *Clb) deleteListener(region, lbID string, port int) error {
	req := tclb.NewDescribeListenersRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.Port = tcommon.Int64Ptr(int64(port))

	ctime := time.Now()
	resp, err := c.sdkWrapper.DescribeListeners(region, req)
	if err != nil {
		cloud.StatRequest("DescribeListeners", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("DescribeListeners", cloud.MetricAPISuccess, ctime, time.Now())

	if len(resp.Response.Listeners) == 0 {
		return nil
	}
	if len(resp.Response.Listeners) > 1 {
		blog.Errorf("response invalid, more than one listener, resp: %s", resp.ToJsonString())
		return fmt.Errorf("response invalid, more than one listener, resp: %s", resp.ToJsonString())
	}

	delreq := tclb.NewDeleteListenerRequest()
	delreq.LoadBalancerId = tcommon.StringPtr(lbID)
	delreq.ListenerId = resp.Response.Listeners[0].ListenerId

	ctime = time.Now()
	err = c.sdkWrapper.DeleteListener(region, delreq)
	if err != nil {
		cloud.StatRequest("DeleteListener", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("DeleteListener", cloud.MetricAPISuccess, ctime, time.Now())
	return nil
}

func (c *Clb) updateListener(region string, ingressListener, cloudListener *networkextensionv1.Listener) error {
	switch ingressListener.Spec.Protocol {
	case ClbProtocolHTTP, ClbProtocolHTTPS:
		if err := c.updateHTTPListener(region, ingressListener, cloudListener); err != nil {
			return err
		}
	case ClbProtocolTCP, ClbProtocolUDP:
		if err := c.update4LayerListener(region, ingressListener, cloudListener); err != nil {
			return err
		}
	default:
		blog.Errorf("invalid listener protocol %s", ingressListener.Spec.Protocol)
		return fmt.Errorf("invalid listener protocol %s", ingressListener.Spec.Protocol)
	}
	return nil
}

func (c *Clb) updateHTTPListener(region string, ingressListener, cloudListener *networkextensionv1.Listener) error {
	if ingressListener.Spec.ListenerAttribute != nil &&
		(!reflect.DeepEqual(ingressListener.Spec.ListenerAttribute, cloudListener.Spec.ListenerAttribute) ||
			!reflect.DeepEqual(ingressListener.Spec.Certificate, cloudListener.Spec.Certificate)) {
		err := c.updateListenerAttrAndCerts(region, ingressListener)
		if err != nil {
			blog.Errorf("updateListenerAttrAndCerts in updateHTTPListener failed, err %s", err.Error())
			return fmt.Errorf("updateListenerAttrAndCerts in updateHTTPListener failed, err %s", err.Error())
		}
	}
	addRules, delRules, updateOldRules, updateRules := getDiffBetweenListenerRule(cloudListener, ingressListener)
	for _, rule := range delRules {
		err := c.deleteListenerRule(region, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID, rule)
		if err != nil {
			return err
		}
	}
	for _, rule := range addRules {
		err := c.addListenerRule(region, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID, rule)
		if err != nil {
			return err
		}
	}
	for index, rule := range updateRules {
		err := c.updateListenerRule(region, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID,
			updateOldRules[index], rule)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Clb) update4LayerListener(region string, ingressListener, cloudListener *networkextensionv1.Listener) error {
	if ingressListener.Spec.ListenerAttribute != nil &&
		!reflect.DeepEqual(ingressListener.Spec.ListenerAttribute, cloudListener.Spec.ListenerAttribute) {
		err := c.updateListenerAttrAndCerts(region, ingressListener)
		if err != nil {
			blog.Errorf("updateListenerAttrAndCerts in update4LayerListener failed, err %s", err.Error())
			return fmt.Errorf("updateListenerAttrAndCerts in update4LayerListener failed, err %s", err.Error())
		}
	}
	addTargets, delTargets, updateWeightTargets := getDiffBetweenTargetGroup(
		cloudListener.Spec.TargetGroup, ingressListener.Spec.TargetGroup)
	// deregister targets
	if len(delTargets) != 0 {
		req := tclb.NewDeregisterTargetsRequest()
		req.LoadBalancerId = tcommon.StringPtr(ingressListener.Spec.LoadbalancerID)
		req.ListenerId = tcommon.StringPtr(cloudListener.Status.ListenerID)
		req.Targets = delTargets
		ctime := time.Now()
		if err := c.sdkWrapper.DeregisterTargets(region, req); err != nil {
			cloud.StatRequest("DeregisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
			return err
		}
		cloud.StatRequest("DeregisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
	}
	// register targets
	if len(addTargets) != 0 {
		req := tclb.NewRegisterTargetsRequest()
		req.LoadBalancerId = tcommon.StringPtr(ingressListener.Spec.LoadbalancerID)
		req.ListenerId = tcommon.StringPtr(cloudListener.Status.ListenerID)
		req.Targets = addTargets
		ctime := time.Now()
		if err := c.sdkWrapper.RegisterTargets(region, req); err != nil {
			cloud.StatRequest("RegisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
			return err
		}
		cloud.StatRequest("RegisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
	}
	// modify weight of targets
	if len(updateWeightTargets) != 0 {
		req := tclb.NewModifyTargetWeightRequest()
		req.LoadBalancerId = tcommon.StringPtr(ingressListener.Spec.LoadbalancerID)
		req.ListenerId = tcommon.StringPtr(cloudListener.Status.ListenerID)
		req.Targets = updateWeightTargets
		ctime := time.Now()
		if err := c.sdkWrapper.ModifyTargetWeight(region, req); err != nil {
			cloud.StatRequest("ModifyTargetWeight", cloud.MetricAPIFailed, ctime, time.Now())
			return err
		}
		cloud.StatRequest("ModifyTargetWeight", cloud.MetricAPISuccess, ctime, time.Now())
	}
	return nil
}

func (c *Clb) updateListenerAttrAndCerts(region string, listener *networkextensionv1.Listener) error {
	if listener.Spec.ListenerAttribute == nil && listener.Spec.Certificate == nil {
		return nil
	}
	req := tclb.NewModifyListenerRequest()
	if listener.Spec.ListenerAttribute != nil {
		attr := listener.Spec.ListenerAttribute
		if attr.SessionTime != 0 {
			req.SessionExpireTime = tcommon.Int64Ptr(int64(attr.SessionTime))
		}
		if len(attr.LbPolicy) != 0 {
			req.Scheduler = tcommon.StringPtr(attr.LbPolicy)
		}
		req.HealthCheck = transIngressHealtchCheck(attr.HealthCheck)
	}
	certs := listener.Spec.Certificate
	req.Certificate = transIngressCertificate(certs)
	ctime := time.Now()
	err := c.sdkWrapper.ModifyListener(region, req)
	if err != nil {
		cloud.StatRequest("ModifyListener", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("ModifyListener", cloud.MetricAPISuccess, ctime, time.Now())
	return nil
}

func (c *Clb) updateRuleAttr(region, lbID, listenerID, locationID string, rule networkextensionv1.ListenerRule) error {
	if rule.ListenerAttribute == nil {
		return nil
	}
	req := tclb.NewModifyRuleRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.ListenerId = tcommon.StringPtr(listenerID)
	req.LocationId = tcommon.StringPtr(rule.RuleID)
	attr := rule.ListenerAttribute
	if attr.SessionTime != 0 {
		req.SessionExpireTime = tcommon.Int64Ptr(int64(attr.SessionTime))
	}
	if len(attr.LbPolicy) != 0 {
		req.Scheduler = tcommon.StringPtr(attr.LbPolicy)
	}
	req.HealthCheck = transIngressHealtchCheck(attr.HealthCheck)

	ctime := time.Now()
	err := c.sdkWrapper.ModifyRule(region, req)
	if err != nil {
		cloud.StatRequest("ModifyListener", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("ModifyListener", cloud.MetricAPISuccess, ctime, time.Now())
	return nil
}

func (c *Clb) updateListenerRule(region, lbID, listenerID string,
	existedRule, newRule networkextensionv1.ListenerRule) error {

	if !reflect.DeepEqual(newRule.ListenerAttribute, existedRule.ListenerAttribute) {
		err := c.updateRuleAttr(region, lbID, listenerID, existedRule.RuleID, newRule)
		if err != nil {
			return err
		}
	}

	addTargets, delTargets, updateWeightTargets := getDiffBetweenTargetGroup(
		existedRule.TargetGroup, newRule.TargetGroup)
	// deregister targets
	if len(delTargets) != 0 {
		req := tclb.NewDeregisterTargetsRequest()
		req.LoadBalancerId = tcommon.StringPtr(lbID)
		req.ListenerId = tcommon.StringPtr(listenerID)
		req.Domain = tcommon.StringPtr(newRule.Domain)
		req.Url = tcommon.StringPtr(newRule.Path)
		req.Targets = delTargets
		ctime := time.Now()
		if err := c.sdkWrapper.DeregisterTargets(region, req); err != nil {
			cloud.StatRequest("DeregisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
			return err
		}
		cloud.StatRequest("DeregisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
	}
	// register targets
	if len(addTargets) != 0 {
		req := tclb.NewRegisterTargetsRequest()
		req.LoadBalancerId = tcommon.StringPtr(lbID)
		req.ListenerId = tcommon.StringPtr(listenerID)
		req.Domain = tcommon.StringPtr(newRule.Domain)
		req.Url = tcommon.StringPtr(newRule.Path)
		req.Targets = addTargets
		ctime := time.Now()
		if err := c.sdkWrapper.RegisterTargets(region, req); err != nil {
			cloud.StatRequest("RegisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
			return err
		}
		cloud.StatRequest("RegisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
	}
	// modify weight of targets
	if len(updateWeightTargets) != 0 {
		req := tclb.NewModifyTargetWeightRequest()
		req.LoadBalancerId = tcommon.StringPtr(lbID)
		req.ListenerId = tcommon.StringPtr(listenerID)
		req.Domain = tcommon.StringPtr(newRule.Domain)
		req.Url = tcommon.StringPtr(newRule.Path)
		req.Targets = updateWeightTargets
		ctime := time.Now()
		if err := c.sdkWrapper.ModifyTargetWeight(region, req); err != nil {
			cloud.StatRequest("ModifyTargetWeight", cloud.MetricAPIFailed, ctime, time.Now())
			return err
		}
		cloud.StatRequest("ModifyTargetWeight", cloud.MetricAPISuccess, ctime, time.Now())
	}
	return nil
}

func (c *Clb) addListenerRule(region, lbID, listenerID string, rule networkextensionv1.ListenerRule) error {
	req := tclb.NewCreateRuleRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.ListenerId = tcommon.StringPtr(listenerID)
	ruleInput := &tclb.RuleInput{
		Domain: tcommon.StringPtr(rule.Domain),
		Url:    tcommon.StringPtr(rule.Path),
	}
	if rule.ListenerAttribute != nil {
		if rule.ListenerAttribute.SessionTime != 0 {
			ruleInput.SessionExpireTime = tcommon.Int64Ptr(int64(rule.ListenerAttribute.SessionTime))
		}
		if len(rule.ListenerAttribute.LbPolicy) != 0 {
			ruleInput.Scheduler = tcommon.StringPtr(rule.ListenerAttribute.LbPolicy)
		}
		ruleInput.HealthCheck = transIngressHealtchCheck(rule.ListenerAttribute.HealthCheck)
	}
	req.Rules = append(req.Rules, ruleInput)
	ctime := time.Now()
	err := c.sdkWrapper.CreateRule(region, req)
	if err != nil {
		cloud.StatRequest("CreateRule", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("CreateRule", cloud.MetricAPISuccess, ctime, time.Now())

	if rule.TargetGroup != nil {
		req := tclb.NewRegisterTargetsRequest()
		req.LoadBalancerId = tcommon.StringPtr(lbID)
		req.ListenerId = tcommon.StringPtr(listenerID)
		req.Domain = tcommon.StringPtr(rule.Domain)
		req.Url = tcommon.StringPtr(rule.Path)
		req.Targets = getTargets(rule.TargetGroup)
		ctime := time.Now()
		if err := c.sdkWrapper.RegisterTargets(region, req); err != nil {
			cloud.StatRequest("RegisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
			return err
		}
		cloud.StatRequest("RegisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
	}
	return nil
}

func (c *Clb) deleteListenerRule(region, lbID, listenerID string, rule networkextensionv1.ListenerRule) error {
	req := tclb.NewDeleteRuleRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.ListenerId = tcommon.StringPtr(listenerID)
	req.LocationIds = tcommon.StringPtrs([]string{rule.RuleID})
	ctime := time.Now()
	err := c.sdkWrapper.DeleteRule(region, req)
	if err != nil {
		cloud.StatRequest("DeleteRule", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("DeleteRule", cloud.MetricAPISuccess, ctime, time.Now())
	return nil
}

func (c *Clb) createSegmentListener(region string, listener *networkextensionv1.Listener) (string, error) {
	req := new(qcloud.CreateForwardLBFourthLayerListenersInput)
	req.LoadBalanceID = listener.Spec.LoadbalancerID
	req.ListenersLoadBalancerPort = listener.Spec.Port
	req.EndPort = listener.Spec.EndPort
	req.ListenersListenerName = listener.GetName()
	//we will validate the field in upper function
	protocol, _ := ProtocolTypeBcs2QCloudMap[listener.Spec.Protocol]
	req.ListenersProtocol = protocol
	if listener.Spec.ListenerAttribute != nil {
		if listener.Spec.ListenerAttribute.SessionTime != 0 {
			req.ListenerExpireTime = listener.Spec.ListenerAttribute.SessionTime
		}
		if listener.Spec.ListenerAttribute.HealthCheck != nil {
			req.ListenerHealthSwitch = DefaultHealthCheckEnabled
			req.ListenerIntervalTime = DefaultHealthCheckIntervalTime
			req.ListenerHealthNum = DefaultHealthCheckHealthNum
			req.ListenerUnHealthNum = DefaultHealthCheckUnhealthNum
			req.ListenerTimeout = DefaultHealthCheckTimeout
			hc := listener.Spec.ListenerAttribute.HealthCheck
			var heatlthSwitch int
			if listener.Spec.ListenerAttribute.HealthCheck.Enabled {
				heatlthSwitch = 1
			} else {
				heatlthSwitch = 0
			}
			req.ListenerHealthSwitch = heatlthSwitch
			if hc.IntervalTime != 0 {
				req.ListenerIntervalTime = hc.IntervalTime
			}
			if hc.HealthNum != 0 {
				req.ListenerHealthNum = hc.HealthNum
			}
			if hc.UnHealthNum != 0 {
				req.ListenerUnHealthNum = hc.UnHealthNum
			}
			if hc.Timeout != 0 {
				req.ListenerTimeout = hc.Timeout
			}
		}
	}
	ctime := time.Now()
	listenerID, err := c.apiWrapper.Create4LayerListener(region, req)
	if err != nil {
		cloud.StatRequest("Create4LayerListener", cloud.MetricAPIFailed, ctime, time.Now())
		return "", err
	}
	cloud.StatRequest("Create4LayerListener", cloud.MetricAPISuccess, ctime, time.Now())

	if listener.Spec.TargetGroup != nil && len(listener.Spec.TargetGroup.Backends) != 0 {
		err = c.registerSegmentListenerTarget(region, listener.Spec.LoadbalancerID, listenerID, listener.Spec.TargetGroup)
		if err != nil {
			return "", err
		}
	}
	return listenerID, nil
}

func (c *Clb) updateSegmentListener(region string, ingressListener, cloudListener *networkextensionv1.Listener) error {
	addTargets, delTargets, _ := getDiffBetweenTargetGroup(
		cloudListener.Spec.TargetGroup, ingressListener.Spec.TargetGroup)
	// deregister targets
	if len(delTargets) != 0 {
		req := new(qcloud.DeregisterInstancesFromForwardLBFourthListenerInput)
		req.LoadBalanceID = ingressListener.Spec.LoadbalancerID
		req.ListenerID = cloudListener.Status.ListenerID
		for _, tg := range delTargets {
			req.Backends = append(req.Backends, qcloud.BackendTarget{
				BackendsIP:   *tg.EniIp,
				BackendsPort: int(*tg.Port),
			})
		}
		ctime := time.Now()
		if err := c.apiWrapper.DeRegInstancesWith4LayerListener(region, req); err != nil {
			cloud.StatRequest("DeRegInstancesWith4LayerListener", cloud.MetricAPIFailed, ctime, time.Now())
			return err
		}
		cloud.StatRequest("DeRegInstancesWith4LayerListener", cloud.MetricAPISuccess, ctime, time.Now())
	}
	// register tagets
	if len(addTargets) != 0 {
		req := new(qcloud.RegisterInstancesWithForwardLBFourthListenerInput)
		req.LoadBalanceID = ingressListener.Spec.LoadbalancerID
		req.ListenerID = cloudListener.Status.ListenerID
		for _, tg := range addTargets {
			req.Backends = append(req.Backends, qcloud.BackendTarget{
				BackendsIP:     *tg.EniIp,
				BackendsPort:   int(*tg.Port),
				BackendsWeight: int(*tg.Weight),
			})
		}
		ctime := time.Now()
		if err := c.apiWrapper.RegInstancesWith4LayerListener(region, req); err != nil {
			cloud.StatRequest("RegInstancesWith4LayerListener", cloud.MetricAPIFailed, ctime, time.Now())
			return err
		}
		cloud.StatRequest("RegInstancesWith4LayerListener", cloud.MetricAPISuccess, ctime, time.Now())
	}
	return nil
}

func (c *Clb) registerSegmentListenerTarget(region string,
	lbID, listenerID string, target *networkextensionv1.ListenerTargetGroup) error {
	req := new(qcloud.RegisterInstancesWithForwardLBFourthListenerInput)
	req.LoadBalanceID = lbID
	req.ListenerID = listenerID

	var backends []qcloud.BackendTarget
	for _, back := range target.Backends {
		backends = append(backends, qcloud.BackendTarget{
			BackendsIP:     back.IP,
			BackendsPort:   back.Port,
			BackendsWeight: 10,
		})
	}
	req.Backends = backends
	ctime := time.Now()
	err := c.apiWrapper.RegInstancesWith4LayerListener(region, req)
	if err != nil {
		cloud.StatRequest("RegInstancesWith4LayerListener", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("RegInstancesWith4LayerListener", cloud.MetricAPISuccess, ctime, time.Now())
	return nil
}

func (c *Clb) getSegmentListenerInfoByPort(region, lbID string, port int) (*networkextensionv1.Listener, error) {
	req := new(qcloud.DescribeForwardLBListenersInput)
	req.LoadBalanceID = lbID
	req.LoadBalancerPort = port
	ctime := time.Now()
	resp, err := c.apiWrapper.DescribeForwardLBListeners(region, req)
	if err != nil {
		cloud.StatRequest("DescribeForwardLBListeners", cloud.MetricAPIFailed, ctime, time.Now())
		return nil, err
	}
	cloud.StatRequest("DescribeForwardLBListeners", cloud.MetricAPISuccess, ctime, time.Now())

	if len(resp.Listeners) == 0 {
		blog.Errorf("segment listener with port %d of clb %s not found", port, lbID)
		return nil, cloud.ErrListenerNotFound
	}
	if len(resp.Listeners) > 1 {
		blog.Errorf("DescribeForwardLBListeners response invalid, more than one listener, resp: %+v", resp)
		return nil, fmt.Errorf("DescribeForwardLBListeners response invalid, more than one listener, resp: %+v", resp)
	}

	respListener := resp.Listeners[0]
	li := &networkextensionv1.Listener{}
	li.Spec.LoadbalancerID = lbID
	li.Spec.Port = port
	li.Spec.Protocol = respListener.ProtocolType

	tg, err := c.getSegmentListenerBackendsByPort(region, lbID, port)
	if err != nil {
		return nil, err
	}
	li.Spec.TargetGroup = tg
	li.Status.ListenerID = respListener.ListenerID
	return li, nil
}

func (c *Clb) getSegmentListenerBackendsByPort(region, lbID string, port int) (
	*networkextensionv1.ListenerTargetGroup, error) {

	req := new(qcloud.DescribeForwardLBBackendsInput)
	req.LoadBalanceID = lbID
	req.LoadBalancerPort = port
	ctime := time.Now()
	resp, err := c.apiWrapper.DescribeForwardLBBackends(region, req)
	if err != nil {
		cloud.StatRequest("DescribeForwardLBBackends", cloud.MetricAPIFailed, ctime, time.Now())
		return nil, err
	}
	cloud.StatRequest("DescribeForwardLBBackends", cloud.MetricAPISuccess, ctime, time.Now())
	if len(resp.Data) == 0 {
		blog.Errorf("segment listener with port %d of clb %s not found", port, lbID)
		return nil, cloud.ErrListenerNotFound
	}
	if len(resp.Data) > 1 {
		blog.Errorf("DescribeForwardLBBackends response invalid, more than one listener, resp: %+v", resp)
		return nil, fmt.Errorf("DescribeForwardLBBackends response invalid, more than one listener, resp: %+v", resp)
	}

	respListener := resp.Data[0]
	tg := new(networkextensionv1.ListenerTargetGroup)
	tg.TargetGroupProtocol = respListener.ProtocolType
	for _, backend := range respListener.Backends {
		tg.Backends = append(tg.Backends, networkextensionv1.ListenerBackend{
			IP:     backend.LanIP,
			Port:   backend.Port,
			Weight: backend.Weight,
		})
	}
	return tg, nil
}

func (c *Clb) deleteSegmentListener(region, lbID string, port int) error {
	req := new(qcloud.DescribeForwardLBListenersInput)
	req.LoadBalanceID = lbID
	req.LoadBalancerPort = port
	ctime := time.Now()
	resp, err := c.apiWrapper.DescribeForwardLBListeners(region, req)
	if err != nil {
		cloud.StatRequest("DescribeForwardLBListeners", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("DescribeForwardLBListeners", cloud.MetricAPISuccess, ctime, time.Now())
	if len(resp.Listeners) == 0 {
		return nil
	}
	if len(resp.Listeners) > 1 {
		blog.Errorf("DescribeForwardLBListeners response invalid, more than one listener, resp: %+v", resp)
		return fmt.Errorf("DescribeForwardLBListeners response invalid, more than one listener, resp: %+v", resp)
	}
	respListener := resp.Listeners[0]

	reqDel := new(qcloud.DeleteForwardLBListenerInput)
	reqDel.LoadBalanceID = lbID
	reqDel.ListenerID = respListener.ListenerID
	ctime = time.Now()
	err = c.apiWrapper.DeleteListener(region, reqDel)
	if err != nil {
		cloud.StatRequest("DeleteListener", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("DeleteListener", cloud.MetricAPISuccess, ctime, time.Now())
	return nil
}
