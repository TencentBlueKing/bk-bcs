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
	"reflect"
	"strconv"

	"bk-bcs/bcs-common/common/blog"

	loadbalance "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
)

func (clb *ClbClient) addListener(ls *loadbalance.CloudListener) error {
	// create name
	listenerID, err := clb.clbAdapter.CreateListener(ls)
	if err != nil {
		return fmt.Errorf("create listener failed, %s", err.Error())
	}
	// set listener id
	ls.Spec.ListenerID = listenerID

	if ls.Spec.Protocol == loadbalance.ClbListenerProtocolTCP || ls.Spec.Protocol == loadbalance.ClbListenerProtocolUDP {
		// register backends for 4 layer protocol
		// when failed, we do not record these backends in cache
		// max backends for each register is LimitationMaxBackendNumEachBind (20)
		effectedBackends := make([]*loadbalance.Backend, 0)
		for i := 0; i < len(ls.Spec.TargetGroup.Backends); i = i + LimitationMaxBackendNumEachBind {
			tmpBackends := GetBackendsSegment(ls.Spec.TargetGroup.Backends, i, LimitationMaxBackendNumEachBind)
			err := clb.clbAdapter.Register4LayerBackends(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, tmpBackends)
			if err != nil {
				blog.Warnf("register 4 layer listener backends %v failed, %s", tmpBackends, err.Error())
				break
			}
			effectedBackends = append(effectedBackends, tmpBackends...)
		}
		ls.Spec.TargetGroup.Backends = effectedBackends

	} else if ls.Spec.Protocol == loadbalance.ClbListenerProtocolHTTP || ls.Spec.Protocol == loadbalance.ClbListenerProtocolHTTPS {
		var successRuleList loadbalance.RuleList
		// each rule corresponds to a target group
		// when failed in creating rule, we do not record this rule in cache
		// the listener is new, so we call **doCreateRule** directly
		for _, rule := range ls.Spec.Rules {
			err := clb.doCreateRule(ls, rule)
			if err != nil {
				blog.Warnf("createRule domain:%s url:%s targetGroup:%v failed, %s", rule.Domain, rule.URL, rule.TargetGroup, err.Error())
				continue
			}
			successRuleList = append(successRuleList, rule)
		}
		ls.Spec.Rules = successRuleList

	} else {
		return fmt.Errorf("add listener failed, unsupported listener protocol %s", ls.Spec.Protocol)
	}

	return nil
}

//update4LayerListener update 4 layer listener
//when listener protocol
func (clb *ClbClient) update4LayerListener(old, cur *loadbalance.CloudListener) error {
	if old.IsEqual(cur) {
		blog.Warnf("the old listener %s is equal to the new one, no need to update", old.GetName())
		return nil
	}

	if !old.Spec.TargetGroup.IsAttrEqual(cur.Spec.TargetGroup) {
		blog.Infof("update attr of listener %s.%s", old.GetName(), old.GetNamespace())
		err := clb.clbAdapter.ModifyListenerAttribute(cur)
		if err != nil {
			blog.Errorf("update attr of listener %s.%s failed, err %s", old.GetName(), old.GetNamespace(), err.Error())
			return fmt.Errorf("update attr of listener %s.%s failed, err %s", old.GetName(), old.GetNamespace(), err.Error())
		}
	}

	//update backends
	backendsDel, backendsNew := old.Spec.TargetGroup.GetDiffBackend(cur.Spec.TargetGroup)

	blog.Infof("get %d backend to del: %v", len(backendsDel), backendsDel)
	blog.Infof("get %d backend to add: %v", len(backendsNew), backendsNew)
	for _, backend := range backendsNew {
		clbBackendsAddMetric.WithLabelValues(backend.IP, strconv.Itoa(backend.Port)).Inc()
	}
	for _, backend := range backendsDel {
		clbBackendsDeleteMetric.WithLabelValues(backend.IP, strconv.Itoa(backend.Port)).Inc()
	}

	//deregister old backend
	if len(backendsDel) > 0 {
		for i := 0; i < len(backendsDel); i = i + LimitationMaxBackendNumEachBind {
			tmpBackends := GetBackendsSegment(backendsDel, i, LimitationMaxBackendNumEachBind)
			err := clb.clbAdapter.DeRegister4LayerBackends(old.Spec.LoadBalancerID, old.Spec.ListenerID, tmpBackends)
			if err != nil {
				blog.Warnf("QCloud4LayerDeRegisterTargets backends %v failed, %s", tmpBackends, err.Error())
				cur.Spec.TargetGroup.AddBackends(tmpBackends)
			}
		}
	}

	//register new backends
	if len(backendsNew) > 0 {
		for i := 0; i < len(backendsNew); i = i + LimitationMaxBackendNumEachBind {
			tmpBackends := GetBackendsSegment(backendsNew, i, LimitationMaxBackendNumEachBind)
			err := clb.clbAdapter.Register4LayerBackends(old.Spec.LoadBalancerID, old.Spec.ListenerID, tmpBackends)
			if err != nil {
				blog.Warnf("QCloud4LayerRegisterBackend backends %v failed, %s", tmpBackends, err.Error())
				cur.Spec.TargetGroup.RemoveBackend(tmpBackends)
			}
		}
	}

	blog.Infof("update 4 layer listener %s done", cur.GetName())
	return nil
}

func (clb *ClbClient) update7LayerListener(old, cur *loadbalance.CloudListener) error {

	if old.IsEqual(cur) {
		blog.Warn("the old listener is equal to the new one, no need to update")
		return nil
	}

	if cur.Spec.Protocol == loadbalance.ClbListenerProtocolHTTPS {
		if !reflect.DeepEqual(old.Spec.TLS, cur.Spec.TLS) {
			blog.Infof("update listener %s.%s tls config", old.GetName(), old.GetNamespace())
			err := clb.clbAdapter.ModifyListenerAttribute(cur)
			if err != nil {
				blog.Errorf("update attr of listener %s.%s failed, err %s", old.GetName(), old.GetNamespace(), err.Error())
				return fmt.Errorf("update attr of listener %s.%s failed, err %s", old.GetName(), old.GetNamespace(), err.Error())
			}
		}
	}

	//get deleted rules and new rules
	delRules, newRules := old.GetDiffRules(cur)

	//get updated rules
	olds, updateRules := old.GetUpdateRules(cur)

	//update rules
	for index, ruleUpdate := range updateRules {
		err := clb.updateRule(old, olds[index], ruleUpdate)
		if err != nil {
			blog.Warnf("updateRule failed listenerId:%s Id:%s domain:%s url:%s, %s",
				old.Spec.ListenerID, olds[index].ID, ruleUpdate.Domain, ruleUpdate.URL, err.Error())
			continue
		}
	}

	//delete rules
	for _, ruleDel := range delRules {
		err := clb.deleteRule(old, ruleDel)
		if err != nil {
			blog.Warnf("createRule failed listenerId:%s Id:%s domain:%s url:%s, %s",
				old.Spec.ListenerID, ruleDel.ID, ruleDel.Domain, ruleDel.URL, err.Error())
			continue
		}
	}

	//add new rules
	for _, ruleNew := range newRules {
		err := clb.createRule(old, ruleNew)
		if err != nil {
			blog.Warnf("createRule failed listenerId:%s domain:%s url:%s, %s",
				old.Spec.ListenerID, ruleNew.Domain, ruleNew.URL, err.Error())
			continue
		}
	}

	return nil

}

func (clb *ClbClient) createRule(ls *loadbalance.CloudListener, rule *loadbalance.Rule) error {

	_, isExisted, err := clb.clbAdapter.DescribeRuleByDomainAndURL(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, rule.Domain, rule.URL)
	if err != nil {
		return fmt.Errorf("QCloudDescribeRuleByDomainAndURL failed, %s", err.Error())
	}
	if isExisted {
		blog.Warnf("Rule domain:%s url:%s to be created already existed in qcloud listener %s, clean it first", rule.Domain, rule.URL, ls.Spec.ListenerID)
		err := clb.clbAdapter.DeleteRule(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, rule.Domain, rule.URL)
		if err != nil {
			return fmt.Errorf("delete rule (domain:%s, url%s) failed, err %s", rule.Domain, rule.URL, err.Error())
		}
	}
	return clb.doCreateRule(ls, rule)
}

//call api to create rule
//TODO: create multiple rules together. temporarily we are not sure if the create rules api is atomic, so create rule one by one
func (clb *ClbClient) doCreateRule(ls *loadbalance.CloudListener, rule *loadbalance.Rule) error {

	var ruleList loadbalance.RuleList
	ruleList = append(ruleList, rule)
	err := clb.clbAdapter.CreateRules(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, ruleList)
	if err != nil {
		return fmt.Errorf("create rule failed, %s", err.Error())
	}

	descRule, isExisted, err := clb.clbAdapter.DescribeRuleByDomainAndURL(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, rule.Domain, rule.URL)
	if err != nil {
		return fmt.Errorf("describe rule (domain %s, url %s) failed after creation", rule.Domain, rule.URL)
	}
	if !isExisted {
		return fmt.Errorf("rule (domain %s, url %s) is not existed after creation", rule.Domain, rule.URL)
	}
	rule.ID = descRule.ID

	// when failed, we do not record these backends in cache
	// max backends for each register is LimitationMaxBackendNumEachBind (20)
	effectedBackends := make([]*loadbalance.Backend, 0)
	for i := 0; i < len(rule.TargetGroup.Backends); i = i + LimitationMaxBackendNumEachBind {
		tmpBackends := GetBackendsSegment(rule.TargetGroup.Backends, i, LimitationMaxBackendNumEachBind)
		err := clb.clbAdapter.Register7LayerBackends(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, rule.ID, tmpBackends)
		if err != nil {
			blog.Warnf("register 7 layer listener backends %v failed, %s", tmpBackends, err.Error())
			break
		}
		effectedBackends = append(effectedBackends, tmpBackends...)
	}
	rule.TargetGroup.Backends = effectedBackends

	blog.Info("Create rule domain:%s url:%s for listener %s successfully", rule.Domain, rule.URL, ls.GetName())
	return nil
}

func (clb *ClbClient) updateRule(ls *loadbalance.CloudListener, ruleOld *loadbalance.Rule, ruleUpdate *loadbalance.Rule) error {

	_, isExisted, err := clb.clbAdapter.DescribeRuleByDomainAndURL(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, ruleOld.Domain, ruleOld.URL)
	if err != nil {
		return fmt.Errorf("QCloudDescribeRule failed, %s", err.Error())
	}
	if !isExisted {
		blog.Warnf("old rule %s %s %s does not exists, try to create new rule", ruleOld.ID, ruleOld.Domain, ruleOld.URL)
		err := clb.doCreateRule(ls, ruleUpdate)
		if err != nil {
			return fmt.Errorf("createRule domain:%s url:%s for listener %s failed, %s", ruleUpdate.Domain, ruleUpdate.URL, ls.GetName(), err.Error())
		}
		return nil
	}
	//because the new rule struct has no id info
	ruleUpdate.ID = ruleOld.ID

	if !ruleOld.TargetGroup.IsAttrEqual(ruleUpdate.TargetGroup) {
		err := clb.clbAdapter.ModifyRuleAttribute(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, ruleUpdate)
		if err != nil {
			blog.Infof("modify rule of lb %s listener %s config failed, err %s", ls.Spec.LoadBalancerID, ls.Spec.ListenerID, err.Error())
			return fmt.Errorf("modify rule of lb %s listener %s config failed, err %s", ls.Spec.LoadBalancerID, ls.Spec.ListenerID, err.Error())
		}
	}

	backendsDel, backendsNew := ruleOld.TargetGroup.GetDiffBackend(ruleUpdate.TargetGroup)
	for _, backend := range backendsNew {
		clbBackendsAddMetric.WithLabelValues(backend.IP, strconv.Itoa(backend.Port)).Inc()
	}
	for _, backend := range backendsDel {
		clbBackendsDeleteMetric.WithLabelValues(backend.IP, strconv.Itoa(backend.Port)).Inc()
	}

	//2.1 deregister old backend
	if len(backendsDel) > 0 {
		for i := 0; i < len(backendsDel); i = i + LimitationMaxBackendNumEachBind {
			tmpBackends := GetBackendsSegment(backendsDel, i, LimitationMaxBackendNumEachBind)
			err := clb.clbAdapter.DeRegister7LayerBackends(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, ruleOld.ID, tmpBackends)
			if err != nil {
				blog.Warnf("Deregister 7 Layer Backends %v failed, %s", tmpBackends, err.Error())
				ruleUpdate.TargetGroup.AddBackends(tmpBackends)
			}
		}
	}

	//2.2 register new backend
	if len(backendsNew) > 0 {
		for i := 0; i < len(backendsNew); i = i + LimitationMaxBackendNumEachBind {
			tmpBackends := GetBackendsSegment(backendsNew, i, LimitationMaxBackendNumEachBind)
			err := clb.clbAdapter.Register7LayerBackends(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, ruleOld.ID, tmpBackends)
			if err != nil {
				blog.Warnf("Register 7Layer Backends %v failed, %s", tmpBackends, err.Error())
				ruleUpdate.TargetGroup.RemoveBackend(tmpBackends)
			}
		}
	}

	blog.Info("update rule id:%s domain:%s url:%s done", ruleUpdate.ID, ruleUpdate.Domain, ruleUpdate.URL)
	return nil
}

func (clb *ClbClient) deleteRule(ls *loadbalance.CloudListener, ruleOld *loadbalance.Rule) error {
	_, isExisted, err := clb.clbAdapter.DescribeRuleByDomainAndURL(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, ruleOld.Domain, ruleOld.URL)
	if err != nil {
		return fmt.Errorf("describe rule by domain %s url %s failed, %s", ruleOld.Domain, ruleOld.URL, err.Error())
	}
	if !isExisted {
		blog.Warnf("old rule (id:%s domain:%s url:%s) does not exists, no need to delete", ruleOld.ID, ruleOld.Domain, ruleOld.URL)
		return nil
	}

	err = clb.clbAdapter.DeleteRule(ls.Spec.LoadBalancerID, ls.Spec.ListenerID, ruleOld.Domain, ruleOld.URL)
	if err != nil {
		return fmt.Errorf("delete rule (id:%s domain:%s url:%s) failed, %s", ruleOld.ID, ruleOld.Domain, ruleOld.URL, err.Error())
	}
	return nil
}
