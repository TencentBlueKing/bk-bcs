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
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// desribe listener info and listener targets with port list, used by batch operation
func (c *Clb) batchDescribeListeners(region, lbID string, ports []int) (
	map[string]*networkextensionv1.Listener, error) {
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
	if len(resp.Response.Listeners) == 0 {
		return nil, nil
	}

	listenerIDs, retListenerMap, ruleIDAttrMap, ruleIDCertMap := transferCloudListener(lbID, resp, portMap)

	// 2. describe listener targets
	if len(listenerIDs) == 0 {
		return retListenerMap, nil
	}
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
		protocol := *retLi.Protocol
		if common.InLayer7Protocol(protocol) {
			for _, retRule := range retLi.Rules {
				rules = append(rules, networkextensionv1.ListenerRule{
					RuleID:      *retRule.LocationId,
					Domain:      *retRule.Domain,
					Path:        *retRule.Url,
					TargetGroup: convertClbBackends(retRule.Targets),
					Certificate: ruleIDCertMap[*retRule.LocationId],
				})
			}
		} else if common.InLayer4Protocol(protocol) {
			tg = convertClbBackends(retLi.Targets)
		} else {
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
		var tmpLiName string
		if retLi.EndPort == nil {
			tmpLiName = common.GetListenerNameWithProtocol(lbID, *retLi.Protocol, int(*retLi.Port), 0)
		} else {
			tmpLiName = common.GetListenerNameWithProtocol(
				lbID, *retLi.Protocol, int(*retLi.Port), int(*retLi.EndPort))
		}

		li, ok := retListenerMap[tmpLiName]
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
	map[string]cloud.Result, error) {
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
	failedListenerIDMap := make(map[string]error)
	if len(tgReq.Targets) != 0 {
		ctime := time.Now()
		failedListenerIDs := c.sdkWrapper.BatchRegisterTargets(region, tgReq)
		if len(failedListenerIDs) != 0 {
			cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
		} else {
			cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
		}
		for id, derr := range failedListenerIDs {
			failedListenerIDMap[id] = derr
		}
	}
	// collect listeners which is created successfully
	retMap := make(map[string]cloud.Result)
	for index, id := range listenerIDs {
		if derr, ok := failedListenerIDMap[id]; !ok {
			retMap[listeners[index].GetName()] = cloud.Result{IsError: false, Res: id}
		} else {
			retMap[listeners[index].GetName()] = cloud.Result{IsError: true, Err: derr}
		}
	}
	return retMap, nil
}

// create multiple 4 layer listener segment
// on tencent cloud, listener segment only support tcp and udp
func (c *Clb) batchCreateSegment4LayerListener(
	region string, listeners []*networkextensionv1.Listener) (map[string]cloud.Result, error) {
	if len(listeners) == 0 {
		return nil, fmt.Errorf("listeners cannot be empty when batch create 4 layer listener segment")
	}

	// on tencent cloud, api cannot create multiple listener segment in one time
	// use goroutine to speed up
	failedListenerNameMap := sync.Map{}
	successListenerNameMap := sync.Map{}
	// 使用channel研制goroutine数量
	ch := make(chan struct{}, MaxSegmentListenerCurrentCreateEachTime)
	wg := sync.WaitGroup{}
	wg.Add(len(listeners))
	for _, li := range listeners {
		ch <- struct{}{}
		go func(li *networkextensionv1.Listener) {
			defer func() {
				wg.Done()
				<-ch
			}()
			// 记录每个监听器是否创建成功
			listenerID, err := c.createL4ListenerWithoutTg(region, li)
			if err != nil {
				err = errors.Wrapf(err, "create 4 layer listener %s/%s failed", li.GetNamespace(), li.GetName())
				blog.Warnf("%v", err)
				failedListenerNameMap.Store(li.GetName(), err)
				return
			}
			successListenerNameMap.Store(li.GetName(), listenerID)
		}(li)
	}
	wg.Wait()

	// collect all targets and register them in one time
	tgReq := tclb.NewBatchRegisterTargetsRequest()
	tgReq.LoadBalancerId = tcommon.StringPtr(listeners[0].Spec.LoadbalancerID)
	for _, li := range listeners {
		// 只处理创建成功的监听器
		if _, ok := failedListenerNameMap.Load(li.GetName()); ok {
			continue
		}
		if li.Spec.TargetGroup != nil && len(li.Spec.TargetGroup.Backends) != 0 {
			for _, backend := range li.Spec.TargetGroup.Backends {
				liName, _ := successListenerNameMap.Load(li.GetName())
				tgReq.Targets = append(tgReq.Targets, &tclb.BatchTarget{
					ListenerId: tcommon.StringPtr(liName.(string)),
					EniIp:      tcommon.StringPtr(backend.IP),
					Port:       tcommon.Int64Ptr(int64(backend.Port)),
					Weight:     tcommon.Int64Ptr(int64(backend.Weight)),
				})
			}
		}
	}
	failedListenerIDMap := make(map[string]error)
	if len(tgReq.Targets) != 0 {
		ctime := time.Now()
		failedListenerIDs := c.sdkWrapper.BatchRegisterTargets(region, tgReq)
		if len(failedListenerIDs) != 0 {
			cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPIFailed, ctime, time.Now())
		} else {
			cloud.StatRequest("BatchRegisterTargets", cloud.MetricAPISuccess, ctime, time.Now())
		}
		for id, derr := range failedListenerIDs {
			failedListenerIDMap[id] = derr
		}
	}
	// collect all listener
	// successListenerNameMap 记录创建成功的监听器
	// failedListenerNameMap 记录创建失败的监听器
	// failedListenerIDMap 记录注册target失败的监听器
	retMap := make(map[string]cloud.Result)
	for _, li := range listeners {
		if liID, ok := successListenerNameMap.Load(li.GetName()); ok {
			liIDStr := liID.(string)
			if err, inOk := failedListenerIDMap[liID.(string)]; inOk {
				retMap[li.GetName()] = cloud.Result{IsError: true, Err: err}
			} else {
				retMap[li.GetName()] = cloud.Result{IsError: false, Res: liIDStr}
			}
		} else {
			err, _ := failedListenerNameMap.Load(li.GetName())
			retMap[li.GetName()] = cloud.Result{IsError: true, Err: err.(error)}
		}
	}
	return retMap, nil
}

// used by batchCreateSegment4LayerListener
func (c *Clb) createL4ListenerWithoutTg(
	region string, listener *networkextensionv1.Listener) (string, error) {
	// construct request for creating listener
	req := tclb.NewCreateListenerRequest()
	req.LoadBalancerId = tcommon.StringPtr(listener.Spec.LoadbalancerID)
	req.Ports = []*int64{
		tcommon.Int64Ptr(int64(listener.Spec.Port)),
	}
	if listener.Spec.EndPort != 0 {
		req.EndPort = tcommon.Uint64Ptr(uint64(listener.Spec.EndPort))
	}
	req.ListenerNames = tcommon.StringPtrs([]string{listener.GetName()})
	req.Protocol = tcommon.StringPtr(listener.Spec.Protocol)
	// translate listenerAttribute to cloud request field
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
	cloudListeners []*networkextensionv1.Listener) ([]error, error) {

	updateErrArr := make([]error, len(ingressListeners))
	backendRegMap := make(map[string][]networkextensionv1.ListenerBackend)
	backendDeregMap := make(map[string][]networkextensionv1.ListenerBackend)
	backendModMap := make(map[string][]networkextensionv1.ListenerBackend)
	for index, ingressListener := range ingressListeners {
		cloudListener := cloudListeners[index]
		if ingressListener.Spec.ListenerAttribute != nil &&
			needUpdateAttribute(cloudListener.Spec.ListenerAttribute, ingressListener.Spec.ListenerAttribute) {
			err := c.updateListenerAttrAndCerts(region, cloudListener.Status.ListenerID, ingressListener)
			if err != nil {
				err = errors.Wrapf(err, "updateListenerAttrAndCerts in update4LayerListener failed")
				blog.Warnf(err.Error())
				updateErrArr[index] = multierror.Append(updateErrArr[index], err)
			}
		}
		// get different targets
		addBackends, delBackends, updateWeightBackends := compareTargetGroup(
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

	failedListenerMap := make(map[string]error)
	// deregister backends of multiple listener
	if len(backendDeregMap) != 0 {
		failedListenerIDs, err := c.batchDeregisterListenerBackend(
			region, ingressListeners[0].Spec.LoadbalancerID, backendDeregMap)
		if err != nil {
			return nil, err
		}
		for _, id := range failedListenerIDs {
			failedListenerMap[id] = multierror.Append(failedListenerMap[id],
				fmt.Errorf("batch deregister listener backend failed"))
		}
	}
	// register backends of multiple listener
	if len(backendRegMap) != 0 {
		failedListenerIDs, err := c.batchRegisterListenerBackend(
			region, ingressListeners[0].Spec.LoadbalancerID, backendRegMap)
		if err != nil {
			return nil, err
		}
		for id, derr := range failedListenerIDs {
			failedListenerMap[id] = multierror.Append(failedListenerMap[id], derr)
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
		if err, ok := failedListenerMap[li.Status.ListenerID]; ok {
			updateErrArr[index] = multierror.Append(updateErrArr[index], err)
		}
	}
	return updateErrArr, nil
}

// batchCreate7LayerListener create multiple 4 layer listener
func (c *Clb) batchCreate7LayerListener(region string, listeners []*networkextensionv1.Listener) (
	map[string]cloud.Result, error) {
	if len(listeners) == 0 {
		return nil, fmt.Errorf("listeners cannot be empty when batch create 7 layer listener")
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
	if listener.Spec.ListenerAttribute != nil {
		req.SniSwitch = tcommon.Int64Ptr(int64(listener.Spec.ListenerAttribute.SniSwitch))
		req.KeepaliveEnable = tcommon.Int64Ptr(int64(listener.Spec.ListenerAttribute.KeepAliveEnable))
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

	failedListenerIDMap := make(map[string]error)
	for liIndex, li := range listeners {
		for _, rule := range li.Spec.Rules {
			// 为每个监听器创建规则
			inErr := c.addListenerRule(region, li.Spec.LoadbalancerID, listenerIDs[liIndex],
				li.Spec.ListenerAttribute, rule)
			if inErr != nil {
				inErr = errors.Wrapf(inErr, "add listener rule %v for listener %s/%s failed", rule, li.GetName(),
					li.GetNamespace())
				blog.Warnf(inErr.Error())
				failedListenerIDMap[listenerIDs[liIndex]] = inErr
			}
		}
	}
	// 收集每个listener的执行结果
	// listener.Name -> Result
	retMap := make(map[string]cloud.Result)
	for index, id := range listenerIDs {
		if inErr, ok := failedListenerIDMap[id]; !ok {
			retMap[listeners[index].GetName()] = cloud.Result{IsError: false, Res: id}
		} else {
			retMap[listeners[index].GetName()] = cloud.Result{IsError: true, Err: inErr}
		}
	}
	return retMap, nil
}

// batchDeleteListener delete multiple listeners
func (c *Clb) batchDeleteListener(region, lbID string, listenerIDs []string) error {
	// It's possible delete all listeners when listenerIds empty
	if len(listenerIDs) == 0 {
		return fmt.Errorf("listenerIDs should not be empty when batch delete listener")
	}

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
	cloudListeners []*networkextensionv1.Listener) ([]error, error) {
	if len(ingressListeners) == 0 {
		return nil, fmt.Errorf("length of listeners cannot be 0")
	}
	if len(ingressListeners) != len(cloudListeners) {
		return nil, fmt.Errorf(
			"length of listeners and cloud listeners should be the same, actually %d, %d",
			len(ingressListeners), len(cloudListeners))
	}

	// 记录每个listener更新时的错误
	listenerErrArr := make([]error, len(ingressListeners))
	backendRegMap := make(map[string]map[string][]networkextensionv1.ListenerBackend)
	backendDeregMap := make(map[string]map[string][]networkextensionv1.ListenerBackend)
	backendModMap := make(map[string]map[string][]networkextensionv1.ListenerBackend)
	for index, li := range ingressListeners {
		backendAdds, backendDels, backendChanges, foundErr := c.updateL7ListenerAndCollectRsChanges(
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

	failedListenerMap := make(map[string]error)
	if len(backendDeregMap) != 0 {
		failedListenerIDs, err := c.batchDeregisterRuleBackend(
			region, ingressListeners[0].Spec.LoadbalancerID, backendDeregMap)
		if err != nil {
			return nil, err
		}
		for _, id := range failedListenerIDs {
			failedListenerMap[id] = errors.New("batch deregister targets failed")
		}
	}
	if len(backendRegMap) != 0 {
		failedListenerIDs, err := c.batchRegisterRuleBackend(
			region, ingressListeners[0].Spec.LoadbalancerID, backendRegMap)
		if err != nil {
			return nil, err
		}
		for id, derr := range failedListenerIDs {
			failedListenerMap[id] = derr
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
		if err, ok := failedListenerMap[li.Status.ListenerID]; ok {
			listenerErrArr[index] = err
		}
	}
	return listenerErrArr, nil
}

// update 7 layer listener and collect rs change
// return the map of the map of targets to register, targets to be deregister, the map of targets to modify weight
func (c *Clb) updateL7ListenerAndCollectRsChanges(
	region string, ingressListener, cloudListener *networkextensionv1.Listener) (
	map[string][]networkextensionv1.ListenerBackend,
	map[string][]networkextensionv1.ListenerBackend,
	map[string][]networkextensionv1.ListenerBackend, error) {

	var resultErr error
	batchRegTargets := make(map[string][]networkextensionv1.ListenerBackend)
	batchDeregTargets := make(map[string][]networkextensionv1.ListenerBackend)
	batchModWeightTargets := make(map[string][]networkextensionv1.ListenerBackend)
	// if listener certificate is defined and is different from remote cloud listener attribute, then do update
	if ingressListener.Spec.Certificate != nil &&
		!reflect.DeepEqual(ingressListener.Spec.Certificate, cloudListener.Spec.Certificate) {
		err := c.updateListenerAttrAndCerts(region, cloudListener.Status.ListenerID, ingressListener)
		if err != nil {
			err = errors.Wrapf(err, "updateListenerAttrAndCerts in updateHTTPListener failed")
			blog.Warnf(err.Error())
			resultErr = multierror.Append(resultErr, err)
		}
	}
	// get different rules
	addRules, delRules, updateOldRules, updateRules := getDiffBetweenListenerRule(cloudListener, ingressListener)
	// do delete rules
	for _, rule := range delRules {
		err := c.deleteListenerRule(region, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID, rule)
		if err != nil {
			err = errors.Wrapf(err, "delete listener rule %v of lb %s listener %s failed", rule,
				cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID)
			blog.Warnf(err.Error())
			resultErr = multierror.Append(resultErr, err)
		}
	}
	// do add rules
	for _, rule := range addRules {
		err := c.addListenerRule(region, cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID,
			cloudListener.Spec.ListenerAttribute, rule)
		if err != nil {
			err = errors.Wrapf(err, "add listener rule %v of lb %s listener %s failed", rule,
				cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID)
			blog.Warnf(err.Error())
			resultErr = multierror.Append(resultErr, err)
		}
	}
	// do update rules
	for index, rule := range updateRules {
		existedRule := updateOldRules[index]
		if needUpdateAttribute(existedRule.ListenerAttribute, rule.ListenerAttribute) {
			err := c.updateRuleAttr(region, cloudListener.Spec.LoadbalancerID,
				cloudListener.Status.ListenerID, existedRule.RuleID, rule)
			if err != nil {
				err = errors.Wrapf(err, "update rule %v of lb %s listener %s attribute failed", rule,
					cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID)
				blog.Warnf(err.Error())
				resultErr = multierror.Append(resultErr, err)
			}
		}
		if !reflect.DeepEqual(rule.Certificate, existedRule.Certificate) {
			err := c.updateDomainAttributes(region, cloudListener.Spec.LoadbalancerID,
				cloudListener.Status.ListenerID, rule)
			if err != nil {
				err = errors.Wrapf(err, "update rule %v of lb %s listener %s domain attribute failed", rule,
					cloudListener.Spec.LoadbalancerID, cloudListener.Status.ListenerID)
				blog.Warnf(err.Error())
				resultErr = multierror.Append(resultErr, err)
			}
		}
		// get different targets
		addBackends, delBackends, updateWeightBackends := compareTargetGroup(
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
	return batchRegTargets, batchDeregTargets, batchModWeightTargets, resultErr
}

// return failed listener id list
func (c *Clb) batchRegisterRuleBackend(region, lbID string,
	ruleBackendMap map[string]map[string][]networkextensionv1.ListenerBackend) (map[string]error, error) {
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

// batchChangeRuleBackendWeight change rule backend weight
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

// batchRegisterListenerBackend register backend to target group
func (c *Clb) batchRegisterListenerBackend(region, lbID string,
	liBackendMap map[string][]networkextensionv1.ListenerBackend) (map[string]error, error) {
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

// batchDeregisterListenerBackend deregister backend from target
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

// batchChangeListenerBackendWeight batch change weight
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

// resolveCreateListener create listener
func (c *Clb) resolveCreateListener(region string, addListeners []*networkextensionv1.Listener) map[string]cloud.Result {
	retMap := make(map[string]cloud.Result)
	addListenerGroups := splitListenersToDiffProtocol(addListeners)
	for _, group := range addListenerGroups {
		if len(group) != 0 {
			// split listeners batch by its attribute and cert
			batches := splitListenersToDiffBatch(group)
			for _, batch := range batches {
				liIDMap, err := c.resolveCreateListenerBatch(region, group[0].Spec.Protocol, batch)
				if err != nil {
					for _, listener := range batch {
						retMap[listener.GetName()] = cloud.Result{IsError: true, Err: err}
					}
					continue
				}

				for liName, res := range liIDMap {
					retMap[liName] = res
				}
			}
		}
	}
	return retMap
}

// resolveCreateListenerBatch listeners in same batch have same lb and attribute&cert
func (c *Clb) resolveCreateListenerBatch(region, protocol string, batch []*networkextensionv1.Listener) (map[string]cloud.Result,
	error) {
	if common.InLayer7Protocol(protocol) {
		liIDMap, err := c.batchCreate7LayerListener(region, batch)
		if err != nil {
			blog.Warnf("batch create 7 layer listener failed, err %s", err.Error())
			return nil, err
		}
		return liIDMap, nil
	} else if common.InLayer4Protocol(protocol) {
		liIDMap, err := c.batchCreate4LayerListener(region, batch)
		if err != nil {
			blog.Warnf("batch create 4 layer listener failed, err %s", err.Error())
			return nil, err
		}
		return liIDMap, nil
	} else {
		blog.Warnf("invalid batch protocol %s", protocol)
		return nil, fmt.Errorf("invalid batch protocol %s", protocol)
	}
}

// resolveUpdateListener update listeners
func (c *Clb) resolveUpdateListener(lbID, region string, updatedListeners []*networkextensionv1.Listener,
	cloudListenerMap map[string]*networkextensionv1.Listener,
) map[string]cloud.Result {
	retMap := make(map[string]cloud.Result)
	// split listener by protocol
	updateListenerGroups := splitListenersToDiffProtocol(updatedListeners)
	for _, group := range updateListenerGroups {
		if len(group) != 0 {
			cloudListenerGroup := make([]*networkextensionv1.Listener, 0)
			for _, li := range group {
				cloudListenerGroup = append(cloudListenerGroup, cloudListenerMap[common.GetListenerNameWithProtocol(
					lbID, li.Spec.Protocol, li.Spec.Port, li.Spec.EndPort)])
			}

			isErrArr, err := c.resolveUpdateListenerGroup(region, group, cloudListenerGroup)
			// 如果err，认为这一批listener全部更新失败
			if err != nil {
				for _, listener := range group {
					retMap[listener.GetName()] = cloud.Result{IsError: true, Err: err}
				}
				continue
			}
			// 根据云接口返回的Err，细分每个listener的失败原因
			for index, inErr := range isErrArr {
				if inErr == nil {
					retMap[group[index].GetName()] = cloud.Result{
						IsError: false,
						Res:     cloudListenerGroup[index].Status.ListenerID}
				} else {
					retMap[group[index].GetName()] = cloud.Result{IsError: true, Err: inErr}
					blog.Warnf("update 7 layer listener %s failed in batch, err: %+v", group[index].GetName(), inErr)
				}
			}
		}
	}
	return retMap
}

// resolveUpdateListenerGroup update listener group.
// group listeners have same protocol
func (c *Clb) resolveUpdateListenerGroup(region string, group []*networkextensionv1.Listener,
	cloudListenerGroup []*networkextensionv1.Listener) ([]error, error) {
	protocol := group[0].Spec.Protocol

	if common.InLayer7Protocol(protocol) {
		// layer7 -> http / https
		isErrArr, err := c.batchUpdate7LayerListeners(region, group, cloudListenerGroup)
		if err != nil {
			blog.Warnf("batch update 7 layer listeners %s failed, err %s", getListenerNames(group), err.Error())
			return nil, err
		}
		return isErrArr, nil
	} else if common.InLayer4Protocol(protocol) {
		// layer4 -< tcp / udp
		isErrArr, err := c.batchUpdate4LayerListener(region, group, cloudListenerGroup)
		if err != nil {
			blog.Infof("batch update 4 layer listeners %s failed, err %s", getListenerNames(group), err.Error())
			return nil, err
		}
		return isErrArr, nil
	} else {
		blog.Warnf("invalid batch protocol %s", group[0].Spec.Protocol)
		return nil, fmt.Errorf("invalid batch protocol %s", group[0].Spec.Protocol)
	}
}
