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
	"strings"

	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// MappingConverter mapping generator
type MappingConverter struct {
	cli              client.Client
	lbIDs            []string
	ingressName      string
	ingressNamespace string
	mapping          *networkextensionv1.IngressPortMapping
}

// NewMappingConverter create mapping generator
func NewMappingConverter(
	cli client.Client,
	lbIDs []string, ingressName, ingressNamespace string,
	mapping *networkextensionv1.IngressPortMapping) *MappingConverter {

	return &MappingConverter{
		cli:              cli,
		lbIDs:            lbIDs,
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

	var retListeners []*networkextensionv1.Listener
	if mg.mapping.StartIndex == 0 && mg.mapping.EndIndex == 0 {
		for _, lbID := range mg.lbIDs {
			for index, pod := range pods {
				startPort := mg.mapping.StartPort + index*segmentLength
				endPort := mg.mapping.StartPort + (index+1)*segmentLength - 1
				listener := mg.generateListener(startPort, endPort, pod)
				retListeners = append(retListeners, listener)
			}
		}
		return retlisteners, nil
	}
	for _, lbID := range mg.lbIDs {
		for i := mg.mapping.StartIndex; i < mg.mapping.EndIndex; i++ {
			startPort := mg.mapping.StartPort + i*segmentLength
			endPort := mg.mapping.StartPort + (i+1)*segmentLength - 1
			var listener *networkextensionv1.Listener
			pod, ok := pods[i]
			if ok {
				listener = mg.generateListener(startPort, endPort, pod)
			} else {
				listener = mg.generateListener(startPort, endPort, nil)
			}
			retListeners = append(retListeners, listener)
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
			blog.Errorf("get statefulset %s/%s failed, err %s", err.Error())
			return nil, fmt.Errorf("get statefulset %s/%s failed, err %s", err.Error())
		}
		podLabelSelector, err = k8smetav1.LabelSelectorAsSelector(sts.Spec.Selector)
		if err != nil {
			blog.Errorf("convert labelselector %+v to selector failed, err %s",
				sts.Spec.Selector, err.Error())
			return nil, fmt.Errorf("convert labelselector %+v to selector failed, err %s",
				sts.Spec.Selector, err.Error())
		}

	case networkextensionv1.WorkloadKindGameStatefulset:
		blog.Errorf("unimplemented workload kind %s", workloadkind)
		return nil, fmt.Errorf("unimplemented workload kind %s", workloadkind)

	default:
		blog.Errorf("unsupported workload kind %s", workloadkind)
		return nil, fmt.Errorf("unsupported workload kind %s", workloadkind)
	}

	podList := k8scorev1.PodList{}
	err := mg.cli.List(context.TODO(), podList, client.MatchingLabelsSelector{podLabelSelector},
		client.ListOptions{Namespace: ns})
	if err != nil {
		blog.Errorf("list pods with selector %+v failed, err %s", podLabelSelector, err.Error())
		return nil, fmt.Errorf("list pods with selector %+v failed, err %s", podLabelSelector, err.Error())
	}
	retPods := make([int]*k8scorev1.Pod)
	for index, pod := range podList.Items {
		podIndex, err := GetPodIndex(pod.GetName())
		if err != nil {
			blog.Errorf("get pod %s index failed, err %s", pod.GetName())
			return nil, fmt.Errorf("get pod %s index failed, err %s", pod.GetName())
		}
		retPods[podIndex] = &podList.Items[index]
	}
	return retPods, nil
}

func (mg *MappingConverter) generateListener(
	lbID string, startPort, endPort int, pod *k8scorev1.Pod) networkextensionv1.Listener {

	li := networkextensionv1.Listener{}
	var listenerName string
	if endPort <= 0 {
		listenerName = GetListenerName(lbID, startPort)
	} else {
		listenerName = GetSegmentListenerName(lbID, startPort, endPort)
	}
	li.SetName(listenerName)
	li.SetNamespace(mg.ingressNamespace)
	li.SetLabels(map[string]string{
		mg.ingressName:      networkextensionv1.LabelValueForIngressName,
		mg.ingressNamespace: networkextensionv1.LabelValueForIngressNamespace,
		networkextensionv1.LabelKeyForLoadbalanceID: lbID,
	})
	li.Spec.Port = startPort
	li.Spec.EndPort = endPort
	li.Spec.Protocol = mg.mapping.Protocol
	li.Spec.LoadbalancerID = lbID

	targetGroup := &networkextensionv1.ListenerTargetGroup{
		TargetGroupProtocol: mg.mapping.Protocol,
	}
	if pod == nil || len(pod.Status.PodIP) == 0 {
		return li
	}
	if startPort == endPort {
		targetGroup.Backends = append(targetGroup.Backends, &networkextensionv1.ListenerBackend{
			IP:     pod.Status.PodIP,
			Port:   startPort,
			Weight: networkextensionv1.DefaultWeight,
		})
	} else {
		targetGroup.Backends = append(targetGroup.Backends, &networkextensionv1.ListenerBackend{
			IP:     pod.Status.PodIP,
			Port:   0,
			Weight: networkextensionv1.DefaultWeight,
		})
	}
	li.Spec.TargetGroup = targetGroup
	return li
}
