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

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	qcloud "github.com/Tencent/bk-bcs/bcs-common/pkg/qcloud/clbv2"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
)

// do create listener
// call next layer function according to listener protocol
func (c *Clb) createListner(region string, listener *networkextensionv1.Listener) (string, error) {
	protocol := listener.Spec.Protocol
	if common.InLayer7Protocol(protocol) {
		return c.create7LayerListener(region, listener)
	} else if common.InLayer4Protocol(protocol) {
		return c.create4LayerListener(region, listener)
	} else {
		blog.Errorf("invalid protocol %s", listener.Spec.Protocol)
		return "", fmt.Errorf("invalid protocol %s", listener.Spec.Protocol)
	}
}

// do create 4 layer listener
// (4 layer listener) --------> backend1
//
//	|--> backend2
//	|--> backend3
//	|--> ...
func (c *Clb) create4LayerListener(region string, listener *networkextensionv1.Listener) (string, error) {
	// construct request for creating listener
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
	listenerIDs, err := c.sdkWrapper.CreateListener(region, req)
	if err != nil {
		cloud.StatRequest("CreateListener", cloud.MetricAPIFailed, ctime, time.Now())
		return "", err
	}
	cloud.StatRequest("CreateListener", cloud.MetricAPISuccess, ctime, time.Now())
	listenerID := listenerIDs[0]

	// if target group is not empty and backends in target group is not empty
	// start to register backend to the listener
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

// do create 7 layer listener
// (7 layer listener) --------> rule1
//
//	|--> rule2
//	|--> rule3
//	|--> ...
//
// domain and url is different in different rules
func (c *Clb) create7LayerListener(region string, listener *networkextensionv1.Listener) (string, error) {
	// construct request for creating listener
	req := tclb.NewCreateListenerRequest()
	req.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadbalancerID)
	req.Ports = []*int64{
		tcommon.Int64Ptr(int64(listener.Spec.Port)),
	}
	req.ListenerNames = tcommon.StringPtrs([]string{listener.GetName()})
	req.Protocol = tcommon.StringPtr(listener.Spec.Protocol)
	req.Certificate = transIngressCertificate(listener.Spec.Certificate)

	ctime := time.Now()
	listenerIDs, err := c.sdkWrapper.CreateListener(region, req)
	if err != nil {
		cloud.StatRequest("CreateListener", cloud.MetricAPIFailed, ctime, time.Now())
		return "", err
	}
	cloud.StatRequest("CreateListener", cloud.MetricAPISuccess, ctime, time.Now())
	listenerID := listenerIDs[0]

	// if rules is not empty, create listener rule
	for _, rule := range listener.Spec.Rules {
		err := c.addListenerRule(region, listener.Spec.LoadbalancerID, listenerID, listener.Spec.ListenerAttribute,
			rule)
		if err != nil {
			return "", err
		}
	}
	return listenerID, nil
}

// get listener info by listener port
// 1. call api to get listener description
// 2. call api to get listener backends
func (c *Clb) getListenerInfoByPort(region, lbID, protocol string, port int) (*networkextensionv1.Listener, error) {
	// construct request
	req := tclb.NewDescribeListenersRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.Port = tcommon.Int64Ptr(int64(port))
	req.Protocol = tcommon.StringPtr(protocol)

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
	// get segment listener end port
	if respListener.EndPort != nil && *respListener.EndPort > 0 {
		li.Spec.EndPort = int(*respListener.EndPort)
	}
	li.Spec.Protocol = strings.ToLower(*respListener.Protocol)
	li.Spec.Certificate = convertCertificate(respListener.Certificate)
	li.Spec.ListenerAttribute = convertListenerAttribute(respListener)
	ruleIDAttrMap := make(map[string]*networkextensionv1.IngressListenerAttribute)
	ruleIDCertMap := make(map[string]*networkextensionv1.IngressListenerCertificate)
	if len(respListener.Rules) != 0 {
		for _, respRule := range respListener.Rules {
			if respRule.LocationId != nil {
				ruleIDAttrMap[*respRule.LocationId] = convertRuleAttribute(respRule)
				ruleIDCertMap[*respRule.LocationId] = convertCertificate(respRule.Certificate)
			}
		}
	}

	// get backends info of listener
	rules, tg, err := c.getListenerBackendsByPort(region, lbID, protocol, port)
	if err != nil {
		return nil, err
	}
	if len(rules) != 0 {
		for index := range rules {
			if ruleAttr, ok := ruleIDAttrMap[rules[index].RuleID]; ok {
				rules[index].ListenerAttribute = ruleAttr
			}
			if ruleCert, ok := ruleIDCertMap[rules[index].RuleID]; ok {
				rules[index].Certificate = ruleCert
			}
		}
	}
	li.Spec.Rules = rules
	li.Spec.TargetGroup = tg
	li.Status.ListenerID = *respListener.ListenerId
	return li, nil
}

// get listener backends by listener pot
func (c *Clb) getListenerBackendsByPort(region, lbID, protocol string, port int) (
	[]networkextensionv1.ListenerRule, *networkextensionv1.ListenerTargetGroup, error) {

	req := tclb.NewDescribeTargetsRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.Port = tcommon.Int64Ptr(int64(port))
	req.Protocol = tcommon.StringPtr(protocol)

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

	// for listeners with different protocol, backends info is slightly different
	pro := *respListener.Protocol
	if common.InLayer7Protocol(pro) {
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
	} else if common.InLayer4Protocol(pro) {
		tg := convertClbBackends(respListener.Targets)
		return nil, tg, nil
	} else {
		blog.Errorf("invalid protocol %s listener", *respListener.Protocol)
		return nil, nil, fmt.Errorf("invalid protocol %s listener", *respListener.Protocol)
	}
}

// delete listener by listener port
func (c *Clb) deleteListener(region, lbID, protocol string, port int) error {
	// first determine if the listener exists
	// there is no need to do delete action when listener doesn't exists
	req := tclb.NewDescribeListenersRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.Port = tcommon.Int64Ptr(int64(port))
	req.Protocol = tcommon.StringPtr(protocol)

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

	//  do delete action
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

// update listener
// call next layer function according to different listener protocol
func (c *Clb) updateListener(region string, ingressListener, cloudListener *networkextensionv1.Listener) error {
	protocol := ingressListener.Spec.Protocol
	if common.InLayer7Protocol(protocol) {
		if err := c.updateHTTPListener(region, ingressListener, cloudListener); err != nil {
			return err
		}
	} else if common.InLayer4Protocol(protocol) {
		if err := c.update4LayerListener(region, ingressListener, cloudListener); err != nil {
			return err
		}
	} else {
		blog.Errorf("invalid listener protocol %s", ingressListener.Spec.Protocol)
		return fmt.Errorf("invalid listener protocol %s", ingressListener.Spec.Protocol)
	}
	return nil
}

// update http and https listener
func (c *Clb) updateHTTPListener(region string, ingressListener, cloudListener *networkextensionv1.Listener) error {
	// if listener certificate is defined and is different from remote cloud listener attribute, then do update
	if ingressListener.Spec.Certificate != nil &&
		!reflect.DeepEqual(ingressListener.Spec.Certificate, cloudListener.Spec.Certificate) {
		err := c.updateListenerAttrAndCerts(region, cloudListener.Status.ListenerID, ingressListener)
		if err != nil {
			blog.Errorf("updateListenerAttrAndCerts in updateHTTPListener failed, err %s", err.Error())
			return fmt.Errorf("updateListenerAttrAndCerts in updateHTTPListener failed, err %s", err.Error())
		}
	}
	// get differet rules
	addRules, delRules, updateOldRules, updateRules := getDiffBetweenListenerRule(cloudListener, ingressListener)
	// do delete rules
	for _, rule := range delRules {
		err := c.deleteListenerRule(region, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID, rule)
		if err != nil {
			return err
		}
	}
	// do add rules
	for _, rule := range addRules {
		err := c.addListenerRule(region, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID,
			cloudListener.Spec.ListenerAttribute, rule)
		if err != nil {
			return err
		}
	}
	// do update rules
	for index, rule := range updateRules {
		err := c.updateListenerRule(region, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID,
			updateOldRules[index], rule)
		if err != nil {
			return err
		}
	}
	return nil
}

// update 4 layer listener
func (c *Clb) update4LayerListener(region string, ingressListener, cloudListener *networkextensionv1.Listener) error {
	// if listener attribute is defined and is different from remote cloud listener attribute, then do update
	if ingressListener.Spec.ListenerAttribute != nil &&
		needUpdateAttribute(cloudListener.Spec.ListenerAttribute, ingressListener.Spec.ListenerAttribute) {
		err := c.updateListenerAttrAndCerts(region, cloudListener.Status.ListenerID, ingressListener)
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

// update listener attributes and certificates
func (c *Clb) updateListenerAttrAndCerts(region, listenerID string, listener *networkextensionv1.Listener) error {
	if listener.Spec.ListenerAttribute == nil && listener.Spec.Certificate == nil {
		return nil
	}
	req := tclb.NewModifyListenerRequest()
	req.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadbalancerID)
	req.ListenerId = tcommon.StringPtr(listenerID)
	if listener.Spec.ListenerAttribute != nil {
		attr := listener.Spec.ListenerAttribute
		req.SessionExpireTime = tcommon.Int64Ptr(int64(attr.SessionTime))
		if len(attr.LbPolicy) != 0 {
			req.Scheduler = tcommon.StringPtr(attr.LbPolicy)
		}
		req.HealthCheck = transIngressHealtchCheck(attr.HealthCheck)

		// 注意：未开启SNI的监听器可以开启SNI；已开启SNI的监听器不能关闭SNI。
		req.SniSwitch = tcommon.Int64Ptr(int64(attr.SniSwitch))
		req.KeepaliveEnable = tcommon.Int64Ptr(int64(attr.KeepAliveEnable))
	} else {
		req.SniSwitch = tcommon.Int64Ptr(0)
		req.KeepaliveEnable = tcommon.Int64Ptr(0)
	}
	// keep alive enable参数仅支持HTTPS/HTTP监听器
	if listener.Spec.Protocol != networkextensionv1.ProtocolHTTPS && listener.Spec.Protocol != networkextensionv1.
		ProtocolHTTP {
		req.KeepaliveEnable = nil
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

// update rule attribute
// include loadbalance policy, health check
func (c *Clb) updateRuleAttr(region, lbID, listenerID, locationID string, rule networkextensionv1.ListenerRule) error {
	if rule.ListenerAttribute == nil {
		return nil
	}
	req := tclb.NewModifyRuleRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.ListenerId = tcommon.StringPtr(listenerID)
	req.LocationId = tcommon.StringPtr(rule.RuleID)
	attr := rule.ListenerAttribute
	req.SessionExpireTime = tcommon.Int64Ptr(int64(attr.SessionTime))
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

// update domain related attributes
// doc https://cloud.tencent.com/document/api/214/38092
func (c *Clb) updateDomainAttributes(region, lbID, listenerID string, rule networkextensionv1.ListenerRule) error {
	req := tclb.NewModifyDomainAttributesRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.ListenerId = tcommon.StringPtr(listenerID)
	req.Domain = tcommon.StringPtr(rule.Domain)

	req.Certificate = transIngressCertificate(rule.Certificate)

	ctime := time.Now()
	err := c.sdkWrapper.ModifyDomainAttributes(region, req)
	if err != nil {
		cloud.StatRequest("ModifyDomainAttributes", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("ModifyDomainAttributes", cloud.MetricAPISuccess, ctime, time.Now())
	return nil
}

// update listener rule
func (c *Clb) updateListenerRule(region, lbID, listenerID string,
	existedRule, newRule networkextensionv1.ListenerRule) error {
	if needUpdateAttribute(existedRule.ListenerAttribute, newRule.ListenerAttribute) {
		err := c.updateRuleAttr(region, lbID, listenerID, existedRule.RuleID, newRule)
		if err != nil {
			return err
		}
	}
	if newRule.Certificate != nil && !reflect.DeepEqual(newRule.Certificate, existedRule.Certificate) {
		err := c.updateDomainAttributes(region, lbID, listenerID, newRule)
		if err != nil {
			return err
		}
	}
	// get different targets
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

// add listener rule
func (c *Clb) addListenerRule(region, lbID, listenerID string, listenerAttribute *networkextensionv1.
IngressListenerAttribute, rule networkextensionv1.ListenerRule) error {
	// construct create rule request
	req := tclb.NewCreateRuleRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.ListenerId = tcommon.StringPtr(listenerID)
	ruleInput := &tclb.RuleInput{
		Domain: tcommon.StringPtr(rule.Domain),
		Url:    tcommon.StringPtr(rule.Path),
	}
	if rule.TargetGroup.TargetGroupProtocol == ClbProtocolGRPC {
		ruleInput.ForwardType = tcommon.StringPtr(ClbProtocolGRPC)
		ruleInput.Http2 = tcommon.BoolPtr(true)
		ruleInput.Quic = tcommon.BoolPtr(false)
	} else if rule.TargetGroup.TargetGroupProtocol == ClbProtocolQUIC {
		ruleInput.Quic = tcommon.BoolPtr(true)
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
	if listenerAttribute != nil && listenerAttribute.SniSwitch == 1 {
		ruleInput.Certificate = transIngressCertificate(rule.Certificate)
	}
	req.Rules = append(req.Rules, ruleInput)
	ctime := time.Now()
	_, err := c.sdkWrapper.CreateRule(region, req)
	if err != nil {
		cloud.StatRequest("CreateRule", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("CreateRule", cloud.MetricAPISuccess, ctime, time.Now())

	// if both target group and backends in target group is not empty, do target registration
	if rule.TargetGroup != nil && len(rule.TargetGroup.Backends) != 0 {
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

// do delete listener rule
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

// create listener with segment
// 端口段：以端口段为规则配置，一个vip的一段端口（首端口-尾端口）绑定一个RS的一段端口。
// 如将vip vport(8000, 8001, 8002……9000) 绑定到 rsip rsport(9000, 9001, 9002……10000)，vport和rsport一一对应
// vport 8000 转发到 rsport 9000
// vport 8001 转发到 rsport 9001
// create listener with segment can only use api interface for tencent cloud, sdk does not support
func (c *Clb) createSegmentListener(region string, listener *networkextensionv1.Listener) (string, error) {
	req := tclb.NewCreateListenerRequest()
	req.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadbalancerID)
	req.Ports = []*int64{
		tcommon.Int64Ptr(int64(listener.Spec.Port)),
	}
	req.EndPort = tcommon.Uint64Ptr(uint64(listener.Spec.EndPort))
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
	listenerIDs, err := c.sdkWrapper.CreateListener(region, req)
	if err != nil {
		cloud.StatRequest("CreateListener", cloud.MetricAPIFailed, ctime, time.Now())
		return "", err
	}
	cloud.StatRequest("CreateListener", cloud.MetricAPISuccess, ctime, time.Now())
	listenerID := listenerIDs[0]

	// if target group is not empty and backends in target group is not empty
	// start to register backend to the listener
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

// updateSegmentListener update listener with port segment
func (c *Clb) updateSegmentListener(region string, ingressListener, cloudListener *networkextensionv1.Listener) error {
	addTargets, delTargets, _ := getDiffBetweenTargetGroup(
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
	// register tagets
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
	return nil
}

// register target to listener with port segment
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

// get listener info with port segment
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

// get backends of listener with port segment
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

// delete listener with port segment
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

func (c *Clb) getBackendHealthStatus(region, ns string, lbIDs []string) (
	map[string][]*cloud.BackendHealthStatus, error) {
	req := new(tclb.DescribeTargetHealthRequest)
	req.LoadBalancerIds = tcommon.StringPtrs(lbIDs)
	resp, err := c.sdkWrapper.DescribeTargetHealth(region, req)
	if err != nil {
		return nil, err
	}
	retMap := make(map[string][]*cloud.BackendHealthStatus)
	for _, retLb := range resp.Response.LoadBalancers {
		for _, listener := range retLb.Listeners {
			for _, rule := range listener.Rules {
				for _, target := range rule.Targets {
					lbID := *retLb.LoadBalancerId
					tmpStatus := &cloud.BackendHealthStatus{
						ListenerID:   *listener.ListenerId,
						ListenerPort: int(*listener.Port),
						Namespace:    ns,
						IP:           *target.IP,
						Port:         int(*target.Port),
						Protocol:     *listener.Protocol,
						Status:       convertHealthStatus(*target.HealthStatusDetial),
					}
					if rule.Domain != nil {
						tmpStatus.Host = *rule.Domain
					}
					if rule.Url != nil {
						tmpStatus.Path = *rule.Url
					}
					retMap[lbID] = append(retMap[lbID], tmpStatus)
				}
			}
		}
	}
	return retMap, nil
}
