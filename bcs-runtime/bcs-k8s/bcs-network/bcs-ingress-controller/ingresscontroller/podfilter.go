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

package ingresscontroller

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/ingresscache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// PodFilter filter for pod event
type PodFilter struct {
	filterName   string
	cli          client.Client
	ingressCache ingresscache.IngressCache
}

// NewPodFilter create pod filter
func NewPodFilter(cli client.Client, ingressCache ingresscache.IngressCache) *PodFilter {
	return &PodFilter{
		filterName:   "pod",
		cli:          cli,
		ingressCache: ingressCache,
	}
}

func isServiceMatchesPod(svc *k8scorev1.Service, pod *k8scorev1.Pod) bool {
	labelSelector := labels.SelectorFromSet(labels.Set(svc.Spec.Selector))
	return labelSelector.Matches(labels.Set(pod.GetLabels()))
}

func (pf *PodFilter) enqueuePodRelatedIngress(pod *k8scorev1.Pod, q workqueue.RateLimitingInterface) {
	ingressMeta := pf.findPodIngresses(pod)
	if len(ingressMeta) == 0 {
		return
	}
	for _, meta := range ingressMeta {
		q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      meta.Name,
			Namespace: meta.Namespace,
		}})
	}
}

func (pf *PodFilter) findPodIngresses(pod *k8scorev1.Pod) []ingresscache.IngressMeta {
	// find ingresses of pod related services
	svcList := &k8scorev1.ServiceList{}
	err := pf.cli.List(context.TODO(), svcList, &client.ListOptions{Namespace: pod.GetNamespace()})
	if err != nil {
		blog.Warnf("list services in namespace %s failed, err %s", pod.GetNamespace(), err.Error())
		return nil
	}
	var retList []ingresscache.IngressMeta
	for _, svc := range svcList.Items {
		if isServiceMatchesPod(&svc, pod) {
			retList = append(retList, pf.ingressCache.GetRelatedIngressOfService(svc.GetNamespace(), svc.GetName())...)
		}
	}
	// find ingresses of pod related workloads
	for _, owner := range pod.GetOwnerReferences() {
		retList = append(retList, pf.ingressCache.GetRelatedIngressOfWorkload(owner.Kind, pod.GetNamespace(),
			owner.Name)...)
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

	oldPod, ok := e.ObjectOld.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv create old object is not Pod, event %+v", e)
		return
	}

	newPod, ok := e.ObjectNew.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv update object is not Pod, event %+v", e)
		return
	}
	if !checkPodNeedReconcile(oldPod, newPod) {
		blog.V(4).Infof("ignore pod[%s/%s] update", newPod.GetNamespace(), newPod.Name)
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
