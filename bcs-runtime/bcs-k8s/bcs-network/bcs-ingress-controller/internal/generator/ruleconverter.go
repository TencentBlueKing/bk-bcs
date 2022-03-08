/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// RuleConverter rule converter
type RuleConverter struct {
	// k8s crd client
	cli client.Client
	// loadbalances for rule conveter
	lbs []*cloud.LoadBalanceObject
	// the name of ingress that rule belongs to
	ingressName string
	// the namespace of ingress that rule belongs to
	ingressNamespace string
	// rule info
	rule *networkextensionv1.IngressRule
	// if true, ingress can only select service, endpoint and workload in the same namespace
	isNamespaced bool
	// if true, allow tcp listener and udp listener use same port
	isTCPUDPPortReuse bool
}

// NewRuleConverter create rule converter
func NewRuleConverter(
	cli client.Client,
	lbs []*cloud.LoadBalanceObject,
	ingressName string,
	ingressNamespace string,
	rule *networkextensionv1.IngressRule) *RuleConverter {

	return &RuleConverter{
		cli:               cli,
		lbs:               lbs,
		ingressName:       ingressName,
		ingressNamespace:  ingressNamespace,
		rule:              rule,
		isNamespaced:      false,
		isTCPUDPPortReuse: false,
	}
}

// SetNamespaced set namespaced flag
func (rc *RuleConverter) SetNamespaced(isNamespaced bool) {
	rc.isNamespaced = isNamespaced
}

// SetTCPUDPPortReuse set isTCPUDPPortReuse flag
func (rc *RuleConverter) SetTCPUDPPortReuse(isTCPUDPPortReuse bool) {
	rc.isTCPUDPPortReuse = isTCPUDPPortReuse
}

// DoConvert do convert action
func (rc *RuleConverter) DoConvert() ([]networkextensionv1.Listener, error) {
	var retListeners []networkextensionv1.Listener
	switch strings.ToLower(rc.rule.Protocol) {
	case networkextensionv1.ProtocolHTTP, networkextensionv1.ProtocolHTTPS:
		for _, lb := range rc.lbs {
			listener, err := rc.generate7LayerListener(lb.Region, lb.LbID)
			if err != nil {
				return nil, err
			}
			retListeners = append(retListeners, *listener)
		}
	case networkextensionv1.ProtocolTCP, networkextensionv1.ProtocolUDP:
		for _, lb := range rc.lbs {
			listener, err := rc.generate4LayerListener(lb.Region, lb.LbID)
			if err != nil {
				return nil, err
			}
			retListeners = append(retListeners, *listener)
		}
	default:
		blog.Errorf("invalid protocol %s", rc.rule.Protocol)
		return nil, fmt.Errorf("invalid protocol %s", rc.rule.Protocol)
	}
	return retListeners, nil
}

// generate 7 layer listener by rule info
func (rc *RuleConverter) generate7LayerListener(region, lbID string) (*networkextensionv1.Listener, error) {
	li := &networkextensionv1.Listener{}
	li.SetName(GetListenerName(lbID, rc.rule.Port))
	li.SetNamespace(rc.ingressNamespace)
	// set ingress name in labels
	// the ingress name in labels is used for checking conficts
	li.SetLabels(map[string]string{
		rc.ingressName: networkextensionv1.LabelValueForIngressName,
		networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
		networkextensionv1.LabelKeyForLoadbalanceID:     lbID,
		networkextensionv1.LabelKeyForLoadbalanceRegion: region,
	})
	li.Finalizers = append(li.Finalizers, constant.FinalizerNameBcsIngressController)
	li.Spec.Port = rc.rule.Port
	li.Spec.Protocol = rc.rule.Protocol
	li.Spec.LoadbalancerID = lbID
	if rc.rule.Certificate != nil {
		li.Spec.Certificate = rc.rule.Certificate
	}

	listenerRules, err := rc.generateListenerRule(rc.rule.Routes)
	if err != nil {
		return nil, err
	}
	li.Spec.Rules = listenerRules
	return li, nil
}

// generate rule of 7 layer listener by rule info
func (rc *RuleConverter) generateListenerRule(l7Routes []networkextensionv1.Layer7Route) (
	[]networkextensionv1.ListenerRule, error) {

	var retListenerRules []networkextensionv1.ListenerRule
	for _, l7Route := range l7Routes {
		liRule := networkextensionv1.ListenerRule{}
		liRule.Domain = l7Route.Domain
		liRule.Path = l7Route.Path
		liRule.ListenerAttribute = l7Route.ListenerAttribute
		targetGroup, err := rc.generateTargetGroup(rc.rule.Protocol, l7Route.Services)
		if err != nil {
			return nil, err
		}
		liRule.TargetGroup = targetGroup
		retListenerRules = append(retListenerRules, liRule)
	}
	sort.Sort(networkextensionv1.ListenerRuleList(retListenerRules))
	return retListenerRules, nil
}

// generate 4 layer listener by rule info
func (rc *RuleConverter) generate4LayerListener(region, lbID string) (*networkextensionv1.Listener, error) {
	li := &networkextensionv1.Listener{}

	if rc.isTCPUDPPortReuse {
		li.SetName(GetListenerNameWithProtocol(lbID, rc.rule.Protocol, rc.rule.Port))
	} else {
		li.SetName(GetListenerName(lbID, rc.rule.Port))
	}
	li.SetNamespace(rc.ingressNamespace)
	// set ingress name in labels
	// the ingress name in labels is used for checking conficts
	li.SetLabels(map[string]string{
		rc.ingressName: networkextensionv1.LabelValueForIngressName,
		networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
		networkextensionv1.LabelKeyForLoadbalanceID:     lbID,
		networkextensionv1.LabelKeyForLoadbalanceRegion: region,
	})
	li.Finalizers = append(li.Finalizers, constant.FinalizerNameBcsIngressController)
	li.Spec.Port = rc.rule.Port
	li.Spec.Protocol = rc.rule.Protocol
	li.Spec.LoadbalancerID = lbID
	if rc.rule.ListenerAttribute != nil {
		li.Spec.ListenerAttribute = rc.rule.ListenerAttribute
	}
	if rc.rule.Certificate != nil {
		li.Spec.Certificate = rc.rule.Certificate
	}

	targetGroup, err := rc.generateTargetGroup(rc.rule.Protocol, rc.rule.Services)
	if err != nil {
		return nil, err
	}
	li.Spec.TargetGroup = targetGroup
	return li, nil
}

// generate target group info
func (rc *RuleConverter) generateTargetGroup(protocol string, routes []networkextensionv1.ServiceRoute) (
	*networkextensionv1.ListenerTargetGroup, error) {

	var retBackends []networkextensionv1.ListenerBackend
	for _, route := range routes {
		backends, err := rc.generateServiceBackendList(&route)
		if err != nil {
			return nil, err
		}
		retBackends = mergeBackendList(retBackends, backends)
	}
	sort.Sort(networkextensionv1.ListenerBackendList(retBackends))
	return &networkextensionv1.ListenerTargetGroup{
		TargetGroupProtocol: protocol,
		Backends:            retBackends,
	}, nil
}

// generate service backend list
func (rc *RuleConverter) generateServiceBackendList(svcRoute *networkextensionv1.ServiceRoute) (
	[]networkextensionv1.ListenerBackend, error) {

	// set namespace when namespaced flag is set
	svcNamespace := svcRoute.ServiceNamespace
	if rc.isNamespaced {
		svcNamespace = rc.ingressNamespace
	}

	// get service
	svc := &k8scorev1.Service{}
	err := rc.cli.Get(context.TODO(), k8stypes.NamespacedName{
		Namespace: svcNamespace,
		Name:      svcRoute.ServiceName,
	}, svc)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get Service %s/%s failed, err %s",
			svcRoute.ServiceName, svcNamespace, err.Error())
	}
	var svcPort *k8scorev1.ServicePort
	for _, port := range svc.Spec.Ports {
		if port.Port == int32(svcRoute.ServicePort) {
			svcPort = &port
			break
		}
	}
	if svcPort == nil {
		blog.Warnf("port %d is not found in service %s/%s",
			svcRoute.ServicePort, svcRoute.ServiceName, svcNamespace)
		return nil, nil
	}

	// subset subset only takes effect when directly connected
	// when directly connected
	// * if no subset, use pod list as backends
	// * if there are subsets, use pod from subset, and do merge
	if svcRoute.IsDirectConnect {
		// to pod directly and no subset
		if len(svcRoute.Subsets) == 0 {
			backends, err := rc.getServiceBackendsFromPods(
				svcNamespace, svc.Spec.Selector, svcPort, svcRoute.GetWeight())
			if err != nil {
				return nil, err
			}
			return backends, nil
		}
		var retBackends []networkextensionv1.ListenerBackend
		// to pod directly and have subset
		for _, subset := range svcRoute.Subsets {
			subsetBackends, err := rc.getSubsetBackends(svc, svcPort, subset)
			if err != nil {
				return nil, err
			}
			retBackends = mergeBackendList(retBackends, subsetBackends)
		}
		return retBackends, nil
	}
	// to node port
	retBackends, err := rc.getNodePortBackends(svc, svcPort, svcRoute.GetWeight())
	if err != nil {
		return nil, err
	}
	return retBackends, nil
}

func mergeBackendList(
	existedList, newList []networkextensionv1.ListenerBackend) []networkextensionv1.ListenerBackend {

	tmpMap := make(map[string]networkextensionv1.ListenerBackend)
	for _, backend := range existedList {
		tmpMap[backend.IP+strconv.Itoa(backend.Port)] = backend
	}
	for _, backend := range newList {
		if _, ok := tmpMap[backend.IP+strconv.Itoa(backend.Port)]; !ok {
			existedList = append(existedList, backend)
		}
	}
	return existedList
}

// get backends from subset
func (rc *RuleConverter) getSubsetBackends(
	svc *k8scorev1.Service, svcPort *k8scorev1.ServicePort, subset networkextensionv1.IngressSubset) (
	[]networkextensionv1.ListenerBackend, error) {
	labels := make(map[string]string)
	for k, v := range svc.Spec.Selector {
		labels[k] = v
	}
	for k, v := range subset.LabelSelector {
		labels[k] = v
	}
	return rc.getServiceBackendsFromPods(svc.GetNamespace(), labels, svcPort, subset.GetWeight())
}

// get backends from pods
func (rc *RuleConverter) getServiceBackendsFromPods(
	ns string, selectorMap map[string]string,
	svcPort *k8scorev1.ServicePort, weight int) (
	[]networkextensionv1.ListenerBackend, error) {

	podList, err := rc.getPodsByLabels(ns, selectorMap)
	if err != nil {
		return nil, err
	}

	var retBackends []networkextensionv1.ListenerBackend
	for _, pod := range podList {
		if len(pod.Status.PodIP) == 0 {
			continue
		}
		backendWeight := rc.getPodWeight(pod, weight)
		if pod.DeletionTimestamp != nil {
			backendWeight = 0
		}
		// if container is unready, client should not visit this pod
		if pod.Status.Phase == k8scorev1.PodRunning {
			ready := true
			for _, c := range pod.Status.Conditions {
				if c.Type == k8scorev1.ContainersReady && c.Status != k8scorev1.ConditionTrue {
					ready = false
					break
				}
			}
			if !ready {
				backendWeight = 0
			}
			blog.Infof("pod name %s namespace %s is running, backendWeight: %d", pod.Name, pod.Namespace, backendWeight)
		}

		found := false
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				if (port.ContainerPort == int32(svcPort.TargetPort.IntValue()) && port.Protocol == svcPort.Protocol) ||
					(port.Name == svcPort.TargetPort.String() && port.Protocol == svcPort.Protocol) {
					retBackends = append(retBackends, networkextensionv1.ListenerBackend{
						IP:     pod.Status.PodIP,
						Port:   int(port.ContainerPort),
						Weight: backendWeight,
					})
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}
	return retBackends, nil
}

// use node port as clb backends
func (rc *RuleConverter) getNodePortBackends(
	svc *k8scorev1.Service, svcPort *k8scorev1.ServicePort, weight int) (
	[]networkextensionv1.ListenerBackend, error) {

	if svcPort.NodePort <= 0 {
		blog.Warnf("get no node port of service %s/%s 's port %+v",
			svc.GetNamespace(), svc.GetName(), svcPort)
		return nil, nil
	}

	pods, err := rc.getPodsByLabels(svc.GetNamespace(), svc.Spec.Selector)
	if err != nil {
		return nil, err
	}

	var retBackends []networkextensionv1.ListenerBackend
	backendMap := make(map[string]networkextensionv1.ListenerBackend)
	for _, pod := range pods {
		if len(pod.Status.HostIP) == 0 {
			continue
		}
		if _, ok := backendMap[pod.Status.HostIP+strconv.Itoa(int(svcPort.NodePort))]; ok {
			continue
		}
		newBackend := networkextensionv1.ListenerBackend{
			IP:     pod.Status.HostIP,
			Port:   int(svcPort.NodePort),
			Weight: weight,
		}
		newBackend.Weight = rc.getPodWeight(pod, weight)
		backendMap[pod.Status.HostIP+strconv.Itoa(int(svcPort.NodePort))] = newBackend
		retBackends = append(retBackends, newBackend)
	}
	return retBackends, nil
}

// get pod by labels
func (rc *RuleConverter) getPodsByLabels(ns string, labels map[string]string) ([]*k8scorev1.Pod, error) {
	if len(labels) == 0 {
		return nil, nil
	}
	if rc.isNamespaced {
		ns = rc.ingressNamespace
	}
	podList := &k8scorev1.PodList{}
	err := rc.cli.List(context.Background(), podList, client.MatchingLabels(labels), &client.ListOptions{Namespace: ns})
	if err != nil {
		blog.Errorf("list pod list failed by labels %+v and ns %s, err %s", labels, ns, err.Error())
		return nil, fmt.Errorf("list pod list failed by labels %+v and ns %s, err %s", labels, ns, err.Error())
	}
	var retPods []*k8scorev1.Pod
	for i := 0; i < len(podList.Items); i++ {
		retPods = append(retPods, &podList.Items[i])
	}
	return retPods, nil
}

// get pod clb-weight from annotations
func (rc *RuleConverter) getPodWeight(pod *k8scorev1.Pod, weight int) int {
	if clbWeightValue, ok := pod.Annotations[networkextensionv1.AnnotationKeyForLoadbalanceWeight]; ok {
		clbWeight, err := strconv.Atoi(clbWeightValue)
		if err != nil {
			blog.Warnf("get pod %s/%s's clb-weight error: %s", pod.Namespace, pod.Name, err.Error())
			return weight
		}
		err = rc.patchPodLBWeightReady(pod)
		if err != nil {
			blog.Warnf("patch pod %s/%s's clb-weight error: %s", pod.Namespace, pod.Name, err.Error())
			return weight
		}
		return clbWeight
	}
	return weight
}

// patch pod annotations for clb weight, if pod lb weight be set, then switch annotation ready to true
func (rc *RuleConverter) patchPodLBWeightReady(pod *k8scorev1.Pod) error {
	if pod.Annotations[networkextensionv1.AnnotationKeyForLoadbalanceWeightReady] == "true" {
		return nil
	}
	patchStruct := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				networkextensionv1.AnnotationKeyForLoadbalanceWeightReady: "true",
			},
		},
	}
	patchData, err := json.Marshal(patchStruct)
	if err != nil {
		return err
	}
	updatePod := &k8scorev1.Pod{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
		},
	}
	return rc.cli.Patch(context.TODO(), updatePod, client.RawPatch(k8stypes.MergePatchType, patchData))
}
