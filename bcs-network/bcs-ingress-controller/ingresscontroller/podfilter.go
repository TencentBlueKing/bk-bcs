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

package ingresscontroller

import (
	"context"

	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/metrics"
)

// PodFilter filter for pod event
type PodFilter struct {
	filterName string
	cli        client.Client
}

// NewPodFilter create pod filter
func NewPodFilter(cli client.Client) *PodFilter {
	return &PodFilter{
		filterName: "pod",
		cli:        cli,
	}
}

func isServiceMatchesPod(svc *k8scorev1.Service, pod *k8scorev1.Pod) bool {
	labelSelector := labels.SelectorFromSet(labels.Set(svc.Spec.Selector))
	return labelSelector.Matches(labels.Set(pod.GetLabels()))
}

func (pf *PodFilter) enqueuePodRelatedIngress(pod *k8scorev1.Pod, q workqueue.RateLimitingInterface) {
	ingresses := pf.findPodIngresses(pod)
	if len(ingresses) == 0 {
		return
	}
	for _, ingress := range ingresses {
		q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      ingress.GetName(),
			Namespace: ingress.GetNamespace(),
		}})
	}
}

func (pf *PodFilter) findPodIngresses(pod *k8scorev1.Pod) []*networkextensionv1.Ingress {
	// find ingresses of pod related services
	svcList := &k8scorev1.ServiceList{}
	err := pf.cli.List(context.TODO(), svcList, &client.ListOptions{Namespace: pod.GetNamespace()})
	if err != nil {
		blog.Warnf("list services in namespace %s failed, err %s", pod.GetNamespace(), err.Error())
		return nil
	}
	ingressList := &networkextensionv1.IngressList{}
	err = pf.cli.List(context.TODO(), ingressList, &client.ListOptions{})
	if err != nil {
		blog.Warnf("list bcs ingresses failed, err %s", err.Error())
		return nil
	}
	var retList []*networkextensionv1.Ingress
	for _, svc := range svcList.Items {
		if isServiceMatchesPod(&svc, pod) {
			retList = append(retList, findIngressesByService(svc.GetName(), svc.GetNamespace(), ingressList)...)
		}
	}

	// find ingresses of pod related workloads
	for _, owner := range pod.GetOwnerReferences() {
		ingresses := findIngressesByWorkload(owner.Kind, owner.Name, pod.GetNamespace(), ingressList)
		if len(ingresses) != 0 {
			retList = append(retList, ingresses...)
		}
	}

	// deduplicate ingresses
	retList = deduplicateIngresses(retList)
	return retList
}

// Create implement EventFilter
func (pf *PodFilter) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeAdd)

	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv create object is not Pod, event %+v", e)
		return
	}
	pf.enqueuePodRelatedIngress(pod, q)
}

// Update implement EventFilter
func (pf *PodFilter) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeUpdate)

	newPod, ok := e.ObjectNew.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv update object is not Pod, event %+v", e)
		return
	}
	pf.enqueuePodRelatedIngress(newPod, q)
}

// Delete implement EventFilter
func (pf *PodFilter) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeDelete)

	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv delete object is not Pod, event %+v", e)
		return
	}
	pf.enqueuePodRelatedIngress(pod, q)
}

// Generic implement EventFilter
func (pf *PodFilter) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeUnknown)

	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv generic object is not Pod, event %+v", e)
		return
	}
	pf.enqueuePodRelatedIngress(pod, q)
}
