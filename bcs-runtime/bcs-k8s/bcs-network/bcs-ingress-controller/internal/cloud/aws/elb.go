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
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	k8scorev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Elb client to operate Elb instance
type Elb struct {
	sdkWrapper *SdkWrapper
}

// NewElb create elb client
func NewElb() (*Elb, error) {
	sdkWrapper, err := NewSdkWrapper()
	if err != nil {
		return nil, err
	}
	return &Elb{
		sdkWrapper: sdkWrapper,
	}, nil
}

// NewElbWithSecret create elb client with k8s secret
func NewElbWithSecret(secret *k8scorev1.Secret, _ client.Client) (cloud.LoadBalance, error) {
	secretIDBytes, ok := secret.Data[EnvNameAWSAccessKeyID]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret %s/%s", EnvNameAWSAccessKeyID,
			secret.Namespace, secret.Name)
	}
	secretKeyBytes, ok := secret.Data[EnvNameAWSAccessKey]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret %s/%s", EnvNameAWSAccessKey,
			secret.Namespace, secret.Name)
	}
	sdkWrapper, err := NewSdkWrapperWithSecretIDKey(string(secretIDBytes), string(secretKeyBytes))
	if err != nil {
		return nil, err
	}
	return &Elb{
		sdkWrapper: sdkWrapper,
	}, nil
}

var _ cloud.LoadBalance = &Elb{}

// DescribeLoadBalancer get loadbalancer object by id or name
func (e *Elb) DescribeLoadBalancer(region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	input := &elbv2.DescribeLoadBalancersInput{}
	if len(lbID) != 0 {
		input.LoadBalancerArns = []string{lbID}
	}
	if len(name) != 0 {
		input.Names = []string{name}
	}

	output, err := e.sdkWrapper.DescribeLoadBalancers(region, input)
	if err != nil {
		blog.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
		return nil, fmt.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
	}

	if len(output.LoadBalancers) == 0 {
		return nil, cloud.ErrLoadbalancerNotFound
	}
	var resplb *types.LoadBalancer
	for _, lb := range output.LoadBalancers {
		if len(lbID) != 0 && lbID == *lb.LoadBalancerArn {
			resplb = &lb
			break
		}
		if len(name) != 0 && name == *lb.LoadBalancerName {
			resplb = &lb
			break
		}
	}
	if resplb == nil {
		blog.Errorf("lb not found in resp %s", common.ToJsonString(output))
		return nil, cloud.ErrLoadbalancerNotFound
	}
	retlb := &cloud.LoadBalanceObject{
		Region:    region,
		Scheme:    string(resplb.Scheme),
		AWSLBType: string(resplb.Type),
	}
	if resplb.LoadBalancerArn != nil {
		retlb.LbID = *resplb.LoadBalancerArn
	}
	if resplb.LoadBalancerName != nil {
		retlb.Name = *resplb.LoadBalancerName
	}
	if resplb.DNSName != nil {
		retlb.DNSName = *resplb.DNSName
	}
	return retlb, nil
}

// DescribeLoadBalancerWithNs get loadbalancer object by id or name with namespace specified
func (e *Elb) DescribeLoadBalancerWithNs(ns, region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	return e.DescribeLoadBalancer(region, lbID, name)
}

// IsNamespaced if client is namespaced
func (e *Elb) IsNamespaced() bool {
	return false
}

// EnsureListener ensure listener to cloud
func (e *Elb) EnsureListener(region string, listener *networkextensionv1.Listener) (string, error) {
	if listener.Spec.LoadbalancerID == "" {
		return "", fmt.Errorf("loadbalancer id is empty")
	}

	switch listener.Spec.Protocol {
	case ElbProtocolHTTP, ElbProtocolHTTPS:
		return e.ensureApplicationLBListener(region, listener)
	case ElbProtocolTCP, ElbProtocolUDP:
		return e.ensureNetworkLBListener(region, listener)
	default:
		blog.Errorf("invalid protocol %s", listener.Spec.Protocol)
		return "", fmt.Errorf("invalid protocol %s", listener.Spec.Protocol)
	}
}

// DeleteListener delete listener by name
func (e *Elb) DeleteListener(region string, listener *networkextensionv1.Listener) error {
	if listener.Spec.EndPort != 0 {
		return e.DeleteSegmentListener(region, listener)
	}
	// 1. get listener id
	input := &elbv2.DescribeListenersInput{LoadBalancerArn: &listener.Spec.LoadbalancerID}
	listeners, err := e.sdkWrapper.DescribeListeners(region, input)
	if err != nil {
		return fmt.Errorf("DescribeListeners failed, err %s", err.Error())
	}
	found := e.findListenerByPort(listeners.Listeners, listener.Spec.Port)
	if found == nil {
		blog.Warnf("listener %s not found", listener.Spec.Port)
		return nil
	}

	// 2. get listener's rules and all target groups
	rules, tgs, err := e.getAllListenerRulesAndTargetGroups(region, *found.ListenerArn)
	if err != nil {
		return err
	}

	// 3. delete listener
	_, err = e.sdkWrapper.DeleteListener(region, &elbv2.DeleteListenerInput{
		ListenerArn: found.ListenerArn})
	if err != nil {
		return fmt.Errorf("DeleteListener failed, err %s", err.Error())
	}

	// 4. delete all listeners' rules
	for _, rule := range rules {
		out, err := e.sdkWrapper.DescribeRules(region, &elbv2.DescribeRulesInput{RuleArns: []string{rule}})
		if err != nil {
			blog.Warnf("DescribeRules failed, err %s", err.Error())
			continue
		}
		// default rule cannot be deleted
		if out.Rules != nil && out.Rules[0].IsDefault {
			continue
		}
		_, err = e.sdkWrapper.DeleteRule(region, &elbv2.DeleteRuleInput{RuleArn: &rule})
		if err != nil {
			return fmt.Errorf("DeleteRule failed, err %s", err.Error())
		}
	}

	// 5. delete all listeners' target groups
	for tg := range tgs {
		_, err = e.sdkWrapper.DescribeTargetGroups(region, &elbv2.DescribeTargetGroupsInput{
			TargetGroupArns: []string{tg}})
		if err != nil {
			blog.Warnf("DescribeTargetGroups failed, err %s", err.Error())
			continue
		}
		_, err := e.sdkWrapper.DeleteTargetGroup(region, &elbv2.DeleteTargetGroupInput{
			TargetGroupArn: &tg})
		if err != nil {
			return fmt.Errorf("DeleteTargetGroup failed, err %s", err.Error())
		}
	}
	return nil
}

// EnsureMultiListeners ensure multiple listeners to cloud
func (e *Elb) EnsureMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) (map[string]string, error) {
	retMap := make(map[string]string)
	for _, listener := range listeners {
		liID, err := e.EnsureListener(region, listener)
		if err != nil {
			return nil, err
		}
		retMap[listener.Name] = liID
	}
	return retMap, nil
}

// DeleteMultiListeners delete multiple listeners from cloud
func (e *Elb) DeleteMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) error {
	for _, listener := range listeners {
		err := e.DeleteListener(region, listener)
		if err != nil {
			return err
		}
	}
	return nil
}

// EnsureSegmentListener ensure listener with port segment
// 端口段：以端口段为规则配置，一个vip的一段端口（首端口-尾端口）绑定一个RS的一段端口。
// 如将vip vport(8000, 8001, 8002……9000) 绑定到 rsip rsport(9000, 9001, 9002……10000)，vport和rsport一一对应
// vport 8000 转发到 rsport 9000
// vport 8001 转发到 rsport 9001
func (e *Elb) EnsureSegmentListener(region string, listener *networkextensionv1.Listener) (string, error) {
	if listener.Spec.EndPort == 0 {
		return e.EnsureListener(region, listener)
	}
	// create listener for each port
	portIndex := 0
	listenerIds := make([]string, 0)
	for i := listener.Spec.Port; i <= listener.Spec.EndPort; i++ {
		// generate single port listener to ensure listener
		li := listener.DeepCopy()
		li.Spec.Port = i
		li.Spec.EndPort = 0
		if li.Spec.TargetGroup != nil {
			for j := range li.Spec.TargetGroup.Backends {
				li.Spec.TargetGroup.Backends[j].Port += portIndex
			}
		}
		portIndex++
		liID, err := e.EnsureListener(region, li)
		if err != nil {
			return "", err
		}
		listenerIds = append(listenerIds, liID)
	}
	return strings.Join(listenerIds, ","), nil
}

// EnsureMultiSegmentListeners ensure multi segment listeners
func (e *Elb) EnsureMultiSegmentListeners(region, lbID string, listeners []*networkextensionv1.Listener) (map[string]string, error) {
	retMap := make(map[string]string)
	for _, listener := range listeners {
		liID, err := e.EnsureSegmentListener(region, listener)
		if err != nil {
			return nil, err
		}
		retMap[listener.Name] = liID
	}
	return retMap, nil
}

// DeleteSegmentListener delete segment listener
func (e *Elb) DeleteSegmentListener(region string, listener *networkextensionv1.Listener) error {
	if listener.Spec.EndPort == 0 {
		return e.DeleteListener(region, listener)
	}
	// delete listener for each port
	portIndex := 0
	for i := listener.Spec.Port; i <= listener.Spec.EndPort; i++ {
		// generate single port listener to ensure listener
		li := listener.DeepCopy()
		li.Spec.Port = i
		li.Spec.EndPort = 0
		if li.Spec.TargetGroup != nil {
			for _, target := range li.Spec.TargetGroup.Backends {
				target.Port += portIndex
			}
		}
		portIndex++
		err := e.DeleteListener(region, li)
		if err != nil {
			blog.Warnf("DeleteListener %s(%s) failed, err %s", li.Spec.LoadbalancerID,
				li.Spec.Port, err.Error())
		}
	}
	return nil
}

// DescribeBackendStatus describe elb backend status, the input ns is no use here, only effects in namespaced cloud client
func (e *Elb) DescribeBackendStatus(region, ns string, lbIDs []string) (map[string][]*cloud.BackendHealthStatus, error) {
	retMap := make(map[string][]*cloud.BackendHealthStatus)
	// 1. get all listeners
	for _, lbID := range lbIDs {
		listeners, err := e.sdkWrapper.DescribeListeners(region, &elbv2.DescribeListenersInput{
			LoadBalancerArn: &lbID,
		})
		if err != nil {
			return nil, fmt.Errorf("DescribeListeners failed, err %s", err.Error())
		}
		// 2. get all listener's target groups
		for _, listener := range listeners.Listeners {
			_, targetGroups, err := e.getAllListenerRulesAndTargetGroups(region, *listener.ListenerArn)
			if err != nil {
				return nil, err
			}
			// 3. get all target groups' health status
			for tg := range targetGroups {
				ths, err := e.sdkWrapper.DescribeTargetHealth(region, &elbv2.DescribeTargetHealthInput{
					TargetGroupArn: &tg,
				})
				if err != nil {
					return nil, fmt.Errorf("DescribeTargetHealth failed, err %s", err.Error())
				}
				for _, th := range ths.TargetHealthDescriptions {
					tmpStatus := &cloud.BackendHealthStatus{
						ListenerID:   *listener.ListenerArn,
						ListenerPort: int(*listener.Port),
						Namespace:    ns,
						IP:           *th.Target.Id,
						Port:         int(*th.Target.Port),
						Protocol:     string(listener.Protocol),
						Status:       convertHealthStatus(th.TargetHealth.State),
					}
					retMap[lbID] = append(retMap[lbID], tmpStatus)
				}
			}
		}
	}
	return retMap, nil
}
