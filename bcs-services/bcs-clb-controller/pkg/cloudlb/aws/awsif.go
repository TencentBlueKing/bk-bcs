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

package aws

import (
	loadbalance "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
)

//awsInterface interface to call aws api
type awsElbInterface interface {
	//loadbalance
	awsCreateLoadBalancer(subNets, securityGroups []string, name, networkType, lbType string) (loadBalancerArn string, err error)
	awsDescribeLoadBalancer(name string) (*loadbalance.CloudLoadBalancer, bool, error)
	//listener
	awsCreateListener(targetGroup, loadBalancerArn, protocol string, port int64) (listenerArn string, err error)
	awsDeleteListener(listenerArn string) error
	awsUpdateListener(listenerArn string, port int64) error
	awsDescribeListener(lbArn, listenerArn string) (listener *loadbalance.CloudListener, isExisted bool, err error)
	//rule
	awsCreateRule(targetGroup, listenerArn, domain, path string, priority int64) (ruleArn string, err error)
	awsUpdateRule(ruleArn, targetGroup, domain, path string) error
	awsDeleteRule(ruleArn string) error
	awsDescribeRule(listenerArn, ruleArn string) (rule *loadbalance.Rule, isExisted bool, err error)
	//target group
	awsCreateTargetGroup(targetGroupName, vpcID, protocol, healthCheckPath string, targetGroupPort int64) (targetGroupArn string, err error)
	awsDeleteTargetGroup(targetGroupArn string) error
	//target
	awsRegisterTargets(targetGroupArn string, backendsRegister loadbalance.BackendList) error
	awsDeRegisterTargets(targetGroupArn string, backendsDeRegister loadbalance.BackendList) error
	//security group
	awsAuthorizeSecurityGroupIngress(securityGroupID *string, startPort, endPort int64) error
}
