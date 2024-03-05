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

// Package np used to re-build networkPolicyInfos
package np

import (
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/controller"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/datainformer"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/iptables"
	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NetworkPolicyHandler saved the informer, and list all networkPolicies.
// Re-build them to net struct.
type NetworkPolicyHandler struct {
	informer datainformer.Interface
}

// NewNetworkPolicyHandler return networkPolicyHandler object.
func NewNetworkPolicyHandler(informer datainformer.Interface) NetworkPolicyHandler {
	return NetworkPolicyHandler{
		informer: informer,
	}
}

// Build used to rebuild and merge all the networkPolicies
func (h NetworkPolicyHandler) Build() ([]controller.NetworkPolicyInfo, error) {
	policies, err := h.informer.ListAllNetworkPolicy()
	if err != nil {
		blog.Errorf("Failed to list networkPolicies from informer, err: %s", err.Error())
		return nil, err
	}

	newPolicies := make([]controller.NetworkPolicyInfo, 0, len(policies))
	for _, policy := range policies {
		newPolicy := controller.NetworkPolicyInfo{
			Name:       policy.Name,
			Namespace:  policy.Namespace,
			Labels:     policy.Spec.PodSelector.MatchLabels,
			PolicyType: evalPolicyType(policy),
		}

		// list all the target pods.
		targetPods, err := h.buildTargetPods(policy, &newPolicy)
		if err != nil {
			return nil, err
		}

		// grab the named ports of all pods
		podsNamedPorts := h.buildPodsPorts(targetPods)

		// calc ingress rules
		if err := h.buildIngressRules(policy, &newPolicy, podsNamedPorts); err != nil {
			return nil, err
		}

		// calc egress rules
		if err := h.buildEgressRules(policy, &newPolicy, podsNamedPorts); err != nil {
			return nil, err
		}

		newPolicies = append(newPolicies, newPolicy)
	}
	return newPolicies, nil
}

// buildIngressRules will calc the source pods' ips of ingress rules.
func (h NetworkPolicyHandler) buildIngressRules(policy *networking.NetworkPolicy,
	newPolicy *controller.NetworkPolicyInfo,
	podsNamedPorts controller.NamedPort2eps) error {
	newPolicy.IngressRules = make([]controller.IngressRule, 0, len(policy.Spec.Ingress))
	for _, rule := range policy.Spec.Ingress {
		ingressRule := controller.IngressRule{}

		if len(rule.From) == 0 {
			ingressRule.MatchAllSource = true
		}
		for _, peer := range rule.From {
			peerPods, err := h.evalPodsPeer(policy, peer)
			if err != nil {
				return err
			}

			for _, peerPod := range peerPods {
				if peerPod.Status.PodIP == "" {
					blog.Warnf("PodIP of %s/%s is empty", peerPod.Namespace, peerPod.Name)
					continue
				}
				ingressRule.SrcPods = append(ingressRule.SrcPods,
					controller.PodInfo{
						Name:      peerPod.Name,
						Namespace: peerPod.Namespace,
						Labels:    peerPod.Labels,
						IP:        peerPod.Status.PodIP,
					})
			}
			ingressRule.SrcIPBlocks = append(ingressRule.SrcIPBlocks, h.evalIPBlockPeer(peer)...)
		}

		// If this field is empty or missing in the spec, this rule matches all ports
		if len(rule.Ports) == 0 {
			ingressRule.MatchAllPorts = true
		} else {
			ingressRule.Ports = h.evalPorts(rule.Ports, podsNamedPorts)
		}
		newPolicy.IngressRules = append(newPolicy.IngressRules, ingressRule)
	}
	return nil
}

// buildEgressRules will calc the source pods' ips of egress rules.
func (h NetworkPolicyHandler) buildEgressRules(policy *networking.NetworkPolicy,
	newPolicy *controller.NetworkPolicyInfo,
	podsNamedPorts controller.NamedPort2eps) error {
	newPolicy.EgressRules = make([]controller.EgressRule, 0, len(policy.Spec.Egress))
	for _, rule := range policy.Spec.Egress {
		egressRule := controller.EgressRule{}

		if len(rule.To) == 0 {
			egressRule.MatchAllDestinations = true
		} else {
			for _, peer := range rule.To {
				peerPods, err := h.evalPodsPeer(policy, peer)
				if err != nil {
					return err
				}

				for _, peerPod := range peerPods {
					if peerPod.Status.PodIP == "" {
						blog.Warnf("PodIP of %s/%s is empty", peerPod.Namespace, peerPod.Name)
						continue
					}
					egressRule.DstPods = append(egressRule.DstPods,
						controller.PodInfo{
							Name:      peerPod.Name,
							Namespace: peerPod.Namespace,
							Labels:    peerPod.Labels,
							IP:        peerPod.Status.PodIP,
						})
				}
				egressRule.DstIPBlocks = append(egressRule.DstIPBlocks, h.evalIPBlockPeer(peer)...)
			}
		}

		// If this field is empty or missing in the spec, this rule matches all ports
		if len(rule.Ports) == 0 {
			egressRule.MatchAllPorts = true
		} else {
			egressRule.Ports = h.evalPorts(rule.Ports, podsNamedPorts)
		}
		newPolicy.EgressRules = append(newPolicy.EgressRules, egressRule)
	}
	return nil
}

// buildTargetPods list all target pods with policy's pod selector
func (h NetworkPolicyHandler) buildTargetPods(
	policy *networking.NetworkPolicy,
	newPolicy *controller.NetworkPolicyInfo) (matchedPods []*api.Pod, err error) {

	matchedPods, err = h.informer.ListPodsByNamespace(policy.Namespace, policy.Spec.PodSelector.MatchLabels)
	if err != nil {
		blog.Errorf("Failed to list pods with policy: %s/%s, err: %s",
			policy.Namespace, policy.Name, err.Error())
		return nil, err
	}
	newPolicy.TargetPods = make(map[string]controller.PodInfo)
	for _, pod := range matchedPods {
		podIP := pod.Status.PodIP
		if podIP == "" {
			blog.Warnf("PodIP of %s/%s is empty", pod.Namespace, pod.Name)
			continue
		}
		newPolicy.TargetPods[podIP] = controller.PodInfo{
			IP:        podIP,
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Labels:    pod.Labels,
		}
	}
	return matchedPods, nil
}

// buildPodsPorts grab the ports of pod/container, and aggregate them.
func (h *NetworkPolicyHandler) buildPodsPorts(pods []*api.Pod) controller.NamedPort2eps {
	podsNamedPorts := make(controller.NamedPort2eps)
	for _, pod := range pods {
		for _, c := range pod.Spec.Containers {
			for _, p := range c.Ports {
				name := p.Name
				if name == "" {
					continue
				}

				protocol := string(p.Protocol)
				containerPort := strconv.Itoa(int(p.ContainerPort))
				if podsNamedPorts[name] == nil {
					podsNamedPorts[name] = make(controller.Protocol2eps)
				}
				if podsNamedPorts[name][protocol] == nil {
					podsNamedPorts[name][protocol] = make(controller.NumericPort2eps)
				}
				if eps, ok := podsNamedPorts[name][protocol][containerPort]; !ok {
					podsNamedPorts[name][protocol][containerPort] = &controller.EndPoints{
						IPs:             []string{pod.Status.PodIP},
						ProtocolAndPort: controller.ProtocolAndPort{Port: containerPort, Protocol: protocol},
					}
				} else {
					eps.IPs = append(eps.IPs, pod.Status.PodIP)
				}
			}
		}
	}
	return podsNamedPorts
}

// evalPodsPeer will return the pods selected by namespaceSelector and podSelector from networkPolicy rules.
func (h NetworkPolicyHandler) evalPodsPeer(policy *networking.NetworkPolicy,
	peer networking.NetworkPolicyPeer) ([]*api.Pod, error) {
	if peer.NamespaceSelector != nil {
		matchedNs, err := h.informer.ListNamespaces(peer.NamespaceSelector.MatchLabels)
		if err != nil {
			blog.Errorf("Failed to list namespaces with policy: %s/%s, err: %s",
				policy.Namespace, policy.Name, err.Error())
			return nil, err
		}

		var pods []*api.Pod
		for _, ns := range matchedNs {
			matchedPods, err := h.informer.ListPodsByNamespace(ns.Name, peer.PodSelector.MatchLabels)
			if err != nil {
				blog.Errorf("Failed to list pods with policy: %s/%s, err: %s",
					policy.Namespace, policy.Name, err.Error())
				return nil, err
			}
			pods = append(pods, matchedPods...)
		}
		return pods, nil
	}

	if peer.PodSelector != nil {
		pods, err := h.informer.ListPodsByNamespace(policy.Namespace, peer.PodSelector.MatchLabels)
		if err != nil {
			blog.Errorf("Failed to list pods with policy: %s/%s, err: %s",
				policy.Namespace, policy.Name, err.Error())
			return nil, err
		}
		return pods, err
	}

	return nil, nil
}

// evalIPBlockPeer will calc the ipBlock filed of networkPolicy rule.
func (h NetworkPolicyHandler) evalIPBlockPeer(peer networking.NetworkPolicyPeer) [][]string {
	if peer.IPBlock == nil {
		return nil
	}

	ipBlock := make([][]string, 0, len(peer.IPBlock.Except)+1)
	if peer.PodSelector == nil && peer.NamespaceSelector == nil && peer.IPBlock != nil {
		if cidr := peer.IPBlock.CIDR; strings.HasSuffix(cidr, "/0") {
			ipBlock = append(ipBlock,
				[]string{"0.0.0.0/1", iptables.OptionTimeout, "0"},
				[]string{"128.0.0.0/1", iptables.OptionTimeout, "0"})
		} else {
			ipBlock = append(ipBlock, []string{cidr, iptables.OptionTimeout, "0"})
		}
		for _, except := range peer.IPBlock.Except {
			if strings.HasSuffix(except, "/0") {
				ipBlock = append(ipBlock,
					[]string{"0.0.0.0/1", iptables.OptionTimeout, "0", iptables.OptionNoMatch},
					[]string{"128.0.0.0/1", iptables.OptionTimeout, "0", iptables.OptionNoMatch})
			} else {
				ipBlock = append(ipBlock, []string{except, iptables.OptionTimeout, "0", iptables.OptionNoMatch})
			}
		}
	}
	return ipBlock
}

// evalPorts used to calc the ports of networkPolicy rule, with the containerPort of selected containers.
func (h NetworkPolicyHandler) evalPorts(ports []networking.NetworkPolicyPort,
	podsNamedPorts controller.NamedPort2eps) (numericPorts []controller.ProtocolAndPort) {
	for _, port := range ports {
		if port.Port == nil {
			numericPorts = append(numericPorts, controller.ProtocolAndPort{Port: "", Protocol: string(*port.Protocol)})
		} else if port.Port.Type == intstr.Int {
			numericPorts = append(numericPorts, controller.ProtocolAndPort{Port: port.Port.String(),
				Protocol: string(*port.Protocol)})
		} else {
			namedPort, ok := podsNamedPorts[port.Port.String()]
			if !ok {
				continue
			}
			nPort, ok := namedPort[string(*port.Protocol)]
			if !ok {
				continue
			}
			for _, p := range nPort {
				numericPorts = append(numericPorts, controller.ProtocolAndPort{
					Protocol: p.Protocol,
					Port:     p.Port,
				})
			}
		}
	}
	return numericPorts
}

// evalPolicyType check if there is explicitly specified PolicyTypes in the spec
func evalPolicyType(policy *networking.NetworkPolicy) controller.NetworkPolicyType {
	if len(policy.Spec.PolicyTypes) > 0 {
		ingressType, egressType := false, false
		for _, policyType := range policy.Spec.PolicyTypes {
			if policyType == networking.PolicyTypeIngress {
				ingressType = true
			}
			if policyType == networking.PolicyTypeEgress {
				egressType = true
			}
		}
		if ingressType && egressType {
			return controller.PolicyTypeBoth
		} else if egressType {
			return controller.PolicyTypeEgress
		} else if ingressType {
			return controller.PolicyTypeIngress
		}
	} else {
		if policy.Spec.Egress != nil && policy.Spec.Ingress != nil {
			return controller.PolicyTypeBoth
		} else if policy.Spec.Egress != nil {
			return controller.PolicyTypeEgress
		} else if policy.Spec.Ingress != nil {
			return controller.PolicyTypeIngress
		}
	}
	return ""
}
