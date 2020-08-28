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
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/constant"
)

// MappingConverter mapping generator
type MappingConverter struct {
	cli              client.Client
	lbs              []*cloud.LoadBalanceObject
	ingressName      string
	ingressNamespace string
	mapping          *networkextensionv1.IngressPortMapping
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

// DoConvert do convert action
func (mg *MappingConverter) DoConvert() ([]networkextensionv1.Listener, error) {
	pods, err := mg.getWorkloadPodMap(mg.mapping.WorkloadKind,
		mg.mapping.WorkloadName, mg.mapping.WorkloadNamespace)
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
			pod := pods[i]
			listeners := mg.generateSegmentListeners(lb.Region, lb.LbID, startPort, endPort, pod)
			retListeners = append(retListeners, listeners...)
		}
	}
	return retListeners, nil
}

// get workload pods map
func (mg *MappingConverter) getWorkloadPodMap(workloadKind, workloadName, workloadNamespace string) (
	map[int]*k8scorev1.Pod, error) {

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
		blog.Errorf("unimplemented workload kind %s", workloadKind)
		return nil, fmt.Errorf("unimplemented workload kind %s", workloadKind)

	default:
		blog.Errorf("unsupported workload kind %s", workloadKind)
		return nil, fmt.Errorf("unsupported workload kind %s", workloadKind)
	}

	podList := &k8scorev1.PodList{}
	err := mg.cli.List(context.TODO(), podList, client.MatchingLabelsSelector{Selector: podLabelSelector},
		&client.ListOptions{Namespace: workloadNamespace})
	if err != nil {
		blog.Errorf("list pods with selector %+v failed, err %s", podLabelSelector, err.Error())
		return nil, fmt.Errorf("list pods with selector %+v failed, err %s", podLabelSelector, err.Error())
	}
	retPods := make(map[int]*k8scorev1.Pod)
	for index, pod := range podList.Items {
		podIndex, err := GetPodIndex(pod.GetName())
		if err != nil {
			blog.Errorf("get pod %s index failed, err %s", pod.GetName())
			return nil, fmt.Errorf("get pod %s index failed, err %s", pod.GetName(), err.Error())
		}
		retPods[podIndex] = &podList.Items[index]
	}
	return retPods, nil
}

func (mg *MappingConverter) generateSegmentListeners(
	region, lbID string, startPort, endPort int, pod *k8scorev1.Pod) []networkextensionv1.Listener {

	var retListeners []networkextensionv1.Listener
	if !mg.mapping.IgnoreSegment && mg.mapping.SegmentLength != 0 && mg.mapping.SegmentLength != 1 {
		listener := mg.generateListener(region, lbID, startPort, endPort, pod)
		retListeners = append(retListeners, listener)
		return retListeners
	}
	// if not use segment mapping feature
	for j := startPort; j <= endPort; j++ {
		listener := mg.generateListener(region, lbID, j, 0, pod)
		retListeners = append(retListeners, listener)
	}
	return retListeners
}

func (mg *MappingConverter) generateListener(
	region, lbID string, startPort, endPort int, pod *k8scorev1.Pod) networkextensionv1.Listener {

	segLabelValue := networkextensionv1.LabelValueTrue
	if endPort == 0 {
		segLabelValue = networkextensionv1.LabelValueFalse
	}

	li := networkextensionv1.Listener{}
	var listenerName string
	listenerName = GetSegmentListenerName(lbID, startPort, endPort)
	li.SetName(listenerName)
	li.SetNamespace(mg.ingressNamespace)
	li.SetLabels(map[string]string{
		mg.ingressName:      networkextensionv1.LabelValueForIngressName,
		mg.ingressNamespace: networkextensionv1.LabelValueForIngressNamespace,
		networkextensionv1.LabelKeyForIsSegmentListener: segLabelValue,
		networkextensionv1.LabelKeyForLoadbalanceID:     lbID,
		networkextensionv1.LabelKeyForLoadbalanceRegion: region,
	})
	li.Finalizers = append(li.Finalizers, constant.FinalizerNameBcsIngressController)
	li.Spec.Port = startPort
	li.Spec.EndPort = endPort
	li.Spec.Protocol = mg.mapping.Protocol
	li.Spec.LoadbalancerID = lbID

	targetGroup := &networkextensionv1.ListenerTargetGroup{
		TargetGroupProtocol: mg.mapping.Protocol,
	}
	if pod == nil || len(pod.Status.PodIP) == 0 {
		li.Spec.TargetGroup = targetGroup
		return li
	}
	targetGroup.Backends = append(targetGroup.Backends, networkextensionv1.ListenerBackend{
		IP:     pod.Status.PodIP,
		Port:   startPort,
		Weight: networkextensionv1.DefaultWeight,
	})
	li.Spec.TargetGroup = targetGroup
	return li
}
