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
	"fmt"
	"strings"

	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sunstruct "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// MappingConverter mapping generator
type MappingConverter struct {
	cli              client.Client
	lbs              []*cloud.LoadBalanceObject
	ingressName      string
	ingressNamespace string
	mapping          *networkextensionv1.IngressPortMapping
	// if true, ingress can only select service, endpoint and workload in the same namespace
	isNamespaced bool
}

// NewMappingConverter create mapping generator
func NewMappingConverter(
	cli client.Client, lbs []*cloud.LoadBalanceObject, ingressName, ingressNamespace string,
	mapping *networkextensionv1.IngressPortMapping) *MappingConverter {

	return &MappingConverter{
		cli:              cli,
		lbs:              lbs,
		ingressName:      ingressName,
		ingressNamespace: ingressNamespace,
		mapping:          mapping,
	}
}

// SetNamespaced set namespaced flag
func (mg *MappingConverter) SetNamespaced(isNamespaced bool) {
	mg.isNamespaced = isNamespaced
}

// get selector by workload
func (mg *MappingConverter) getWorkloadSelector(workloadKind, workloadName, workloadNamespace string) (
	k8slabels.Selector, error) {

	var podLabelSelector k8slabels.Selector
	switch strings.ToLower(workloadKind) {
	case networkextensionv1.WorkloadKindStatefulset:
		sts := &k8sappsv1.StatefulSet{}
		err := mg.cli.Get(context.TODO(), k8stypes.NamespacedName{
			Namespace: workloadNamespace,
			Name:      workloadName,
		}, sts)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, nil
			}
			blog.Errorf("get statefulset %s/%s failed, err %s", workloadName, workloadNamespace, err.Error())
			return nil, fmt.Errorf("get statefulset %s/%s failed, err %s", workloadName, workloadNamespace, err.Error())
		}
		podLabelSelector, err = k8smetav1.LabelSelectorAsSelector(sts.Spec.Selector)
		if err != nil {
			blog.Errorf("convert labelselector %+v to selector failed, err %s",
				sts.Spec.Selector, err.Error())
			return nil, fmt.Errorf("convert labelselector %+v to selector failed, err %s",
				sts.Spec.Selector, err.Error())
		}

	case networkextensionv1.WorkloadKindGameStatefulset:
		gsts := &k8sunstruct.Unstructured{}
		gsts.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "tkex.tencent.com",
			Version: "v1alpha1",
			Kind:    "GameStatefulSet",
		})
		err := mg.cli.Get(context.TODO(), k8stypes.NamespacedName{
			Namespace: workloadNamespace,
			Name:      workloadName,
		}, gsts)

		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, nil
			}
			blog.Errorf("get game statefulset %s/%s failed, err %s",
				workloadName, workloadNamespace, err.Error())
			return nil, fmt.Errorf("get game statefulset %s/%s failed, err %s",
				workloadName, workloadNamespace, err.Error())
		}
		selectorObj := GetSpecLabelSelectorFromMap(gsts.Object)
		if selectorObj == nil {
			blog.Warnf("found no selector in game statefulset %s/%s", workloadName, workloadNamespace)
			return nil, nil
		}
		podLabelSelector, err = k8smetav1.LabelSelectorAsSelector(selectorObj)
		if err != nil {
			blog.Errorf("convert labelselector %+v to selector failed, err %s", selectorObj, err.Error())
			return nil, fmt.Errorf("convert labelselector %+v to selector failed, err %s", selectorObj, err.Error())
		}

	default:
		blog.Errorf("unsupported workload kind %s", workloadKind)
		return nil, fmt.Errorf("unsupported workload kind %s", workloadKind)
	}

	return podLabelSelector, nil
}

// get workload pods map
func (mg *MappingConverter) getWorkloadPodMap(workloadKind, workloadName, workloadNamespace string) (
	map[int]*k8scorev1.Pod, error) {

	podLabelSelector, err := mg.getWorkloadSelector(workloadKind, workloadName, workloadNamespace)
	if err != nil {
		return nil, err
	}
	retPods := make(map[int]*k8scorev1.Pod)

	if podLabelSelector == nil {
		return retPods, nil
	}

	podList := &k8scorev1.PodList{}
	err = mg.cli.List(context.TODO(), podList, client.MatchingLabelsSelector{Selector: podLabelSelector},
		&client.ListOptions{Namespace: workloadNamespace})
	if err != nil {
		blog.Errorf("list pods with selector %+v failed, err %s", podLabelSelector, err.Error())
		return nil, fmt.Errorf("list pods with selector %+v failed, err %s", podLabelSelector, err.Error())
	}

	for index, pod := range podList.Items {
		if !isPodOwner(workloadKind, workloadName, &pod) {
			blog.Warnf("pod %s/%s is not owner by %s/%s",
				pod.GetName(), pod.GetNamespace(), workloadKind, workloadName)
			continue
		}
		podIndex, err := GetPodIndex(pod.GetName())
		if err != nil {
			blog.Errorf("get pod %s index failed, err %s", pod.GetName(), err.Error())
			return nil, fmt.Errorf("get pod %s index failed, err %s", pod.GetName(), err.Error())
		}
		retPods[podIndex] = &podList.Items[index]
	}
	return retPods, nil
}

// DoConvert do convert action
func (mg *MappingConverter) DoConvert() ([]networkextensionv1.Listener, error) {

	// set namespace when namespaced flag is set
	workloadNamespace := mg.mapping.WorkloadNamespace
	if mg.isNamespaced {
		workloadNamespace = mg.ingressNamespace
	}

	pods, err := mg.getWorkloadPodMap(mg.mapping.WorkloadKind,
		mg.mapping.WorkloadName, workloadNamespace)
	if err != nil {
		return nil, err
	}

	segmentLength := mg.mapping.SegmentLength
	if segmentLength == 0 {
		segmentLength = 1
	}

	var retListeners []networkextensionv1.Listener
	for _, lb := range mg.lbs {
		for i := mg.mapping.StartIndex; i < mg.mapping.EndIndex; i++ {
			startPort := mg.mapping.StartPort + i*segmentLength
			endPort := mg.mapping.StartPort + (i+1)*segmentLength - 1

			rsStartPort := startPort
			if mg.mapping.RsStartPort > 0 {
				rsStartPort = mg.mapping.RsStartPort + i*segmentLength
			}

			// if rs port fixed
			if mg.mapping.IsRsPortFixed {
				rsStartPort = mg.mapping.StartPort
				if mg.mapping.RsStartPort > 0 {
					rsStartPort = mg.mapping.RsStartPort
				}
			}

			pod := pods[i]

			// contruct converter for every single segment listener
			newSegConverter := &segmentListenerConverter{
				ingressName:      mg.ingressName,
				ingressNamespace: mg.ingressNamespace,
				protocol:         mg.mapping.Protocol,
				region:           lb.Region,
				lbID:             lb.LbID,
				startPort:        startPort,
				endPort:          endPort,
				rsStartPort:      rsStartPort,
				hostPort:         mg.mapping.HostPort,
				ignoreSegment:    mg.mapping.IgnoreSegment,
				segmentLength:    mg.mapping.SegmentLength,
				pod:              pod,
				listenerAttr:     mg.mapping.ListenerAttribute,
				routes:           mg.mapping.Routes,
				certs:            mg.mapping.Certificate,
			}
			listeners, err := newSegConverter.generateSegmentListener()
			if err != nil {
				return nil, fmt.Errorf("generate segment listeners failed, err %s", err.Error())
			}
			retListeners = append(retListeners, listeners...)
		}
	}
	return retListeners, nil
}

// segmentListenerConverter converter for segment listener
type segmentListenerConverter struct {
	ingressName      string
	ingressNamespace string
	protocol         string
	region           string
	lbID             string
	startPort        int
	endPort          int
	rsStartPort      int
	hostPort         bool
	ignoreSegment    bool
	segmentLength    int
	pod              *k8scorev1.Pod
	listenerAttr     *networkextensionv1.IngressListenerAttribute
	routes           []networkextensionv1.IngressPortMappingLayer7Route
	certs            *networkextensionv1.IngressListenerCertificate
}

func (slc *segmentListenerConverter) generateSegmentListener() ([]networkextensionv1.Listener, error) {
	var retListeners []networkextensionv1.Listener
	if !slc.ignoreSegment && slc.segmentLength != 0 && slc.segmentLength != 1 {
		listener, err := slc.generateListener(slc.startPort, slc.endPort, slc.rsStartPort)
		if err != nil {
			return nil, err
		}
		retListeners = append(retListeners, listener)
		return retListeners, nil
	}
	// if not use segment mapping feature
	for j, rs := slc.startPort, slc.rsStartPort; j <= slc.endPort; j, rs = j+1, rs+1 {
		listener, err := slc.generateListener(j, 0, rs)
		if err != nil {
			return nil, err
		}
		retListeners = append(retListeners, listener)
	}
	return retListeners, nil
}

func (slc *segmentListenerConverter) generateListener(start, end, rsStart int) (networkextensionv1.Listener, error) {
	segLabelValue := networkextensionv1.LabelValueTrue
	if end == 0 {
		segLabelValue = networkextensionv1.LabelValueFalse
	}

	// only tcp and udp support segment feature
	if end != 0 && strings.ToLower(slc.protocol) != "tcp" && strings.ToLower(slc.protocol) != "udp" {
		return networkextensionv1.Listener{}, fmt.Errorf("only tcp and udp support segment feature")
	}

	li := networkextensionv1.Listener{}
	var listenerName string
	listenerName = GetSegmentListenerName(slc.lbID, start, end)
	li.SetName(listenerName)
	li.SetNamespace(slc.ingressNamespace)
	li.SetLabels(map[string]string{
		slc.ingressName: networkextensionv1.LabelValueForIngressName,
		networkextensionv1.LabelKeyForIsSegmentListener: segLabelValue,
		networkextensionv1.LabelKeyForLoadbalanceID:     slc.lbID,
		networkextensionv1.LabelKeyForLoadbalanceRegion: slc.region,
	})
	li.Finalizers = append(li.Finalizers, constant.FinalizerNameBcsIngressController)
	li.Spec.Port = start
	li.Spec.EndPort = end
	li.Spec.Protocol = slc.protocol
	li.Spec.LoadbalancerID = slc.lbID
	li.Spec.ListenerAttribute = slc.listenerAttr

	if slc.certs == nil && len(slc.routes) == 0 {
		li.Spec.TargetGroup = slc.generateListenerTargetGroup(rsStart)
		return li, nil
	}
	if slc.certs != nil {
		li.Spec.Certificate = slc.certs
	}
	li.Spec.Rules = slc.generateListenerRules(rsStart)
	return li, nil
}

func (slc *segmentListenerConverter) generateListenerRules(rsPort int) []networkextensionv1.ListenerRule {
	var retRules []networkextensionv1.ListenerRule
	for _, r := range slc.routes {
		liRule := networkextensionv1.ListenerRule{}
		liRule.Domain = r.Domain
		liRule.Path = r.Path
		liRule.ListenerAttribute = r.ListenerAttribute
		liRule.TargetGroup = slc.generateListenerTargetGroup(rsPort)
		retRules = append(retRules, liRule)
	}
	return retRules
}

func (slc *segmentListenerConverter) generateListenerTargetGroup(rsPort int) *networkextensionv1.ListenerTargetGroup {
	targetGroup := &networkextensionv1.ListenerTargetGroup{
		TargetGroupProtocol: slc.protocol,
	}
	if slc.pod == nil || len(slc.pod.Status.PodIP) == 0 {
		return targetGroup
	}
	backend := networkextensionv1.ListenerBackend{
		IP:     slc.pod.Status.PodIP,
		Port:   rsPort,
		Weight: networkextensionv1.DefaultWeight,
	}
	// if hostPort is specified, use hostPort as backend port
	if hostPort := GetPodHostPortByPort(slc.pod, int32(rsPort)); slc.hostPort &&
		hostPort != 0 {
		backend.IP = slc.pod.Status.HostIP
		backend.Port = int(hostPort)
	}
	targetGroup.Backends = append(targetGroup.Backends, backend)
	return targetGroup
}
