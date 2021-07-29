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

package qcloudif

import (
	cloudListenerType "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"
)

//ClbAdapter interface to operate clb
type ClbAdapter interface {
	//loadbalance
	CreateLoadBalance(lb *cloudListenerType.CloudLoadBalancer) (lbID string, vips []string, err error)
	DescribeLoadBalance(name string) (*cloudListenerType.CloudLoadBalancer, bool, error)

	//listener
	CreateListener(listener *cloudListenerType.CloudListener) (listenerID string, err error)
	DeleteListener(lbID, listenerID string) error
	DescribeListener(lbID, listenerID string, port int) (listener *cloudListenerType.CloudListener, isExisted bool, err error)
	ModifyListenerAttribute(listener *cloudListenerType.CloudListener) (err error)

	//rule
	CreateRules(lbID, listenerID string, rules cloudListenerType.RuleList) error
	DeleteRule(lbID, listenerID, domain, url string) error
	DescribeRuleByDomainAndURL(loadBalanceID, listenerID, Domain, URL string) (rule *cloudListenerType.Rule, isExisted bool, err error)
	ModifyRuleAttribute(loadBalanceID, listenerID string, rule *cloudListenerType.Rule) error

	//7 layer backend
	Register7LayerBackends(lbID, listenerID, ruleID string, backendsRegister cloudListenerType.BackendList) error
	DeRegister7LayerBackends(lbID, listenerID, ruleID string, backendsDeRegister cloudListenerType.BackendList) error

	//4 layer backend
	Register4LayerBackends(lbID, listenerID string, backendsRegister cloudListenerType.BackendList) error
	DeRegister4LayerBackends(lbID, listenerID string, backendsDeRegister cloudListenerType.BackendList) error

	// list all listener
	ListListener(lbID string) ([]*cloudListenerType.CloudListener, error)
}
