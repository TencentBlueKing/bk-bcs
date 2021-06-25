/*
 * Modifications Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * refer to https://github.com/cloudnativelabs/kube-router
 */

package networkpolicy

import (
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/controller"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/datainformer"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/iptables"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/metrics"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/options"

	sysipt "github.com/coreos/go-iptables/iptables"
	"github.com/prometheus/client_golang/prometheus"
	api "k8s.io/api/core/v1"
	apiextensions "k8s.io/api/extensions/v1beta1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Network policy controller provides both ingress and egress filtering for the pods as per the defined network
// policies. Two different types of iptables chains are used. Each pod running on the node which either
// requires ingress or egress filtering gets a pod specific chains. Each network policy has a iptables chain, which
// has rules expressed through ipsets matching source and destination pod ip's. In the FORWARD chain of the
// filter table a rule is added to jump the traffic originating (in case of egress network policy) from the pod
// or destined (in case of ingress network policy) to the pod specific iptables chain. Each
// pod specific iptables chain has rules to jump to the network polices chains, that pod matches. So packet
// originating/destined from/to pod goes through fitler table's, FORWARD chain, followed by pod specific chain,
// followed by one or more network policy chains, till there is a match which will accept the packet, or gets
// dropped by the rule in the pod chain, if there is no match.

// NetworkPolicyController strcut to hold information required by NetworkPolicyController
type NetworkPolicyController struct {
	nodeIP net.IP
	// nodeHostName    string
	mu              sync.Mutex
	syncPeriod      time.Duration
	MetricsEnabled  bool
	readyForUpdates bool

	// list of all active network policies expressed as networkPolicyInfo
	networkPoliciesInfo *[]controller.NetworkPolicyInfo
	ipSetHandler        *iptables.IPSet

	PodEventHandler           cache.ResourceEventHandler
	NamespaceEventHandler     cache.ResourceEventHandler
	NetworkPolicyEventHandler cache.ResourceEventHandler

	// data informer
	dataInfr       datainformer.Interface
	dataInfrSynced bool
}

func (npc *NetworkPolicyController) SetDataInformerSynced() {
	npc.dataInfrSynced = true
}

// Run runs forever till we receive notification on stopCh
func (npc *NetworkPolicyController) Run(stopCh <-chan struct{}, wg *sync.WaitGroup) error {
	t := time.NewTicker(npc.syncPeriod)
	defer t.Stop()
	defer wg.Done()

	blog.Info("Starting network policy controller")

	// loop forever till notified to stop on stopCh
	for {
		select {
		case <-stopCh:
			blog.Info("Shutting down network policies controller")
			return nil
		default:
		}

		blog.V(1).Info("Performing periodic sync of iptables to reflect network policies")
		err := npc.Sync()
		if err != nil {
			blog.Errorf("nsIPTablesError during periodic sync of network policies in network policy controller. nsIPTablesError: " + err.Error())
			blog.Errorf("Skipping sending heartbeat from network policy controller as periodic sync failed.")
			metrics.ControllerIPTablesSyncError.Inc()
		}

		npc.readyForUpdates = true
		select {
		case <-stopCh:
			blog.Infof("Shutting down network policies controller")
			return nil
		case <-t.C:
		}
	}
}

// OnPodUpdate handles updates to pods from the Kubernetes api server
func (npc *NetworkPolicyController) OnPodUpdate(obj interface{}) {
	pod := obj.(*api.Pod)
	blog.V(2).Infof("Received update to pod: %s/%s", pod.Namespace, pod.Name)

	if !npc.readyForUpdates {
		blog.V(3).Infof("Skipping update to pod: %s/%s, controller still performing bootup full-sync", pod.Namespace, pod.Name)
		return
	}

	err := npc.Sync()
	if err != nil {
		blog.Errorf("nsIPTablesError syncing network policy for the update to pod: %s/%s nsIPTablesError: %s", pod.Namespace, pod.Name, err)
		metrics.ControllerIPTablesSyncError.Inc()
	}
}

// OnNetworkPolicyUpdate handles updates to network policy from the kubernetes api server
func (npc *NetworkPolicyController) OnNetworkPolicyUpdate(obj interface{}) {
	netpol := obj.(*networking.NetworkPolicy)
	blog.V(2).Infof("Received update for network policy: %s/%s", netpol.Namespace, netpol.Name)

	if !npc.readyForUpdates {
		blog.V(3).Infof("Skipping update to network policy: %s/%s, controller still performing bootup full-sync", netpol.Namespace, netpol.Name)
		return
	}

	err := npc.Sync()
	if err != nil {
		blog.Errorf("nsIPTablesError syncing network policy for the update to network policy: %s/%s nsIPTablesError: %s", netpol.Namespace, netpol.Name, err)
		metrics.ControllerIPTablesSyncError.Inc()
	}
}

// OnNamespaceUpdate handles updates to namespace from kubernetes api server
func (npc *NetworkPolicyController) OnNamespaceUpdate(obj interface{}) {
	namespace := obj.(*api.Namespace)
	blog.V(2).Infof("Received update for namespace: %s", namespace.Name)

	err := npc.Sync()
	if err != nil {
		blog.Errorf("nsIPTablesError syncing on namespace update: %s", err)
		metrics.ControllerIPTablesSyncError.Inc()
	}
}

// Sync synchronizes iptables to desired state of network policies
func (npc *NetworkPolicyController) Sync() error {

	var err error
	npc.mu.Lock()
	defer npc.mu.Unlock()

	start := time.Now()
	syncVersion := strconv.FormatInt(start.UnixNano(), 10)
	defer func() {
		endTime := time.Since(start)
		if npc.MetricsEnabled {
			metrics.ControllerIPTablesSyncTime.Observe(endTime.Seconds())
		}
		blog.V(1).Infof("sync iptables took %v", endTime)
	}()

	blog.V(1).Infof("Starting sync of iptables with version: %s", syncVersion)
	npc.networkPoliciesInfo, err = npc.buildNetworkPoliciesInfo()
	if err != nil {
		return errors.New("Aborting sync. Failed to build network policies: " + err.Error())
	}

	activePolicyChains, activePolicyIPSets, err := npc.syncNetworkPolicyChains(syncVersion)
	if err != nil {
		return errors.New("Aborting sync. Failed to sync network policy chains: " + err.Error())
	}

	activePodFwChains, err := npc.syncPodFirewallChains(syncVersion)
	if err != nil {
		return errors.New("Aborting sync. Failed to sync pod firewalls: " + err.Error())
	}

	err = cleanupStaleRules(activePolicyChains, activePodFwChains, activePolicyIPSets)
	if err != nil {
		return errors.New("Aborting sync. Failed to cleanup stale iptables rules: " + err.Error())
	}

	return nil
}

// Configure iptables rules representing each network policy. All pod's matched by
// network policy spec podselector labels are grouped together in one ipset which
// is used for matching destination ip address. Each ingress rule in the network
// policyspec is evaluated to set of matching pods, which are grouped in to a
// ipset used for source ip addr matching.
func (npc *NetworkPolicyController) syncNetworkPolicyChains(version string) (map[string]bool, map[string]bool, error) {
	start := time.Now()
	defer func() {
		endTime := time.Since(start)
		metrics.ControllerPolicyChainsSyncTime.Observe(endTime.Seconds())
		blog.V(2).Infof("Syncing network policy chains took %v", endTime)
	}()
	activePolicyChains := make(map[string]bool)
	activePolicyIPSets := make(map[string]bool)

	iptablesCmdHandler, err := iptables.New()
	if err != nil {
		blog.Fatalf("Failed to initialize iptables executor due to: %s", err.Error())
	}

	// run through all network policies
	for _, policy := range *npc.networkPoliciesInfo {

		// ensure there is a unique chain per network policy in filter table
		policyChainName := networkPolicyChainName(policy.Namespace, policy.Name, version)
		err := iptablesCmdHandler.NewChain("filter", policyChainName)
		if err != nil && err.(*sysipt.Error).ExitStatus() != 1 {
			return nil, nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}

		activePolicyChains[policyChainName] = true

		// create a ipset for all destination pod ip's matched by the policy spec PodSelector
		targetDestPodIPSetName := policyDestinationPodIPSetName(policy.Namespace, policy.Name)
		targetDestPodIPSet, err := npc.ipSetHandler.Create(targetDestPodIPSetName, iptables.TypeHashIP, iptables.OptionTimeout, "0")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create ipset: %s", err.Error())
		}

		// create a ipset for all source pod ip's matched by the policy spec PodSelector
		targetSourcePodIPSetName := policySourcePodIPSetName(policy.Namespace, policy.Name)
		targetSourcePodIPSet, err := npc.ipSetHandler.Create(targetSourcePodIPSetName, iptables.TypeHashIP, iptables.OptionTimeout, "0")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create ipset: %s", err.Error())
		}

		activePolicyIPSets[targetDestPodIPSet.Name] = true
		activePolicyIPSets[targetSourcePodIPSet.Name] = true

		currnetPodIPs := make([]string, 0, len(policy.TargetPods))
		for ip := range policy.TargetPods {
			currnetPodIPs = append(currnetPodIPs, ip)
		}

		err = targetSourcePodIPSet.Refresh(currnetPodIPs, iptables.OptionTimeout, "0")
		if err != nil {
			blog.Errorf("failed to refresh targetSourcePodIPSet: " + err.Error())
		}
		err = targetDestPodIPSet.Refresh(currnetPodIPs, iptables.OptionTimeout, "0")
		if err != nil {
			blog.Errorf("failed to refresh targetDestPodIPSet: " + err.Error())
		}

		err = npc.processIngressRules(policy, targetDestPodIPSetName, activePolicyIPSets, version)
		if err != nil {
			return nil, nil, err
		}

		err = npc.processEgressRules(policy, targetSourcePodIPSetName, activePolicyIPSets, version)
		if err != nil {
			return nil, nil, err
		}
	}

	blog.V(2).Infof("IPtables chains in the filter table are synchronized with the network policies.")

	return activePolicyChains, activePolicyIPSets, nil
}

func (npc *NetworkPolicyController) processIngressRules(policy controller.NetworkPolicyInfo,
	targetDestPodIPSetName string, activePolicyIPSets map[string]bool, version string) error {

	// From network policy spec: "If field 'Ingress' is empty then this NetworkPolicy does not allow any traffic "
	// so no whitelist rules to be added to the network policy
	if policy.IngressRules == nil {
		return nil
	}

	iptablesCmdHandler, err := iptables.New()
	if err != nil {
		return fmt.Errorf("Failed to initialize iptables executor due to: %s", err.Error())
	}

	policyChainName := networkPolicyChainName(policy.Namespace, policy.Name, version)

	// run through all the ingress rules in the spec and create iptables rules
	// in the chain for the network policy
	for i, ingressRule := range policy.IngressRules {

		if len(ingressRule.SrcPods) != 0 {
			srcPodIPSetName := policyIndexedSourcePodIPSetName(policy.Namespace, policy.Name, i)
			srcPodIPSet, err := npc.ipSetHandler.Create(srcPodIPSetName, iptables.TypeHashIP, iptables.OptionTimeout, "0")
			if err != nil {
				return fmt.Errorf("failed to create ipset: %s", err.Error())
			}

			activePolicyIPSets[srcPodIPSet.Name] = true

			ingressRuleSrcPodIPs := make([]string, 0, len(ingressRule.SrcPods))
			for _, pod := range ingressRule.SrcPods {
				ingressRuleSrcPodIPs = append(ingressRuleSrcPodIPs, pod.IP)
			}
			err = srcPodIPSet.Refresh(ingressRuleSrcPodIPs, iptables.OptionTimeout, "0")
			if err != nil {
				blog.Errorf("failed to refresh srcPodIPSet: " + err.Error())
			}

			if len(ingressRule.Ports) != 0 {
				// case where 'ports' details and 'from' details specified in the ingress rule
				// so match on specified source and destination ip's and specified port (if any) and protocol
				for _, portProtocol := range ingressRule.Ports {
					comment := "rule to ACCEPT traffic from source pods to dest pods selected by policy name " +
						policy.Name + " namespace " + policy.Namespace
					if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, srcPodIPSetName, targetDestPodIPSetName, portProtocol.Protocol, portProtocol.Port); err != nil {
						return err
					}
				}
			}

			if len(ingressRule.NamedPorts) != 0 {
				for j, endPoints := range ingressRule.NamedPorts {
					namedPortIPSetName := policyIndexedIngressNamedPortIPSetName(policy.Namespace, policy.Name, i, j)
					namedPortIPSet, err := npc.ipSetHandler.Create(namedPortIPSetName, iptables.TypeHashIP, iptables.OptionTimeout, "0")
					if err != nil {
						return fmt.Errorf("failed to create ipset: %s", err.Error())
					}
					activePolicyIPSets[namedPortIPSet.Name] = true
					err = namedPortIPSet.Refresh(endPoints.IPs, iptables.OptionTimeout, "0")
					if err != nil {
						blog.Errorf("failed to refresh namedPortIPSet: " + err.Error())
					}
					comment := "rule to ACCEPT traffic from source pods to dest pods selected by policy name " +
						policy.Name + " namespace " + policy.Namespace
					if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, srcPodIPSetName, namedPortIPSetName, endPoints.Protocol, endPoints.Port); err != nil {
						return err
					}
				}
			}

			if len(ingressRule.Ports) == 0 && len(ingressRule.NamedPorts) == 0 {
				// case where no 'ports' details specified in the ingress rule but 'from' details specified
				// so match on specified source and destination ip with all port and protocol
				comment := "rule to ACCEPT traffic from source pods to dest pods selected by policy name " +
					policy.Name + " namespace " + policy.Namespace
				if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, srcPodIPSetName, targetDestPodIPSetName, "", ""); err != nil {
					return err
				}
			}
		}

		// case where only 'ports' details specified but no 'from' details in the ingress rule
		// so match on all sources, with specified port (if any) and protocol
		if ingressRule.MatchAllSource && !ingressRule.MatchAllPorts {
			for _, portProtocol := range ingressRule.Ports {
				comment := "rule to ACCEPT traffic from all sources to dest pods selected by policy name: " +
					policy.Name + " namespace " + policy.Namespace
				if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, "", targetDestPodIPSetName, portProtocol.Protocol, portProtocol.Port); err != nil {
					return err
				}
			}

			for j, endPoints := range ingressRule.NamedPorts {
				namedPortIPSetName := policyIndexedIngressNamedPortIPSetName(policy.Namespace, policy.Name, i, j)
				namedPortIPSet, err := npc.ipSetHandler.Create(namedPortIPSetName, iptables.TypeHashIP, iptables.OptionTimeout, "0")
				if err != nil {
					return fmt.Errorf("failed to create ipset: %s", err.Error())
				}

				activePolicyIPSets[namedPortIPSet.Name] = true

				err = namedPortIPSet.Refresh(endPoints.IPs, iptables.OptionTimeout, "0")
				if err != nil {
					blog.Errorf("failed to refresh namedPortIPSet: " + err.Error())
				}
				comment := "rule to ACCEPT traffic from all sources to dest pods selected by policy name: " +
					policy.Name + " namespace " + policy.Namespace
				if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, "", namedPortIPSetName, endPoints.Protocol, endPoints.Port); err != nil {
					return err
				}
			}
		}

		// case where nether ports nor from details are speified in the ingress rule
		// so match on all ports, protocol, source IP's
		if ingressRule.MatchAllSource && ingressRule.MatchAllPorts {
			comment := "rule to ACCEPT traffic from all sources to dest pods selected by policy name: " +
				policy.Name + " namespace " + policy.Namespace
			if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, "", targetDestPodIPSetName, "", ""); err != nil {
				return err
			}
		}

		if len(ingressRule.SrcIPBlocks) != 0 {
			srcIPBlockIPSetName := policyIndexedSourceIPBlockIPSetName(policy.Namespace, policy.Name, i)
			srcIPBlockIPSet, err := npc.ipSetHandler.Create(srcIPBlockIPSetName, iptables.TypeHashNet, iptables.OptionTimeout, "0")
			if err != nil {
				return fmt.Errorf("failed to create ipset: %s", err.Error())
			}
			activePolicyIPSets[srcIPBlockIPSet.Name] = true
			err = srcIPBlockIPSet.RefreshWithBuiltinOptions(ingressRule.SrcIPBlocks)
			if err != nil {
				blog.Errorf("failed to refresh srcIPBlockIPSet: " + err.Error())
			}
			if !ingressRule.MatchAllPorts {
				for _, portProtocol := range ingressRule.Ports {
					comment := "rule to ACCEPT traffic from specified ipBlocks to dest pods selected by policy name: " +
						policy.Name + " namespace " + policy.Namespace
					if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, srcIPBlockIPSetName, targetDestPodIPSetName, portProtocol.Protocol, portProtocol.Port); err != nil {
						return err
					}
				}

				for j, endPoints := range ingressRule.NamedPorts {
					namedPortIPSetName := policyIndexedIngressNamedPortIPSetName(policy.Namespace, policy.Name, i, j)
					namedPortIPSet, err := npc.ipSetHandler.Create(namedPortIPSetName, iptables.TypeHashIP, iptables.OptionTimeout, "0")
					if err != nil {
						return fmt.Errorf("failed to create ipset: %s", err.Error())
					}

					activePolicyIPSets[namedPortIPSet.Name] = true

					err = namedPortIPSet.Refresh(endPoints.IPs, iptables.OptionTimeout, "0")
					if err != nil {
						blog.Errorf("failed to refresh namedPortIPSet: " + err.Error())
					}
					comment := "rule to ACCEPT traffic from specified ipBlocks to dest pods selected by policy name: " +
						policy.Name + " namespace " + policy.Namespace
					if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, srcIPBlockIPSetName, namedPortIPSetName, endPoints.Protocol, endPoints.Port); err != nil {
						return err
					}
				}
			}
			if ingressRule.MatchAllPorts {
				comment := "rule to ACCEPT traffic from specified ipBlocks to dest pods selected by policy name: " +
					policy.Name + " namespace " + policy.Namespace
				if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, srcIPBlockIPSetName, targetDestPodIPSetName, "", ""); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (npc *NetworkPolicyController) processEgressRules(policy controller.NetworkPolicyInfo,
	targetSourcePodIPSetName string, activePolicyIPSets map[string]bool, version string) error {

	// From network policy spec: "If field 'Ingress' is empty then this NetworkPolicy does not allow any traffic "
	// so no whitelist rules to be added to the network policy
	if policy.EgressRules == nil {
		return nil
	}

	iptablesCmdHandler, err := iptables.New()
	if err != nil {
		return fmt.Errorf("Failed to initialize iptables executor due to: %s", err.Error())
	}

	policyChainName := networkPolicyChainName(policy.Namespace, policy.Name, version)

	// run through all the egress rules in the spec and create iptables rules
	// in the chain for the network policy
	for i, egressRule := range policy.EgressRules {

		if len(egressRule.DstPods) != 0 {
			dstPodIPSetName := policyIndexedDestinationPodIPSetName(policy.Namespace, policy.Name, i)
			dstPodIPSet, err := npc.ipSetHandler.Create(dstPodIPSetName, iptables.TypeHashIP, iptables.OptionTimeout, "0")
			if err != nil {
				return fmt.Errorf("failed to create ipset: %s", err.Error())
			}

			activePolicyIPSets[dstPodIPSet.Name] = true

			egressRuleDstPodIPs := make([]string, 0, len(egressRule.DstPods))
			for _, pod := range egressRule.DstPods {
				egressRuleDstPodIPs = append(egressRuleDstPodIPs, pod.IP)
			}
			err = dstPodIPSet.Refresh(egressRuleDstPodIPs, iptables.OptionTimeout, "0")
			if err != nil {
				blog.Errorf("failed to refresh dstPodIPSet: " + err.Error())
			}
			if len(egressRule.Ports) != 0 {
				// case where 'ports' details and 'from' details specified in the egress rule
				// so match on specified source and destination ip's and specified port (if any) and protocol
				for _, portProtocol := range egressRule.Ports {
					comment := "rule to ACCEPT traffic from source pods to dest pods selected by policy name " +
						policy.Name + " namespace " + policy.Namespace
					if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, targetSourcePodIPSetName, dstPodIPSetName, portProtocol.Protocol, portProtocol.Port); err != nil {
						return err
					}
				}
			}

			if len(egressRule.NamedPorts) != 0 {
				for j, endPoints := range egressRule.NamedPorts {
					namedPortIPSetName := policyIndexedEgressNamedPortIPSetName(policy.Namespace, policy.Name, i, j)
					namedPortIPSet, err := npc.ipSetHandler.Create(namedPortIPSetName, iptables.TypeHashIP, iptables.OptionTimeout, "0")
					if err != nil {
						return fmt.Errorf("failed to create ipset: %s", err.Error())
					}

					activePolicyIPSets[namedPortIPSet.Name] = true

					err = namedPortIPSet.Refresh(endPoints.IPs, iptables.OptionTimeout, "0")
					if err != nil {
						blog.Errorf("failed to refresh namedPortIPSet: " + err.Error())
					}
					comment := "rule to ACCEPT traffic from source pods to dest pods selected by policy name " +
						policy.Name + " namespace " + policy.Namespace
					if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, targetSourcePodIPSetName, namedPortIPSetName, endPoints.Protocol, endPoints.Port); err != nil {
						return err
					}
				}

			}

			if len(egressRule.Ports) == 0 && len(egressRule.NamedPorts) == 0 {
				// case where no 'ports' details specified in the ingress rule but 'from' details specified
				// so match on specified source and destination ip with all port and protocol
				comment := "rule to ACCEPT traffic from source pods to dest pods selected by policy name " +
					policy.Name + " namespace " + policy.Namespace
				if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, targetSourcePodIPSetName, dstPodIPSetName, "", ""); err != nil {
					return err
				}
			}
		}

		// case where only 'ports' details specified but no 'to' details in the egress rule
		// so match on all sources, with specified port (if any) and protocol
		if egressRule.MatchAllDestinations && !egressRule.MatchAllPorts {
			for _, portProtocol := range egressRule.Ports {
				comment := "rule to ACCEPT traffic from source pods to all destinations selected by policy name: " +
					policy.Name + " namespace " + policy.Namespace
				if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, targetSourcePodIPSetName, "", portProtocol.Protocol, portProtocol.Port); err != nil {
					return err
				}
			}
		}

		// case where nether ports nor from details are speified in the egress rule
		// so match on all ports, protocol, source IP's
		if egressRule.MatchAllDestinations && egressRule.MatchAllPorts {
			comment := "rule to ACCEPT traffic from source pods to all destinations selected by policy name: " +
				policy.Name + " namespace " + policy.Namespace
			if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, targetSourcePodIPSetName, "", "", ""); err != nil {
				return err
			}
		}
		if len(egressRule.DstIPBlocks) != 0 {
			dstIPBlockIPSetName := policyIndexedDestinationIPBlockIPSetName(policy.Namespace, policy.Name, i)
			dstIPBlockIPSet, err := npc.ipSetHandler.Create(dstIPBlockIPSetName, iptables.TypeHashNet, iptables.OptionTimeout, "0")
			if err != nil {
				return fmt.Errorf("failed to create ipset: %s", err.Error())
			}
			activePolicyIPSets[dstIPBlockIPSet.Name] = true
			err = dstIPBlockIPSet.RefreshWithBuiltinOptions(egressRule.DstIPBlocks)
			if err != nil {
				blog.Errorf("failed to refresh dstIPBlockIPSet: " + err.Error())
			}
			if !egressRule.MatchAllPorts {
				for _, portProtocol := range egressRule.Ports {
					comment := "rule to ACCEPT traffic from source pods to specified ipBlocks selected by policy name: " +
						policy.Name + " namespace " + policy.Namespace
					if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, targetSourcePodIPSetName, dstIPBlockIPSetName, portProtocol.Protocol, portProtocol.Port); err != nil {
						return err
					}
				}
			}
			if egressRule.MatchAllPorts {
				comment := "rule to ACCEPT traffic from source pods to specified ipBlocks selected by policy name: " +
					policy.Name + " namespace " + policy.Namespace
				if err := npc.appendRuleToPolicyChain(iptablesCmdHandler, policyChainName, comment, targetSourcePodIPSetName, dstIPBlockIPSetName, "", ""); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (npc *NetworkPolicyController) appendRuleToPolicyChain(iptablesCmdHandler iptables.Interface, policyChainName, comment, srcIPSetName, dstIPSetName, protocol, dPort string) error {
	if iptablesCmdHandler == nil {
		return fmt.Errorf("Failed to run iptables command: iptablesCmdHandler is nil")
	}
	args := make([]string, 0)
	if comment != "" {
		args = append(args, "-m", "comment", "--comment", comment)
	}
	if srcIPSetName != "" {
		args = append(args, "-m", "set", "--set", srcIPSetName, "src")
	}
	if dstIPSetName != "" {
		args = append(args, "-m", "set", "--set", dstIPSetName, "dst")
	}
	if protocol != "" {
		args = append(args, "-p", protocol)
	}
	if dPort != "" {
		args = append(args, "--dport", dPort)
	}
	args = append(args, "-j", "ACCEPT")
	err := iptablesCmdHandler.AppendUnique("filter", policyChainName, args...)
	if err != nil {
		return fmt.Errorf("Failed to run iptables command: %s", err.Error())
	}
	return nil
}

func (npc *NetworkPolicyController) syncPodFirewallChains(version string) (map[string]bool, error) {

	activePodFwChains := make(map[string]bool)

	iptablesCmdHandler, err := iptables.New()
	if err != nil {
		blog.Fatalf("Failed to initialize iptables executor: %s", err.Error())
	}

	// loop through the pods running on the node which to which ingress network policies to be applied
	ingressNetworkPolicyEnabledPods, err := npc.getIngressNetworkPolicyEnabledPods(npc.nodeIP.String())
	if err != nil {
		return nil, err
	}
	for _, pod := range *ingressNetworkPolicyEnabledPods {

		// below condition occurs when we get trasient update while removing or adding pod
		// subseqent update will do the correct action
		if len(pod.IP) == 0 || pod.IP == "" {
			continue
		}

		// ensure pod specific firewall chain exist for all the pods that need ingress firewall
		podFwChainName := podFirewallChainName(pod.Namespace, pod.Name, version)
		err = iptablesCmdHandler.NewChain("filter", podFwChainName)
		if err != nil && err.(*sysipt.Error).ExitStatus() != 1 {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
		activePodFwChains[podFwChainName] = true

		// add entries in pod firewall to run through required network policies
		for _, policy := range *npc.networkPoliciesInfo {
			if _, ok := policy.TargetPods[pod.IP]; ok {
				comment := "run through nw policy " + policy.Name
				policyChainName := networkPolicyChainName(policy.Namespace, policy.Name, version)
				args := []string{"-m", "comment", "--comment", comment, "-j", policyChainName}
				exists, err := iptablesCmdHandler.Exists("filter", podFwChainName, args...)
				if err != nil {
					return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
				}
				if !exists {
					err := iptablesCmdHandler.Insert("filter", podFwChainName, 1, args...)
					if err != nil && err.(*sysipt.Error).ExitStatus() != 1 {
						return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
					}
				}
			}
		}

		comment := "rule to permit the traffic traffic to pods when source is the pod's local node"
		args := []string{"-m", "comment", "--comment", comment, "-m", "addrtype", "--src-type", "LOCAL", "-d", pod.IP, "-j", "ACCEPT"}
		exists, err := iptablesCmdHandler.Exists("filter", podFwChainName, args...)
		if err != nil {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
		if !exists {
			err := iptablesCmdHandler.Insert("filter", podFwChainName, 1, args...)
			if err != nil {
				return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
			}
		}

		// ensure statefull firewall, that permits return traffic for the traffic originated by the pod
		comment = "rule for stateful firewall for pod"
		args = []string{"-m", "comment", "--comment", comment, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"}
		exists, err = iptablesCmdHandler.Exists("filter", podFwChainName, args...)
		if err != nil {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
		if !exists {
			err := iptablesCmdHandler.Insert("filter", podFwChainName, 1, args...)
			if err != nil {
				return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
			}
		}

		// ensure there is rule in filter table and FORWARD chain to jump to pod specific firewall chain
		// this rule applies to the traffic getting routed (coming for other node pods)
		comment = "rule to jump traffic destined to POD name:" + pod.Name + " namespace: " + pod.Namespace +
			" to chain " + podFwChainName
		args = []string{"-m", "comment", "--comment", comment, "-d", pod.IP, "-j", podFwChainName}
		exists, err = iptablesCmdHandler.Exists("filter", "FORWARD", args...)
		if err != nil {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
		if !exists {
			err := iptablesCmdHandler.Insert("filter", "FORWARD", 1, args...)
			if err != nil {
				return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
			}
		}

		// ensure there is rule in filter table and OUTPUT chain to jump to pod specific firewall chain
		// this rule applies to the traffic from a pod getting routed back to another pod on same node by service proxy
		exists, err = iptablesCmdHandler.Exists("filter", "OUTPUT", args...)
		if err != nil {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
		if !exists {
			err := iptablesCmdHandler.Insert("filter", "OUTPUT", 1, args...)
			if err != nil {
				return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
			}
		}

		// ensure there is rule in filter table and forward chain to jump to pod specific firewall chain
		// this rule applies to the traffic getting switched (coming for same node pods)
		comment = "rule to jump traffic destined to POD name:" + pod.Name + " namespace: " + pod.Namespace +
			" to chain " + podFwChainName
		args = []string{"-m", "physdev", "--physdev-is-bridged",
			"-m", "comment", "--comment", comment,
			"-d", pod.IP,
			"-j", podFwChainName}
		exists, err = iptablesCmdHandler.Exists("filter", "FORWARD", args...)
		if err != nil {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
		if !exists {
			err = iptablesCmdHandler.Insert("filter", "FORWARD", 1, args...)
			if err != nil {
				return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
			}
		}

		// add default DROP rule at the end of chain
		comment = "default rule to REJECT traffic destined for POD name:" + pod.Name + " namespace: " + pod.Namespace
		args = []string{"-m", "comment", "--comment", comment, "-j", "REJECT"}
		err = iptablesCmdHandler.AppendUnique("filter", podFwChainName, args...)
		if err != nil {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
	}

	// loop through the pods running on the node which egress network policies to be applied
	egressNetworkPolicyEnabledPods, err := npc.getEgressNetworkPolicyEnabledPods(npc.nodeIP.String())
	if err != nil {
		return nil, err
	}
	for _, pod := range *egressNetworkPolicyEnabledPods {

		// below condition occurs when we get trasient update while removing or adding pod
		// subseqent update will do the correct action
		if len(pod.IP) == 0 || pod.IP == "" {
			continue
		}

		// ensure pod specific firewall chain exist for all the pods that need egress firewall
		podFwChainName := podFirewallChainName(pod.Namespace, pod.Name, version)
		err = iptablesCmdHandler.NewChain("filter", podFwChainName)
		if err != nil && err.(*sysipt.Error).ExitStatus() != 1 {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
		activePodFwChains[podFwChainName] = true

		// add entries in pod firewall to run through required network policies
		for _, policy := range *npc.networkPoliciesInfo {
			if _, ok := policy.TargetPods[pod.IP]; ok {
				comment := "run through nw policy " + policy.Name
				policyChainName := networkPolicyChainName(policy.Namespace, policy.Name, version)
				args := []string{"-m", "comment", "--comment", comment, "-j", policyChainName}
				exists, err := iptablesCmdHandler.Exists("filter", podFwChainName, args...)
				if err != nil {
					return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
				}
				if !exists {
					err := iptablesCmdHandler.Insert("filter", podFwChainName, 1, args...)
					if err != nil && err.(*sysipt.Error).ExitStatus() != 1 {
						return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
					}
				}
			}
		}

		// ensure statefull firewall, that permits return traffic for the traffic originated by the pod
		comment := "rule for stateful firewall for pod"
		args := []string{"-m", "comment", "--comment", comment, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"}
		exists, err := iptablesCmdHandler.Exists("filter", podFwChainName, args...)
		if err != nil {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
		if !exists {
			err := iptablesCmdHandler.Insert("filter", podFwChainName, 1, args...)
			if err != nil {
				return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
			}
		}

		// ensure there is rule in filter table and FORWARD chain to jump to pod specific firewall chain
		// this rule applies to the traffic getting routed (coming for other node pods)
		comment = "rule to jump traffic from POD name:" + pod.Name + " namespace: " + pod.Namespace +
			" to chain " + podFwChainName
		args = []string{"-m", "comment", "--comment", comment, "-s", pod.IP, "-j", podFwChainName}
		exists, err = iptablesCmdHandler.Exists("filter", "FORWARD", args...)
		if err != nil {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
		if !exists {
			err := iptablesCmdHandler.Insert("filter", "FORWARD", 1, args...)
			if err != nil {
				return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
			}
		}

		// ensure there is rule in filter table and forward chain to jump to pod specific firewall chain
		// this rule applies to the traffic getting switched (coming for same node pods)
		comment = "rule to jump traffic from POD name:" + pod.Name + " namespace: " + pod.Namespace +
			" to chain " + podFwChainName
		args = []string{"-m", "physdev", "--physdev-is-bridged",
			"-m", "comment", "--comment", comment,
			"-s", pod.IP,
			"-j", podFwChainName}
		exists, err = iptablesCmdHandler.Exists("filter", "FORWARD", args...)
		if err != nil {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
		if !exists {
			err = iptablesCmdHandler.Insert("filter", "FORWARD", 1, args...)
			if err != nil {
				return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
			}
		}

		// add default DROP rule at the end of chain
		comment = "default rule to REJECT traffic destined for POD name:" + pod.Name + " namespace: " + pod.Namespace
		args = []string{"-m", "comment", "--comment", comment, "-j", "REJECT"}
		err = iptablesCmdHandler.AppendUnique("filter", podFwChainName, args...)
		if err != nil {
			return nil, fmt.Errorf("Failed to run iptables command: %s", err.Error())
		}
	}

	return activePodFwChains, nil
}

func cleanupStaleRules(activePolicyChains, activePodFwChains, activePolicyIPSets map[string]bool) error {

	cleanupPodFwChains := make([]string, 0)
	cleanupPolicyChains := make([]string, 0)
	cleanupPolicyIPSets := make([]*iptables.Set, 0)

	iptablesCmdHandler, err := iptables.New()
	if err != nil {
		blog.Fatalf("failed to initialize iptables command executor due to %s", err.Error())
	}
	ipsets, err := iptables.NewIPSet(false)
	if err != nil {
		blog.Fatalf("failed to create ipsets command executor due to %s", err.Error())
	}
	err = ipsets.Save()
	if err != nil {
		blog.Fatalf("failed to initialize ipsets command executor due to %s", err.Error())
	}

	// get the list of chains created for pod firewall and network policies
	chains, err := iptablesCmdHandler.ListChains("filter")
	for _, chain := range chains {
		if strings.HasPrefix(chain, controller.KubeNetworkPolicyChainPrefix) {
			if _, ok := activePolicyChains[chain]; !ok {
				cleanupPolicyChains = append(cleanupPolicyChains, chain)
			}
		}
		if strings.HasPrefix(chain, controller.KubePodFirewallChainPrefix) {
			if _, ok := activePodFwChains[chain]; !ok {
				cleanupPodFwChains = append(cleanupPodFwChains, chain)
			}
		}
	}
	for _, set := range ipsets.Sets {
		if strings.HasPrefix(set.Name, controller.KubeSourceIPSetPrefix) ||
			strings.HasPrefix(set.Name, controller.KubeDestinationIPSetPrefix) {
			if _, ok := activePolicyIPSets[set.Name]; !ok {
				cleanupPolicyIPSets = append(cleanupPolicyIPSets, set)
			}
		}
	}

	// cleanup FORWARD chain rules to jump to pod firewall
	for _, chain := range cleanupPodFwChains {

		forwardChainRules, err := iptablesCmdHandler.List("filter", "FORWARD")
		if err != nil {
			return fmt.Errorf("failed to list rules in filter table, FORWARD chain due to %s", err.Error())
		}
		outputChainRules, err := iptablesCmdHandler.List("filter", "OUTPUT")
		if err != nil {
			return fmt.Errorf("failed to list rules in filter table, OUTPUT chain due to %s", err.Error())
		}

		// TODO delete rule by spec, than rule number to avoid extra loop
		var realRuleNo int
		for i, rule := range forwardChainRules {
			if strings.Contains(rule, chain) {
				err = iptablesCmdHandler.Delete("filter", "FORWARD", strconv.Itoa(i-realRuleNo))
				if err != nil {
					return fmt.Errorf("failed to delete rule: %s from the FORWARD chain of filter table due to %s", rule, err.Error())
				}
				realRuleNo++
			}
		}
		realRuleNo = 0
		for i, rule := range outputChainRules {
			if strings.Contains(rule, chain) {
				err = iptablesCmdHandler.Delete("filter", "OUTPUT", strconv.Itoa(i-realRuleNo))
				if err != nil {
					return fmt.Errorf("failed to delete rule: %s from the OUTPUT chain of filter table due to %s", rule, err.Error())
				}
				realRuleNo++
			}
		}
	}

	// cleanup pod firewall chain
	for _, chain := range cleanupPodFwChains {
		blog.V(2).Infof("Found pod fw chain to cleanup: %s", chain)
		err = iptablesCmdHandler.ClearChain("filter", chain)
		if err != nil {
			return fmt.Errorf("Failed to flush the rules in chain %s due to %s", chain, err.Error())
		}
		err = iptablesCmdHandler.DeleteChain("filter", chain)
		if err != nil {
			return fmt.Errorf("Failed to delete the chain %s due to %s", chain, err.Error())
		}
		blog.V(2).Infof("Deleted pod specific firewall chain: %s from the filter table", chain)
	}

	// cleanup network policy chains
	for _, policyChain := range cleanupPolicyChains {
		blog.V(2).Infof("Found policy chain to cleanup %s", policyChain)

		// first clean up any references from pod firewall chain
		for podFwChain := range activePodFwChains {
			podFwChainRules, err := iptablesCmdHandler.List("filter", podFwChain)
			if err != nil {

			}
			for i, rule := range podFwChainRules {
				if strings.Contains(rule, policyChain) {
					err = iptablesCmdHandler.Delete("filter", podFwChain, strconv.Itoa(i))
					if err != nil {
						return fmt.Errorf("Failed to delete rule %s from the chain %s", rule, podFwChain)
					}
					break
				}
			}
		}

		err = iptablesCmdHandler.ClearChain("filter", policyChain)
		if err != nil {
			return fmt.Errorf("Failed to flush the rules in chain %s due to  %s", policyChain, err)
		}
		err = iptablesCmdHandler.DeleteChain("filter", policyChain)
		if err != nil {
			return fmt.Errorf("Failed to flush the rules in chain %s due to %s", policyChain, err)
		}
		blog.V(2).Infof("Deleted network policy chain: %s from the filter table", policyChain)
	}

	// cleanup network policy ipsets
	for _, set := range cleanupPolicyIPSets {
		err = set.Destroy()
		if err != nil {
			return fmt.Errorf("Failed to delete ipset %s due to %s", set.Name, err)
		}
	}
	return nil
}

func (npc *NetworkPolicyController) getIngressNetworkPolicyEnabledPods(nodeIP string) (*map[string]controller.PodInfo, error) {
	nodePods := make(map[string]controller.PodInfo)

	podList, err := npc.dataInfr.ListAllPods()
	if err != nil {
		return nil, fmt.Errorf("list all pods failed, err %s", err.Error())
	}

	for _, pod := range podList {
		if strings.Compare(pod.Status.HostIP, nodeIP) != 0 {
			continue
		}
		for _, policy := range *npc.networkPoliciesInfo {
			if policy.Namespace != pod.ObjectMeta.Namespace {
				continue
			}
			_, ok := policy.TargetPods[pod.Status.PodIP]
			if ok && (policy.PolicyType == "both" || policy.PolicyType == "ingress") {
				blog.V(2).Infof("Found pod name: " + pod.ObjectMeta.Name + " namespace: " + pod.ObjectMeta.Namespace + " for which network policies need to be applied.")
				nodePods[pod.Status.PodIP] = controller.PodInfo{IP: pod.Status.PodIP,
					Name:      pod.ObjectMeta.Name,
					Namespace: pod.ObjectMeta.Namespace,
					Labels:    pod.ObjectMeta.Labels}
				break
			}
		}
	}
	return &nodePods, nil

}

func (npc *NetworkPolicyController) getEgressNetworkPolicyEnabledPods(nodeIP string) (*map[string]controller.PodInfo, error) {

	nodePods := make(map[string]controller.PodInfo)

	podList, err := npc.dataInfr.ListAllPods()
	if err != nil {
		return nil, fmt.Errorf("list all pods failed, err %s", err.Error())
	}

	for _, pod := range podList {

		if strings.Compare(pod.Status.HostIP, nodeIP) != 0 {
			continue
		}
		for _, policy := range *npc.networkPoliciesInfo {
			if policy.Namespace != pod.ObjectMeta.Namespace {
				continue
			}
			_, ok := policy.TargetPods[pod.Status.PodIP]
			if ok && (policy.PolicyType == "both" || policy.PolicyType == "egress") {
				blog.V(2).Infof("Found pod name: " + pod.ObjectMeta.Name + " namespace: " + pod.ObjectMeta.Namespace + " for which network policies need to be applied.")
				nodePods[pod.Status.PodIP] = controller.PodInfo{IP: pod.Status.PodIP,
					Name:      pod.ObjectMeta.Name,
					Namespace: pod.ObjectMeta.Namespace,
					Labels:    pod.ObjectMeta.Labels}
				break
			}
		}
	}
	return &nodePods, nil
}

func (npc *NetworkPolicyController) processNetworkPolicyPorts(npPorts []networking.NetworkPolicyPort, namedPort2eps controller.NamedPort2eps) (numericPorts []controller.ProtocolAndPort, namedPorts []controller.EndPoints) {
	numericPorts, namedPorts = make([]controller.ProtocolAndPort, 0), make([]controller.EndPoints, 0)
	for _, npPort := range npPorts {
		if npPort.Port == nil {
			numericPorts = append(numericPorts, controller.ProtocolAndPort{Port: "", Protocol: string(*npPort.Protocol)})
		} else if npPort.Port.Type == intstr.Int {
			numericPorts = append(numericPorts, controller.ProtocolAndPort{Port: npPort.Port.String(), Protocol: string(*npPort.Protocol)})
		} else {
			if protocol2eps, ok := namedPort2eps[npPort.Port.String()]; ok {
				if numericPort2eps, ok := protocol2eps[string(*npPort.Protocol)]; ok {
					for _, eps := range numericPort2eps {
						namedPorts = append(namedPorts, *eps)
					}
				}
			}
		}
	}
	return
}

func (npc *NetworkPolicyController) processBetaNetworkPolicyPorts(npPorts []apiextensions.NetworkPolicyPort, namedPort2eps controller.NamedPort2eps) (numericPorts []controller.ProtocolAndPort, namedPorts []controller.EndPoints) {
	numericPorts, namedPorts = make([]controller.ProtocolAndPort, 0), make([]controller.EndPoints, 0)
	for _, npPort := range npPorts {
		if npPort.Port == nil {
			numericPorts = append(numericPorts, controller.ProtocolAndPort{Port: "", Protocol: string(*npPort.Protocol)})
		} else if npPort.Port.Type == intstr.Int {
			numericPorts = append(numericPorts, controller.ProtocolAndPort{Port: npPort.Port.String(), Protocol: string(*npPort.Protocol)})
		} else {
			if protocol2eps, ok := namedPort2eps[npPort.Port.String()]; ok {
				if numericPort2eps, ok := protocol2eps[string(*npPort.Protocol)]; ok {
					for _, eps := range numericPort2eps {
						namedPorts = append(namedPorts, *eps)
					}
				}
			}
		}
	}
	return
}

func (npc *NetworkPolicyController) buildNetworkPoliciesInfo() (*[]controller.NetworkPolicyInfo, error) {

	NetworkPolicies := make([]controller.NetworkPolicyInfo, 0)
	npList, err := npc.dataInfr.ListAllNetworkPolicy()
	if err != nil {
		return nil, fmt.Errorf("dataInformer list network policy falied, err %s", err.Error())
	}
	//construct local network policy info according NetworkPolicy definition
	for _, policy := range npList {
		newPolicy := controller.NetworkPolicyInfo{
			Name:       policy.Name,
			Namespace:  policy.Namespace,
			Labels:     policy.Spec.PodSelector.MatchLabels,
			PolicyType: "ingress",
		}

		// check if there is explicitly specified PolicyTypes in the spec
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
				newPolicy.PolicyType = controller.PolicyTypeBoth
			} else if egressType {
				newPolicy.PolicyType = controller.PolicyTypeEgress
			} else if ingressType {
				newPolicy.PolicyType = controller.PolicyTypeIngress
			}
		} else {
			if policy.Spec.Egress != nil && policy.Spec.Ingress != nil {
				newPolicy.PolicyType = controller.PolicyTypeBoth
			} else if policy.Spec.Egress != nil {
				newPolicy.PolicyType = controller.PolicyTypeEgress
			} else if policy.Spec.Ingress != nil {
				newPolicy.PolicyType = controller.PolicyTypeIngress
			}
		}

		matchingPods, err := npc.ListPodsByNamespaceAndLabels(policy.Namespace, policy.Spec.PodSelector.MatchLabels)
		newPolicy.TargetPods = make(map[string]controller.PodInfo)
		namedPort2IngressEps := make(controller.NamedPort2eps)
		if err == nil {
			for _, matchingPod := range matchingPods {
				if matchingPod.Status.PodIP == "" {
					continue
				}
				newPolicy.TargetPods[matchingPod.Status.PodIP] = controller.PodInfo{IP: matchingPod.Status.PodIP,
					Name:      matchingPod.ObjectMeta.Name,
					Namespace: matchingPod.ObjectMeta.Namespace,
					Labels:    matchingPod.ObjectMeta.Labels}
				//grab all port information from pod
				npc.grabNamedPortFromPod(matchingPod, &namedPort2IngressEps)
			}
		}

		if policy.Spec.Ingress == nil {
			newPolicy.IngressRules = nil
		} else {
			newPolicy.IngressRules = make([]controller.IngressRule, 0)
		}

		if policy.Spec.Egress == nil {
			newPolicy.EgressRules = nil
		} else {
			newPolicy.EgressRules = make([]controller.EgressRule, 0)
		}

		for _, specIngressRule := range policy.Spec.Ingress {
			ingressRule := controller.IngressRule{}
			ingressRule.SrcPods = make([]controller.PodInfo, 0)
			ingressRule.SrcIPBlocks = make([][]string, 0)

			// If this field is empty or missing in the spec, this rule matches all sources
			if len(specIngressRule.From) == 0 {
				ingressRule.MatchAllSource = true
			} else {
				ingressRule.MatchAllSource = false
				for _, peer := range specIngressRule.From {
					if peerPods, err := npc.evalPodPeer(policy, peer); err == nil {
						for _, peerPod := range peerPods {
							if peerPod.Status.PodIP == "" {
								continue
							}
							ingressRule.SrcPods = append(ingressRule.SrcPods,
								controller.PodInfo{IP: peerPod.Status.PodIP,
									Name:      peerPod.ObjectMeta.Name,
									Namespace: peerPod.ObjectMeta.Namespace,
									Labels:    peerPod.ObjectMeta.Labels})
						}
					}
					ingressRule.SrcIPBlocks = append(ingressRule.SrcIPBlocks, npc.evalIPBlockPeer(peer)...)
				}
			}

			ingressRule.Ports = make([]controller.ProtocolAndPort, 0)
			ingressRule.NamedPorts = make([]controller.EndPoints, 0)
			// If this field is empty or missing in the spec, this rule matches all ports
			if len(specIngressRule.Ports) == 0 {
				ingressRule.MatchAllPorts = true
			} else {
				ingressRule.MatchAllPorts = false
				ingressRule.Ports, ingressRule.NamedPorts = npc.processNetworkPolicyPorts(specIngressRule.Ports, namedPort2IngressEps)
			}

			newPolicy.IngressRules = append(newPolicy.IngressRules, ingressRule)
		}

		for _, specEgressRule := range policy.Spec.Egress {
			egressRule := controller.EgressRule{}
			egressRule.DstPods = make([]controller.PodInfo, 0)
			egressRule.DstIPBlocks = make([][]string, 0)
			namedPort2EgressEps := make(controller.NamedPort2eps)

			// If this field is empty or missing in the spec, this rule matches all sources
			if len(specEgressRule.To) == 0 {
				egressRule.MatchAllDestinations = true
			} else {
				egressRule.MatchAllDestinations = false
				for _, peer := range specEgressRule.To {
					if peerPods, err := npc.evalPodPeer(policy, peer); err == nil {
						for _, peerPod := range peerPods {
							if peerPod.Status.PodIP == "" {
								continue
							}
							egressRule.DstPods = append(egressRule.DstPods,
								controller.PodInfo{IP: peerPod.Status.PodIP,
									Name:      peerPod.ObjectMeta.Name,
									Namespace: peerPod.ObjectMeta.Namespace,
									Labels:    peerPod.ObjectMeta.Labels})
							npc.grabNamedPortFromPod(peerPod, &namedPort2EgressEps)
						}

					}
					egressRule.DstIPBlocks = append(egressRule.DstIPBlocks, npc.evalIPBlockPeer(peer)...)
				}
			}

			egressRule.Ports = make([]controller.ProtocolAndPort, 0)
			egressRule.NamedPorts = make([]controller.EndPoints, 0)
			// If this field is empty or missing in the spec, this rule matches all ports
			if len(specEgressRule.Ports) == 0 {
				egressRule.MatchAllPorts = true
			} else {
				egressRule.MatchAllPorts = false
				egressRule.Ports, egressRule.NamedPorts = npc.processNetworkPolicyPorts(specEgressRule.Ports, namedPort2EgressEps)
			}

			newPolicy.EgressRules = append(newPolicy.EgressRules, egressRule)
		}
		NetworkPolicies = append(NetworkPolicies, newPolicy)
	}

	return &NetworkPolicies, nil
}

//evalPodPeer list all specified pod according policy for peers
func (npc *NetworkPolicyController) evalPodPeer(policy *networking.NetworkPolicy, peer networking.NetworkPolicyPeer) ([]*api.Pod, error) {

	var matchingPods []*api.Pod
	matchingPods = make([]*api.Pod, 0)
	var err error
	// spec can have both PodSelector AND NamespaceSelector
	if peer.NamespaceSelector != nil {
		namespaces, err := npc.ListNamespaceByLabels(peer.NamespaceSelector.MatchLabels)
		if err != nil {
			return nil, errors.New("Failed to build network policies info due to " + err.Error())
		}

		var podSelectorLabels map[string]string
		if peer.PodSelector != nil {
			podSelectorLabels = peer.PodSelector.MatchLabels
		}
		for _, namespace := range namespaces {
			namespacePods, err := npc.ListPodsByNamespaceAndLabels(namespace.Name, podSelectorLabels)
			if err != nil {
				return nil, errors.New("Failed to build network policies info due to " + err.Error())
			}
			matchingPods = append(matchingPods, namespacePods...)
		}
	} else if peer.PodSelector != nil {
		matchingPods, err = npc.ListPodsByNamespaceAndLabels(policy.Namespace, peer.PodSelector.MatchLabels)
	}

	return matchingPods, err
}

//ListPodsByNamespaceAndLabels list all pods by namespace & label
func (npc *NetworkPolicyController) ListPodsByNamespaceAndLabels(namespace string, labelsToMatch labels.Set) (ret []*api.Pod, err error) {
	return npc.dataInfr.ListPodsByNamespace(namespace, labelsToMatch)
}

//ListNamespaceByLabels list namespace by label
func (npc *NetworkPolicyController) ListNamespaceByLabels(set labels.Set) ([]*api.Namespace, error) {
	return npc.dataInfr.ListNamespaces(set)
}

func (npc *NetworkPolicyController) evalIPBlockPeer(peer networking.NetworkPolicyPeer) [][]string {
	ipBlock := make([][]string, 0)
	if peer.PodSelector == nil && peer.NamespaceSelector == nil && peer.IPBlock != nil {
		if cidr := peer.IPBlock.CIDR; strings.HasSuffix(cidr, "/0") {
			ipBlock = append(ipBlock, []string{"0.0.0.0/1", iptables.OptionTimeout, "0"}, []string{"128.0.0.0/1", iptables.OptionTimeout, "0"})
		} else {
			ipBlock = append(ipBlock, []string{cidr, iptables.OptionTimeout, "0"})
		}
		for _, except := range peer.IPBlock.Except {
			if strings.HasSuffix(except, "/0") {
				ipBlock = append(ipBlock, []string{"0.0.0.0/1", iptables.OptionTimeout, "0", iptables.OptionNoMatch}, []string{"128.0.0.0/1", iptables.OptionTimeout, "0", iptables.OptionNoMatch})
			} else {
				ipBlock = append(ipBlock, []string{except, iptables.OptionTimeout, "0", iptables.OptionNoMatch})
			}
		}
	}
	return ipBlock
}

func (npc *NetworkPolicyController) grabNamedPortFromPod(pod *api.Pod, namedPort2eps *controller.NamedPort2eps) {
	if pod == nil || namedPort2eps == nil {
		return
	}
	for k := range pod.Spec.Containers {
		for _, port := range pod.Spec.Containers[k].Ports {
			name := port.Name
			protocol := string(port.Protocol)
			containerPort := strconv.Itoa(int(port.ContainerPort))

			if (*namedPort2eps)[name] == nil {
				(*namedPort2eps)[name] = make(controller.Protocol2eps)
			}
			if (*namedPort2eps)[name][protocol] == nil {
				(*namedPort2eps)[name][protocol] = make(controller.NumericPort2eps)
			}
			if eps, ok := (*namedPort2eps)[name][protocol][containerPort]; !ok {
				(*namedPort2eps)[name][protocol][containerPort] = &controller.EndPoints{
					IPs:             []string{pod.Status.PodIP},
					ProtocolAndPort: controller.ProtocolAndPort{Port: containerPort, Protocol: protocol},
				}
			} else {
				eps.IPs = append(eps.IPs, pod.Status.PodIP)
			}
		}
	}
}

func podFirewallChainName(namespace, podName string, version string) string {
	hash := sha256.Sum256([]byte(namespace + podName + version))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubePodFirewallChainPrefix + encoded[:16]
}

func networkPolicyChainName(namespace, policyName string, version string) string {
	hash := sha256.Sum256([]byte(namespace + policyName + version))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeNetworkPolicyChainPrefix + encoded[:16]
}

func policySourcePodIPSetName(namespace, policyName string) string {
	hash := sha256.Sum256([]byte(namespace + policyName))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeSourceIPSetPrefix + encoded[:16]
}

func policyDestinationPodIPSetName(namespace, policyName string) string {
	hash := sha256.Sum256([]byte(namespace + policyName))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeDestinationIPSetPrefix + encoded[:16]
}

func policyIndexedSourcePodIPSetName(namespace, policyName string, ingressRuleNo int) string {
	hash := sha256.Sum256([]byte(namespace + policyName + "ingressrule" + strconv.Itoa(ingressRuleNo) + "pod"))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeSourceIPSetPrefix + encoded[:16]
}

func policyIndexedDestinationPodIPSetName(namespace, policyName string, egressRuleNo int) string {
	hash := sha256.Sum256([]byte(namespace + policyName + "egressrule" + strconv.Itoa(egressRuleNo) + "pod"))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeDestinationIPSetPrefix + encoded[:16]
}

func policyIndexedSourceIPBlockIPSetName(namespace, policyName string, ingressRuleNo int) string {
	hash := sha256.Sum256([]byte(namespace + policyName + "ingressrule" + strconv.Itoa(ingressRuleNo) + "ipblock"))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeSourceIPSetPrefix + encoded[:16]
}

func policyIndexedDestinationIPBlockIPSetName(namespace, policyName string, egressRuleNo int) string {
	hash := sha256.Sum256([]byte(namespace + policyName + "egressrule" + strconv.Itoa(egressRuleNo) + "ipblock"))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeDestinationIPSetPrefix + encoded[:16]
}

func policyIndexedIngressNamedPortIPSetName(namespace, policyName string, ingressRuleNo, namedPortNo int) string {
	hash := sha256.Sum256([]byte(namespace + policyName + "ingressrule" + strconv.Itoa(ingressRuleNo) + strconv.Itoa(namedPortNo) + "namedport"))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeDestinationIPSetPrefix + encoded[:16]
}

func policyIndexedEgressNamedPortIPSetName(namespace, policyName string, egressRuleNo, namedPortNo int) string {
	hash := sha256.Sum256([]byte(namespace + policyName + "egressrule" + strconv.Itoa(egressRuleNo) + strconv.Itoa(namedPortNo) + "namedport"))
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return controller.KubeDestinationIPSetPrefix + encoded[:16]
}

// Cleanup cleanup configurations done
func (npc *NetworkPolicyController) Cleanup() {

	blog.Info("Cleaning up iptables configuration permanently done by kube-router")

	iptablesCmdHandler, err := iptables.New()
	if err != nil {
		blog.Errorf("Failed to initialize iptables executor: %s", err.Error())
	}

	// delete jump rules in FORWARD chain to pod specific firewall chain
	forwardChainRules, err := iptablesCmdHandler.List("filter", "FORWARD")
	if err != nil {
		blog.Errorf("Failed to delete iptables rules as part of cleanup")
		return
	}

	// TODO: need a better way to delte rule with out using number
	var realRuleNo int
	for i, rule := range forwardChainRules {
		if strings.Contains(rule, controller.KubePodFirewallChainPrefix) {
			err = iptablesCmdHandler.Delete("filter", "FORWARD", strconv.Itoa(i-realRuleNo))
			realRuleNo++
		}
	}

	// delete jump rules in OUTPUT chain to pod specific firewall chain
	forwardChainRules, err = iptablesCmdHandler.List("filter", "OUTPUT")
	if err != nil {
		blog.Errorf("Failed to delete iptables rules as part of cleanup")
		return
	}

	// TODO: need a better way to delte rule with out using number
	realRuleNo = 0
	for i, rule := range forwardChainRules {
		if strings.Contains(rule, controller.KubePodFirewallChainPrefix) {
			err = iptablesCmdHandler.Delete("filter", "OUTPUT", strconv.Itoa(i-realRuleNo))
			realRuleNo++
		}
	}

	// flush and delete pod specific firewall chain
	chains, err := iptablesCmdHandler.ListChains("filter")
	for _, chain := range chains {
		if strings.HasPrefix(chain, controller.KubePodFirewallChainPrefix) {
			err = iptablesCmdHandler.ClearChain("filter", chain)
			if err != nil {
				blog.Errorf("Failed to cleanup iptables rules: " + err.Error())
				return
			}
			err = iptablesCmdHandler.DeleteChain("filter", chain)
			if err != nil {
				blog.Errorf("Failed to cleanup iptables rules: " + err.Error())
				return
			}
		}
	}

	// flush and delete per network policy specific chain
	chains, err = iptablesCmdHandler.ListChains("filter")
	for _, chain := range chains {
		if strings.HasPrefix(chain, controller.KubeNetworkPolicyChainPrefix) {
			err = iptablesCmdHandler.ClearChain("filter", chain)
			if err != nil {
				blog.Errorf("Failed to cleanup iptables rules: " + err.Error())
				return
			}
			err = iptablesCmdHandler.DeleteChain("filter", chain)
			if err != nil {
				blog.Errorf("Failed to cleanup iptables rules: " + err.Error())
				return
			}
		}
	}

	// delete all ipsets
	ipset, err := iptables.NewIPSet(false)
	if err != nil {
		blog.Errorf("Failed to clean up ipsets: " + err.Error())
	}
	err = ipset.Save()
	if err != nil {
		blog.Errorf("Failed to clean up ipsets: " + err.Error())
	}
	err = ipset.DestroyAllWithin()
	if err != nil {
		blog.Errorf("Failed to clean up ipsets: " + err.Error())
	}
	blog.Infof("Successfully cleaned the iptables configuration done by kube-router")
}

func (npc *NetworkPolicyController) newPodEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			npc.OnPodUpdate(obj)

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPoObj := newObj.(*api.Pod)
			oldPoObj := oldObj.(*api.Pod)
			if newPoObj.Status.Phase != oldPoObj.Status.Phase || newPoObj.Status.PodIP != oldPoObj.Status.PodIP {
				// for the network policies, we are only interested in pod status phase change or IP change
				npc.OnPodUpdate(newObj)
			}
		},
		DeleteFunc: func(obj interface{}) {
			npc.OnPodUpdate(obj)
		},
	}
}

func (npc *NetworkPolicyController) newNamespaceEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			npc.OnNamespaceUpdate(obj)

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			npc.OnNamespaceUpdate(newObj)

		},
		DeleteFunc: func(obj interface{}) {
			npc.OnNamespaceUpdate(obj)

		},
	}
}

func (npc *NetworkPolicyController) newNetworkPolicyEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			npc.OnNetworkPolicyUpdate(obj)

		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			npc.OnNetworkPolicyUpdate(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			npc.OnNetworkPolicyUpdate(obj)

		},
	}
}

// GetPodEventHandler get pod event handler
func (npc *NetworkPolicyController) GetPodEventHandler() cache.ResourceEventHandler {
	return npc.PodEventHandler
}

// GetNamespaceEventHandler get namespace event handler
func (npc *NetworkPolicyController) GetNamespaceEventHandler() cache.ResourceEventHandler {
	return npc.NamespaceEventHandler
}

// GetNetworkPolicyEventHandler get network policy event handler
func (npc *NetworkPolicyController) GetNetworkPolicyEventHandler() cache.ResourceEventHandler {
	return npc.NetworkPolicyEventHandler
}

// NewNetworkPolicyController returns new NetworkPolicyController object
// add data informer for pod, namepsace and network policy discovery
// add iptables sync error metric
func NewNetworkPolicyController(
	clientset kubernetes.Interface,
	informer datainformer.Interface,
	config *options.NetworkPolicyOption) (controller.Controller, error) {
	npc := NetworkPolicyController{}
	npc.dataInfr = informer

	//Register the metrics for this controller
	prometheus.MustRegister(metrics.ControllerIPTablesSyncTime)
	prometheus.MustRegister(metrics.ControllerPolicyChainsSyncTime)
	prometheus.MustRegister(metrics.ControllerIPTablesSyncError)

	npc.syncPeriod = time.Duration(config.IPTableSyncPeriod) * time.Second

	nodeIP, err := controller.GetNetworkInterfaceIP(config.NetworkInterface)
	if err != nil {
		return nil, err
	}
	npc.nodeIP = nodeIP

	ipset, err := iptables.NewIPSet(false)
	if err != nil {
		return nil, err
	}
	err = ipset.Save()
	if err != nil {
		return nil, err
	}
	npc.ipSetHandler = ipset

	npc.PodEventHandler = npc.newPodEventHandler()
	npc.NamespaceEventHandler = npc.newNamespaceEventHandler()
	npc.NetworkPolicyEventHandler = npc.newNetworkPolicyEventHandler()

	return &npc, nil
}
