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
	"context"
	"time"

	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
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

	// check if related portBinding created success
	go pf.checkPortBindingCreate(pod)
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
		blog.Warnf("recv create new object is not Pod, event %+v", e)
		return
	}

	if !checkPortPoolAnnotationForPod(newPod) {
		blog.V(4).Infof("ignore pod[%s/%s] update", newPod.GetNamespace(), newPod.Name)
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

	go pf.checkPortBindingDelete(pod)
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

// checkPortBindingCreate check if related portbinding create successfully
func (pf *PodFilter) checkPortBindingCreate(pod *k8scorev1.Pod) {
	blog.Infof("starts to check related portbinding %s/%s status", pod.GetNamespace(), pod.GetName())
	timeout := time.After(time.Minute)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			blog.Warnf("portbinding '%s/%s' is not ready, inc fail metric", pod.GetNamespace(), pod.GetName())
			metrics.IncreaseFailMetric(metrics.ObjectPortbinding, metrics.EventTypeAdd)
			return
		case <-ticker.C:
			portBinding := &networkextensionv1.PortBinding{}
			err := pf.cli.Get(context.TODO(), types.NamespacedName{
				Namespace: pod.GetNamespace(),
				Name:      pod.GetName(),
			}, portBinding)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					blog.V(5).Infof("not found portbinding '%s/%s' related to created pod",
						pod.GetNamespace(), pod.GetName())
					continue
				}
				blog.Warnf("failed to get portbinding '%s/%s' related to created pod: %s",
					pod.GetNamespace(), pod.GetName(), err.Error())
				continue
			}

			if portBinding.Status.Status == constant.PortBindingStatusReady {
				blog.Infof("portbinding '%s/%s' is ready", pod.GetNamespace(), pod.GetName())
				return
			}
		}
	}
}

// checkPortBindingDelete check if related portbinding delete successfully
func (pf *PodFilter) checkPortBindingDelete(pod *k8scorev1.Pod) {
	blog.Infof("starts to check portbinding %s/%s clean", pod.GetNamespace(), pod.GetName())
	timeout := time.After(time.Minute)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			blog.Warnf("portbinding '%s/%s' clean not finished, inc fail metric", pod.GetNamespace(),
				pod.GetName())
			metrics.IncreaseFailMetric(metrics.ObjectPortbinding, metrics.EventTypeDelete)
			return
		case <-ticker.C:
			portBinding := &networkextensionv1.PortBinding{}
			err := pf.cli.Get(context.TODO(), types.NamespacedName{
				Namespace: pod.GetNamespace(),
				Name:      pod.GetName(),
			}, portBinding)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					blog.Infof("portbinding '%s/%s' clean finish",
						pod.GetNamespace(), pod.GetName())
					return
				}
				blog.Warnf("failed to get portbinding '%s/%s' related to created pod: %s",
					pod.GetNamespace(), pod.GetName(), err.Error())
				continue
			}
		}
	}
}
