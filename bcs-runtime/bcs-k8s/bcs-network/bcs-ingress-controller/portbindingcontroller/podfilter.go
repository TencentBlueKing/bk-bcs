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

package portbindingcontroller

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
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

// Create implement EventFilter
func (pf *PodFilter) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeAdd)

	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv create object is not Pod, event %+v", e)
		return
	}
	if !checkPortPoolAnnotation(pod.Annotations) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      pod.GetName(),
		Namespace: pod.GetNamespace(),
	}})

	// check if related portBinding created success
	go checkPortBindingCreate(pf.cli, pod.GetNamespace(), pod.GetName())
}

// Update implement EventFilter
func (pf *PodFilter) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeUpdate)

	oldPod, okOld := e.ObjectOld.(*k8scorev1.Pod)
	newPod, okNew := e.ObjectNew.(*k8scorev1.Pod)
	if !okOld || !okNew {
		blog.Warnf("recv create object is not Pod, event %+v", e)
		return
	}

	if !checkPortPoolAnnotation(newPod.Annotations) && !checkPortPoolAnnotation(oldPod.Annotations) {
		return
	}

	if !checkPodNeedReconcile(oldPod, newPod) {
		blog.V(4).Infof("ignore pod[%s/%s] update", newPod.GetNamespace(), newPod.Name)
		return
	}

	// find portbinding related to updated pod
	portBinding := &networkextensionv1.PortBinding{}
	err := pf.cli.Get(context.TODO(), types.NamespacedName{
		Namespace: newPod.GetNamespace(),
		Name:      newPod.GetName(),
	}, portBinding)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Warnf("not found portbinding '%s/%s' related to updated pod",
				newPod.GetNamespace(), newPod.GetName())
			return
		}
		blog.Errorf("failed to get portbinding '%s/%s' related to updated pod: %s",
			newPod.GetNamespace(), newPod.GetName(), err.Error())
		return
	}

	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      newPod.GetName(),
		Namespace: newPod.GetNamespace(),
	}})

	// 如果删除portpool相关annotation，认为用户不再需要绑定端口，会在portBinding reconcile过程中删除相关PortBinding
	if !checkPortPoolAnnotation(newPod.Annotations) && checkPortPoolAnnotation(oldPod.Annotations) {
		go checkPortBindingDelete(pf.cli, newPod.GetNamespace(), newPod.GetName())
	}
}

// Delete implement EventFilter
func (pf *PodFilter) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeDelete)

	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv delete object is not Pod, event %+v", e)
		return
	}
	if !checkPortPoolAnnotation(pod.Annotations) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      pod.GetName(),
		Namespace: pod.GetNamespace(),
	}})

	go checkPortBindingDelete(pf.cli, pod.GetNamespace(), pod.GetName())
}

// Generic implement EventFilter
func (pf *PodFilter) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeUnknown)
	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv delete object is not Pod, event %+v", e)
		return
	}
	if !checkPortPoolAnnotation(pod.Annotations) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      pod.GetName(),
		Namespace: pod.GetNamespace(),
	}})
}
