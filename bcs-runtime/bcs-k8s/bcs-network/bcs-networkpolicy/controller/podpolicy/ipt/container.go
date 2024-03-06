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

package ipt

import (
	"crypto/sha256"
	"encoding/base32"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/controller"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/iptables"
	docker "github.com/fsouza/go-dockerclient"
)

// containerHandler used to sync iptables for single container
type containerHandler struct {
	iptablesClient iptables.Interface
	version        string

	container *docker.APIContainers

	ipSetHandler *iptables.IPSet
	activeChains map[string]string
	activeIPSets map[string]string
}

// Sync for-loop every network policy for container, sync iptables for ingress and egress rules.
// If ingress or egress have global rules(e.g. deny-all/allow-all), mark it and append it
// to the end of the INPUT/OUTPUT chain.
func (c *containerHandler) Sync(policyInfos []controller.NetworkPolicyInfo) error {
	start := time.Now()
	defer func() {
		end := time.Since(start)
		blog.Infof("Sync for container took: %v", end)
	}()

	// Accept localhost traffic.
	if err := c.acceptLocalHost(); err != nil {
		return err
	}

	globalIngressPolicy := make([]controller.NetworkPolicyInfo, 0)
	globalEgressPolicy := make([]controller.NetworkPolicyInfo, 0)
	for _, policy := range policyInfos {
		// if container's namespace not same with policy's namespace
		// then skip this policy
		if c.container.Labels == nil || c.container.Labels[controller.ContainerNamespaceLabel] != policy.Namespace {
			continue
		}

		// policy's labels should be subset of container's labels
		// if not then skipp the policy
		labelSame := true
		for k, v := range policy.Labels {
			cv, ok := c.container.Labels[k]
			if !ok || cv != v {
				labelSame = false
				break
			}
		}
		if !labelSame {
			continue
		}

		blog.Infof("NetworkPolicy %s/%s hit this container.", policy.Namespace, policy.Name)
		// sync ingress rules
		if policy.PolicyType == controller.PolicyTypeBoth || policy.PolicyType == controller.PolicyTypeIngress {
			if len(policy.IngressRules) == 0 {
				globalIngressPolicy = append(globalIngressPolicy, policy)
			} else {
				if err := c.syncIngressRules(&policy); err != nil {
					return err
				}
			}
		}

		// sync egress rules
		if policy.PolicyType == controller.PolicyTypeBoth || policy.PolicyType == controller.PolicyTypeEgress {
			if len(policy.EgressRules) == 0 {
				// mark the global ingress rule
				globalEgressPolicy = append(globalEgressPolicy, policy)
			} else {
				if err := c.syncEgressRules(&policy); err != nil {
					return err
				}
			}
		}
	}
	// append the global ingress rule to the end of INPUT chain
	for _, policy := range globalIngressPolicy {
		if err := c.denyAllIngress(policy); err != nil {
			return err
		}
	}
	// append the global egress rule to the end of OUTPUT chain
	for _, policy := range globalEgressPolicy {
		if err := c.denyAllEgress(policy); err != nil {
			return err
		}
	}

	// cleanup stale iptables
	if err := c.cleanupStaleIptables(); err != nil {
		return err
	}
	return nil
}

// acceptLocalHost allow localhost traffic
func (c *containerHandler) acceptLocalHost() error {
	if err := c.iptablesClient.AppendUnique("filter", "INPUT",
		"-s", "127.0.0.1", "-j", "ACCEPT"); err != nil {
		blog.Errorf("Append accept 127.0.0.1 input traffic failed, err: %s", err.Error())
		return err
	}
	if err := c.iptablesClient.AppendUnique("filter", "OUTPUT",
		"-d", "127.0.0.1", "-j", "ACCEPT"); err != nil {
		blog.Errorf("Append accept 127.0.0.1 output traffic failed, err: %s", err.Error())
		return err
	}
	return nil
}

// syncIngressRules used to sync ingress rules for the container
func (c *containerHandler) syncIngressRules(policy *controller.NetworkPolicyInfo) error {
	// create iptables chain
	chain := networkPolicyInputChainName(policy.Namespace, policy.Name, c.version)
	if err := c.newChain("filter", chain); err != nil {
		blog.Errorf("Create chain %s failed, err: %s", chain, err.Error())
		return err
	}
	if err := c.iptablesClient.AppendUnique("filter", "INPUT", "-j", chain); err != nil {
		blog.Errorf("Create relation between chain INPUT with %s failed, err: %s", chain, err.Error())
		return err
	}

	for i, rule := range policy.IngressRules {
		// ingress rules with source-pods
		if !rule.MatchAllSource && len(rule.SrcPods) != 0 {
			// create ipSet
			ipSourceSet := ipSetSourcePodName(policy.Namespace, policy.Name, strconv.Itoa(i), ruleIngress)
			if _, ok := c.activeIPSets[ipSourceSet]; ok {
				blog.Infof("IPSet %s is already created.", ipSourceSet)
			} else {
				set, err := c.ipSetHandler.Create(ipSourceSet, iptables.TypeHashIP, iptables.OptionTimeout, "0")
				if err != nil {
					blog.Errorf("Failed to create ingress sourcePods ipSet, err: %s", err.Error())
					return err
				}
				blog.Infof("Create ipSet %s success.", ipSourceSet)

				podsIPs := make([]string, 0, len(rule.SrcPods))
				for i := range rule.SrcPods {
					if rule.SrcPods[i].IP == "" {
						blog.Warnf("Pod %s has empty ip.", rule.SrcPods[i].Name)
						continue
					}
					podsIPs = append(podsIPs, rule.SrcPods[i].IP)
				}
				if err = set.Refresh(podsIPs, iptables.OptionTimeout, "0"); err != nil {
					blog.Errorf("Failed to refresh ingress sourcePods ipSet, podIPs: %v, err: %s",
						podsIPs, err.Error())
					return err
				}

				c.activeIPSets[ipSourceSet] = ipSourceSet
			}

			// append iptables
			if rule.MatchAllPorts {
				if err := c.appendRuleToChain(iptablesRule{
					SrcIPSetName: ipSourceSet,
					Behavior:     behaviorAccept,
					ChainName:    chain,
					Comment: "ACCEPT traffic from source-pods with policy " +
						policy.Namespace + "/" + policy.Name + strconv.Itoa(i),
				}); err != nil {
					blog.Errorf("Failed to append rule chain: %s", chain)
					return err
				}
			} else {
				for j := range rule.Ports {
					if err := c.appendRuleToChain(iptablesRule{
						SrcIPSetName: ipSourceSet,
						Behavior:     behaviorAccept,
						ChainName:    chain,
						Protocol:     rule.Ports[j].Protocol,
						DPort:        rule.Ports[j].Port,
						Comment: "ACCEPT traffic from source-pods to port " + rule.Ports[j].Port +
							" with policy " + policy.Namespace + "/" + policy.Name + strconv.Itoa(i),
					}); err != nil {
						blog.Errorf("Failed to append rule chain: %s", chain)
						return err
					}
				}
			}
		}

		// ingress rules with ip-blocks
		if len(rule.SrcIPBlocks) != 0 {
			// create ipSet
			ipBlockSet := ipSetIPBlockName(policy.Namespace, policy.Name, strconv.Itoa(i), ruleIngress)
			if _, ok := c.activeIPSets[ipBlockSet]; ok {
				blog.Infof("IPSet %s is already created.", ipBlockSet)
			} else {
				set, err := c.ipSetHandler.Create(ipBlockSet, iptables.TypeHashNet, iptables.OptionTimeout, "0")
				if err != nil {
					blog.Errorf("Failed to create ingress ipBlock ipSet, err: %s", err.Error())
					return err
				}
				blog.Infof("Create ipSet %s success.", ipBlockSet)

				if err = set.RefreshWithBuiltinOptions(rule.SrcIPBlocks); err != nil {
					blog.Errorf("Failed to refresh ingress ipBlock ipSet, err: %s", err.Error())
					return err
				}

				c.activeIPSets[ipBlockSet] = ipBlockSet
			}

			// append iptables
			if rule.MatchAllPorts {
				if err := c.appendRuleToChain(iptablesRule{
					SrcIPSetName: ipBlockSet,
					Behavior:     behaviorReject,
					ChainName:    chain,
					Comment: "REJECT traffic from ip-blocks with policy " +
						policy.Namespace + "/" + policy.Name + strconv.Itoa(i),
				}); err != nil {
					blog.Errorf("Failed to append rule chain: %s", chain)
					return err
				}
			} else {
				for j := range rule.Ports {
					if err := c.appendRuleToChain(iptablesRule{
						SrcIPSetName: ipBlockSet,
						Behavior:     behaviorReject,
						ChainName:    chain,
						Protocol:     rule.Ports[j].Protocol,
						DPort:        rule.Ports[j].Port,
						Comment: "REJECT traffic from ip-blocks to port " + rule.Ports[j].Port +
							" with policy " + policy.Namespace + "/" + policy.Name + strconv.Itoa(i),
					}); err != nil {
						blog.Errorf("Failed to append rule chain: %s", chain)
						return err
					}
				}
			}
		}
	}
	return nil
}

// denyAllIngress will reject all traffic to the container.
// This function will append only one rule:
//   - REJECT all ingress traffic
func (c *containerHandler) denyAllIngress(policy controller.NetworkPolicyInfo) error {
	// create iptables chain
	chain := networkPolicyInputChainName(policy.Namespace, policy.Name, c.version)
	if err := c.newChain("filter", chain); err != nil {
		blog.Errorf("Create chain %s failed, err: %s", chain, err.Error())
		return err
	}
	if err := c.iptablesClient.AppendUnique("filter", "INPUT", "-j", chain); err != nil {
		blog.Errorf("Create relation between chain INPUT with %s failed, err: %s", chain, err.Error())
		return err
	}
	blog.Infof("Create iptables chain %s success.", chain)

	// append iptables
	if err := c.appendRuleToChain(iptablesRule{
		Behavior:  behaviorReject,
		ChainName: chain,
		Comment:   "REJECT all ingress traffic with policy " + policy.Namespace + "/" + policy.Name,
	}); err != nil {
		blog.Errorf("Failed to append rule chain: %s", chain)
		return err
	}
	return nil
}

// syncEgressRules used to sync policy's egress rules for the container
func (c *containerHandler) syncEgressRules(policy *controller.NetworkPolicyInfo) error {
	// create iptables chain
	chain := networkPolicyOutputChainName(policy.Namespace, policy.Name, c.version)
	if err := c.newChain("filter", chain); err != nil {
		blog.Errorf("Create chain %s failed, err: %s", chain, err.Error())
		return err
	}
	if err := c.iptablesClient.AppendUnique("filter", "OUTPUT", "-j", chain); err != nil {
		blog.Errorf("Create relation between chain OUTPUT with %s failed, err: %s", chain, err.Error())
		return err
	}

	for i, rule := range policy.EgressRules {
		// egress rules with dst-ipBlocks
		if len(rule.DstIPBlocks) != 0 {
			// create ipSet
			ipBlockSet := ipSetIPBlockName(policy.Namespace, policy.Name, strconv.Itoa(i), ruleEgress)
			if _, ok := c.activeIPSets[ipBlockSet]; ok {
				blog.Infof("IPSet %s is already created.", ipBlockSet)
			} else {
				set, err := c.ipSetHandler.Create(ipBlockSet, iptables.TypeHashNet, iptables.OptionTimeout, "0")
				if err != nil {
					blog.Errorf("Failed to create egress ipBlock ipSet, err: %s", err.Error())
					return err
				}
				blog.Infof("Create ipSet %s success.", ipBlockSet)

				if err = set.RefreshWithBuiltinOptions(rule.DstIPBlocks); err != nil {
					blog.Errorf("Failed to refresh egress ipBlock ipSet, err: %s", err.Error())
					return err
				}

				c.activeIPSets[ipBlockSet] = ipBlockSet
			}

			// append iptables
			if rule.MatchAllPorts {
				if err := c.appendRuleToChain(iptablesRule{
					DstIPSetName: ipBlockSet,
					Behavior:     behaviorReject,
					ChainName:    chain,
					Comment: "REJECT traffic to ip-blocks with policy " +
						policy.Namespace + "/" + policy.Name + strconv.Itoa(i),
				}); err != nil {
					blog.Errorf("Failed to append rule chain: %s", chain)
					return err
				}
			} else {
				for j := range rule.Ports {
					if err := c.appendRuleToChain(iptablesRule{
						DstIPSetName: ipBlockSet,
						Behavior:     behaviorReject,
						ChainName:    chain,
						Protocol:     rule.Ports[j].Protocol,
						DPort:        rule.Ports[j].Port,
						Comment: "REJECT traffic to ip-blocks to port " + rule.Ports[j].Port +
							" with policy " + policy.Namespace + "/" + policy.Name + strconv.Itoa(i),
					}); err != nil {
						blog.Errorf("Failed to append rule chain: %s", chain)
						return err
					}
				}
			}
		}

		// egress rules with dstPods
		if !rule.MatchAllDestinations && len(rule.DstPods) != 0 {
			// create ipSet
			ipDstSet := ipSetSourcePodName(policy.Namespace, policy.Name, strconv.Itoa(i), ruleEgress)
			if _, ok := c.activeIPSets[ipDstSet]; ok {
				blog.Infof("IPSet %s is already created.", ipDstSet)
			} else {
				set, err := c.ipSetHandler.Create(ipDstSet, iptables.TypeHashIP, iptables.OptionTimeout, "0")
				if err != nil {
					blog.Errorf("Failed to create egress dstPods ipSet, err: %s", err.Error())
					return err
				}
				blog.Infof("Create ipSet %s success.", ipDstSet)

				podsIPs := make([]string, 0, len(rule.DstPods))
				for i := range rule.DstPods {
					podsIPs = append(podsIPs, rule.DstPods[i].IP)
				}
				if err = set.Refresh(podsIPs, iptables.OptionTimeout, "0"); err != nil {
					blog.Errorf("Failed to refresh egress dstPods ipSet, err: %s", err.Error())
					return err
				}

				c.activeIPSets[ipDstSet] = ipDstSet
			}

			// append iptables
			if rule.MatchAllPorts {
				if err := c.appendRuleToChain(iptablesRule{
					DstIPSetName: ipDstSet,
					Behavior:     behaviorReject,
					ChainName:    chain,
					Comment: "REJECT traffic to dstPods with policy " +
						policy.Namespace + "/" + policy.Name + strconv.Itoa(i),
				}); err != nil {
					blog.Errorf("Failed to append rule chain: %s", chain)
					return err
				}
			} else {
				for j := range rule.Ports {
					if err := c.appendRuleToChain(iptablesRule{
						DstIPSetName: ipDstSet,
						Behavior:     behaviorReject,
						ChainName:    chain,
						Protocol:     rule.Ports[j].Protocol,
						DPort:        rule.Ports[j].Port,
						Comment: "REJECT traffic to target-pods to port " + rule.Ports[j].Port +
							" with policy " + policy.Namespace + "/" + policy.Name + strconv.Itoa(i),
					}); err != nil {
						blog.Errorf("Failed to append rule chain: %s", chain)
						return err
					}
				}
			}
		}
	}
	return nil
}

// denyAllEgress will reject all traffic from the container,
// This func will append only one rule:
//   - REJECT all egress traffic
func (c *containerHandler) denyAllEgress(policy controller.NetworkPolicyInfo) error {
	// create iptables chain
	chain := networkPolicyOutputChainName(policy.Namespace, policy.Name, c.version)
	if err := c.newChain("filter", chain); err != nil {
		blog.Errorf("Create chain %s failed, err: %s", chain, err.Error())
		return err
	}
	if err := c.iptablesClient.AppendUnique("filter", "OUTPUT", "-j", chain); err != nil {
		blog.Errorf("Create relation between chain OUTPUT with %s failed, err: %s", chain, err.Error())
		return err
	}

	// sync iptables
	if err := c.appendRuleToChain(iptablesRule{
		Behavior:  behaviorReject,
		ChainName: chain,
		Comment:   "REJECT all egress traffic with policy " + policy.Namespace + "/" + policy.Name,
	}); err != nil {
		blog.Errorf("Failed to append rule chain: %s", chain)
		return err
	}
	return nil
}

// cleanupStaleIptables will cleanup all the stale iptables, after iptables synced for container.
func (c *containerHandler) cleanupStaleIptables() error {
	chains, err := c.iptablesClient.ListChains("filter")
	if err != nil {
		blog.Errorf("List chains failed, err: %s", err.Error())
		return err
	}

	for _, chain := range chains {
		if !strings.HasPrefix(chain, controller.KubeNetworkPolicyChainPrefix) {
			continue
		}
		if _, ok := c.activeChains[chain]; ok {
			continue
		}

		// delete relation in INPUT chain
		inputChainRules, err := c.iptablesClient.List("filter", "INPUT")
		if err != nil {
			blog.Errorf("List rules of INPUT chain failed, err: %s", err.Error())
			return err
		}
		for i := range inputChainRules {
			if !strings.Contains(inputChainRules[i], chain) {
				continue
			}
			if !strings.HasPrefix(inputChainRules[i], "-A INPUT") {
				continue
			}
			deleteArgs := strings.Split(strings.Split(inputChainRules[i], "-A INPUT ")[1], " ")
			if err = c.iptablesClient.Delete("filter", "INPUT", deleteArgs...); err != nil {
				blog.Errorf("Delete rule of filter-INPUT chain failed, rule: %s, err: %s",
					inputChainRules[i], err.Error())
				return err
			}
		}

		// delete relation in OUTPUT chain
		outputChainRules, err := c.iptablesClient.List("filter", "OUTPUT")
		if err != nil {
			blog.Errorf("List rules of OUTPUT chain failed, err: %s", err.Error())
			return err
		}
		for i := range outputChainRules {
			if !strings.Contains(outputChainRules[i], chain) {
				continue
			}
			if !strings.HasPrefix(outputChainRules[i], "-A OUTPUT") {
				continue
			}
			deleteArgs := strings.Split(strings.Split(outputChainRules[i], "-A OUTPUT ")[1], " ")
			if err = c.iptablesClient.Delete("filter", "OUTPUT", deleteArgs...); err != nil {
				blog.Errorf("Delete rule of filter-OUTPUT chain failed, rule: %s, err: %s",
					outputChainRules[i], err.Error())
				return err
			}
		}

		// delete custom chain
		if err := c.iptablesClient.ClearChain("filter", chain); err != nil {
			blog.Errorf("Clear chain %s failed, err: %s", chain, err.Error())
			return err
		}
		if err := c.iptablesClient.DeleteChain("filter", chain); err != nil {
			blog.Errorf("Delete chain %s failed, err: %s", chain, err.Error())
			return err
		}
		blog.Infof("Stale Chain %s is deleted.", chain)
	}
	return nil
}

func (c *containerHandler) newChain(table, chain string) error {
	_, err := c.iptablesClient.List("filter", chain)
	if err == nil {
		c.activeChains[chain] = chain
		// chain exist
		return nil
	}
	if err = c.iptablesClient.NewChain(table, chain); err != nil {
		return err
	}
	c.activeChains[chain] = chain
	return nil
}

type iptablesRule struct {
	Behavior     BehaviorType
	ChainName    string
	Comment      string
	SrcIPSetName string
	DstIPSetName string
	Protocol     string
	DPort        string
}

func (c *containerHandler) appendRuleToChain(rule iptablesRule) error {
	args := make([]string, 0)
	args = append(args, "-m", "state", "--state", "NEW")
	if rule.SrcIPSetName != "" {
		args = append(args, "-m", "set", "--match-set", rule.SrcIPSetName, "src")
	}
	if rule.DstIPSetName != "" {
		args = append(args, "-m", "set", "--match-set", rule.DstIPSetName, "dst")
	}
	if rule.Protocol != "" {
		args = append(args, "-p", rule.Protocol)
	}
	if rule.DPort != "" {
		args = append(args, "--dport", rule.DPort)
	}
	if rule.Comment != "" {
		args = append(args, "-m", "comment", "--comment", rule.Comment)
	}
	args = append(args, "-j", string(rule.Behavior))
	err := c.iptablesClient.AppendUnique("filter", rule.ChainName, args...)
	if err != nil {
		blog.Errorf("Failed to run iptables command, err: %s", rule, err.Error())
		return err
	}
	return nil
}

func networkPolicyInputChainName(namespace, policyName, version string) string {
	hash := sha256.Sum256([]byte(namespace + policyName + version))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeNetworkPolicyChainPrefix + "IN-" + encoded[:13]
}

func networkPolicyOutputChainName(namespace, policyName, version string) string {
	hash := sha256.Sum256([]byte(namespace + policyName + version))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeNetworkPolicyChainPrefix + "OUT-" + encoded[:12]
}

func ipSetIPBlockName(namespace, policyName, index string, ruleType RuleType) string {
	return ipSetName(namespace, policyName, index, ruleType, "IPBlock")
}

func ipSetSourcePodName(namespace, policyName, index string, ruleType RuleType) string {
	return ipSetName(namespace, policyName, index, ruleType, "SourcePod")
}

func ipSetName(namespace, policyName, index string, ruleType RuleType, suffix string) string {
	hash := sha256.Sum256([]byte(namespace + policyName + string(ruleType) + index + suffix))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeSourceIPSetPrefix + encoded[:16]
}
