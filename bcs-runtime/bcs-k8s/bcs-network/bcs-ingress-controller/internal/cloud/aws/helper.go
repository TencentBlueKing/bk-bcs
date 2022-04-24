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

package aws

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/aws/aws-sdk-go-v2/aws"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// do ensure network lb listener, only support one target group
// (network lb listener) --> (target group) --> backend1
//                                         |--> backend2
//                                         |--> backend3
//                                         |--> ...
func (e *Elb) ensureNetworkLBListener(region string, listener *networkextensionv1.Listener) (string, error) {
	// 1. ensure target group
	targetGroup, err := e.ensureTargetGroup(region, listener)
	if err != nil {
		return "", err
	}

	// 2. ensure listener
	listenerArn, err := e.ensureListenerSelf(region, listener, targetGroup)
	if err != nil {
		return "", err
	}
	return listenerArn, nil
}

// do create application lb listener, support multiple target groups
// (application lb listener) --> rule1  --> tartget group1 --> backend1
//                          |--> rule2 |--> tartget group2
//                          |--> rule3
//                          |--> ...
//
// domain and url is different in different rules
func (e *Elb) ensureApplicationLBListener(region string, listener *networkextensionv1.Listener) (string, error) {
	// 1. ensure target group, create a default target group for listener default action with empty backend
	targetGroup, err := e.ensureTargetGroup(region, listener)
	if err != nil {
		return "", err
	}

	// 2. ensure listener
	listenerArn, err := e.ensureListenerSelf(region, listener, targetGroup)
	if err != nil {
		return "", err
	}

	// 3. ensure rules
	if err := e.ensureRule(region, listener, listenerArn); err != nil {
		return "", err
	}
	return listenerArn, nil
}

func (e *Elb) ensureTargetGroup(region string, listener *networkextensionv1.Listener) (string, error) {
	// get lb information
	lbs, err := e.sdkWrapper.DescribeLoadBalancers(region, &elbv2.DescribeLoadBalancersInput{LoadBalancerArns: []string{listener.Spec.LoadbalancerID}})
	if err != nil {
		return "", fmt.Errorf("DescribeLoadBalancers failed, %s", err.Error())
	}
	if len(lbs.LoadBalancers) == 0 {
		return "", fmt.Errorf("loadbalancer %s not found", listener.Spec.LoadbalancerID)
	}
	lb := lbs.LoadBalancers[0]

	// 1. get target group by names
	name := e.generateTargetGroupName(*lb.LoadBalancerName, listener.Spec.Protocol, listener.Spec.Port)
	tg, err := e.sdkWrapper.DescribeTargetGroups(region, &elbv2.DescribeTargetGroupsInput{
		Names: []string{name}})
	if err != nil {
		return "", fmt.Errorf("DescribeTargetGroups failed, %s", err.Error())
	}

	// 2. if not exist or have different protocol, create target group
	var targetGroup *types.TargetGroup
	if len(tg.TargetGroups) == 0 {
		targetGroup, err = e.createTargetGroup(region, listener, name, lb.VpcId)
		if err != nil {
			return "", err
		}
	} else {
		targetGroup = &tg.TargetGroups[0]
		// if target group's protocol is different with listener, then recreate target group.
		if tg.TargetGroups[0].Protocol != types.ProtocolEnum(listener.Spec.Protocol) {
			_, err := e.sdkWrapper.DeleteTargetGroup(region, &elbv2.DeleteTargetGroupInput{
				TargetGroupArn: tg.TargetGroups[0].TargetGroupArn})
			if err != nil {
				return "", fmt.Errorf("DeleteTargetGroup failed, %s", err.Error())
			}
			targetGroup, err = e.createTargetGroup(region, listener, name, lb.VpcId)
			if err != nil {
				return "", err
			}
		}
	}

	// 3. ensure target ip and port
	if err := e.ensureTargetGroupTarget(region, listener.Spec.TargetGroup, targetGroup.TargetGroupArn); err != nil {
		return "", err
	}

	// 4. ensure target group's health check
	if err := e.ensureTargetGroupHealthCheck(region, listener, targetGroup); err != nil {
		return "", err
	}

	// 5. ensure target group attributes
	if err := e.ensureTargetGroupAttributes(region, listener, targetGroup.TargetGroupArn); err != nil {
		return "", err
	}
	return *targetGroup.TargetGroupArn, nil
}

func (e *Elb) generateTargetGroupName(lbName, protocol string, port int) string {
	return fmt.Sprintf("%s-%s-%d", lbName, protocol, port)
}

func (e *Elb) ensureListenerSelf(region string, listener *networkextensionv1.Listener,
	defaultTargetGroup string) (string, error) {
	// get cloud listener
	input := &elbv2.DescribeListenersInput{LoadBalancerArn: &listener.Spec.LoadbalancerID}
	listeners, err := e.sdkWrapper.DescribeListeners(region, input)
	if err != nil {
		return "", fmt.Errorf("DescribeListeners failed, err %s", err.Error())
	}

	// find listener
	found := e.findListenerByPort(listeners.Listeners, listener.Spec.Port)
	listenerInput := e.generateListenerInput(listener, defaultTargetGroup)

	// if not found, create listener
	if found == nil {
		output, err := e.sdkWrapper.CreateListener(region, listenerInput)
		if err != nil {
			return "", fmt.Errorf("CreateListener failed, err %s", err.Error())
		}
		if len(output.Listeners) == 0 {
			return "", fmt.Errorf("CreateLIstener failed, response listeners is nil")
		}
		return *output.Listeners[0].ListenerArn, nil
	}

	// if found, update listener
	_, err = e.sdkWrapper.ModifyListener(region, &elbv2.ModifyListenerInput{
		ListenerArn:    found.ListenerArn,
		Certificates:   listenerInput.Certificates,
		DefaultActions: listenerInput.DefaultActions,
		Port:           found.Port,
		Protocol:       listenerInput.Protocol,
	})
	if err != nil {
		return "", fmt.Errorf("ModifyListener failed, err %s", err.Error())
	}
	return *found.ListenerArn, nil
}

// check exist rules and except rules
// if except rules is not int exist rules, the add it
// if exist rules is not in except rules, then delete it
// if exist rules is in except rules, then update it
func (e *Elb) ensureRule(region string, listener *networkextensionv1.Listener, listenerArn string) error {
	ruleTargetGroup, err := e.ensureRuleTargetGroup(region, listener)
	if err != nil {
		return err
	}
	exceptRules := e.generateExceptRules(listener.Spec.Rules, ruleTargetGroup)
	rules, err := e.sdkWrapper.DescribeRules(region, &elbv2.DescribeRulesInput{ListenerArn: &listenerArn})
	if err != nil {
		return fmt.Errorf("DescribeRules failed, err %s", err.Error())
	}

	var add []elbv2.CreateRuleInput
	var modify []elbv2.ModifyRuleInput
	var del []types.Rule
	for _, r := range rules.Rules {
		if r.IsDefault {
			continue
		}
		var found *types.Rule
		for _, v := range exceptRules {
			if isSameRuleCondition(r.Conditions, v.Conditions) {
				found = &v
			}
		}
		if found == nil {
			del = append(del, types.Rule{
				RuleArn:    r.RuleArn,
				Actions:    r.Actions,
				Conditions: r.Conditions,
			})
		} else {
			modify = append(modify, elbv2.ModifyRuleInput{
				RuleArn:    r.RuleArn,
				Conditions: r.Conditions,
				Actions:    found.Actions,
			})
		}
	}

	for _, r := range exceptRules {
		found := false
		for _, v := range rules.Rules {
			if v.IsDefault {
				continue
			}
			if isSameRuleCondition(r.Conditions, v.Conditions) {
				found = true
			}
		}
		if !found {
			add = append(add, elbv2.CreateRuleInput{
				Conditions:  r.Conditions,
				Actions:     r.Actions,
				Priority:    e.nextPriority(rules.Rules, add),
				ListenerArn: &listenerArn,
			})
		}
	}

	for _, v := range add {
		if _, err := e.sdkWrapper.CreateRule(region, &v); err != nil {
			return fmt.Errorf("CreateRule failed, err %s", err.Error())
		}
	}
	for _, v := range modify {
		if _, err := e.sdkWrapper.ModifyRule(region, &v); err != nil {
			return fmt.Errorf("ModifyRule failed, err %s", err.Error())
		}
	}
	for _, v := range del {
		if _, err := e.sdkWrapper.DeleteRule(region, &elbv2.DeleteRuleInput{RuleArn: v.RuleArn}); err != nil {
			return fmt.Errorf("DeleteRule failed, err %s", err.Error())
		}
		for _, tg := range v.Actions {
			if _, err := e.sdkWrapper.DeleteTargetGroup(region, &elbv2.DeleteTargetGroupInput{
				TargetGroupArn: tg.TargetGroupArn,
			}); err != nil {
				return fmt.Errorf("DeleteTargetGroup failed, err %s", err.Error())
			}
		}
	}
	return nil
}

// ensure all rules's backend are created, every backend is created by one target group
func (e *Elb) ensureRuleTargetGroup(region string, listener *networkextensionv1.Listener) (map[string]string, error) {
	// get lb information
	lbs, err := e.sdkWrapper.DescribeLoadBalancers(region, &elbv2.DescribeLoadBalancersInput{LoadBalancerArns: []string{listener.Spec.LoadbalancerID}})
	if err != nil {
		return nil, fmt.Errorf("DescribeLoadBalancers failed, %s", err.Error())
	}
	if len(lbs.LoadBalancers) == 0 {
		return nil, fmt.Errorf("loadbalancer %s not found", listener.Spec.LoadbalancerID)
	}
	lb := lbs.LoadBalancers[0]
	ruleTgMap := make(map[string]string, 0)
	for _, rule := range listener.Spec.Rules {
		if rule.TargetGroup == nil {
			continue
		}
		tgName := getRuleTargetGroupName(rule.Domain, rule.Path)
		tg, err := e.sdkWrapper.DescribeTargetGroups(region, &elbv2.DescribeTargetGroupsInput{
			Names: []string{tgName}})
		if err != nil {
			return nil, fmt.Errorf("DescribeTargetGroups failed, err %s", err.Error())
		}

		// ensure target group
		tgArn := ""
		if tg.TargetGroups == nil {
			// create target group
			input := &elbv2.CreateTargetGroupInput{
				Name:                &tgName,
				TargetType:          types.TargetTypeEnumIp,
				VpcId:               lb.VpcId,
				Port:                aws.Int32(int32(listener.Spec.Port)),
				Protocol:            types.ProtocolEnum(listener.Spec.Protocol),
				HealthCheckProtocol: types.ProtocolEnum(listener.Spec.Protocol),
			}
			if rule.ListenerAttribute != nil && rule.ListenerAttribute.BackendInsecure {
				input.Protocol = types.ProtocolEnumHttp
			}
			setHealthCheck(input, &rule)
			out, err := e.sdkWrapper.CreateTargetGroup(region, input)
			if err != nil {
				return nil, fmt.Errorf("CreateTargetGroup failed, err %s", err.Error())
			}
			tgArn = *out.TargetGroups[0].TargetGroupArn
		} else {
			// update target group
			input := &elbv2.ModifyTargetGroupInput{
				TargetGroupArn: tg.TargetGroups[0].TargetGroupArn,
			}
			setModifyHealthCheck(input, &rule)
			_, err := e.sdkWrapper.ModifyTargetGroup(region, input)
			if err != nil {
				return nil, fmt.Errorf("ModifyTargetGroup failed, err %s", err.Error())
			}
			tgArn = *tg.TargetGroups[0].TargetGroupArn
		}

		// ensure target group's target
		if err := e.ensureTargetGroupTarget(region, rule.TargetGroup, &tgArn); err != nil {
			return nil, err
		}

		// ensure target group's attribute
		if err := e.ensureRuleTargetGroupAttributes(region, rule, &tgArn); err != nil {
			return nil, err
		}

		ruleTgMap[tgName] = tgArn
	}

	return ruleTgMap, nil
}

func setHealthCheck(input *elbv2.CreateTargetGroupInput, rule *networkextensionv1.ListenerRule) {
	if rule.ListenerAttribute == nil || rule.ListenerAttribute.HealthCheck == nil {
		return
	}
	hc := rule.ListenerAttribute.HealthCheck
	if hc.HealthNum != 0 {
		input.HealthyThresholdCount = aws.Int32(int32(hc.HealthNum))
	}
	if hc.HTTPCheckPath != "" {
		input.HealthCheckPath = aws.String(hc.HTTPCheckPath)
	}
	if hc.UnHealthNum != 0 {
		input.UnhealthyThresholdCount = aws.Int32(int32(hc.UnHealthNum))
	}
	if hc.Timeout != 0 {
		input.HealthCheckTimeoutSeconds = aws.Int32(int32(hc.Timeout))
	}
	if hc.IntervalTime != 0 {
		input.HealthCheckIntervalSeconds = aws.Int32(int32(hc.IntervalTime))
	}
	if len(hc.HTTPCodeValues) != 0 {
		input.Matcher = &types.Matcher{HttpCode: aws.String(hc.HTTPCodeValues)}
	}
	if len(hc.HealthCheckProtocol) != 0 {
		input.HealthCheckProtocol = types.ProtocolEnum(hc.HealthCheckProtocol)
	}
	if hc.HealthCheckPort != 0 {
		input.HealthCheckPort = aws.String(strconv.Itoa(hc.HealthCheckPort))
	}
}

func setModifyHealthCheck(input *elbv2.ModifyTargetGroupInput, rule *networkextensionv1.ListenerRule) {
	if rule.ListenerAttribute == nil || rule.ListenerAttribute.HealthCheck == nil {
		return
	}
	hc := rule.ListenerAttribute.HealthCheck
	if hc.HealthNum != 0 {
		input.HealthyThresholdCount = aws.Int32(int32(hc.HealthNum))
	}
	if hc.HTTPCheckPath != "" {
		input.HealthCheckPath = aws.String(hc.HTTPCheckPath)
	}
	if hc.UnHealthNum != 0 {
		input.UnhealthyThresholdCount = aws.Int32(int32(hc.UnHealthNum))
	}
	if hc.Timeout != 0 {
		input.HealthCheckTimeoutSeconds = aws.Int32(int32(hc.Timeout))
	}
	if hc.IntervalTime != 0 {
		input.HealthCheckIntervalSeconds = aws.Int32(int32(hc.IntervalTime))
	}
	if len(hc.HTTPCodeValues) != 0 {
		input.Matcher = &types.Matcher{HttpCode: aws.String(hc.HTTPCodeValues)}
	}
	if len(hc.HealthCheckProtocol) != 0 {
		input.HealthCheckProtocol = types.ProtocolEnum(hc.HealthCheckProtocol)
	}
	if hc.HealthCheckPort != 0 {
		input.HealthCheckPort = aws.String(strconv.Itoa(hc.HealthCheckPort))
	}
}

// md5(domain+path)
func getRuleTargetGroupName(domain, path string) string {
	return fmt.Sprintf("%x", (md5.Sum([]byte(domain + path))))
}

// generate rules from listener rules
func (e *Elb) generateExceptRules(rules []networkextensionv1.ListenerRule,
	ruleTgMap map[string]string) []types.Rule {
	var except []types.Rule
	for _, rule := range rules {
		if rule.TargetGroup == nil {
			continue
		}
		tgName := getRuleTargetGroupName(rule.Domain, rule.Path)
		action := types.Action{
			Type: types.ActionTypeEnumForward,
			ForwardConfig: &types.ForwardActionConfig{
				TargetGroups: []types.TargetGroupTuple{
					{TargetGroupArn: aws.String(ruleTgMap[tgName])},
				},
			},
		}
		conditions := []types.RuleCondition{
			{Field: aws.String("host-header"), HostHeaderConfig: &types.HostHeaderConditionConfig{
				Values: []string{rule.Domain},
			}},
		}
		if len(rule.Path) != 0 {
			conditions = append(conditions, types.RuleCondition{Field: aws.String("path-pattern"),
				PathPatternConfig: &types.PathPatternConditionConfig{
					Values: []string{rule.Path},
				}},
			)
		}
		except = append(except, types.Rule{
			Actions:    []types.Action{action},
			Conditions: conditions,
		})
	}
	return except
}

func (e *Elb) nextPriority(rules []types.Rule, add []elbv2.CreateRuleInput) *int32 {
	highest := int32(0)
	for _, r := range rules {
		i, err := strconv.Atoi(*r.Priority)
		if err != nil {
			continue
		}
		if int32(i) > highest {
			highest = int32(i)
		}
	}
	next := highest + int32(len(add)) + int32(1)
	return &next
}

func (e *Elb) generateListenerInput(listener *networkextensionv1.Listener,
	defaultTargetGroup string) *elbv2.CreateListenerInput {
	listenerInput := &elbv2.CreateListenerInput{
		DefaultActions: []types.Action{{
			Type:           types.ActionTypeEnumForward,
			TargetGroupArn: &defaultTargetGroup,
		}},
		LoadBalancerArn: &listener.Spec.LoadbalancerID,
		Protocol:        types.ProtocolEnum(listener.Spec.Protocol),
		Port:            aws.Int32(int32(listener.Spec.Port)),
	}
	if listener.Spec.Protocol == ElbProtocolHTTPS && listener.Spec.Certificate != nil {
		listenerInput.Certificates = append(listenerInput.Certificates, types.Certificate{
			CertificateArn: &listener.Spec.Certificate.CertID,
		})
	}
	return listenerInput
}

func (e *Elb) findListenerByPort(listeners []types.Listener, port int) *types.Listener {
	var foundListener *types.Listener
	for _, li := range listeners {
		if li.Port == nil {
			continue
		}
		if int(*li.Port) == port {
			return &li
		}
	}
	return foundListener
}

func (e *Elb) createTargetGroup(region string, listener *networkextensionv1.Listener,
	tgName string, vpcID *string) (
	*types.TargetGroup, error) {
	newTgInput := &elbv2.CreateTargetGroupInput{
		Name:                &tgName,
		Port:                aws.Int32(int32(listener.Spec.Port)),
		TargetType:          types.TargetTypeEnumIp,
		VpcId:               vpcID,
		Protocol:            types.ProtocolEnum(listener.Spec.Protocol),
		HealthCheckProtocol: types.ProtocolEnum(listener.Spec.Protocol),
	}
	if listener.Spec.Protocol == ElbProtocolHTTP || listener.Spec.Protocol == ElbProtocolHTTPS {
		if listener.Spec.ListenerAttribute != nil &&
			listener.Spec.ListenerAttribute.HealthCheck != nil {
			hc := listener.Spec.ListenerAttribute.HealthCheck
			if hc.HealthNum != 0 {
				newTgInput.HealthyThresholdCount = aws.Int32(int32(hc.HealthNum))
			}
			if hc.HTTPCheckPath != "" {
				newTgInput.HealthCheckPath = aws.String(hc.HTTPCheckPath)
			}
			if hc.UnHealthNum != 0 {
				newTgInput.UnhealthyThresholdCount = aws.Int32(int32(hc.UnHealthNum))
			}
			if hc.Timeout != 0 {
				newTgInput.HealthCheckTimeoutSeconds = aws.Int32(int32(hc.Timeout))
			}
			if hc.IntervalTime != 0 {
				newTgInput.HealthCheckIntervalSeconds = aws.Int32(int32(hc.IntervalTime))
			}
			if len(hc.HTTPCodeValues) != 0 {
				newTgInput.Matcher = &types.Matcher{HttpCode: aws.String(hc.HTTPCodeValues)}
			}
			if len(hc.HealthCheckProtocol) != 0 {
				newTgInput.HealthCheckProtocol = types.ProtocolEnum(hc.HealthCheckProtocol)
			}
			if hc.HealthCheckPort != 0 {
				newTgInput.HealthCheckPort = aws.String(strconv.Itoa(hc.HealthCheckPort))
			}
		}
	}
	if listener.Spec.Protocol == ElbProtocolTCP || listener.Spec.Protocol == ElbProtocolUDP {
		newTgInput.HealthCheckProtocol = types.ProtocolEnumTcp
		if listener.Spec.ListenerAttribute != nil &&
			listener.Spec.ListenerAttribute.HealthCheck != nil {
			hc := listener.Spec.ListenerAttribute.HealthCheck
			if hc.HealthNum != 0 {
				newTgInput.HealthyThresholdCount = aws.Int32(int32(hc.HealthNum))
				newTgInput.UnhealthyThresholdCount = aws.Int32(int32(hc.HealthNum))
			}
		}
	}
	tg, err := e.sdkWrapper.CreateTargetGroup(region, newTgInput)
	if err != nil {
		return nil, fmt.Errorf("CreateTargetGroup failed, %s", err.Error())
	}
	if len(tg.TargetGroups) == 0 {
		return nil, fmt.Errorf("CreateTargetGroup failed, empty output")
	}
	return &tg.TargetGroups[0], nil
}

func (e *Elb) ensureTargetGroupTarget(region string, listenerTg *networkextensionv1.ListenerTargetGroup,
	targetGroupArn *string) error {
	th, err := e.sdkWrapper.DescribeTargetHealth(region, &elbv2.DescribeTargetHealthInput{
		TargetGroupArn: targetGroupArn})
	if err != nil {
		return fmt.Errorf("DescribeTargetHealth failed, %s", err.Error())
	}

	var registers, deregisters []types.TargetDescription

	var backends []networkextensionv1.ListenerBackend
	if listenerTg != nil {
		backends = listenerTg.Backends
	}
	for _, backend := range backends {
		var found bool
		for _, t := range th.TargetHealthDescriptions {
			if t.Target != nil && *t.Target.Id == backend.IP && *t.Target.Port == int32(backend.Port) {
				found = true
				break
			}
		}
		if !found {
			registers = append(registers, types.TargetDescription{
				Id:   aws.String(backend.IP),
				Port: aws.Int32(int32(backend.Port)),
			})
		}
	}
	for _, t := range th.TargetHealthDescriptions {
		if t.Target == nil {
			continue
		}
		var found bool
		for _, backend := range backends {
			if *t.Target.Id == backend.IP && *t.Target.Port == int32(backend.Port) {
				found = true
				break
			}
		}
		if !found {
			deregisters = append(deregisters, types.TargetDescription{
				Id:   t.Target.Id,
				Port: t.Target.Port,
			})
		}
	}

	if len(registers) > 0 {
		if _, err := e.sdkWrapper.RegisterTargets(region, &elbv2.RegisterTargetsInput{
			TargetGroupArn: targetGroupArn, Targets: registers}); err != nil {
			return fmt.Errorf("RegisterTargets failed, %s", err.Error())
		}
	}
	if len(deregisters) > 0 {
		if _, err := e.sdkWrapper.DeregisterTargets(region, &elbv2.DeregisterTargetsInput{
			TargetGroupArn: targetGroupArn, Targets: deregisters}); err != nil {
			return fmt.Errorf("DeregisterTargets failed, %s", err.Error())
		}
	}
	return nil
}

func (e *Elb) ensureTargetGroupHealthCheck(region string, listener *networkextensionv1.Listener,
	targetGroup *types.TargetGroup) error {
	input := &elbv2.ModifyTargetGroupInput{TargetGroupArn: targetGroup.TargetGroupArn}
	if listener.Spec.TargetGroup != nil && len(listener.Spec.TargetGroup.Backends) > 0 {
		input.HealthCheckPort = aws.String(strconv.Itoa(listener.Spec.TargetGroup.Backends[0].Port))
	}
	if listener.Spec.Protocol == ElbProtocolHTTP || listener.Spec.Protocol == ElbProtocolHTTPS {
		if listener.Spec.ListenerAttribute != nil &&
			listener.Spec.ListenerAttribute.HealthCheck != nil {
			hc := listener.Spec.ListenerAttribute.HealthCheck
			if hc.HealthNum != 0 {
				input.HealthyThresholdCount = aws.Int32(int32(hc.HealthNum))
			}
			if hc.HTTPCheckPath != "" {
				input.HealthCheckPath = aws.String(hc.HTTPCheckPath)
			}
			if hc.UnHealthNum != 0 {
				input.UnhealthyThresholdCount = aws.Int32(int32(hc.UnHealthNum))
			}
			if hc.Timeout != 0 {
				input.HealthCheckTimeoutSeconds = aws.Int32(int32(hc.Timeout))
			}
			if hc.IntervalTime != 0 {
				input.HealthCheckIntervalSeconds = aws.Int32(int32(hc.IntervalTime))
			}
			if len(hc.HTTPCodeValues) != 0 {
				input.Matcher = &types.Matcher{HttpCode: aws.String(hc.HTTPCodeValues)}
			}
			if len(hc.HealthCheckProtocol) != 0 {
				input.HealthCheckProtocol = types.ProtocolEnum(hc.HealthCheckProtocol)
			}
			if hc.HealthCheckPort != 0 {
				input.HealthCheckPort = aws.String(strconv.Itoa(hc.HealthCheckPort))
			}
		}
	}
	if listener.Spec.Protocol == ElbProtocolTCP || listener.Spec.Protocol == ElbProtocolUDP {
		input.HealthCheckProtocol = types.ProtocolEnumTcp
		if listener.Spec.ListenerAttribute != nil &&
			listener.Spec.ListenerAttribute.HealthCheck != nil {
			hc := listener.Spec.ListenerAttribute.HealthCheck
			if hc.HealthNum != 0 {
				input.HealthyThresholdCount = aws.Int32(int32(hc.HealthNum))
				input.UnhealthyThresholdCount = aws.Int32(int32(hc.HealthNum))
			}
			if hc.HealthCheckPort != 0 {
				input.HealthCheckPort = aws.String(strconv.Itoa(hc.HealthCheckPort))
			}
		}
	}

	_, err := e.sdkWrapper.ModifyTargetGroup(region, input)
	if err != nil {
		return fmt.Errorf("ModifyTargetGroup failed, %s", err.Error())
	}
	return nil
}

func (e *Elb) ensureTargetGroupAttributes(region string, listener *networkextensionv1.Listener,
	targetGroupArn *string) error {
	attrs := make([]types.TargetGroupAttribute, 0)
	if listener.Spec.ListenerAttribute == nil {
		return nil
	}
	for _, v := range listener.Spec.ListenerAttribute.AWSAttributes {
		attrs = append(attrs, types.TargetGroupAttribute{
			Key: aws.String(v.Key), Value: aws.String(v.Value),
		})
	}
	if len(attrs) == 0 {
		return nil
	}
	_, err := e.sdkWrapper.ModifyTargetGroupAttributes(region, &elbv2.ModifyTargetGroupAttributesInput{
		TargetGroupArn: targetGroupArn,
		Attributes:     attrs,
	})
	return err
}

func (e *Elb) ensureRuleTargetGroupAttributes(region string, rule networkextensionv1.ListenerRule,
	targetGroupArn *string) error {
	attrs := make([]types.TargetGroupAttribute, 0)
	if rule.ListenerAttribute == nil {
		return nil
	}
	for _, v := range rule.ListenerAttribute.AWSAttributes {
		attrs = append(attrs, types.TargetGroupAttribute{
			Key: aws.String(v.Key), Value: aws.String(v.Value),
		})
	}
	if len(attrs) == 0 {
		return nil
	}
	_, err := e.sdkWrapper.ModifyTargetGroupAttributes(region, &elbv2.ModifyTargetGroupAttributesInput{
		TargetGroupArn: targetGroupArn,
		Attributes:     attrs,
	})
	return err
}

func (e *Elb) getAllListenerRulesAndTargetGroups(region, listenerArn string) ([]string, map[string]bool, error) {
	ruleInput := &elbv2.DescribeRulesInput{ListenerArn: &listenerArn}
	rules, err := e.sdkWrapper.DescribeRules(region, ruleInput)
	if err != nil {
		return nil, nil, fmt.Errorf("DescribeRules failed, err %s", err.Error())
	}

	targetGroups := make(map[string]bool, 0)
	ruleIDs := make([]string, 0)
	for _, rule := range rules.Rules {
		ruleIDs = append(ruleIDs, *rule.RuleArn)
		for _, action := range rule.Actions {
			if action.ForwardConfig != nil {
				for _, v := range action.ForwardConfig.TargetGroups {
					targetGroups[*v.TargetGroupArn] = true
				}
			}
		}
	}
	return ruleIDs, targetGroups, nil
}

func convertHealthStatus(status types.TargetHealthStateEnum) string {
	var statusStr string
	switch status {
	case types.TargetHealthStateEnumHealthy:
		statusStr = cloud.BackendHealthStatusHealthy
	case types.TargetHealthStateEnumUnhealthy:
		statusStr = cloud.BackendHealthStatusUnhealthy
	default:
		statusStr = cloud.BackendHealthStatusUnknown
	}
	return statusStr
}

// has same rule condition
// condition has host and path, check if the rule has same host and path
func isSameRuleCondition(a, b []types.RuleCondition) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		same := false
		for _, v1 := range b {
			if *v.Field == *v1.Field {
				if reflect.DeepEqual(v.HostHeaderConfig, v1.HostHeaderConfig) ||
					reflect.DeepEqual(v.PathPatternConfig, v1.PathPatternConfig) {
					same = true
				}
			}
		}
		if !same {
			return false
		}
	}
	return true
}
