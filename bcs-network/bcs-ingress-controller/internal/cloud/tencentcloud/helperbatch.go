/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
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
)

// desribe listener info and listener targets with port list, used by batch operation
func (c *Clb) batchDescribeListeners(region, lbID string, ports []int) (
	map[int]*networkextensionv1.Listener, error) {
	if len(ports) == 0 {
		return nil, nil
	}
	portMap := make(map[int]struct{})
	for _, port := range ports {
		portMap[port] = struct{}{}
	}
	// 1. desribe listener attributes
	req := tclb.NewDescribeListenersRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	ctime := time.Now()
	resp, err := c.sdkWrapper.DescribeListeners(region, req)
	if err != nil {
		cloud.StatRequest("DescribeListeners", cloud.MetricAPIFailed, ctime, time.Now())
		return nil, err
	}
	cloud.StatRequest("DescribeListeners", cloud.MetricAPISuccess, ctime, time.Now())
	retListenerMap := make(map[int]*networkextensionv1.Listener)
	var listenerIDs []string
	if len(resp.Response.Listeners) == 0 {
		return nil, nil
	}
	ruleIDAttrMap := make(map[string]*networkextensionv1.IngressListenerAttribute)
	for _, cloudLi := range resp.Response.Listeners {
		// only care about listener with given ports
		if _, ok := portMap[int(*cloudLi.Port)]; !ok {
			continue
		}
		listenerIDs = append(listenerIDs, *cloudLi.ListenerId)
		li := &networkextensionv1.Listener{}
		li.Spec.LoadbalancerID = lbID
		li.Spec.Port = int(*cloudLi.Port)
		// get segment listener end port
		if cloudLi.EndPort != nil && *cloudLi.EndPort > 0 {
			li.Spec.EndPort = 0
		}
		li.Spec.Protocol = strings.ToLower(*cloudLi.Protocol)
		li.Spec.Certificate = convertCertificate(cloudLi.Certificate)
		li.Spec.ListenerAttribute = convertListenerAttribute(cloudLi)
		if len(cloudLi.Rules) != 0 {
			for _, respRule := range cloudLi.Rules {
				if respRule.LocationId != nil {
					ruleIDAttrMap[*respRule.LocationId] = convertRuleAttribute(respRule)
				}
			}
		}
		li.Status.ListenerID = *cloudLi.ListenerId
		retListenerMap[li.Spec.Port] = li
	}

	// 2. describe listener targets
	dtReq := tclb.NewDescribeTargetsRequest()
	dtReq.LoadBalancerId = tcommon.StringPtr(lbID)
	dtReq.ListenerIds = tcommon.StringPtrs(listenerIDs)
	ctime = time.Now()
	dtResp, err := c.sdkWrapper.DescribeTargets(region, dtReq)
	if err != nil {
		cloud.StatRequest("DescribeTargets", cloud.MetricAPIFailed, ctime, time.Now())
		return nil, err
	}
	cloud.StatRequest("DescribeTargets", cloud.MetricAPISuccess, ctime, time.Now())

	// 3. combine the listener properties and the listener back-end information into a complete listener definition
	for _, retLi := range dtResp.Response.Listeners {
		// only care about listener with given ports
		if _, ok := portMap[int(*retLi.Port)]; !ok {
			continue
		}
		var rules []networkextensionv1.ListenerRule
		var tg *networkextensionv1.ListenerTargetGroup
		switch *retLi.Protocol {
		case ClbProtocolHTTP, ClbProtocolHTTPS:
			for _, retRule := range retLi.Rules {
				rules = append(rules, networkextensionv1.ListenerRule{
					RuleID:      *retRule.LocationId,
					Domain:      *retRule.Domain,
					Path:        *retRule.Url,
					TargetGroup: convertClbBackends(retRule.Targets),
				})
			}
		case ClbProtocolTCP, ClbProtocolUDP:
			tg = convertClbBackends(retLi.Targets)
		default:
			blog.Errorf("invalid protocol %s listener", *retLi.Protocol)
			return nil, fmt.Errorf("invalid protocol %s listener", *retLi.Protocol)
		}
		if len(rules) != 0 {
			for index := range rules {
				if ruleAttr, ok := ruleIDAttrMap[rules[index].RuleID]; ok {
					rules[index].ListenerAttribute = ruleAttr
				}
			}
		}
		li, ok := retListenerMap[int(*retLi.Port)]
		if !ok {
			blog.Errorf("listener %d is not found in DescribeListeners", int(*retLi.Port))
			return nil, fmt.Errorf("listener %d is not found in DescribeListeners", int(*retLi.Port))
		}
		li.Spec.Rules = rules
		li.Spec.TargetGroup = tg
		li.Status.ListenerID = *retLi.ListenerId
	}
	return retListenerMap, nil
}

// batchCreate4LayerListener create multiple listener, collect all targets and register targets togeteher
// in upper layer, already split listener into different batch, the attribute of listeners should be the same
func (c *Clb) batchCreate4LayerListener(
	region string, listeners []*networkextensionv1.Listener) (
	map[string]string, error) {
	if len(listeners) == 0 {
		return nil, fmt.Errorf("listeners cannot be empty when batch create 4 layer listener")
	}
	// do create listeners with same listener attributes
	req := tclb.NewCreateListenerRequest()
	req.LoadBalancerId = tcommon.StringPtr(listeners[0].Spec.LoadbalancerID)
	var portList []*int64
	var listenerNames []*string
	for _, li := range listeners {
		listenerNames = append(listenerNames, tcommon.StringPtr(li.GetName()))
		portList = append(portList, tcommon.Int64Ptr(int64(li.Spec.Port)))
	}
	req.Ports = portList
	req.ListenerNames = listenerNames
	listener := listeners[0]
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
		return nil, err
	}
	cloud.StatRequest("CreateListener", cloud.MetricAPISuccess, ctime, time.Now())
	if len(listenerIDs) != len(listeners) {
		blog.Errorf("number of created listeners %d is not equal input number of input listeners %d",
			len(listenerIDs), len(listenerIDs))
		return nil, fmt.Errorf("number of created listeners %d is not equal input number of input listeners %d",
			len(listenerIDs), len(listenerIDs))
	}

	// collect all new targets for all listeners
	tgReq := tclb.NewBatchRegisterTargetsRequest()
	tgReq.LoadBalancerId = tcommon.StringPtr(listeners[0].Spec.LoadbalancerID)
	for liIndex, li := range listeners {
		if li.Spec.TargetGroup != nil && len(li.Spec.TargetGroup.Backends) != 0 {
			for _, backend := range li.Spec.TargetGroup.Backends {
				tgReq.Targets = append(tgReq.Targets, &tclb.BatchTarget{
					ListenerId: tcommon.StringPtr(listenerIDs[liIndex]),
					EniIp:      tcommon.StringPtr(backend.IP),
					Port:       tcommon.Int64Ptr(int64(backend.Port)),
					Weight:     tcommon.Int64Ptr(int64(backend.Weight)),
				})
			}
		}
	}
	// register all targets
	failedListenerIDMap := make(map[string]struct{})
	if len(tgReq.Targets) != 0 {
		ctime := time.Now()
		failedListenerIDs := c.sdkWrapper.BatchRegisterTargets(region, tgReq)
		if len(failedListenerIDs) != 0 {
			cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
		} else {
			cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
		}
		for _, id := range failedListenerIDs {
			failedListenerIDMap[id] = struct{}{}
		}
	}
	// collect listeners which is created successfully
	retMap := make(map[string]string)
	for index, id := range listenerIDs {
		if _, ok := failedListenerIDMap[id]; !ok {
			retMap[listeners[index].GetName()] = id
		}
	}
	return retMap, nil
}

// create multiple 4 layer listener segment
// on tencent cloud, listener segment only support tcp and udp
func (c *Clb) batchCreateSegment4LayerListener(
	region string, listeners []*networkextensionv1.Listener) (map[string]string, error) {
	if len(listeners) == 0 {
		return nil, fmt.Errorf("listeners cannot be empty when batch create 4 layer listener segment")
	}

	// on tencent cloud, api cannot create multiple listener segment in one time
	failedListenerNameMap := make(map[string]struct{})
	successListenerNameMap := make(map[string]string)
	for _, li := range listeners {
		listenerID, err := c.create4LayerListenerWithoutTargetGroup(region, li)
		if err != nil {
			blog.Warnf("create 4 layer listener %s/%s failed, err %s", li.GetName(), li.GetNamespace(), err.Error())
			failedListenerNameMap[li.GetName()] = struct{}{}
		}
		successListenerNameMap[li.GetName()] = listenerID
	}

	// collect all targets and register them in one time
	tgReq := tclb.NewBatchRegisterTargetsRequest()
	tgReq.LoadBalancerId = tcommon.StringPtr(listeners[0].Spec.LoadbalancerID)
	for _, li := range listeners {
		if _, ok := failedListenerNameMap[li.GetName()]; ok {
			continue
		}
		if li.Spec.TargetGroup != nil && len(li.Spec.TargetGroup.Backends) != 0 {
			for _, backend := range li.Spec.TargetGroup.Backends {
				tgReq.Targets = append(tgReq.Targets, &tclb.BatchTarget{
					ListenerId: tcommon.StringPtr(successListenerNameMap[li.GetName()]),
					EniIp:      tcommon.StringPtr(backend.IP),
					Port:       tcommon.Int64Ptr(int64(backend.Port)),
					Weight:     tcommon.Int64Ptr(int64(backend.Weight)),
				})
			}
		}
	}
	failedListenerIDMap := make(map[string]struct{})
	if len(tgReq.Targets) != 0 {
		ctime := time.Now()
		failedListenerIDs := c.sdkWrapper.BatchRegisterTargets(region, tgReq)
		if len(failedListenerIDs) != 0 {
			cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
		} else {
			cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
		}
		for _, id := range failedListenerIDs {
			failedListenerIDMap[id] = struct{}{}
		}
	}
	// collect all listener which is created successfully
	retMap := make(map[string]string)
	for liName, liID := range successListenerNameMap {
		if _, ok := failedListenerIDMap[liID]; ok {
			continue
		}
		retMap[liName] = liID
	}
	return retMap, nil
}

// used by batchCreateSegment4LayerListener
func (c *Clb) create4LayerListenerWithoutTargetGroup(
	region string, listener *networkextensionv1.Listener) (string, error) {
	// construct request for creating listener
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

	return listenerID, nil
}

// batchUpdate4LayerListener updates multiple 4 layer listener in one time
func (c *Clb) batchUpdate4LayerListener(
	region string, ingressListeners []*networkextensionv1.Listener,
	cloudListeners []*networkextensionv1.Listener) ([]bool, error) {

	updateErrArr := make([]bool, len(ingressListeners))
	backendRegMap := make(map[string][]networkextensionv1.ListenerBackend)
	backendDeregMap := make(map[string][]networkextensionv1.ListenerBackend)
	backendModMap := make(map[string][]networkextensionv1.ListenerBackend)
	for index, ingressListener := range ingressListeners {
		cloudListener := cloudListeners[index]
		if ingressListener.Spec.ListenerAttribute != nil &&
			needUpdateAttribute(cloudListener.Spec.ListenerAttribute, ingressListener.Spec.ListenerAttribute) {
			err := c.updateListenerAttrAndCerts(region, cloudListener.Status.ListenerID, ingressListener)
			if err != nil {
				blog.Warnf("updateListenerAttrAndCerts in update4LayerListener failed, err %s", err.Error())
				updateErrArr[index] = true
			}
		}
		// get different targets
		addBackends, delBackends, updateWeightBackends := getDiffBackendListBetweenTargetGroup(
			cloudListener.Spec.TargetGroup, ingressListener.Spec.TargetGroup)
		if len(addBackends) != 0 {
			backendRegMap[cloudListener.Status.ListenerID] = addBackends
		}
		if len(delBackends) != 0 {
			backendDeregMap[cloudListener.Status.ListenerID] = delBackends
		}
		if len(updateWeightBackends) != 0 {
			backendModMap[cloudListener.Status.ListenerID] = updateWeightBackends
		}
	}

	failedListenerMap := map[string]struct{}{}
	// deregister backends of multiple listener
	if len(backendDeregMap) != 0 {
		failedListenerIDs, err := c.batchDeregisterListenerBackend(
			region, ingressListeners[0].Spec.LoadbalancerID, backendDeregMap)
		if err != nil {
			return nil, err
		}
		for _, id := range failedListenerIDs {
			failedListenerMap[id] = struct{}{}
		}
	}
	// register backends of multiple listener
	if len(backendRegMap) != 0 {
		failedListenerIDs, err := c.batchRegisterListenerBackend(
			region, ingressListeners[0].Spec.LoadbalancerID, backendRegMap)
		if err != nil {
			return nil, err
		}
		for _, id := range failedListenerIDs {
			failedListenerMap[id] = struct{}{}
		}
	}
	// modify backends weights of multiple listener
	if len(backendModMap) != 0 {
		if err := c.batchChangeListenerBackendWeight(
			region, ingressListeners[0].Spec.LoadbalancerID, backendModMap); err != nil {
			return nil, err
		}
	}
	// collect register and de register err
	for index, li := range cloudListeners {
		if _, ok := failedListenerMap[li.Status.ListenerID]; ok {
			updateErrArr[index] = true
		}
	}
	return updateErrArr, nil
}

// batchCreate7LayerListener create multiple 4 layer listener
func (c *Clb) batchCreate7LayerListener(region string, listeners []*networkextensionv1.Listener) (
	map[string]string, error) {
	if len(listeners) == 0 {
		return nil, fmt.Errorf("listeners cannot be empty when batch create 4 layer listener")
	}
	req := tclb.NewCreateListenerRequest()
	req.LoadBalancerId = tcommon.StringPtr(listeners[0].Spec.LoadbalancerID)
	var portList []*int64
	var listenerNames []*string
	for _, li := range listeners {
		listenerNames = append(listenerNames, tcommon.StringPtr(li.GetName()))
		portList = append(portList, tcommon.Int64Ptr(int64(li.Spec.Port)))
	}
	req.Ports = portList
	req.ListenerNames = listenerNames
	listener := listeners[0]
	req.Protocol = tcommon.StringPtr(listener.Spec.Protocol)
	req.Certificate = transIngressCertificate(listener.Spec.Certificate)

	ctime := time.Now()
	listenerIDs, err := c.sdkWrapper.CreateListener(region, req)
	if err != nil {
		cloud.StatRequest("CreateListener", cloud.MetricAPIFailed, ctime, time.Now())
		return nil, err
	}
	cloud.StatRequest("CreateListener", cloud.MetricAPISuccess, ctime, time.Now())
	if len(listenerIDs) != len(listeners) {
		blog.Errorf("number of created listeners %d is not equal input number of input listeners %d",
			len(listenerIDs), len(listenerIDs))
		return nil, fmt.Errorf("number of created listeners %d is not equal input number of input listeners %d",
			len(listenerIDs), len(listenerIDs))
	}

	failedListenerIDMap := make(map[string]struct{})
	for liIndex, listener := range listeners {
		for _, rule := range listener.Spec.Rules {
			err := c.addListenerRule(region, listener.Spec.LoadbalancerID, listenerIDs[liIndex], rule)
			if err != nil {
				blog.Warnf("add listener rule %v for listener %s/%s failed, err %s",
					rule, listener.GetName(), listener.GetNamespace(), err.Error())
				failedListenerIDMap[listenerIDs[liIndex]] = struct{}{}
			}
		}
	}
	retMap := make(map[string]string)
	for index, id := range listenerIDs {
		if _, ok := failedListenerIDMap[id]; !ok {
			retMap[listeners[index].GetName()] = id
		}
	}
	return retMap, nil
}

// batchDeleteListener delete multiple listeners
func (c *Clb) batchDeleteListener(region, lbID string, listenerIDs []string) error {
	req := tclb.NewDeleteLoadBalancerListenersRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.ListenerIds = tcommon.StringPtrs(listenerIDs)

	ctime := time.Now()
	err := c.sdkWrapper.DeleteLoadbalanceListenners(region, req)
	if err != nil {
		cloud.StatRequest("DeleteLoadbalanceListenners", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("DeleteLoadbalanceListenners", cloud.MetricAPISuccess, ctime, time.Now())
	return nil
}

// batchUpdate7LayerListeners update multiple 7 layer listeners
func (c *Clb) batchUpdate7LayerListeners(region string, ingressListeners []*networkextensionv1.Listener,
	cloudListeners []*networkextensionv1.Listener) ([]bool, error) {
	if len(ingressListeners) == 0 {
		return nil, fmt.Errorf("length of listeners cannot be 0")
	}
	if len(ingressListeners) != len(cloudListeners) {
		return nil, fmt.Errorf(
			"length of listeners and cloud listeners should be the same, actually %d, %d",
			len(ingressListeners), len(cloudListeners))
	}

	listenerErrArr := make([]bool, len(ingressListeners))
	backendRegMap := make(map[string]map[string][]networkextensionv1.ListenerBackend)
	backendDeregMap := make(map[string]map[string][]networkextensionv1.ListenerBackend)
	backendModMap := make(map[string]map[string][]networkextensionv1.ListenerBackend)
	for index, li := range ingressListeners {
		backendAdds, backendDels, backendChanges, foundErr := c.updateHTTPListenerAndCollectRsChanges(
			region, li, cloudListeners[index])
		if len(backendAdds) != 0 {
			backendRegMap[cloudListeners[index].Status.ListenerID] = backendAdds
		}
		if len(backendDels) != 0 {
			backendDeregMap[cloudListeners[index].Status.ListenerID] = backendDels
		}
		if len(backendChanges) != 0 {
			backendModMap[cloudListeners[index].Status.ListenerID] = backendChanges
		}
		listenerErrArr[index] = foundErr
	}

	failedListenerMap := map[string]struct{}{}
	if len(backendDeregMap) != 0 {
		failedListenerIDs, err := c.batchDeregisterRuleBackend(
			region, ingressListeners[0].Spec.LoadbalancerID, backendDeregMap)
		if err != nil {
			return nil, err
		}
		for _, id := range failedListenerIDs {
			failedListenerMap[id] = struct{}{}
		}
	}
	if len(backendRegMap) != 0 {
		failedListenerIDs, err := c.batchRegisterRuleBackend(
			region, ingressListeners[0].Spec.LoadbalancerID, backendRegMap)
		if err != nil {
			return nil, err
		}
		for _, id := range failedListenerIDs {
			failedListenerMap[id] = struct{}{}
		}
	}
	if len(backendModMap) != 0 {
		if err := c.batchChangeRuleBackendWeight(
			region, ingressListeners[0].Spec.LoadbalancerID, backendModMap); err != nil {
			return nil, err
		}
	}
	// collect register and de register err
	for index, li := range cloudListeners {
		if _, ok := failedListenerMap[li.Status.ListenerID]; ok {
			listenerErrArr[index] = true
		}
	}
	return listenerErrArr, nil
}

// update 7 layer listener and collect rs change
// return the map of the map of targets to register, targets to be deregister, the map of targets to modify weight
func (c *Clb) updateHTTPListenerAndCollectRsChanges(
	region string, ingressListener, cloudListener *networkextensionv1.Listener) (
	map[string][]networkextensionv1.ListenerBackend,
	map[string][]networkextensionv1.ListenerBackend,
	map[string][]networkextensionv1.ListenerBackend, bool) {

	foundErr := false
	batchRegTargets := make(map[string][]networkextensionv1.ListenerBackend)
	batchDeregTargets := make(map[string][]networkextensionv1.ListenerBackend)
	batchModWeightTargets := make(map[string][]networkextensionv1.ListenerBackend)
	// if listener certificate is defined and is different from remote cloud listener attribute, then do update
	if ingressListener.Spec.Certificate != nil &&
		!reflect.DeepEqual(ingressListener.Spec.Certificate, cloudListener.Spec.Certificate) {
		err := c.updateListenerAttrAndCerts(region, cloudListener.Status.ListenerID, ingressListener)
		if err != nil {
			blog.Warnf("updateListenerAttrAndCerts in updateHTTPListener failed, err %s", err.Error())
			foundErr = true
		}
	}
	// get differet rules
	addRules, delRules, updateOldRules, updateRules := getDiffBetweenListenerRule(cloudListener, ingressListener)
	// do delete rules
	for _, rule := range delRules {
		err := c.deleteListenerRule(region, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID, rule)
		if err != nil {
			blog.Warnf("delete listener rule %v of lb %s listener %s failed, err %s",
				rule, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID, err.Error())
			foundErr = true
		}
	}
	// do add rules
	for _, rule := range addRules {
		err := c.addListenerRule(region, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID, rule)
		if err != nil {
			blog.Warnf("add listener rule %v of lb %s listener %s failed, err %s",
				rule, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID, err.Error())
			foundErr = true
		}
	}
	// do update rules
	for index, rule := range updateRules {
		existedRule := updateOldRules[index]
		if needUpdateAttribute(existedRule.ListenerAttribute, rule.ListenerAttribute) {
			err := c.updateRuleAttr(region, cloudListener.Spec.LoadbalancerID,
				cloudListener.Status.ListenerID, existedRule.RuleID, rule)
			if err != nil {
				blog.Warnf("update rule %v of lb %s listener %s attribute failed, err %s",
					rule, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID, err.Error())
				foundErr = true
			}
		}
		// get different targets
		addBackends, delBackends, updateWeightBackends := getDiffBackendListBetweenTargetGroup(
			existedRule.TargetGroup, rule.TargetGroup)
		if len(addBackends) != 0 {
			batchRegTargets[existedRule.RuleID] = addBackends
		}
		if len(delBackends) != 0 {
			batchDeregTargets[existedRule.RuleID] = delBackends
		}
		if len(updateWeightBackends) != 0 {
			batchModWeightTargets[existedRule.RuleID] = updateWeightBackends
		}
	}
	return batchRegTargets, batchDeregTargets, batchModWeightTargets, foundErr
}

// return failed listener id list
func (c *Clb) batchRegisterRuleBackend(region, lbID string,
	ruleBackendMap map[string]map[string][]networkextensionv1.ListenerBackend) ([]string, error) {
	req := tclb.NewBatchRegisterTargetsRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	for listenerID, liRuleMap := range ruleBackendMap {
		for ruleID, backendList := range liRuleMap {
			for _, backend := range backendList {
				req.Targets = append(req.Targets, &tclb.BatchTarget{
					ListenerId: tcommon.StringPtr(listenerID),
					LocationId: tcommon.StringPtr(ruleID),
					EniIp:      tcommon.StringPtr(backend.IP),
					Port:       tcommon.Int64Ptr(int64(backend.Port)),
					Weight:     tcommon.Int64Ptr(int64(backend.Weight)),
				})
			}
		}
	}
	if len(req.Targets) == 0 {
		return nil, fmt.Errorf("BatchRegisterTargets targets cannot be empty")
	}
	ctime := time.Now()
	failedListenerIDs := c.sdkWrapper.BatchRegisterTargets(region, req)
	if len(failedListenerIDs) != 0 {
		cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
	} else {
		cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
	}
	return failedListenerIDs, nil
}

// return failed listener id list
func (c *Clb) batchDeregisterRuleBackend(region, lbID string,
	ruleBackendMap map[string]map[string][]networkextensionv1.ListenerBackend) ([]string, error) {
	req := tclb.NewBatchDeregisterTargetsRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	for listenerID, liRuleMap := range ruleBackendMap {
		for ruleID, backendList := range liRuleMap {
			for _, backend := range backendList {
				req.Targets = append(req.Targets, &tclb.BatchTarget{
					ListenerId: tcommon.StringPtr(listenerID),
					LocationId: tcommon.StringPtr(ruleID),
					EniIp:      tcommon.StringPtr(backend.IP),
					Port:       tcommon.Int64Ptr(int64(backend.Port)),
					Weight:     tcommon.Int64Ptr(int64(backend.Weight)),
				})
			}
		}
	}
	if len(req.Targets) == 0 {
		return nil, fmt.Errorf("BatchDeregisterTargets targets cannot be empty")
	}
	ctime := time.Now()
	failedListenerIDs := c.sdkWrapper.BatchDeregisterTargets(region, req)
	if len(failedListenerIDs) != 0 {
		cloud.StatRequest("BatchDeregisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
	} else {
		cloud.StatRequest("BatchDeregisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
	}
	return failedListenerIDs, nil
}

func (c *Clb) batchChangeRuleBackendWeight(region, lbID string,
	ruleBackendMap map[string]map[string][]networkextensionv1.ListenerBackend) error {
	req := tclb.NewBatchModifyTargetWeightRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	for listenerID, liRuleMap := range ruleBackendMap {
		for ruleID, backendList := range liRuleMap {
			newWeightRule := &tclb.RsWeightRule{
				ListenerId: tcommon.StringPtr(listenerID),
				LocationId: tcommon.StringPtr(ruleID),
			}
			for _, backend := range backendList {
				newWeightRule.Targets = append(newWeightRule.Targets, &tclb.Target{
					EniIp:  tcommon.StringPtr(backend.IP),
					Port:   tcommon.Int64Ptr(int64(backend.Port)),
					Weight: tcommon.Int64Ptr(int64(backend.Weight)),
				})
			}
			req.ModifyList = append(req.ModifyList, newWeightRule)
		}
	}
	ctime := time.Now()
	if err := c.sdkWrapper.BatchModifyTargetWeight(region, req); err != nil {
		cloud.StatRequest("BatchModifyTargetWeight", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("BatchModifyTargetWeight", cloud.MetricAPISuccess, ctime, time.Now())
	return nil
}

func (c *Clb) batchRegisterListenerBackend(region, lbID string,
	liBackendMap map[string][]networkextensionv1.ListenerBackend) ([]string, error) {
	req := tclb.NewBatchRegisterTargetsRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	for listenerID, backendList := range liBackendMap {
		for _, backend := range backendList {
			req.Targets = append(req.Targets, &tclb.BatchTarget{
				ListenerId: tcommon.StringPtr(listenerID),
				EniIp:      tcommon.StringPtr(backend.IP),
				Port:       tcommon.Int64Ptr(int64(backend.Port)),
				Weight:     tcommon.Int64Ptr(int64(backend.Weight)),
			})
		}
	}
	if len(req.Targets) == 0 {
		return nil, fmt.Errorf("BatchRegisterTargets targets cannot be empty")
	}
	ctime := time.Now()
	failedListenerIDs := c.sdkWrapper.BatchRegisterTargets(region, req)
	if len(failedListenerIDs) != 0 {
		cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
	} else {
		cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
	}
	return failedListenerIDs, nil
}

func (c *Clb) batchDeregisterListenerBackend(region, lbID string,
	liBackendMap map[string][]networkextensionv1.ListenerBackend) ([]string, error) {
	req := tclb.NewBatchDeregisterTargetsRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	for listenerID, backendList := range liBackendMap {
		for _, backend := range backendList {
			req.Targets = append(req.Targets, &tclb.BatchTarget{
				ListenerId: tcommon.StringPtr(listenerID),
				EniIp:      tcommon.StringPtr(backend.IP),
				Port:       tcommon.Int64Ptr(int64(backend.Port)),
				Weight:     tcommon.Int64Ptr(int64(backend.Weight)),
			})
		}
	}
	if len(req.Targets) == 0 {
		return nil, fmt.Errorf("BatchDeregisterTargets targets cannot be empty")
	}
	ctime := time.Now()
	failedListenerIDs := c.sdkWrapper.BatchDeregisterTargets(region, req)
	if len(failedListenerIDs) != 0 {
		cloud.StatRequest("BatchDeregisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
	} else {
		cloud.StatRequest("BatchDeregisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
	}
	return failedListenerIDs, nil
}

func (c *Clb) batchChangeListenerBackendWeight(region, lbID string,
	liBackendMap map[string][]networkextensionv1.ListenerBackend) error {
	req := tclb.NewBatchModifyTargetWeightRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	for listenerID, backendList := range liBackendMap {
		newWeightRule := &tclb.RsWeightRule{
			ListenerId: tcommon.StringPtr(listenerID),
		}
		for _, backend := range backendList {
			newWeightRule.Targets = append(newWeightRule.Targets, &tclb.Target{
				EniIp:  tcommon.StringPtr(backend.IP),
				Port:   tcommon.Int64Ptr(int64(backend.Port)),
				Weight: tcommon.Int64Ptr(int64(backend.Weight)),
			})
		}
		req.ModifyList = append(req.ModifyList, newWeightRule)
	}
	ctime := time.Now()
	if err := c.sdkWrapper.BatchModifyTargetWeight(region, req); err != nil {
		cloud.StatRequest("BatchModifyTargetWeight", cloud.MetricAPIFailed, ctime, time.Now())
		return err
	}
	cloud.StatRequest("BatchModifyTargetWeight", cloud.MetricAPISuccess, ctime, time.Now())
	return nil
}
