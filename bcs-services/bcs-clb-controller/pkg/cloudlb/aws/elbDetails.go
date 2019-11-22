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
	"bk-bcs/bcs-common/common/blog"
	"fmt"

	loadbalance "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
)

func (aws *ElbClient) getRulePriority() int64 {
	return 0
}

//createRulesFor7LayerListener
func (aws *ElbClient) createRulesFor7LayerListener(ls *loadbalance.CloudListener) error {
	//each rule corresponds to a target group
	for _, rule := range ls.Spec.Rules {
		err := aws.createRule(ls.Spec.ListenerID, rule)
		if err != nil {
			return fmt.Errorf("createTargetAndRule failed, %s", err.Error())
		}
	}
	return nil
}

//update4LayerListener
func (aws *ElbClient) update4LayerListener(old, cur *loadbalance.CloudListener) error {
	//check for service port change
	if cur.Spec.ListenPort != old.Spec.ListenPort {
		//update listen port
		err := aws.awsAPI.awsUpdateListener(cur.Spec.ListenerID, int64(cur.Spec.ListenPort))
		if err != nil {
			blog.Errorf("awsUpdateListener failed, %s", err.Error())
		}
		//no need to update default target group serivce port
		//because target group service port does not affect forwarding behavior
	}
	//update default target group backends
	backendsDel, backendsNew := old.Spec.TargetGroup.GetDiffBackend(cur.Spec.TargetGroup)

	//register new backend
	err := aws.awsAPI.awsRegisterTargets(old.Spec.TargetGroup.ID, backendsNew)
	if err != nil {
		return fmt.Errorf("awsRegisterTargets register new backend failed, %s", err.Error())
	}
	//deregister old backend
	err = aws.awsAPI.awsDeRegisterTargets(cur.Spec.TargetGroup.ID, backendsDel)
	if err != nil {
		return fmt.Errorf("awsDeRegisterTargets deregister backend failed, %s", err.Error())
	}
	return nil
}

//update7LayerListener
func (aws *ElbClient) update7LayerListener(old, cur *loadbalance.CloudListener) error {

	//check for service port change
	if cur.Spec.ListenPort != old.Spec.ListenPort {
		//update listen port
		err := aws.awsAPI.awsUpdateListener(cur.Spec.ListenerID, int64(cur.Spec.ListenPort))
		if err != nil {
			blog.Errorf("awsUpdateListener failed, %s", err.Error())
		}
		//no need to update default target group serivce port
		//because target group service port does not affect forwarding behavior
	}

	//get deleted rules and new rules
	delRules, newRules := old.GetDiffRules(cur)

	//get updated rules
	//TODO:
	updatedRules, _ := old.GetUpdateRules(cur)

	//update rules
	for _, ruleUpdate := range updatedRules {
		ruleOld, err := old.GetRuleByID(ruleUpdate.ID)
		if err != nil {
			return fmt.Errorf("GetRuleByID error, %s", err.Error())
		}
		err = aws.updateRule(cur.Spec.ListenerID, ruleOld, ruleUpdate)
		if err != nil {
			return fmt.Errorf("updateRule failed, %s", err.Error())
		}
	}

	//add new rules
	for _, ruleNew := range newRules {
		err := aws.createRule(cur.Spec.ListenerID, ruleNew)
		if err != nil {
			return fmt.Errorf("createRule failed, %s", err.Error())
		}
	}

	//delete rules
	for _, ruleDel := range delRules {
		err := aws.deleteRule(cur.Spec.ListenerID, ruleDel)
		if err != nil {
			return fmt.Errorf("deleteRule failed, %s", err.Error())
		}
	}
	return nil
}

//createRule
//when create a new rule, we will create new target group
func (aws *ElbClient) createRule(listenerID string, rule *loadbalance.Rule) error {
	//1. create target group
	//TODO:
	TargetGroupArn, err := aws.awsAPI.awsCreateTargetGroup(rule.TargetGroup.Name, "", rule.TargetGroup.Protocol,
		rule.TargetGroup.HealthCheck.HTTPCheckPath, int64(rule.TargetGroup.Port))

	//set target group id
	rule.TargetGroup.ID = TargetGroupArn

	//2. create rule
	ruleArn, err := aws.awsAPI.awsCreateRule(rule.TargetGroup.ID, listenerID, rule.Domain, rule.URL, aws.getRulePriority())
	if err != nil {
		return fmt.Errorf("awsCreateRule failed, %s", err.Error())
	}
	//set rule id
	rule.ID = ruleArn

	//3. register backends to targetgroup
	err = aws.awsAPI.awsRegisterTargets(rule.TargetGroup.ID, rule.TargetGroup.Backends)
	if err != nil {
		return fmt.Errorf("awsRegisterTargets failed, %s", err.Error())
	}
	return nil
}

//updateRule
//call createRule when rule is not existed
func (aws *ElbClient) updateRule(listenerID string, ruleOld *loadbalance.Rule, ruleUpdate *loadbalance.Rule) error {
	_, isExisted, err := aws.awsAPI.awsDescribeRule(listenerID, ruleUpdate.ID)
	if err != nil {
		return fmt.Errorf("awsDescribeRule failed, %s", err.Error())
	}
	if !isExisted {
		err := aws.createRule(listenerID, ruleUpdate)
		if err != nil {
			return fmt.Errorf("createRule failed, %s", err.Error())
		}
		return nil
	}

	//do update
	//1. update rule info
	err = aws.awsAPI.awsUpdateRule(ruleUpdate.ID, ruleUpdate.TargetGroup.ID, ruleUpdate.Domain, ruleUpdate.URL)
	if err != nil {
		return fmt.Errorf("awsUpdateRule failed, %s", err.Error())
	}
	//2. update backend in targetgroup
	backendsDel, backendsNew := ruleOld.TargetGroup.GetDiffBackend(ruleUpdate.TargetGroup)

	if len(backendsNew) != 0 {
		//2.1 register new backend
		err = aws.awsAPI.awsRegisterTargets(ruleUpdate.TargetGroup.ID, backendsNew)
		if err != nil {
			return fmt.Errorf("awsRegisterTargets failed, %s", err.Error())
		}
	}

	if len(backendsDel) != 0 {
		//2.2 deregister old backend
		err = aws.awsAPI.awsDeRegisterTargets(ruleUpdate.TargetGroup.ID, backendsDel)
		if err != nil {
			return fmt.Errorf("awsDeRegisterTargets failed, %s", err.Error())
		}
	}
	return nil
}

//deleteRule
//when rule to delete is not existed, just log error, do not return error
func (aws *ElbClient) deleteRule(listenerID string, ruleNew *loadbalance.Rule) error {
	_, isExisted, err := aws.awsAPI.awsDescribeRule(listenerID, ruleNew.ID)
	if err != nil {
		return fmt.Errorf("awsDescribeRule failed, %s", err.Error())
	}
	if !isExisted {
		blog.Errorf("deleteRule failed, %s %s is not existed, no need to delete", listenerID, ruleNew.ID)
		return nil
	}
	err = aws.awsAPI.awsDeleteRule(ruleNew.ID)
	if err != nil {
		return fmt.Errorf("awsDeleteRule failed, %s", err.Error())
	}
	return nil
}

//addSecurityGroupIngressPermissionToBackends add ip and port to security group
//if backend ip is in security group, just need to add port
//if backend ip not in security group, add both ip and port
func (aws *ElbClient) addSecurityGroupIngressPermissionToBackends(securityGroupID string, backends loadbalance.BackendList) error {
	blog.Errorf("addSecurityGroupIngressPermissionToBackends unimplemented")
	return nil
}

//addBackendsToSecurityGroup
func (aws *ElbClient) addBackendsToSecurityGroup(securityGroupID string, backends loadbalance.BackendList) error {
	blog.Errorf("addSecurityGroupIngressPermissionToBackends unimplemented")
	return nil
}
