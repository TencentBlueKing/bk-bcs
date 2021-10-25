/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

package portbindingcontroller

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/metrics"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	if !checkPortPoolAnnotationForPod(pod) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      pod.GetName(),
		Namespace: pod.GetNamespace(),
	}})
}

// Update implement EventFilter
func (pf *PodFilter) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeUpdate)
}

// Delete implement EventFilter
func (pf *PodFilter) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeDelete)

	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv delete object is not Pod, event %+v", e)
		return
	}
	if !checkPortPoolAnnotationForPod(pod) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      pod.GetName(),
		Namespace: pod.GetNamespace(),
	}})
}

// Generic implement EventFilter
func (pf *PodFilter) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(pf.filterName, metrics.EventTypeUnknown)
	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv delete object is not Pod, event %+v", e)
		return
	}
	if !checkPortPoolAnnotationForPod(pod) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      pod.GetName(),
		Namespace: pod.GetNamespace(),
	}})
}
